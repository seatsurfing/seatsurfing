package test

import (
	"log"
	"runtime/debug"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestBookingRepositoryPresenceReport(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user1 := CreateTestUserInOrgWithName(org, "u1@test.com", UserRoleUser)
	user2 := CreateTestUserInOrgWithName(org, "u2@test.com", UserRoleUser)
	user3 := CreateTestUserInOrgWithName(org, "u3@test.com", UserRoleUser)

	// Prepare
	l := &Location{
		Name:           "Test",
		OrganizationID: org.ID,
	}
	GetLocationRepository().Create(l)
	s1 := &Space{Name: "Test 1", LocationID: l.ID}
	GetSpaceRepository().Create(s1)

	tomorrow := time.Now().Add(24 * time.Hour)
	tomorrow = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 8, 0, 0, 0, tomorrow.Location())

	// Create booking
	b1_1 := &Booking{
		UserID:  user1.ID,
		SpaceID: s1.ID,
		Enter:   tomorrow.Add(0 * time.Hour),
		Leave:   tomorrow.Add(8 * time.Hour),
	}
	GetBookingRepository().Create(b1_1)
	b1_2 := &Booking{
		UserID:  user1.ID,
		SpaceID: s1.ID,
		Enter:   tomorrow.Add((24 + 0) * time.Hour),
		Leave:   tomorrow.Add((24 + 8) * time.Hour),
	}
	GetBookingRepository().Create(b1_2)
	b2_1 := &Booking{
		UserID:  user2.ID,
		SpaceID: s1.ID,
		Enter:   tomorrow.Add((24*2 + 0) * time.Hour),
		Leave:   tomorrow.Add((24*2 + 8) * time.Hour),
	}
	GetBookingRepository().Create(b2_1)

	end := tomorrow.Add(24 * 7 * time.Hour)
	res, err := GetBookingRepository().GetPresenceReport(org.ID, nil, tomorrow, end, 99999, 0)

	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 3, len(res))
	const DateFormat string = "2006-01-02"

	CheckTestString(t, user1.Email, res[0].User.Email)
	CheckTestInt(t, 1, res[0].Presence[tomorrow.Add(24*0*time.Hour).Format(DateFormat)])
	CheckTestInt(t, 1, res[0].Presence[tomorrow.Add(24*1*time.Hour).Format(DateFormat)])
	CheckTestInt(t, 0, res[0].Presence[tomorrow.Add(24*2*time.Hour).Format(DateFormat)])
	CheckTestInt(t, 0, res[0].Presence[tomorrow.Add(24*3*time.Hour).Format(DateFormat)])
	CheckTestInt(t, 0, res[0].Presence[tomorrow.Add(24*4*time.Hour).Format(DateFormat)])
	CheckTestInt(t, 0, res[0].Presence[tomorrow.Add(24*5*time.Hour).Format(DateFormat)])
	CheckTestInt(t, 0, res[0].Presence[tomorrow.Add(24*6*time.Hour).Format(DateFormat)])
	CheckTestInt(t, 0, res[0].Presence[tomorrow.Add(24*7*time.Hour).Format(DateFormat)])

	CheckTestString(t, user2.Email, res[1].User.Email)
	CheckTestInt(t, 0, res[1].Presence[tomorrow.Add(24*0*time.Hour).Format(DateFormat)])
	CheckTestInt(t, 0, res[1].Presence[tomorrow.Add(24*1*time.Hour).Format(DateFormat)])
	CheckTestInt(t, 1, res[1].Presence[tomorrow.Add(24*2*time.Hour).Format(DateFormat)])
	CheckTestInt(t, 0, res[1].Presence[tomorrow.Add(24*3*time.Hour).Format(DateFormat)])
	CheckTestInt(t, 0, res[1].Presence[tomorrow.Add(24*4*time.Hour).Format(DateFormat)])
	CheckTestInt(t, 0, res[1].Presence[tomorrow.Add(24*5*time.Hour).Format(DateFormat)])
	CheckTestInt(t, 0, res[1].Presence[tomorrow.Add(24*6*time.Hour).Format(DateFormat)])
	CheckTestInt(t, 0, res[1].Presence[tomorrow.Add(24*7*time.Hour).Format(DateFormat)])

	CheckTestString(t, user3.Email, res[2].User.Email)
	CheckTestInt(t, 0, res[2].Presence[tomorrow.Add(24*0*time.Hour).Format(DateFormat)])
	CheckTestInt(t, 0, res[2].Presence[tomorrow.Add(24*1*time.Hour).Format(DateFormat)])
	CheckTestInt(t, 0, res[2].Presence[tomorrow.Add(24*2*time.Hour).Format(DateFormat)])
	CheckTestInt(t, 0, res[2].Presence[tomorrow.Add(24*3*time.Hour).Format(DateFormat)])
	CheckTestInt(t, 0, res[2].Presence[tomorrow.Add(24*4*time.Hour).Format(DateFormat)])
	CheckTestInt(t, 0, res[2].Presence[tomorrow.Add(24*5*time.Hour).Format(DateFormat)])
	CheckTestInt(t, 0, res[2].Presence[tomorrow.Add(24*6*time.Hour).Format(DateFormat)])
	CheckTestInt(t, 0, res[2].Presence[tomorrow.Add(24*7*time.Hour).Format(DateFormat)])
}

