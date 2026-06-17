package api

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

type SeatsurfingPlugin interface {
	GetPublicRoutes() map[string]Route
	GetUnauthorizedRoutes() []string
	GetRepositories() []Repository
	GetAdminUIMenuItems() []AdminUIMenuItem
	OnTimer()
	OnInit()
	GetAdminWelcomeScreen() *AdminWelcomeScreen
	GetPublicSettings(organizationID string) []*PluginSetting
	OnUserCreated(userID string)
	OnUserUpdated(userID string)
	OnBeforeUserDelete(userID string)
	OnOrganizationCreated(organizationID string)
	OnOrganizationUpdated(organizationID string)
	OnBeforeOrganizationDelete(organizationID string)
	OnBookingCreated(bookingID string)
	OnBookingUpdated(bookingID string)
	OnBookingDeleted(bookingID string)
}

type AdminUIMenuItem struct {
	ID         string
	Title      string
	Source     string
	Visibility string // "admin", "spaceadmin"
	Icon       string
}

type AdminWelcomeScreen struct {
	Source            string
	SkipOnSettingTrue string
}

type PluginSetting struct {
	Name        string
	Value       string
	SettingType SettingType
}

type SettingType int

const (
	SettingTypeInt             SettingType = 1
	SettingTypeBool            SettingType = 2
	SettingTypeString          SettingType = 3
	SettingTypeIntArray        SettingType = 4
	SettingTypeEncryptedString SettingType = 5
)

type PluginRPC struct {
	client *rpc.Client
}

func (p *PluginRPC) GetPublicRoutes() map[string]Route {
	var resp map[string]Route
	err := p.client.Call("Plugin.GetPublicRoutes", new(interface{}), &resp)
	if err != nil {
		return make(map[string]Route)
	}
	return resp
}

func (p *PluginRPC) GetUnauthorizedRoutes() []string {
	var resp []string
	err := p.client.Call("Plugin.GetUnauthorizedRoutes", new(interface{}), &resp)
	if err != nil {
		return []string{}
	}
	return resp
}

func (p *PluginRPC) GetRepositories() []Repository {
	var resp []Repository
	err := p.client.Call("Plugin.GetRepositories", new(interface{}), &resp)
	if err != nil {
		return []Repository{}
	}
	return resp
}

func (p *PluginRPC) GetAdminUIMenuItems() []AdminUIMenuItem {
	var resp []AdminUIMenuItem
	err := p.client.Call("Plugin.GetAdminUIMenuItems", new(interface{}), &resp)
	if err != nil {
		return []AdminUIMenuItem{}
	}
	return resp
}

func (p *PluginRPC) OnTimer() {
	p.client.Call("Plugin.OnTimer", new(interface{}), new(interface{}))
}

func (p *PluginRPC) OnInit() {
	p.client.Call("Plugin.OnInit", new(interface{}), new(interface{}))
}

func (p *PluginRPC) GetAdminWelcomeScreen() *AdminWelcomeScreen {
	var resp AdminWelcomeScreen
	err := p.client.Call("Plugin.GetAdminWelcomeScreen", new(interface{}), &resp)
	if err != nil {
		return nil
	}
	return &resp
}

func (p *PluginRPC) GetPublicSettings(organizationID string) []*PluginSetting {
	var resp []*PluginSetting
	err := p.client.Call("Plugin.GetPublicSettings", organizationID, &resp)
	if err != nil {
		return []*PluginSetting{}
	}
	return resp
}

func (p *PluginRPC) OnUserCreated(userID string) {
	p.client.Call("Plugin.OnUserCreated", userID, new(interface{}))
}

func (p *PluginRPC) OnUserUpdated(userID string) {
	p.client.Call("Plugin.OnUserUpdated", userID, new(interface{}))
}

func (p *PluginRPC) OnBeforeUserDelete(userID string) {
	p.client.Call("Plugin.OnBeforeUserDelete", userID, new(interface{}))
}

func (p *PluginRPC) OnOrganizationCreated(organizationID string) {
	p.client.Call("Plugin.OnOrganizationCreated", organizationID, new(interface{}))
}

func (p *PluginRPC) OnOrganizationUpdated(organizationID string) {
	p.client.Call("Plugin.OnOrganizationUpdated", organizationID, new(interface{}))
}

