package test

import (
	"slices"
	"testing"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestSpaceGetAllInTimeApprovedField(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	location, space := CreateTestLocationAndSpace(org)

	base := time.Now().Add(1 * time.Hour)
	queryStart := base.Add(-1 * time.Hour)
	queryEnd := base.Add(3 * time.Hour)

	// bookings ordered by enter_time ASC: b1 first, b2 second
	b1 := &Booking{UserID: user.ID, SpaceID: space.ID, Enter: base, Leave: base.Add(1 * time.Hour), Approved: true}
	GetBookingRepository().Create(b1)
	b2 := &Booking{UserID: user.ID, SpaceID: space.ID, Enter: base.Add(1 * time.Hour), Leave: base.Add(2 * time.Hour), Approved: false}
	GetBookingRepository().Create(b2)

	res, err := GetSpaceRepository().GetAllInTime(location.ID, queryStart, queryEnd)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, len(res))
	CheckTestInt(t, 2, len(res[0].Bookings))
	CheckTestBool(t, true, res[0].Bookings[0].Approved)
	CheckTestBool(t, false, res[0].Bookings[1].Approved)
}

func TestSpacesCount(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")

	l1 := &Location{
		OrganizationID: org.ID,
		Name:           "L1",
	}
	GetLocationRepository().Create(l1)

	s1 := &Space{
		LocationID: l1.ID,
		Name:       "S1",
	}
	GetSpaceRepository().Create(s1)
	s2 := &Space{
		LocationID: l1.ID,
		Name:       "S2",
	}
	GetSpaceRepository().Create(s2)

	res, err := GetSpaceRepository().GetCount(org.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 2, res)
}

func TestSpacesCountMap(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")

	l1 := &Location{OrganizationID: org.ID, Name: "L1"}
	l2 := &Location{OrganizationID: org.ID, Name: "L2"}
	GetLocationRepository().Create(l1)
	GetLocationRepository().Create(l2)

	GetSpaceRepository().Create(&Space{LocationID: l1.ID, Name: "S1.1"})
	GetSpaceRepository().Create(&Space{LocationID: l1.ID, Name: "S1.2"})
	GetSpaceRepository().Create(&Space{LocationID: l1.ID, Name: "S1.3"})
	GetSpaceRepository().Create(&Space{LocationID: l2.ID, Name: "S2.1"})
	GetSpaceRepository().Create(&Space{LocationID: l2.ID, Name: "S2.2"})

	res, err := GetSpaceRepository().GetTotalCountMap(org.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 2, len(res))
	CheckTestInt(t, 3, res[l1.ID])
	CheckTestInt(t, 2, res[l2.ID])
}

func TestSpacesApproversCRUD(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")

	l1 := &Location{OrganizationID: org.ID, Name: "L1"}
	GetLocationRepository().Create(l1)
	s1 := &Space{LocationID: l1.ID, Name: "S1"}
	GetSpaceRepository().Create(s1)
	s2 := &Space{LocationID: l1.ID, Name: "S2"}
	GetSpaceRepository().Create(s2)

	g1 := &Group{
		OrganizationID: org.ID,
		Name:           "G1",
	}
	GetGroupRepository().Create(g1)

	g2 := &Group{
		OrganizationID: org.ID,
		Name:           "G2",
	}
	GetGroupRepository().Create(g2)

	list, err := GetSpaceRepository().GetApproverGroupIDs(s1.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 0, len(list))

	err = GetSpaceRepository().AddApprovers(s1, []string{g1.ID, g2.ID})
	CheckTestBool(t, true, err == nil)

	err = GetSpaceRepository().AddApprovers(s2, []string{g2.ID})
	CheckTestBool(t, true, err == nil)

	list, err = GetSpaceRepository().GetApproverGroupIDs(s1.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 2, len(list))
	CheckTestBool(t, true, slices.Contains(list, g1.ID))
	CheckTestBool(t, true, slices.Contains(list, g2.ID))

	list, err = GetSpaceRepository().GetApproverGroupIDs(s2.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, len(list))
	CheckTestBool(t, true, slices.Contains(list, g2.ID))

	err = GetSpaceRepository().RemoveApprovers(s1, []string{g2.ID})
	CheckTestBool(t, true, err == nil)

	list, err = GetSpaceRepository().GetApproverGroupIDs(s1.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, len(list))
	CheckTestBool(t, true, slices.Contains(list, g1.ID))

	list, err = GetSpaceRepository().GetApproverGroupIDs(s2.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, len(list))
	CheckTestBool(t, true, slices.Contains(list, g2.ID))
}

