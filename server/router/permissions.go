package router

import (
	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
)

func CanPasswordLogin(user *User) bool {
	if user.PasswordPending {
		return false
	}
	return CanResetPassword(user)
}

// PasswordLoginDenialReason returns the auth event error code explaining why
// CanPasswordLogin() is false for this user ("" if password login is allowed).
func PasswordLoginDenialReason(user *User) string {
	if user.PasswordPending {
		return AuthErrorPasswordPending
	}
	if user.HashedPassword == "" {
		return AuthErrorNoPasswordSet
	}
	if user.AuthProviderID != "" {
		return AuthErrorBoundToAuthProvider
	}
	if user.Disabled {
		return AuthErrorUserDisabled
	}
	if user.Role == UserRoleServiceAccountRO || user.Role == UserRoleServiceAccountRW {
		return AuthErrorServiceAccount
	}
	return ""
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
	if user.Role == UserRoleServiceAccountRO || user.Role == UserRoleServiceAccountRW {
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
