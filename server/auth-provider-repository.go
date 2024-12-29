package main

import (
	"sync"
)

type AuthProviderRepository struct {
}

type AuthProviderType int

const (
	OAuth2 AuthProviderType = 1
)

type AuthProvider struct {
	ID                 string
	OrganizationID     string
	Name               string
	ProviderType       int
	AuthURL            string
	TokenURL           string
	AuthStyle          int
	Scopes             string
	UserInfoURL        string
	UserInfoEmailField string
	ClientID           string
	ClientSecret       string
	LogoutURL          string
}

var authProviderRepository *AuthProviderRepository
var authProviderRepositoryOnce sync.Once

func GetAuthProviderRepository() *AuthProviderRepository {
	authProviderRepositoryOnce.Do(func() {
		authProviderRepository = &AuthProviderRepository{}
		_, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS auth_providers (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"organization_id uuid NOT NULL, " +
			"name VARCHAR NOT NULL, " +
			"provider_type INT NOT NULL, " +
			"auth_url VARCHAR NOT NULL, " +
			"token_url VARCHAR NOT NULL, " +
			"auth_style INT NOT NULL, " +
			"scopes VARCHAR NOT NULL, " +
			"userinfo_url VARCHAR NOT NULL, " +
			"userinfo_email_field VARCHAR NOT NULL, " +
			"client_id VARCHAR NOT NULL, " +
			"client_secret VARCHAR NOT NULL, " +
			"PRIMARY KEY (id))")
		if err != nil {
			panic(err)
		}
		_, err = GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_auth_providers_organization_id ON auth_providers(organization_id)")
		if err != nil {
			panic(err)
		}
	})
	return authProviderRepository
}

func (r *AuthProviderRepository) RunSchemaUpgrade(curVersion, targetVersion int) {
	if curVersion < 17 {
		if _, err := GetDatabase().DB().Exec("ALTER TABLE auth_providers " +
			"ADD COLUMN logout_url VARCHAR NOT NULL DEFAULT ''"); err != nil {
			panic(err)
		}
	}
}

func (r *AuthProviderRepository) Create(e *AuthProvider) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO auth_providers "+
		"(organization_id, name, provider_type, auth_url, token_url, auth_style, scopes, userinfo_url, userinfo_email_field, client_id, client_secret, logout_url) "+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) "+
		"RETURNING id",
		e.OrganizationID, e.Name, e.ProviderType, e.AuthURL, e.TokenURL, e.AuthStyle, e.Scopes, e.UserInfoURL, e.UserInfoEmailField, e.ClientID, e.ClientSecret, e.LogoutURL).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return nil
}

func (r *AuthProviderRepository) GetOne(id string) (*AuthProvider, error) {
	e := &AuthProvider{}
	err := GetDatabase().DB().QueryRow("SELECT id, organization_id, name, provider_type, auth_url, token_url, auth_style, scopes, userinfo_url, userinfo_email_field, client_id, client_secret, logout_url "+
		"FROM auth_providers "+
		"WHERE id = $1",
		id).Scan(&e.ID, &e.OrganizationID, &e.Name, &e.ProviderType, &e.AuthURL, &e.TokenURL, &e.AuthStyle, &e.Scopes, &e.UserInfoURL, &e.UserInfoEmailField, &e.ClientID, &e.ClientSecret, &e.LogoutURL)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *AuthProviderRepository) GetAll(organizationID string) ([]*AuthProvider, error) {
	var result []*AuthProvider
	rows, err := GetDatabase().DB().Query("SELECT id, organization_id, name, provider_type, auth_url, token_url, auth_style, scopes, userinfo_url, userinfo_email_field, client_id, client_secret, logout_url "+
		"FROM auth_providers "+
		"WHERE organization_id = $1 "+
		"ORDER BY name", organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &AuthProvider{}
		err = rows.Scan(&e.ID, &e.OrganizationID, &e.Name, &e.ProviderType, &e.AuthURL, &e.TokenURL, &e.AuthStyle, &e.Scopes, &e.UserInfoURL, &e.UserInfoEmailField, &e.ClientID, &e.ClientSecret, &e.LogoutURL)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *AuthProviderRepository) Update(e *AuthProvider) error {
	_, err := GetDatabase().DB().Exec("UPDATE auth_providers SET "+
		"organization_id = $1, "+
		"name = $2, "+
		"provider_type = $3, "+
		"auth_url = $4, "+
		"token_url = $5, "+
		"auth_style = $6, "+
		"scopes = $7, "+
		"userinfo_url = $8, "+
		"userinfo_email_field = $9, "+
		"client_id = $10, "+
		"client_secret = $11, "+
		"logout_url = $12 "+
		"WHERE id = $13",
		e.OrganizationID, e.Name, e.ProviderType, e.AuthURL, e.TokenURL, e.AuthStyle, e.Scopes, e.UserInfoURL, e.UserInfoEmailField, e.ClientID, e.ClientSecret, e.LogoutURL, e.ID)
	return err
}

func (r *AuthProviderRepository) Delete(e *AuthProvider) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM auth_providers WHERE id = $1", e.ID)
	return err
}

func (r *AuthProviderRepository) DeleteAll(organizationID string) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM auth_providers WHERE organization_id = $1", organizationID)
	return err
}
