package router

import (
	"bytes"
	"errors"
	"log"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

type BookingRouter struct {
}

type BookingMailNotification int

const (
	BookingMailNotificationCreated BookingMailNotification = iota
	BookingMailNotificationDeclined
)

type BookingRequest struct {
	Enter     time.Time `json:"enter" validate:"required"`
	Leave     time.Time `json:"leave" validate:"required"`
	UserEmail string    `json:"userEmail"`
}

type CreateBookingRequest struct {
	SpaceID string `json:"spaceId" validate:"required"`
	Subject string `json:"subject"`
	BookingRequest
}

type PreCreateBookingRequest struct {
	LocationID string `json:"locationID" validate:"required"`
	BookingRequest
}

type GetBookingResponse struct {
	ID          string           `json:"id"`
	UserID      string           `json:"userId"`
	UserEmail   string           `json:"userEmail"`
	Approved    bool             `json:"approved"`
	Space       GetSpaceResponse `json:"space"`
	RecurringID string           `json:"recurringId"`
	CreateBookingRequest
}

type GetBookingFilterRequest struct {
	Start      time.Time `json:"start" validate:"required"`
	End        time.Time `json:"end" validate:"required"`
	LocationID string    `json:"locationId"`
}

type GetPresenceReportResult struct {
	Users     []GetUserInfoSmall `json:"users"`
	Dates     []string           `json:"dates"`
	Presences [][]int            `json:"presences"`
}

type DebugTimeIssuesRequest struct {
	Time time.Time `json:"time" validate:"required"`
}

type DebugTimeIssuesResponse struct {
	Timezone                string    `json:"tz"`
	Error                   string    `json:"error"`
	ReceivedTime            string    `json:"receivedTime"`
	ReceivedTimeTransformed string    `json:"receivedTimeTransformed"`
	Database                time.Time `json:"dbTime"`
	Result                  time.Time `json:"result"`
}

type GetPendingApprovalsCountResponse struct {
	Count int `json:"count"`
}

type SetBookingApprovalRequest struct {
	Approved bool `json:"approved"`
}

type CaldavConfig struct {
	URL      string
	Username string
	Password string
	Path     string
}

func (router *BookingRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/pendingapprovals/count", router.getPendingApprovalsCount).Methods("GET")
	s.HandleFunc("/pendingapprovals/", router.getPendingApprovals).Methods("GET")
	s.HandleFunc("/debugtimeissues/", router.debugTimeIssues).Methods("POST")
	s.HandleFunc("/report/presence/", router.getPresenceReport).Methods("GET")
	s.HandleFunc("/filter/", router.getFiltered).Methods("GET")
	s.HandleFunc("/precheck/", router.preBookingCreateCheck).Methods("POST")
	s.HandleFunc("/{id}/approve", router.approveBooking).Methods("POST")
	s.HandleFunc("/{id}/ical", router.getIcal).Methods("GET")
	s.HandleFunc("/{id}", router.getOne).Methods("GET")
	s.HandleFunc("/{id}", router.update).Methods("PUT")
	s.HandleFunc("/{id}", router.delete).Methods("DELETE")
	s.HandleFunc("/", router.create).Methods("POST")
	s.HandleFunc("/", router.getAll).Methods("GET")
}

func (router *BookingRouter) approveBooking(w http.ResponseWriter, r *http.Request) {
	requestUser := GetRequestUser(r)
	if !CanSpaceAdminOrg(requestUser, requestUser.OrganizationID) {
		SendForbidden(w)
		return
	}
	vars := mux.Vars(r)
	e, err := GetBookingRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}

	if !router.isValidApproverForSpace(requestUser.ID, e.SpaceID) {
		SendForbidden(w)
		return
	}
	m := &SetBookingApprovalRequest{}
	if UnmarshalBody(r, m) != nil {
		SendBadRequest(w)
		return
	}
	if e.Approved {
		SendUpdated(w)
		return
	}
	if !m.Approved {
		if err := GetBookingRepository().Delete(e); err != nil {
			log.Println(err)
			SendInternalServerError(w)
			return
		}
		go router.onBookingApproved(&e.Booking)
		SendUpdated(w)
		return
	}
	e.Approved = true
	if err := GetBookingRepository().Update(&e.Booking); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	go router.onBookingApproved(&e.Booking)
	SendUpdated(w)
}

