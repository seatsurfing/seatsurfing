package service

import (
	"encoding/json"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
)

// LocationService holds the location business logic shared by the REST API
// (package router) and the plugin host API.
type LocationService struct{}

var locationService *LocationService
var locationServiceOnce sync.Once

func GetLocationService() *LocationService {
	locationServiceOnce.Do(func() {
		locationService = &LocationService{}
	})
	return locationService
}

// UserLocation is a location as seen by a specific user.
type UserLocation struct {
	Location *Location
	// AllowedBookers are the location's allowed booker groups; none means
	// everyone may book.
	AllowedBookers []*LocationGroup
	// Attributes are the location's attribute values (without the synthetic
	// search attributes).
	Attributes []*SpaceAttributeValue
	// AllowedForUser reports whether the user is in AllowedBookers (or there
	// are none).
	AllowedForUser bool
}

// SearchLocationsForUser returns the locations of user's organization that
// match attributes (all of them if there are none), including the synthetic
// numSpaces, numFreeSpaces and buddyOnSite attributes evaluated for the time
// between enter and leave.
func (s *LocationService) SearchLocationsForUser(user *User, enter, leave time.Time, attributes []SearchAttribute) ([]*UserLocation, error) {
	list, err := GetLocationRepository().GetAll(user.OrganizationID)
	if err != nil {
		return nil, err
	}
	attributeValues, err := GetSpaceAttributeValueRepository().GetAll(user.OrganizationID, SpaceAttributeValueEntityTypeLocation)
	if err != nil {
		return nil, err
	}
	locationAttributeValues := attributeValues
	if searchInputContains(attributes, SearchAttributeNumSpaces) {
		attributeValues, err = s.attachNumSpaces(attributeValues, user.OrganizationID)
		if err != nil {
			return nil, err
		}
	}
	if searchInputContains(attributes, SearchAttributeNumFreeSpaces) {
		attributeValues, err = s.attachNumFreeSpaces(attributeValues, user.OrganizationID, enter, leave)
		if err != nil {
			return nil, err
		}
	}
	if searchInputContains(attributes, SearchAttributeBuddyOnSite) {
		attributeValues, err = s.attachBuddiesOnSite(attributeValues, user, enter, leave)
		if err != nil {
			return nil, err
		}
	}

	locationIDs := []string{}
	for _, e := range list {
		locationIDs = append(locationIDs, e.ID)
	}
	allowedBookers, err := GetLocationRepository().GetAllAllowedBookersForLocationList(locationIDs)
	if err != nil {
		return nil, err
	}
	userGroups, err := GetGroupRepository().GetAllWhereUserIsMember(user.ID)
	if err != nil {
		return nil, err
	}

	res := []*UserLocation{}
	for _, e := range list {
		if !MatchesSearchAttributes(e.ID, &attributes, attributeValues) {
			continue
		}
		item := &UserLocation{Location: e, AllowedBookers: []*LocationGroup{}, Attributes: []*SpaceAttributeValue{}}
		for _, ab := range allowedBookers {
			if ab.LocationID == e.ID {
				item.AllowedBookers = append(item.AllowedBookers, ab)
			}
		}
		for _, v := range locationAttributeValues {
			if v.EntityID == e.ID {
				item.Attributes = append(item.Attributes, v)
			}
		}
		item.AllowedForUser = s.IsUserAllowedToBookLocation(item.AllowedBookers, userGroups)
		res = append(res, item)
	}
	return res, nil
}

// IsUserAllowedToBookLocation reports whether a member of userGroups may book
// at a location with allowedBookers (no allowed bookers: everyone may).
func (s *LocationService) IsUserAllowedToBookLocation(allowedBookers []*LocationGroup, userGroups []*Group) bool {
	restricted := false
	for _, allowedBooker := range allowedBookers {
		restricted = true
		for _, userGroup := range userGroups {
			if allowedBooker.GroupID == userGroup.ID {
				return true
			}
		}
	}
	return !restricted
}

