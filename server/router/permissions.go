package router

import (
	"log"
	"net/http"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
)

func CanPasswordLogin(user *User) bool {
	if user.PasswordPending {
		return false
	}
	return CanResetPassword(user)
}

func CanResetPassword(user *User) bool {
	if user.HashedPassword == "" {
		return false
	}
	if user.AuthProviderID != "" {
		return false
	}
	if user.Disabled {
		return false
	}
	if user.AccountType.IsServiceAccount() {
		return false
	}
	return true
}

func CanUpdatePassword(user *User) bool {
	if user.PasswordPending {
		return false
	}
	if !user.PasswordUpdateRequired {
		return false
	}
	return CanResetPassword(user)
}

// ─── Role-based access control ───────────────────────────────────────────────

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

// CheckPermission is the handler-level guard: it writes 403 and reports false
// when the request user lacks the required level.
func CheckPermission(w http.ResponseWriter, user *User, organizationID string, p Permission, min PermissionLevel) bool {
	if !HasPermission(user, organizationID, p, min) {
		SendForbidden(w)
		return false
	}
	return true
}

// PermissionsToRestModel converts a resolved permission map into the
// name-keyed form returned to clients. It always returns a non-nil map so the
// JSON is an object rather than null.
func PermissionsToRestModel(perms map[Permission]PermissionLevel) map[string]int {
	res := make(map[string]int, len(perms))
	for p, level := range perms {
		res[string(p)] = int(level)
	}
	return res
}

// ─── Lock-out prevention ─────────────────────────────────────────────────────

// OrgRetainsAdminWithout reports whether the organization would still have at
// least one administrator if the given users were to lose their access. An
// administrator is any enabled, non-service-account user holding the built-in
// organization administrator role. Pass the users being deleted, disabled, or
// whose assignments are being replaced.
func OrgRetainsAdminWithout(organizationID string, excludeUserIDs ...string) bool {
	found, err := GetUserRoleRepository().HasAdminUser(organizationID, excludeUserIDs)
	if err != nil {
		// Fail closed: refusing a change is recoverable, locking an
		// organization out of its own administration is not.
		log.Println(err)
		return false
	}
	return found
}

// UserWouldRetainAdmin reports whether the user would still hold the built-in
// organization administrator role if their assignments from the given source
// were replaced by newRoleIDs. Assignments from other sources are left in
// place and so still count.
func UserWouldRetainAdmin(user *User, newRoleIDs []string, source string) bool {
	if rolesIncludeAdmin(user.OrganizationID, newRoleIDs) {
		return true
	}
	other, err := GetUserRoleRepository().GetAssignmentsExcludingSource(user.ID, source)
	if err != nil {
		log.Println(err)
		return false
	}
	return rolesIncludeAdmin(user.OrganizationID, other)
}

// rolesIncludeAdmin reports whether any of the given role IDs is a system role
// of the organization. The organization administrator role is the only system
// role.
func rolesIncludeAdmin(organizationID string, roleIDs []string) bool {
	for _, id := range roleIDs {
		role, err := GetRoleRepository().GetOne(id)
		if err == nil && role.OrganizationID == organizationID && role.System {
			return true
		}
	}
	return false
}

// CheckOrgRetainsAdmin writes the appropriate error response and reports false
// when a change would leave the organization without an administrator.
func CheckOrgRetainsAdmin(w http.ResponseWriter, organizationID string, excludeUserIDs ...string) bool {
	if !OrgRetainsAdminWithout(organizationID, excludeUserIDs...) {
		SendBadRequestCode(w, ResponseCodeRoleWouldLeaveOrgWithoutAdmin)
		return false
	}
	return true
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

// ResultingPermissions computes the access a user would end up with if their
// manual role assignments were replaced by the given roles. Assignments from
// other sources, such as an identity provider, are left in place and so are
// included in the result.
//
// It exists so that a change can be judged before it is written: a refused
// request must not leave the assignments modified.
func ResultingPermissions(user *User, roleIDs []string) map[Permission]PermissionLevel {
	return ResultingPermissionsFromSource(user, roleIDs, RoleAssignmentSourceManual)
}

// ResultingPermissionsFromSource computes the access a user would end up with
// if the assignments from one source were replaced by the given roles.
func ResultingPermissionsFromSource(user *User, roleIDs []string, source string) map[Permission]PermissionLevel {
	res := make(map[Permission]PermissionLevel)
	merge := func(perms map[Permission]PermissionLevel) {
		for p, level := range perms {
			if level > res[p] {
				res[p] = level
			}
		}
	}
	for _, roleID := range roleIDs {
		role, err := GetRoleRepository().GetOne(roleID)
		if err != nil || role.OrganizationID != user.OrganizationID {
			continue
		}
		merge(role.Permissions)
	}
	current, err := GetUserRoleRepository().GetAssignmentsExcludingSource(user.ID, source)
	if err != nil {
		log.Println(err)
		return res
	}
	for _, roleID := range current {
		role, err := GetRoleRepository().GetOne(roleID)
		if err != nil {
			continue
		}
		merge(role.Permissions)
	}
	return res
}

// ─── Identity provider reconciliation ────────────────────────────────────────

// ReconcileUserFromIdP applies an auth provider's mappings to a user, based on
// the groups their identity provider reports. Roles and group memberships that
// this mechanism granted are replaced; anything an administrator assigned by
// hand is left alone, so a manual grant survives a login and a revoked
// provider group is actually revoked.
//
// It is called on every login through the provider, which is what makes the
// provider authoritative for the access it governs.
func ReconcileUserFromIdP(user *User, provider *AuthProvider, groups []string) {
	if user == nil || provider == nil {
		return
	}
	mappings, err := GetAuthProviderMappingRepository().GetAll(provider.ID)
	if err != nil {
		log.Println(err)
		return
	}
	if len(mappings) == 0 {
		return
	}

	reported := make(map[string]bool, len(groups))
	for _, g := range groups {
		reported[g] = true
	}
	var roleIDs, groupIDs []string
	for _, m := range mappings {
		if !reported[m.ClaimValue] {
			continue
		}
		switch m.TargetType {
		case AuthProviderMappingTargetRole:
			roleIDs = append(roleIDs, m.TargetID)
		case AuthProviderMappingTargetGroup:
			groupIDs = append(groupIDs, m.TargetID)
		}
	}

	// An organization must not be able to lose its last administrator because
	// somebody edited a group in the identity provider. Where that would be
	// the effect, the roles are left as they are and the operator is told.
	stillAdmin := UserWouldRetainAdmin(user, roleIDs, RoleAssignmentSourceOIDC)
	if stillAdmin || OrgRetainsAdminWithout(user.OrganizationID, user.ID) {
		if err := GetUserRoleRepository().SetRolesForUser(user.ID, roleIDs, RoleAssignmentSourceOIDC); err != nil {
			log.Println(err)
		}
	} else {
		log.Printf("Skipped identity provider role reconciliation for user %s: it would leave organization %s without an administrator\n",
			user.ID, user.OrganizationID)
	}

	if err := GetGroupRepository().SetMembershipsForUser(user.ID, groupIDs, RoleAssignmentSourceOIDC); err != nil {
		log.Println(err)
	}
}
