package test

import (
	"testing"

	"github.com/google/uuid"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

// restoreLegacyRoleColumn recreates the users.role column, putting the schema
// back into the state a pre-version-53 installation is in. The upgrade drops
// it again, so it has to be put back before each of these tests.
func restoreLegacyRoleColumn(t *testing.T) {
	t.Helper()
	if _, err := GetDatabase().DB().Exec(
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS role INT NOT NULL DEFAULT 0"); err != nil {
		t.Fatal(err)
	}
}

// createLegacyUser creates a user carrying the pre-version-53 role integer, as
// an installation being upgraded would have. The column is written directly:
// nothing in the code writes it any more.
func createLegacyUser(t *testing.T, org *Organization, role UserRole) *User {
	t.Helper()
	user := &User{
		Email:          uuid.New().String() + "@test.com",
		OrganizationID: org.ID,
	}
	if err := GetUserRepository().Create(user); err != nil {
		t.Fatal(err)
	}
	if _, err := GetDatabase().DB().Exec(
		"UPDATE users SET role = $1, account_type = 0 WHERE id = $2", int(role), user.ID); err != nil {
		t.Fatal(err)
	}
	return user
}

// runLegacyRoleMigration replays the version 53 upgrade over the current data,
// including the drop of the legacy column at the end.
func runLegacyRoleMigration(t *testing.T) {
	t.Helper()
	GetDatabase().DB().Exec("TRUNCATE user_roles")
	GetDatabase().DB().Exec("TRUNCATE role_permissions")
	GetDatabase().DB().Exec("TRUNCATE roles")
	GetUserRepository().RunSchemaUpgrade(52, 53)
	GetRoleRepository().RunSchemaUpgrade(52, 53)
	if _, err := GetDatabase().DB().Exec("ALTER TABLE users DROP COLUMN IF EXISTS role"); err != nil {
		t.Fatal(err)
	}
}

// legacyRoleColumnExists reports whether users.role is still present.
func legacyRoleColumnExists(t *testing.T) bool {
	t.Helper()
	var n int
	if err := GetDatabase().DB().QueryRow(
		"SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'role'").Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n > 0
}

func effectivePermissions(t *testing.T, user *User) map[Permission]PermissionLevel {
	t.Helper()
	perms, err := GetUserRoleRepository().GetEffectivePermissions(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	return perms
}

// The upgrade must not change what any existing user can do. Each legacy role
// is asserted against the access the removed Can*Org helpers would have
// granted: CanSpaceAdminOrg was role >= 10, CanAdminOrg was role >= 20 — which
// is why the service account roles 21 and 22, sitting above 20, were org
// admins in all but name.
func TestMigrationPreservesEffectiveAccess(t *testing.T) {
	ClearTestDB()
	restoreLegacyRoleColumn(t)
	org := CreateTestOrg("test.com")

	regular := createLegacyUser(t, org, UserRoleUser)
	spaceAdmin := createLegacyUser(t, org, UserRoleSpaceAdmin)
	orgAdmin := createLegacyUser(t, org, UserRoleOrgAdmin)
	serviceRO := createLegacyUser(t, org, UserRoleServiceAccountRO)
	serviceRW := createLegacyUser(t, org, UserRoleServiceAccountRW)
	superAdmin := createLegacyUser(t, org, UserRoleSuperAdmin)

	runLegacyRoleMigration(t)

	// A regular user gains nothing: baseline access is not represented as a
	// permission and can not be revoked.
	CheckTestInt(t, 0, len(effectivePermissions(t, regular)))

	// Space admin becomes the Floor Plan Administrator role.
	perms := effectivePermissions(t, spaceAdmin)
	for _, p := range []Permission{PermissionAreas, PermissionSpaceAttributes, PermissionBookings, PermissionApprovals} {
		if perms[p] != PermissionLevelAdmin {
			t.Fatalf("expected space admin to retain admin on %s, got %d", p, perms[p])
		}
	}
	CheckTestInt(t, int(PermissionLevelRead), int(perms[PermissionAnalytics]))
	CheckTestInt(t, int(PermissionLevelRead), int(perms[PermissionUsers]))
	CheckTestInt(t, int(PermissionLevelRead), int(perms[PermissionGroups]))
	// Space admins were never able to reach these.
	for _, p := range []Permission{PermissionOrgSettings, PermissionAuthProviders, PermissionAuditLog, PermissionServiceAccounts, PermissionRoles} {
		if perms[p] != PermissionLevelNone {
			t.Fatalf("expected space admin to have no access to %s, got %d", p, perms[p])
		}
	}

	// Org admins, former super admins and both service account kinds all had
	// org admin access and keep it.
	for label, user := range map[string]*User{
		"org admin":          orgAdmin,
		"super admin":        superAdmin,
		"service account RO": serviceRO,
		"service account RW": serviceRW,
	} {
		perms := effectivePermissions(t, user)
		for _, d := range GetPermissionDefinitions() {
			if perms[d.Key] != d.MaxLevel() {
				t.Fatalf("expected %s to hold %s at level %d, got %d", label, d.Key, d.MaxLevel(), perms[d.Key])
			}
		}
	}
}

// The read-only restriction on service accounts lives in the HTTP method
// check, not in the permission level, so both kinds keep full permissions and
// are distinguished by their account type instead.
func TestMigrationSetsAccountType(t *testing.T) {
	ClearTestDB()
	restoreLegacyRoleColumn(t)
	org := CreateTestOrg("test.com")
	regular := createLegacyUser(t, org, UserRoleUser)
	serviceRO := createLegacyUser(t, org, UserRoleServiceAccountRO)
	serviceRW := createLegacyUser(t, org, UserRoleServiceAccountRW)

	runLegacyRoleMigration(t)

	reload := func(u *User) *User {
		e, err := GetUserRepository().GetOne(u.ID)
		if err != nil {
			t.Fatal(err)
		}
		return e
	}
	CheckTestInt(t, int(AccountTypePerson), int(reload(regular).AccountType))
	CheckTestInt(t, int(AccountTypeServiceAccountRO), int(reload(serviceRO).AccountType))
	CheckTestInt(t, int(AccountTypeServiceAccountRW), int(reload(serviceRW).AccountType))
	// The legacy column is dropped once its contents have been converted.
	CheckTestBool(t, false, legacyRoleColumnExists(t))
	CheckTestBool(t, false, reload(regular).AccountType.IsServiceAccount())
	CheckTestBool(t, true, reload(serviceRO).AccountType.IsServiceAccount())
	CheckTestBool(t, true, reload(serviceRW).AccountType.IsServiceAccount())
}

func TestMigrationSeedsRolesPerOrganization(t *testing.T) {
	ClearTestDB()
	restoreLegacyRoleColumn(t)
	org1 := CreateTestOrg("test.com")
	org2 := CreateTestOrg("test2.com")
	admin1 := createLegacyUser(t, org1, UserRoleOrgAdmin)
	admin2 := createLegacyUser(t, org2, UserRoleOrgAdmin)

	runLegacyRoleMigration(t)

	for _, org := range []*Organization{org1, org2} {
		roles, err := GetRoleRepository().GetAll(org.ID)
		if err != nil {
			t.Fatal(err)
		}
		CheckTestInt(t, 3, len(roles))
		orgAdminRole, err := GetRoleRepository().GetByName(org.ID, RoleNameOrgAdmin)
		if err != nil {
			t.Fatal(err)
		}
		// The organization administrator role is protected from editing.
		CheckTestBool(t, true, orgAdminRole.System)
		if _, err := GetRoleRepository().GetByName(org.ID, RoleNameFloorPlanAdmin); err != nil {
			t.Fatal(err)
		}
		if _, err := GetRoleRepository().GetByName(org.ID, RoleNameApiAccess); err != nil {
			t.Fatal(err)
		}
	}

	// Each administrator is assigned their own organization's role.
	roles1, _ := GetUserRoleRepository().GetRolesForUser(admin1.ID)
	CheckTestInt(t, 1, len(roles1))
	CheckTestString(t, org1.ID, roles1[0].OrganizationID)
	roles2, _ := GetUserRoleRepository().GetRolesForUser(admin2.ID)
	CheckTestInt(t, 1, len(roles2))
	CheckTestString(t, org2.ID, roles2[0].OrganizationID)
}

// The upgrade runs again on every start-up until the version is bumped, and
// re-runs on plugin reconnect, so it must not duplicate roles or assignments.
func TestMigrationIsIdempotent(t *testing.T) {
	ClearTestDB()
	restoreLegacyRoleColumn(t)
	org := CreateTestOrg("test.com")
	admin := createLegacyUser(t, org, UserRoleOrgAdmin)

	runLegacyRoleMigration(t)
	GetRoleRepository().RunSchemaUpgrade(52, 53)
	GetRoleRepository().RunSchemaUpgrade(52, 53)

	roles, err := GetRoleRepository().GetAll(org.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 3, len(roles))
	assigned, _ := GetUserRoleRepository().GetRoleIDsForUser(admin.ID)
	CheckTestInt(t, 1, len(assigned))
}
