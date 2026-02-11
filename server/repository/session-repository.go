package repository

import (
	"sync"
	"time"
)

type SessionRepository struct {
}

type Session struct {
	ID           string
	UserID       string
	Device       string
	Created      time.Time
	LastActivity time.Time
	ExpiresAt    time.Time
}

var sesionRepository *SessionRepository
var sessionRepositoryOnce sync.Once

func GetSessionRepository() *SessionRepository {
	sessionRepositoryOnce.Do(func() {
		sesionRepository = &SessionRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS sessions (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"user_id uuid NOT NULL, " +
			"device VARCHAR NOT NULL, " +
			"created TIMESTAMP NOT NULL, " +
			"last_activity TIMESTAMP NOT NULL DEFAULT NOW(), " +
			"expires_at TIMESTAMP NOT NULL DEFAULT (NOW() + INTERVAL '90 days'), " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
		if _, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)"); err != nil {
			panic(err)
		}
		if _, err := GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at)"); err != nil {
			panic(err)
		}
	})
	return sesionRepository
}

func (r *SessionRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	// no schema changes yet
}

func (r *SessionRepository) Create(e *Session) error {
	var id string
	if e.ExpiresAt.IsZero() {
		e.ExpiresAt = time.Now().Add(90 * 24 * time.Hour) // 90 days default
	}
	if e.LastActivity.IsZero() {
		e.LastActivity = e.Created
	}
	err := GetDatabase().DB().QueryRow("INSERT INTO sessions "+
		"(user_id, device, created, last_activity, expires_at) "+
		"VALUES ($1, $2, $3, $4, $5) "+
		"RETURNING id",
		e.UserID, e.Device, e.Created, e.LastActivity, e.ExpiresAt).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *SessionRepository) GetOne(id string) (*Session, error) {
	e := &Session{}
	err := GetDatabase().DB().QueryRow("SELECT id, user_id, device, created, last_activity, expires_at "+
		"FROM sessions "+
		"WHERE id = $1 AND expires_at > NOW()",
		id).Scan(&e.ID, &e.UserID, &e.Device, &e.Created, &e.LastActivity, &e.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *SessionRepository) GetOfUser(u *User) ([]*Session, error) {
	rows, err := GetDatabase().DB().Query("SELECT id, user_id, device, created, last_activity, expires_at "+
		"FROM sessions "+
		"WHERE user_id = $1 AND expires_at > NOW() "+
		"ORDER BY last_activity DESC",
		u.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		e := &Session{}
		err := rows.Scan(&e.ID, &e.UserID, &e.Device, &e.Created, &e.LastActivity, &e.ExpiresAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, e)
	}
	return sessions, nil
}

func (r *SessionRepository) Delete(e *Session) error {
	_, err := GetDatabase().DB().Exec("WITH deleted_rows as ("+
		"DELETE FROM sessions WHERE id = $1 RETURNING id) "+
		"DELETE FROM refresh_tokens WHERE session_id IN (SELECT id::uuid FROM deleted_rows)", e.ID)
	return err
}

func (r *SessionRepository) DeleteOfUser(u *User) error {
	_, err := GetDatabase().DB().Exec("WITH deleted_rows as ("+
		"DELETE FROM sessions WHERE user_id = $1 RETURNING id) "+
		"DELETE FROM refresh_tokens WHERE session_id IN (SELECT id::uuid FROM deleted_rows)", u.ID)
	return err
}

func (r *SessionRepository) UpdateActivity(sessionID string) error {
	_, err := GetDatabase().DB().Exec("UPDATE sessions SET last_activity = NOW() WHERE id = $1", sessionID)
	return err
}

func (r *SessionRepository) DeleteExpired() error {
	_, err := GetDatabase().DB().Exec("WITH deleted_rows as (" +
		"DELETE FROM sessions WHERE expires_at < NOW() RETURNING id) " +
		"DELETE FROM refresh_tokens WHERE session_id IN (SELECT id::uuid FROM deleted_rows)")
	return err
}

func (r *SessionRepository) GetActiveSessionCount(u *User) (int, error) {
	var count int
	err := GetDatabase().DB().QueryRow("SELECT COUNT(*) FROM sessions WHERE user_id = $1 AND expires_at > NOW()", u.ID).Scan(&count)
	return count, err
}

func (r *SessionRepository) DeleteOldestSession(u *User) error {
	_, err := GetDatabase().DB().Exec(
		"WITH oldest_session AS ("+
			"SELECT id FROM sessions WHERE user_id = $1 AND expires_at > NOW() "+
			"ORDER BY last_activity ASC LIMIT 1"+
			") "+
			"DELETE FROM sessions WHERE id IN (SELECT id FROM oldest_session)",
		u.ID)
	return err
}
