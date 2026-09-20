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

// PublicBookingRouter serves the fully unauthenticated public-booking flow:
// listing public-bookable spaces, requesting a booking (double opt-in),
// and confirming it. Every handler re-validates that public booking is
// actually enabled for the org/space rather than trusting the unauthenticated
// route whitelist, following the same principle as KioskRouter.
type PublicBookingRouter struct {
}

const publicBookingConfirmExpiry = 30 * time.Minute
const maxPendingPublicBookingRequestsPerEmail = 5

type GetPublicBookableSpaceResponse struct {
	SpaceID        string `json:"spaceId"`
	SpaceName      string `json:"spaceName"`
	LocationID     string `json:"locationId"`
	LocationName   string `json:"locationName"`
	RequireSubject bool   `json:"requireSubject"`
}

type GetPublicBookableSpacesResponse struct {
	Spaces           []GetPublicBookableSpaceResponse `json:"spaces"`
	MaxDaysInAdvance int                              `json:"maxDaysInAdvance"`
}

type CreatePublicBookingRequest struct {
	SpaceID  string    `json:"spaceId" validate:"required,uuid"`
	Enter    time.Time `json:"enter" validate:"required"`
	Leave    time.Time `json:"leave" validate:"required"`
	Name     string    `json:"name" validate:"required,max=256"`
	Email    string    `json:"email" validate:"required,email,max=256"`
	Subject  string    `json:"subject" validate:"omitempty,max=256"`
	Language string    `json:"language" validate:"omitempty,max=10"`
}

type ConfirmPublicBookingResponse struct {
	Status string    `json:"status"`
	Enter  time.Time `json:"enter"`
	Leave  time.Time `json:"leave"`
}

func (router *PublicBookingRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/{orgId}/spaces", router.getSpaces).Methods("GET")
	s.HandleFunc("/{orgId}/request", router.request).Methods("POST")
	s.HandleFunc("/confirm/{id}", router.confirm).Methods("POST")
}

func (router *PublicBookingRouter) isPublicBookingEnabledForOrg(orgID string) bool {
	featureEnabled, _ := GetSettingsRepository().GetBool(orgID, SettingFeaturePublicBooking.Name)
	if !featureEnabled {
		return false
	}
	enabled, _ := GetSettingsRepository().GetBool(orgID, SettingPublicBookingEnabled.Name)
	return enabled
}

