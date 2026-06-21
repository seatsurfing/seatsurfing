package api

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

type PluginHTTPRequest struct {
	Method   string
	Path     string
	RawQuery string
	Headers  map[string][]string
	Body     []byte
	UserID   string
}

type PluginHTTPResponse struct {
	StatusCode int
	Headers    map[string][]string
	Body       []byte
}

type SeatsurfingPlugin interface {
	GetRoutePrefix() []string
	GetUnauthorizedRoutes() []string
	RunSchemaUpdates()
	GetAdminUIMenuItems() []AdminUIMenuItem
	OnTimer()
	OnInit(hostAPIBrokerID uint32)
	GetAdminWelcomeScreen() *AdminWelcomeScreen
	GetPublicSettings(organizationID string) []*PluginSetting
	HandleHTTPRequest(req PluginHTTPRequest) PluginHTTPResponse
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
	Client *rpc.Client
	Broker *plugin.MuxBroker
}

func (p *PluginRPC) GetRoutePrefix() []string {
	var resp []string
	err := p.Client.Call("Plugin.GetRoutePrefix", new(any), &resp)
	if err != nil {
		return []string{}
	}
	return resp
}

func (p *PluginRPC) GetUnauthorizedRoutes() []string {
	var resp []string
	err := p.Client.Call("Plugin.GetUnauthorizedRoutes", new(any), &resp)
	if err != nil {
		return []string{}
	}
	return resp
}

func (p *PluginRPC) RunSchemaUpdates() {
	p.Client.Call("Plugin.RunSchemaUpdates", new(any), new(any))
}

func (p *PluginRPC) GetAdminUIMenuItems() []AdminUIMenuItem {
	var resp []AdminUIMenuItem
	err := p.Client.Call("Plugin.GetAdminUIMenuItems", new(any), &resp)
	if err != nil {
		return []AdminUIMenuItem{}
	}
	return resp
}

func (p *PluginRPC) OnTimer() {
	p.Client.Call("Plugin.OnTimer", new(any), new(any))
}

func (p *PluginRPC) OnInit(brokerID uint32) {
	p.Client.Call("Plugin.OnInit", brokerID, new(any))
}

func (p *PluginRPC) GetAdminWelcomeScreen() *AdminWelcomeScreen {
	var resp AdminWelcomeScreen
	err := p.Client.Call("Plugin.GetAdminWelcomeScreen", new(any), &resp)
	if err != nil {
		return nil
	}
	return &resp
}

func (p *PluginRPC) GetPublicSettings(organizationID string) []*PluginSetting {
	var resp []*PluginSetting
	err := p.Client.Call("Plugin.GetPublicSettings", organizationID, &resp)
	if err != nil {
		return []*PluginSetting{}
	}
	return resp
}

func (p *PluginRPC) HandleHTTPRequest(req PluginHTTPRequest) PluginHTTPResponse {
	var resp PluginHTTPResponse
	if err := p.Client.Call("Plugin.HandleHTTPRequest", req, &resp); err != nil {
		return PluginHTTPResponse{StatusCode: 502}
	}
	return resp
}

func (p *PluginRPC) OnUserCreated(userID string) {
	p.Client.Call("Plugin.OnUserCreated", userID, new(any))
}

func (p *PluginRPC) OnUserUpdated(userID string) {
	p.Client.Call("Plugin.OnUserUpdated", userID, new(any))
}

func (p *PluginRPC) OnBeforeUserDelete(userID string) {
	p.Client.Call("Plugin.OnBeforeUserDelete", userID, new(any))
}

func (p *PluginRPC) OnOrganizationCreated(organizationID string) {
	p.Client.Call("Plugin.OnOrganizationCreated", organizationID, new(any))
}

func (p *PluginRPC) OnOrganizationUpdated(organizationID string) {
	p.Client.Call("Plugin.OnOrganizationUpdated", organizationID, new(any))
}

func (p *PluginRPC) OnBeforeOrganizationDelete(organizationID string) {
	p.Client.Call("Plugin.OnBeforeOrganizationDelete", organizationID, new(any))
}

