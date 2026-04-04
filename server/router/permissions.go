package router

import (
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
	if user.Role == UserRoleServiceAccountRO || user.Role == UserRoleServiceAccountRW {
		return false
	}
	return true
}
