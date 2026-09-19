package router

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

// PublicBookingRouter serves the fully unauthenticated anonymous-booking flow:
// listing anonymous-bookable spaces, requesting a booking (double opt-in),
// and confirming it. Every handler re-validates that anonymous booking is
// actually enabled for the org/space rather than trusting the unauthenticated
// route whitelist, following the same principle as KioskRouter.
type PublicBookingRouter struct {
}

const anonymousBookingConfirmExpiry = 30 * time.Minute
const maxPendingAnonymousBookingRequestsPerEmail = 2

type GetAnonymousBookableSpaceResponse struct {
	SpaceID        string `json:"spaceId"`
	SpaceName      string `json:"spaceName"`
	LocationID     string `json:"locationId"`
	LocationName   string `json:"locationName"`
	RequireSubject bool   `json:"requireSubject"`
}

type GetAnonymousBookableSpacesResponse struct {
	Spaces           []GetAnonymousBookableSpaceResponse `json:"spaces"`
	MaxDaysInAdvance int                                  `json:"maxDaysInAdvance"`
}

type CreateAnonymousBookingRequest struct {
	SpaceID string    `json:"spaceId" validate:"required,uuid"`
	Enter   time.Time `json:"enter" validate:"required"`
	Leave   time.Time `json:"leave" validate:"required"`
	Name    string    `json:"name" validate:"required,max=256"`
	Email   string    `json:"email" validate:"required,email,max=256"`
	Subject string    `json:"subject" validate:"omitempty,max=256"`
}

type ConfirmAnonymousBookingResponse struct {
	Status string `json:"status"`
}

func (router *PublicBookingRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/{orgId}/spaces", router.getSpaces).Methods("GET")
	s.HandleFunc("/{orgId}/request", router.request).Methods("POST")
	s.HandleFunc("/confirm/{id}", router.confirm).Methods("POST")
}

func (router *PublicBookingRouter) isAnonymousBookingEnabledForOrg(orgID string) bool {
	enabled, _ := GetSettingsRepository().GetBool(orgID, SettingAnonymousBookingEnabled.Name)
	return enabled
}

