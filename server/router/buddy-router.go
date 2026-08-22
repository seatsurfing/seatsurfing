package router

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
)

type BuddyRouter struct {
}

type BuddyBooking struct {
	Enter time.Time `json:"enter"`
	Leave time.Time `json:"leave"`
	Desk  string    `json:"desk"`
	Room  string    `json:"room"`
}

type BuddyRequest struct {
	BuddyID           string        `json:"buddyId" validate:"required,uuid"`
	BuddyEmail        string        `json:"buddyEmail"`
	BuddyFirstBooking *BuddyBooking `json:"buddyFirstBooking"`
}

type CreateBuddyRequest struct {
	BuddyRequest
}

type GetBuddyResponse struct {
	ID string `json:"id"`
	CreateBuddyRequest
}

type SearchBuddyResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

func (router *BuddyRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/{id}", router.delete).Methods("DELETE")
	s.HandleFunc("/search", router.search).Methods("GET")
	s.HandleFunc("/", router.create).Methods("PUT")
	s.HandleFunc("/", router.getAll).Methods("GET")
}

func (router *BuddyRouter) isFeatureEnabled(orgID string) bool {
	showNames, _ := GetSettingsRepository().GetBool(orgID, SettingShowNames.Name)
	if !showNames {
		return false
	}
	disableBuddies, _ := GetSettingsRepository().GetBool(orgID, SettingDisableBuddies.Name)
	return !disableBuddies
}

func (router *BuddyRouter) getAll(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !router.isFeatureEnabled(user.OrganizationID) {
		SendForbidden(w)
		return
	}

	list, err := GetBuddyRepository().GetAllByOwner(user.ID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []*GetBuddyResponse{}
	for _, e := range list {
		m := router.copyToRestModel(e)
		res = append(res, m)
	}
	SendJSON(w, res)
}

func (router *BuddyRouter) search(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !router.isFeatureEnabled(user.OrganizationID) {
		SendForbidden(w)
		return
	}

	keyword := strings.TrimSpace(r.URL.Query().Get("q"))
	res := []*SearchBuddyResponse{}
	if len(keyword) < 3 || len(keyword) > 64 {
		SendJSON(w, res)
		return
	}

	list, err := GetUserRepository().GetByKeyword(user.OrganizationID, keyword)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	for _, e := range list {
		if e.ID == user.ID {
			continue
		}
		res = append(res, &SearchBuddyResponse{
			ID:        e.ID,
			Email:     e.Email,
			Firstname: e.Firstname,
			Lastname:  e.Lastname,
		})
	}
	SendJSON(w, res)
}

func (router *BuddyRouter) delete(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !router.isFeatureEnabled(user.OrganizationID) {
		SendForbidden(w)
		return
	}

	vars := mux.Vars(r)
	e, err := GetBuddyRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	if e.OwnerID != user.ID {
		SendForbidden(w)
		return
	}
	if err := GetBuddyRepository().Delete(e); err != nil {
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *BuddyRouter) create(w http.ResponseWriter, r *http.Request) {
	var m CreateBuddyRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}

	requestUser := GetRequestUser(r)
	if !router.isFeatureEnabled(requestUser.OrganizationID) {
		SendForbidden(w)
		return
	}

	buddyUser, err := GetUserRepository().GetOne(m.BuddyID)
	if err != nil {
		SendBadRequest(w)
		return
	}

	if buddyUser.OrganizationID != requestUser.OrganizationID {
		SendForbidden(w)
		return
	}

	e := &Buddy{}
	e.BuddyID = buddyUser.ID
	e.OwnerID = requestUser.ID
	if err := GetBuddyRepository().Create(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendCreated(w, e.ID)
}

func (router *BuddyRouter) copyToRestModel(e *BuddyDetails) *GetBuddyResponse {
	m := &GetBuddyResponse{}
	m.ID = e.ID
	m.BuddyID = e.BuddyID
	m.BuddyEmail = e.BuddyEmail
	// Assuming GetOne returns a pointer to BookingDetails
	bookingDetails, _ := GetBookingRepository().GetFirstUpcomingOrCurrentBookingByUserID(e.BuddyID)
	if bookingDetails == nil {
		m.BuddyFirstBooking = nil
		return m
	}
	// Use * to dereference the pointer
	actualBookingDetails := *bookingDetails

	m.BuddyFirstBooking = &BuddyBooking{
		Enter: actualBookingDetails.Enter,
		Leave: actualBookingDetails.Leave,
		Desk:  actualBookingDetails.Space.Name,
		Room:  actualBookingDetails.Space.Location.Name,
	}

	return m
}
