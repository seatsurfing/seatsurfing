package repository

import (
	"log"
	"strconv"
	"strings"
	"sync"

	. "github.com/seatsurfing/seatsurfing/server/api"
)

type RoleStore struct {
}

var roleRepository *RoleStore
var roleRepositoryOnce sync.Once

func GetRoleRepository() *RoleStore {
	roleRepositoryOnce.Do(func() {
		roleRepository = &RoleStore{}
		if _, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS roles (" +
			"id uuid DEFAULT uuid_generate_v4(), " +
			"organization_id uuid NOT NULL, " +
			"name VARCHAR NOT NULL, " +
			"description VARCHAR NOT NULL DEFAULT '', " +
			"system BOOLEAN NOT NULL DEFAULT FALSE, " +
			"auto_grant_plugin_permissions BOOLEAN NOT NULL DEFAULT FALSE, " +
			"PRIMARY KEY (id))"); err != nil {
			panic(err)
		}
		if _, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS role_permissions (" +
			"role_id uuid NOT NULL, " +
			"permission VARCHAR NOT NULL, " +
			"level INT NOT NULL, " +
			"PRIMARY KEY (role_id, permission))"); err != nil {
			panic(err)
		}
		if _, err := GetDatabase().DB().Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_org_name ON roles (organization_id, LOWER(name))"); err != nil {
			panic(err)
		}
	})
	return roleRepository
}

func (r *RoleStore) RunSchemaUpgrade(curVersion, targetVersion int) {
	if curVersion < 55 {
		// The presence report was split out of the analytics permission
		// because it covers personal data. Every role that could already see
		// it keeps that access.
		if _, err := GetDatabase().DB().Exec(
			"INSERT INTO role_permissions (role_id, permission, level) "+
				"SELECT role_id, $1, level FROM role_permissions WHERE permission = $2 "+
				"ON CONFLICT (role_id, permission) DO NOTHING",
			string(PermissionPresenceReport), string(PermissionAnalytics)); err != nil {
			panic(err)
		}
	}
	if curVersion < 54 {
		// Its own version rather than 53: a database already upgraded to 53
		// would never see this column otherwise.
		if _, err := GetDatabase().DB().Exec("ALTER TABLE roles " +
			"ADD COLUMN IF NOT EXISTS auto_grant_plugin_permissions BOOLEAN NOT NULL DEFAULT FALSE"); err != nil {
			panic(err)
		}
		// Mark the already-seeded API access role as tracking plugin
		// permissions, which is what it was seeded to do.
		if _, err := GetDatabase().DB().Exec(
			"UPDATE roles SET auto_grant_plugin_permissions = TRUE WHERE name = $1", RoleNameApiAccess); err != nil {
			panic(err)
		}
	}
	if curVersion < 53 {
		r.migrateLegacyRoles()
	}
}

// legacyRoleColumnExists reports whether users.role is still present. The
// version 53 upgrade drops it once its contents have been converted, so an
// upgrade interrupted between the drop and the version bump would otherwise
// fail on its next run.
func legacyRoleColumnExists() bool {
	var n int
	if err := GetDatabase().DB().QueryRow(
		"SELECT COUNT(*) FROM information_schema.columns " +
			"WHERE table_name = 'users' AND column_name = 'role'").Scan(&n); err != nil {
		log.Println(err)
		return false
	}
	return n > 0
}

