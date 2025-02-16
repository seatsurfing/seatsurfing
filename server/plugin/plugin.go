package plugin

import . "github.com/seatsurfing/seatsurfing/server/router"

type SeatsurfingPlugin interface {
	GetPublicRoutes() map[string]Route
	GetBackplaneRoutes() map[string]Route
}
