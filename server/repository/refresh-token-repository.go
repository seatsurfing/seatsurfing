package repository

import (
	"sync"
	"time"
)

type RefreshTokenRepository struct {
}

type RefreshToken struct {
	ID      string
	UserID  string
	Created time.Time
	Expiry  time.Time
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
	// No updates yet
}

func (r *RefreshTokenRepository) Create(e *RefreshToken) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO refresh_tokens "+
		"(user_id, created, expiry) "+
		"VALUES ($1, $2, $3) "+
		"RETURNING id",
		e.UserID, e.Created, e.Expiry).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *RefreshTokenRepository) GetOne(id string) (*RefreshToken, error) {
	e := &RefreshToken{}
	err := GetDatabase().DB().QueryRow("SELECT id, user_id, created, expiry "+
		"FROM refresh_tokens "+
		"WHERE id = $1",
		id).Scan(&e.ID, &e.UserID, &e.Created, &e.Expiry)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *RefreshTokenRepository) Delete(e *RefreshToken) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM refresh_tokens WHERE id = $1", e.ID)
	return err
}

func (r *RefreshTokenRepository) DeleteExpired() error {
	now := time.Now()
	_, err := GetDatabase().DB().Exec("DELETE FROM refresh_tokens WHERE expiry < $1", now)
	return err
}

func (r *RefreshTokenRepository) DeleteOfUser(u *User) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM refresh_tokens WHERE user_id = $1", u.ID)
	return err
}
