package test

import (
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/service"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestSpaceServiceAvailabilityRedactsOtherBookers(t *testing.T) {
	org, user, location, space := setupServiceTest(t)
	other := CreateTestUserInOrg(org)
	space2 := &Space{LocationID: location.ID, Enabled: true}
	GetSpaceRepository().Create(space2)
	enter, leave := serviceTestSlot(0, 8, 17)
	if _, bErr := GetBookingService().CreateBooking(other, &BookingInput{SpaceID: space.ID, Enter: enter, Leave: leave}); bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}
	if _, bErr := GetBookingService().CreateBooking(user, &BookingInput{SpaceID: space2.ID, Enter: enter, Leave: leave}); bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}
	location, _ = GetLocationRepository().GetOne(location.ID)
	enter, _ = GetLocationRepository().AttachTimezoneInformation(enter, location)
	leave, _ = GetLocationRepository().AttachTimezoneInformation(leave, location)

	list, err := GetSpaceService().GetAvailabilityForUser(user, location, "", enter, leave, nil)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 2, len(list))
	for _, s := range list {
		CheckTestBool(t, false, s.Available)
		CheckTestInt(t, 1, len(s.Bookings))
		if s.Space.ID == space.ID {
			CheckTestString(t, "", s.Bookings[0].UserEmail)
		} else {
			CheckTestString(t, user.Email, s.Bookings[0].UserEmail)
		}
	}

	GetSettingsRepository().Set(org.ID, SettingShowNames.Name, "1")
	list, _ = GetSpaceService().GetAvailabilityForUser(user, location, space.ID, enter, leave, nil)
	CheckTestInt(t, 1, len(list))
	CheckTestString(t, other.Email, list[0].Bookings[0].UserEmail)
}

func TestSpaceServiceRequiresApproval(t *testing.T) {
	org, user, _, space := setupServiceTest(t)
	group := CreateTestGroup(org, user)
	GetSpaceRepository().AddApprovers(space, []string{group.ID})
	CheckTestBool(t, false, GetSpaceService().RequiresApproval(org.ID, space))
	GetSettingsRepository().Set(org.ID, SettingFeatureGroups.Name, "1")
	CheckTestBool(t, true, GetSpaceService().RequiresApproval(org.ID, space))
}
