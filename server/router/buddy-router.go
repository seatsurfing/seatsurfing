package router

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

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
	BuddyID           string        `json:"buddyId" validate:"required"`
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

func (router *BuddyRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/{id}", router.delete).Methods("DELETE")
	s.HandleFunc("/", router.create).Methods("POST")
	s.HandleFunc("/", router.getAll).Methods("GET")
}

func (router *BuddyRouter) getAll(w http.ResponseWriter, r *http.Request) {
	list, err := GetBuddyRepository().GetAllByOwner(GetRequestUserID(r))
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

func (router *BuddyRouter) delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetBuddyRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	if e.OwnerID != GetRequestUserID(r) {
		SendForbidden(w)
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

	buddyUser, err := GetUserRepository().GetOne(m.BuddyID)
	if err != nil {
		SendBadRequest(w)
		return
	}

	e := &Buddy{}
	e.BuddyID = buddyUser.ID
	e.OwnerID = GetRequestUserID(r)
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
	bookingDetails, _ := GetBookingRepository().GetFirstUpcomingBookingByUserID(e.BuddyID)
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
