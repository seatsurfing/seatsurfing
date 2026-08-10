package router

import "sync"

var unauthorizedRoutes = []string{
	"/auth/",
	"/organization/domain/",
	"/organization/deleteorg/",
	"/auth-provider/org/",
	"/admin/",
	"/ui/",
	"/confluence",
	"/robots.txt",
	"/healthcheck/",
	"/kiosk/",
}

var unauthorizedRoutesMu sync.RWMutex

// AddUnauthorizedRoutes must be safe to call more than once with the same
// routes: under the connect-driven plugin lifecycle it is invoked on every
// plugin (re)connection, not just once at host startup (a plugin restarting
// independently re-registers on reconnect). Duplicate entries are silently
// skipped rather than appended again.
func AddUnauthorizedRoutes(routes []string) {
	unauthorizedRoutesMu.Lock()
	defer unauthorizedRoutesMu.Unlock()
	for _, r := range routes {
		exists := false
		for _, existing := range unauthorizedRoutes {
			if existing == r {
				exists = true
				break
			}
		}
		if !exists {
			unauthorizedRoutes = append(unauthorizedRoutes, r)
		}
	}
}

func getUnauthorizedRoutes() []string {
	unauthorizedRoutesMu.RLock()
	defer unauthorizedRoutesMu.RUnlock()
	out := make([]string, len(unauthorizedRoutes))
	copy(out, unauthorizedRoutes)
	return out
}
