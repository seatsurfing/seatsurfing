package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

type BackplaneOrganizationRouter struct {
}

type BackplaneOrganizationGetNumUsersResponse struct {
	OrganizationID string `json:"organizationId"`
	NumUsers       int    `json:"numUsers"`
}

type BackplaneOrganizationSetCloudEnabledRequest struct {
	CloudEnabled bool `json:"cloudEnabled"`
}

func (router *BackplaneOrganizationRouter) setupRoutes(s *mux.Router) {
	s.HandleFunc("/{id}/numUsers", router.getNumUsers).Methods("GET")
	s.HandleFunc("/{id}/cloudEnabled", router.setCloudEnabled).Methods("PUT")
}

func (router *BackplaneOrganizationRouter) getNumUsers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	org, err := GetOrganizationRepository().GetOne(vars["id"])
	if err != nil || org == nil {
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}
	//numUsers, err := GetUsexrRepository().GetNumUsers(org.ID)
}

func (router *BackplaneOrganizationRouter) setCloudEnabled(w http.ResponseWriter, r *http.Request) {
}
