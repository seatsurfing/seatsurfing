package test

import (
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

// A full pass over the feature: a Keycloak group grants a narrow role, the
// user gains exactly that access, and losing the group revokes it.
func TestEndToEndKeycloakGroupGrantsNarrowRole(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	CreateTestUserOrgAdmin(org)
	user := CreateTestUserInOrg(org)
	provider := createTestAuthProvider(org, "groups")

	role := CreateTestRole(org, "Group Coordinator", map[Permission]PermissionLevel{
		PermissionGroups: PermissionLevelWrite,
	})
	addMapping(provider, "/engineering/backend", AuthProviderMappingTargetRole, role.ID)

	// Before login: no administrative access at all.
	CheckTestBool(t, false, HasAnyPermission(user, org.ID))

	ReconcileUserFromIdP(user, provider, []string{"/engineering/backend", "/unmapped"})

	// Exactly the mapped access, and nothing more.
	CheckTestBool(t, true, HasPermission(user, org.ID, PermissionGroups, PermissionLevelWrite))
	CheckTestBool(t, false, HasPermission(user, org.ID, PermissionGroups, PermissionLevelAdmin))
	CheckTestBool(t, false, HasPermission(user, org.ID, PermissionUsers, PermissionLevelRead))
	CheckTestBool(t, false, HasPermission(user, org.ID, PermissionOrgSettings, PermissionLevelAdmin))

	// A leaf name alone must not match a full path.
	other := CreateTestUserInOrg(org)
	ReconcileUserFromIdP(other, provider, []string{"backend"})
	CheckTestBool(t, false, HasAnyPermission(other, org.ID))

	// Losing the group revokes it again.
	ReconcileUserFromIdP(user, provider, []string{})
	CheckTestBool(t, false, HasAnyPermission(user, org.ID))
}
