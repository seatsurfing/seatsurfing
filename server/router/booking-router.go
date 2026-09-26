package router

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/emersion/go-ical"
	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	"github.com/seatsurfing/seatsurfing/server/service"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

type BookingRouter struct {
}

type BookingMailNotification int

const (
	BookingMailNotificationCreated BookingMailNotification = iota
	BookingMailNotificationDeclined
	BookingMailNotificationUpdated
	BookingMailNotificationApproved
	BookingMailNotificationDeleted
)

type BookingRequest struct {
	Enter     time.Time `json:"enter" validate:"required"`
	Leave     time.Time `json:"leave" validate:"required"`
	UserEmail string    `json:"userEmail"`
}

type CreateBookingRequest struct {
	SpaceID string `json:"spaceId" validate:"required"`
	Subject string `json:"subject" validate:"omitempty,max=256"`
	BookingRequest
}

type PreCreateBookingRequest struct {
	LocationID string `json:"locationId" validate:"required,uuid"`
	BookingRequest
}

type GetBookingResponse struct {
	ID            string           `json:"id"`
	UserID        string           `json:"userId"`
	UserEmail     string           `json:"userEmail"`
	UserFirstname string           `json:"userFirstname"`
	UserLastname  string           `json:"userLastname"`
	Public        bool             `json:"public"`
	Approved      bool             `json:"approved"`
	Space         GetSpaceResponse `json:"space"`
	RecurringID   string           `json:"recurringId"`
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
	s.HandleFunc("/report/presence/", router.getPresenceReport).Methods("GET")
	s.HandleFunc("/filter/", router.getFiltered).Methods("GET")
	s.HandleFunc("/current/", router.getCurrent).Methods("GET")
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
	vars := mux.Vars(r)
	e, err := GetBookingRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
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
	if !HasPermission(requestUser, location.OrganizationID, PermissionApprovals, PermissionLevelAdmin) {
		SendForbidden(w)
		return
	}

	if !router.isValidApproverForSpace(requestUser.ID, e.SpaceID) {
		SendForbidden(w)
		return
	}
	if e.PublicID != "" {
		publicBookingEnabled, _ := GetSettingsRepository().GetBool(e.Space.Location.OrganizationID, SettingPublicBookingEnabled.Name)
		if !publicBookingEnabled {
			SendForbidden(w)
			return
		}
	}
	m := &SetBookingApprovalRequest{}
	if UnmarshalBody(r, m) != nil {
		SendBadRequest(w)
		return
	}

	if e.Leave.Before(time.Now().Add(-24 * time.Hour)) {
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
		go router.onBookingDeclinedOrApproved(&e.Booking)
		SendUpdated(w)
		return
	}
	e.Approved = true
	if err := GetBookingRepository().Update(&e.Booking); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	go router.onBookingDeclinedOrApproved(&e.Booking)
	SendUpdated(w)
}

