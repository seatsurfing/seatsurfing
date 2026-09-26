package test

import (
	"testing"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/service"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

// serviceTestSlot returns wall-clock times on a Monday far enough in the
// future to be bookable.
func serviceTestSlot(dayOffset, fromHour, toHour int) (time.Time, time.Time) {
	base := time.Date(2030, 9, 2+dayOffset, 0, 0, 0, 0, time.UTC)
	return base.Add(time.Duration(fromHour) * time.Hour), base.Add(time.Duration(toHour) * time.Hour)
}

func setupServiceTest(t *testing.T) (*Organization, *User, *Location, *Space) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org.ID, SettingMaxDaysInAdvance.Name, "5000")
	user := CreateTestUserInOrg(org)
	location, space := CreateTestLocationAndSpace(org)
	return org, user, location, space
}

func TestBookingServiceCreateBooking(t *testing.T) {
	_, user, _, space := setupServiceTest(t)
	created := make(chan string, 1)
	svc := GetBookingService()
	svc.SetOnCreated(func(e *Booking) { created <- e.ID })
	defer svc.SetOnCreated(nil)

	enter, leave := serviceTestSlot(0, 8, 17)
	e, bErr := svc.CreateBooking(user, &BookingInput{SpaceID: space.ID, Subject: "Focus", Enter: enter, Leave: leave})
	if bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}
	CheckStringNotEmpty(t, e.ID)
	CheckTestString(t, user.ID, e.UserID)
	CheckTestBool(t, true, e.Approved)
	select {
	case id := <-created:
		CheckTestString(t, e.ID, id)
	case <-time.After(5 * time.Second):
		t.Fatal("OnCreated was not called")
	}

	_, bErr = svc.CreateBooking(user, &BookingInput{SpaceID: space.ID, Enter: enter, Leave: leave})
	CheckTestBool(t, true, bErr != nil && bErr.Kind == BookingErrorConflict)
	CheckTestInt(t, BookingCodeSlotConflict, bErr.Code)

	list, err := svc.GetUpcomingBookingsForUser(user)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(list))
}

func TestBookingServiceCreateBookingInvalidInput(t *testing.T) {
	_, user, _, space := setupServiceTest(t)
	svc := GetBookingService()
	enter, leave := serviceTestSlot(0, 8, 17)

	_, bErr := svc.CreateBooking(user, &BookingInput{Enter: enter, Leave: leave})
	CheckTestBool(t, true, bErr != nil && bErr.Kind == BookingErrorInvalid)
	_, bErr = svc.CreateBooking(user, &BookingInput{SpaceID: space.ID, Subject: CreateTestString(257), Enter: enter, Leave: leave})
	CheckTestBool(t, true, bErr != nil && bErr.Kind == BookingErrorInvalid)
	_, bErr = svc.CreateBooking(nil, &BookingInput{SpaceID: space.ID, Enter: enter, Leave: leave})
	CheckTestBool(t, true, bErr != nil && bErr.Kind == BookingErrorForbidden)

	org2 := CreateTestOrg("other.com")
	foreignUser := CreateTestUserInOrg(org2)
	_, bErr = svc.CreateBooking(foreignUser, &BookingInput{SpaceID: space.ID, Enter: enter, Leave: leave})
	CheckTestBool(t, true, bErr != nil && bErr.Kind == BookingErrorForbidden)
}

func TestBookingServicePrepareCommitForOtherUser(t *testing.T) {
	org, _, _, space := setupServiceTest(t)
	admin := CreateTestUserOrgAdmin(org)
	other := CreateTestUserInOrg(org)
	svc := GetBookingService()
	enter, leave := serviceTestSlot(0, 8, 17)

	p, bErr := svc.PrepareCreate(admin, &BookingInput{SpaceID: space.ID, Enter: enter, Leave: leave})
	if bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}
	CheckTestString(t, admin.ID, p.Booking.UserID)
	p.Booking.UserID = other.ID
	if bErr := svc.CommitCreate(admin, p); bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}
	stored, err := GetBookingRepository().GetOne(p.Booking.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestString(t, other.ID, stored.UserID)
}

func TestBookingServiceDisabledSpaceOnlyForSpaceAdmins(t *testing.T) {
	org, user, _, space := setupServiceTest(t)
	space.Enabled = false
	GetSpaceRepository().Update(space)
	svc := GetBookingService()
	enter, leave := serviceTestSlot(0, 8, 17)

	_, bErr := svc.CreateBooking(user, &BookingInput{SpaceID: space.ID, Enter: enter, Leave: leave})
	CheckTestBool(t, true, bErr != nil && bErr.Kind == BookingErrorInvalid)
	admin := CreateTestUserOrgAdmin(org)
	_, bErr = svc.CreateBooking(admin, &BookingInput{SpaceID: space.ID, Enter: enter, Leave: leave})
	CheckTestBool(t, true, bErr == nil)
}

