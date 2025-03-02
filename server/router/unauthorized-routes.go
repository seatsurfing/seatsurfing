package router

import (
	"sync"

	"github.com/seatsurfing/seatsurfing/server/plugin"
)

var unauthorizedRoutes = []string{
	"/auth/",
	"/organization/domain/",
	"/auth-provider/org/",
	"/signup/",
	"/admin/",
	"/ui/",
	"/confluence",
	"/booking/debugtimeissues/",
	"/robots.txt",
}

var unauthorizedRoutesOnce sync.Once

func getUnauthorizedRoutes() []string {
	unauthorizedRoutesOnce.Do(func() {
		for _, plg := range plugin.GetPlugins() {
			unauthorizedRoutes = append(unauthorizedRoutes, (*plg).GetUnauthorizedRoutes()...)
		}
	})
	return unauthorizedRoutes
}
