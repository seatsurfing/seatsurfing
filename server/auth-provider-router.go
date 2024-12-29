package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

type AuthProviderRouter struct {
}

type CreateAuthProviderRequest struct {
	Name               string `json:"name" validate:"required"`
	ProviderType       int    `json:"providerType" validate:"required"`
	AuthURL            string `json:"authUrl" validate:"required"`
	TokenURL           string `json:"tokenUrl" validate:"required"`
	AuthStyle          int    `json:"authStyle"`
	Scopes             string `json:"scopes" validate:"required"`
	UserInfoURL        string `json:"userInfoUrl" validate:"required"`
	UserInfoEmailField string `json:"userInfoEmailField" validate:"required"`
	ClientID           string `json:"clientId" validate:"required"`
	ClientSecret       string `json:"clientSecret" validate:"required"`
	LogoutURL          string `json:"logoutUrl"`
}

type GetAuthProviderResponse struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organizationId"`
	CreateAuthProviderRequest
}

type GetAuthProviderPublicResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (router *AuthProviderRouter) setupRoutes(s *mux.Router) {
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
func (router *AuthProviderRouter) update(w http.ResponseWriter, r *http.Request) {
	var m CreateAuthProviderRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	vars := mux.Vars(r)
	e, err := GetAuthProviderRepository().GetOne(vars["id"])
	if err != nil {
		SendBadRequest(w)
		return
	}
	user := GetRequestUser(r)
	if !CanAdminOrg(user, e.OrganizationID) {
		SendForbidden(w)
		return
	}
	eNew := router.copyFromRestModel(&m)
	eNew.ID = e.ID
	eNew.OrganizationID = e.OrganizationID
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
	user := GetRequestUser(r)
	e := router.copyFromRestModel(&m)
	e.OrganizationID = user.OrganizationID
	if !CanAdminOrg(user, e.OrganizationID) {
		SendForbidden(w)
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
	e := &AuthProvider{}
	e.Name = m.Name
	e.ClientID = m.ClientID
	e.ClientSecret = m.ClientSecret
	e.AuthURL = m.AuthURL
	e.TokenURL = m.TokenURL
	e.AuthStyle = m.AuthStyle
	e.Scopes = m.Scopes
	e.UserInfoURL = m.UserInfoURL
	e.UserInfoEmailField = m.UserInfoEmailField
	e.ProviderType = m.ProviderType
	e.LogoutURL = m.LogoutURL
	return e
}

func (router *AuthProviderRouter) copyToRestModel(e *AuthProvider) *GetAuthProviderResponse {
	m := &GetAuthProviderResponse{}
	m.ID = e.ID
	m.OrganizationID = e.OrganizationID
	m.Name = e.Name
	m.ClientID = e.ClientID
	m.ClientSecret = e.ClientSecret
	m.AuthURL = e.AuthURL
	m.TokenURL = e.TokenURL
	m.AuthStyle = e.AuthStyle
	m.Scopes = e.Scopes
	m.UserInfoURL = e.UserInfoURL
	m.UserInfoEmailField = e.UserInfoEmailField
	m.ProviderType = e.ProviderType
	m.LogoutURL = e.LogoutURL
	return m
}
