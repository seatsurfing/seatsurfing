package repository

import (
	"strconv"
	"strings"
	"sync"

	"github.com/lib/pq"

	. "github.com/seatsurfing/seatsurfing/server/api"
)

type UserRoleStore struct {
}

var userRoleRepository *UserRoleStore
var userRoleRepositoryOnce sync.Once

func GetUserRoleRepository() *UserRoleStore {
	userRoleRepositoryOnce.Do(func() {
		userRoleRepository = &UserRoleStore{}
		if _, err := GetDatabase().DB().Exec("CREATE TABLE IF NOT EXISTS user_roles (" +
			"user_id uuid NOT NULL, " +
			"role_id uuid NOT NULL, " +
			// scope_type and scope_id are reserved for a later scoped
			// assignment feature and are always empty for now. Empty strings
			// rather than NULL keep them usable in the primary key.
			"scope_type VARCHAR NOT NULL DEFAULT '', " +
			"scope_id VARCHAR NOT NULL DEFAULT '', " +
			"source VARCHAR NOT NULL DEFAULT 'manual', " +
			"PRIMARY KEY (user_id, role_id, scope_type, scope_id))"); err != nil {
			panic(err)
		}
		if _, err := GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles (user_id)"); err != nil {
			panic(err)
		}
		if _, err := GetDatabase().DB().Exec("CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles (role_id)"); err != nil {
			panic(err)
		}
	})
	return userRoleRepository
}

func (r *UserRoleStore) RunSchemaUpgrade(curVersion, targetVersion int) {
}

// GetEffectivePermissions resolves the highest level granted to the user by
// any of their assigned roles. Permissions granted at PermissionLevelNone are
// omitted, as are keys the catalogue does not know: a role may reference a
// permission belonging to a plugin that is currently offline, and an unknown
// key must never grant access.
func (r *UserRoleStore) GetEffectivePermissions(userID string) (map[Permission]PermissionLevel, error) {
	result := make(map[Permission]PermissionLevel)

	// A system role grants everything, resolved against the catalogue as it
	// stands right now rather than against what was stored when the role was
	// seeded. Plugins register their permissions when they connect, which can
	// be long after that, and a plugin may be installed later still - a
	// snapshot would leave the organization's administrator unable to reach
	// the plugin's own screens.
	holdsSystemRole, err := r.HoldsSystemRole(userID)
	if err != nil {
		return nil, err
	}
	if holdsSystemRole {
		for _, d := range GetPermissionDefinitions() {
			result[d.Key] = d.MaxLevel()
		}
	}

	rows, err := GetDatabase().DB().Query("SELECT rp.permission, MAX(rp.level) "+
		"FROM user_roles ur "+
		"INNER JOIN role_permissions rp ON rp.role_id = ur.role_id "+
		"WHERE ur.user_id = $1 "+
		"GROUP BY rp.permission",
		userID)
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
		if level <= int(PermissionLevelNone) {
			continue
		}
		if _, known := GetPermissionDefinition(Permission(permission)); !known {
			continue
		}
		if PermissionLevel(level) > result[Permission(permission)] {
			result[Permission(permission)] = PermissionLevel(level)
		}
	}
	return result, nil
}