func (router *PublicBookingRouter) getSpaces(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgId"]
	if !router.isAnonymousBookingEnabledForOrg(orgID) {
		SendNotFound(w)
		return
	}
	spaces, err := GetSpaceRepository().GetAllAnonymousBookable(orgID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []GetAnonymousBookableSpaceResponse{}
	for _, space := range spaces {
		location, err := GetLocationRepository().GetOne(space.LocationID)
		if err != nil {
			continue
		}
		res = append(res, GetAnonymousBookableSpaceResponse{
			SpaceID:        space.ID,
			SpaceName:      space.Name,
			LocationID:     location.ID,
			LocationName:   location.Name,
			RequireSubject: space.RequireSubject,
		})
	}
	maxDaysInAdvance, _ := GetSettingsRepository().GetInt(orgID, SettingMaxDaysInAdvance.Name)
	SendJSON(w, GetAnonymousBookableSpacesResponse{
		Spaces:           res,
		MaxDaysInAdvance: maxDaysInAdvance,
	})
}

// validateSpaceAndTimes re-checks every condition that makes a slot
// requestable: org/space anonymous-booking enablement, space/location
// enablement, and the org's normal booking-duration/advance/weekday rules
// (reusing BookingRouter's checks with a nil user, which they support).
func (router *PublicBookingRouter) validateSpaceAndTimes(orgID string, m *CreateAnonymousBookingRequest) (*Space, *Location, time.Time, time.Time, bool) {
	space, err := GetSpaceRepository().GetOne(m.SpaceID)
	if err != nil {
		return nil, nil, time.Time{}, time.Time{}, false
	}
	location, err := GetLocationRepository().GetOne(space.LocationID)
	if err != nil || location.OrganizationID != orgID {
		return nil, nil, time.Time{}, time.Time{}, false
	}
	if !space.AnonymousBookingEnabled || !space.Enabled || !location.Enabled {
		return nil, nil, time.Time{}, time.Time{}, false
	}
	enter, err := GetLocationRepository().AttachTimezoneInformation(m.Enter, location)
	if err != nil {
		return nil, nil, time.Time{}, time.Time{}, false
	}
	leave, err := GetLocationRepository().AttachTimezoneInformation(m.Leave, location)
	if err != nil || !leave.After(enter) {
		return nil, nil, time.Time{}, time.Time{}, false
	}
	if enter.Year() != leave.Year() || enter.YearDay() != leave.YearDay() {
		// Anonymous bookings must not span across a day boundary.
		return nil, nil, time.Time{}, time.Time{}, false
	}
	bookingRequest := &BookingRequest{Enter: enter, Leave: leave}
	bookingRouter := &BookingRouter{}
	if !bookingRouter.IsValidBookingDuration(bookingRequest, orgID, nil) {
		return nil, nil, time.Time{}, time.Time{}, false
	}
	if valid, _ := bookingRouter.IsValidBookingAdvance(bookingRequest, orgID, nil); !valid {
		return nil, nil, time.Time{}, time.Time{}, false
	}
	if !bookingRouter.isValidMinHoursBooking(bookingRequest, orgID, nil) {
		return nil, nil, time.Time{}, time.Time{}, false
	}
	if !IsLocationWeekdayBookable(location, nil, enter, leave) {
		return nil, nil, time.Time{}, time.Time{}, false
	}
	return space, location, enter, leave, true
}

func (router *PublicBookingRouter) request(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgId"]
	if !router.isAnonymousBookingEnabledForOrg(orgID) {
		SendNotFound(w)
		return
	}
	var m CreateAnonymousBookingRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	name := strings.TrimSpace(m.Name)
	if !IsValidHumanName(name) {
		SendBadRequest(w)
		return
	}
	org, err := GetOrganizationRepository().GetOne(orgID)
	if err != nil || org == nil {
		SendBadRequest(w)
		return
	}
	space, location, enter, leave, valid := router.validateSpaceAndTimes(orgID, &m)
	if !valid {
		SendBadRequest(w)
		return
	}
	globalRequireSubjectSetting, _ := GetSettingsRepository().GetInt(orgID, SettingSubjectDefault.Name)
	if globalRequireSubjectSetting != SettingSubjectDefaultDisabled && space.RequireSubject && len(strings.TrimSpace(m.Subject)) < 3 {
		SendBadRequestCode(w, ResponseCodeBookingSubjectRequired)
		return
	}

	// Cap the number of emails (confirmation or "sorry") sent to any one
	// address within the confirmation window, to prevent using this public
	// endpoint to mail-bomb an arbitrary address. Every accepted request
	// below records a marker via the auth_states table (see below), whether
	// or not the slot turned out to be free, so this check also catches
	// repeated requests against an already-booked slot.
	if router.countPendingRequestsForEmail(m.Email) >= maxPendingAnonymousBookingRequestsPerEmail {
		// Don't disclose rate limiting to the caller.
		SendUpdated(w)
		return
	}

	conflicts, err := GetBookingRepository().GetConflicts(space.ID, enter, leave, "")
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	if len(conflicts) > 0 {
		router.createAuthState(AnonymousBookingRequestPayload{Email: m.Email})
		router.sendUnavailableMail(org, location, space, name, m.Email, enter, leave, m.Subject)
		SendUpdated(w)
		return
	}

	authState, err := router.createAuthState(AnonymousBookingRequestPayload{
		SpaceID: space.ID,
		Enter:   enter,
		Leave:   leave,
		Name:    name,
		Email:   m.Email,
		Subject: m.Subject,
	})
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	router.sendConfirmMail(org, location, space, name, m.Email, enter, leave, authState.ID)
	SendUpdated(w)
}

// countPendingRequestsForEmail counts non-expired anonymous-booking auth
// states (confirmations and "sorry" markers alike) for the given email.
func (router *PublicBookingRouter) countPendingRequestsForEmail(email string) int {
	pending, err := GetAuthStateRepository().GetActiveByType(AuthAnonymousBooking)
	if err != nil {
		return 0
	}
	count := 0
	for _, state := range pending {
		var payload AnonymousBookingRequestPayload
		if json.Unmarshal([]byte(state.Payload), &payload) == nil && strings.EqualFold(payload.Email, email) {
			count++
		}
	}
	return count
}

func (router *PublicBookingRouter) createAuthState(payload AnonymousBookingRequestPayload) (*AuthState, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	authState := &AuthState{
		AuthProviderID: GetSettingsRepository().GetNullUUID(),
		Expiry:         time.Now().Add(anonymousBookingConfirmExpiry),
		AuthStateType:  AuthAnonymousBooking,
		Payload:        string(payloadJSON),
	}
	if err := GetAuthStateRepository().Create(authState); err != nil {
		return nil, err
	}
	return authState, nil
}

func (router *PublicBookingRouter) confirm(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	authState, err := GetAuthStateRepository().GetOneActive(id)
	if err != nil || authState.AuthStateType != AuthAnonymousBooking {
		SendNotFound(w)
		return
	}
	var payload AnonymousBookingRequestPayload
	if json.Unmarshal([]byte(authState.Payload), &payload) != nil {
		SendNotFound(w)
		return
	}
	// Single-use: remove the state now regardless of outcome below.
	defer GetAuthStateRepository().Delete(authState)

	space, err := GetSpaceRepository().GetOne(payload.SpaceID)
	if err != nil {
		SendNotFound(w)
		return
	}
	location, err := GetLocationRepository().GetOne(space.LocationID)
	if err != nil {
		SendNotFound(w)
		return
	}
	org, err := GetOrganizationRepository().GetOne(location.OrganizationID)
	if err != nil || org == nil {
		SendNotFound(w)
		return
	}
	if !router.isAnonymousBookingEnabledForOrg(org.ID) || !space.AnonymousBookingEnabled || !space.Enabled || !location.Enabled {
		SendNotFound(w)
		return
	}

	conflicts, err := GetBookingRepository().GetConflicts(space.ID, payload.Enter, payload.Leave, "")
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	if len(conflicts) > 0 {
		router.sendUnavailableMail(org, location, space, payload.Name, payload.Email, payload.Enter, payload.Leave, payload.Subject)
		SendJSON(w, ConfirmAnonymousBookingResponse{Status: "unavailable"})
		return
	}

	anonymousBooking := &AnonymousBooking{
		Name:  payload.Name,
		Email: payload.Email,
	}
	if err := GetAnonymousBookingRepository().Create(anonymousBooking); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	booking := &Booking{
		SpaceID:     space.ID,
		Enter:       payload.Enter,
		Leave:       payload.Leave,
		Subject:     payload.Subject,
		Approved:    false, // anonymous booking is only allowed on spaces that always have an approver group
		AnonymousID: NullUUID(anonymousBooking.ID),
	}
	if err := GetBookingRepository().Create(booking); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}

	bookingRouter := &BookingRouter{}
	go bookingRouter.sendApprovalRequestNotifications(booking)

	SendJSON(w, ConfirmAnonymousBookingResponse{Status: "pending"})
}

