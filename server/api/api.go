package api

type PluginHTTPRequest struct {
	Method   string
	Path     string
	RawQuery string
	Headers  map[string][]string
	Body     []byte
	UserID   string
	// Host is the Host the request was sent to (http.Request.Host); Go does
	// not keep it in Headers.
	Host string
	// RemoteAddr is the host-side http.Request.RemoteAddr.
	RemoteAddr string
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
	// GetBookingUIIntegrations returns the booking UI's own integrations this
	// plugin contributes: zero, one, or more nav items shown in the booking
	// UI's top navigation bar, each opening a panel beside the map.
	// organizationID lets a plugin hide an integration for organizations that
	// have not enabled the underlying feature (same shape as
	// GetPublicSettings).
	GetBookingUIIntegrations(organizationID string) []BookingUIIntegration
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

// BookingUIIntegration describes one nav item a plugin contributes to the
// booking UI's top navigation bar. Clicking it opens a panel beside the
// search/map view that hosts the custom element named by TagName, mounted
// from the JS module at Source (same mechanism as AdminUIMenuItem).
type BookingUIIntegration struct {
	ID     string
	Title  string
	Source string
	Icon   string
	// TagName is the custom element tag to mount for this integration's panel
	// once the JS module at Source has been loaded.
	TagName string
	// Width is the panel's desired width in pixels. Zero/unset falls back to
	// a host-defined default (200px). The host clamps this to a sane range.
	Width int
	// RequiredPermission is the permission a user must hold to see this item,
	// and RequiredLevel the level at which. Empty means visible to every
	// authenticated user, since the booking UI itself has no admin gate.
	RequiredPermission Permission
	RequiredLevel      PermissionLevel
	// RequiredPermissionsAny lists alternative permissions, any one of which
	// (at RequiredLevel) is enough to see this item. Takes precedence over
	// RequiredPermission when non-empty.
	RequiredPermissionsAny []Permission
	// Titles optionally maps a language code (e.g. "de", "en-GB") to a
	// localized title. The frontend picks the entry for the signed-in user's
	// exact language, then its base language, then falls back to Title.
	Titles map[string]string
}

// PluginEmbed custom elements mounted for a BookingUIIntegration (and for
// AdminUIMenuItem/AdminWelcomeScreen) may dispatch three DOM CustomEvents
// that the host listens for:
//
//   - "plugin-token-expired": the element's access token was rejected by the
//     backend. The host refreshes it (reusing the same refresh-token flow and
//     mutex as its own requests) and sets the new value as the element's
//     accessToken property, or falls back to the host's normal
//     session-expired handling if the refresh itself fails.
//   - "plugin-data-changed": booking data changed as a side effect of using
//     the integration (e.g. a chat assistant created or cancelled a
//     booking). The booking UI reloads the current search results in
//     response, keeping the selected location, date and view.
//   - "plugin-close": the plugin has its own close control (e.g. a chat
//     assistant's close button) and wants the host to close the panel it's
//     embedded in. For a BookingUIIntegration this is the same as clicking
//     the integration's own nav item again.
//
// The host also proactively pushes a refreshed access token to every mounted
// element whenever it refreshes its own (see ui/src/util/Ajax.ts and
// ui/src/components/PluginEmbed.tsx), so "plugin-token-expired" only matters
// for a token that expires between those refreshes.

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
