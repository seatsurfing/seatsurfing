package api

import "github.com/seatsurfing/seatsurfing/server/api/pluginpb"

// Conversion functions between the hand-written PluginHTTPRequest/Response/
// AdminUIMenuItem/etc. types (api.go) and their generated protobuf wire
// counterparts in pluginpb. Exported (unlike entity_grpc_convert.go's
// helpers) because both the host's plugin_grpc_client.go (package api) and
// the plugin binary's grpc_server.go (package main, in plugin-cloud-features,
// a different module importing this package) need them.

func headerValuesToProto(h map[string][]string) map[string]*pluginpb.HeaderValues {
	out := make(map[string]*pluginpb.HeaderValues, len(h))
	for k, v := range h {
		out[k] = &pluginpb.HeaderValues{Values: v}
	}
	return out
}

func headerValuesFromProto(h map[string]*pluginpb.HeaderValues) map[string][]string {
	out := make(map[string][]string, len(h))
	for k, v := range h {
		if v != nil {
			out[k] = v.Values
		}
	}
	return out
}

func PluginHTTPRequestToProto(req PluginHTTPRequest) *pluginpb.HttpRequest {
	return &pluginpb.HttpRequest{
		Method:   req.Method,
		Path:     req.Path,
		RawQuery: req.RawQuery,
		Headers:  headerValuesToProto(req.Headers),
		Body:     req.Body,
		UserId:   req.UserID,
	}
}

func PluginHTTPRequestFromProto(p *pluginpb.HttpRequest) PluginHTTPRequest {
	if p == nil {
		return PluginHTTPRequest{}
	}
	return PluginHTTPRequest{
		Method:   p.Method,
		Path:     p.Path,
		RawQuery: p.RawQuery,
		Headers:  headerValuesFromProto(p.Headers),
		Body:     p.Body,
		UserID:   p.UserId,
	}
}

func PluginHTTPResponseToProto(resp PluginHTTPResponse) *pluginpb.HttpResponse {
	return &pluginpb.HttpResponse{
		StatusCode: int32(resp.StatusCode),
		Headers:    headerValuesToProto(resp.Headers),
		Body:       resp.Body,
	}
}

func PluginHTTPResponseFromProto(p *pluginpb.HttpResponse) PluginHTTPResponse {
	if p == nil {
		return PluginHTTPResponse{}
	}
	return PluginHTTPResponse{
		StatusCode: int(p.StatusCode),
		Headers:    headerValuesFromProto(p.Headers),
		Body:       p.Body,
	}
}

func AdminUIMenuItemsToProto(items []AdminUIMenuItem) []*pluginpb.AdminUIMenuItem {
	out := make([]*pluginpb.AdminUIMenuItem, 0, len(items))
	for _, it := range items {
		out = append(out, &pluginpb.AdminUIMenuItem{
			Id:         it.ID,
			Title:      it.Title,
			Source:     it.Source,
			Visibility: it.Visibility,
			Icon:       it.Icon,
		})
	}
	return out
}

func AdminUIMenuItemsFromProto(items []*pluginpb.AdminUIMenuItem) []AdminUIMenuItem {
	out := make([]AdminUIMenuItem, 0, len(items))
	for _, it := range items {
		if it == nil {
			continue
		}
		out = append(out, AdminUIMenuItem{
			ID:         it.Id,
			Title:      it.Title,
			Source:     it.Source,
			Visibility: it.Visibility,
			Icon:       it.Icon,
		})
	}
	return out
}

func AdminWelcomeScreenToProto(s *AdminWelcomeScreen) *pluginpb.AdminWelcomeScreenReply {
	if s == nil {
		return &pluginpb.AdminWelcomeScreenReply{Present: false}
	}
	return &pluginpb.AdminWelcomeScreenReply{
		Present: true,
		Screen: &pluginpb.AdminWelcomeScreen{
			Source:            s.Source,
			SkipOnSettingTrue: s.SkipOnSettingTrue,
		},
	}
}

func AdminWelcomeScreenFromProto(p *pluginpb.AdminWelcomeScreenReply) *AdminWelcomeScreen {
	if p == nil || !p.Present || p.Screen == nil {
		return nil
	}
	return &AdminWelcomeScreen{
		Source:            p.Screen.Source,
		SkipOnSettingTrue: p.Screen.SkipOnSettingTrue,
	}
}

func PluginSettingsToProto(settings []*PluginSetting) []*pluginpb.PluginSetting {
	out := make([]*pluginpb.PluginSetting, 0, len(settings))
	for _, s := range settings {
		if s == nil {
			continue
		}
		out = append(out, &pluginpb.PluginSetting{
			Name:        s.Name,
			Value:       s.Value,
			SettingType: int32(s.SettingType),
		})
	}
	return out
}

func PluginSettingsFromProto(settings []*pluginpb.PluginSetting) []*PluginSetting {
	out := make([]*PluginSetting, 0, len(settings))
	for _, s := range settings {
		if s == nil {
			continue
		}
		out = append(out, &PluginSetting{
			Name:        s.Name,
			Value:       s.Value,
			SettingType: SettingType(s.SettingType),
		})
	}
	return out
}
