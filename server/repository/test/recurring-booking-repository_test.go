package test

import (
	"testing"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestRecurringBookingRepositoryDailyCRUD(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrgWithName(org, "u1@test.com", UserRoleUser)

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

	details := &CadenceDailyDetails{
		Cycle: 1,
	}
	rb := &RecurringBooking{
		UserID:  user.ID,
		SpaceID: s1.ID,
		Enter:   tomorrow.Add(0 * time.Hour),
		Leave:   tomorrow.Add(8 * time.Hour),
		Subject: "Test Subject",
		Cadence: CadenceDaily,
		Details: details,
		End:     tomorrow.AddDate(0, 0, 60),
	}
	err := GetRecurringBookingRepository().Create(rb)
	CheckTestBool(t, true, err == nil)

	rb2, err := GetRecurringBookingRepository().GetOne(rb.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestString(t, rb.ID, rb2.ID)
	CheckTestString(t, rb.UserID, rb2.UserID)
	CheckTestString(t, rb.SpaceID, rb2.SpaceID)
	CheckTestString(t, rb.Enter.Format(time.DateTime), rb2.Enter.Format(time.DateTime))
	CheckTestString(t, rb.Leave.Format(time.DateTime), rb2.Leave.Format(time.DateTime))
	CheckTestString(t, rb.Subject, rb2.Subject)
	CheckTestInt(t, int(CadenceDaily), int(rb2.Cadence))
	CheckTestInt(t, 1, rb2.Details.(CadenceDailyDetails).Cycle)
}

func TestRecurringBookingRepositoryWeeklyCRUD(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrgWithName(org, "u1@test.com", UserRoleUser)

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

	details := &CadenceWeeklyDetails{
		Cycle:    2,
		Weekdays: []time.Weekday{time.Monday, time.Wednesday},
	}
	rb := &RecurringBooking{
		UserID:  user.ID,
		SpaceID: s1.ID,
		Enter:   tomorrow.Add(0 * time.Hour),
		Leave:   tomorrow.Add(8 * time.Hour),
		Subject: "Test Subject",
		Cadence: CadenceWeekly,
		Details: details,
		End:     tomorrow.AddDate(0, 0, 60),
	}
	err := GetRecurringBookingRepository().Create(rb)
	CheckTestBool(t, true, err == nil)

	rb2, err := GetRecurringBookingRepository().GetOne(rb.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestString(t, rb.ID, rb2.ID)
	CheckTestString(t, rb.UserID, rb2.UserID)
	CheckTestString(t, rb.SpaceID, rb2.SpaceID)
	CheckTestString(t, rb.Enter.Format(time.DateTime), rb2.Enter.Format(time.DateTime))
	CheckTestString(t, rb.Leave.Format(time.DateTime), rb2.Leave.Format(time.DateTime))
	CheckTestString(t, rb.Subject, rb2.Subject)
	CheckTestInt(t, int(CadenceWeekly), int(rb2.Cadence))
	CheckTestInt(t, 2, rb2.Details.(CadenceWeeklyDetails).Cycle)
	CheckTestInt(t, 2, len(rb2.Details.(CadenceWeeklyDetails).Weekdays))
	CheckTestInt(t, int(time.Monday), int(rb2.Details.(CadenceWeeklyDetails).Weekdays[0]))
	CheckTestInt(t, int(time.Wednesday), int(rb2.Details.(CadenceWeeklyDetails).Weekdays[1]))
}

func TestRecurringBookingRepositoryCreateDailyBookingsCadence3(t *testing.T) {
	ClearTestDB()

	rb := &RecurringBooking{
		UserID:  "user1",
		SpaceID: "space1",
		Enter:   time.Date(2023, 10, 1, 9, 0, 0, 0, time.UTC),
		Leave:   time.Date(2023, 10, 1, 17, 0, 0, 0, time.UTC),
		Subject: "Test Daily Booking",
		Cadence: CadenceDaily,
		Details: &CadenceDailyDetails{
			Cycle: 3,
		},
		End: time.Date(2023, 10, 20, 0, 0, 0, 0, time.UTC),
	}
	bookings := GetRecurringBookingRepository().CreateBookings(rb)

	expectedEnter := []time.Time{
		time.Date(2023, 10, 1, 9, 0, 0, 0, time.UTC),
		time.Date(2023, 10, 4, 9, 0, 0, 0, time.UTC),
		time.Date(2023, 10, 7, 9, 0, 0, 0, time.UTC),
		time.Date(2023, 10, 10, 9, 0, 0, 0, time.UTC),
		time.Date(2023, 10, 13, 9, 0, 0, 0, time.UTC),
		time.Date(2023, 10, 16, 9, 0, 0, 0, time.UTC),
		time.Date(2023, 10, 19, 9, 0, 0, 0, time.UTC),
	}
	expectedLeave := []time.Time{
		time.Date(2023, 10, 1, 17, 0, 0, 0, time.UTC),
		time.Date(2023, 10, 4, 17, 0, 0, 0, time.UTC),
		time.Date(2023, 10, 7, 17, 0, 0, 0, time.UTC),
		time.Date(2023, 10, 10, 17, 0, 0, 0, time.UTC),
		time.Date(2023, 10, 13, 17, 0, 0, 0, time.UTC),
		time.Date(2023, 10, 16, 17, 0, 0, 0, time.UTC),
		time.Date(2023, 10, 19, 17, 0, 0, 0, time.UTC),
	}

	CheckTestInt(t, 7, len(bookings))
	for i, booking := range bookings {
		CheckTestString(t, rb.UserID, booking.UserID)
		CheckTestString(t, rb.SpaceID, booking.SpaceID)
		CheckTestString(t, rb.Subject, booking.Subject)
		CheckTestString(t, expectedEnter[i].Format(time.DateTime), booking.Enter.Format(time.DateTime))
		CheckTestString(t, expectedLeave[i].Format(time.DateTime), booking.Leave.Format(time.DateTime))
	}
}

func TestRecurringBookingRepositoryCreateDailyBookingsDaylightSaving(t *testing.T) {
	ClearTestDB()

	rb := &RecurringBooking{
		UserID:  "user1",
		SpaceID: "space1",
		Enter:   time.Date(2023, 10, 28, 9, 0, 0, 0, time.UTC),
		Leave:   time.Date(2023, 10, 28, 17, 0, 0, 0, time.UTC),
		Subject: "Test Daily Booking",
		Cadence: CadenceDaily,
		Details: &CadenceDailyDetails{
			Cycle: 1,
		},
		End: time.Date(2023, 10, 31, 0, 0, 0, 0, time.UTC),
	}
	bookings := GetRecurringBookingRepository().CreateBookings(rb)

	expectedEnter := []time.Time{
		time.Date(2023, 10, 28, 9, 0, 0, 0, time.UTC),
		time.Date(2023, 10, 29, 9, 0, 0, 0, time.UTC),
		time.Date(2023, 10, 30, 9, 0, 0, 0, time.UTC),
	}
	expectedLeave := []time.Time{
		time.Date(2023, 10, 28, 17, 0, 0, 0, time.UTC),
		time.Date(2023, 10, 29, 17, 0, 0, 0, time.UTC),
		time.Date(2023, 10, 30, 17, 0, 0, 0, time.UTC),
	}

	CheckTestInt(t, 3, len(bookings))
	for i, booking := range bookings {
		CheckTestString(t, rb.UserID, booking.UserID)
		CheckTestString(t, rb.SpaceID, booking.SpaceID)
		CheckTestString(t, rb.Subject, booking.Subject)
		CheckTestString(t, expectedEnter[i].Format(time.DateTime), booking.Enter.Format(time.DateTime))
		CheckTestString(t, expectedLeave[i].Format(time.DateTime), booking.Leave.Format(time.DateTime))
	}
}

func TestRecurringBookingRepositoryCreateWeeklyBookingsCadence1(t *testing.T) {
	ClearTestDB()

	rb := &RecurringBooking{
		UserID:  "user1",
		SpaceID: "space1",
		Enter:   time.Date(2023, 10, 2, 9, 0, 0, 0, time.UTC),
		Leave:   time.Date(2023, 10, 2, 17, 0, 0, 0, time.UTC),
		Subject: "Test Weekly Booking",
		Cadence: CadenceWeekly,
		Details: &CadenceWeeklyDetails{
			Cycle:    1,
			Weekdays: []time.Weekday{time.Monday, time.Wednesday, time.Friday},
		},
		End: time.Date(2023, 11, 2, 0, 0, 0, 0, time.UTC),
	}
	bookings := GetRecurringBookingRepository().CreateBookings(rb)

	expectedEnter := []time.Time{
		time.Date(2023, 10, 2, 9, 0, 0, 0, time.UTC),  // Monday
		time.Date(2023, 10, 4, 9, 0, 0, 0, time.UTC),  // Wednesday
		time.Date(2023, 10, 6, 9, 0, 0, 0, time.UTC),  // Friday
		time.Date(2023, 10, 9, 9, 0, 0, 0, time.UTC),  // Monday
		time.Date(2023, 10, 11, 9, 0, 0, 0, time.UTC), // Wednesday
		time.Date(2023, 10, 13, 9, 0, 0, 0, time.UTC), // Friday
		time.Date(2023, 10, 16, 9, 0, 0, 0, time.UTC), // Monday
		time.Date(2023, 10, 18, 9, 0, 0, 0, time.UTC), // Wednesday
		time.Date(2023, 10, 20, 9, 0, 0, 0, time.UTC), // Friday
		time.Date(2023, 10, 23, 9, 0, 0, 0, time.UTC), // Monday
		time.Date(2023, 10, 25, 9, 0, 0, 0, time.UTC), // Wednesday
		time.Date(2023, 10, 27, 9, 0, 0, 0, time.UTC), // Friday
		time.Date(2023, 10, 30, 9, 0, 0, 0, time.UTC), // Monday
		time.Date(2023, 11, 1, 9, 0, 0, 0, time.UTC),  // Wednesday
	}

	expectedLeave := []time.Time{
		time.Date(2023, 10, 2, 17, 0, 0, 0, time.UTC),  // Monday
		time.Date(2023, 10, 4, 17, 0, 0, 0, time.UTC),  // Wednesday
		time.Date(2023, 10, 6, 17, 0, 0, 0, time.UTC),  // Friday
		time.Date(2023, 10, 9, 17, 0, 0, 0, time.UTC),  // Monday
		time.Date(2023, 10, 11, 17, 0, 0, 0, time.UTC), // Wednesday
		time.Date(2023, 10, 13, 17, 0, 0, 0, time.UTC), // Friday
		time.Date(2023, 10, 16, 17, 0, 0, 0, time.UTC), // Monday
		time.Date(2023, 10, 18, 17, 0, 0, 0, time.UTC), // Wednesday
		time.Date(2023, 10, 20, 17, 0, 0, 0, time.UTC), // Friday
		time.Date(2023, 10, 23, 17, 0, 0, 0, time.UTC), // Monday
		time.Date(2023, 10, 25, 17, 0, 0, 0, time.UTC), // Wednesday
		time.Date(2023, 10, 27, 17, 0, 0, 0, time.UTC), // Friday
		time.Date(2023, 10, 30, 17, 0, 0, 0, time.UTC), // Monday
		time.Date(2023, 11, 1, 17, 0, 0, 0, time.UTC),  // Wednesday
	}

	CheckTestInt(t, 14, len(bookings))
	for i, booking := range bookings {
		CheckTestString(t, rb.UserID, booking.UserID)
		CheckTestString(t, rb.SpaceID, booking.SpaceID)
		CheckTestString(t, rb.Subject, booking.Subject)
		CheckTestString(t, expectedEnter[i].Format(time.DateTime), booking.Enter.Format(time.DateTime))
		CheckTestString(t, expectedLeave[i].Format(time.DateTime), booking.Leave.Format(time.DateTime))
	}
}

func TestRecurringBookingRepositoryCreateWeeklyBookingsCadence2(t *testing.T) {
	ClearTestDB()

	rb := &RecurringBooking{
		UserID:  "user1",
		SpaceID: "space1",
		Enter:   time.Date(2023, 10, 3, 9, 0, 0, 0, time.UTC),
		Leave:   time.Date(2023, 10, 3, 17, 0, 0, 0, time.UTC),
		Subject: "Test Weekly Booking",
		Cadence: CadenceWeekly,
		Details: &CadenceWeeklyDetails{
			Cycle:    2,
			Weekdays: []time.Weekday{time.Tuesday},
		},
		End: time.Date(2023, 11, 2, 0, 0, 0, 0, time.UTC),
	}
	bookings := GetRecurringBookingRepository().CreateBookings(rb)

	expectedEnter := []time.Time{
		time.Date(2023, 10, 3, 9, 0, 0, 0, time.UTC),  // Tuesday
		time.Date(2023, 10, 17, 9, 0, 0, 0, time.UTC), // Tuesday
		time.Date(2023, 10, 31, 9, 0, 0, 0, time.UTC), // Tuesday
	}

	expectedLeave := []time.Time{
		time.Date(2023, 10, 3, 17, 0, 0, 0, time.UTC),  // Tuesday
		time.Date(2023, 10, 17, 17, 0, 0, 0, time.UTC), // Tuesday
		time.Date(2023, 10, 31, 17, 0, 0, 0, time.UTC), // Tuesday
	}

	CheckTestInt(t, 3, len(bookings))
	for i, booking := range bookings {
		CheckTestString(t, rb.UserID, booking.UserID)
		CheckTestString(t, rb.SpaceID, booking.SpaceID)
		CheckTestString(t, rb.Subject, booking.Subject)
		CheckTestString(t, expectedEnter[i].Format(time.DateTime), booking.Enter.Format(time.DateTime))
		CheckTestString(t, expectedLeave[i].Format(time.DateTime), booking.Leave.Format(time.DateTime))
	}
}