func (router *BookingRouter) getPendingApprovalsCount(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !HasPermission(user, user.OrganizationID, PermissionApprovals, PermissionLevelAdmin) {
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
	if !HasPermission(user, user.OrganizationID, PermissionApprovals, PermissionLevelAdmin) {
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

func (router *BookingRouter) validateBookingFilters(w http.ResponseWriter, r *http.Request) (*User, string, string, bool) {
	user := GetRequestUser(r)
	if !HasPermission(user, user.OrganizationID, PermissionBookings, PermissionLevelRead) {
		SendForbidden(w)
		return nil, "", "", false
	}

	filterUserEmail := r.URL.Query().Get("user")
	if filterUserEmail != "" {
		filterUser, _ := GetUserRepository().GetByEmail(user.OrganizationID, filterUserEmail)
		if filterUser == nil {
			SendBadRequest(w)
			return nil, "", "", false
		}
	}

	filterLocationId := r.URL.Query().Get("location")
	if filterLocationId != "" {
		if !ValidateGUID(filterLocationId) {
			SendBadRequest(w)
			return nil, "", "", false
		}
		filterLocation, _ := GetLocationRepository().GetOne(filterLocationId)
		if filterLocation == nil || filterLocation.OrganizationID != user.OrganizationID {
			SendBadRequest(w)
			return nil, "", "", false
		}
	}

	return user, filterUserEmail, filterLocationId, true
}

func (router *BookingRouter) sendBookingList(w http.ResponseWriter, list []*BookingDetails) {
	res := make([]*GetBookingResponse, 0, len(list))
	for _, e := range list {
		res = append(res, router.copyToRestModel(e))
	}
	SendJSON(w, res)
}

func (router *BookingRouter) getFiltered(w http.ResponseWriter, r *http.Request) {
	user, filterUserEmail, filterLocationId, ok := router.validateBookingFilters(w, r)
	if !ok {
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

	list, err := GetBookingRepository().GetAllByOrg(user.OrganizationID, start, end, filterUserEmail, filterLocationId)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	router.sendBookingList(w, list)
}

func (router *BookingRouter) getCurrent(w http.ResponseWriter, r *http.Request) {
	user, filterUserEmail, filterLocationId, ok := router.validateBookingFilters(w, r)
	if !ok {
		return
	}

	list, err := GetBookingRepository().GetAllCurrentByOrg(user.OrganizationID, filterUserEmail, filterLocationId)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	router.sendBookingList(w, list)
}

func (router *BookingRouter) getIcal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetBookingRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
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
	requestUser := GetRequestUser(r)
	if !CanAccessOrg(requestUser, location.OrganizationID) && e.UserID != GetRequestUserID(r) {
		SendForbidden(w)
		return
	}
	if e.UserID != GetRequestUserID(r) && !HasPermission(requestUser, location.OrganizationID, PermissionBookings, PermissionLevelRead) {
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
	icalEvent := caldavClient.GetCaldavEvent([]*CalDAVEvent{calDavEvent})
	var buf bytes.Buffer
	if err := ical.NewEncoder(&buf).Encode(icalEvent); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	w.Header().Set("Content-Type", "text/calendar")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+router.getICalFilename(calDavEvent)+"\"")
	w.Write(buf.Bytes())
}

func (router *BookingRouter) getICalFilename(calDavEvent *CalDAVEvent) string {
	filename := fmt.Sprintf("seatsurfing-%s-%s.ics", calDavEvent.Start.Format("20060102"), calDavEvent.Start.Format("1504"))
	return filename
}

func (router *BookingRouter) getOne(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetBookingRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
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
	requestUser := GetRequestUser(r)
	if !CanAccessOrg(requestUser, location.OrganizationID) && e.UserID != GetRequestUserID(r) {
		SendForbidden(w)
		return
	}
	if e.UserID != GetRequestUserID(r) && !HasPermission(requestUser, location.OrganizationID, PermissionBookings, PermissionLevelRead) {
		SendForbidden(w)
		return
	}
	res := router.copyToRestModel(e)
	SendJSON(w, res)
}

func (router *BookingRouter) getAll(w http.ResponseWriter, r *http.Request) {
	list, err := service.GetBookingService().GetUpcomingBookingsForUser(GetRequestUser(r))
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []*GetBookingResponse{}
	for _, e := range list {
		res = append(res, router.copyToRestModel(e))
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
	if e.PublicID != "" {
		SendForbidden(w)
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

	if e.Space.Location.OrganizationID != location.OrganizationID {
		SendForbidden(w)
		return
	}
	if !CanAccessOrg(requestUser, location.OrganizationID) {
		SendForbidden(w)
		return
	}

	if e.UserID != requestUser.ID && !HasPermission(requestUser, location.OrganizationID, PermissionBookings, PermissionLevelAdmin) {
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
	eNew.PublicID = e.PublicID
	eNew.Approved = e.Approved
	if m.UserEmail != "" {
		if !HasPermission(requestUser, location.OrganizationID, PermissionBookings, PermissionLevelAdmin) {
			SendForbidden(w)
			return
		}
		if m.UserEmail == requestUser.Email {
			eNew.UserID = requestUser.ID
		} else {
			eNew.UserID, err = router.bookForUser(requestUser, m.UserEmail, w)
			if err != nil {
				SendInternalServerError(w)
				return
			}
		}
	}

	// See the matching comment in create(): this serializes the checks and
	// write below against any other concurrent create/update for the same
	// user or location.
	releaseLock, err := AcquireBookingCreateLock(eNew.UserID, location.ID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	defer releaseLock()

	bookingReq := &service.BookingInput{
		SpaceID: m.SpaceID,
		Subject: m.Subject,
		Enter:   eNew.Enter,
		Leave:   eNew.Leave,
	}

	if valid, code := service.GetBookingService().CheckBooking(bookingReq, location, requestUser, eNew.ID, 0); !valid {
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
		SendAlreadyExists(w)
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
	if (e.UserID != GetRequestUserID(r)) && !HasPermission(GetRequestUser(r), location.OrganizationID, PermissionBookings, PermissionLevelAdmin) {
		SendForbidden(w)
		return
	}
	requestUser := GetRequestUser(r)

	// leave must not be in past
	now := time.Now().UTC()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if e.Booking.Leave.Before(now) {
		SendBadRequest(w)
		return
	}

	// Check for the date, if the booking request is too close with SettingsMaxHoursBeforeDelete and the deletion can not be performed
	if service.GetBookingService().IsValidBookingHoursBeforeDelete(e, requestUser, location.OrganizationID) {
		go router.onBookingDeleted(&e.Booking, true)
		if err := GetBookingRepository().Delete(e); err != nil {
			SendInternalServerError(w)
			return
		}
		SendUpdated(w)
		return
	}
	SendForbiddenCode(w, ResponseCodeBookingMaxHoursBeforeDelete)
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
	bookingReq := &service.BookingInput{
		Enter: enterNew,
		Leave: leaveNew,
	}
	if valid, code := service.GetBookingService().CheckBooking(bookingReq, location, requestUser, "", 0); !valid {
		SendBadRequestCode(w, code)
		return
	}
	SendUpdated(w)
}

func init() {
	// Notifications, CalDAV and plugin hooks for bookings created through the
	// booking service (REST API and plugin host API alike).
	service.GetBookingService().SetOnCreated(func(e *Booking) {
		(&BookingRouter{}).onBookingCreated(e)
	})
}

// sendBookingError writes a booking service error the way the booking
// endpoints always have.
func sendBookingError(w http.ResponseWriter, e *service.BookingError) {
	switch e.Kind {
	case service.BookingErrorForbidden:
		SendForbidden(w)
	case service.BookingErrorConflict:
		SendAlreadyExistsCode(w, e.Code)
	case service.BookingErrorInternal:
		SendInternalServerError(w)
	default:
		if e.Code != 0 {
			SendBadRequestCode(w, e.Code)
		} else {
			SendBadRequest(w)
		}
	}
}

func (router *BookingRouter) create(w http.ResponseWriter, r *http.Request) {
	var m CreateBookingRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	requestUser := GetRequestUser(r)
	bookings := service.GetBookingService()
	p, bErr := bookings.PrepareCreate(requestUser, &service.BookingInput{
		SpaceID: m.SpaceID,
		Subject: m.Subject,
		Enter:   m.Enter,
		Leave:   m.Leave,
	})
	if bErr != nil {
		sendBookingError(w, bErr)
		return
	}
	if m.UserEmail != "" && m.UserEmail != requestUser.Email {
		if !HasPermission(requestUser, p.Location.OrganizationID, PermissionBookings, PermissionLevelAdmin) {
			SendForbidden(w)
			return
		}
		var err error
		p.Booking.UserID, err = router.bookForUser(requestUser, m.UserEmail, w)
		if err != nil {
			SendInternalServerError(w)
			return
		}
	}
	if bErr := bookings.CommitCreate(requestUser, p); bErr != nil {
		sendBookingError(w, bErr)
		return
	}
	SendCreated(w, p.Booking.ID)
}

func (router *BookingRouter) bookForUser(requestUser *User, userEmail string, w http.ResponseWriter) (string, error) {
	if !HasPermission(requestUser, requestUser.OrganizationID, PermissionBookings, PermissionLevelAdmin) {
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
	if !HasPermission(user, user.OrganizationID, PermissionPresenceReport, PermissionLevelRead) {
		SendForbidden(w)
		return
	}
	hideReports, _ := GetSettingsRepository().GetBool(user.OrganizationID, SettingHideReports.Name)
	if hideReports {
		SendNotFound(w)
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

	// max 31 days
	diff := end.Sub(start)
	maxDuration := 31 * 24 * time.Hour
	if diff > maxDuration {
		SendBadRequestCode(w, ResponseCodePresenceReportDateRangeTooLong)
		return
	}

	locationID := r.URL.Query().Get("locationId")
	var location *Location = nil
	if locationID != "" {
		if !ValidateGUID(locationID) {
			SendBadRequest(w)
			return
		}
		location, _ = GetLocationRepository().GetOne(locationID)
		if location == nil {
			SendBadRequest(w)
			return
		}
		if location.OrganizationID != user.OrganizationID {
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
	// Guarded like numDates above: an empty result - no bookings in the range -
	// would otherwise index items[0] and panic, answering 500 instead of an
	// empty report.
	if numUsers > 0 {
		i := 0
		for date := range items[0].Presence {
			res.Dates[i] = date
			i++
		}
	}
	sort.Strings(res.Dates)
	for i, item := range items {
		res.Users[i] = GetUserInfoSmall{
			UserID:    item.User.ID,
			Email:     item.User.Email,
			Firstname: item.User.Firstname,
			Lastname:  item.User.Lastname,
		}
		res.Presences[i] = make([]int, numDates)
		for j, date := range res.Dates {
			res.Presences[i][j] = item.Presence[date]
		}
	}
	SendJSON(w, res)
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
			decryptedPassword, err := DecryptString(pref.Value)
			if err != nil {
				log.Println("Error decrypting CalDAV password for user " + userID + ": " + err.Error())
				continue
			}
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
	enterTime, err := GetLocationRepository().AttachTimezoneInformation(e.Enter, location)
	if err != nil {
		return nil, err
	}
	leaveTime, err := GetLocationRepository().AttachTimezoneInformation(e.Leave, location)
	if err != nil {
		return nil, err
	}
	caldavEvent := &CalDAVEvent{
		ID:       e.ID,
		Title:    "Seat Reservation: " + space.Name + ", " + location.Name,
		Location: space.Name + ", " + location.Name,
		Start:    enterTime,
		End:      leaveTime,
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

func (router *BookingRouter) sendMailNotification(e *Booking, notification BookingMailNotification) {
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

	recipientEmail := ""
	recipientName := ""
	var org *Organization
	language := ""
	if e.UserID == "" {
		// Public booking: no user account, no opt-out preference.
		// Recipient info (including language, as submitted with the
		// original booking request) is taken from e.PublicName/Email/
		// Language as already loaded by the caller - the public_bookings
		// row itself may already be gone by now (e.g. declining a booking
		// deletes it before this notification is sent).
		var err error
		recipientEmail = e.PublicEmail
		recipientName = SafeRecipientName(e.PublicName, e.PublicEmail)
		org, err = GetOrganizationRepository().GetOne(location.OrganizationID)
		if err != nil || org == nil {
			log.Println(err)
			return
		}
		language = org.Language
		if e.PublicLanguage != "" {
			language = e.PublicLanguage
		}
	} else {
		active, err := GetUserPreferencesRepository().GetBool(e.UserID, PreferenceMailNotifications.Name)
		if err != nil || !active {
			return
		}
		user, err := GetUserRepository().GetOne(e.UserID)
		if err != nil || user == nil {
			log.Println(err)
			return
		}
		org, err = GetOrganizationRepository().GetOne(user.OrganizationID)
		if err != nil || org == nil {
			log.Println(err)
			return
		}
		recipientEmail = user.Email
		recipientName = user.GetSafeRecipientName()
		language = org.Language
		if userLang, err := GetUserPreferencesRepository().Get(e.UserID, PreferenceMailLanguage.Name); err == nil && userLang != "" {
			language = userLang
		}
	}

	attachments := []*MailAttachment{}
	if notification == BookingMailNotificationCreated || notification == BookingMailNotificationUpdated || notification == BookingMailNotificationApproved {
		calDavEvent, err := router.getCalDavEventFromBooking(e)
		if err != nil {
			log.Println(err)
			return
		}
		caldavClient := &CalDAVClient{}
		icalEvent := caldavClient.GetCaldavEvent([]*CalDAVEvent{calDavEvent})
		var buf bytes.Buffer
		if err := ical.NewEncoder(&buf).Encode(icalEvent); err != nil {
			log.Println(err)
			return
		}
		attachments = append(attachments, &MailAttachment{
			Filename: router.getICalFilename(calDavEvent),
			MimeType: "text/calendar",
			Data:     buf.Bytes(),
		})
	}

	subject := e.Subject
	if subject == "" {
		subject = "—"
	}
	domain, err := GetOrganizationRepository().GetPrimaryDomain(org)
	if err != nil {
		log.Println(err)
		return
	}
	vars := map[string]string{
		"orgDomain":     FormatURL(domain.DomainName) + "/",
		"recipientName": recipientName,
		"date":          e.Enter.Format("2006-01-02 15:04") + " - " + e.Leave.Format("2006-01-02 15:04"),
		"areaName":      location.Name,
		"spaceName":     space.Name,
		"subject":       subject,
	}
	// Public bookers have no account, so their approved/declined mails use
	// dedicated templates without the "your bookings" button (there is no
	// bookings page for them to sign into).
	isPublicBooking := e.UserID == ""
	if isPublicBooking && notification == BookingMailNotificationApproved {
		vars["detailsUrl"] = FormatURL(domain.DomainName) + "/ui/book/details/" + e.PublicExternalID + "/"
	}
	template := GetEmailTemplatePathBookingCreated()
	if notification == BookingMailNotificationUpdated {
		template = GetEmailTemplatePathBookingUpdated()
	} else if notification == BookingMailNotificationDeclined {
		if isPublicBooking {
			template = GetEmailTemplatePathPublicBookingDeclined()
		} else {
			template = GetEmailTemplatePathBookingDeclined()
		}
	} else if notification == BookingMailNotificationApproved {
		if isPublicBooking {
			template = GetEmailTemplatePathPublicBookingApproved()
		} else {
			template = GetEmailTemplatePathBookingApproved()
		}
	} else if notification == BookingMailNotificationDeleted {
		if isPublicBooking {
			template = GetEmailTemplatePathPublicBookingDeleted()
		} else {
			template = GetEmailTemplatePathBookingDeleted()
		}
	}
	if err := SendEmailWithAttachmentsAndOrg(&MailAddress{Address: recipientEmail}, template, language, vars, attachments, org.ID); err != nil {
		log.Println(err)
		return
	}
	now := time.Now().UTC()
	if err := GetBookingRepository().UpdateLastInfoMailSentAt(e.ID, &now); err != nil {
		log.Println(err)
	}
}

func (router *BookingRouter) onBookingUpdated(e *Booking) {
	router.updateCalDavEvent(e)
	for _, plg := range GetPlugins() {
		plg.OnBookingUpdated(e.ID)
	}
	router.sendMailNotification(e, BookingMailNotificationUpdated)
}

func (router *BookingRouter) onBookingDeclinedOrApproved(e *Booking) {
	if !e.Approved {
		router.onBookingDeleted(e, false)
		router.sendMailNotification(e, BookingMailNotificationDeclined)
	} else {
		router.createCalDavEvent(e)
		for _, plg := range GetPlugins() {
			plg.OnBookingCreated(e.ID)
		}
		router.sendMailNotification(e, BookingMailNotificationApproved)
	}
}

func (router *BookingRouter) onBookingCreated(e *Booking) {
	if e.Approved {
		router.createCalDavEvent(e)
		for _, plg := range GetPlugins() {
			plg.OnBookingCreated(e.ID)
		}
		router.sendMailNotification(e, BookingMailNotificationCreated)
	} else {
		// Booking requires approval - notify approvers
		router.sendApprovalRequestNotifications(e)
	}
}

func (router *BookingRouter) sendApprovalRequestNotifications(e *Booking) {
	// Get the space to find approver groups
	space, err := GetSpaceRepository().GetOne(e.SpaceID)
	if err != nil {
		log.Println("Error getting space:", err)
		return
	}

	// Get approver group IDs for this space
	approverGroupIDs, err := GetSpaceRepository().GetApproverGroupIDs(e.SpaceID)
	if err != nil || len(approverGroupIDs) == 0 {
		log.Println("Error getting approver groups or no approvers:", err)
		return
	}

	// Get location for timezone information
	location, err := GetLocationRepository().GetOne(space.LocationID)
	if err != nil {
		log.Println("Error getting location:", err)
		return
	}

	// Get organization for language settings
	org, err := GetOrganizationRepository().GetOne(location.OrganizationID)
	if err != nil {
		log.Println("Error getting organization:", err)
		return
	}

	// Get booking requester info (a real user, or a public booker)
	bookingUserEmail := ""
	if e.UserID == "" {
		publicBooking, err := GetPublicBookingRepository().GetOne(string(e.PublicID))
		if err != nil || publicBooking == nil {
			log.Println("Error getting public booking:", err)
			return
		}
		bookingUserEmail = publicBooking.Email
	} else {
		bookingUser, err := GetUserRepository().GetOne(e.UserID)
		if err != nil {
			log.Println("Error getting booking user:", err)
			return
		}
		bookingUserEmail = bookingUser.Email
	}

	// Collect all unique approver user IDs who have the preference enabled
	approverUserIDs := make(map[string]bool)
	for _, groupID := range approverGroupIDs {
		group, err := GetGroupRepository().GetOne(groupID)
		if err != nil {
			log.Println("Error getting group:", err)
			continue
		}

		memberIDs, err := GetGroupRepository().GetMemberUserIDs(group)
		if err != nil {
			log.Println("Error getting group members:", err)
			continue
		}

		for _, userID := range memberIDs {
			// Check if user has approval notifications enabled
			notificationsEnabled, err := GetUserPreferencesRepository().GetBool(userID, PreferenceApprovalNotifications.Name)
			if err == nil && notificationsEnabled {
				approverUserIDs[userID] = true
			}
		}
	}

	subject := e.Subject
	if subject == "" {
		subject = "—"
	}
	domain, err := GetOrganizationRepository().GetPrimaryDomain(org)
	if err != nil {
		log.Println(err)
		return
	}

	// Send email to each approver
	for userID := range approverUserIDs {
		approver, err := GetUserRepository().GetOne(userID)
		if err != nil {
			log.Println("Error getting approver user:", err)
			continue
		}

		vars := map[string]string{
			"orgDomain":     FormatURL(domain.DomainName) + "/",
			"recipientName": approver.GetSafeRecipientName(),
			"userEmail":     bookingUserEmail,
			"date":          e.Enter.Format("2006-01-02 15:04") + " - " + e.Leave.Format("2006-01-02 15:04"),
			"areaName":      location.Name,
			"spaceName":     space.Name,
			"subject":       subject,
		}

		approverLang := org.Language
		if userLang, err := GetUserPreferencesRepository().Get(approver.ID, PreferenceMailLanguage.Name); err == nil && userLang != "" {
			approverLang = userLang
		}
		template := GetEmailTemplatePathBookingApprovalRequest()
		if err := SendEmailWithOrg(&MailAddress{Address: approver.Email}, template, approverLang, vars, org.ID); err != nil {
			log.Println("Error sending approval notification email:", err)
		}
	}
}

func (router *BookingRouter) onBookingDeleted(e *Booking, sendNotification bool) {
	for _, plg := range GetPlugins() {
		plg.OnBookingDeleted(e.ID)
	}
	caldavClient, caldavEvent, path, err := router.initCaldavEvent(e)
	if err == nil {
		if e.CalDavID != "" {
			caldavEvent.ID = e.CalDavID
			if err := caldavClient.DeleteEvent(path, caldavEvent); err != nil {
				log.Println(err)
			}
		}
	}

	if sendNotification {
		router.sendMailNotification(e, BookingMailNotificationDeleted)
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
	m.UserFirstname = e.UserFirstname
	m.UserLastname = e.UserLastname
	if e.UserID == "" {
		m.Public = true
		m.UserEmail = e.PublicEmail
		m.UserFirstname = e.PublicName
	}
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
