package service

import (
	"math"
	"sort"
	"sync"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
)

// BuddyService holds the buddy business logic shared by the REST API
// (package router) and the plugin host API.
type BuddyService struct{}

var buddyService *BuddyService
var buddyServiceOnce sync.Once

func GetBuddyService() *BuddyService {
	buddyServiceOnce.Do(func() {
		buddyService = &BuddyService{}
	})
	return buddyService
}

// IsEnabled reports whether the buddies feature is available for an
// organization: it requires names to be shown (buddies would otherwise be
// anonymous) and must not be explicitly disabled.
func (s *BuddyService) IsEnabled(organizationID string) bool {
	showNames, _ := GetSettingsRepository().GetBool(organizationID, SettingShowNames.Name)
	if !showNames {
		return false
	}
	disableBuddies, _ := GetSettingsRepository().GetBool(organizationID, SettingDisableBuddies.Name)
	return !disableBuddies
}

// GetBuddies returns user's buddies, or an empty list if the feature is
// disabled for their organization.
func (s *BuddyService) GetBuddies(user *User) ([]*BuddyDetails, error) {
	if !s.IsEnabled(user.OrganizationID) {
		return []*BuddyDetails{}, nil
	}
	return GetBuddyRepository().GetAllByOwner(user.ID)
}

// BuddyScheduleEntry is where one of user's buddies is booked in a time slot.
type BuddyScheduleEntry struct {
	Buddy        BuddyDetails
	LocationID   string
	LocationName string
	SpaceID      string
	SpaceName    string
	SpaceX       uint
	SpaceY       uint
	// Enter and Leave carry the booking's location's time zone.
	Enter time.Time
	Leave time.Time
}

// GetSchedule returns where the buddies in buddyIDs (or all of user's
// buddies, if buddyIDs is empty) are booked overlapping [enter, leave). A
// buddy with no such booking is simply absent from the result; an ID that is
// not one of user's buddies is ignored. Returns an empty slice if the
// buddies feature is disabled for user's organization.
func (s *BuddyService) GetSchedule(user *User, buddyIDs []string, enter, leave time.Time) ([]*BuddyScheduleEntry, error) {
	buddies, err := s.GetBuddies(user)
	if err != nil {
		return nil, err
	}
	wanted := map[string]bool{}
	for _, id := range buddyIDs {
		wanted[id] = true
	}
	startOfDay := time.Date(enter.Year(), enter.Month(), enter.Day(), 0, 0, 0, 0, enter.Location())
	res := []*BuddyScheduleEntry{}
	for _, buddy := range buddies {
		if len(wanted) > 0 && !wanted[buddy.ID] {
			continue
		}
		bookings, err := GetBookingRepository().GetAllByUser(buddy.BuddyID, startOfDay)
		if err != nil {
			return nil, err
		}
		for _, b := range bookings {
			if !b.Enter.Before(leave) || !b.Leave.After(enter) {
				continue
			}
			res = append(res, &BuddyScheduleEntry{
				Buddy:        *buddy,
				LocationID:   b.Space.Location.ID,
				LocationName: b.Space.Location.Name,
				SpaceID:      b.Space.ID,
				SpaceName:    b.Space.Name,
				SpaceX:       b.Space.X,
				SpaceY:       b.Space.Y,
				Enter:        b.Enter,
				Leave:        b.Leave,
			})
		}
	}
	return res, nil
}

// NearBuddySpace is a space available to a user, ranked by proximity to
// where their buddies are seated.
type NearBuddySpace struct {
	Space            Space
	Available        bool
	Allowed          bool
	ApprovalRequired bool
	// Distance is the (average, if several) Euclidean distance in floor-plan
	// units to the buddies seated in this space's location.
	Distance     float64
	NearBuddyIDs []string
}

// FindNearBuddies picks the location holding the most of the buddies in
// buddyIDs (or all of user's buddies, if buddyIDs is empty) for
// [enter, leave), then returns the spaces available to user there, sorted
// by ascending distance to those buddies' seats. locationID/locationName
// identify the chosen location, and foundBuddyIDs are the buddy IDs (from
// BuddyDetails.ID) actually seated there - a caller can diff this against
// the requested buddyIDs to report the ones not found. All are zero values
// if none of the requested buddies have a booking in the time slot.
func (s *BuddyService) FindNearBuddies(user *User, buddyIDs []string, enter, leave time.Time) (spaces []*NearBuddySpace, locationID, locationName string, foundBuddyIDs []string, err error) {
	entries, err := s.GetSchedule(user, buddyIDs, enter, leave)
	if err != nil {
		return nil, "", "", nil, err
	}
	if len(entries) == 0 {
		return nil, "", "", nil, nil
	}

	byLocation := map[string][]*BuddyScheduleEntry{}
	for _, e := range entries {
		byLocation[e.LocationID] = append(byLocation[e.LocationID], e)
	}
	locationIDs := make([]string, 0, len(byLocation))
	for id := range byLocation {
		locationIDs = append(locationIDs, id)
	}
	sort.Slice(locationIDs, func(i, j int) bool {
		return byLocation[locationIDs[i]][0].LocationName < byLocation[locationIDs[j]][0].LocationName
	})
	best := locationIDs[0]
	for _, id := range locationIDs[1:] {
		if distinctBuddyCount(byLocation[id]) > distinctBuddyCount(byLocation[best]) {
			best = id
		}
	}
	winning := byLocation[best]
	locationID = best
	locationName = winning[0].LocationName
	foundBuddyIDs = distinctBuddyIDs(winning)

	location, err := GetLocationRepository().GetOne(locationID)
	if err != nil || location == nil {
		return nil, "", "", nil, err
	}
	list, err := GetSpaceService().GetAvailabilityForUser(user, location, "", enter, leave, nil)
	if err != nil {
		return nil, "", "", nil, err
	}

	occupiedByBuddy := map[string]bool{}
	for _, e := range winning {
		occupiedByBuddy[e.SpaceID] = true
	}
	res := []*NearBuddySpace{}
	for _, e := range list {
		if occupiedByBuddy[e.Space.ID] {
			continue
		}
		sum := 0.0
		for _, buddy := range winning {
			dx := float64(e.Space.X) - float64(buddy.SpaceX)
			dy := float64(e.Space.Y) - float64(buddy.SpaceY)
			sum += math.Hypot(dx, dy)
		}
		res = append(res, &NearBuddySpace{
			Space:            e.Space,
			Available:        e.Available,
			Allowed:          e.Allowed,
			ApprovalRequired: e.ApprovalRequired,
			Distance:         sum / float64(len(winning)),
			NearBuddyIDs:     foundBuddyIDs,
		})
	}
	sort.Slice(res, func(i, j int) bool { return res[i].Distance < res[j].Distance })
	return res, locationID, locationName, foundBuddyIDs, nil
}

func distinctBuddyIDs(entries []*BuddyScheduleEntry) []string {
	seen := map[string]bool{}
	res := []string{}
	for _, e := range entries {
		if !seen[e.Buddy.ID] {
			seen[e.Buddy.ID] = true
			res = append(res, e.Buddy.ID)
		}
	}
	return res
}

func distinctBuddyCount(entries []*BuddyScheduleEntry) int {
	return len(distinctBuddyIDs(entries))
}
