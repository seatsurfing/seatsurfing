package repository

import (
	"sync"
	"time"
)

type SessionRepository struct {
}

type Session struct {
	ID      string
	UserID  string
	Device  string
	Created time.Time
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
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)")
		if err != nil {
			panic(err)
		}
	})
	return sesionRepository
}

func (r *SessionRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	// No updates yet
}

func (r *SessionRepository) Create(e *Session) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO sessions "+
		"(user_id, device, created) "+
		"VALUES ($1, $2, $3) "+
		"RETURNING id",
		e.UserID, e.Device, e.Created).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *SessionRepository) GetOne(id string) (*Session, error) {
	e := &Session{}
	err := GetDatabase().DB().QueryRow("SELECT id, user_id, device, created "+
		"FROM sessions "+
		"WHERE id = $1",
		id).Scan(&e.ID, &e.UserID, &e.Device, &e.Created)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *SessionRepository) GetOfUser(u *User) ([]*Session, error) {
	rows, err := GetDatabase().DB().Query("SELECT id, user_id, device, created "+
		"FROM sessions "+
		"WHERE user_id = $1",
		u.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		e := &Session{}
		err := rows.Scan(&e.ID, &e.UserID, &e.Device, &e.Created)
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
