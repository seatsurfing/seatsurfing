package test

import (
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestSpaceAttributeRepositoryCRUD(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")

	sa1 := &SpaceAttribute{
		OrganizationID:     org.ID,
		Label:              "Test 123",
		Type:               SettingTypeBool,
		SpaceApplicable:    true,
		LocationApplicable: true,
	}
	err := GetSpaceAttributeRepository().Create(sa1)
	CheckTestBool(t, true, err == nil)

	sa2 := &SpaceAttribute{
		OrganizationID:     org.ID,
		Label:              "Test 456",
		Type:               SettingTypeString,
		SpaceApplicable:    false,
		LocationApplicable: false,
	}
	err = GetSpaceAttributeRepository().Create(sa2)
	CheckTestBool(t, true, err == nil)

	sa11, err := GetSpaceAttributeRepository().GetOne(sa1.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestString(t, sa1.ID, sa11.ID)
	CheckTestString(t, sa1.Label, sa11.Label)
	CheckTestInt(t, int(sa1.Type), int(sa11.Type))
	CheckTestBool(t, sa1.LocationApplicable, sa11.LocationApplicable)
	CheckTestBool(t, sa1.SpaceApplicable, sa11.SpaceApplicable)

	sa21, err := GetSpaceAttributeRepository().GetOne(sa2.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestString(t, sa2.ID, sa21.ID)
	CheckTestString(t, sa2.Label, sa21.Label)
	CheckTestInt(t, int(sa2.Type), int(sa21.Type))
	CheckTestBool(t, sa2.LocationApplicable, sa21.LocationApplicable)
	CheckTestBool(t, sa2.SpaceApplicable, sa21.SpaceApplicable)

	list, err := GetSpaceAttributeRepository().GetAll(org.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 2, len(list))
	CheckTestString(t, sa1.Label, list[0].Label)
	CheckTestString(t, sa2.Label, list[1].Label)

	GetSpaceAttributeRepository().Delete(sa1)

	list, err = GetSpaceAttributeRepository().GetAll(org.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, len(list))
	CheckTestString(t, sa2.Label, list[0].Label)
}
