package test

import (
	"slices"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestGroupCRUD(t *testing.T) {
	ClearTestDB()
	org1 := CreateTestOrg("test.com")
	org2 := CreateTestOrg("test.com")

	g1 := &Group{
		OrganizationID: org1.ID,
		Name:           "G1",
	}
	err := GetGroupRepository().Create(g1)
	CheckTestBool(t, true, err == nil)

	g2 := &Group{
		OrganizationID: org1.ID,
		Name:           "G2",
	}
	err = GetGroupRepository().Create(g2)
	CheckTestBool(t, true, err == nil)

	g3 := &Group{
		OrganizationID: org2.ID,
		Name:           "G3",
	}
	err = GetGroupRepository().Create(g3)
	CheckTestBool(t, true, err == nil)

	res, err := GetGroupRepository().GetAll(org1.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 2, len(res))
	CheckTestString(t, g1.Name, res[0].Name)
	CheckTestString(t, g2.Name, res[1].Name)

	res, err = GetGroupRepository().GetAll(org2.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, len(res))
	CheckTestString(t, g3.Name, res[0].Name)

	GetGroupRepository().Delete(g2)
	res, err = GetGroupRepository().GetAll(org1.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, len(res))
	CheckTestString(t, g1.Name, res[0].Name)

	err = GetGroupRepository().DeleteAll(org1.ID)
	CheckTestBool(t, true, err == nil)

	err = GetGroupRepository().DeleteAll(org2.ID)
	CheckTestBool(t, true, err == nil)
}

func TestGroupMembers(t *testing.T) {
	org := CreateTestOrg("test.com")

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

	u1 := CreateTestUserInOrg(org)
	u2 := CreateTestUserInOrg(org)
	u3 := CreateTestUserInOrg(org)

	err := GetGroupRepository().AddMembers(g1, []string{u1.ID, u2.ID})
	CheckTestBool(t, true, err == nil)
	err = GetGroupRepository().AddMembers(g2, []string{u2.ID, u3.ID})
	CheckTestBool(t, true, err == nil)

	res1, err := GetGroupRepository().GetMemberUserIDs(g1)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 2, len(res1))
	CheckTestBool(t, true, slices.Contains(res1, u1.ID))
	CheckTestBool(t, true, slices.Contains(res1, u2.ID))
	CheckTestBool(t, false, slices.Contains(res1, u3.ID))

	res2, err := GetGroupRepository().GetMemberUserIDs(g2)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 2, len(res2))
	CheckTestBool(t, false, slices.Contains(res2, u1.ID))
	CheckTestBool(t, true, slices.Contains(res2, u2.ID))
	CheckTestBool(t, true, slices.Contains(res2, u3.ID))

	err = GetGroupRepository().RemoveMembers(g2, []string{u2.ID})
	CheckTestBool(t, true, err == nil)

	res1, err = GetGroupRepository().GetMemberUserIDs(g1)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 2, len(res1))
	CheckTestBool(t, true, slices.Contains(res1, u1.ID))
	CheckTestBool(t, true, slices.Contains(res1, u2.ID))
	CheckTestBool(t, false, slices.Contains(res1, u3.ID))

	res2, err = GetGroupRepository().GetMemberUserIDs(g2)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, len(res2))
	CheckTestBool(t, false, slices.Contains(res2, u1.ID))
	CheckTestBool(t, false, slices.Contains(res2, u2.ID))
	CheckTestBool(t, true, slices.Contains(res2, u3.ID))
}
