package api

type SeatsurfingPlugin interface {
	GetPublicRoutes() map[string]Route
	GetBackplaneRoutes() map[string]Route
	GetRepositories() []Repository
}
