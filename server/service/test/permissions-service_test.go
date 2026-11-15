package test

import (
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/service"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestCanManageUserLowerOrEqualPrivileges(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	actor := CreateTestUserWithPermissions(org, map[Permission]PermissionLevel{
		PermissionUsers:    PermissionLevelAdmin,
		PermissionBookings: PermissionLevelAdmin,
	})
	plain := CreateTestUserInOrg(org)
	lower := CreateTestUserWithPermissions(org, map[Permission]PermissionLevel{PermissionBookings: PermissionLevelRead})
	equal := CreateTestUserWithPermissions(org, map[Permission]PermissionLevel{
		PermissionUsers:    PermissionLevelAdmin,
		PermissionBookings: PermissionLevelAdmin,
	})

	CheckTestBool(t, true, CanManageUser(actor, plain))
	CheckTestBool(t, true, CanManageUser(actor, lower))
	CheckTestBool(t, true, CanManageUser(actor, equal))
	CheckTestBool(t, true, CanManageUser(actor, actor))
}

func TestCanManageUserHigherPrivileges(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	actor := CreateTestUserWithPermissions(org, map[Permission]PermissionLevel{PermissionUsers: PermissionLevelAdmin})
	orgAdmin := CreateTestUserOrgAdmin(org)
	higherLevel := CreateTestUserWithPermissions(org, map[Permission]PermissionLevel{PermissionUsers: PermissionLevelAdmin, PermissionBookings: PermissionLevelRead})
	otherPerm := CreateTestUserWithPermissions(org, map[Permission]PermissionLevel{PermissionRoles: PermissionLevelRead})

	CheckTestBool(t, false, CanManageUser(actor, orgAdmin))
	CheckTestBool(t, false, CanManageUser(actor, higherLevel))
	CheckTestBool(t, false, CanManageUser(actor, otherPerm))

	// The organization administrator may manage everybody in the organization.
	CheckTestBool(t, true, CanManageUser(orgAdmin, actor))
	CheckTestBool(t, true, CanManageUser(orgAdmin, higherLevel))
}

func TestCanManageUserForeignOrg(t *testing.T) {
	ClearTestDB()
	org1 := CreateTestOrg("test1.com")
	org2 := CreateTestOrg("test2.com")
	admin1 := CreateTestUserOrgAdmin(org1)
	user2 := CreateTestUserInOrg(org2)

	CheckTestBool(t, false, CanManageUser(admin1, user2))
	CheckTestBool(t, false, CanManageUser(nil, user2))
	CheckTestBool(t, false, CanManageUser(admin1, nil))
}
