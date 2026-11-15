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

// CanGrantPermissions reports whether a user may hand out the given permission
// set: nobody may grant a level above their own. This generalizes the rule
// that already applied to the old role ladder, where a user could not raise
// anyone above their own rank.
func CanGrantPermissions(user *User, organizationID string, perms map[Permission]PermissionLevel) bool {
	own := GetEffectivePermissions(user, organizationID)
	for p, level := range perms {
		if level <= PermissionLevelNone {
			continue
		}
		if own[p] < level {
			return false
		}
	}
	return true
}

// CanManageUser reports whether actor may modify or delete target's account:
// target must be in actor's organization and must not hold any permission at
// a level above actor's own. Otherwise a delegated user administrator could
// take over a more privileged account (e.g. by resetting its password or
// second factor) and so gain access they could never be granted directly.
//
// Changes to one's own account are not judged here; the self-service rules
// of the individual operations apply to those.
func CanManageUser(actor *User, target *User) bool {
	if actor == nil || target == nil || !CanAccessOrg(actor, target.OrganizationID) {
		return false
	}
	if actor.ID == target.ID {
		return true
	}
	return CanGrantPermissions(actor, target.OrganizationID, GetEffectivePermissions(target, target.OrganizationID))
}

// hasNoAdminRestrictions reports whether the organization lifts booking
// restrictions for booking admins and user is one.
func hasNoAdminRestrictions(user *User, organizationID string) bool {
	noAdminRestrictions, _ := GetSettingsRepository().GetBool(organizationID, SettingNoAdminRestrictions.Name)
	return noAdminRestrictions && HasPermission(user, organizationID, PermissionBookings, PermissionLevelAdmin)
}
