package api

type SeatsurfingPlugin interface {
	GetPublicRoutes() map[string]Route
	GetUnauthorizedRoutes() []string
	GetRepositories() []Repository
	GetAdminUIMenuItems() []AdminUIMenuItem
	OnTimer()
	OnInit()
	GetAdminWelcomeScreen() *AdminWelcomeScreen
	GetPublicSettings(organizationID string) []*PluginSetting
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