func (router *PublicBookingRouter) sendConfirmMail(org *Organization, location *Location, space *Space, name, email string, enter, leave time.Time, confirmID string) {
	domain, err := GetOrganizationRepository().GetPrimaryDomain(org)
	if err != nil {
		log.Println(err)
		return
	}
	vars := map[string]string{
		"orgDomain":     FormatURL(domain.DomainName) + "/",
		"recipientName": SafeRecipientName(name, email),
		"date":          enter.Format("2006-01-02 15:04") + " - " + leave.Format("2006-01-02 15:04"),
		"areaName":      location.Name,
		"spaceName":     space.Name,
		"confirmID":     confirmID,
	}
	if err := SendEmailWithOrg(&MailAddress{Address: email}, GetEmailTemplatePathAnonymousBookingConfirm(), org.Language, vars, org.ID); err != nil {
		log.Println(err)
	}
}

func (router *PublicBookingRouter) sendUnavailableMail(org *Organization, location *Location, space *Space, name, email string, enter, leave time.Time, subject string) {
	domain, err := GetOrganizationRepository().GetPrimaryDomain(org)
	if err != nil {
		log.Println(err)
		return
	}
	vars := map[string]string{
		"orgDomain":     FormatURL(domain.DomainName) + "/",
		"recipientName": SafeRecipientName(name, email),
		"date":          enter.Format("2006-01-02 15:04") + " - " + leave.Format("2006-01-02 15:04"),
		"areaName":      location.Name,
		"spaceName":     space.Name,
		"subject":       subject,
	}
	if err := SendEmailWithOrg(&MailAddress{Address: email}, GetEmailTemplatePathAnonymousBookingUnavailable(), org.Language, vars, org.ID); err != nil {
		log.Println(err)
	}
}