// EnsureBuiltInRoles creates the roles every organization starts with and
// returns their IDs. It is idempotent, so it is safe to call on every
// organization creation and on a re-run of the schema upgrade.
//
// Only the organization administrator role is marked as a system role. The
// other two are ordinary roles that administrators may edit or delete:
// reproducing the old fixed ladder is a starting point, not a constraint.
func (r *RoleStore) EnsureBuiltInRoles(organizationID string) (orgAdminRoleID, floorPlanRoleID, apiRoleID string) {
	orgAdminRoleID = r.seedRole(organizationID, RoleNameOrgAdmin,
		"Full access to every functionality.", true, allPermissionsAtMaxLevel())
	floorPlanRoleID = r.seedRole(organizationID, RoleNameFloorPlanAdmin,
		"Manages areas, floor plans, spaces and approvals.", false, floorPlanAdminPermissions())
	// Read-only and read/write service accounts are granted the same
	// permissions: the read-only restriction is enforced by the HTTP method
	// check in handleServiceAccountAuth, not by permission level. Granting a
	// lower level here would revoke access such accounts have today.
	// Plugins register their permissions when they connect, which is after
	// this runs. The role therefore tracks later registrations, so that a
	// service account migrated from the old model - where roles 21 and 22
	// reached every plugin's admin endpoints - keeps working.
	apiRoleID = r.seedRole(organizationID, RoleNameApiAccess,
		"Grants a service account access to the REST API.", false, allPermissionsAtMaxLevel(), true)
	return orgAdminRoleID, floorPlanRoleID, apiRoleID
}

// migrateLegacyRoles seeds the built-in roles for every organization and
// assigns them according to each user's legacy users.role value, so that
// effective access is unchanged by the upgrade.
//
// This mirrors the version 13 upgrade in user-repository.go, which converted
// the org_admin/super_admin booleans into the users.role column.
func (r *RoleStore) migrateLegacyRoles() {
	orgIDs, err := r.getAllOrganizationIDs()
	if err != nil {
		panic(err)
	}
	// The built-in roles are seeded regardless; only reading the old ladder
	// depends on the column still being there.
	hasLegacyColumn := legacyRoleColumnExists()
	for _, orgID := range orgIDs {
		orgAdminRoleID, floorPlanRoleID, apiRoleID := r.EnsureBuiltInRoles(orgID)
		if !hasLegacyColumn {
			continue
		}

		// Former super admins become administrators of their own organization.
		r.assignUsersWithLegacyRole(orgID, int(UserRoleSuperAdmin), orgAdminRoleID, "super admin")
		r.assignUsersWithLegacyRole(orgID, int(UserRoleOrgAdmin), orgAdminRoleID, "org admin")
		r.assignUsersWithLegacyRole(orgID, int(UserRoleSpaceAdmin), floorPlanRoleID, "space admin")
		r.assignUsersWithLegacyRole(orgID, int(UserRoleServiceAccountRO), apiRoleID, "service account (read-only)")
		r.assignUsersWithLegacyRole(orgID, int(UserRoleServiceAccountRW), apiRoleID, "service account (read/write)")
	}
}

