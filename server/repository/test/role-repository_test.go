package test

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestRoleCRUD(t *testing.T) {
	ClearTestDB()
	org1 := CreateTestOrg("test.com")
	org2 := CreateTestOrg("test2.com")

	r1 := &Role{
		OrganizationID: org1.ID,
		Name:           "Group Manager",
		Description:    "Manages groups",
		Permissions: map[Permission]PermissionLevel{
			PermissionGroups: PermissionLevelAdmin,
		},
	}
	if err := GetRoleRepository().Create(r1); err != nil {
		t.Fatal(err)
	}
	CheckStringNotEmpty(t, r1.ID)

	// Read
	e, err := GetRoleRepository().GetOne(r1.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestString(t, "Group Manager", e.Name)
	CheckTestString(t, "Manages groups", e.Description)
	CheckTestBool(t, false, e.System)
	CheckTestInt(t, 1, len(e.Permissions))
	CheckTestInt(t, int(PermissionLevelAdmin), int(e.Permissions[PermissionGroups]))

	// Update replaces the permission set entirely
	e.Name = "Group Reader"
	e.Permissions = map[Permission]PermissionLevel{
		PermissionGroups:   PermissionLevelRead,
		PermissionAuditLog: PermissionLevelRead,
	}
	if err := GetRoleRepository().Update(e); err != nil {
		t.Fatal(err)
	}
	e, err = GetRoleRepository().GetOne(r1.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestString(t, "Group Reader", e.Name)
	CheckTestInt(t, 2, len(e.Permissions))
	CheckTestInt(t, int(PermissionLevelRead), int(e.Permissions[PermissionGroups]))

	// Organization scoping
	r2 := &Role{OrganizationID: org2.ID, Name: "Other Org Role"}
	if err := GetRoleRepository().Create(r2); err != nil {
		t.Fatal(err)
	}
	// Every organization is seeded with the built-in roles, so the custom one
	// is additional.
	list, err := GetRoleRepository().GetAll(org1.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 4, len(list))
	var names []string
	for _, e := range list {
		names = append(names, e.Name)
	}
	if !slices.Contains(names, "Group Reader") {
		t.Fatalf("expected the custom role in the list, got %v", names)
	}
	// The role created in the other organization is not visible here.
	if slices.Contains(names, "Other Org Role") {
		t.Fatalf("expected organization scoping, got %v", names)
	}

	// Delete
	if err := GetRoleRepository().Delete(e); err != nil {
		t.Fatal(err)
	}
	if _, err := GetRoleRepository().GetOne(r1.ID); err == nil {
		t.Fatal("expected deleted role to be gone")
	}
}

// A level of none is not stored: absence is what "not granted" means.
func TestRolePermissionLevelNoneNotStored(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	role := CreateTestRole(org, "Mixed", map[Permission]PermissionLevel{
		PermissionGroups: PermissionLevelAdmin,
		PermissionUsers:  PermissionLevelNone,
	})
	e, err := GetRoleRepository().GetOne(role.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(e.Permissions))
	if _, ok := e.Permissions[PermissionUsers]; ok {
		t.Fatal("expected a level of none not to be stored")
	}
}

func TestRoleNameUniquePerOrg(t *testing.T) {
	ClearTestDB()
	org1 := CreateTestOrg("test.com")
	org2 := CreateTestOrg("test2.com")
	CreateTestRole(org1, "Duplicate", nil)

	// Same name in the same organization is rejected, case-insensitively.
	err := GetRoleRepository().Create(&Role{OrganizationID: org1.ID, Name: "duplicate"})
	if err == nil {
		t.Fatal("expected duplicate role name in the same organization to be rejected")
	}
	// The same name in another organization is fine.
	if err := GetRoleRepository().Create(&Role{OrganizationID: org2.ID, Name: "Duplicate"}); err != nil {
		t.Fatal(err)
	}
}

func TestEffectivePermissionsUnionTakesHighestLevel(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	readerRole := CreateTestRole(org, "Reader", map[Permission]PermissionLevel{
		PermissionGroups:   PermissionLevelRead,
		PermissionAuditLog: PermissionLevelRead,
	})
	groupAdminRole := CreateTestRole(org, "Group Admin", map[Permission]PermissionLevel{
		PermissionGroups: PermissionLevelAdmin,
	})
	AssignTestRole(user, readerRole)
	AssignTestRole(user, groupAdminRole)

	perms, err := GetUserRoleRepository().GetEffectivePermissions(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 2, len(perms))
	// The higher of the two grants for groups wins.
	CheckTestInt(t, int(PermissionLevelAdmin), int(perms[PermissionGroups]))
	CheckTestInt(t, int(PermissionLevelRead), int(perms[PermissionAuditLog]))
	// A permission neither role grants is absent.
	if _, ok := perms[PermissionUsers]; ok {
		t.Fatal("expected ungranted permission to be absent")
	}
}

// A role may reference a permission belonging to a plugin that is offline.
// Such a key must never grant access.
func TestEffectivePermissionsIgnoresUnknownKeys(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	role := CreateTestRole(org, "Plugin Role", map[Permission]PermissionLevel{
		Permission("plugin.nonexistent.feature"): PermissionLevelAdmin,
		PermissionGroups:                         PermissionLevelRead,
	})
	AssignTestRole(user, role)

	perms, err := GetUserRoleRepository().GetEffectivePermissions(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(perms))
	CheckTestInt(t, int(PermissionLevelRead), int(perms[PermissionGroups]))
}

func TestUserRoleAssignment(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	role1 := CreateTestRole(org, "R1", map[Permission]PermissionLevel{PermissionGroups: PermissionLevelAdmin})
	role2 := CreateTestRole(org, "R2", map[Permission]PermissionLevel{PermissionUsers: PermissionLevelRead})

	// Adding twice is a no-op, not an error.
	if err := GetUserRoleRepository().Add(user.ID, role1.ID, RoleAssignmentSourceManual); err != nil {
		t.Fatal(err)
	}
	if err := GetUserRoleRepository().Add(user.ID, role1.ID, RoleAssignmentSourceManual); err != nil {
		t.Fatal(err)
	}
	ids, err := GetUserRoleRepository().GetRoleIDsForUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(ids))

	if err := GetUserRoleRepository().Add(user.ID, role2.ID, RoleAssignmentSourceManual); err != nil {
		t.Fatal(err)
	}
	count, err := GetUserRoleRepository().GetUserCountForRole(role2.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, count)

	if err := GetUserRoleRepository().Remove(user.ID, role1.ID); err != nil {
		t.Fatal(err)
	}
	ids, err = GetUserRoleRepository().GetRoleIDsForUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(ids))
	CheckTestString(t, role2.ID, ids[0])

	// Deleting a role removes its assignments.
	if err := GetRoleRepository().Delete(role2); err != nil {
		t.Fatal(err)
	}
	ids, err = GetUserRoleRepository().GetRoleIDsForUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 0, len(ids))
}

