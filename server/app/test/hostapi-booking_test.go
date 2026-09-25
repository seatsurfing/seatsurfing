package test

import (
	"net/http"
	"testing"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/app"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

// hostBookingDay returns wall-clock times on a Monday far enough in the
// future to be bookable, built in UTC as the host API expects.
func hostBookingDay(dayOffset, fromHour, toHour int) (time.Time, time.Time) {
	base := time.Date(2030, 9, 2+dayOffset, 0, 0, 0, 0, time.UTC)
	return base.Add(time.Duration(fromHour) * time.Hour), base.Add(time.Duration(toHour) * time.Hour)
}

func setupHostBookingTest(t *testing.T) (*Organization, *User, *Location, *Space) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org.ID, SettingMaxDaysInAdvance.Name, "5000")
	user := CreateTestUserInOrg(org)
	location, space := CreateTestLocationAndSpace(org)
	return org, user, location, space
}

func TestHostAPICreateAndListBookings(t *testing.T) {
	_, user, _, space := setupHostBookingTest(t)
	h := NewHostAPI()

	enter, leave := hostBookingDay(0, 8, 17)
	res, err := h.CreateBookingForUser(user.ID, space.ID, enter, leave, "Focus time")
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, http.StatusCreated, res.StatusCode)
	CheckStringNotEmpty(t, res.BookingID)
	CheckTestBool(t, true, res.Approved)

	// The same slot again conflicts.
	res, err = h.CreateBookingForUser(user.ID, space.ID, enter, leave, "")
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, http.StatusConflict, res.StatusCode)
	CheckTestInt(t, ResponseCodeBookingSlotConflict, res.ErrorCode)

	list, err := h.GetUpcomingBookingsForUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(list))
	CheckTestString(t, space.ID, list[0].SpaceID)
	CheckTestString(t, "Focus time", list[0].Subject)
	CheckTestString(t, "08:00", list[0].Enter.Format("15:04"))
}

func TestHostAPICreateBookingRespectsLimits(t *testing.T) {
	org, user, location, space := setupHostBookingTest(t)
	h := NewHostAPI()
	GetSettingsRepository().Set(org.ID, SettingMaxBookingsPerUser.Name, "1")
	space2 := &Space{LocationID: location.ID, Enabled: true}
	GetSpaceRepository().Create(space2)

	enter, leave := hostBookingDay(0, 8, 17)
	res, _ := h.CreateBookingForUser(user.ID, space.ID, enter, leave, "")
	CheckTestInt(t, http.StatusCreated, res.StatusCode)

	enter, leave = hostBookingDay(1, 8, 17)
	res, _ = h.CreateBookingForUser(user.ID, space2.ID, enter, leave, "")
	CheckTestInt(t, http.StatusBadRequest, res.StatusCode)
	CheckTestInt(t, ResponseCodeBookingTooManyUpcomingBookings, res.ErrorCode)

	// Leave before enter is rejected as well.
	res, _ = h.CreateBookingForUser(user.ID, space2.ID, leave, enter, "")
	CheckTestInt(t, http.StatusBadRequest, res.StatusCode)
}

func TestHostAPICreateBookingDisabledSpace(t *testing.T) {
	_, user, _, space := setupHostBookingTest(t)
	space.Enabled = false
	GetSpaceRepository().Update(space)

	enter, leave := hostBookingDay(0, 8, 17)
	res, err := NewHostAPI().CreateBookingForUser(user.ID, space.ID, enter, leave, "")
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, http.StatusBadRequest, res.StatusCode)
}

func TestHostAPIAllowedBookersAndApproval(t *testing.T) {
	org, user, location, space := setupHostBookingTest(t)
	h := NewHostAPI()
	GetSettingsRepository().Set(org.ID, SettingFeatureGroups.Name, "1")

	otherUser := CreateTestUserInOrg(org)
	restrictedGroup := CreateTestGroup(org, otherUser)
	GetSpaceRepository().AddAllowedBookers(space, []string{restrictedGroup.ID})

	approvalSpace := &Space{LocationID: location.ID, Enabled: true, Name: "Approval"}
	GetSpaceRepository().Create(approvalSpace)
	approverGroup := CreateTestGroup(org, otherUser)
	GetSpaceRepository().AddApprovers(approvalSpace, []string{approverGroup.ID})

	enter, leave := hostBookingDay(0, 8, 17)
	spaces, err := h.GetSpaceAvailabilityForUser(user.ID, location.ID, enter, leave, nil)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 2, len(spaces))
	for _, s := range spaces {
		switch s.Space.ID {
		case space.ID:
			CheckTestBool(t, false, s.Allowed)
		case approvalSpace.ID:
			CheckTestBool(t, true, s.Allowed)
			CheckTestBool(t, true, s.ApprovalRequired)
		default:
			t.Fatalf("unexpected space %s", s.Space.ID)
		}
	}

	res, _ := h.CreateBookingForUser(user.ID, space.ID, enter, leave, "")
	CheckTestInt(t, http.StatusBadRequest, res.StatusCode)
	CheckTestInt(t, ResponseCodeBookingNotAllowedBooker, res.ErrorCode)

	res, _ = h.CreateBookingForUser(user.ID, approvalSpace.ID, enter, leave, "")
	CheckTestInt(t, http.StatusCreated, res.StatusCode)
	CheckTestBool(t, false, res.Approved)
}

