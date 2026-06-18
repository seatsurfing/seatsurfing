package pluginapi

import "github.com/seatsurfing/seatsurfing/server/repository"

// Re-export SettingName variable declarations from the repository package so
// the plugin binary does not need to import repository directly.
var (
	SettingShowNames             = repository.SettingShowNames
	SettingFeatureNoUserLimit    = repository.SettingFeatureNoUserLimit
	SettingFeatureCustomDomains  = repository.SettingFeatureCustomDomains
	SettingFeatureGroups         = repository.SettingFeatureGroups
	SettingFeatureAuthProviders  = repository.SettingFeatureAuthProviders
	SettingFeatureRecurringBookings = repository.SettingFeatureRecurringBookings
	SettingFeatureKioskMode      = repository.SettingFeatureKioskMode
)
