package test

import (
	"errors"
	"testing"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	"github.com/seatsurfing/seatsurfing/server/service"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func createTestRefreshTokenWithSession(t *testing.T, domain string) (*RefreshToken, *Session) {
	org := CreateTestOrg(domain)
	user := CreateTestUserInOrg(org)
	session := &Session{UserID: user.ID, Device: "test", Created: time.Now()}
	if err := GetSessionRepository().Create(session); err != nil {
		t.Fatal(err)
	}
	token, err := service.GetRefreshTokenService().Issue(user.ID, session.ID)
	CheckTestIsNil(t, err)
	return token, session
}

func TestRefreshTokenServiceIssue(t *testing.T) {
	ClearTestDB()
	token, session := createTestRefreshTokenWithSession(t, "test.com")
	CheckStringNotEmpty(t, token.ID)
	CheckTestString(t, session.ID, token.SessionID)
	CheckTestBool(t, true, token.Expiry.After(time.Now().Add(service.RefreshTokenLifetime-time.Minute)))
}

func TestRefreshTokenServiceConsume(t *testing.T) {
	ClearTestDB()
	token, _ := createTestRefreshTokenWithSession(t, "test.com")

	consumed, err := service.GetRefreshTokenService().Consume(token.ID)
	CheckTestIsNil(t, err)
	CheckTestString(t, token.ID, consumed.ID)
}

func TestRefreshTokenServiceConsumeInvalid(t *testing.T) {
	ClearTestDB()
	_, err := service.GetRefreshTokenService().Consume("00000000-0000-0000-0000-000000000000")
	CheckTestBool(t, true, errors.Is(err, service.ErrRefreshTokenInvalid))

	token, _ := createTestRefreshTokenWithSession(t, "test.com")
	if _, err := GetDatabase().DB().Exec("UPDATE refresh_tokens SET expiry = $2 WHERE id = $1", token.ID, time.Now().Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	_, err = service.GetRefreshTokenService().Consume(token.ID)
	CheckTestBool(t, true, errors.Is(err, service.ErrRefreshTokenInvalid))
}

func TestRefreshTokenServiceReplayWithinGracePeriod(t *testing.T) {
	ClearTestDB()
	token, session := createTestRefreshTokenWithSession(t, "test.com")

	_, err := service.GetRefreshTokenService().Consume(token.ID)
	CheckTestIsNil(t, err)
	_, err = service.GetRefreshTokenService().Consume(token.ID)
	CheckTestBool(t, true, errors.Is(err, service.ErrRefreshTokenReplayed))

	// Session is untouched
	_, err = GetSessionRepository().GetOne(session.ID)
	CheckTestIsNil(t, err)
}

func TestRefreshTokenServiceReuseRevokesSession(t *testing.T) {
	ClearTestDB()
	token, session := createTestRefreshTokenWithSession(t, "test.com")
	otherToken, err := service.GetRefreshTokenService().Issue(session.UserID, session.ID)
	CheckTestIsNil(t, err)

	_, err = service.GetRefreshTokenService().Consume(token.ID)
	CheckTestIsNil(t, err)
	// Pretend the rotation happened long ago
	if _, err := GetDatabase().DB().Exec("UPDATE refresh_tokens SET used_at = $2 WHERE id = $1",
		token.ID, time.Now().Add(-service.RefreshTokenReuseGracePeriod-time.Minute)); err != nil {
		t.Fatal(err)
	}

	_, err = service.GetRefreshTokenService().Consume(token.ID)
	CheckTestBool(t, true, errors.Is(err, service.ErrRefreshTokenReused))

	// The whole session family is gone, including tokens that were still unused
	_, err = GetSessionRepository().GetOne(session.ID)
	CheckTestBool(t, true, err != nil)
	_, err = service.GetRefreshTokenService().Consume(otherToken.ID)
	CheckTestBool(t, true, errors.Is(err, service.ErrRefreshTokenInvalid))
}

func TestRefreshTokenServiceReuseDoesNotAffectOtherSessions(t *testing.T) {
	ClearTestDB()
	token, _ := createTestRefreshTokenWithSession(t, "test.com")
	_, otherSession := createTestRefreshTokenWithSession(t, "other.com")

	_, err := service.GetRefreshTokenService().Consume(token.ID)
	CheckTestIsNil(t, err)
	if _, err := GetDatabase().DB().Exec("UPDATE refresh_tokens SET used_at = $2 WHERE id = $1",
		token.ID, time.Now().Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	_, err = service.GetRefreshTokenService().Consume(token.ID)
	CheckTestBool(t, true, errors.Is(err, service.ErrRefreshTokenReused))

	_, err = GetSessionRepository().GetOne(otherSession.ID)
	CheckTestIsNil(t, err)
}
