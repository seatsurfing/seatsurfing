package router

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

func AddUnauthorizedRoutes(routes []string) {
	unauthorizedRoutes = append(unauthorizedRoutes, routes...)
}

func getUnauthorizedRoutes() []string {
	return unauthorizedRoutes
}
