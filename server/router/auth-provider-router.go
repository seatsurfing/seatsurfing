package router

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

type AuthProviderRouter struct {
}

type CreateAuthProviderRequest struct {
	Name                   string `json:"name" validate:"required,max=256"`
	ProviderType           int    `json:"providerType" validate:"required"`
	AuthURL                string `json:"authUrl" validate:"required,max=512"`
	TokenURL               string `json:"tokenUrl" validate:"required,max=512"`
	AuthStyle              int    `json:"authStyle"`
	Scopes                 string `json:"scopes" validate:"required,max=256"`
	UserInfoURL            string `json:"userInfoUrl" validate:"required,max=512"`
	UserInfoEmailField     string `json:"userInfoEmailField" validate:"required,max=256"`
	UserInfoFirstnameField string `json:"userInfoFirstnameField" validate:"max=256"`
	UserInfoLastnameField  string `json:"userInfoLastnameField" validate:"max=256"`
	ClientID               string `json:"clientId" validate:"required,max=256"`
	ClientSecret           string `json:"clientSecret" validate:"required,max=256"`
	LogoutURL              string `json:"logoutUrl" validate:"max=256"`
	ProfilePageURL         string `json:"profilePageUrl" validate:"max=256"`
}

type GetAuthProviderResponse struct {
	ID             string `json:"id"`
	ReadOnly       bool   `json:"readOnly"`
	OrganizationID string `json:"organizationId"`
	CreateAuthProviderRequest
}

type GetAuthProviderPublicResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (router *AuthProviderRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/org/{id}", router.listPublicForOrg).Methods("GET")
	s.HandleFunc("/{id}", router.getOne).Methods("GET")
	s.HandleFunc("/{id}", router.update).Methods("PUT")
	s.HandleFunc("/{id}", router.delete).Methods("DELETE")
	s.HandleFunc("/", router.create).Methods("POST")
	s.HandleFunc("/", router.getAll).Methods("GET")
}

