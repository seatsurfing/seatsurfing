package repository

import (
	"sync"
	"time"
)

type AuthStateRepository struct {
}

type AuthStateType int

const (
	AuthRequestState         AuthStateType = 1
	AuthResponseCache        AuthStateType = 2
	AuthAtlassian            AuthStateType = 3
	AuthMergeRequest         AuthStateType = 4
	AuthResetPasswordRequest AuthStateType = 5
)

type AuthState struct {
	ID             string
	AuthProviderID string
	Expiry         time.Time
	AuthStateType  AuthStateType
	Payload        string
}

var authStateRepository *AuthStateRepository
var authStateRepositoryOnce sync.Once

func GetAuthStateRepository() *AuthStateRepository {
	authStateRepositoryOnce.Do(func() {
		authStateRepository = &AuthStateRepository{}
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

func (r *AuthStateRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	// No updates yet
}

func (r *AuthStateRepository) Create(e *AuthState) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO auth_states "+
		"(auth_provider_id, expiry, auth_state_type, payload) "+
		"VALUES ($1, $2, $3, $4) "+
		"RETURNING id",
		e.AuthProviderID, e.Expiry, e.AuthStateType, e.Payload).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *AuthStateRepository) GetOne(id string) (*AuthState, error) {
	e := &AuthState{}
	err := GetDatabase().DB().QueryRow("SELECT id, auth_provider_id, expiry, auth_state_type, payload "+
		"FROM auth_states "+
		"WHERE id = $1",
		id).Scan(&e.ID, &e.AuthProviderID, &e.Expiry, &e.AuthStateType, &e.Payload)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *AuthStateRepository) Delete(e *AuthState) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM auth_states WHERE id = $1", e.ID)
	return err
}

func (r *AuthStateRepository) DeleteExpired() error {
	now := time.Now()
	_, err := GetDatabase().DB().Exec("DELETE FROM auth_states WHERE expiry < $1", now)
	return err
}

func (r *AuthStateRepository) GetByAuthProviderID(authProviderID string) ([]*AuthState, error) {
	var result []*AuthState
	rows, err := GetDatabase().DB().Query("SELECT id, auth_provider_id, expiry, auth_state_type, payload "+
		"FROM auth_states "+
		"WHERE auth_provider_id = $1",
		authProviderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &AuthState{}
		err = rows.Scan(&e.ID, &e.AuthProviderID, &e.Expiry, &e.AuthStateType, &e.Payload)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}
