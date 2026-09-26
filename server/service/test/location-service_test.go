package test

import (
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/service"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestLocationServiceSearch(t *testing.T) {
	org, user, location, _ := setupServiceTest(t)
	restricted := &Location{OrganizationID: org.ID, Name: "Restricted", Enabled: true}
	GetLocationRepository().Create(restricted)
	GetLocationRepository().ReplaceAllowedBookers(restricted, []string{CreateTestGroup(org, nil).ID})
	attr := &SpaceAttribute{OrganizationID: org.ID, Label: "Parking", Type: SettingTypeBool, LocationApplicable: true}
	GetSpaceAttributeRepository().Create(attr)
	GetSpaceAttributeValueRepository().Set(attr.ID, location.ID, SpaceAttributeValueEntityTypeLocation, "1")
	enter, leave := serviceTestSlot(0, 8, 17)

	all, err := GetLocationService().SearchLocationsForUser(user, enter, leave, nil)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 2, len(all))
	for _, l := range all {
		if l.Location.ID == location.ID {
			CheckTestBool(t, true, l.AllowedForUser)
			CheckTestInt(t, 1, len(l.Attributes))
		} else {
			CheckTestBool(t, false, l.AllowedForUser)
			CheckTestInt(t, 1, len(l.AllowedBookers))
		}
	}

	// Synthetic attributes are used for matching but not returned.
	found, err := GetLocationService().SearchLocationsForUser(user, enter, leave, []SearchAttribute{{AttributeID: SearchAttributeNumFreeSpaces, Comparator: "gte", Value: "1"}})
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(found))
	CheckTestString(t, location.ID, found[0].Location.ID)
	CheckTestInt(t, 1, len(found[0].Attributes))
	CheckTestString(t, attr.ID, found[0].Attributes[0].AttributeID)
}

func TestLocationServiceWeekdays(t *testing.T) {
	org, user, location, _ := setupServiceTest(t)
	location.BookableDays = "1,2,3,4,5"
	monday, _ := serviceTestSlot(0, 8, 17)
	saturday, _ := serviceTestSlot(5, 8, 17)
	svc := GetLocationService()
	CheckTestBool(t, true, svc.IsLocationWeekdayBookable(location, user, monday, monday.Add(8)))
	CheckTestBool(t, false, svc.IsLocationWeekdayBookable(location, user, saturday, saturday.Add(8)))
	CheckTestInt(t, 5, len(WeekdaysFromString(location.BookableDays)))

	admin := CreateTestUserOrgAdmin(org)
	GetSettingsRepository().Set(org.ID, SettingNoAdminRestrictions.Name, "1")
	CheckTestBool(t, true, svc.IsLocationWeekdayBookable(location, admin, saturday, saturday.Add(8)))
}

func TestValidateSearchAttributes(t *testing.T) {
	CheckTestIsNil(t, ValidateSearchAttributes([]SearchAttribute{
		{AttributeID: "00000000-0000-0000-0000-000000000000", Comparator: "eq", Value: "1"},
		{AttributeID: SearchAttributeBuddyOnSite, Comparator: "contains", Value: "*"},
	}))
	CheckTestBool(t, true, ValidateSearchAttributes([]SearchAttribute{{AttributeID: "not-a-uuid"}}) != nil)
	CheckTestBool(t, true, ValidateSearchAttributes([]SearchAttribute{{Comparator: "like"}}) != nil)
	CheckTestBool(t, true, ValidateSearchAttributes([]SearchAttribute{{Value: CreateTestString(257)}}) != nil)
}
