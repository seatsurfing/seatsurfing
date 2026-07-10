package api

import (
	"context"
	"log"
	"time"

	"github.com/seatsurfing/seatsurfing/server/api/commonpb"
	"github.com/seatsurfing/seatsurfing/server/api/pluginpb"
	"google.golang.org/grpc"
)

// PluginGRPC runs in the HOST process. It implements SeatsurfingPlugin by
// forwarding every call to a plugin process over gRPC. The SeatsurfingPlugin
// interface has no context.Context parameter (kept as-is to avoid touching
// every call site), so PluginGRPC synthesizes a deadline per call from
// callTimeout - without one, a hung plugin across the network would block
// the calling host goroutine indefinitely.
type PluginGRPC struct {
	conn        *grpc.ClientConn
	client      pluginpb.SeatsurfingPluginServiceClient
	callTimeout time.Duration
}

func NewPluginGRPC(conn *grpc.ClientConn, callTimeout time.Duration) *PluginGRPC {
	return &PluginGRPC{
		conn:        conn,
		client:      pluginpb.NewSeatsurfingPluginServiceClient(conn),
		callTimeout: callTimeout,
	}
}

var _ SeatsurfingPlugin = (*PluginGRPC)(nil)

func (p *PluginGRPC) ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), p.callTimeout)
}

func (p *PluginGRPC) GetRoutePrefix() []string {
	ctx, cancel := p.ctx()
	defer cancel()
	resp, err := p.client.GetRoutePrefix(ctx, &commonpb.Empty{})
	if err != nil {
		return []string{}
	}
	return resp.Values
}

func (p *PluginGRPC) GetUnauthorizedRoutes() []string {
	ctx, cancel := p.ctx()
	defer cancel()
	resp, err := p.client.GetUnauthorizedRoutes(ctx, &commonpb.Empty{})
	if err != nil {
		return []string{}
	}
	return resp.Values
}

// RunSchemaUpdates deliberately has no error return on the SeatsurfingPlugin
// interface, but its *caller* in the connect-driven lifecycle watcher needs
// to know about failure so it can keep the plugin instance not-ready rather
// than silently treating a failed migration as success. RunSchemaUpdatesErr
// exposes that.
func (p *PluginGRPC) RunSchemaUpdates() {
	if err := p.RunSchemaUpdatesErr(); err != nil {
		log.Println("RunSchemaUpdates RPC error:", err)
	}
}

func (p *PluginGRPC) RunSchemaUpdatesErr() error {
	// Schema migrations can legitimately take longer than the default
	// per-call timeout; give this one more room.
	ctx, cancel := context.WithTimeout(context.Background(), p.callTimeout*4)
	defer cancel()
	_, err := p.client.RunSchemaUpdates(ctx, &commonpb.Empty{})
	return err
}

func (p *PluginGRPC) GetAdminUIMenuItems() []AdminUIMenuItem {
	ctx, cancel := p.ctx()
	defer cancel()
	resp, err := p.client.GetAdminUIMenuItems(ctx, &commonpb.Empty{})
	if err != nil {
		log.Println("GetAdminUIMenuItems RPC error:", err)
		return []AdminUIMenuItem{}
	}
	return AdminUIMenuItemsFromProto(resp.Items)
}

func (p *PluginGRPC) OnTimer() {
	ctx, cancel := p.ctx()
	defer cancel()
	if _, err := p.client.OnTimer(ctx, &commonpb.Empty{}); err != nil {
		log.Println("OnTimer RPC error:", err)
	}
}

// OnInit must be safe to call more than once - the host re-invokes it on
// every reconnection (see OnInitErr).
func (p *PluginGRPC) OnInit() {
	if err := p.OnInitErr(); err != nil {
		log.Println("OnInit RPC error:", err)
	}
}

func (p *PluginGRPC) OnInitErr() error {
	ctx, cancel := p.ctx()
	defer cancel()
	_, err := p.client.OnInit(ctx, &commonpb.Empty{})
	return err
}

func (p *PluginGRPC) GetAdminWelcomeScreen() *AdminWelcomeScreen {
	ctx, cancel := p.ctx()
	defer cancel()
	resp, err := p.client.GetAdminWelcomeScreen(ctx, &commonpb.Empty{})
	if err != nil {
		log.Println("GetAdminWelcomeScreen RPC error:", err)
		return nil
	}
	return AdminWelcomeScreenFromProto(resp)
}

func (p *PluginGRPC) GetPublicSettings(organizationID string) []*PluginSetting {
	ctx, cancel := p.ctx()
	defer cancel()
	resp, err := p.client.GetPublicSettings(ctx, &pluginpb.GetPublicSettingsRequest{OrganizationId: organizationID})
	if err != nil {
		return []*PluginSetting{}
	}
	return PluginSettingsFromProto(resp.Settings)
}

func (p *PluginGRPC) HandleHTTPRequest(req PluginHTTPRequest) PluginHTTPResponse {
	ctx, cancel := p.ctx()
	defer cancel()
	resp, err := p.client.HandleHTTPRequest(ctx, PluginHTTPRequestToProto(req))
	if err != nil {
		return PluginHTTPResponse{StatusCode: 502}
	}
	return PluginHTTPResponseFromProto(resp)
}

func (p *PluginGRPC) onIDHook(name string, id string, call func(ctx context.Context, req *pluginpb.IdRequest, opts ...grpc.CallOption) (*commonpb.Empty, error)) {
	ctx, cancel := p.ctx()
	defer cancel()
	if _, err := call(ctx, &pluginpb.IdRequest{Id: id}); err != nil {
		log.Printf("%s RPC error: %v", name, err)
	}
}

func (p *PluginGRPC) OnUserCreated(userID string) {
	p.onIDHook("OnUserCreated", userID, p.client.OnUserCreated)
}
func (p *PluginGRPC) OnUserUpdated(userID string) {
	p.onIDHook("OnUserUpdated", userID, p.client.OnUserUpdated)
}
func (p *PluginGRPC) OnBeforeUserDelete(userID string) {
	p.onIDHook("OnBeforeUserDelete", userID, p.client.OnBeforeUserDelete)
}
func (p *PluginGRPC) OnOrganizationCreated(organizationID string) {
	p.onIDHook("OnOrganizationCreated", organizationID, p.client.OnOrganizationCreated)
}
func (p *PluginGRPC) OnOrganizationUpdated(organizationID string) {
	p.onIDHook("OnOrganizationUpdated", organizationID, p.client.OnOrganizationUpdated)
}
func (p *PluginGRPC) OnBeforeOrganizationDelete(organizationID string) {
	p.onIDHook("OnBeforeOrganizationDelete", organizationID, p.client.OnBeforeOrganizationDelete)
}
func (p *PluginGRPC) OnBookingCreated(bookingID string) {
	p.onIDHook("OnBookingCreated", bookingID, p.client.OnBookingCreated)
}
func (p *PluginGRPC) OnBookingUpdated(bookingID string) {
	p.onIDHook("OnBookingUpdated", bookingID, p.client.OnBookingUpdated)
}
func (p *PluginGRPC) OnBookingDeleted(bookingID string) {
	p.onIDHook("OnBookingDeleted", bookingID, p.client.OnBookingDeleted)
}
