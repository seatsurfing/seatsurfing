package test

import (
	"bytes"
	"image"
	"image/png"
	"testing"
	"time"

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

func TestLocationServiceBlockedPeriod(t *testing.T) {
	org, user, location, _ := setupServiceTest(t)
	svc := GetLocationService()
	day := func(d int) *time.Time {
		v := time.Date(2030, 9, d, 0, 0, 0, 0, time.UTC)
		return &v
	}
	slot := func(d int) (time.Time, time.Time) {
		return serviceTestSlot(d-2, 8, 17)
	}

	enter, leave := slot(5)
	CheckTestBool(t, false, svc.IsLocationBlocked(location, user, enter, leave))

	location.BlockedFrom = day(4)
	location.BlockedUntil = day(6)
	enter, leave = slot(3)
	CheckTestBool(t, false, svc.IsLocationBlocked(location, user, enter, leave))
	enter, leave = slot(4)
	CheckTestBool(t, true, svc.IsLocationBlocked(location, user, enter, leave))
	enter, leave = slot(6)
	CheckTestBool(t, true, svc.IsLocationBlocked(location, user, enter, leave))
	enter, leave = slot(7)
	CheckTestBool(t, false, svc.IsLocationBlocked(location, user, enter, leave))

	// Multi-day booking overlapping the start of the period
	enter, _ = slot(2)
	_, leave = slot(4)
	CheckTestBool(t, true, svc.IsLocationBlocked(location, user, enter, leave))
	// Leave at midnight of the first blocked day does not touch it
	CheckTestBool(t, false, svc.IsLocationBlocked(location, user, enter, *day(4)))

	// Open end
	location.BlockedUntil = nil
	enter, leave = slot(25)
	CheckTestBool(t, true, svc.IsLocationBlocked(location, user, enter, leave))
	enter, leave = slot(3)
	CheckTestBool(t, false, svc.IsLocationBlocked(location, user, enter, leave))

	// Open start
	location.BlockedFrom = nil
	location.BlockedUntil = day(6)
	enter, leave = slot(2)
	CheckTestBool(t, true, svc.IsLocationBlocked(location, user, enter, leave))
	enter, leave = slot(7)
	CheckTestBool(t, false, svc.IsLocationBlocked(location, user, enter, leave))

	admin := CreateTestUserOrgAdmin(org)
	enter, leave = slot(2)
	CheckTestBool(t, true, svc.IsLocationBlocked(location, admin, enter, leave))
	GetSettingsRepository().Set(org.ID, SettingNoAdminRestrictions.Name, "1")
	CheckTestBool(t, false, svc.IsLocationBlocked(location, admin, enter, leave))
	CheckTestBool(t, true, svc.IsLocationBlocked(location, nil, enter, leave))
}

func TestLocationServiceIsValidBlockedPeriod(t *testing.T) {
	svc := GetLocationService()
	from := time.Date(2030, 9, 4, 0, 0, 0, 0, time.UTC)
	until := time.Date(2030, 9, 6, 0, 0, 0, 0, time.UTC)
	CheckTestBool(t, true, svc.IsValidBlockedPeriod(nil, nil))
	CheckTestBool(t, true, svc.IsValidBlockedPeriod(&from, nil))
	CheckTestBool(t, true, svc.IsValidBlockedPeriod(nil, &until))
	CheckTestBool(t, true, svc.IsValidBlockedPeriod(&from, &until))
	CheckTestBool(t, true, svc.IsValidBlockedPeriod(&from, &from))
	CheckTestBool(t, false, svc.IsValidBlockedPeriod(&until, &from))
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

func TestLocationServiceHideDisallowedSetting(t *testing.T) {
	org, user, _, _ := setupServiceTest(t)
	restricted := &Location{OrganizationID: org.ID, Name: "Restricted", Enabled: true}
	GetLocationRepository().Create(restricted)
	GetLocationRepository().ReplaceAllowedBookers(restricted, []string{CreateTestGroup(org, nil).ID})
	svc := GetLocationService()

	// Enabled: the restricted location is hidden.
	GetSettingsRepository().Set(org.ID, SettingHideDisallowedLocations.Name, "1")
	enabled, err := svc.IsHideDisallowedLocationsEnabled(org.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestBool(t, true, enabled)
	visible, err := svc.IsLocationVisibleForUser(user, restricted, true)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestBool(t, false, visible)

	// Missing setting: defaults to disabled without an error.
	GetSettingsRepository().Delete(org.ID, SettingHideDisallowedLocations.Name)
	enabled, err = svc.IsHideDisallowedLocationsEnabled(org.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestBool(t, false, enabled)
	visible, err = svc.IsLocationVisibleForUser(user, restricted, true)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestBool(t, true, visible)
}

func encodeTestPNG(t *testing.T, width, height int) []byte {
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewGray(image.Rect(0, 0, width, height))); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func checkLocationMapInvalid(t *testing.T, data []byte) {
	if _, err := GetLocationService().ParseLocationMap(data); err == nil {
		t.Fatalf("expected error for %q", data[:min(len(data), 40)])
	}
}

func TestLocationServiceParseLocationMap(t *testing.T) {
	svc := GetLocationService()

	m, err := svc.ParseLocationMap(encodeTestPNG(t, 200, 100))
	if err != nil {
		t.Fatal(err)
	}
	CheckTestString(t, "png", m.MimeType)
	CheckTestUint(t, 200, m.Width)
	CheckTestUint(t, 100, m.Height)

	m, err = svc.ParseLocationMap([]byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg" width="300" height="150px"></svg>`))
	if err != nil {
		t.Fatal(err)
	}
	CheckTestString(t, "svg+xml", m.MimeType)
	CheckTestUint(t, 300, m.Width)
	CheckTestUint(t, 150, m.Height)
}

func TestLocationServiceParseLocationMapInvalid(t *testing.T) {
	_, err := GetLocationService().ParseLocationMap(make([]byte, MaxLocationMapSize+1))
	CheckTestBool(t, true, err == ErrLocationMapTooLarge)
	_, err = GetLocationService().ParseLocationMap([]byte("no image"))
	CheckTestBool(t, true, err == ErrLocationMapInvalid)
	checkLocationMapInvalid(t, []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="10001" height="150"></svg>`))
	checkLocationMapInvalid(t, []byte("no image"))
	checkLocationMapInvalid(t, []byte(`<html width="300" height="150"></html>`))
	checkLocationMapInvalid(t, []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="100%" height="150"></svg>`))
	checkLocationMapInvalid(t, []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="0" height="150"></svg>`))
	checkLocationMapInvalid(t, encodeTestPNG(t, MaxLocationMapDimension+1, 1))
}
