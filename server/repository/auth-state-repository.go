package repository

import (
	"sync"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/api"
)

type AuthStateStore struct {
}

var authStateRepository *AuthStateStore
var authStateRepositoryOnce sync.Once

func GetAuthStateRepository() *AuthStateStore {
	authStateRepositoryOnce.Do(func() {
		authStateRepository = &AuthStateStore{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS auth_states (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"auth_provider_id uuid NOT NULL, " +
			"expiry TIMESTAMP NOT NULL, " +
			"auth_state_type INT NOT NULL, " +
			"payload VARCHAR NULL, " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
	})
	return authStateRepository
}

func (r *AuthStateStore) RunSchemaUpgrade(curVersion, targetVersion int) {
	if curVersion < 57 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE auth_states ADD COLUMN IF NOT EXISTS key VARCHAR NULL"); err != nil {
			panic(err)
		}
		if _, err := GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_auth_states_key_type ON auth_states(key, auth_state_type)"); err != nil {
			panic(err)
		}
		if _, err := GetDatabase().DB().Exec("ALTER TABLE auth_states DROP COLUMN IF EXISTS auth_provider_id"); err != nil {
			panic(err)
		}
	}
}

func (r *AuthStateStore) Create(e *AuthState) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO auth_states "+
		"(expiry, auth_state_type, payload, key) "+
		"VALUES ($1, $2, $3, $4) "+
		"RETURNING id",
		e.Expiry, e.AuthStateType, e.Payload, NullString(e.Key)).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *AuthStateStore) GetOne(id string) (*AuthState, error) {
	e := &AuthState{}
	var key NullString
	err := GetDatabase().DB().QueryRow("SELECT id, expiry, auth_state_type, payload, key "+
		"FROM auth_states "+
		"WHERE id = $1",
		id).Scan(&e.ID, &e.Expiry, &e.AuthStateType, &e.Payload, &key)
	if err != nil {
		return nil, err
	}
	e.Key = string(key)
	return e, nil
}

func (r *AuthStateStore) Delete(e *AuthState) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM auth_states WHERE id = $1", e.ID)
	return err
}

// GetOneActive returns the auth state with the given ID if it has not expired yet.
// Expiry is compared in SQL as timestamps round-trip as wall-clock time.
func (r *AuthStateStore) GetOneActive(id string) (*AuthState, error) {
	e := &AuthState{}
	var key NullString
	err := GetDatabase().DB().QueryRow("SELECT id, expiry, auth_state_type, payload, key "+
		"FROM auth_states "+
		"WHERE id = $1 AND expiry > $2",
		id, time.Now()).Scan(&e.ID, &e.Expiry, &e.AuthStateType, &e.Payload, &key)
	if err != nil {
		return nil, err
	}
	e.Key = string(key)
	return e, nil
}

// MarkForDeletion keeps the auth state valid for a short grace window instead
// of deleting it immediately, so duplicate verification requests for the same
// state (e.g. a double-mounted client) still succeed. The periodic cleanup
// removes the state once the shortened expiry has passed.
func (r *AuthStateStore) MarkForDeletion(e *AuthState, graceWindow time.Duration) error {
	_, err := GetDatabase().DB().Exec("UPDATE auth_states SET expiry = LEAST(expiry, $1) WHERE id = $2",
		time.Now().Add(graceWindow), e.ID)
	return err
}

func (r *AuthStateStore) GetActiveByPayloadAndType(payload string, authStateType AuthStateType) ([]*AuthState, error) {
	var result []*AuthState
	now := time.Now()
	rows, err := GetDatabase().DB().Query("SELECT id, expiry, auth_state_type, payload, key "+
		"FROM auth_states "+
		"WHERE payload = $1 AND auth_state_type = $2 AND expiry > $3",
		payload, authStateType, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &AuthState{}
		var key NullString
		if err := rows.Scan(&e.ID, &e.Expiry, &e.AuthStateType, &e.Payload, &key); err != nil {
			return nil, err
		}
		e.Key = string(key)
		result = append(result, e)
	}
	return result, nil
}

// GetActiveByKeyAndType returns all non-expired auth states of the given type
// whose Key matches, e.g. an OAuth provider ID, user ID or email used to look
// up or rate-limit states where the payload is a structured (JSON) blob.
func (r *AuthStateStore) GetActiveByKeyAndType(key string, authStateType AuthStateType) ([]*AuthState, error) {
	var result []*AuthState
	rows, err := GetDatabase().DB().Query("SELECT id, expiry, auth_state_type, payload, key "+
		"FROM auth_states "+
		"WHERE key = $1 AND auth_state_type = $2 AND expiry > $3",
		key, authStateType, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &AuthState{}
		var k NullString
		if err := rows.Scan(&e.ID, &e.Expiry, &e.AuthStateType, &e.Payload, &k); err != nil {
			return nil, err
		}
		e.Key = string(k)
		result = append(result, e)
	}
	return result, nil
}

func (r *AuthStateStore) DeleteExpired() error {
	now := time.Now()
	_, err := GetDatabase().DB().Exec("DELETE FROM auth_states WHERE expiry < $1", now)
	return err
}