func TestBookingRepositoryGetBookingsRequiringApproval(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	adminUser := CreateTestUserOrgAdmin(org)
	GetSettingsRepository().Set(org.ID, SettingMaxDaysInAdvance.Name, "5000")

	group := &Group{
		Name:           "Group 1",
		OrganizationID: org.ID,
	}
	GetGroupRepository().Create(group)
	GetGroupRepository().AddMembers(group, []string{adminUser.ID})
	location := &Location{
		Name:           "Location 1",
		OrganizationID: org.ID,
	}
	if err := GetLocationRepository().Create(location); err != nil {
		t.Fatalf("Expected nil error, but got %s\n%s", err, debug.Stack())
	}
	space := &Space{
		Name:       "H234",
		X:          50,
		Y:          100,
		Width:      200,
		Height:     300,
		Rotation:   90,
		LocationID: location.ID,
	}
	if err := GetSpaceRepository().Create(space); err != nil {
		t.Fatalf("Expected nil error, but got %s\n%s", err, debug.Stack())
	}
	if err := GetSpaceRepository().AddApprovers(space, []string{group.ID}); err != nil {
		t.Fatalf("Expected nil error, but got %s\n%s", err, debug.Stack())
	}

	bookings, err := GetBookingRepository().GetBookingsRequiringApproval(adminUser.ID)
	if err != nil {
		t.Fatalf("Expected nil error, but got %s\n%s", err, debug.Stack())
	}
	CheckTestInt(t, 0, len(bookings))
	count, err := GetBookingRepository().GetBookingsCountRequiringApproval(adminUser.ID)
	if err != nil {
		t.Fatalf("Expected nil error, but got %s\n%s", err, debug.Stack())
	}
	CheckTestInt(t, 0, count)

	booking := &Booking{
		UserID:   user.ID,
		SpaceID:  space.ID,
		Enter:    time.Now().Add(2 * time.Hour),
		Leave:    time.Now().Add(4 * time.Hour),
		Approved: false,
	}
	GetBookingRepository().Create(booking)

	bookings, err = GetBookingRepository().GetBookingsRequiringApproval(adminUser.ID)
	if err != nil {
		t.Fatalf("Expected nil error, but got %s\n%s", err, debug.Stack())
	}
	CheckTestInt(t, 1, len(bookings))
	CheckTestString(t, booking.ID, bookings[0].ID)
	count, err = GetBookingRepository().GetBookingsCountRequiringApproval(adminUser.ID)
	if err != nil {
		t.Fatalf("Expected nil error, but got %s\n%s", err, debug.Stack())
	}
	CheckTestInt(t, 1, count)

	booking.Approved = true
	GetBookingRepository().Update(booking)

	bookings, err = GetBookingRepository().GetBookingsRequiringApproval(adminUser.ID)
	if err != nil {
		t.Fatalf("Expected nil error, but got %s\n%s", err, debug.Stack())
	}
	CheckTestInt(t, 0, len(bookings))
	count, err = GetBookingRepository().GetBookingsCountRequiringApproval(adminUser.ID)
	if err != nil {
		t.Fatalf("Expected nil error, but got %s\n%s", err, debug.Stack())
	}
	CheckTestInt(t, 0, count)

}

