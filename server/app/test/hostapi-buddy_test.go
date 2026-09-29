package test

import (
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/app"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestHostAPIBuddiesDisabled(t *testing.T) {
	org, user, _, space := setupHostBookingTest(t)
	h := NewHostAPI()
	buddyUser := CreateTestUserInOrg(org)
	e := &Buddy{OwnerID: user.ID, BuddyID: buddyUser.ID}
	if err := GetBuddyRepository().Create(e); err != nil {
		t.Fatal(err)
	}

	enter, leave := hostBookingDay(0, 8, 17)
	if _, bErr := h.CreateBookingForUser(buddyUser.ID, space.ID, enter, leave, ""); bErr != nil {
		t.Fatal(bErr)
	}

	// SettingShowNames is off by default, so the feature is disabled.
	buddies, err := h.GetBuddiesForUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 0, len(buddies))

	schedule, err := h.GetBuddyScheduleForUser(user.ID, nil, enter, leave)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 0, len(schedule))

	near, err := h.FindSpacesNearBuddiesForUser(user.ID, nil, enter, leave)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestString(t, "", near.LocationID)
	CheckTestInt(t, 0, len(near.Spaces))
}

func TestHostAPIBuddiesScheduleAndNearBuddies(t *testing.T) {
	org, user, location, space := setupHostBookingTest(t)
	GetSettingsRepository().Set(org.ID, SettingShowNames.Name, "1")
	h := NewHostAPI()

	buddyUser := CreateTestUserInOrg(org)
	e := &Buddy{OwnerID: user.ID, BuddyID: buddyUser.ID}
	if err := GetBuddyRepository().Create(e); err != nil {
		t.Fatal(err)
	}

	buddies, err := h.GetBuddiesForUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(buddies))
	CheckTestString(t, buddyUser.ID, buddies[0].BuddyUserID)
	CheckTestString(t, buddyUser.Email, buddies[0].Email)
	buddyID := buddies[0].ID

	nearSpace := &Space{LocationID: location.ID, Enabled: true, X: 10, Y: 0}
	if err := GetSpaceRepository().Create(nearSpace); err != nil {
		t.Fatal(err)
	}

	enter, leave := hostBookingDay(0, 8, 17)
	res, err := h.CreateBookingForUser(buddyUser.ID, space.ID, enter, leave, "")
	if err != nil {
		t.Fatal(err)
	}
	CheckTestBool(t, true, res.Approved)

	schedule, err := h.GetBuddyScheduleForUser(user.ID, []string{buddyID}, enter, leave)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(schedule))
	CheckTestString(t, space.ID, schedule[0].SpaceID)
	CheckTestString(t, location.ID, schedule[0].LocationID)

	near, err := h.FindSpacesNearBuddiesForUser(user.ID, []string{buddyID}, enter, leave)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestString(t, location.ID, near.LocationID)
	CheckTestInt(t, 1, len(near.FoundBuddyIDs))
	CheckTestString(t, buddyID, near.FoundBuddyIDs[0])
	CheckTestInt(t, 1, len(near.Spaces))
	CheckTestString(t, nearSpace.ID, near.Spaces[0].Space.ID)
}

func TestHostAPIBuddiesDisabledUser(t *testing.T) {
	_, user, _, _ := setupHostBookingTest(t)
	h := NewHostAPI()
	user.Disabled = true
	GetUserRepository().Update(user)

	enter, leave := hostBookingDay(0, 8, 17)
	if _, err := h.GetBuddiesForUser(user.ID); err == nil {
		t.Fatal("expected error for disabled user")
	}
	if _, err := h.GetBuddyScheduleForUser(user.ID, nil, enter, leave); err == nil {
		t.Fatal("expected error for disabled user")
	}
	if _, err := h.FindSpacesNearBuddiesForUser(user.ID, nil, enter, leave); err == nil {
		t.Fatal("expected error for disabled user")
	}
}
