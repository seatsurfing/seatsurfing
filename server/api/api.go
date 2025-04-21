package api

type SeatsurfingPlugin interface {
	GetPublicRoutes() map[string]Route
	GetUnauthorizedRoutes() []string
	GetRepositories() []Repository
	GetAdminUIMenuItems() []AdminUIMenuItem
	OnTimer()
	OnInit()
	GetAdminWelcomeScreen() *AdminWelcomeScreen
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