func TestHostAPISpaceAvailabilityAndAttributes(t *testing.T) {
	org, user, location, space := setupHostBookingTest(t)
	h := NewHostAPI()
	space2 := &Space{LocationID: location.ID, Enabled: true, Name: "With monitor"}
	GetSpaceRepository().Create(space2)

	attr := &SpaceAttribute{OrganizationID: org.ID, Label: "Monitor", Type: SettingTypeBool, SpaceApplicable: true}
	if err := GetSpaceAttributeRepository().Create(attr); err != nil {
		t.Fatal(err)
	}
	GetSpaceAttributeValueRepository().Set(attr.ID, space2.ID, SpaceAttributeValueEntityTypeSpace, "1")

	defs, err := h.GetSpaceAttributes(org.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(defs))
	CheckTestString(t, "Monitor", defs[0].Label)

	enter, leave := hostBookingDay(0, 8, 17)
	res, _ := h.CreateBookingForUser(user.ID, space.ID, enter, leave, "")
	CheckTestInt(t, http.StatusCreated, res.StatusCode)

	spaces, err := h.GetSpaceAvailabilityForUser(user.ID, location.ID, enter, leave, nil)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 2, len(spaces))
	for _, s := range spaces {
		if s.Space.ID == space.ID {
			CheckTestBool(t, false, s.Available)
		} else {
			CheckTestBool(t, true, s.Available)
			CheckTestInt(t, 1, len(s.Attributes))
			CheckTestString(t, "1", s.Attributes[0].Value)
		}
	}

	filtered, err := h.GetSpaceAvailabilityForUser(user.ID, location.ID, enter, leave, []SearchAttributeFilter{{AttributeID: attr.ID, Comparator: "eq", Value: "1"}})
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(filtered))
	CheckTestString(t, space2.ID, filtered[0].Space.ID)

	// Invalid comparators are rejected.
	_, err = h.GetSpaceAvailabilityForUser(user.ID, location.ID, enter, leave, []SearchAttributeFilter{{AttributeID: attr.ID, Comparator: "like", Value: "1"}})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestHostAPISearchLocations(t *testing.T) {
	org, user, location, _ := setupHostBookingTest(t)
	h := NewHostAPI()
	location2 := &Location{OrganizationID: org.ID, Name: "Restricted", Enabled: true}
	GetLocationRepository().Create(location2)
	otherUser := CreateTestUserInOrg(org)
	group := CreateTestGroup(org, otherUser)
	GetLocationRepository().ReplaceAllowedBookers(location2, []string{group.ID})

	attr := &SpaceAttribute{OrganizationID: org.ID, Label: "Parking", Type: SettingTypeBool, LocationApplicable: true}
	GetSpaceAttributeRepository().Create(attr)
	GetSpaceAttributeValueRepository().Set(attr.ID, location.ID, SpaceAttributeValueEntityTypeLocation, "1")

	enter, leave := hostBookingDay(0, 8, 17)
	all, err := h.SearchLocationsForUser(user.ID, enter, leave, nil)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 2, len(all))
	for _, l := range all {
		if l.Location.ID == location.ID {
			CheckTestBool(t, true, l.Allowed)
			CheckTestInt(t, 1, len(l.Attributes))
		} else {
			CheckTestBool(t, false, l.Allowed)
		}
	}

	filtered, err := h.SearchLocationsForUser(user.ID, enter, leave, []SearchAttributeFilter{{AttributeID: attr.ID, Comparator: "eq", Value: "1"}})
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(filtered))
	CheckTestString(t, location.ID, filtered[0].Location.ID)

	free, err := h.SearchLocationsForUser(user.ID, enter, leave, []SearchAttributeFilter{{AttributeID: "numFreeSpaces", Comparator: "gte", Value: "1"}})
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(free))
	CheckTestString(t, location.ID, free[0].Location.ID)
}

func TestHostAPIBookingRejectsForeignOrgAndDisabledUsers(t *testing.T) {
	_, user, location, space := setupHostBookingTest(t)
	h := NewHostAPI()
	org2 := CreateTestOrg("other.com")
	foreignUser := CreateTestUserInOrg(org2)
	GetSettingsRepository().Set(org2.ID, SettingMaxDaysInAdvance.Name, "5000")

	enter, leave := hostBookingDay(0, 8, 17)
	if _, err := h.GetSpaceAvailabilityForUser(foreignUser.ID, location.ID, enter, leave, nil); err == nil {
		t.Fatal("expected error for foreign org location")
	}
	res, err := h.CreateBookingForUser(foreignUser.ID, space.ID, enter, leave, "")
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, http.StatusForbidden, res.StatusCode)
	locations, err := h.SearchLocationsForUser(foreignUser.ID, enter, leave, nil)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 0, len(locations))

	user.Disabled = true
	GetUserRepository().Update(user)
	if _, err := h.CreateBookingForUser(user.ID, space.ID, enter, leave, ""); err == nil {
		t.Fatal("expected error for disabled user")
	}
	if _, err := h.GetUpcomingBookingsForUser(user.ID); err == nil {
		t.Fatal("expected error for disabled user")
	}
	if _, err := h.GetUpcomingBookingsForUser("00000000-0000-0000-0000-000000000000"); err == nil {
		t.Fatal("expected error for unknown user")
	}
}