func TestSpacesAllowedBookersCRUD(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")

	l1 := &Location{OrganizationID: org.ID, Name: "L1"}
	GetLocationRepository().Create(l1)
	s1 := &Space{LocationID: l1.ID, Name: "S1"}
	GetSpaceRepository().Create(s1)
	s2 := &Space{LocationID: l1.ID, Name: "S2"}
	GetSpaceRepository().Create(s2)

	g1 := &Group{
		OrganizationID: org.ID,
		Name:           "G1",
	}
	GetGroupRepository().Create(g1)

	g2 := &Group{
		OrganizationID: org.ID,
		Name:           "G2",
	}
	GetGroupRepository().Create(g2)

	list, err := GetSpaceRepository().GetAllowedBookersGroupIDs(s1)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 0, len(list))

	err = GetSpaceRepository().AddAllowedBookers(s1, []string{g1.ID, g2.ID})
	CheckTestBool(t, true, err == nil)

	err = GetSpaceRepository().AddAllowedBookers(s2, []string{g2.ID})
	CheckTestBool(t, true, err == nil)

	list, err = GetSpaceRepository().GetAllowedBookersGroupIDs(s1)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 2, len(list))
	CheckTestBool(t, true, slices.Contains(list, g1.ID))
	CheckTestBool(t, true, slices.Contains(list, g2.ID))

	list, err = GetSpaceRepository().GetAllowedBookersGroupIDs(s2)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, len(list))
	CheckTestBool(t, true, slices.Contains(list, g2.ID))

	err = GetSpaceRepository().RemoveAllowedBookers(s1, []string{g2.ID})
	CheckTestBool(t, true, err == nil)

	list, err = GetSpaceRepository().GetAllowedBookersGroupIDs(s1)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, len(list))
	CheckTestBool(t, true, slices.Contains(list, g1.ID))

	list, err = GetSpaceRepository().GetAllowedBookersGroupIDs(s2)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, len(list))
	CheckTestBool(t, true, slices.Contains(list, g2.ID))
}

func TestSpaceDeleteCascades(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	location, space := CreateTestLocationAndSpace(org)
	otherSpace := &Space{LocationID: location.ID, Name: "Other"}
	CheckTestIsNil(t, GetSpaceRepository().Create(otherSpace))

	group := &Group{OrganizationID: org.ID, Name: "G1"}
	CheckTestIsNil(t, GetGroupRepository().Create(group))
	CheckTestIsNil(t, GetSpaceRepository().AddApprovers(space, []string{group.ID}))
	CheckTestIsNil(t, GetSpaceRepository().AddAllowedBookers(space, []string{group.ID}))

	attribute := &SpaceAttribute{OrganizationID: org.ID, Label: "Attr", Type: SettingTypeString, SpaceApplicable: true}
	CheckTestIsNil(t, GetSpaceAttributeRepository().Create(attribute))
	CheckTestIsNil(t, GetSpaceAttributeValueRepository().Set(attribute.ID, space.ID, SpaceAttributeValueEntityTypeSpace, "value"))

	base := time.Now().Add(1 * time.Hour)
	booking := &Booking{UserID: user.ID, SpaceID: space.ID, Enter: base, Leave: base.Add(1 * time.Hour)}
	CheckTestIsNil(t, GetBookingRepository().Create(booking))
	otherBooking := &Booking{UserID: user.ID, SpaceID: otherSpace.ID, Enter: base, Leave: base.Add(1 * time.Hour)}
	CheckTestIsNil(t, GetBookingRepository().Create(otherBooking))

	recurringBooking := &RecurringBooking{
		UserID:  user.ID,
		SpaceID: space.ID,
		Enter:   base,
		Leave:   base.Add(1 * time.Hour),
		Subject: "Recurring",
		Cadence: CadenceDaily,
		Details: &CadenceDailyDetails{Cycle: 1},
		End:     base.AddDate(0, 0, 30),
	}
	CheckTestIsNil(t, GetRecurringBookingRepository().Create(recurringBooking))

	anonBooking := &AnonymousBooking{Name: "Anon", Email: "anon@test.com"}
	CheckTestIsNil(t, GetAnonymousBookingRepository().Create(anonBooking))
	booking2 := &Booking{SpaceID: space.ID, AnonymousID: NullUUID(anonBooking.ID), Enter: base.Add(2 * time.Hour), Leave: base.Add(3 * time.Hour)}
	CheckTestIsNil(t, GetBookingRepository().Create(booking2))

	otherAnonBooking := &AnonymousBooking{Name: "OtherAnon", Email: "otheranon@test.com"}
	CheckTestIsNil(t, GetAnonymousBookingRepository().Create(otherAnonBooking))
	otherBooking2 := &Booking{SpaceID: otherSpace.ID, AnonymousID: NullUUID(otherAnonBooking.ID), Enter: base.Add(2 * time.Hour), Leave: base.Add(3 * time.Hour)}
	CheckTestIsNil(t, GetBookingRepository().Create(otherBooking2))

	CheckTestIsNil(t, GetSpaceRepository().Delete(space))

	// Dependent data for the deleted space is gone
	_, err := GetBookingRepository().GetOne(booking.ID)
	CheckTestBool(t, true, err != nil)
	_, err = GetRecurringBookingRepository().GetOne(recurringBooking.ID)
	CheckTestBool(t, true, err != nil)
	_, err = GetAnonymousBookingRepository().GetOne(anonBooking.ID)
	CheckTestBool(t, true, err != nil)
	approvers, err := GetSpaceRepository().GetApproverGroupIDs(space.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 0, len(approvers))
	allowedBookers, err := GetSpaceRepository().GetAllowedBookersGroupIDs(space)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 0, len(allowedBookers))
	attrValues, err := GetSpaceAttributeValueRepository().GetAllForEntity(space.ID, SpaceAttributeValueEntityTypeSpace)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 0, len(attrValues))

	// Data belonging to other spaces is untouched
	_, err = GetBookingRepository().GetOne(otherBooking.ID)
	CheckTestIsNil(t, err)
	_, err = GetSpaceRepository().GetOne(otherSpace.ID)
	CheckTestIsNil(t, err)
	_, err = GetAnonymousBookingRepository().GetOne(otherAnonBooking.ID)
	CheckTestIsNil(t, err)
}
