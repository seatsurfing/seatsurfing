package router

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/repository"
)

type GroupRouter struct {
}

type CreateGroupRequest struct {
	Name string `json:"name" validate:"required"`
}

type GetGroupResponse struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organizationId"`
	CreateGroupRequest
}

func (router *GroupRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/{id}/member/remove", router.removeMembers).Methods("POST")
	s.HandleFunc("/{id}/member", router.getMembers).Methods("GET")
	s.HandleFunc("/{id}/member", router.addMembers).Methods("PUT")
	s.HandleFunc("/{id}", router.getOne).Methods("GET")
	s.HandleFunc("/{id}", router.update).Methods("PUT")
	s.HandleFunc("/{id}", router.delete).Methods("DELETE")
	s.HandleFunc("/", router.create).Methods("POST")
	s.HandleFunc("/", router.getAll).Methods("GET")
}

func (router *GroupRouter) getOne(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetGroupRepository().GetOne(vars["id"])
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

func (router *GroupRouter) getAll(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CanAdminOrg(user, user.OrganizationID) {
		SendForbidden(w)
		return
	}
	list, err := GetGroupRepository().GetAll(user.OrganizationID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []*GetGroupResponse{}
	for _, e := range list {
		m := router.copyToRestModel(e)
		res = append(res, m)
	}
	SendJSON(w, res)
}

func (router *GroupRouter) update(w http.ResponseWriter, r *http.Request) {
	var m CreateGroupRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	vars := mux.Vars(r)
	e, err := GetGroupRepository().GetOne(vars["id"])
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
	if err := GetGroupRepository().Update(eNew); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *GroupRouter) delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetGroupRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	user := GetRequestUser(r)
	if !CanAdminOrg(user, e.OrganizationID) {
		SendForbidden(w)
		return
	}
	if err := GetGroupRepository().Delete(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *GroupRouter) create(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CanAdminOrg(user, user.OrganizationID) {
		SendForbidden(w)
		return
	}
	featureGroups, _ := GetSettingsRepository().GetBool(user.OrganizationID, SettingFeatureGroups.Name)
	if !featureGroups {
		SendPaymentRequired(w)
		return
	}
	var m CreateGroupRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	e := router.copyFromRestModel(&m)
	e.OrganizationID = user.OrganizationID
	if err := GetGroupRepository().Create(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendCreated(w, e.ID)
}

func (router *GroupRouter) addMembers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetGroupRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	user := GetRequestUser(r)
	if !CanAdminOrg(user, e.OrganizationID) {
		SendForbidden(w)
		return
	}
	var members []string
	if UnmarshalBody(r, &members) != nil {
		SendBadRequest(w)
		return
	}
	ok, err := GetUserRepository().UsersExistAndBelongToOrg(e.OrganizationID, members)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	if !ok {
		SendBadRequest(w)
		return
	}
	if err := GetGroupRepository().AddMembers(e, members); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *GroupRouter) removeMembers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetGroupRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	user := GetRequestUser(r)
	if !CanAdminOrg(user, e.OrganizationID) {
		SendForbidden(w)
		return
	}
	var members []string
	if UnmarshalBody(r, &members) != nil {
		SendBadRequest(w)
		return
	}
	if err := GetGroupRepository().RemoveMembers(e, members); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *GroupRouter) getMembers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetGroupRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	user := GetRequestUser(r)
	if !CanAdminOrg(user, e.OrganizationID) {
		SendForbidden(w)
		return
	}
	members, err := GetGroupRepository().GetMemberUserIDs(e)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	users, err := GetUserRepository().GetAllByIDs(members)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	ur := &UserRouter{}
	res := []*GetUserResponse{}
	for _, e := range users {
		m := ur.copyToRestModel(e, true)
		res = append(res, m)
	}
	SendJSON(w, res)
}

func (router *GroupRouter) copyFromRestModel(m *CreateGroupRequest) *Group {
	e := &Group{}
	e.Name = m.Name
	return e
}

func (router *GroupRouter) copyToRestModel(e *Group) *GetGroupResponse {
	m := &GetGroupResponse{}
	m.ID = e.ID
	m.OrganizationID = e.OrganizationID
	m.Name = e.Name
	return m
}
