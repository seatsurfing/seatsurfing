package service

import (
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

// BookingService holds the booking business and validation logic shared by
// the REST API (package router) and the plugin host API.
type BookingService struct {
	onCreatedMu sync.RWMutex
	onCreated   func(e *Booking)
}

var bookingService *BookingService
var bookingServiceOnce sync.Once

func GetBookingService() *BookingService {
	bookingServiceOnce.Do(func() {
		bookingService = &BookingService{}
	})
	return bookingService
}

// Booking error codes, sent by the REST API as X-Error-Code.
const (
	BookingCodeSlotConflict              = 1001
	BookingCodeLocationMaxConcurrent     = 1002
	BookingCodeTooManyUpcomingBookings   = 1003
	BookingCodeTooManyDaysInAdvance      = 1004
	BookingCodeInvalidBookingDuration    = 1005
	BookingCodeMaxConcurrentForUser      = 1006
	BookingCodeInvalidMinBookingDuration = 1007
	BookingCodeMaxHoursBeforeDelete      = 1008
	BookingCodeNotAllowedBooker          = 1009
	BookingCodeSubjectRequired           = 1010
	BookingCodeInPast                    = 1011
	BookingCodeInvalidSubject            = 1012
	BookingCodeInvalidWeekday            = 1013
)

// BookingErrorKind classifies why a booking operation failed.
type BookingErrorKind int

const (
	BookingErrorInvalid   BookingErrorKind = iota + 1 // the request violates a booking rule or is malformed
	BookingErrorForbidden                             // the user may not perform the operation
	BookingErrorConflict                              // the space is already booked in the time slot
	BookingErrorInternal                              // unexpected failure
)

// BookingError describes why a booking could not be created. Code is one of
// the BookingCode* values, or 0.
type BookingError struct {
	Kind BookingErrorKind
	Code int
}

func (e *BookingError) Error() string {
	return fmt.Sprintf("booking failed: kind %d, code %d", e.Kind, e.Code)
}

// BookingInput is a booking to validate or create.
type BookingInput struct {
	SpaceID string
	Subject string
	// Enter and Leave are wall-clock times; their time zone is replaced by
	// the location's.
	Enter time.Time
	Leave time.Time
}

// PreparedBooking is a validated booking not yet stored, see PrepareCreate.
type PreparedBooking struct {
	// Booking is to be created. Its UserID defaults to the requesting user
	// and may be changed before CommitCreate.
	Booking  *Booking
	Space    *Space
	Location *Location
}

// SetOnCreated registers the function called (asynchronously) after a
// booking was created, e.g. to send notifications. The router registers it.
func (s *BookingService) SetOnCreated(f func(e *Booking)) {
	s.onCreatedMu.Lock()
	defer s.onCreatedMu.Unlock()
	s.onCreated = f
}

func (s *BookingService) notifyCreated(e *Booking) {
	s.onCreatedMu.RLock()
	f := s.onCreated
	s.onCreatedMu.RUnlock()
	if f != nil {
		go f(e)
	}
}

// CreateBooking creates a booking for requestUser themselves, applying all
// booking rules.
func (s *BookingService) CreateBooking(requestUser *User, in *BookingInput) (*Booking, *BookingError) {
	p, bErr := s.PrepareCreate(requestUser, in)
	if bErr != nil {
		return nil, bErr
	}
	if bErr := s.CommitCreate(requestUser, p); bErr != nil {
		return nil, bErr
	}
	return p.Booking, nil
}

// PrepareCreate performs the checks that don't depend on who the booking is
// for and converts in into a booking for requestUser. Callers booking for
// someone else set PreparedBooking.Booking.UserID before CommitCreate.
func (s *BookingService) PrepareCreate(requestUser *User, in *BookingInput) (*PreparedBooking, *BookingError) {
	if requestUser == nil {
		return nil, &BookingError{Kind: BookingErrorForbidden}
	}
	if in.SpaceID == "" || len(in.Subject) > 256 {
		return nil, &BookingError{Kind: BookingErrorInvalid}
	}
	space, err := GetSpaceRepository().GetOne(in.SpaceID)
	if err != nil {
		return nil, &BookingError{Kind: BookingErrorInvalid}
	}
	location, err := GetLocationRepository().GetOne(space.LocationID)
	if err != nil {
		return nil, &BookingError{Kind: BookingErrorInvalid}
	}
	// Check organization access before any space-specific rule, so nothing
	// about another organization's spaces is revealed.
	if !CanAccessOrg(requestUser, location.OrganizationID) {
		return nil, &BookingError{Kind: BookingErrorForbidden}
	}
	globalRequireSubjectSetting, _ := GetSettingsRepository().GetInt(location.OrganizationID, SettingSubjectDefault.Name)
	if globalRequireSubjectSetting != SettingSubjectDefaultDisabled {
		if space.RequireSubject && len(strings.TrimSpace(in.Subject)) < 3 {
			return nil, &BookingError{Kind: BookingErrorInvalid, Code: BookingCodeSubjectRequired}
		}
	}
	// Disabled locations and spaces can only be booked by space admins.
	if (!location.Enabled || !space.Enabled) && !HasPermission(requestUser, location.OrganizationID, PermissionAreas, PermissionLevelAdmin) {
		return nil, &BookingError{Kind: BookingErrorInvalid}
	}
	enter, err := GetLocationRepository().AttachTimezoneInformation(in.Enter, location)
	if err != nil {
		return nil, &BookingError{Kind: BookingErrorInternal}
	}
	leave, err := GetLocationRepository().AttachTimezoneInformation(in.Leave, location)
	if err != nil {
		return nil, &BookingError{Kind: BookingErrorInternal}
	}
	e := &Booking{
		UserID:  requestUser.ID,
		SpaceID: in.SpaceID,
		Subject: in.Subject,
		Enter:   enter,
		Leave:   leave,
	}
	return &PreparedBooking{Booking: e, Space: space, Location: location}, nil
}

// CommitCreate validates a prepared booking against the booking rules and
// stores it.
func (s *BookingService) CommitCreate(requestUser *User, p *PreparedBooking) *BookingError {
	e := p.Booking
	// Hold the create lock for the rest of this function: it serializes this
	// request against any other concurrent create/update for the same user
	// or location, so the checks below and the insert they guard can't race
	// with another request's checks and insert.
	releaseLock, err := AcquireBookingCreateLock(e.UserID, p.Location.ID)
	if err != nil {
		log.Println(err)
		return &BookingError{Kind: BookingErrorInternal}
	}
	defer releaseLock()

	in := &BookingInput{SpaceID: e.SpaceID, Subject: e.Subject, Enter: e.Enter, Leave: e.Leave}
	if valid, code := s.CheckBooking(in, p.Location, requestUser, "", 0); !valid {
		return &BookingError{Kind: BookingErrorInvalid, Code: code}
	}
	conflicts, err := GetBookingRepository().GetConflicts(e.SpaceID, e.Enter, e.Leave, "")
	if err != nil {
		log.Println(err)
		return &BookingError{Kind: BookingErrorInternal}
	}
	if len(conflicts) > 0 {
		return &BookingError{Kind: BookingErrorConflict, Code: BookingCodeSlotConflict}
	}
	e.Approved = !GetSpaceService().RequiresApproval(p.Location.OrganizationID, p.Space)
	if err := GetBookingRepository().Create(e); err != nil {
		log.Println(err)
		return &BookingError{Kind: BookingErrorInternal}
	}
	s.notifyCreated(e)
	return nil
}

// GetUpcomingBookingsForUser returns the user's own bookings that have not
// ended yet, judged by the wall-clock time at each booking's location.
func (s *BookingService) GetUpcomingBookingsForUser(user *User) ([]*BookingDetails, error) {
	startTime := time.Now().UTC().Add(time.Hour * -12)
	list, err := GetBookingRepository().GetAllByUser(user.ID, startTime)
	if err != nil {
		return nil, err
	}
	defaultTz, err := GetSettingsRepository().Get(user.OrganizationID, SettingDefaultTimezone.Name)
	if err != nil {
		defaultTz = "UTC"
	}
	nowAtOrg, _ := GetUTCNowInTimezone(defaultTz)
	res := []*BookingDetails{}
	for _, e := range list {
		var nowAtLocation time.Time
		if e.Space.Location.Timezone == "" {
			nowAtLocation = nowAtOrg
		} else {
			nowAtLocation, _ = GetUTCNowInTimezone(e.Space.Location.Timezone)
		}
		if e.Leave.After(nowAtLocation) {
			res = append(res, e)
		}
	}
	return res, nil
}

// ─── Validation ──────────────────────────────────────────────────────────────

// CheckBooking validates a booking to be created (bookingID empty) or
// updated against all booking rules. in.Enter and in.Leave must already
// carry the location's time zone; an empty in.SpaceID skips the space-level
// checks. upcomingBookingsMarkup counts bookings about to be created in the
// same request (recurring bookings). It returns false and a BookingCode* on
// violation.
func (s *BookingService) CheckBooking(in *BookingInput, location *Location, requestUser *User, bookingID string, upcomingBookingsMarkup int) (bool, int) {
	if valid, code := s.isValidBookingRequest(in, location, requestUser, location.OrganizationID, bookingID, upcomingBookingsMarkup); !valid {
		return false, code
	}
	if !s.isValidConcurrent(in, location, bookingID) {
		return false, BookingCodeLocationMaxConcurrent
	}
	if !GetLocationService().IsLocationWeekdayBookable(location, requestUser, in.Enter, in.Leave) {
		return false, BookingCodeInvalidWeekday
	}
	return true, 0
}

func (s *BookingService) isValidBookingRequest(in *BookingInput, location *Location, user *User, orgID string, bookingID string, upcomingBookingsMarkup int) (bool, int) {
	if !IsValidBookingSubject(in.Subject) {
		return false, BookingCodeInvalidSubject
	}
	isUpdate := bookingID != ""
	if !s.IsValidBookingDuration(in.Enter, in.Leave, orgID, user) {
		return false, BookingCodeInvalidBookingDuration
	}
	valid, errorCode := s.IsValidBookingAdvance(in.Enter, in.Leave, orgID, user)
	if !valid {
		return false, errorCode
	}
	if !s.isValidMaxConcurrentBookingsForUser(orgID, user, in, bookingID) {
		return false, BookingCodeMaxConcurrentForUser
	}
	if !s.IsValidMinHoursBooking(in.Enter, in.Leave, orgID, user) {
		return false, BookingCodeInvalidMinBookingDuration
	}
	if !isUpdate {
		if !s.IsValidMaxUpcomingBookings(orgID, user, upcomingBookingsMarkup) {
			return false, BookingCodeTooManyUpcomingBookings
		}
	}
	if in.SpaceID == "" {
		return true, 0
	}

	// check allowed space and location bookers
	groupMemberships, _ := GetGroupRepository().GetAllWhereUserIsMember(user.ID)
	allowedSpaceBookers, _ := GetSpaceRepository().GetAllAllowedBookersForSpaceList([]string{in.SpaceID})
	if len(allowedSpaceBookers) > 0 {
		allowed := false
		for _, allowedBooker := range allowedSpaceBookers {
			for _, group := range groupMemberships {
				if group.ID == allowedBooker.GroupID {
					allowed = true
					break
				}
			}
		}
		if !allowed {
			return false, BookingCodeNotAllowedBooker
		}
	}
	allowedLocationBookers, _ := GetLocationRepository().GetAllAllowedBookersForLocation(location.ID)
	if len(allowedLocationBookers) > 0 {
		allowed := false
		for _, allowedBooker := range allowedLocationBookers {
			for _, group := range groupMemberships {
				if group.ID == allowedBooker.GroupID {
					allowed = true
					break
				}
			}
		}
		if !allowed {
			return false, BookingCodeNotAllowedBooker
		}
	}
	return true, 0
}

// IsValidBookingDuration checks the organization's maximum booking duration
// (and full-day granularity for daily-basis bookings). user may be nil.
func (s *BookingService) IsValidBookingDuration(enter, leave time.Time, orgID string, user *User) bool {
	if hasNoAdminRestrictions(user, orgID) {
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
		now := enter
		for now.Before(leave) {
			hoursOnDate := getHoursOnDate(&now)
			now = now.AddDate(0, 0, 1)
			correction += (hoursOnDate - 24)
		}
		durationNotRounded := int(math.Round(leave.Sub(enter).Minutes()) / 60)
		return ((durationNotRounded-correction)%24 == 0) && (durationNotRounded <= (maxDurationHours + correction))
	}

	// For non-daily-basis bookings, check exact duration
	duration := math.Floor(leave.Sub(enter).Minutes()) / 60
	if duration < 0 || duration > float64(maxDurationHours) {
		return false
	}
	return true
}

func getHoursOnDate(t *time.Time) int {
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	end := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
	durationNotRounded := int(math.Round(end.Sub(start).Minutes()) / 60)
	return durationNotRounded
}

// IsValidBookingAdvance checks that a booking is not in the past and not
// further ahead than the organization allows. user may be nil.
func (s *BookingService) IsValidBookingAdvance(enter, leave time.Time, orgID string, user *User) (bool, int) {
	noAdminRestrictions, _ := GetSettingsRepository().GetBool(orgID, SettingNoAdminRestrictions.Name)
	maxAdvanceDays, _ := GetSettingsRepository().GetInt(orgID, SettingMaxDaysInAdvance.Name)
	dailyBasisBooking, _ := GetSettingsRepository().GetBool(orgID, SettingDailyBasisBooking.Name)

	nowExact := time.Now().UTC()
	// booking already started today and hasn't ended yet -> always valid
	if enter.Before(nowExact) && enter.Year() == nowExact.Year() && enter.YearDay() == nowExact.YearDay() && leave.After(nowExact) {
		return true, 0
	}

	// allow Enter-Date in past if at least this morning
	now := time.Date(nowExact.Year(), nowExact.Month(), nowExact.Day(), 0, 0, 0, 0, nowExact.Location())
	if dailyBasisBooking {
		now = now.Add(-12 * time.Hour)
	}
	if leave.Before(now) { // Leave must not be in past
		return false, BookingCodeInPast
	}
	advanceDays := math.Floor(enter.Sub(now).Hours() / 24)
	if advanceDays >= 0 && noAdminRestrictions && HasPermission(user, orgID, PermissionBookings, PermissionLevelAdmin) {
		return true, 0
	}

	if advanceDays < 0 {
		return false, BookingCodeInPast
	}
	if advanceDays > float64(maxAdvanceDays) {
		return false, BookingCodeTooManyDaysInAdvance
	}
	return true, 0
}

// IsValidMaxUpcomingBookings checks the organization's maximum number of
// upcoming bookings per user, counting upcomingBookingsMarkup more.
func (s *BookingService) IsValidMaxUpcomingBookings(orgID string, user *User, upcomingBookingsMarkup int) bool {
	if hasNoAdminRestrictions(user, orgID) {
		return true
	}
	maxUpcoming, _ := GetSettingsRepository().GetInt(orgID, SettingMaxBookingsPerUser.Name)
	curUpcoming, _ := GetBookingRepository().GetAllByUser(user.ID, time.Now().UTC())
	return len(curUpcoming)+upcomingBookingsMarkup < maxUpcoming
}

func (s *BookingService) isValidMaxConcurrentBookingsForUser(orgID string, user *User, in *BookingInput, bookingID string) bool {
	if hasNoAdminRestrictions(user, orgID) {
		return true
	}
	maxConcurrent, _ := GetSettingsRepository().GetInt(orgID, SettingMaxConcurrentBookingsPerUser.Name)
	// 0 = no limit
	if maxConcurrent == 0 {
		return true
	}
	curAtTime, _ := GetBookingRepository().GetTimeRangeByUser(user.ID, in.Enter, in.Leave, bookingID)
	return len(curAtTime) < maxConcurrent
}

func (s *BookingService) isValidConcurrent(in *BookingInput, location *Location, bookingID string) bool {
	if location.MaxConcurrentBookings == 0 {
		return true
	}
	bookings, err := GetBookingRepository().GetConcurrent(location, in.Enter, in.Leave, bookingID)
	if err != nil {
		log.Println(err)
		return false
	}
	if bookings >= int(location.MaxConcurrentBookings) {
		return false
	}
	return true
}

// IsValidMinHoursBooking checks the organization's minimum booking duration.
// user may be nil.
func (s *BookingService) IsValidMinHoursBooking(enter, leave time.Time, organizationID string, user *User) bool {
	if hasNoAdminRestrictions(user, organizationID) {
		return true
	}
	min_hours, err := GetSettingsRepository().GetInt(organizationID, SettingMinBookingDurationHours.Name)
	if err != nil {
		log.Println(err)
		return false
	}
	if min_hours == 0 {
		return true
	}

	enterTime := enter
	leaveTime := leave

	// if daily based bookings is *NOT* enabled, we have to add 1s to the leave time
	dailyBasisBooking, err := GetSettingsRepository().GetBool(organizationID, SettingDailyBasisBooking.Name)
	if err != nil {
		log.Println(err)
		return false
	}
	if !dailyBasisBooking {
		leaveTime = leaveTime.Add(time.Second)
	}

	difference_in_hours := int64(leaveTime.Sub(enterTime).Hours())
	return difference_in_hours >= int64(min_hours)
}

// IsValidBookingHoursBeforeDelete checks the organization's minimum time
// between deleting a booking and its start.
func (s *BookingService) IsValidBookingHoursBeforeDelete(e *BookingDetails, user *User, organizationID string) bool {

	// test if user is admin and "no admin" restrictions is enabled
	noAdminRestrictions, err := GetSettingsRepository().GetBool(organizationID, SettingNoAdminRestrictions.Name)
	if err != nil {
		log.Println(err)
		return false
	}
	if noAdminRestrictions && HasPermission(user, organizationID, PermissionBookings, PermissionLevelAdmin) {
		return true
	}

	// test "max hour before delete" settings
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
	if max_hours == 0 {
		return true
	}

	// get the enter time in the location's time zone
	location, err := GetLocationRepository().GetOne(e.Space.Location.ID)
	if err != nil {
		log.Println(err)
		return false
	}
	enterTime, err := GetLocationRepository().AttachTimezoneInformation(e.Enter, location)
	if err != nil {
		log.Println(err)
		return false
	}

	now := time.Now()
	difference_in_hours := int64(enterTime.Sub(now).Hours()) // int64 rounds down
	return difference_in_hours >= int64(max_hours)
}