// IsLocationWeekdayBookable checks whether every calendar day in [enter, leave)
// falls on one of the location's bookable weekdays, honoring the org's
// no-admin-restrictions setting for those who manage other people's bookings.
func (s *LocationService) IsLocationWeekdayBookable(location *Location, user *User, enter, leave time.Time) bool {
	if location.BookableDays == "" {
		return true
	}
	if hasNoAdminRestrictions(user, location.OrganizationID) {
		return true
	}
	allowedDays := map[time.Weekday]bool{}
	for _, s := range strings.Split(location.BookableDays, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			continue
		}
		allowedDays[time.Weekday(n)] = true
	}
	day := time.Date(enter.Year(), enter.Month(), enter.Day(), 0, 0, 0, 0, enter.Location())
	// leave is exclusive: the last day to check is the calendar day just before leave,
	// so a leave of exactly midnight does not pull in the following day.
	lastInstant := leave.Add(-time.Nanosecond)
	lastDay := time.Date(lastInstant.Year(), lastInstant.Month(), lastInstant.Day(), 0, 0, 0, 0, lastInstant.Location())
	for !day.After(lastDay) {
		if !allowedDays[day.Weekday()] {
			return false
		}
		day = day.AddDate(0, 0, 1)
	}
	return true
}

// WeekdaysFromString parses a location's comma-separated bookable weekdays.
func WeekdaysFromString(csv string) []int {
	days := []int{}
	if csv == "" {
		return days
	}
	for _, part := range strings.Split(csv, ",") {
		d, err := strconv.Atoi(part)
		if err == nil {
			days = append(days, d)
		}
	}
	return days
}

func searchInputContains(m []SearchAttribute, attributeID string) bool {
	for _, e := range m {
		if e.AttributeID == attributeID {
			return true
		}
	}
	return false
}

func (s *LocationService) attachNumSpaces(attributeValues []*SpaceAttributeValue, organizationID string) ([]*SpaceAttributeValue, error) {
	totalSpaces, err := GetSpaceRepository().GetTotalCountMap(organizationID)
	if err != nil {
		return nil, err
	}
	for k, v := range totalSpaces {
		attributeValues = append(attributeValues, &SpaceAttributeValue{
			AttributeID: SearchAttributeNumSpaces,
			EntityID:    k,
			EntityType:  SpaceAttributeValueEntityTypeLocation,
			Value:       strconv.Itoa(v),
		})
	}
	return attributeValues, nil
}

func (s *LocationService) attachNumFreeSpaces(attributeValues []*SpaceAttributeValue, organizationID string, enter, leave time.Time) ([]*SpaceAttributeValue, error) {
	freeSpaces, err := GetSpaceRepository().GetFreeCountMap(organizationID, enter, leave)
	if err != nil {
		return nil, err
	}
	for k, v := range freeSpaces {
		attributeValues = append(attributeValues, &SpaceAttributeValue{
			AttributeID: SearchAttributeNumFreeSpaces,
			EntityID:    k,
			EntityType:  SpaceAttributeValueEntityTypeLocation,
			Value:       strconv.Itoa(v),
		})
	}
	return attributeValues, nil
}

func (s *LocationService) attachBuddiesOnSite(attributeValues []*SpaceAttributeValue, user *User, enter, leave time.Time) ([]*SpaceAttributeValue, error) {
	buddies, err := GetBuddyRepository().GetAllByOwner(user.ID)
	if err != nil {
		return nil, err
	}
	usersOnSite, err := GetSpaceRepository().GetBookingUserIDMap(user.OrganizationID, enter, leave)
	if err != nil {
		return nil, err
	}
	buddiesOnSite := make(map[string][]string)
	for locationID, userIDs := range usersOnSite {
		buddiesOnSite[locationID] = []string{}
		for _, buddy := range buddies {
			if slices.Contains(userIDs, buddy.BuddyID) {
				buddiesOnSite[locationID] = append(buddiesOnSite[locationID], buddy.ID)
			}
		}
	}
	for k, v := range buddiesOnSite {
		json, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		attributeValues = append(attributeValues, &SpaceAttributeValue{
			AttributeID: SearchAttributeBuddyOnSite,
			EntityID:    k,
			EntityType:  SpaceAttributeValueEntityTypeLocation,
			Value:       string(json),
		})
	}
	return attributeValues, nil
}