func (p *PluginRPC) OnBookingCreated(bookingID string) {
	p.Client.Call("Plugin.OnBookingCreated", bookingID, new(any))
}

func (p *PluginRPC) OnBookingUpdated(bookingID string) {
	p.Client.Call("Plugin.OnBookingUpdated", bookingID, new(any))
}

func (p *PluginRPC) OnBookingDeleted(bookingID string) {
	p.Client.Call("Plugin.OnBookingDeleted", bookingID, new(any))
}

type PluginRPCServer struct {
	Impl   SeatsurfingPlugin
	Broker *plugin.MuxBroker
}

func (s *PluginRPCServer) GetRoutePrefix(args any, resp *[]string) error {
	*resp = s.Impl.GetRoutePrefix()
	return nil
}

func (s *PluginRPCServer) GetUnauthorizedRoutes(args any, resp *[]string) error {
	*resp = s.Impl.GetUnauthorizedRoutes()
	return nil
}

func (s *PluginRPCServer) RunSchemaUpdates(args any, resp *any) error {
	s.Impl.RunSchemaUpdates()
	return nil
}

func (s *PluginRPCServer) GetAdminUIMenuItems(args any, resp *[]AdminUIMenuItem) error {
	*resp = s.Impl.GetAdminUIMenuItems()
	return nil
}

func (s *PluginRPCServer) OnTimer(args any, resp *any) error {
	s.Impl.OnTimer()
	return nil
}

// PluginSideBroker is set in the plugin process when Server() is called.
// The plugin binary can read this to Dial the host's HostAPI broker stream.
var PluginSideBroker *plugin.MuxBroker

func (s *PluginRPCServer) OnInit(brokerID uint32, resp *any) error {
	PluginSideBroker = s.Broker
	s.Impl.OnInit(brokerID)
	return nil
}

func (s *PluginRPCServer) GetAdminWelcomeScreen(args any, resp *AdminWelcomeScreen) error {
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

func (s *PluginRPCServer) HandleHTTPRequest(req PluginHTTPRequest, resp *PluginHTTPResponse) error {
	*resp = s.Impl.HandleHTTPRequest(req)
	return nil
}

func (s *PluginRPCServer) OnUserCreated(userID string, resp *any) error {
	s.Impl.OnUserCreated(userID)
	return nil
}

func (s *PluginRPCServer) OnUserUpdated(userID string, resp *any) error {
	s.Impl.OnUserUpdated(userID)
	return nil
}

func (s *PluginRPCServer) OnBeforeUserDelete(userID string, resp *any) error {
	s.Impl.OnBeforeUserDelete(userID)
	return nil
}

func (s *PluginRPCServer) OnOrganizationCreated(organizationID string, resp *any) error {
	s.Impl.OnOrganizationCreated(organizationID)
	return nil
}

func (s *PluginRPCServer) OnOrganizationUpdated(organizationID string, resp *any) error {
	s.Impl.OnOrganizationUpdated(organizationID)
	return nil
}

func (s *PluginRPCServer) OnBeforeOrganizationDelete(organizationID string, resp *any) error {
	s.Impl.OnBeforeOrganizationDelete(organizationID)
	return nil
}

func (s *PluginRPCServer) OnBookingCreated(bookingID string, resp *any) error {
	s.Impl.OnBookingCreated(bookingID)
	return nil
}

func (s *PluginRPCServer) OnBookingUpdated(bookingID string, resp *any) error {
	s.Impl.OnBookingUpdated(bookingID)
	return nil
}

func (s *PluginRPCServer) OnBookingDeleted(bookingID string, resp *any) error {
	s.Impl.OnBookingDeleted(bookingID)
	return nil
}

type SeatsurfingPluginImpl struct {
	Impl SeatsurfingPlugin
}

func (p *SeatsurfingPluginImpl) Server(b *plugin.MuxBroker) (any, error) {
	return &PluginRPCServer{Impl: p.Impl, Broker: b}, nil
}

func (SeatsurfingPluginImpl) Client(b *plugin.MuxBroker, c *rpc.Client) (any, error) {
	return &PluginRPC{Client: c, Broker: b}, nil
}
