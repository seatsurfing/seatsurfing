package app

import (
	"time"

	"github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	"github.com/seatsurfing/seatsurfing/server/service"
)

// User-scoped buddy operations of the host API. They are implemented by
// service.BuddyService, which the REST handlers use as well, so a plugin
// acting on behalf of a user gets exactly that user's permissions and
// respects the same feature settings (SettingShowNames, SettingDisableBuddies).

func buddyDetailsToAPI(e *BuddyDetails) *api.BuddyInfo {
	return &api.BuddyInfo{
		ID:          e.ID,
		BuddyUserID: e.BuddyID,
		Email:       e.BuddyEmail,
		Firstname:   e.BuddyFirstname,
		Lastname:    e.BuddyLastname,
	}
}

func buddyScheduleEntryToAPI(e *service.BuddyScheduleEntry) *api.BuddyScheduleInfo {
	return &api.BuddyScheduleInfo{
		Buddy:        *buddyDetailsToAPI(&e.Buddy),
		LocationID:   e.LocationID,
		LocationName: e.LocationName,
		SpaceID:      e.SpaceID,
		SpaceName:    e.SpaceName,
		Enter:        e.Enter,
		Leave:        e.Leave,
	}
}

func (h *hostAPIImpl) GetBuddiesForUser(userID string) ([]*api.BuddyInfo, error) {
	user, err := getActiveUser(userID)
	if err != nil {
		return nil, err
	}
	list, err := service.GetBuddyService().GetBuddies(user)
	if err != nil {
		return nil, err
	}
	res := make([]*api.BuddyInfo, 0, len(list))
	for _, e := range list {
		res = append(res, buddyDetailsToAPI(e))
	}
	return res, nil
}

func (h *hostAPIImpl) GetBuddyScheduleForUser(userID string, buddyIDs []string, enter, leave time.Time) ([]*api.BuddyScheduleInfo, error) {
	user, err := getActiveUser(userID)
	if err != nil {
		return nil, err
	}
	list, err := service.GetBuddyService().GetSchedule(user, buddyIDs, enter, leave)
	if err != nil {
		return nil, err
	}
	res := make([]*api.BuddyScheduleInfo, 0, len(list))
	for _, e := range list {
		res = append(res, buddyScheduleEntryToAPI(e))
	}
	return res, nil
}

func (h *hostAPIImpl) FindSpacesNearBuddiesForUser(userID string, buddyIDs []string, enter, leave time.Time) (*api.NearBuddiesResult, error) {
	user, err := getActiveUser(userID)
	if err != nil {
		return nil, err
	}
	spaces, locationID, locationName, foundBuddyIDs, err := service.GetBuddyService().FindNearBuddies(user, buddyIDs, enter, leave)
	if err != nil {
		return nil, err
	}
	res := &api.NearBuddiesResult{
		LocationID:    locationID,
		LocationName:  locationName,
		Spaces:        make([]*api.NearBuddySpaceInfo, 0, len(spaces)),
		FoundBuddyIDs: foundBuddyIDs,
	}
	for _, s := range spaces {
		res.Spaces = append(res.Spaces, &api.NearBuddySpaceInfo{
			Space:            s.Space,
			Available:        s.Available,
			Allowed:          s.Allowed,
			ApprovalRequired: s.ApprovalRequired,
			Distance:         s.Distance,
			NearBuddyIDs:     s.NearBuddyIDs,
		})
	}
	return res, nil
}
