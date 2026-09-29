package test

import (
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/service"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func makeTestBuddy(owner, buddy *User) *Buddy {
	e := &Buddy{OwnerID: owner.ID, BuddyID: buddy.ID}
	if err := GetBuddyRepository().Create(e); err != nil {
		panic(err)
	}
	return e
}

func createTestSpaceAt(location *Location, x, y uint) *Space {
	space := &Space{LocationID: location.ID, Enabled: true, X: x, Y: y}
	if err := GetSpaceRepository().Create(space); err != nil {
		panic(err)
	}
	return space
}

func TestBuddyServiceIsEnabled(t *testing.T) {
	org, _, _, _ := setupServiceTest(t)
	CheckTestBool(t, false, GetBuddyService().IsEnabled(org.ID))
	GetSettingsRepository().Set(org.ID, SettingShowNames.Name, "1")
	CheckTestBool(t, true, GetBuddyService().IsEnabled(org.ID))
	GetSettingsRepository().Set(org.ID, SettingDisableBuddies.Name, "1")
	CheckTestBool(t, false, GetBuddyService().IsEnabled(org.ID))
}

func TestBuddyServiceGetScheduleDisabled(t *testing.T) {
	org, user, _, space := setupServiceTest(t)
	buddyUser := CreateTestUserInOrg(org)
	makeTestBuddy(user, buddyUser)
	enter, leave := serviceTestSlot(0, 8, 17)
	if _, bErr := GetBookingService().CreateBooking(buddyUser, &BookingInput{SpaceID: space.ID, Enter: enter, Leave: leave}); bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}
	// SettingShowNames is off by default, so the feature is disabled.
	list, err := GetBuddyService().GetSchedule(user, nil, enter, leave)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 0, len(list))
}

func TestBuddyServiceGetSchedule(t *testing.T) {
	org, user, location, space := setupServiceTest(t)
	GetSettingsRepository().Set(org.ID, SettingShowNames.Name, "1")
	buddyUser := CreateTestUserInOrg(org)
	notABuddy := CreateTestUserInOrg(org)
	buddy := makeTestBuddy(user, buddyUser)
	otherSpace := createTestSpaceAt(location, 20, 0)

	enter, leave := serviceTestSlot(0, 8, 17)
	if _, bErr := GetBookingService().CreateBooking(buddyUser, &BookingInput{SpaceID: space.ID, Enter: enter, Leave: leave}); bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}
	// A booking outside the requested slot must not show up.
	otherEnter, otherLeave := serviceTestSlot(5, 8, 17)
	if _, bErr := GetBookingService().CreateBooking(buddyUser, &BookingInput{SpaceID: space.ID, Enter: otherEnter, Leave: otherLeave}); bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}
	// A non-buddy's booking must never show up, even if requested explicitly.
	if _, bErr := GetBookingService().CreateBooking(notABuddy, &BookingInput{SpaceID: otherSpace.ID, Enter: enter, Leave: leave}); bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}

	list, err := GetBuddyService().GetSchedule(user, []string{buddy.ID, notABuddy.ID}, enter, leave)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(list))
	CheckTestString(t, buddyUser.ID, list[0].Buddy.BuddyID)
	CheckTestString(t, space.ID, list[0].SpaceID)

	// Requesting all buddies (empty buddyIDs) returns the same result.
	list, err = GetBuddyService().GetSchedule(user, nil, enter, leave)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(list))
}

func TestBuddyServiceFindNearBuddiesNoneOnSite(t *testing.T) {
	org, user, _, _ := setupServiceTest(t)
	GetSettingsRepository().Set(org.ID, SettingShowNames.Name, "1")
	buddyUser := CreateTestUserInOrg(org)
	makeTestBuddy(user, buddyUser)
	enter, leave := serviceTestSlot(0, 8, 17)

	spaces, locationID, _, found, err := GetBuddyService().FindNearBuddies(user, nil, enter, leave)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 0, len(spaces))
	CheckTestString(t, "", locationID)
	CheckTestInt(t, 0, len(found))
}

func TestBuddyServiceFindNearBuddiesSingleLocation(t *testing.T) {
	org, user, location, buddySpace := setupServiceTest(t)
	GetSettingsRepository().Set(org.ID, SettingShowNames.Name, "1")
	buddyUser := CreateTestUserInOrg(org)
	buddy := makeTestBuddy(user, buddyUser)

	// buddySpace (from setupServiceTest) is at (0, 0); place the buddy there
	// and two more candidate spaces at increasing distance.
	near := createTestSpaceAt(location, 10, 0)
	far := createTestSpaceAt(location, 100, 0)

	enter, leave := serviceTestSlot(0, 8, 17)
	if _, bErr := GetBookingService().CreateBooking(buddyUser, &BookingInput{SpaceID: buddySpace.ID, Enter: enter, Leave: leave}); bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}

	spaces, locationID, locationName, found, err := GetBuddyService().FindNearBuddies(user, []string{buddy.ID}, enter, leave)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestString(t, location.ID, locationID)
	CheckTestString(t, location.Name, locationName)
	CheckTestInt(t, 1, len(found))
	CheckTestString(t, buddy.ID, found[0])
	// The buddy's own occupied space must not be offered back.
	CheckTestInt(t, 2, len(spaces))
	CheckTestString(t, near.ID, spaces[0].Space.ID)
	CheckTestString(t, far.ID, spaces[1].Space.ID)
	if spaces[0].Distance >= spaces[1].Distance {
		t.Fatalf("expected %s to be closer than %s, got distances %v and %v", near.ID, far.ID, spaces[0].Distance, spaces[1].Distance)
	}
}

func TestBuddyServiceFindNearBuddiesMultipleLocations(t *testing.T) {
	org, user, location1, space1 := setupServiceTest(t)
	GetSettingsRepository().Set(org.ID, SettingShowNames.Name, "1")
	location2, space2 := CreateTestLocationAndSpace(org)

	buddyA := CreateTestUserInOrg(org)
	buddyB := CreateTestUserInOrg(org)
	buddyC := CreateTestUserInOrg(org)
	bA := makeTestBuddy(user, buddyA)
	bB := makeTestBuddy(user, buddyB)
	bC := makeTestBuddy(user, buddyC)

	enter, leave := serviceTestSlot(0, 8, 17)
	// location1 holds two of the requested buddies, location2 only one.
	if _, bErr := GetBookingService().CreateBooking(buddyA, &BookingInput{SpaceID: space1.ID, Enter: enter, Leave: leave}); bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}
	space1b := createTestSpaceAt(location1, 50, 0)
	if _, bErr := GetBookingService().CreateBooking(buddyB, &BookingInput{SpaceID: space1b.ID, Enter: enter, Leave: leave}); bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}
	if _, bErr := GetBookingService().CreateBooking(buddyC, &BookingInput{SpaceID: space2.ID, Enter: enter, Leave: leave}); bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}

	_, locationID, _, found, err := GetBuddyService().FindNearBuddies(user, []string{bA.ID, bB.ID, bC.ID}, enter, leave)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestString(t, location1.ID, locationID)
	CheckTestInt(t, 2, len(found))
	if location2.ID == locationID {
		t.Fatalf("expected the majority location %s, got %s", location1.ID, locationID)
	}
}
