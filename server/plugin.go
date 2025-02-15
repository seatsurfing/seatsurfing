package main

type SeatsurfingPlugin interface {
	GetPublicRoutes() map[string]Route
	GetBackplaneRoutes() map[string]Route
}
