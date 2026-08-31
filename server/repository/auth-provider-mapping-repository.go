package repository

import (
	"sync"

	. "github.com/seatsurfing/seatsurfing/server/api"
)

type AuthProviderMappingStore struct {
}

var authProviderMappingRepository *AuthProviderMappingStore
var authProviderMappingRepositoryOnce sync.Once

func GetAuthProviderMappingRepository() *AuthProviderMappingStore {
	authProviderMappingRepositoryOnce.Do(func() {
		authProviderMappingRepository = &AuthProviderMappingStore{}
		if _, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS auth_provider_mappings (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"auth_provider_id uuid NOT NULL, " +
			"claim_value VARCHAR NOT NULL, " +
			"target_type VARCHAR NOT NULL, " +
			"target_id uuid NOT NULL, " +
			"PRIMARY KEY (id))"); err != nil {
			panic(err)
		}
		if _, err := GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_auth_provider_mappings_provider " +
			"ON auth_provider_mappings (auth_provider_id)"); err != nil {
			panic(err)
		}
	})
	return authProviderMappingRepository
}

func (r *AuthProviderMappingStore) RunSchemaUpgrade(curVersion, targetVersion int) {
}

func (r *AuthProviderMappingStore) Create(e *AuthProviderMapping) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO auth_provider_mappings "+
		"(auth_provider_id, claim_value, target_type, target_id) "+
		"VALUES ($1, $2, $3, $4) "+
		"RETURNING id",
		e.AuthProviderID, e.ClaimValue, e.TargetType, e.TargetID).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *AuthProviderMappingStore) GetAll(authProviderID string) ([]*AuthProviderMapping, error) {
	var result []*AuthProviderMapping
	rows, err := GetDatabase().DB().Query("SELECT id, auth_provider_id, claim_value, target_type, target_id "+
		"FROM auth_provider_mappings "+
		"WHERE auth_provider_id = $1 "+
		"ORDER BY claim_value",
		authProviderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &AuthProviderMapping{}
		if err := rows.Scan(&e.ID, &e.AuthProviderID, &e.ClaimValue, &e.TargetType, &e.TargetID); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

// DeleteAllForProvider removes every mapping of an auth provider, used when
// the provider itself is deleted or its mappings are replaced wholesale.
func (r *AuthProviderMappingStore) DeleteAllForProvider(authProviderID string) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM auth_provider_mappings WHERE auth_provider_id = $1", authProviderID)
	return err
}

// DeleteAllForTarget removes mappings pointing at a role or group that is
// being deleted, so that no mapping outlives what it targets.
func (r *AuthProviderMappingStore) DeleteAllForTarget(targetType, targetID string) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM auth_provider_mappings WHERE target_type = $1 AND target_id = $2", targetType, targetID)
	return err
}
