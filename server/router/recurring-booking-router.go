package router

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/repository"
)

type RecurringBookingRouter struct {
}

type CreateRecurringBookingRequest struct {
	SpaceID  string         `json:"spaceId" validate:"required"`
	Subject  string         `json:"subject"`
	Enter    time.Time      `json:"enter" validate:"required"`
	Leave    time.Time      `json:"leave" validate:"required"`
	End      time.Time      `json:"end" validate:"required"`
	Cadence  Cadence        `json:"cadence" validate:"required,min=1,max=2"`
	Cycle    int            `json:"cycle" validate:"required,min=1"`
	Weekdays []time.Weekday `json:"weekdays"`
}

type GetRecurringBookingResponse struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`
	CreateRecurringBookingRequest
}

func (router *RecurringBookingRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/{id}", router.getOne).Methods("GET")
	s.HandleFunc("/", router.create).Methods("POST")
}

func (router *RecurringBookingRouter) getOne(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	e, err := GetRecurringBookingRepository().GetOne(id)
	if err != nil || e == nil {
		SendNotFound(w)
		return
	}
	requestUser := GetRequestUser(r)
	if !CanAccessOrg(requestUser, requestUser.OrganizationID) && e.UserID != GetRequestUserID(r) {
		SendForbidden(w)
		return
	}
	if e.UserID != GetRequestUserID(r) && !CanSpaceAdminOrg(requestUser, requestUser.OrganizationID) {
		SendForbidden(w)
		return
	}
	space, err := GetSpaceRepository().GetOne(e.SpaceID)
	if err != nil || space == nil {
		SendNotFound(w)
		return
	}
	location, err := GetLocationRepository().GetOne(space.LocationID)
	if err != nil || location == nil {
		SendNotFound(w)
		return
	}
	res := router.copyToRestModel(e, location)
	SendJSON(w, res)
}

func (router *RecurringBookingRouter) create(w http.ResponseWriter, r *http.Request) {
	var m CreateRecurringBookingRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	space, err := GetSpaceRepository().GetOne(m.SpaceID)
	if err != nil {
		SendBadRequest(w)
		return
	}
	if space.RequireSubject && len(strings.TrimSpace(m.Subject)) < 3 {
		SendBadRequestCode(w, ResponseCodeBookingSubjectRequired)
		return
	}
	location, err := GetLocationRepository().GetOne(space.LocationID)
	if err != nil {
		SendBadRequest(w)
		return
	}
	requestUser := GetRequestUser(r)
	if !CanAccessOrg(requestUser, location.OrganizationID) {
		SendForbidden(w)
		return
	}
	e, err := router.copyFromRestModel(&m, location)
	if err != nil {
		SendInternalServerError(w)
		return
	}
	e.UserID = GetRequestUserID(r)
	// TODO Validate basic parameter (such as in advance, ...)
	// ...
	if err := GetRecurringBookingRepository().Create(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	// TODO create bookings
	// ..
	SendCreated(w, e.ID)
}

func (router *RecurringBookingRouter) copyFromRestModel(m *CreateRecurringBookingRequest, location *Location) (*RecurringBooking, error) {
	e := &RecurringBooking{}
	e.SpaceID = m.SpaceID
	e.Subject = m.Subject
	e.Enter = m.Enter
	e.Leave = m.Leave
	enterNew, err := GetLocationRepository().AttachTimezoneInformation(e.Enter, location)
	if err != nil {
		return nil, err
	}
	e.Enter = enterNew
	leaveNew, err := GetLocationRepository().AttachTimezoneInformation(e.Leave, location)
	if err != nil {
		return nil, err
	}
	e.Leave = leaveNew
	endNew, err := GetLocationRepository().AttachTimezoneInformation(m.End, location)
	if err != nil {
		return nil, err
	}
	e.End = endNew
	e.Cadence = m.Cadence
	if m.Cadence == CadenceDaily {
		e.Details = &CadenceDailyDetails{
			Cycle: m.Cycle,
		}
	} else if m.Cadence == CadenceWeekly {
		e.Details = &CadenceWeeklyDetails{
			Cycle:    m.Cycle,
			Weekdays: m.Weekdays,
		}
	} else {
		return nil, errors.New("invalid cadence")
	}
	return e, nil
}

func (router *RecurringBookingRouter) copyToRestModel(e *RecurringBooking, location *Location) *GetRecurringBookingResponse {
	m := &GetRecurringBookingResponse{}
	m.ID = e.ID
	m.UserID = e.UserID
	m.SpaceID = e.SpaceID
	m.Subject = e.Subject
	m.Enter, _ = GetLocationRepository().AttachTimezoneInformation(e.Enter, location)
	m.Leave, _ = GetLocationRepository().AttachTimezoneInformation(e.Leave, location)
	m.End, _ = GetLocationRepository().AttachTimezoneInformation(e.End, location)
	m.Cadence = e.Cadence
	if e.Cadence == CadenceDaily {
		if details, ok := e.Details.(*CadenceDailyDetails); ok {
			m.Cycle = details.Cycle
		}
	} else if e.Cadence == CadenceWeekly {
		if details, ok := e.Details.(*CadenceWeeklyDetails); ok {
			m.Cycle = details.Cycle
			m.Weekdays = details.Weekdays
		}
	}
	return m
}
