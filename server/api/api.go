package api

type SeatsurfingPlugin interface {
	GetPublicRoutes() map[string]Route
	GetRepositories() []Repository
}
