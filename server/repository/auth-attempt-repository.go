package repository

import (
	"sync"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/config"
)

type AuthAttemptRepository struct {
}

type AuthAttempt struct {
	ID         string
	UserID     string
	Email      string
	Timestamp  time.Time
	Successful bool
}

var authAttemptRepository *AuthAttemptRepository
var authAttemptRepositoryOnce sync.Once

func GetAuthAttemptRepository() *AuthAttemptRepository {
	authAttemptRepositoryOnce.Do(func() {
		authAttemptRepository = &AuthAttemptRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS auth_attempts (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"user_id uuid NULL, " +
			"email VARCHAR NOT NULL, " +
			"timestamp TIMESTAMP NOT NULL, " +
			"successful BOOLEAN, " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_auth_attempts_user_id ON auth_attempts(user_id)")
		if err != nil {
			panic(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_auth_attempts_email ON auth_attempts(email)")
		if err != nil {
			panic(err)
		}
	})
	return authAttemptRepository
}

func (r *AuthAttemptRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	// No updates yet
}

func (r *AuthAttemptRepository) Create(e *AuthAttempt) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO auth_attempts "+
		"(user_id, email, timestamp, successful) "+
		"VALUES ($1, $2, $3, $4) "+
		"RETURNING id",
		e.UserID, e.Email, e.Timestamp, e.Successful).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *AuthAttemptRepository) RecordLoginAttempt(user *User, success bool) error {
	e := &AuthAttempt{
		UserID:     user.ID,
		Email:      user.Email,
		Timestamp:  time.Now(),
		Successful: success,
	}
	if err := r.Create(e); err != nil {
		return err
	}
	if err := r.checkBanUser(user); err != nil {
		return err
	}
	return nil
}

func (r *AuthAttemptRepository) checkBanUser(user *User) error {
	var lastSuccessfulLogin time.Time
	if err := GetDatabase().DB().QueryRow("SELECT timestamp FROM auth_attempts WHERE user_id = $1 AND successful = TRUE ORDER BY timestamp DESC LIMIT 1",
		user.ID).Scan(&lastSuccessfulLogin); err != nil {
		lastSuccessfulLogin = time.Unix(0, 0)
	}
	var numFailedLogins int
	limit := time.Now().Add(time.Second * time.Duration(GetConfig().LoginProtectionSlidingWindowSeconds*-1))
	if err := GetDatabase().DB().QueryRow("SELECT COUNT(id) FROM auth_attempts "+
		"WHERE user_id = $1 AND timestamp > $2 AND timestamp > $3",
		user.ID, limit, lastSuccessfulLogin).Scan(&numFailedLogins); err != nil {
		return err
	}
	if numFailedLogins >= GetConfig().LoginProtectionMaxFails {
		banExpiry := time.Now().Add(time.Minute * time.Duration(GetConfig().LoginProtectionBanMinutes))
		user.Disabled = true
		user.BanExpiry = &banExpiry
		if err := GetUserRepository().Update(user); err != nil {
			return err
		}
	}
	return nil
}
