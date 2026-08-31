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
	// GetBasePath returns the single canonical URL prefix under which all of
	// this plugin's endpoints are nested
	GetBasePath() string
	// GetRoutePrefix returns the plugin's legacy flat top-level prefixes,
	// kept mounted alongside the base path for backward compatibility with
	// externally-configured URLs
	GetRoutePrefix() []string
	GetUnauthorizedRoutes() []string
	RunSchemaUpdates()
	GetAdminUIMenuItems() []AdminUIMenuItem
	// GetPermissionDefinitions returns the permissions this plugin
	// contributes to the role editor. Keys must begin with
	// PluginPermissionPrefix. Returning none is valid: the plugin's menu items
	// then fall back to the coarse Visibility field.
	GetPermissionDefinitions() []PermissionDefinition
	OnTimer()
	// Implementations must be safe to call more than once: the host
	// re-invokes OnInit on every reconnection, not only once at startup.
	OnInit()
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
	ID     string
	Title  string
	Source string
	// Deprecated: superseded by RequiredPermission. Retained so that a plugin
	// built against the old contract still places its items correctly:
	// "admin" maps to org_settings at admin level, "spaceadmin" to holding any
	// permission at all.
	Visibility string
	Icon       string
	// TagName is the custom element tag to mount for this menu item's UI
	// once the JS module at Source has been loaded.
	TagName string
	// RequiredPermission is the permission a user must hold to see this item,
	// and RequiredLevel the level at which. An empty RequiredPermission falls
	// back to Visibility.
	RequiredPermission Permission
	RequiredLevel      PermissionLevel
	// RequiredPermissionsAny lists alternative permissions, any one of which
	// (at RequiredLevel) is enough to see this item. Takes precedence over
	// RequiredPermission when non-empty, for a plugin whose single UI surface
	// spans several independently-grantable permissions.
	RequiredPermissionsAny []Permission
}

type AdminWelcomeScreen struct {
	Source            string
	SkipOnSettingTrue string
	TagName           string
	// RequiredPermission is the permission a user must hold to see this
	// screen, and RequiredLevel the level at which. An empty
	// RequiredPermission falls back to requiring org_settings at admin level.
	RequiredPermission Permission
	RequiredLevel      PermissionLevel
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
