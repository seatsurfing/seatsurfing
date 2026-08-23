package api

import (
	"sort"
	"strings"
	"sync"
)

// Permission identifies a functionality whose access can be granted by a role.
type Permission string

// PermissionLevel is the degree of access a role grants for a Permission.
// Levels are ordered: a check for a minimum level is a >= comparison.
type PermissionLevel int

const (
	PermissionLevelNone  PermissionLevel = 0
	PermissionLevelRead  PermissionLevel = 10
	PermissionLevelWrite PermissionLevel = 20
	PermissionLevelAdmin PermissionLevel = 30
)

// Built-in permissions. Plugins contribute additional keys prefixed with
// PluginPermissionPrefix via RegisterPermission.
const (
	PermissionAreas           Permission = "areas"
	PermissionSpaceAttributes Permission = "space_attributes"
	PermissionBookings        Permission = "bookings"
	PermissionApprovals       Permission = "approvals"
	PermissionAnalytics       Permission = "analytics"
	PermissionUsers           Permission = "users"
	PermissionGroups          Permission = "groups"
	PermissionRoles           Permission = "roles"
	PermissionServiceAccounts Permission = "service_accounts"
	PermissionAuthProviders   Permission = "auth_providers"
	PermissionOrgSettings     Permission = "org_settings"
	PermissionAuditLog        Permission = "audit_log"
)

const PluginPermissionPrefix = "plugin."

// PermissionDefinition declares a permission and the levels that are
// meaningful for it. Levels that would be indistinguishable from the baseline
// access every user already has are deliberately not offered: for example
// every user must be able to read locations and spaces in order to book at
// all, so PermissionAreas offers only none and admin.
type PermissionDefinition struct {
	Key Permission
	// AllowedLevels always includes PermissionLevelNone and is sorted ascending.
	AllowedLevels []PermissionLevel
	// PluginID is empty for built-in permissions.
	PluginID string
}

// MaxLevel returns the highest level this permission can be granted at.
func (d *PermissionDefinition) MaxLevel() PermissionLevel {
	if len(d.AllowedLevels) == 0 {
		return PermissionLevelNone
	}
	return d.AllowedLevels[len(d.AllowedLevels)-1]
}

// AllowsLevel reports whether level is one of the declared levels.
func (d *PermissionDefinition) AllowsLevel(level PermissionLevel) bool {
	for _, l := range d.AllowedLevels {
		if l == level {
			return true
		}
	}
	return false
}

var builtInPermissions = []PermissionDefinition{
	{Key: PermissionAreas, AllowedLevels: []PermissionLevel{PermissionLevelNone, PermissionLevelAdmin}},
	{Key: PermissionSpaceAttributes, AllowedLevels: []PermissionLevel{PermissionLevelNone, PermissionLevelAdmin}},
	{Key: PermissionBookings, AllowedLevels: []PermissionLevel{PermissionLevelNone, PermissionLevelRead, PermissionLevelAdmin}},
	{Key: PermissionApprovals, AllowedLevels: []PermissionLevel{PermissionLevelNone, PermissionLevelAdmin}},
	{Key: PermissionAnalytics, AllowedLevels: []PermissionLevel{PermissionLevelNone, PermissionLevelRead}},
	{Key: PermissionUsers, AllowedLevels: []PermissionLevel{PermissionLevelNone, PermissionLevelRead, PermissionLevelAdmin}},
	{Key: PermissionGroups, AllowedLevels: []PermissionLevel{PermissionLevelNone, PermissionLevelRead, PermissionLevelWrite, PermissionLevelAdmin}},
	{Key: PermissionRoles, AllowedLevels: []PermissionLevel{PermissionLevelNone, PermissionLevelRead, PermissionLevelAdmin}},
	{Key: PermissionServiceAccounts, AllowedLevels: []PermissionLevel{PermissionLevelNone, PermissionLevelAdmin}},
	{Key: PermissionAuthProviders, AllowedLevels: []PermissionLevel{PermissionLevelNone, PermissionLevelAdmin}},
	{Key: PermissionOrgSettings, AllowedLevels: []PermissionLevel{PermissionLevelNone, PermissionLevelAdmin}},
	{Key: PermissionAuditLog, AllowedLevels: []PermissionLevel{PermissionLevelNone, PermissionLevelRead}},
}

var (
	permissionRegistry   = map[Permission]PermissionDefinition{}
	permissionRegistryMx sync.RWMutex
)

func init() {
	for _, d := range builtInPermissions {
		permissionRegistry[d.Key] = d
	}
}

// RegisterPermission adds or replaces a plugin-contributed permission.
// Plugins re-register on every reconnect, so this must be idempotent.
// Built-in permissions can not be overridden.
func RegisterPermission(d PermissionDefinition) {
	if d.Key == "" || !strings.HasPrefix(string(d.Key), PluginPermissionPrefix) {
		return
	}
	if len(d.AllowedLevels) == 0 {
		d.AllowedLevels = []PermissionLevel{PermissionLevelNone, PermissionLevelAdmin}
	}
	if !d.AllowsLevel(PermissionLevelNone) {
		d.AllowedLevels = append([]PermissionLevel{PermissionLevelNone}, d.AllowedLevels...)
	}
	sort.Slice(d.AllowedLevels, func(i, j int) bool { return d.AllowedLevels[i] < d.AllowedLevels[j] })
	permissionRegistryMx.Lock()
	defer permissionRegistryMx.Unlock()
	permissionRegistry[d.Key] = d
}

// UnregisterPluginPermissions drops every permission contributed by pluginID.
func UnregisterPluginPermissions(pluginID string) {
	if pluginID == "" {
		return
	}
	permissionRegistryMx.Lock()
	defer permissionRegistryMx.Unlock()
	for key, d := range permissionRegistry {
		if d.PluginID == pluginID {
			delete(permissionRegistry, key)
		}
	}
}

// GetPermissionDefinitions returns every known permission, built-ins first in
// catalogue order, then plugin permissions sorted by key.
func GetPermissionDefinitions() []PermissionDefinition {
	permissionRegistryMx.RLock()
	defer permissionRegistryMx.RUnlock()
	result := make([]PermissionDefinition, 0, len(permissionRegistry))
	for _, d := range builtInPermissions {
		if cur, ok := permissionRegistry[d.Key]; ok {
			result = append(result, cur)
		}
	}
	var pluginPerms []PermissionDefinition
	for _, d := range permissionRegistry {
		if d.PluginID != "" {
			pluginPerms = append(pluginPerms, d)
		}
	}
	sort.Slice(pluginPerms, func(i, j int) bool { return pluginPerms[i].Key < pluginPerms[j].Key })
	return append(result, pluginPerms...)
}

// GetPermissionDefinition looks up a single permission. Unknown keys resolve
// to not-found rather than an error: a role may reference a permission
// belonging to a plugin that is currently offline.
func GetPermissionDefinition(key Permission) (PermissionDefinition, bool) {
	permissionRegistryMx.RLock()
	defer permissionRegistryMx.RUnlock()
	d, ok := permissionRegistry[key]
	return d, ok
}

// IsValidPermissionLevel reports whether key is known and declares level.
func IsValidPermissionLevel(key Permission, level PermissionLevel) bool {
	d, ok := GetPermissionDefinition(key)
	if !ok {
		return false
	}
	return d.AllowsLevel(level)
}