func (p *PluginRPC) OnBeforeOrganizationDelete(organizationID string) {
	p.client.Call("Plugin.OnBeforeOrganizationDelete", organizationID, new(interface{}))
}

func (p *PluginRPC) OnBookingCreated(bookingID string) {
	p.client.Call("Plugin.OnBookingCreated", bookingID, new(interface{}))
}

func (p *PluginRPC) OnBookingUpdated(bookingID string) {
	p.client.Call("Plugin.OnBookingUpdated", bookingID, new(interface{}))
}

func (p *PluginRPC) OnBookingDeleted(bookingID string) {
	p.client.Call("Plugin.OnBookingDeleted", bookingID, new(interface{}))
}

type PluginRPCServer struct {
	Impl SeatsurfingPlugin
}

func (s *PluginRPCServer) GetPublicRoutes(args interface{}, resp *map[string]Route) error {
	*resp = s.Impl.GetPublicRoutes()
	return nil
}

func (s *PluginRPCServer) GetUnauthorizedRoutes(args interface{}, resp *[]string) error {
	*resp = s.Impl.GetUnauthorizedRoutes()
	return nil
}

func (s *PluginRPCServer) GetRepositories(args interface{}, resp *[]Repository) error {
	*resp = s.Impl.GetRepositories()
	return nil
}

func (s *PluginRPCServer) GetAdminUIMenuItems(args interface{}, resp *[]AdminUIMenuItem) error {
	*resp = s.Impl.GetAdminUIMenuItems()
	return nil
}

func (s *PluginRPCServer) OnTimer(args interface{}, resp *interface{}) error {
	s.Impl.OnTimer()
	return nil
}

func (s *PluginRPCServer) OnInit(args interface{}, resp *interface{}) error {
	s.Impl.OnInit()
	return nil
}

func (s *PluginRPCServer) GetAdminWelcomeScreen(args interface{}, resp *AdminWelcomeScreen) error {
	result := s.Impl.GetAdminWelcomeScreen()
	if result != nil {
		*resp = *result
	}
	return nil
}

func (s *PluginRPCServer) GetPublicSettings(organizationID string, resp *[]*PluginSetting) error {
	*resp = s.Impl.GetPublicSettings(organizationID)
	return nil
}

func (s *PluginRPCServer) OnUserCreated(userID string, resp *interface{}) error {
	s.Impl.OnUserCreated(userID)
	return nil
}

func (s *PluginRPCServer) OnUserUpdated(userID string, resp *interface{}) error {
	s.Impl.OnUserUpdated(userID)
	return nil
}

func (s *PluginRPCServer) OnBeforeUserDelete(userID string, resp *interface{}) error {
	s.Impl.OnBeforeUserDelete(userID)
	return nil
}

func (s *PluginRPCServer) OnOrganizationCreated(organizationID string, resp *interface{}) error {
	s.Impl.OnOrganizationCreated(organizationID)
	return nil
}

func (s *PluginRPCServer) OnOrganizationUpdated(organizationID string, resp *interface{}) error {
	s.Impl.OnOrganizationUpdated(organizationID)
	return nil
}

func (s *PluginRPCServer) OnBeforeOrganizationDelete(organizationID string, resp *interface{}) error {
	s.Impl.OnBeforeOrganizationDelete(organizationID)
	return nil
}

func (s *PluginRPCServer) OnBookingCreated(bookingID string, resp *interface{}) error {
	s.Impl.OnBookingCreated(bookingID)
	return nil
}

func (s *PluginRPCServer) OnBookingUpdated(bookingID string, resp *interface{}) error {
	s.Impl.OnBookingUpdated(bookingID)
	return nil
}

func (s *PluginRPCServer) OnBookingDeleted(bookingID string, resp *interface{}) error {
	s.Impl.OnBookingDeleted(bookingID)
	return nil
}

type SeatsurfingPluginImpl struct {
	Impl SeatsurfingPlugin
}

func (p *SeatsurfingPluginImpl) Server(*plugin.MuxBroker) (interface{}, error) {
	return &PluginRPCServer{Impl: p.Impl}, nil
}

func (SeatsurfingPluginImpl) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &PluginRPC{client: c}, nil
}