// Reconciliation from an identity provider must replace only what it granted,
// leaving assignments an administrator made by hand untouched.
func TestSetRolesForUserIsScopedToSource(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	manualRole := CreateTestRole(org, "Manual", map[Permission]PermissionLevel{PermissionGroups: PermissionLevelAdmin})
	oidcRole1 := CreateTestRole(org, "Oidc1", map[Permission]PermissionLevel{PermissionUsers: PermissionLevelRead})
	oidcRole2 := CreateTestRole(org, "Oidc2", map[Permission]PermissionLevel{PermissionAuditLog: PermissionLevelRead})

	if err := GetUserRoleRepository().Add(user.ID, manualRole.ID, RoleAssignmentSourceManual); err != nil {
		t.Fatal(err)
	}
	if err := GetUserRoleRepository().SetRolesForUser(user.ID, []string{oidcRole1.ID}, RoleAssignmentSourceOIDC); err != nil {
		t.Fatal(err)
	}
	ids, _ := GetUserRoleRepository().GetRoleIDsForUser(user.ID)
	CheckTestInt(t, 2, len(ids))

	// Replacing the identity provider's set drops oidcRole1 but keeps the manual one.
	if err := GetUserRoleRepository().SetRolesForUser(user.ID, []string{oidcRole2.ID}, RoleAssignmentSourceOIDC); err != nil {
		t.Fatal(err)
	}
	ids, _ = GetUserRoleRepository().GetRoleIDsForUser(user.ID)
	CheckTestInt(t, 2, len(ids))
	if !slices.Contains(ids, manualRole.ID) {
		t.Fatal("expected the manual assignment to survive reconciliation")
	}
	if !slices.Contains(ids, oidcRole2.ID) {
		t.Fatal("expected the new provider assignment to be present")
	}
	if slices.Contains(ids, oidcRole1.ID) {
		t.Fatal("expected the superseded provider assignment to be removed")
	}
}

// Backs the invariant that an organization is never left without an
// administrator.
func TestGetUserIDsWithPermissions(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	required := map[Permission]PermissionLevel{
		PermissionRoles: PermissionLevelAdmin,
		PermissionUsers: PermissionLevelAdmin,
	}

	admin := CreateTestUserWithPermissions(org, map[Permission]PermissionLevel{
		PermissionRoles: PermissionLevelAdmin,
		PermissionUsers: PermissionLevelAdmin,
	})
	// Holds only one of the two required permissions.
	CreateTestUserWithPermissions(org, map[Permission]PermissionLevel{
		PermissionRoles: PermissionLevelAdmin,
	})
	// Holds both, but below the required level.
	CreateTestUserWithPermissions(org, map[Permission]PermissionLevel{
		PermissionRoles: PermissionLevelRead,
		PermissionUsers: PermissionLevelRead,
	})

	ids, err := GetUserRoleRepository().GetUserIDsWithPermissions(org.ID, required, nil)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(ids))
	CheckTestString(t, admin.ID, ids[0])

	// Excluding the only administrator leaves none: this is what makes a
	// lock-out detectable before it happens.
	ids, err = GetUserRoleRepository().GetUserIDsWithPermissions(org.ID, required, []string{admin.ID})
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 0, len(ids))
}

// Disabled users and service accounts can not be the last administrator.
func TestGetUserIDsWithPermissionsExcludesDisabledAndServiceAccounts(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	required := map[Permission]PermissionLevel{PermissionRoles: PermissionLevelAdmin}
	adminRole := CreateTestRole(org, "Admin", map[Permission]PermissionLevel{PermissionRoles: PermissionLevelAdmin})

	disabled := &User{Email: uuid.New().String() + "@test.com", OrganizationID: org.ID, Disabled: true}
	if err := GetUserRepository().Create(disabled); err != nil {
		t.Fatal(err)
	}
	AssignTestRole(disabled, adminRole)

	serviceAccount := &User{Email: uuid.New().String() + "@test.com", OrganizationID: org.ID, AccountType: AccountTypeServiceAccountRW}
	if err := GetUserRepository().Create(serviceAccount); err != nil {
		t.Fatal(err)
	}
	AssignTestRole(serviceAccount, adminRole)

	ids, err := GetUserRoleRepository().GetUserIDsWithPermissions(org.ID, required, nil)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 0, len(ids))
}
