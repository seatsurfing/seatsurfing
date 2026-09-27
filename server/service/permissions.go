package service

import (
	"log"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
)

// GetEffectivePermissions resolves the access a user has within an
// organization: for each permission, the highest level granted by any of their
// assigned roles. A user is never granted anything outside their own
// organization.
//
// Permissions absent from the result are not granted. Note that this covers
// only administrative functionality: every authenticated user additionally has
// a fixed baseline (their own bookings, buddies, preferences, profile and MFA,
// and read access to locations, spaces and availability) which is not
// represented here and can not be revoked.
func GetEffectivePermissions(user *User, organizationID string) map[Permission]PermissionLevel {
	if user == nil || user.OrganizationID != organizationID {
		return map[Permission]PermissionLevel{}
	}
	perms, err := GetUserRoleRepository().GetEffectivePermissions(user.ID)
	if err != nil {
		// Fail closed: an unreadable assignment must not grant access.
		log.Println(err)
		return map[Permission]PermissionLevel{}
	}
	return perms
}

// HasPermission reports whether the user holds at least the given level for a
// permission within the organization.
func HasPermission(user *User, organizationID string, p Permission, min PermissionLevel) bool {
	if min <= PermissionLevelNone {
		return true
	}
	return GetEffectivePermissions(user, organizationID)[p] >= min
}

// HasAnyPermission reports whether the user holds any administrative
// permission at all within the organization. It backs the checks that used to
// ask "is this user some kind of admin", such as whether to show the link into
// the administration UI or whether an admins-only MFA policy applies.
func HasAnyPermission(user *User, organizationID string) bool {
	return len(GetEffectivePermissions(user, organizationID)) > 0
}

// CanAccessOrg reports organization membership. It is not a privilege check:
// every authenticated user has baseline access to their own organization.
func CanAccessOrg(user *User, organizationID string) bool {
	return user != nil && user.OrganizationID == organizationID
}

// hasNoAdminRestrictions reports whether the organization lifts booking
// restrictions for booking admins and user is one.
func hasNoAdminRestrictions(user *User, organizationID string) bool {
	noAdminRestrictions, _ := GetSettingsRepository().GetBool(organizationID, SettingNoAdminRestrictions.Name)
	return noAdminRestrictions && HasPermission(user, organizationID, PermissionBookings, PermissionLevelAdmin)
}