func (router *AuthProviderRouter) listPublicForOrg(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	org, err := GetOrganizationRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	list, err := GetAuthProviderRepository().GetAll(org.ID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []*GetAuthProviderPublicResponse{}
	for _, e := range list {
		m := &GetAuthProviderPublicResponse{}
		m.ID = e.ID
		m.Name = e.Name
		res = append(res, m)
	}
	SendJSON(w, res)
}

func (router *AuthProviderRouter) getOne(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetAuthProviderRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	user := GetRequestUser(r)
	if !CanAdminOrg(user, e.OrganizationID) {
		SendForbidden(w)
		return
	}
	res := router.copyToRestModel(e)
	SendJSON(w, res)
}

func (router *AuthProviderRouter) getAll(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CanAdminOrg(user, user.OrganizationID) {
		SendForbidden(w)
		return
	}
	list, err := GetAuthProviderRepository().GetAll(user.OrganizationID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []*GetAuthProviderResponse{}
	for _, e := range list {
		m := router.copyToRestModel(e)
		res = append(res, m)
	}
	SendJSON(w, res)
}
func (router *AuthProviderRouter) validateCreateAuthProviderRequest(m *CreateAuthProviderRequest) bool {
	if m.ProviderType != int(OAuth2) {
		return false
	}
	if m.AuthStyle < 0 || m.AuthStyle > 2 {
		return false
	}
	if !ValidateURL(m.AuthURL) || !ValidateURL(m.TokenURL) || !ValidateURL(m.UserInfoURL) {
		return false
	}
	if m.LogoutURL != "" && !ValidateURL(m.LogoutURL) {
		return false
	}
	if m.ProfilePageURL != "" && !ValidateURL(m.ProfilePageURL) {
		return false
	}
	return true
}

func (router *AuthProviderRouter) update(w http.ResponseWriter, r *http.Request) {
	var m CreateAuthProviderRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	if !router.validateCreateAuthProviderRequest(&m) {
		SendBadRequest(w)
		return
	}
	vars := mux.Vars(r)
	e, err := GetAuthProviderRepository().GetOne(vars["id"])
	if err != nil {
		SendBadRequest(w)
		return
	}
	if e.ReadOnly {
		SendForbidden(w)
		return
	}
	user := GetRequestUser(r)
	if !CanAdminOrg(user, e.OrganizationID) {
		SendForbidden(w)
		return
	}

	existingAuthProvider, err := GetAuthProviderRepository().GetByName(e.OrganizationID, e.Name)
	if err == nil && existingAuthProvider != nil && existingAuthProvider.ID != e.ID {
		SendAlreadyExistsCode(w, ResponseCodeAuthProviderAlreadyExists)
		return
	}

	eNew := router.copyFromRestModel(&m)
	eNew.ID = e.ID
	eNew.OrganizationID = e.OrganizationID
	eNew.ReadOnly = e.ReadOnly
	if err := GetAuthProviderRepository().Update(eNew); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *AuthProviderRouter) delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetAuthProviderRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	user := GetRequestUser(r)
	if !CanAdminOrg(user, e.OrganizationID) {
		SendForbidden(w)
		return
	}
	if e.ReadOnly {
		SendForbidden(w)
		return
	}

	// Check if any users are bound to this auth provider
	hasUsers, err := GetUserRepository().HasAnyUserWithAuthProvider(e.ID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	if hasUsers {
		SendBadRequest(w)
		return
	}

	if err := GetAuthProviderRepository().Delete(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *AuthProviderRouter) create(w http.ResponseWriter, r *http.Request) {
	var m CreateAuthProviderRequest
	if err := UnmarshalValidateBody(r, &m); err != nil {
		log.Println(err)
		SendBadRequest(w)
		return
	}
	if !router.validateCreateAuthProviderRequest(&m) {
		SendBadRequest(w)
		return
	}

	user := GetRequestUser(r)
	if !CanAdminOrg(user, user.OrganizationID) {
		SendForbidden(w)
		return
	}

	e := router.copyFromRestModel(&m)
	e.OrganizationID = user.OrganizationID
	e.ReadOnly = false

	featureAuthProviders, _ := GetSettingsRepository().GetBool(user.OrganizationID, SettingFeatureAuthProviders.Name)
	if !featureAuthProviders {
		SendPaymentRequired(w)
		return
	}

	existingAuthProvider, err := GetAuthProviderRepository().GetByName(e.OrganizationID, e.Name)
	if err == nil && existingAuthProvider != nil {
		SendAlreadyExistsCode(w, ResponseCodeAuthProviderAlreadyExists)
		return
	}

	if err := GetAuthProviderRepository().Create(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendCreated(w, e.ID)
}

func (router *AuthProviderRouter) copyFromRestModel(m *CreateAuthProviderRequest) *AuthProvider {
	ClientSecretEncrypted, _ := EncryptString(m.ClientSecret)
	e := &AuthProvider{}
	e.Name = m.Name
	e.ClientID = m.ClientID
	e.ClientSecret = ClientSecretEncrypted
	e.AuthURL = m.AuthURL
	e.TokenURL = m.TokenURL
	e.AuthStyle = m.AuthStyle
	e.Scopes = m.Scopes
	e.UserInfoURL = m.UserInfoURL
	e.UserInfoEmailField = m.UserInfoEmailField
	e.UserInfoFirstnameField = m.UserInfoFirstnameField
	e.UserInfoLastnameField = m.UserInfoLastnameField
	e.ProviderType = m.ProviderType
	e.LogoutURL = m.LogoutURL
	e.ProfilePageURL = m.ProfilePageURL
	return e
}

func (router *AuthProviderRouter) copyToRestModel(e *AuthProvider) *GetAuthProviderResponse {
	ClientSecretDecrypted, err := DecryptString(e.ClientSecret)
	if err != nil || ClientSecretDecrypted == "" {
		// backward compatibility (client secret is stored in plain text)
		ClientSecretDecrypted = e.ClientSecret
	}
	m := &GetAuthProviderResponse{}
	m.ID = e.ID
	m.OrganizationID = e.OrganizationID
	m.Name = e.Name
	m.ClientID = e.ClientID
	m.ClientSecret = ClientSecretDecrypted
	m.AuthURL = e.AuthURL
	m.TokenURL = e.TokenURL
	m.AuthStyle = e.AuthStyle
	m.Scopes = e.Scopes
	m.UserInfoURL = e.UserInfoURL
	m.UserInfoEmailField = e.UserInfoEmailField
	m.UserInfoFirstnameField = e.UserInfoFirstnameField
	m.UserInfoLastnameField = e.UserInfoLastnameField
	m.ProviderType = e.ProviderType
	m.LogoutURL = e.LogoutURL
	m.ProfilePageURL = e.ProfilePageURL
	m.ReadOnly = e.ReadOnly
	return m
}