func (r *RoleStore) getAllOrganizationIDs() ([]string, error) {
	var result []string
	rows, err := GetDatabase().DB().Query("SELECT id FROM organizations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, nil
}

// seedRole creates the role if the organization does not already have one by
// that name, and returns its ID either way. Re-running the upgrade is safe.
func (r *RoleStore) seedRole(orgID, name, description string, system bool, perms map[Permission]PermissionLevel, autoGrantPluginPermissions ...bool) string {
	e, err := r.GetByName(orgID, name)
	if err == nil && e != nil {
		return e.ID
	}
	e = &Role{
		OrganizationID: orgID,
		Name:           name,
		Description:    description,
		System:         system,
		Permissions:    perms,
	}
	if len(autoGrantPluginPermissions) > 0 {
		e.AutoGrantPluginPermissions = autoGrantPluginPermissions[0]
	}
	if err := r.Create(e); err != nil {
		panic(err)
	}
	return e.ID
}

func (r *RoleStore) assignUsersWithLegacyRole(orgID string, legacyRole int, roleID, label string) {
	res, err := GetDatabase().DB().Exec("INSERT INTO user_roles (user_id, role_id, scope_type, scope_id, source) "+
		"SELECT id, $1, '', '', $2 FROM users WHERE organization_id = $3 AND role = $4 "+
		"ON CONFLICT DO NOTHING",
		roleID, RoleAssignmentSourceManual, orgID, legacyRole)
	if err != nil {
		panic(err)
	}
	if n, err := res.RowsAffected(); err == nil && n > 0 {
		log.Printf("Migrated %d %s user(s) in organization %s to role %s\n", n, label, orgID, roleID)
	}
}

func allPermissionsAtMaxLevel() map[Permission]PermissionLevel {
	result := make(map[Permission]PermissionLevel)
	for _, d := range GetPermissionDefinitions() {
		result[d.Key] = d.MaxLevel()
	}
	return result
}

// floorPlanAdminPermissions reproduces the access of the legacy space admin
// role: full control over areas, spaces, bookings and approvals, and read
// access to analytics, users and groups.
func floorPlanAdminPermissions() map[Permission]PermissionLevel {
	return map[Permission]PermissionLevel{
		PermissionAreas:           PermissionLevelAdmin,
		PermissionSpaceAttributes: PermissionLevelAdmin,
		PermissionBookings:        PermissionLevelAdmin,
		PermissionApprovals:       PermissionLevelAdmin,
		PermissionAnalytics:       PermissionLevelRead,
		PermissionPresenceReport:  PermissionLevelRead,
		PermissionUsers:           PermissionLevelRead,
		PermissionGroups:          PermissionLevelRead,
	}
}

func (r *RoleStore) Create(e *Role) error {
	var id string
	err := GetDatabase().DB().QueryRow("INSERT INTO roles "+
		"(organization_id, name, description, system, auto_grant_plugin_permissions) "+
		"VALUES ($1, $2, $3, $4, $5) "+
		"RETURNING id",
		e.OrganizationID, e.Name, e.Description, e.System, e.AutoGrantPluginPermissions).Scan(&id)
	if err != nil {
		return err
	}
	e.ID = id
	return r.setPermissions(e)
}

func (r *RoleStore) Update(e *Role) error {
	if _, err := GetDatabase().DB().Exec("UPDATE roles SET "+
		"name = $1, "+
		"description = $2, "+
		"auto_grant_plugin_permissions = $3 "+
		"WHERE id = $4",
		e.Name, e.Description, e.AutoGrantPluginPermissions, e.ID); err != nil {
		return err
	}
	return r.setPermissions(e)
}

// setPermissions replaces the role's permission rows. Levels of
// PermissionLevelNone and unknown permission keys are not stored.
func (r *RoleStore) setPermissions(e *Role) error {
	if _, err := GetDatabase().DB().Exec("DELETE FROM role_permissions WHERE role_id = $1", e.ID); err != nil {
		return err
	}
	if len(e.Permissions) == 0 {
		return nil
	}
	sqlStr := "INSERT INTO role_permissions (role_id, permission, level) VALUES "
	vals := []interface{}{}
	i := 1
	for permission, level := range e.Permissions {
		if level <= PermissionLevelNone {
			continue
		}
		sqlStr += "($" + strconv.Itoa(i) + ", $" + strconv.Itoa(i+1) + ", $" + strconv.Itoa(i+2) + "),"
		i += 3
		vals = append(vals, e.ID, string(permission), int(level))
	}
	if len(vals) == 0 {
		return nil
	}
	sqlStr = strings.TrimSuffix(sqlStr, ",")
	_, err := GetDatabase().DB().Exec(sqlStr, vals...)
	return err
}

func (r *RoleStore) GetOne(id string) (*Role, error) {
	e := &Role{}
	err := GetDatabase().DB().QueryRow("SELECT id, organization_id, name, description, system, auto_grant_plugin_permissions "+
		"FROM roles "+
		"WHERE id = $1",
		id).Scan(&e.ID, &e.OrganizationID, &e.Name, &e.Description, &e.System, &e.AutoGrantPluginPermissions)
	if err != nil {
		return nil, err
	}
	perms, err := r.getPermissions(e.ID)
	if err != nil {
		return nil, err
	}
	e.Permissions = perms
	return e, nil
}

func (r *RoleStore) GetByName(organizationID, name string) (*Role, error) {
	e := &Role{}
	err := GetDatabase().DB().QueryRow("SELECT id, organization_id, name, description, system, auto_grant_plugin_permissions "+
		"FROM roles "+
		"WHERE organization_id = $1 AND LOWER(name) = $2",
		organizationID, strings.ToLower(name)).Scan(&e.ID, &e.OrganizationID, &e.Name, &e.Description, &e.System, &e.AutoGrantPluginPermissions)
	if err != nil {
		return nil, err
	}
	perms, err := r.getPermissions(e.ID)
	if err != nil {
		return nil, err
	}
	e.Permissions = perms
	return e, nil
}

func (r *RoleStore) GetAll(organizationID string) ([]*Role, error) {
	var result []*Role
	rows, err := GetDatabase().DB().Query("SELECT id, organization_id, name, description, system, auto_grant_plugin_permissions "+
		"FROM roles "+
		"WHERE organization_id = $1 "+
		"ORDER BY name",
		organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &Role{}
		if err := rows.Scan(&e.ID, &e.OrganizationID, &e.Name, &e.Description, &e.System, &e.AutoGrantPluginPermissions); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	for _, e := range result {
		perms, err := r.getPermissions(e.ID)
		if err != nil {
			return nil, err
		}
		e.Permissions = perms
	}
	return result, nil
}

func (r *RoleStore) getPermissions(roleID string) (map[Permission]PermissionLevel, error) {
	result := make(map[Permission]PermissionLevel)
	rows, err := GetDatabase().DB().Query("SELECT permission, level "+
		"FROM role_permissions "+
		"WHERE role_id = $1",
		roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var permission string
		var level int
		if err := rows.Scan(&permission, &level); err != nil {
			return nil, err
		}
		result[Permission(permission)] = PermissionLevel(level)
	}
	return result, nil
}

// GrantPluginPermissions adds the given plugin-contributed permissions, at
// their highest level, to every role that tracks them. Called when a plugin
// registers, which happens on each connect, so it must be idempotent.
//
// This exists because the seeded roles are created during the schema upgrade,
// before any plugin has connected, and a plugin may be installed long after
// that. Without it a service account migrated from the old model would lose
// access to plugin endpoints such as SCIM, which it previously reached by
// virtue of being an organization administrator in all but name.
func (r *RoleStore) GrantPluginPermissions(defs []PermissionDefinition) error {
	for _, d := range defs {
		if d.Key == "" || !strings.HasPrefix(string(d.Key), PluginPermissionPrefix) {
			continue
		}
		if _, err := GetDatabase().DB().Exec(
			"INSERT INTO role_permissions (role_id, permission, level) "+
				"SELECT id, $1, $2 FROM roles WHERE auto_grant_plugin_permissions IS TRUE "+
				"ON CONFLICT (role_id, permission) DO NOTHING",
			string(d.Key), int(d.MaxLevel())); err != nil {
			return err
		}
	}
	return nil
}

func (r *RoleStore) Delete(e *Role) error {
	if err := GetAuthProviderMappingRepository().DeleteAllForTarget(AuthProviderMappingTargetRole, e.ID); err != nil {
		return err
	}
	if _, err := GetDatabase().DB().Exec("DELETE FROM user_roles WHERE role_id = $1", e.ID); err != nil {
		return err
	}
	if _, err := GetDatabase().DB().Exec("DELETE FROM role_permissions WHERE role_id = $1", e.ID); err != nil {
		return err
	}
	_, err := GetDatabase().DB().Exec("DELETE FROM roles WHERE id = $1", e.ID)
	return err
}

func (r *RoleStore) DeleteAll(organizationID string) error {
	if _, err := GetDatabase().DB().Exec("DELETE FROM user_roles WHERE "+
		"role_id IN (SELECT id FROM roles WHERE organization_id = $1)", organizationID); err != nil {
		return err
	}
	if _, err := GetDatabase().DB().Exec("DELETE FROM role_permissions WHERE "+
		"role_id IN (SELECT id FROM roles WHERE organization_id = $1)", organizationID); err != nil {
		return err
	}
	_, err := GetDatabase().DB().Exec("DELETE FROM roles WHERE organization_id = $1", organizationID)
	return err
}
