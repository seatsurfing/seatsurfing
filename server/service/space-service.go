package service

import (
	"sync"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
)

// SpaceService holds the space business logic shared by the REST API
// (package router) and the plugin host API.
type SpaceService struct{}

var spaceService *SpaceService
var spaceServiceOnce sync.Once

func GetSpaceService() *SpaceService {
	spaceServiceOnce.Do(func() {
		spaceService = &SpaceService{}
	})
	return spaceService
}

// UserSpaceAvailability is a space's availability as seen by a specific user.
type UserSpaceAvailability struct {
	Space     Space
	Available bool
	// Allowed reports whether the user may book the space: they are an
	// allowed booker of the space and its location, and the time slot falls
	// on the location's bookable weekdays.
	Allowed          bool
	ApprovalRequired bool
	Attributes       []*SpaceAttributeValue
	// Bookings overlapping the time slot. The booker is only included for
	// the user's own bookings, or if the user may see other bookers' names.
	Bookings []*UserSpaceAvailabilityBooking
}

type UserSpaceAvailabilityBooking struct {
	BookingID     string
	RecurringID   string
	UserID        string
	UserEmail     string
	UserFirstname string
	UserLastname  string
	// Enter and Leave carry the location's time zone.
	Enter    time.Time
	Leave    time.Time
	Subject  string
	Approved bool
}

// GetAvailabilityForUser returns the spaces of location (or only spaceID, if
// non-empty) matching attributes, with their availability between enter and
// leave and whether user may book them. enter and leave must already carry
// the location's time zone. The caller must have verified that user may
// access the location's organization.
func (s *SpaceService) GetAvailabilityForUser(user *User, location *Location, spaceID string, enter, leave time.Time, attributes []SearchAttribute) ([]*UserSpaceAvailability, error) {
	showNames := false
	if HasPermission(user, location.OrganizationID, PermissionBookings, PermissionLevelRead) {
		showNames = true
	} else {
		showNames, _ = GetSettingsRepository().GetBool(location.OrganizationID, SettingShowNames.Name)
	}
	list, err := GetSpaceRepository().GetAllInTime(location.ID, enter, leave)
	if err != nil {
		return nil, err
	}
	spaceIds := []string{}
	for _, e := range list {
		spaceIds = append(spaceIds, e.Space.ID)
	}
	attributeValues, err := GetSpaceAttributeValueRepository().GetAllForEntityList(spaceIds, SpaceAttributeValueEntityTypeSpace)
	if err != nil {
		return nil, err
	}
	userGroups, err := GetGroupRepository().GetAllWhereUserIsMember(user.ID)
	if err != nil {
		return nil, err
	}
	spaceAllowedBookers, err := GetSpaceRepository().GetAllAllowedBookersForSpaceList(spaceIds)
	if err != nil {
		return nil, err
	}
	locationAllowedBookers, err := GetLocationRepository().GetAllAllowedBookersForLocation(location.ID)
	if err != nil {
		return nil, err
	}
	approvers, err := GetSpaceRepository().GetAllApproversForSpaceList(spaceIds)
	if err != nil {
		return nil, err
	}
	// Approver groups only take effect while the groups feature is enabled
	// (see RequiresApproval), so the hint shown to the user must match.
	if groupsEnabled, _ := GetSettingsRepository().GetBool(location.OrganizationID, SettingFeatureGroups.Name); !groupsEnabled {
		approvers = nil
	}
	locations := GetLocationService()
	isAllowedToBookLocation := locations.IsUserAllowedToBookLocation(locationAllowedBookers, userGroups)
	isValidWeekday := locations.IsLocationWeekdayBookable(location, user, enter, leave)
	res := []*UserSpaceAvailability{}
	for _, e := range list {
		if spaceID != "" && e.ID != spaceID {
			continue
		}
		if !MatchesSearchAttributes(e.ID, &attributes, attributeValues) {
			continue
		}
		item := &UserSpaceAvailability{
			Space:            e.Space,
			Available:        e.Available,
			Allowed:          isAllowedToBookLocation && s.IsUserAllowedToBookSpace(&e.Space, spaceAllowedBookers, userGroups) && isValidWeekday,
			ApprovalRequired: s.IsApprovalRequired(&e.Space, approvers),
			Attributes:       []*SpaceAttributeValue{},
			Bookings:         []*UserSpaceAvailabilityBooking{},
		}
		for _, v := range attributeValues {
			if v.EntityType == SpaceAttributeValueEntityTypeSpace && v.EntityID == e.ID {
				item.Attributes = append(item.Attributes, v)
			}
		}
		for _, booking := range e.Bookings {
			bookingEnter, _ := GetLocationRepository().AttachTimezoneInformation(booking.Enter, location)
			bookingLeave, _ := GetLocationRepository().AttachTimezoneInformation(booking.Leave, location)
			entry := &UserSpaceAvailabilityBooking{
				BookingID:   booking.BookingID,
				RecurringID: booking.RecurringID,
				Enter:       bookingEnter,
				Leave:       bookingLeave,
				Subject:     booking.Subject,
				Approved:    booking.Approved,
			}
			if showNames || user.Email == booking.UserEmail {
				entry.UserID = booking.UserID
				entry.UserEmail = booking.UserEmail
				entry.UserFirstname = booking.UserFirstname
				entry.UserLastname = booking.UserLastname
			}
			item.Bookings = append(item.Bookings, entry)
		}
		res = append(res, item)
	}
	return res, nil
}

