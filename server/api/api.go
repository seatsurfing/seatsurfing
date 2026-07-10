package api

type PluginHTTPRequest struct {
	Method   string
	Path     string
	RawQuery string
	Headers  map[string][]string
	Body     []byte
	UserID   string
}

type PluginHTTPResponse struct {
	StatusCode int
	Headers    map[string][]string
	Body       []byte
}

type SeatsurfingPlugin interface {
	GetRoutePrefix() []string
	GetUnauthorizedRoutes() []string
	RunSchemaUpdates()
	GetAdminUIMenuItems() []AdminUIMenuItem
	OnTimer()
	// hostAPIBrokerID is unused; the plugin learns the host's HostAPI
	// address from its own static config instead. Implementations must be
	// safe to call more than once: the host re-invokes OnInit on every
	// reconnection, not only once at startup.
	OnInit(hostAPIBrokerID uint32)
	GetAdminWelcomeScreen() *AdminWelcomeScreen
	GetPublicSettings(organizationID string) []*PluginSetting
	HandleHTTPRequest(req PluginHTTPRequest) PluginHTTPResponse
	OnUserCreated(userID string)
	OnUserUpdated(userID string)
	OnBeforeUserDelete(userID string)
	OnOrganizationCreated(organizationID string)
	OnOrganizationUpdated(organizationID string)
	OnBeforeOrganizationDelete(organizationID string)
	OnBookingCreated(bookingID string)
	OnBookingUpdated(bookingID string)
	OnBookingDeleted(bookingID string)
}

type AdminUIMenuItem struct {
	ID         string
	Title      string
	Source     string
	Visibility string // "admin", "spaceadmin"
	Icon       string
}

type AdminWelcomeScreen struct {
	Source            string
	SkipOnSettingTrue string
}

type PluginSetting struct {
	Name        string
	Value       string
	SettingType SettingType
}

type SettingType int

const (
	SettingTypeInt             SettingType = 1
	SettingTypeBool            SettingType = 2
	SettingTypeString          SettingType = 3
	SettingTypeIntArray        SettingType = 4
	SettingTypeEncryptedString SettingType = 5
)
