package api

type SeatsurfingPlugin interface {
	GetPublicRoutes() map[string]Route
	GetUnauthorizedRoutes() []string
	GetRepositories() []Repository
	GetAdminUIMenuItems() []AdminUIMenuItem
}

type AdminUIMenuItem struct {
	ID         string
	Title      string
	Source     string
	Visibility string // "admin", "spaceadmin"
	Icon       string
}
