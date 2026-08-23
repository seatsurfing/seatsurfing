package test

import (
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

// The organization administrator role is documented as always granting full
// access. Plugins register their permissions asynchronously when they connect
// over gRPC, which can happen long after the role was seeded - and a plugin
// may be installed months later. The role must therefore cover permissions
// that did not exist when it was created.
func TestSystemRoleCoversLaterRegisteredPluginPermissions(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)

	const pluginPermission = Permission("plugin.test.laterfeature")
	RegisterPermission(PermissionDefinition{
		Key:           pluginPermission,
		AllowedLevels: []PermissionLevel{PermissionLevelNone, PermissionLevelAdmin},
		PluginID:      "test-plugin",
	})
	defer UnregisterPluginPermissions("test-plugin")

	perms, err := GetUserRoleRepository().GetEffectivePermissions(admin.ID)
	if err != nil {
		t.Fatal(err)
	}
	if perms[pluginPermission] != PermissionLevelAdmin {
		t.Fatalf("expected the organization administrator to hold %s at admin level, got %d",
			pluginPermission, perms[pluginPermission])
	}
}

// A role that is not a system role must not gain permissions it was never
// granted, however they come to be registered.
func TestOrdinaryRoleDoesNotGainLaterPermissions(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserWithPermissions(org, map[Permission]PermissionLevel{
		PermissionGroups: PermissionLevelAdmin,
	})

	const pluginPermission = Permission("plugin.test.laterfeature")
	RegisterPermission(PermissionDefinition{
		Key:           pluginPermission,
		AllowedLevels: []PermissionLevel{PermissionLevelNone, PermissionLevelAdmin},
		PluginID:      "test-plugin",
	})
	defer UnregisterPluginPermissions("test-plugin")

	perms, err := GetUserRoleRepository().GetEffectivePermissions(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if perms[pluginPermission] != PermissionLevelNone {
		t.Fatalf("expected an ordinary role not to gain %s, got %d", pluginPermission, perms[pluginPermission])
	}
}

// A service account migrated from the old model reached every plugin's admin
// endpoints, because roles 21 and 22 satisfied IsOrgAdmin. The seeded API
// access role therefore has to pick up plugin permissions as they register -
// plugins connect after the schema upgrade has already seeded it, and one may
// be installed months later. Without this, SCIM provisioning breaks silently
// on upgrade.
func TestApiAccessRoleTracksLaterPluginPermissions(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	serviceAccount := CreateTestServiceAccountRW(org)

	const scim = Permission("plugin.plus.scim")
	defs := []PermissionDefinition{{
		Key:           scim,
		AllowedLevels: []PermissionLevel{PermissionLevelNone, PermissionLevelAdmin},
		PluginID:      "plus-features",
	}}
	for _, d := range defs {
		RegisterPermission(d)
	}
	defer UnregisterPluginPermissions("plus-features")

	// Before the plugin connects, the permission does not exist for anyone.
	perms, _ := GetUserRoleRepository().GetEffectivePermissions(serviceAccount.ID)
	CheckTestInt(t, int(PermissionLevelNone), int(perms[scim]))

	// The plugin connects and registers.
	if err := GetRoleRepository().GrantPluginPermissions(defs); err != nil {
		t.Fatal(err)
	}
	perms, _ = GetUserRoleRepository().GetEffectivePermissions(serviceAccount.ID)
	CheckTestInt(t, int(PermissionLevelAdmin), int(perms[scim]))

	// Registering again on a reconnect must not duplicate the grant.
	if err := GetRoleRepository().GrantPluginPermissions(defs); err != nil {
		t.Fatal(err)
	}
	perms, _ = GetUserRoleRepository().GetEffectivePermissions(serviceAccount.ID)
	CheckTestInt(t, int(PermissionLevelAdmin), int(perms[scim]))
}

// Once an administrator has edited the role, its permission set is theirs.
func TestEditedApiAccessRoleStopsTracking(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")

	role, err := GetRoleRepository().GetByName(org.ID, RoleNameApiAccess)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestBool(t, true, role.AutoGrantPluginPermissions)

	// An edit through the role router clears the flag; simulate that.
	role.AutoGrantPluginPermissions = false
	role.Permissions = map[Permission]PermissionLevel{PermissionBookings: PermissionLevelRead}
	if err := GetRoleRepository().Update(role); err != nil {
		t.Fatal(err)
	}

	const scim = Permission("plugin.plus.scim")
	defs := []PermissionDefinition{{
		Key:           scim,
		AllowedLevels: []PermissionLevel{PermissionLevelNone, PermissionLevelAdmin},
		PluginID:      "plus-features",
	}}
	for _, d := range defs {
		RegisterPermission(d)
	}
	defer UnregisterPluginPermissions("plus-features")
	if err := GetRoleRepository().GrantPluginPermissions(defs); err != nil {
		t.Fatal(err)
	}

	reloaded, err := GetRoleRepository().GetOne(role.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, granted := reloaded.Permissions[scim]; granted {
		t.Fatal("expected an edited role not to gain plugin permissions behind the administrator's back")
	}
	// Ordinary roles never tracked them in the first place.
	floorPlan, _ := GetRoleRepository().GetByName(org.ID, RoleNameFloorPlanAdmin)
	CheckTestBool(t, false, floorPlan.AutoGrantPluginPermissions)
}
