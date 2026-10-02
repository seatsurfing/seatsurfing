package test

import (
	. "github.com/seatsurfing/seatsurfing/server/api"
)

// fakeBookingUIPlugin is a minimal SeatsurfingPlugin used to exercise the
// booking UI integrations pipeline (permission filtering, width clamping)
// without a real out-of-process plugin. Every method besides
// GetBookingUIIntegrations is a no-op.
//
// Register it with api.RegisterPlugin and always defer
// api.ResetPluginsForTest(): the registry has no supported way to unregister
// a plugin in production, so tests must clean up themselves or the fake
// leaks into every other test sharing the process.
type fakeBookingUIPlugin struct {
	integrations []BookingUIIntegration
}

var _ SeatsurfingPlugin = (*fakeBookingUIPlugin)(nil)

func (p *fakeBookingUIPlugin) GetBasePath() string             { return "" }
func (p *fakeBookingUIPlugin) GetRoutePrefix() []string        { return nil }
func (p *fakeBookingUIPlugin) GetUnauthorizedRoutes() []string { return nil }
func (p *fakeBookingUIPlugin) RunSchemaUpdates()               {}
func (p *fakeBookingUIPlugin) GetAdminUIMenuItems() []AdminUIMenuItem {
	return nil
}
func (p *fakeBookingUIPlugin) GetBookingUIIntegrations(organizationID string) []BookingUIIntegration {
	return p.integrations
}
func (p *fakeBookingUIPlugin) GetPermissionDefinitions() []PermissionDefinition {
	return nil
}
func (p *fakeBookingUIPlugin) OnTimer() {}
func (p *fakeBookingUIPlugin) OnInit()  {}
func (p *fakeBookingUIPlugin) GetAdminWelcomeScreen() *AdminWelcomeScreen {
	return nil
}
func (p *fakeBookingUIPlugin) GetPublicSettings(organizationID string) []*PluginSetting {
	return nil
}
func (p *fakeBookingUIPlugin) HandleHTTPRequest(req PluginHTTPRequest) PluginHTTPResponse {
	return PluginHTTPResponse{StatusCode: 404}
}
func (p *fakeBookingUIPlugin) OnUserCreated(userID string)                      {}
func (p *fakeBookingUIPlugin) OnUserUpdated(userID string)                      {}
func (p *fakeBookingUIPlugin) OnBeforeUserDelete(userID string)                 {}
func (p *fakeBookingUIPlugin) OnOrganizationCreated(organizationID string)      {}
func (p *fakeBookingUIPlugin) OnOrganizationUpdated(organizationID string)      {}
func (p *fakeBookingUIPlugin) OnBeforeOrganizationDelete(organizationID string) {}
func (p *fakeBookingUIPlugin) OnBookingCreated(bookingID string)                {}
func (p *fakeBookingUIPlugin) OnBookingUpdated(bookingID string)                {}
func (p *fakeBookingUIPlugin) OnBookingDeleted(bookingID string)                {}
