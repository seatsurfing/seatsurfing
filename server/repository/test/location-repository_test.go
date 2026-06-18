package test

import (
	"testing"

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