func (router *PublicBookingRouter) getSpaces(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgId"]
	if !router.isPublicBookingEnabledForOrg(orgID) {
		SendNotFound(w)
		return
	}
	spaces, err := GetSpaceRepository().GetAllPublicBookable(orgID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []GetPublicBookableSpaceResponse{}
	for _, space := range spaces {
		location, err := GetLocationRepository().GetOne(space.LocationID)
		if err != nil {
			continue
		}
		res = append(res, GetPublicBookableSpaceResponse{
			SpaceID:        space.ID,
			SpaceName:      space.Name,
			LocationID:     location.ID,
			LocationName:   location.Name,
			RequireSubject: space.RequireSubject,
		})
	}
	maxDaysInAdvance, _ := GetSettingsRepository().GetInt(orgID, SettingMaxDaysInAdvance.Name)
	SendJSON(w, GetPublicBookableSpacesResponse{
		Spaces:           res,
		MaxDaysInAdvance: maxDaysInAdvance,
	})
}

// validateSpaceAndTimes re-checks every condition that makes a slot
// requestable: org/space public-booking enablement, space/location
// enablement, and the org's normal booking-duration/advance/weekday rules
// (reusing BookingRouter's checks with a nil user, which they support).
func (router *PublicBookingRouter) validateSpaceAndTimes(orgID string, m *CreatePublicBookingRequest) (*Space, *Location, time.Time, time.Time, bool) {
	space, err := GetSpaceRepository().GetOne(m.SpaceID)
	if err != nil {
		return nil, nil, time.Time{}, time.Time{}, false
	}
	location, err := GetLocationRepository().GetOne(space.LocationID)
	if err != nil || location.OrganizationID != orgID {
		return nil, nil, time.Time{}, time.Time{}, false
	}
	if !space.PublicBookingEnabled || !space.Enabled || !location.Enabled {
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
		// Public bookings must not span across a day boundary.
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

// normalizePublicBookingLanguage maps the UI locale submitted with a
// public booking request (e.g. "de", "en-GB", "fr") down to one of the
// backend's supported email languages, defaulting to English for anything
// that isn't German.
func normalizePublicBookingLanguage(lang string) string {
	if strings.HasPrefix(strings.ToLower(lang), "de") {
		return "de"
	}
	return "en"
}

func (router *PublicBookingRouter) request(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgId"]
	if !router.isPublicBookingEnabledForOrg(orgID) {
		SendNotFound(w)
		return
	}
	var m CreatePublicBookingRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	language := normalizePublicBookingLanguage(m.Language)
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

	// Cap the number of confirmation emails sent to any one address within
	// the confirmation window, to prevent using this public endpoint to
	// mail-bomb an arbitrary address. Availability is deliberately not
	// checked here anymore; it's re-checked in confirm() instead, so every
	// marker created below is single-use and gets deleted there (see the
	// defer in confirm()) whether or not the slot turns out to still be
	// free. That keeps a single opt-in step in charge of releasing it,
	// rather than requiring a second mechanism for markers that never get
	// a confirm click.
	if router.countPendingRequestsForEmail(m.Email) >= maxPendingPublicBookingRequestsPerEmail {
		// Don't disclose rate limiting to the caller.
		SendUpdated(w)
		return
	}

	authState, err := router.createAuthState(PublicBookingRequestPayload{
		SpaceID:  space.ID,
		Enter:    enter,
		Leave:    leave,
		Name:     name,
		Email:    m.Email,
		Subject:  m.Subject,
		Language: language,
	})
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	router.sendConfirmMail(org, location, space, name, m.Email, language, enter, leave, authState.ID)
	SendUpdated(w)
}

// countPendingRequestsForEmail counts non-expired public-booking auth
// states for the given email, via the indexed Key column rather than
// scanning and decoding every payload.
func (router *PublicBookingRouter) countPendingRequestsForEmail(email string) int {
	pending, err := GetAuthStateRepository().GetActiveByKeyAndType(strings.ToLower(email), AuthPublicBooking)
	if err != nil {
		return 0
	}
	return len(pending)
}

func (router *PublicBookingRouter) createAuthState(payload PublicBookingRequestPayload) (*AuthState, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	authState := &AuthState{
		Expiry:        time.Now().Add(publicBookingConfirmExpiry),
		AuthStateType: AuthPublicBooking,
		Payload:       string(payloadJSON),
		Key:           strings.ToLower(payload.Email),
	}
	if err := GetAuthStateRepository().Create(authState); err != nil {
		return nil, err
	}
	return authState, nil
}

func (router *PublicBookingRouter) confirm(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	authState, err := GetAuthStateRepository().ClaimActive(id)
	if err != nil || authState.AuthStateType != AuthPublicBooking {
		SendNotFound(w)
		return
	}
	var payload PublicBookingRequestPayload
	if json.Unmarshal([]byte(authState.Payload), &payload) != nil {
		SendNotFound(w)
		return
	}

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
	if !router.isPublicBookingEnabledForOrg(org.ID) || !space.PublicBookingEnabled || !space.Enabled || !location.Enabled {
		SendNotFound(w)
		return
	}

	releaseLock, err := AcquireBookingCreateLock("", location.ID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	defer releaseLock()

	conflicts, err := GetBookingRepository().GetConflicts(space.ID, payload.Enter, payload.Leave, "")
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	if len(conflicts) > 0 {
		SendJSON(w, ConfirmPublicBookingResponse{Status: "unavailable", Enter: payload.Enter, Leave: payload.Leave})
		return
	}
	if location.MaxConcurrentBookings > 0 {
		concurrent, err := GetBookingRepository().GetConcurrent(location, payload.Enter, payload.Leave, "")
		if err != nil {
			log.Println(err)
			SendInternalServerError(w)
			return
		}
		if concurrent >= int(location.MaxConcurrentBookings) {
			SendJSON(w, ConfirmPublicBookingResponse{Status: "unavailable", Enter: payload.Enter, Leave: payload.Leave})
			return
		}
	}

	publicBooking := &PublicBooking{
		Name:     payload.Name,
		Email:    payload.Email,
		Language: payload.Language,
	}
	if err := GetPublicBookingRepository().Create(publicBooking); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	booking := &Booking{
		SpaceID:  space.ID,
		Enter:    payload.Enter,
		Leave:    payload.Leave,
		Subject:  payload.Subject,
		Approved: false, // public booking is only allowed on spaces that always have an approver group
		PublicID: NullUUID(publicBooking.ID),
	}
	if err := GetBookingRepository().Create(booking); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}

	bookingRouter := &BookingRouter{}
	go bookingRouter.sendApprovalRequestNotifications(booking)

	SendJSON(w, ConfirmPublicBookingResponse{Status: "pending", Enter: payload.Enter, Leave: payload.Leave})
}

func (router *PublicBookingRouter) sendConfirmMail(org *Organization, location *Location, space *Space, name, email, language string, enter, leave time.Time, confirmID string) {
	domain, err := GetOrganizationRepository().GetPrimaryDomain(org)
	if err != nil {
		log.Println(err)
		return
	}
	if language == "" {
		language = org.Language
	}
	vars := map[string]string{
		"orgDomain":     FormatURL(domain.DomainName) + "/",
		"recipientName": SafeRecipientName(name, email),
		"date":          enter.Format("2006-01-02 15:04") + " - " + leave.Format("2006-01-02 15:04"),
		"areaName":      location.Name,
		"spaceName":     space.Name,
		"confirmID":     confirmID,
	}
	if err := SendEmailWithOrg(&MailAddress{Address: email}, GetEmailTemplatePathPublicBookingConfirm(), language, vars, org.ID); err != nil {
		log.Println(err)
	}
}
