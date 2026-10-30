package service

import (
	"database/sql"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	. "github.com/seatsurfing/seatsurfing/server/repository"
)

// RefreshTokenLifetime is how long an issued refresh token stays valid.
const RefreshTokenLifetime = 28 * 24 * time.Hour

// RefreshTokenReuseGracePeriod is the time after rotation during which a
// replay of the same refresh token is merely rejected. This absorbs benign
// races (several browser tabs, client retries after a lost response). A replay
// after the grace period is treated as token theft and revokes the session.
const RefreshTokenReuseGracePeriod = 15 * time.Second

var (
	// ErrRefreshTokenInvalid is returned for unknown, expired and malformed tokens.
	ErrRefreshTokenInvalid = errors.New("refresh token invalid")
	// ErrRefreshTokenReplayed is returned if a token that was rotated a moment
	// ago is presented again. The session is left untouched.
	ErrRefreshTokenReplayed = errors.New("refresh token already used")
	// ErrRefreshTokenReused is returned if a rotated token is presented again
	// after the grace period. The session has been revoked.
	ErrRefreshTokenReused = errors.New("refresh token reuse detected")
)

// RefreshTokenService holds the refresh token rotation logic.
type RefreshTokenService struct{}

var refreshTokenService *RefreshTokenService
var refreshTokenServiceOnce sync.Once

func GetRefreshTokenService() *RefreshTokenService {
	refreshTokenServiceOnce.Do(func() {
		refreshTokenService = &RefreshTokenService{}
	})
	return refreshTokenService
}

// Issue creates a new refresh token for the given user and session.
func (s *RefreshTokenService) Issue(userID, sessionID string) (*RefreshToken, error) {
	// Timestamps are stored in timezone-less columns and read back as UTC, so
	// always write UTC to keep elapsed-time comparisons correct.
	now := time.Now().UTC()
	e := &RefreshToken{
		UserID:    userID,
		SessionID: sessionID,
		Created:   now,
		Expiry:    now.Add(RefreshTokenLifetime),
	}
	if err := GetRefreshTokenRepository().Create(e); err != nil {
		return nil, err
	}
	return e, nil
}

// Consume validates a refresh token and marks it as used. It is atomic: of
// several concurrent calls with the same token at most one succeeds. Presenting
// an already used token fails with ErrRefreshTokenReplayed within the grace
// period and, afterwards, revokes the token's session and fails with
// ErrRefreshTokenReused.
func (s *RefreshTokenService) Consume(id string) (*RefreshToken, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrRefreshTokenInvalid
	}
	id = parsed.String()
	now := time.Now().UTC()
	token, err := GetRefreshTokenRepository().Consume(id, now)
	if err == nil {
		return token, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	// Not consumable: unknown, expired or already used.
	existing, err := GetRefreshTokenRepository().GetOne(id)
	if err != nil || existing == nil || existing.UsedAt == nil || !existing.Expiry.After(now) {
		return nil, ErrRefreshTokenInvalid
	}
	if now.Sub(*existing.UsedAt) <= RefreshTokenReuseGracePeriod {
		return nil, ErrRefreshTokenReplayed
	}
	s.revokeSession(existing)
	return nil, ErrRefreshTokenReused
}

func (s *RefreshTokenService) revokeSession(token *RefreshToken) {
	session, err := GetSessionRepository().GetOne(token.SessionID)
	if err != nil || session == nil {
		return
	}
	log.Printf("Refresh token reuse detected, revoking session %s of user %s\n", session.ID, session.UserID)
	if err := GetSessionRepository().Delete(session); err != nil {
		log.Println("Error revoking session after refresh token reuse: " + err.Error())
	}
}