func TestBookingRepositoryRecurringUUID(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	location := &Location{
		Name:           "Test",
		OrganizationID: org.ID,
	}
	GetLocationRepository().Create(location)
	space := &Space{
		Name:       "Test 1",
		LocationID: location.ID,
	}
	GetSpaceRepository().Create(space)

	recurringID := uuid.New()
	log.Println("Recurring ID 1:", recurringID.String())
	booking := &Booking{
		UserID:      user.ID,
		SpaceID:     space.ID,
		Enter:       time.Now().Add(1 * time.Hour),
		Leave:       time.Now().Add(2 * time.Hour),
		RecurringID: api.NullUUID(recurringID.String()),
	}
	log.Println("Recurring ID 2:", string(booking.RecurringID))
	err := GetBookingRepository().Create(booking)
	CheckTestBool(t, true, err == nil)

	bookingFromDB, err := GetBookingRepository().GetOne(booking.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestBool(t, true, bookingFromDB != nil)
	CheckTestString(t, recurringID.String(), string(bookingFromDB.RecurringID))
}

func TestBookingRepositoryRecurringUUIDNull(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	location := &Location{
		Name:           "Test",
		OrganizationID: org.ID,
	}
	GetLocationRepository().Create(location)
	space := &Space{
		Name:       "Test 1",
		LocationID: location.ID,
	}
	GetSpaceRepository().Create(space)

	booking := &Booking{
		UserID:      user.ID,
		SpaceID:     space.ID,
		Enter:       time.Now().Add(1 * time.Hour),
		Leave:       time.Now().Add(2 * time.Hour),
		RecurringID: api.NullUUID(""),
	}
	err := GetBookingRepository().Create(booking)
	CheckTestBool(t, true, err == nil)

	bookingFromDB, err := GetBookingRepository().GetOne(booking.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestBool(t, true, bookingFromDB != nil)
	CheckTestString(t, "", string(bookingFromDB.RecurringID))
}

func TestBookingRepositoryGetAllByOrgDateFiltering(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	location := &Location{
		Name:           "Test",
		OrganizationID: org.ID,
	}
	GetLocationRepository().Create(location)
	space := &Space{
		Name:       "Test 1",
		LocationID: location.ID,
	}
	GetSpaceRepository().Create(space)

	time8 := time.Date(2025, 1, 1, 8, 0, 0, 0, time.Local)
	time9 := time.Date(2025, 1, 1, 9, 0, 0, 0, time.Local)
	time10 := time.Date(2025, 1, 1, 10, 0, 0, 0, time.Local)
	time11 := time.Date(2025, 1, 1, 11, 0, 0, 0, time.Local)
	time12 := time.Date(2025, 1, 1, 12, 0, 0, 0, time.Local)

	// create booking from 09:00 to 11:00
	booking := &Booking{
		UserID:      user.ID,
		SpaceID:     space.ID,
		Enter:       time9,
		Leave:       time11,
		RecurringID: api.NullUUID(""),
	}
	GetBookingRepository().Create(booking)

	bookings_8_10, _ := GetBookingRepository().GetAllByOrg(org.ID, time8, time10)
	CheckTestInt(t, 0, len(bookings_8_10))

	bookings_10_12, _ := GetBookingRepository().GetAllByOrg(org.ID, time10, time12)
	CheckTestInt(t, 0, len(bookings_10_12))

	bookings_9_11, _ := GetBookingRepository().GetAllByOrg(org.ID, time9, time11)
	CheckTestInt(t, 1, len(bookings_9_11))

	bookings_8_12, _ := GetBookingRepository().GetAllByOrg(org.ID, time8, time12)
	CheckTestInt(t, 1, len(bookings_8_12))
}
