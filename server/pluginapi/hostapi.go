// Package pluginapi defines the host-side API contract that plugins call back into.
// It lives in a separate package (not api) to avoid the import cycle:
// repository → api, so api cannot import repository types.
package pluginapi

import (
	"github.com/seatsurfing/seatsurfing/server/repository"
)

// SettingsRepository covers the settings methods the plugin calls.
type SettingsRepository interface {
	Get(organizationID, name string) (string, error)
	GetBool(organizationID, name string) (bool, error)
	GetInt(organizationID, name string) (int, error)
	GetNullUUID() string
	GetOrgIDsByValue(name, value string) ([]string, error)
	Set(organizationID, name, value string) error
	Delete(organizationID, name string) error
}

// UserRepository covers the user methods the plugin calls.
type UserRepository interface {
	GetOne(id string) (*repository.User, error)
	GetAll(organizationID string, maxResults, offset int) ([]*repository.User, error)
	GetByEmail(organizationID, email string) (*repository.User, error)
	GetCount(organizationID string) (int, error)
	GetHashedPassword(password string) string
	GetUsersWithEmail(email string) ([]*repository.User, error)
	IsOrgAdmin(user *repository.User) bool
	IsSuperAdmin(user *repository.User) bool
	Create(e *repository.User) error
	Update(e *repository.User) error
	Delete(e *repository.User) error
}

// OrganizationRepository covers the organisation methods the plugin calls.
type OrganizationRepository interface {
	GetOne(id string) (*repository.Organization, error)
	GetAll() ([]*repository.Organization, error)
	GetOneByDomain(domain string) (*repository.Organization, error)
	GetByEmail(email string) (*repository.Organization, error)
	GetAllDaysPassedSinceSignup(daysPassed int, settingExists string) ([]*repository.Organization, error)
	GetPrimaryDomain(e *repository.Organization) (*repository.Domain, error)
	Create(e *repository.Organization) error
	Update(e *repository.Organization) error
	Delete(e *repository.Organization) error
	AddDomain(e *repository.Organization, domain string, active bool) error
	SetPrimaryDomain(e *repository.Organization, domain string) error
	CreateSampleData(org *repository.Organization) error
}

// GroupRepository covers the group methods the plugin calls.
type GroupRepository interface {
	GetOne(id string) (*repository.Group, error)
	GetAll(organizationID string) ([]*repository.Group, error)
	GetByName(organizationID, name string) (*repository.Group, error)
	GetMemberUserIDs(e *repository.Group) ([]string, error)
	AddMembers(e *repository.Group, userIDs []string) error
	RemoveMembers(e *repository.Group, userIDs []string) error
	Create(e *repository.Group) error
	Update(e *repository.Group) error
	Delete(e *repository.Group) error
}

// BookingRepository covers the booking methods the plugin calls.
type BookingRepository interface {
	GetOne(id string) (*repository.BookingDetails, error)
}

// SpaceRepository covers the space methods the plugin calls.
type SpaceRepository interface {
	GetOne(id string) (*repository.Space, error)
	GetCount(organizationID string) (int, error)
}

// LocationRepository covers the location methods the plugin calls.
type LocationRepository interface {
	GetOne(id string) (*repository.Location, error)
	GetCount(organizationID string) (int, error)
	GetTimezone(location *repository.Location) string
}

// AuthProviderRepository covers the auth-provider methods the plugin calls.
type AuthProviderRepository interface {
	Create(e *repository.AuthProvider) error
	Update(e *repository.AuthProvider) error
}

// AuthStateRepository covers the auth-state methods the plugin calls.
type AuthStateRepository interface {
	Create(e *repository.AuthState) error
}

// HostAPI is the full host-side interface exposed to plugins via RPC.
type HostAPI interface {
	GetSettingsRepository() SettingsRepository
	GetUserRepository() UserRepository
	GetOrganizationRepository() OrganizationRepository
	GetGroupRepository() GroupRepository
	GetBookingRepository() BookingRepository
	GetSpaceRepository() SpaceRepository
	GetLocationRepository() LocationRepository
	GetAuthProviderRepository() AuthProviderRepository
	GetAuthStateRepository() AuthStateRepository

	SendEmail(recipient, subject, body, language, orgID string) error
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
	IsValidLanguageCode(code string) bool
	DisablePasswordLogin() bool
}
