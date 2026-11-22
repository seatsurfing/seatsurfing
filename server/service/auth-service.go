package service

import (
	"sync"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
)

// AuthService holds authentication-related business rules shared by the
// login entry points.
type AuthService struct{}

var authService *AuthService
var authServiceOnce sync.Once

func GetAuthService() *AuthService {
	authServiceOnce.Do(func() {
		authService = &AuthService{}
	})
	return authService
}

// IsBoundToAuthProvider reports whether the user's configured authentication
// method is the given auth provider. A user without an auth provider binding
// uses another authentication method (e.g. password) and must not be able to
// sign in via an identity provider; the binding is only ever set explicitly by
// an administrator or when the account is created through that provider.
func (s *AuthService) IsBoundToAuthProvider(user *User, provider *AuthProvider) bool {
	if user == nil || provider == nil {
		return false
	}
	bound := string(user.AuthProviderID)
	if bound == "" || bound == GetSettingsRepository().GetNullUUID() {
		return false
	}
	return bound == provider.ID
}
