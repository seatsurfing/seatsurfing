package test

import (
	"slices"
	"testing"
	"time"

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