func (router *BookingRouter) getPendingApprovalsCount(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CanSpaceAdminOrg(user, user.OrganizationID) {
		SendForbidden(w)
		return
	}
	count, err := GetBookingRepository().GetBookingsCountRequiringApproval(user.ID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := &GetPendingApprovalsCountResponse{
		Count: count,
	}
	SendJSON(w, res)
}

func (router *BookingRouter) getPendingApprovals(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CanSpaceAdminOrg(user, user.OrganizationID) {
		SendForbidden(w)
		return
	}
	list, err := GetBookingRepository().GetBookingsRequiringApproval(user.ID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []*GetBookingResponse{}
	for _, e := range list {
		m := router.copyToRestModel(e)
		res = append(res, m)
	}
	SendJSON(w, res)
}

func (router *BookingRouter) debugTimeIssues(w http.ResponseWriter, r *http.Request) {
	var m DebugTimeIssuesRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	tz := "America/Los_Angeles"
	res := &DebugTimeIssuesResponse{
		Timezone:     tz,
		ReceivedTime: m.Time.String(),
		Error:        "No error",
	}
	_, err := time.LoadLocation(tz)
	if err != nil {
		res.Error = "Could not load timezone: " + err.Error()
		SendJSON(w, res)
		return
	}
	timeNew, err := AttachTimezoneInformationTz(m.Time, tz)
	if err != nil {
		res.Error = "Could not attach timezone information (incoming): " + err.Error()
		SendJSON(w, res)
		return
	}
	res.ReceivedTimeTransformed = timeNew.String()
	e := &DebugTimeIssueItem{
		Created: timeNew,
	}
	if err := GetDebugTimeIssuesRepository().Create(e); err != nil {
		res.Error = "Could not create database record: " + err.Error()
		SendJSON(w, res)
		return
	}
	defer GetDebugTimeIssuesRepository().Delete(e)
	e2, err := GetDebugTimeIssuesRepository().GetOne(e.ID)
	if err != nil {
		res.Error = "Could not load database record: " + err.Error()
		SendJSON(w, res)
		return
	}
	res.Database = e2.Created
	timeToSend, err := AttachTimezoneInformationTz(e2.Created, tz)
	if err != nil {
		res.Error = "Could not attach timezone information (outgoing): " + err.Error()
		SendJSON(w, res)
		return
	}
	res.Result = timeToSend
	SendJSON(w, res)
}

func (router *BookingRouter) getFiltered(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CanSpaceAdminOrg(user, user.OrganizationID) {
		SendForbidden(w)
		return
	}
	start, err := time.Parse(time.RFC3339Nano, r.URL.Query().Get("start"))
	if err != nil {
		SendBadRequest(w)
		return
	}
	end, err := time.Parse(time.RFC3339Nano, r.URL.Query().Get("end"))
	if err != nil {
		SendBadRequest(w)
		return
	}
	list, err := GetBookingRepository().GetAllByOrg(user.OrganizationID, start, end)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []*GetBookingResponse{}
	for _, e := range list {
		m := router.copyToRestModel(e)
		res = append(res, m)
	}
	SendJSON(w, res)
}

func (router *BookingRouter) getIcal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetBookingRepository().GetOne(vars["id"])
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
	calDavEvent, err := router.getCalDavEventFromBooking(&e.Booking)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	caldavClient := &CalDAVClient{}
	icalEvent := caldavClient.GetCaldavEvent(calDavEvent)
	var buf bytes.Buffer
	if err := ical.NewEncoder(&buf).Encode(icalEvent); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	w.Header().Set("Content-Type", "text/calendar")
	w.Header().Set("Content-Disposition", "attachment; filename=\"seatsurfing.ics\"")
	w.Write(buf.Bytes())
}

func (router *BookingRouter) getOne(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetBookingRepository().GetOne(vars["id"])
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
	res := router.copyToRestModel(e)
	SendJSON(w, res)
}

func (router *BookingRouter) getAll(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now().UTC().Add(time.Hour * -12)
	list, err := GetBookingRepository().GetAllByUser(GetRequestUserID(r), startTime)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	user := GetRequestUser(r)
	defaultTz, err := GetSettingsRepository().Get(user.OrganizationID, SettingDefaultTimezone.Name)
	if err != nil {
		defaultTz = "UTC"
	}
	nowAtOrg, _ := GetUTCNowInTimezone(defaultTz)
	res := []*GetBookingResponse{}
	for _, e := range list {
		var nowAtLocation time.Time
		if e.Space.Location.Timezone == "" {
			nowAtLocation = nowAtOrg
		} else {
			nowAtLocation, _ = GetUTCNowInTimezone(e.Space.Location.Timezone)
		}
		includeEntity := e.Leave.After(nowAtLocation)
		if includeEntity {
			m := router.copyToRestModel(e)
			res = append(res, m)
		}
	}
	SendJSON(w, res)
}

func (router *BookingRouter) update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	e, err := GetBookingRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	var m CreateBookingRequest
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
	if e.UserID != requestUser.ID && !CanSpaceAdminOrg(requestUser, location.OrganizationID) {
		SendForbidden(w)
		return
	}
	eNew, err := router.copyFromRestModel(&m, location)
	if err != nil {
		SendInternalServerError(w)
		return
	}
	eNew.ID = e.ID
	eNew.CalDavID = e.CalDavID
	eNew.UserID = e.UserID
	if m.UserEmail != "" && m.UserEmail != requestUser.Email {
		if !CanSpaceAdminOrg(requestUser, location.OrganizationID) {
			SendForbidden(w)
			return
		}
		eNew.UserID, err = router.bookForUser(requestUser, m.UserEmail, w)
		if err != nil {
			SendInternalServerError(w)
			return
		}
	}
	bookingReq := &CreateBookingRequest{
		SpaceID: m.SpaceID,
		BookingRequest: BookingRequest{
			Enter: eNew.Enter,
			Leave: eNew.Leave,
		},
	}

	if valid, code := router.checkBookingCreateUpdate(bookingReq, location, requestUser, eNew.ID, 0); !valid {
		SendBadRequestCode(w, code)
		return
	}
	conflicts, err := GetBookingRepository().GetConflicts(eNew.SpaceID, eNew.Enter, eNew.Leave, eNew.ID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	if len(conflicts) > 0 {
		SendAleadyExists(w)
		return
	}
	if err := GetBookingRepository().Update(eNew); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	go router.onBookingUpdated(eNew)
	SendUpdated(w)
}

func (router *BookingRouter) delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetBookingRepository().GetOne(vars["id"])
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
	requestUser := GetRequestUser(r)
	// Check for the date, If the BookingRequest is to close with SettingsMaxHoursBeforeDelete, the Delete can not be performed.
	if router.isValidBookingHoursBeforeDelete(e, requestUser, location.OrganizationID) {
		go router.onBookingDeleted(&e.Booking)
		if err := GetBookingRepository().Delete(e); err != nil {
			SendInternalServerError(w)
			return
		}
		SendUpdated(w)
		return
	}
	SendForbiddenCode(w, ResponseCodeBookingMaxHoursBeforeDelete)
}

