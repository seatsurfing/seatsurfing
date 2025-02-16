package router

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/repository"
)

type SpaceAttributeRouter struct {
}

type CreateSpaceAttributeRequest struct {
	Label              string `json:"label" validate:"required"`
	Type               int    `json:"type"`
	SpaceApplicable    bool   `json:"spaceApplicable"`
	LocationApplicable bool   `json:"locationApplicable"`
}

type GetSpaceAttributeResponse struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organizationId"`
	CreateSpaceAttributeRequest
}

func (router *SpaceAttributeRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/{id}", router.getOne).Methods("GET")
	s.HandleFunc("/{id}", router.update).Methods("PUT")
	s.HandleFunc("/{id}", router.delete).Methods("DELETE")
	s.HandleFunc("/", router.create).Methods("POST")
	s.HandleFunc("/", router.getAll).Methods("GET")
}

func (router *SpaceAttributeRouter) getOne(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetSpaceAttributeRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	user := GetRequestUser(r)
	if !CanAccessOrg(user, e.OrganizationID) {
		SendForbidden(w)
		return
	}
	res := router.copyToRestModel(e)
	SendJSON(w, res)
}

func (router *SpaceAttributeRouter) getAll(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	list, err := GetSpaceAttributeRepository().GetAll(user.OrganizationID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []*GetSpaceAttributeResponse{}
	for _, e := range list {
		m := router.copyToRestModel(e)
		res = append(res, m)
	}
	SendJSON(w, res)
}

func (router *SpaceAttributeRouter) update(w http.ResponseWriter, r *http.Request) {
	var m CreateSpaceAttributeRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	vars := mux.Vars(r)
	e, err := GetSpaceAttributeRepository().GetOne(vars["id"])
	if err != nil {
		SendBadRequest(w)
		return
	}
	user := GetRequestUser(r)
	if !CanSpaceAdminOrg(user, e.OrganizationID) {
		SendForbidden(w)
		return
	}
	eNew := router.copyFromRestModel(&m)
	eNew.ID = e.ID
	eNew.OrganizationID = e.OrganizationID
	if err := GetSpaceAttributeRepository().Update(eNew); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *SpaceAttributeRouter) delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetSpaceAttributeRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	user := GetRequestUser(r)
	if !CanSpaceAdminOrg(user, e.OrganizationID) {
		SendForbidden(w)
		return
	}
	if err := GetSpaceAttributeRepository().Delete(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *SpaceAttributeRouter) create(w http.ResponseWriter, r *http.Request) {
	var m CreateSpaceAttributeRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	user := GetRequestUser(r)
	e := router.copyFromRestModel(&m)
	e.OrganizationID = user.OrganizationID
	if !CanSpaceAdminOrg(user, e.OrganizationID) {
		SendForbidden(w)
		return
	}
	if err := GetSpaceAttributeRepository().Create(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendCreated(w, e.ID)
}

func (router *SpaceAttributeRouter) copyFromRestModel(m *CreateSpaceAttributeRequest) *SpaceAttribute {
	e := &SpaceAttribute{}
	e.Label = m.Label
	e.Type = SettingType(m.Type)
	e.SpaceApplicable = m.SpaceApplicable
	e.LocationApplicable = m.LocationApplicable
	return e
}

func (router *SpaceAttributeRouter) copyToRestModel(e *SpaceAttribute) *GetSpaceAttributeResponse {
	m := &GetSpaceAttributeResponse{}
	m.ID = e.ID
	m.OrganizationID = e.OrganizationID
	m.Label = e.Label
	m.Type = int(e.Type)
	m.SpaceApplicable = e.SpaceApplicable
	m.LocationApplicable = e.LocationApplicable
	return m
}
