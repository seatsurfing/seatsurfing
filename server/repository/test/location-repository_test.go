package test

import (
	"testing"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestLocationsCount(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")

	l1 := &Location{
		OrganizationID: org.ID,
		Name:           "L1",
	}
	GetLocationRepository().Create(l1)

	res, err := GetLocationRepository().GetCount(org.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, res)
}

func TestLocationDeleteCleansUpSpaceRelations(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")

	l1 := &Location{OrganizationID: org.ID, Name: "L1"}
	GetLocationRepository().Create(l1)

	s1 := &Space{LocationID: l1.ID, Name: "S1"}
	GetSpaceRepository().Create(s1)

	g1 := &Group{OrganizationID: org.ID, Name: "G1"}
	GetGroupRepository().Create(g1)

	err := GetSpaceRepository().AddApprovers(s1, []string{g1.ID})
	CheckTestBool(t, true, err == nil)

	err = GetSpaceRepository().AddAllowedBookers(s1, []string{g1.ID})
	CheckTestBool(t, true, err == nil)

	approvers, err := GetSpaceRepository().GetApproverGroupIDs(s1.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, len(approvers))

	bookers, err := GetSpaceRepository().GetAllowedBookersGroupIDs(s1)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, len(bookers))

	err = GetLocationRepository().Delete(l1)
	CheckTestBool(t, true, err == nil)

	approvers, err = GetSpaceRepository().GetApproverGroupIDs(s1.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 0, len(approvers))

	bookers, err = GetSpaceRepository().GetAllowedBookersGroupIDs(s1)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 0, len(bookers))

	count, err := GetLocationRepository().GetCount(org.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 0, count)
}

func TestLocationDeleteCleansUpBookings(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	l1 := &Location{OrganizationID: org.ID, Name: "L1"}
	CheckTestIsNil(t, GetLocationRepository().Create(l1))
	s1 := &Space{LocationID: l1.ID, Name: "S1"}
	CheckTestIsNil(t, GetSpaceRepository().Create(s1))

	l2 := &Location{OrganizationID: org.ID, Name: "L2"}
	CheckTestIsNil(t, GetLocationRepository().Create(l2))
	s2 := &Space{LocationID: l2.ID, Name: "S2"}
	CheckTestIsNil(t, GetSpaceRepository().Create(s2))

	base := time.Now().Add(1 * time.Hour)
	booking := &Booking{UserID: user.ID, SpaceID: s1.ID, Enter: base, Leave: base.Add(1 * time.Hour)}
	CheckTestIsNil(t, GetBookingRepository().Create(booking))
	otherBooking := &Booking{UserID: user.ID, SpaceID: s2.ID, Enter: base, Leave: base.Add(1 * time.Hour)}
	CheckTestIsNil(t, GetBookingRepository().Create(otherBooking))

	recurringBooking := &RecurringBooking{
		UserID:  user.ID,
		SpaceID: s1.ID,
		Enter:   base,
		Leave:   base.Add(1 * time.Hour),
		Subject: "Recurring",
		Cadence: CadenceDaily,
		Details: &CadenceDailyDetails{Cycle: 1},
		End:     base.AddDate(0, 0, 30),
	}
	CheckTestIsNil(t, GetRecurringBookingRepository().Create(recurringBooking))
	otherRecurringBooking := &RecurringBooking{
		UserID:  user.ID,
		SpaceID: s2.ID,
		Enter:   base,
		Leave:   base.Add(1 * time.Hour),
		Subject: "Recurring",
		Cadence: CadenceDaily,
		Details: &CadenceDailyDetails{Cycle: 1},
		End:     base.AddDate(0, 0, 30),
	}
	CheckTestIsNil(t, GetRecurringBookingRepository().Create(otherRecurringBooking))

	pubBooking := &PublicBooking{Name: "Pub", Email: "pub@test.com"}
	CheckTestIsNil(t, GetPublicBookingRepository().Create(pubBooking))
	publicBooking := &Booking{SpaceID: s1.ID, PublicID: NullUUID(pubBooking.ID), Enter: base.Add(2 * time.Hour), Leave: base.Add(3 * time.Hour)}
	CheckTestIsNil(t, GetBookingRepository().Create(publicBooking))

	otherPubBooking := &PublicBooking{Name: "OtherPub", Email: "otherpub@test.com"}
	CheckTestIsNil(t, GetPublicBookingRepository().Create(otherPubBooking))
	otherPublicBooking := &Booking{SpaceID: s2.ID, PublicID: NullUUID(otherPubBooking.ID), Enter: base.Add(2 * time.Hour), Leave: base.Add(3 * time.Hour)}
	CheckTestIsNil(t, GetBookingRepository().Create(otherPublicBooking))

	CheckTestIsNil(t, GetLocationRepository().Delete(l1))

	// Dependent data for the deleted location's space is gone
	_, err := GetBookingRepository().GetOne(booking.ID)
	CheckTestBool(t, true, err != nil)
	_, err = GetRecurringBookingRepository().GetOne(recurringBooking.ID)
	CheckTestBool(t, true, err != nil)
	_, err = GetPublicBookingRepository().GetOne(pubBooking.ID)
	CheckTestBool(t, true, err != nil)

	// Data belonging to the other location is untouched
	_, err = GetBookingRepository().GetOne(otherBooking.ID)
	CheckTestIsNil(t, err)
	_, err = GetRecurringBookingRepository().GetOne(otherRecurringBooking.ID)
	CheckTestIsNil(t, err)
	_, err = GetPublicBookingRepository().GetOne(otherPubBooking.ID)
	CheckTestIsNil(t, err)
}