// HoldsSystemRole reports whether any of the user's roles is a system role.
func (r *UserRoleStore) HoldsSystemRole(userID string) (bool, error) {
	var count int
	err := GetDatabase().DB().QueryRow("SELECT COUNT(*) "+
		"FROM user_roles ur "+
		"INNER JOIN roles r ON r.id = ur.role_id "+
		"WHERE ur.user_id = $1 AND r.system IS TRUE",
		userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRoleStore) GetRoleIDsForUser(userID string) ([]string, error) {
	var result []string
	rows, err := GetDatabase().DB().Query("SELECT role_id "+
		"FROM user_roles "+
		"WHERE user_id = $1",
		userID)
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

// GetAssignmentsExcludingSource returns the role IDs assigned to a user by
// any source other than the given one. It supports judging a replacement of
// one source's assignments without losing sight of the others.
func (r *UserRoleStore) GetAssignmentsExcludingSource(userID, source string) ([]string, error) {
	var result []string
	rows, err := GetDatabase().DB().Query("SELECT role_id "+
		"FROM user_roles "+
		"WHERE user_id = $1 AND source <> $2",
		userID, source)
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

func (r *UserRoleStore) GetRolesForUser(userID string) ([]*Role, error) {
	var result []*Role
	rows, err := GetDatabase().DB().Query("SELECT r.id, r.organization_id, r.name, r.description, r.system "+
		"FROM user_roles ur "+
		"INNER JOIN roles r ON r.id = ur.role_id "+
		"WHERE ur.user_id = $1 "+
		"ORDER BY r.name",
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &Role{}
		if err := rows.Scan(&e.ID, &e.OrganizationID, &e.Name, &e.Description, &e.System); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func (r *UserRoleStore) GetUserIDsForRole(roleID string) ([]string, error) {
	var result []string
	rows, err := GetDatabase().DB().Query("SELECT user_id "+
		"FROM user_roles "+
		"WHERE role_id = $1",
		roleID)
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

func (r *UserRoleStore) GetUserCountForRole(roleID string) (int, error) {
	var count int
	err := GetDatabase().DB().QueryRow("SELECT COUNT(*) FROM user_roles WHERE role_id = $1", roleID).Scan(&count)
	return count, err
}

// Add assigns a role to a user. Assigning a role the user already holds is a
// no-op.
func (r *UserRoleStore) Add(userID, roleID, source string) error {
	if source == "" {
		source = RoleAssignmentSourceManual
	}
	_, err := GetDatabase().DB().Exec("INSERT INTO user_roles (user_id, role_id, scope_type, scope_id, source) "+
		"VALUES ($1, $2, '', '', $3) "+
		"ON CONFLICT DO NOTHING",
		userID, roleID, source)
	return err
}

func (r *UserRoleStore) Remove(userID, roleID string) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2", userID, roleID)
	return err
}

// SetRolesForUser replaces the user's assignments for the given source,
// leaving assignments from other sources untouched. This lets identity
// provider reconciliation revoke only what it granted, without disturbing
// roles an administrator assigned by hand.
func (r *UserRoleStore) SetRolesForUser(userID string, roleIDs []string, source string) error {
	if source == "" {
		source = RoleAssignmentSourceManual
	}
	if _, err := GetDatabase().DB().Exec("DELETE FROM user_roles WHERE user_id = $1 AND source = $2", userID, source); err != nil {
		return err
	}
	if len(roleIDs) == 0 {
		return nil
	}
	sqlStr := "INSERT INTO user_roles (user_id, role_id, scope_type, scope_id, source) VALUES "
	vals := []interface{}{}
	i := 1
	for _, roleID := range roleIDs {
		sqlStr += "($" + strconv.Itoa(i) + ", $" + strconv.Itoa(i+1) + ", '', '', $" + strconv.Itoa(i+2) + "),"
		i += 3
		vals = append(vals, userID, roleID, source)
	}
	sqlStr = strings.TrimSuffix(sqlStr, ",")
	sqlStr += " ON CONFLICT DO NOTHING"
	_, err := GetDatabase().DB().Exec(sqlStr, vals...)
	return err
}

func (r *UserRoleStore) DeleteAllForUser(userID string) error {
	_, err := GetDatabase().DB().Exec("DELETE FROM user_roles WHERE user_id = $1", userID)
	return err
}

// GetUserIDsWithPermissions returns the enabled, non-service-account users of
// an organization whose effective access meets every one of the given minimum
// levels. It backs the invariant that an organization is never left without an
// administrator.
func (r *UserRoleStore) GetUserIDsWithPermissions(organizationID string, required map[Permission]PermissionLevel, excludeUserIDs []string) ([]string, error) {
	if len(required) == 0 {
		return nil, nil
	}
	// One EXISTS clause per required permission: the user must hold at least
	// one role granting that permission at or above the minimum level.
	// A nil slice would bind as NULL, and "NOT (id = ANY(NULL))" is NULL
	// rather than true, which would exclude every row.
	if excludeUserIDs == nil {
		excludeUserIDs = []string{}
	}
	var clauses []string
	vals := []interface{}{organizationID, pq.Array(excludeUserIDs)}
	i := 3
	for permission, level := range required {
		clauses = append(clauses, "EXISTS (SELECT 1 FROM user_roles ur "+
			"INNER JOIN role_permissions rp ON rp.role_id = ur.role_id "+
			"WHERE ur.user_id = u.id AND rp.permission = $"+strconv.Itoa(i)+" AND rp.level >= $"+strconv.Itoa(i+1)+")")
		vals = append(vals, string(permission), int(level))
		i += 2
	}
	sqlStr := "SELECT u.id FROM users u " +
		"WHERE u.organization_id = $1 " +
		"AND u.disabled IS NOT TRUE " +
		"AND u.account_type = " + strconv.Itoa(int(AccountTypePerson)) + " " +
		"AND NOT (u.id = ANY($2)) " +
		"AND " + strings.Join(clauses, " AND ")
	var result []string
	rows, err := GetDatabase().DB().Query(sqlStr, vals...)
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
