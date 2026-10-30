package repository

import (
	"sync"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/api"
)

type RefreshTokenRepository struct {
}

type RefreshToken struct {
	ID        string
	UserID    string
	SessionID string
	Created   time.Time
	Expiry    time.Time
	// UsedAt is set once the token has been rotated. Used tokens are kept
	// (until they expire) so that a replay can be told apart from an unknown
	// token and treated as a compromise signal.
	UsedAt *time.Time
}

var refreshTokenRepository *RefreshTokenRepository
var refreshTokenRepositoryOnce sync.Once

func GetRefreshTokenRepository() *RefreshTokenRepository {
	refreshTokenRepositoryOnce.Do(func() {
		refreshTokenRepository = &RefreshTokenRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS refresh_tokens (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"user_id uuid NOT NULL, " +
			"created TIMESTAMP NOT NULL, " +
			"expiry TIMESTAMP NOT NULL, " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
	})
	return refreshTokenRepository
}

func (r *RefreshTokenRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	if curVersion < 34 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE refresh_tokens " +
			"ADD COLUMN IF NOT EXISTS session_id uuid NOT NULL DEFAULT '" + GetSettingsRepository().GetNullUUID() + "'"); err != nil {
			panic(err)
		}
	}
	if curVersion < 66 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE refresh_tokens " +
			"ADD COLUMN IF NOT EXISTS used_at TIMESTAMP NULL"); err != nil {
			panic(err)
		}
		if _, err := GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_refresh_tokens_session_id " +
			"ON refresh_tokens(session_id)"); err != nil {
			panic(err)
		}
	}
}

func (r *RefreshTokenRepository) Create(e *RefreshToken) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO refresh_tokens "+
		"(user_id, session_id, created, expiry) "+
		"VALUES ($1, $2, $3, $4) "+
		"RETURNING id",
		e.UserID, e.SessionID, e.Created, e.Expiry).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *RefreshTokenRepository) GetOne(id string) (*RefreshToken, error) {
	e := &RefreshToken{}
	err := GetDatabase().DB().QueryRow("SELECT id, user_id, session_id, created, expiry, used_at "+
		"FROM refresh_tokens "+
		"WHERE id = $1",
		id).Scan(&e.ID, &e.UserID, &e.SessionID, &e.Created, &e.Expiry, &e.UsedAt)
	if err != nil {
		return nil, err
	}
	return e, nil
}

// Consume atomically marks an unused, unexpired refresh token as used and
// returns it. Concurrent calls for the same token serialize on the row, so
// exactly one of them succeeds; all others get sql.ErrNoRows.
func (r *RefreshTokenRepository) Consume(id string, now time.Time) (*RefreshToken, error) {
	e := &RefreshToken{}
	err := GetDatabase().DB().QueryRow("UPDATE refresh_tokens SET used_at = $2 "+
		"WHERE id = $1 AND used_at IS NULL AND expiry > $2 "+
		"RETURNING id, user_id, session_id, created, expiry, used_at",
		id, now).Scan(&e.ID, &e.UserID, &e.SessionID, &e.Created, &e.Expiry, &e.UsedAt)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *RefreshTokenRepository) Delete(e *RefreshToken) error {
	// Hint: Do not delete session here, because it might be used by other refresh tokens.
	// Instead, delete the session when deleting expired tokens or when deleting all tokens of a user.
	// This function is only used when refreshing a token, so the session should not be deleted here.
	_, err := GetDatabase().DB().Exec("DELETE FROM refresh_tokens WHERE id = $1", e.ID)
	return err
}

func (r *RefreshTokenRepository) DeleteExpired() error {
	now := time.Now()
	// Only delete refresh tokens that are expired
	// Do NOT delete the sessions here - sessions are managed separately and may have other valid refresh tokens
	// Orphaned sessions (with no valid refresh tokens) will be cleaned up by session expiry
	_, err := GetDatabase().DB().Exec("DELETE FROM refresh_tokens WHERE expiry < $1", now)
	return err
}

func (r *RefreshTokenRepository) DeleteOfUser(u *User) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM refresh_tokens WHERE user_id = $1", u.ID)
	return err
}

func (r *RefreshTokenRepository) GetCountBySession(sessionID string) (int, error) {
	var count int
	err := GetDatabase().DB().QueryRow("SELECT COUNT(*) FROM refresh_tokens WHERE session_id = $1 AND used_at IS NULL AND expiry > NOW()", sessionID).Scan(&count)
	return count, err
}