func TestBookingServiceCheckBooking(t *testing.T) {
	org, user, location, space := setupServiceTest(t)
	svc := GetBookingService()
	enter, leave := serviceTestSlot(0, 8, 17)
	location, _ = GetLocationRepository().GetOne(location.ID)

	valid, code := svc.CheckBooking(&BookingInput{SpaceID: space.ID, Enter: enter, Leave: leave}, location, user, "", 0)
	CheckTestBool(t, true, valid)
	CheckTestInt(t, 0, code)

	valid, code = svc.CheckBooking(&BookingInput{SpaceID: space.ID, Enter: time.Now().AddDate(0, 0, -3), Leave: time.Now().AddDate(0, 0, -3).Add(time.Hour)}, location, user, "", 0)
	CheckTestBool(t, false, valid)
	CheckTestInt(t, BookingCodeInPast, code)

	GetSettingsRepository().Set(org.ID, SettingMaxBookingsPerUser.Name, "1")
	valid, code = svc.CheckBooking(&BookingInput{SpaceID: space.ID, Enter: enter, Leave: leave}, location, user, "", 1)
	CheckTestBool(t, false, valid)
	CheckTestInt(t, BookingCodeTooManyUpcomingBookings, code)

	location.BookableDays = "2,3"
	valid, code = svc.CheckBooking(&BookingInput{Enter: enter, Leave: leave}, location, user, "", 0)
	CheckTestBool(t, false, valid)
	CheckTestInt(t, BookingCodeInvalidWeekday, code)
}

func TestBookingServiceForeignSpaceRevealsNothing(t *testing.T) {
	_, _, _, space := setupServiceTest(t)
	space.RequireSubject = true
	GetSpaceRepository().Update(space)
	org2 := CreateTestOrg("other.com")
	foreignUser := CreateTestUserInOrg(org2)
	enter, leave := serviceTestSlot(0, 8, 17)

	// Without a subject, a same-organization user would get
	// BookingCodeSubjectRequired; a foreign user only learns "forbidden".
	_, bErr := GetBookingService().CreateBooking(foreignUser, &BookingInput{SpaceID: space.ID, Enter: enter, Leave: leave})
	CheckTestBool(t, true, bErr != nil && bErr.Kind == BookingErrorForbidden)
	CheckTestInt(t, 0, bErr.Code)
}

func TestBookingServiceDeleteBooking(t *testing.T) {
	org, user, _, space := setupServiceTest(t)
	svc := GetBookingService()
	deleted := make(chan string, 1)
	svc.SetOnDeleted(func(e *Booking) { deleted <- e.ID })
	defer svc.SetOnDeleted(nil)

	enter, leave := serviceTestSlot(0, 8, 17)
	e, bErr := svc.CreateBooking(user, &BookingInput{SpaceID: space.ID, Enter: enter, Leave: leave})
	if bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}

	// Other users of the organization and users of other organizations may
	// not delete it.
	otherUser := CreateTestUserInOrg(org)
	bErr = svc.DeleteBooking(otherUser, e.ID)
	CheckTestBool(t, true, bErr != nil && bErr.Kind == BookingErrorForbidden)
	foreignUser := CreateTestUserInOrg(CreateTestOrg("other.com"))
	bErr = svc.DeleteBooking(foreignUser, e.ID)
	CheckTestBool(t, true, bErr != nil && bErr.Kind == BookingErrorForbidden)

	if bErr := svc.DeleteBooking(user, e.ID); bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}
	select {
	case id := <-deleted:
		CheckTestString(t, e.ID, id)
	case <-time.After(5 * time.Second):
		t.Fatal("OnDeleted was not called")
	}
	bErr = svc.DeleteBooking(user, e.ID)
	CheckTestBool(t, true, bErr != nil && bErr.Kind == BookingErrorNotFound)
}

func TestBookingServiceDeleteBookingRules(t *testing.T) {
	org, user, _, space := setupServiceTest(t)
	svc := GetBookingService()
	GetSettingsRepository().Set(org.ID, SettingEnableMaxHourBeforeDelete.Name, "1")
	GetSettingsRepository().Set(org.ID, SettingMaxHoursBeforeDelete.Name, "24")

	// Starting too soon.
	soon := &Booking{UserID: user.ID, SpaceID: space.ID, Enter: time.Now().UTC().Add(2 * time.Hour), Leave: time.Now().UTC().Add(3 * time.Hour)}
	GetBookingRepository().Create(soon)
	bErr := svc.DeleteBooking(user, soon.ID)
	CheckTestBool(t, true, bErr != nil && bErr.Kind == BookingErrorForbidden)
	CheckTestInt(t, BookingCodeMaxHoursBeforeDelete, bErr.Code)

	// Already ended.
	past := &Booking{UserID: user.ID, SpaceID: space.ID, Enter: time.Now().UTC().Add(-72 * time.Hour), Leave: time.Now().UTC().Add(-71 * time.Hour)}
	GetBookingRepository().Create(past)
	bErr = svc.DeleteBooking(user, past.ID)
	CheckTestBool(t, true, bErr != nil && bErr.Kind == BookingErrorInvalid)

	// Bookings admins may delete other users' bookings, ignoring the time
	// limit only with "no admin restrictions".
	admin := CreateTestUserOrgAdmin(org)
	bErr = svc.DeleteBooking(admin, soon.ID)
	CheckTestBool(t, true, bErr != nil && bErr.Code == BookingCodeMaxHoursBeforeDelete)
	GetSettingsRepository().Set(org.ID, SettingNoAdminRestrictions.Name, "1")
	if bErr := svc.DeleteBooking(admin, soon.ID); bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}
}
