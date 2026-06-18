package pluginapi

import "github.com/seatsurfing/seatsurfing/server/repository"

// Entity type aliases — the plugin dot-imports pluginapi and gets these
// types without directly importing the repository package.

type User     = repository.User
type UserRole = repository.UserRole

const (
	UserRoleUser             = repository.UserRoleUser
	UserRoleSpaceAdmin       = repository.UserRoleSpaceAdmin
	UserRoleOrgAdmin         = repository.UserRoleOrgAdmin
	UserRoleServiceAccountRO = repository.UserRoleServiceAccountRO
	UserRoleServiceAccountRW = repository.UserRoleServiceAccountRW
	UserRoleSuperAdmin       = repository.UserRoleSuperAdmin
	DefaultUserLimit         = repository.DefaultUserLimit
)

type Organization = repository.Organization
type Domain       = repository.Domain

type Group = repository.Group

type Space        = repository.Space
type SpaceDetails = repository.SpaceDetails

type Location = repository.Location

type Booking        = repository.Booking
type BookingDetails = repository.BookingDetails

type AuthProvider     = repository.AuthProvider
type AuthProviderType = repository.AuthProviderType

const OAuth2 = repository.OAuth2

type AuthState     = repository.AuthState
type AuthStateType = repository.AuthStateType

const (
	AuthRequestState         = repository.AuthRequestState
	AuthResponseCache        = repository.AuthResponseCache
	AuthAtlassian            = repository.AuthAtlassian
	AuthMergeRequest         = repository.AuthMergeRequest
	AuthResetPasswordRequest = repository.AuthResetPasswordRequest
)

// AuthStateLoginPayload mirrors the struct defined in router/auth-router.go.
// Used only for JSON marshaling into AuthState.Payload.
type AuthStateLoginPayload struct {
	UserID    string `json:"userId"`
	LoginType string `json:"loginType"`
}

// SettingName is the type used to declare named settings with their type metadata.
type SettingName = repository.SettingName

// GetDatabase delegates to the repository singleton so plugin-owned
// repositories can create their own tables and run queries without
// importing the repository package directly.
func GetDatabase() *repository.Database {
	return repository.GetDatabase()
}