func (router *BookingRouter) checkBookingCreateUpdate(m *CreateBookingRequest, location *Location, requestUser *User, bookingID string, upcomingBookingsMarkup int) (bool, int) {
	if valid, code := router.isValidBookingRequest(m, requestUser, location.OrganizationID, bookingID, upcomingBookingsMarkup); !valid {
		return false, code
	}
	if !router.isValidConcurrent(m, location, bookingID) {
		return false, ResponseCodeBookingLocationMaxConcurrent
	}
	return true, 0
}

func (router *BookingRouter) preBookingCreateCheck(w http.ResponseWriter, r *http.Request) {
	var m PreCreateBookingRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	location, err := GetLocationRepository().GetOne(m.LocationID)
	if err != nil {
		SendBadRequest(w)
		return
	}
	requestUser := GetRequestUser(r)
	if !CanAccessOrg(requestUser, location.OrganizationID) {
		SendForbidden(w)
		return
	}
	enterNew, err := GetLocationRepository().AttachTimezoneInformation(m.Enter, location)
	if err != nil {
		SendInternalServerError(w)
		return
	}
	leaveNew, err := GetLocationRepository().AttachTimezoneInformation(m.Leave, location)
	if err != nil {
		SendInternalServerError(w)
		return
	}
	bookingReq := &CreateBookingRequest{
		SpaceID: "",
		BookingRequest: BookingRequest{
			Enter: enterNew,
			Leave: leaveNew,
		},
	}
	if valid, code := router.checkBookingCreateUpdate(bookingReq, location, requestUser, "", 0); !valid {
		SendBadRequestCode(w, code)
		return
	}
	SendUpdated(w)
}

