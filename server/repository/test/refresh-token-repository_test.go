package test

import (
	"database/sql"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func createTestRefreshToken(t *testing.T, expiry time.Time) *RefreshToken {
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	session := &Session{UserID: user.ID, Device: "test", Created: time.Now()}
	if err := GetSessionRepository().Create(session); err != nil {
		t.Fatal(err)
	}
	token := &RefreshToken{UserID: user.ID, SessionID: session.ID, Created: time.Now(), Expiry: expiry}
	if err := GetRefreshTokenRepository().Create(token); err != nil {
		t.Fatal(err)
	}
	return token
}

func TestRefreshTokenRepositoryConsume(t *testing.T) {
	ClearTestDB()
	token := createTestRefreshToken(t, time.Now().Add(time.Hour))

	consumed, err := GetRefreshTokenRepository().Consume(token.ID, time.Now())
	CheckTestIsNil(t, err)
	CheckTestString(t, token.ID, consumed.ID)
	CheckTestString(t, token.UserID, consumed.UserID)
	CheckTestString(t, token.SessionID, consumed.SessionID)
	CheckTestBool(t, true, consumed.UsedAt != nil)

	// A used token cannot be consumed again, but stays readable
	_, err = GetRefreshTokenRepository().Consume(token.ID, time.Now())
	CheckTestBool(t, true, errors.Is(err, sql.ErrNoRows))
	stored, err := GetRefreshTokenRepository().GetOne(token.ID)
	CheckTestIsNil(t, err)
	CheckTestBool(t, true, stored.UsedAt != nil)

	// Used tokens do not count as live tokens of the session
	count, err := GetRefreshTokenRepository().GetCountBySession(token.SessionID)
	CheckTestIsNil(t, err)
	CheckTestInt(t, 0, count)
}

func TestRefreshTokenRepositoryConsumeExpired(t *testing.T) {
	ClearTestDB()
	token := createTestRefreshToken(t, time.Now().Add(-time.Minute))

	_, err := GetRefreshTokenRepository().Consume(token.ID, time.Now())
	CheckTestBool(t, true, errors.Is(err, sql.ErrNoRows))
}

func TestRefreshTokenRepositoryConsumeUnknown(t *testing.T) {
	ClearTestDB()
	_, err := GetRefreshTokenRepository().Consume("00000000-0000-0000-0000-000000000000", time.Now())
	CheckTestBool(t, true, errors.Is(err, sql.ErrNoRows))
}

func TestRefreshTokenRepositoryConsumeConcurrent(t *testing.T) {
	ClearTestDB()
	token := createTestRefreshToken(t, time.Now().Add(time.Hour))

	const workers = 20
	var successes atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if _, err := GetRefreshTokenRepository().Consume(token.ID, time.Now()); err == nil {
				successes.Add(1)
			}
		}()
	}
	close(start)
	wg.Wait()
	CheckTestInt(t, 1, int(successes.Load()))
}
