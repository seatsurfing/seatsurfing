package api

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
	GetOne(id string) (*User, error)
	GetAll(organizationID string, maxResults, offset int) ([]*User, error)
	GetByEmail(organizationID, email string) (*User, error)
	GetCount(organizationID string) (int, error)
	GetHashedPassword(password string) string
	GetUsersWithEmail(email string) ([]*User, error)
	IsOrgAdmin(user *User) bool
	IsSuperAdmin(user *User) bool
	Create(e *User) error
	Update(e *User) error
	Delete(e *User) error
}

// OrganizationRepository covers the organisation methods the plugin calls.
type OrganizationRepository interface {
	GetOne(id string) (*Organization, error)
	GetAll() ([]*Organization, error)
	GetOneByDomain(domain string) (*Organization, error)
	GetByEmail(email string) (*Organization, error)
	GetAllDaysPassedSinceSignup(daysPassed int, settingExists string) ([]*Organization, error)
	GetPrimaryDomain(e *Organization) (*Domain, error)
	Create(e *Organization) error
	Update(e *Organization) error
	Delete(e *Organization) error
	AddDomain(e *Organization, domain string, active bool) error
	SetPrimaryDomain(e *Organization, domain string) error
	CreateSampleData(org *Organization) error
}

// GroupRepository covers the group methods the plugin calls.
type GroupRepository interface {
	GetOne(id string) (*Group, error)
	GetAll(organizationID string) ([]*Group, error)
	GetByName(organizationID, name string) (*Group, error)
	GetMemberUserIDs(e *Group) ([]string, error)
	AddMembers(e *Group, userIDs []string) error
	RemoveMembers(e *Group, userIDs []string) error
	Create(e *Group) error
	Update(e *Group) error
	Delete(e *Group) error
}

// BookingRepository covers the booking methods the plugin calls.
type BookingRepository interface {
	GetOne(id string) (*BookingDetails, error)
}

// SpaceRepository covers the space methods the plugin calls.
type SpaceRepository interface {
	GetOne(id string) (*Space, error)
	GetCount(organizationID string) (int, error)
}

// LocationRepository covers the location methods the plugin calls.
type LocationRepository interface {
	GetOne(id string) (*Location, error)
	GetCount(organizationID string) (int, error)
	GetTimezone(location *Location) string
}

// AuthProviderRepository covers the auth-provider methods the plugin calls.
type AuthProviderRepository interface {
	Create(e *AuthProvider) error
	Update(e *AuthProvider) error
}

// AuthStateRepository covers the auth-state methods the plugin calls.
type AuthStateRepository interface {
	Create(e *AuthState) error
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
