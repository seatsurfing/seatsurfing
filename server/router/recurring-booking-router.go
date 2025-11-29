package router

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/util"
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

type CreateRecurringBookingResponse struct {
	Enter     time.Time `json:"enter"`
	Leave     time.Time `json:"leave"`
	Success   bool      `json:"success"`
	ErrorCode int       `json:"errorCode,omitempty"`
	ID        string    `json:"id"`
}

type GetRecurringBookingResponse struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`
	CreateRecurringBookingRequest
}

func (router *RecurringBookingRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/precheck", router.preBookingCreateCheck).Methods("POST")
	s.HandleFunc("/{id}/ical", router.getIcal).Methods("GET")
	s.HandleFunc("/{id}", router.getOne).Methods("GET")
	s.HandleFunc("/{id}", router.delete).Methods("DELETE")
	s.HandleFunc("/", router.create).Methods("POST")
}

func (router *RecurringBookingRouter) preBookingCreateCheck(w http.ResponseWriter, r *http.Request) {
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
	featureRecurringBookings, _ := GetSettingsRepository().GetBool(requestUser.OrganizationID, SettingFeatureRecurringBookings.Name)
	if !featureRecurringBookings {
		SendPaymentRequired(w)
		return
	}
	e, err := router.copyFromRestModel(&m, location)
	if err != nil {
		SendInternalServerError(w)
		return
	}
	bookingRouter := &BookingRouter{}
	bookings := GetRecurringBookingRepository().CreateBookings(e)
	res := make([]CreateRecurringBookingResponse, 0)
	for idx, b := range bookings {
		bookingReq := &CreateBookingRequest{
			SpaceID: b.SpaceID,
			BookingRequest: BookingRequest{
				Enter: b.Enter,
				Leave: b.Leave,
			},
		}
		valid, code := bookingRouter.checkBookingCreateUpdate(bookingReq, location, requestUser, "", idx)
		if valid {
			conflicts, _ := GetBookingRepository().GetConflicts(e.SpaceID, b.Enter, b.Leave, "")
			if len(conflicts) > 0 {
				valid = false
				code = ResponseCodeBookingSlotConflict
			}
		}
		item := CreateRecurringBookingResponse{
			Enter:     b.Enter,
			Leave:     b.Leave,
			Success:   valid,
			ErrorCode: code,
			ID:        b.ID,
		}
		res = append(res, item)
	}
	SendJSON(w, res)
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

func (router *RecurringBookingRouter) delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetRecurringBookingRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	space, err := GetSpaceRepository().GetOne(e.SpaceID)
	if err != nil {
		SendBadRequest(w)
		return
	}
	location, err := GetLocationRepository().GetOne(space.LocationID)
	if err != nil {
		SendBadRequest(w)
		return
	}
	if !CanAccessOrg(GetRequestUser(r), location.OrganizationID) {
		SendForbidden(w)
		return
	}
	if (e.UserID != GetRequestUserID(r)) && !CanSpaceAdminOrg(GetRequestUser(r), location.OrganizationID) {
		SendForbidden(w)
		return
	}
	if err := GetRecurringBookingRepository().Delete(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
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
	featureRecurringBookings, _ := GetSettingsRepository().GetBool(requestUser.OrganizationID, SettingFeatureRecurringBookings.Name)
	if !featureRecurringBookings {
		SendPaymentRequired(w)
		return
	}
	e, err := router.copyFromRestModel(&m, location)
	if err != nil {
		SendInternalServerError(w)
		return
	}
	e.UserID = GetRequestUserID(r)
	if err := GetRecurringBookingRepository().Create(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	bookingRouter := &BookingRouter{}
	spaceRequiresApproval := bookingRouter.getSpaceRequiresApproval(location.OrganizationID, space)
	bookings := GetRecurringBookingRepository().CreateBookings(e)
	res := make([]CreateRecurringBookingResponse, 0)
	for _, b := range bookings {
		bookingReq := &CreateBookingRequest{
			SpaceID: b.SpaceID,
			BookingRequest: BookingRequest{
				Enter: b.Enter,
				Leave: b.Leave,
			},
		}
		valid, code := bookingRouter.checkBookingCreateUpdate(bookingReq, location, requestUser, "", 0)
		if valid {
			conflicts, _ := GetBookingRepository().GetConflicts(e.SpaceID, b.Enter, b.Leave, "")
			if len(conflicts) > 0 {
				valid = false
				code = ResponseCodeBookingSlotConflict
			}
		}
		if valid {
			b.Approved = !spaceRequiresApproval
			if err := GetBookingRepository().Create(b); err != nil {
				log.Println(err)
				SendInternalServerError(w)
				return
			}
		}
		item := CreateRecurringBookingResponse{
			Enter:     b.Enter,
			Leave:     b.Leave,
			Success:   valid,
			ErrorCode: code,
			ID:        b.ID,
		}
		res = append(res, item)
	}
	go router.onBookingCreated(e, bookings, spaceRequiresApproval)
	w.Header().Set("X-Object-ID", e.ID)
	w.WriteHeader(http.StatusCreated)
	SendJSON(w, res)
}

func (router *RecurringBookingRouter) getIcal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetRecurringBookingRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
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
	bookings, err := GetBookingRepository().GetAllByRecurringID(e.ID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	if len(bookings) == 0 {
		SendNotFound(w)
		return
	}
	calDavEvents := []*CalDAVEvent{}
	bookingRouter := &BookingRouter{}
	for _, b := range bookings {
		calDavEvent, err := bookingRouter.getCalDavEventFromBooking(&b.Booking)
		if err != nil {
			log.Println(err)
			return
		}
		calDavEvents = append(calDavEvents, calDavEvent)
	}
	caldavClient := &CalDAVClient{}
	icalEvent := caldavClient.GetCaldavEvent(calDavEvents)
	var buf bytes.Buffer
	if err := ical.NewEncoder(&buf).Encode(icalEvent); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	w.Header().Set("Content-Type", "text/calendar")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+bookingRouter.getICalFilename(calDavEvents[0])+"\"")
	w.Write(buf.Bytes())
}

func (router *RecurringBookingRouter) onBookingCreated(e *RecurringBooking, bookings []*Booking, approvalRequired bool) {
	if len(bookings) == 0 {
		return
	}
	if !approvalRequired {
		br := &BookingRouter{}
		for _, b := range bookings {
			br.createCalDavEvent(b)
		}
		router.sendMailNotification(e, bookings)
	}
}

func (router *RecurringBookingRouter) sendMailNotification(e *RecurringBooking, bookings []*Booking) {
	active, err := GetUserPreferencesRepository().GetBool(e.UserID, PreferenceMailNotifications.Name)
	if err != nil || !active {
		return
	}
	user, err := GetUserRepository().GetOne(e.UserID)
	if err != nil || user == nil {
		log.Println(err)
		return
	}
	org, err := GetOrganizationRepository().GetOne(user.OrganizationID)
	if err != nil || org == nil {
		log.Println(err)
		return
	}
	space, err := GetSpaceRepository().GetOne(e.SpaceID)
	if err != nil {
		log.Println(err)
		return
	}
	location, err := GetLocationRepository().GetOne(space.LocationID)
	if err != nil {
		log.Println(err)
		return
	}
	attachments := []*MailAttachment{}
	calDavEvents := []*CalDAVEvent{}
	bookingRouter := &BookingRouter{}
	for _, b := range bookings {
		calDavEvent, err := bookingRouter.getCalDavEventFromBooking(b)
		if err != nil {
			log.Println(err)
			return
		}
		calDavEvents = append(calDavEvents, calDavEvent)
	}
	caldavClient := &CalDAVClient{}
	icalEvent := caldavClient.GetCaldavEvent(calDavEvents)
	var buf bytes.Buffer
	if err := ical.NewEncoder(&buf).Encode(icalEvent); err != nil {
		log.Println(err)
		return
	}
	attachments = append(attachments, &MailAttachment{
		Filename: bookingRouter.getICalFilename(calDavEvents[0]),
		MimeType: "text/calendar",
		Data:     buf.Bytes(),
	})

	subject := e.Subject
	if subject == "" {
		subject = "—"
	}
	vars := map[string]string{
		"recipientName": GetLocalPartFromEmailAddress(user.Email),
		"date":          e.Enter.Format("2006-01-02 15:04") + " - " + e.Leave.Format("2006-01-02 15:04"),
		"areaName":      location.Name,
		"spaceName":     space.Name,
		"subject":       subject,
	}
	template := GetEmailTemplatePathRecurringBookingCreated()
	if err := SendEmailWithAttachments(&MailAddress{Address: user.Email}, template, org.Language, vars, attachments); err != nil {
		log.Println(err)
		return
	}
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