func (router *BookingRouter) create(w http.ResponseWriter, r *http.Request) {
	var m CreateBookingRequest
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
	if m.UserEmail != "" && m.UserEmail != requestUser.Email {
		if !CanSpaceAdminOrg(requestUser, location.OrganizationID) {
			SendForbidden(w)
			return
		}
		e.UserID, err = router.bookForUser(requestUser, m.UserEmail, w)
		if err != nil {
			SendInternalServerError(w)
			return
		}
	}
	bookingReq := &CreateBookingRequest{
		SpaceID: m.SpaceID,
		BookingRequest: BookingRequest{
			Enter: e.Enter,
			Leave: e.Leave,
		},
	}

	if valid, code := router.checkBookingCreateUpdate(bookingReq, location, requestUser, "", 0); !valid {
		SendBadRequestCode(w, code)
		return
	}
	conflicts, err := GetBookingRepository().GetConflicts(e.SpaceID, e.Enter, e.Leave, "")
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	if len(conflicts) > 0 {
		SendAleadyExists(w)
		return
	}
	e.Approved = !router.getSpaceRequiresApproval(location.OrganizationID, space)
	if err := GetBookingRepository().Create(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	go router.onBookingCreated(e)
	SendCreated(w, e.ID)
}

func (router *BookingRouter) bookForUser(requestUser *User, userEmail string, w http.ResponseWriter) (string, error) {
	if !CanSpaceAdminOrg(requestUser, requestUser.OrganizationID) {
		SendForbidden(w)
		return "", errors.New("Forbidden")
	}
	bookForUser, err := GetUserRepository().GetByEmail(requestUser.OrganizationID, userEmail)
	if bookForUser == nil || err != nil {
		org, err := GetOrganizationRepository().GetOne(requestUser.OrganizationID)
		if err != nil || org == nil {
			SendInternalServerError(w)
			return "", errors.New("InternalServerError")
		}
		if allowed, _ := GetSettingsRepository().GetBool(org.ID, SettingAllowBookingsNonExistingUsers.Name); !allowed {
			SendForbidden(w)
			return "", errors.New("Forbidden")
		}
		if !GetUserRepository().CanCreateUser(org) {
			SendInternalServerError(w)
			return "", errors.New("InternalServerError")
		}
		if !GetOrganizationRepository().IsValidCustomDomainForOrg(userEmail, org) {
			SendBadRequest(w)
			return "", errors.New("BadRequest")
		}
		user := &User{
			Email:          userEmail,
			AtlassianID:    NullString(""),
			OrganizationID: org.ID,
			Role:           UserRoleUser,
		}
		err = GetUserRepository().Create(user)
		if err != nil {
			SendInternalServerError(w)
			return "", errors.New("InternalServerError")
		}
		bookForUser, err = GetUserRepository().GetByEmail(org.ID, userEmail)
		if err != nil {
			SendInternalServerError(w)
			return "", errors.New("InternalServerError")
		}
	}

	if bookForUser == nil {
		SendNotFound(w)
		return "", errors.New("NotFound")
	}
	return bookForUser.ID, nil
}

func (router *BookingRouter) getPresenceReport(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CanSpaceAdminOrg(user, user.OrganizationID) {
		SendForbidden(w)
		return
	}
	start, err := time.Parse(time.RFC3339Nano, r.URL.Query().Get("start"))
	if err != nil {
		SendBadRequest(w)
		return
	}
	end, err := time.Parse(time.RFC3339Nano, r.URL.Query().Get("end"))
	if err != nil {
		SendBadRequest(w)
		return
	}
	locationID := r.URL.Query().Get("locationId")
	var location *Location = nil
	if locationID != "" {
		location, _ = GetLocationRepository().GetOne(locationID)
		if location == nil {
			SendNotFound(w)
			return
		}
		if !GetUserRepository().IsSuperAdmin(user) && location.OrganizationID != user.OrganizationID {
			SendForbidden(w)
			return
		}
	}
	items, err := GetBookingRepository().GetPresenceReport(user.OrganizationID, location, start, end, 1000, 0)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	numUsers := len(items)
	numDates := 0
	if numUsers > 0 {
		numDates = len(items[0].Presence)
	}
	res := &GetPresenceReportResult{
		Users:     make([]GetUserInfoSmall, numUsers),
		Dates:     make([]string, numDates),
		Presences: make([][]int, numUsers),
	}
	i := 0
	for date := range items[0].Presence {
		res.Dates[i] = date
		i++
	}
	sort.Strings(res.Dates)
	for i, item := range items {
		res.Users[i] = GetUserInfoSmall{
			UserID: item.User.ID,
			Email:  item.User.Email,
		}
		res.Presences[i] = make([]int, numDates)
		for j, date := range res.Dates {
			res.Presences[i][j] = item.Presence[date]
		}
	}
	SendJSON(w, res)
}

func (router *BookingRouter) IsValidBookingDuration(m *BookingRequest, orgID string, user *User) bool {
	noAdminRestrictions, _ := GetSettingsRepository().GetBool(orgID, SettingNoAdminRestrictions.Name)
	if noAdminRestrictions && CanSpaceAdminOrg(user, orgID) {
		return true
	}
	dailyBasisBooking, _ := GetSettingsRepository().GetBool(orgID, SettingDailyBasisBooking.Name)
	maxDurationHours, _ := GetSettingsRepository().GetInt(orgID, SettingMaxBookingDurationHours.Name)
	if dailyBasisBooking && (maxDurationHours%24 != 0) {
		maxDurationHours += (24 - (maxDurationHours % 24))
	}

	// Due to daylight saving time, days can have more or less than 24 hours
	if dailyBasisBooking {
		correction := 0
		now := m.Enter
		for now.Before(m.Leave) {
			hoursOnDate := router.getHoursOnDate(&now)
			now = now.AddDate(0, 0, 1)
			correction += (hoursOnDate - 24)
		}
		durationNotRounded := int(math.Round(m.Leave.Sub(m.Enter).Minutes()) / 60)
		return ((durationNotRounded-correction)%24 == 0) && (durationNotRounded <= (maxDurationHours + correction))
	}

	// For non-daily-basis bookings, check exact duration
	duration := math.Floor(m.Leave.Sub(m.Enter).Minutes()) / 60
	if duration < 0 || duration > float64(maxDurationHours) {
		return false
	}
	return true
}

func (router *BookingRouter) getHoursOnDate(t *time.Time) int {
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	end := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
	durationNotRounded := int(math.Round(end.Sub(start).Minutes()) / 60)
	return durationNotRounded
}

func (router *BookingRouter) IsValidBookingAdvance(m *BookingRequest, orgID string, user *User) bool {
	noAdminRestrictions, _ := GetSettingsRepository().GetBool(orgID, SettingNoAdminRestrictions.Name)
	maxAdvanceDays, _ := GetSettingsRepository().GetInt(orgID, SettingMaxDaysInAdvance.Name)
	dailyBasisBooking, _ := GetSettingsRepository().GetBool(orgID, SettingDailyBasisBooking.Name)
	// allow Enter-Date in past if at least this morning
	now := time.Now().UTC()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if dailyBasisBooking {
		now = now.Add(-12 * time.Hour)
	}
	if m.Leave.Before(now) { // Leave must not be in past
		return false
	}
	advanceDays := math.Floor(m.Enter.Sub(now).Hours() / 24)
	if advanceDays >= 0 && noAdminRestrictions && CanSpaceAdminOrg(user, orgID) {
		return true
	}
	if advanceDays < 0 || advanceDays > float64(maxAdvanceDays) {
		return false
	}
	return true
}

func (router *BookingRouter) IsValidMaxUpcomingBookings(orgID string, user *User, upcomingBookingsMarkup int) bool {
	noAdminRestrictions, _ := GetSettingsRepository().GetBool(orgID, SettingNoAdminRestrictions.Name)
	if noAdminRestrictions && CanSpaceAdminOrg(user, orgID) {
		return true
	}
	maxUpcoming, _ := GetSettingsRepository().GetInt(orgID, SettingMaxBookingsPerUser.Name)
	curUpcoming, _ := GetBookingRepository().GetAllByUser(user.ID, time.Now().UTC())
	return len(curUpcoming)+upcomingBookingsMarkup < maxUpcoming
}

func (router *BookingRouter) isValidMaxConcurrentBookingsForUser(orgID string, user *User, m *BookingRequest, bookingID string) bool {
	noAdminRestrictions, _ := GetSettingsRepository().GetBool(orgID, SettingNoAdminRestrictions.Name)
	if noAdminRestrictions && CanSpaceAdminOrg(user, orgID) {
		return true
	}
	maxConcurrent, _ := GetSettingsRepository().GetInt(orgID, SettingMaxConcurrentBookingsPerUser.Name)
	// 0 = no limit
	if maxConcurrent == 0 {
		return true
	}
	curAtTime, _ := GetBookingRepository().GetTimeRangeByUser(user.ID, m.Enter, m.Leave, bookingID)
	return len(curAtTime) < maxConcurrent
}

func (router *BookingRouter) isValidBookingRequest(m *CreateBookingRequest, user *User, orgID string, bookingID string, upcomingBookingsMarkup int) (bool, int) {
	isUpdate := bookingID != ""
	if !router.IsValidBookingDuration(&m.BookingRequest, orgID, user) {
		return false, ResponseCodeBookingInvalidBookingDuration
	}
	if !router.IsValidBookingAdvance(&m.BookingRequest, orgID, user) {
		return false, ResponseCodeBookingTooManyDaysInAdvance
	}
	if !router.isValidMaxConcurrentBookingsForUser(orgID, user, &m.BookingRequest, bookingID) {
		return false, ResponseCodeBookingMaxConcurrentForUser
	}
	if !router.isValidMinHoursBooking(&m.BookingRequest, orgID, user) {
		return false, ResponseCodeBookingInvalidMinBookingDuration
	}
	if !isUpdate {
		if !router.IsValidMaxUpcomingBookings(orgID, user, upcomingBookingsMarkup) {
			return false, ResponseCodeBookingTooManyUpcomingBookings
		}
	}
	if m.SpaceID == "" {
		return true, 0
	}
	groupMemberships, _ := GetGroupRepository().GetAllWhereUserIsMember(user.ID)
	allowedBookers, _ := GetSpaceRepository().GetAllAllowedBookersForSpaceList([]string{m.SpaceID})
	if len(allowedBookers) > 0 {
		allowed := false
		for _, allowedBooker := range allowedBookers {
			for _, group := range groupMemberships {
				if group.ID == allowedBooker.GroupID {
					allowed = true
					break
				}
			}
		}
		if !allowed {
			return false, ResponseCodeBookingNotAllowedBooker
		}
	}
	return true, 0
}

func (router *BookingRouter) isValidConcurrent(m *CreateBookingRequest, location *Location, bookingID string) bool {
	if location.MaxConcurrentBookings == 0 {
		return true
	}
	bookings, err := GetBookingRepository().GetConcurrent(location, m.Enter, m.Leave, bookingID)
	if err != nil {
		log.Println(err)
		return false
	}
	if bookings >= int(location.MaxConcurrentBookings) {
		return false
	}
	return true
}

func (router *BookingRouter) isValidBookingHoursBeforeDelete(e *BookingDetails, user *User, organizationID string) bool {
	noAdminRestrictions, _ := GetSettingsRepository().GetBool(organizationID, SettingNoAdminRestrictions.Name)
	if noAdminRestrictions && CanSpaceAdminOrg(user, organizationID) {
		return true
	}
	enable_check, err := GetSettingsRepository().GetBool(organizationID, SettingEnableMaxHourBeforeDelete.Name)
	if err != nil {
		log.Println(err)
		return false
	}
	if !enable_check {
		return true
	}
	max_hours, err := GetSettingsRepository().GetInt(organizationID, SettingMaxHoursBeforeDelete.Name)
	if err != nil {
		log.Println(err)
		return false
	}
	enterTime := e.Enter
	now := time.Now().UTC()
	difference_in_hours := int64(enterTime.Sub(now).Hours())
	return difference_in_hours > int64(max_hours) || (max_hours == 0)
}

func (router *BookingRouter) isValidMinHoursBooking(e *BookingRequest, organizationID string, user *User) bool {
	noAdminRestrictions, _ := GetSettingsRepository().GetBool(organizationID, SettingNoAdminRestrictions.Name)
	if noAdminRestrictions && CanSpaceAdminOrg(user, organizationID) {
		return true
	}
	min_hours, err := GetSettingsRepository().GetInt(organizationID, SettingMinBookingDurationHours.Name)
	if err != nil {
		log.Println(err)
		return false
	}
	enterTime := e.Enter
	leaveTime := e.Leave
	difference_in_hours := int64(leaveTime.Sub(enterTime).Hours())
	return difference_in_hours >= int64(min_hours)
}

func (router *BookingRouter) getCalDavConfig(userID string) (*CaldavConfig, error) {
	prefs, err := GetUserPreferencesRepository().GetAll(userID)
	if err != nil {
		return nil, err
	}
	res := &CaldavConfig{}
	for _, pref := range prefs {
		if pref.Name == PreferenceCalDAVURL.Name {
			if _, err := url.ParseRequestURI(pref.Value); err == nil {
				res.URL = pref.Value
			}
		} else if pref.Name == PreferenceCalDAVUser.Name && len(pref.Value) > 0 {
			res.Username = pref.Value
		} else if pref.Name == PreferenceCalDAVPass.Name && len(pref.Value) > 0 {
			decryptedPassword := DecryptString(pref.Value)
			if decryptedPassword != "" {
				res.Password = decryptedPassword
			}
		} else if pref.Name == PreferenceCalDAVPath.Name && len(pref.Value) > 0 {
			res.Path = pref.Value
		}
	}
	if res.URL == "" || res.Username == "" || res.Password == "" || res.Path == "" {
		return nil, errors.New("caldav not configured completely")
	}
	return res, nil
}

func (router *BookingRouter) initCaldavEvent(e *Booking) (*CalDAVClient, *CalDAVEvent, string, error) {
	config, err := router.getCalDavConfig(e.UserID)
	if err != nil {
		return nil, nil, "", err
	}
	caldavClient := &CalDAVClient{}
	if err := caldavClient.Connect(config.URL, config.Username, config.Password); err != nil {
		log.Println(err)
		return nil, nil, "", err
	}
	caldavEvent, err := router.getCalDavEventFromBooking(e)
	if err != nil {
		return nil, nil, "", err
	}
	return caldavClient, caldavEvent, config.Path, nil
}

func (router *BookingRouter) getCalDavEventFromBooking(e *Booking) (*CalDAVEvent, error) {
	space, err := GetSpaceRepository().GetOne(e.SpaceID)
	if err != nil {
		return nil, err
	}
	location, err := GetLocationRepository().GetOne(space.LocationID)
	if err != nil {
		return nil, err
	}
	caldavEvent := &CalDAVEvent{
		Title:    "Seat Reservation: " + space.Name + ", " + location.Name,
		Location: space.Name + ", " + location.Name,
		Start:    e.Enter,
		End:      e.Leave,
	}
	return caldavEvent, nil
}

func (router *BookingRouter) createCalDavEvent(e *Booking) {
	caldavClient, caldavEvent, path, err := router.initCaldavEvent(e)
	if err != nil {
		return
	}
	if err := caldavClient.CreateEvent(path, caldavEvent); err != nil {
		log.Println(err)
		return
	}
	e.CalDavID = caldavEvent.ID
	GetBookingRepository().Update(e)
}

func (router *BookingRouter) updateCalDavEvent(e *Booking) {
	caldavClient, caldavEvent, path, err := router.initCaldavEvent(e)
	if err != nil {
		return
	}
	if e.CalDavID != "" {
		caldavEvent.ID = e.CalDavID
	}
	if err := caldavClient.CreateEvent(path, caldavEvent); err != nil {
		log.Println(err)
		return
	}
	e.CalDavID = caldavEvent.ID
	GetBookingRepository().Update(e)
}

func (router *BookingRouter) isValidApproverForSpace(userID, spaceID string) bool {
	approverGroups, err := GetSpaceRepository().GetApproverGroupIDs(spaceID)
	if err != nil {
		log.Println(err)
		return false
	}
	if len(approverGroups) == 0 {
		return true
	}
	userGroups, err := GetGroupRepository().GetAllWhereUserIsMember(userID)
	if err != nil {
		log.Println(err)
		return false
	}
	for _, group := range userGroups {
		for _, approverGroup := range approverGroups {
			if group.ID == approverGroup {
				return true
			}
		}
	}
	return false
}

func (router *BookingRouter) getSpaceRequiresApproval(orgID string, e *Space) bool {
	groupsEnabled, _ := GetSettingsRepository().GetBool(orgID, SettingFeatureGroups.Name)
	if !groupsEnabled {
		return false
	}
	approvers, _ := GetSpaceRepository().GetApproverGroupIDs(e.ID)
	return len(approvers) > 0
}

func (router *BookingRouter) sendMailNotification(e *Booking, notification BookingMailNotification) {
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
	if notification == BookingMailNotificationCreated {
		calDavEvent, err := router.getCalDavEventFromBooking(e)
		if err != nil {
			log.Println(err)
			return
		}
		caldavClient := &CalDAVClient{}
		icalEvent := caldavClient.GetCaldavEvent(calDavEvent)
		var buf bytes.Buffer
		if err := ical.NewEncoder(&buf).Encode(icalEvent); err != nil {
			log.Println(err)
			return
		}
		attachments = append(attachments, &MailAttachment{
			Filename: "seatsurfing.ics",
			MimeType: "text/calendar",
			Data:     buf.Bytes(),
		})
	}
	vars := map[string]string{
		"recipientName": GetLocalPartFromEmailAddress(user.Email),
		"date":          e.Enter.Format("2006-01-02 15:04") + " - " + e.Leave.Format("2006-01-02 15:04"),
		"areaName":      location.Name,
		"spaceName":     space.Name,
	}
	template := GetEmailTemplatePathBookingCreated()
	if notification == BookingMailNotificationDeclined {
		template = GetEmailTemplatePathBookingDeclined()
	}
	if err := SendEmailWithAttachments(&MailAddress{Address: user.Email}, template, org.Language, vars, attachments); err != nil {
		log.Println(err)
		return
	}
}

func (router *BookingRouter) onBookingUpdated(e *Booking) {
	router.updateCalDavEvent(e)
	router.sendMailNotification(e, BookingMailNotificationCreated)
}

func (router *BookingRouter) onBookingApproved(e *Booking) {
	if !e.Approved {
		router.onBookingDeleted(e)
		router.sendMailNotification(e, BookingMailNotificationDeclined)
		return
	}
	router.onBookingCreated(e)
}

func (router *BookingRouter) onBookingCreated(e *Booking) {
	if e.Approved {
		router.createCalDavEvent(e)
		router.sendMailNotification(e, BookingMailNotificationCreated)
	}
}

func (router *BookingRouter) onBookingDeleted(e *Booking) {
	caldavClient, caldavEvent, path, err := router.initCaldavEvent(e)
	if err != nil {
		return
	}
	if e.CalDavID == "" {
		return
	}
	caldavEvent.ID = e.CalDavID
	if err := caldavClient.DeleteEvent(path, caldavEvent); err != nil {
		log.Println(err)
		return
	}
}

func (router *BookingRouter) copyFromRestModel(m *CreateBookingRequest, location *Location) (*Booking, error) {
	e := &Booking{}
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
	return e, nil
}

func (router *BookingRouter) copyToRestModel(e *BookingDetails) *GetBookingResponse {
	m := &GetBookingResponse{}
	m.ID = e.ID
	m.UserID = e.UserID
	m.UserEmail = e.UserEmail
	m.SpaceID = e.SpaceID
	m.Subject = e.Subject
	m.Enter, _ = GetLocationRepository().AttachTimezoneInformation(e.Enter, &e.Space.Location)
	m.Leave, _ = GetLocationRepository().AttachTimezoneInformation(e.Leave, &e.Space.Location)
	m.Space.ID = e.Space.ID
	m.RecurringID = string(e.RecurringID)
	m.Approved = e.Approved
	m.Space.LocationID = e.Space.LocationID
	m.Space.Name = e.Space.Name
	m.Space.Location = &GetLocationResponse{
		ID: e.Space.Location.ID,
		CreateLocationRequest: CreateLocationRequest{
			Name: e.Space.Location.Name,
		},
	}
	return m
}