// GetPublicAvailability reports for each public-bookable space of the
// organization whether it can still be booked between enter and leave. enter
// and leave are wall-clock times which are interpreted in the time zone of
// each space's location. A space is unavailable if it is already booked, its
// location is not bookable on that weekday, or the location's concurrent
// booking limit is reached.
func (s *SpaceService) GetPublicAvailability(organizationID string, enter, leave time.Time) (map[string]bool, error) {
	spaces, err := GetSpaceRepository().GetAllPublicBookable(organizationID)
	if err != nil {
		return nil, err
	}
	spacesByLocation := map[string][]*Space{}
	for _, space := range spaces {
		spacesByLocation[space.LocationID] = append(spacesByLocation[space.LocationID], space)
	}
	res := map[string]bool{}
	for locationID, locationSpaces := range spacesByLocation {
		location, err := GetLocationRepository().GetOne(locationID)
		if err != nil {
			return nil, err
		}
		available, booked, err := s.getPublicLocationAvailability(location, enter, leave)
		if err != nil {
			return nil, err
		}
		for _, space := range locationSpaces {
			res[space.ID] = available && !booked[space.ID]
		}
	}
	return res, nil
}

// IsPublicSlotAvailable reports whether space in location can still be booked
// publicly between enter and leave, which must already carry the location's
// time zone. The slot is unavailable if the location is not bookable on that
// weekday, the location's concurrent booking limit is reached, or the space
// already has an overlapping booking.
func (s *SpaceService) IsPublicSlotAvailable(space *Space, location *Location, enter, leave time.Time) (bool, error) {
	available, err := s.isPublicLocationSlotAvailable(location, enter, leave)
	if err != nil || !available {
		return false, err
	}
	conflicts, err := GetBookingRepository().GetConflicts(space.ID, enter, leave, "")
	if err != nil {
		return false, err
	}
	return len(conflicts) == 0, nil
}

func (s *SpaceService) getPublicLocationAvailability(location *Location, enter, leave time.Time) (bool, map[string]bool, error) {
	booked := map[string]bool{}
	enter, err := GetLocationRepository().AttachTimezoneInformation(enter, location)
	if err != nil {
		return false, booked, err
	}
	leave, err = GetLocationRepository().AttachTimezoneInformation(leave, location)
	if err != nil {
		return false, booked, err
	}
	available, err := s.isPublicLocationSlotAvailable(location, enter, leave)
	if err != nil || !available {
		return false, booked, err
	}
	spaceIDs, err := GetBookingRepository().GetBookedSpaceIDs(location.OrganizationID, location.ID, enter, leave)
	if err != nil {
		return false, booked, err
	}
	for _, id := range spaceIDs {
		booked[id] = true
	}
	return true, booked, nil
}

// isPublicLocationSlotAvailable checks the location-wide conditions of a
// public booking slot: bookable weekday and concurrent booking limit.
func (s *SpaceService) isPublicLocationSlotAvailable(location *Location, enter, leave time.Time) (bool, error) {
	if !GetLocationService().IsLocationWeekdayBookable(location, nil, enter, leave) {
		return false, nil
	}
	if location.MaxConcurrentBookings > 0 {
		concurrent, err := GetBookingRepository().GetConcurrent(location, enter, leave, "")
		if err != nil {
			return false, err
		}
		if concurrent >= int(location.MaxConcurrentBookings) {
			return false, nil
		}
	}
	return true, nil
}

// IsApprovalRequired reports whether any of approvers belongs to space e.
func (s *SpaceService) IsApprovalRequired(e *Space, approvers []*SpaceGroup) bool {
	for _, approver := range approvers {
		if approver.SpaceID == e.ID {
			return true
		}
	}
	return false
}

// RequiresApproval reports whether new bookings of space e need approval.
func (s *SpaceService) RequiresApproval(orgID string, e *Space) bool {
	groupsEnabled, _ := GetSettingsRepository().GetBool(orgID, SettingFeatureGroups.Name)
	if !groupsEnabled {
		return false
	}
	approvers, _ := GetSpaceRepository().GetApproverGroupIDs(e.ID)
	return len(approvers) > 0
}

// IsUserAllowedToBookSpace reports whether a member of userGroups may book
// space e with allowedBookers (no allowed bookers for e: everyone may).
func (s *SpaceService) IsUserAllowedToBookSpace(e *Space, allowedBookers []*SpaceGroup, userGroups []*Group) bool {
	restricted := false
	for _, allowedBooker := range allowedBookers {
		if allowedBooker.SpaceID == e.ID {
			restricted = true
			for _, userGroup := range userGroups {
				if allowedBooker.GroupID == userGroup.ID {
					return true
				}
			}
		}
	}
	return !restricted
}
