package api

import "testing"

func TestPluginHTTPRequestRoundTrip(t *testing.T) {
	req := PluginHTTPRequest{
		Method:   "POST",
		Path:     "/subscription/webhook",
		RawQuery: "a=b",
		Headers:  map[string][]string{"Content-Type": {"application/json"}, "X-Multi": {"a", "b"}},
		Body:     []byte(`{"ok":true}`),
		UserID:   "u1",
	}
	got := PluginHTTPRequestFromProto(PluginHTTPRequestToProto(req))
	if got.Method != req.Method || got.Path != req.Path || got.RawQuery != req.RawQuery || got.UserID != req.UserID {
		t.Fatalf("round-trip mismatch:\n got=%+v\nwant=%+v", got, req)
	}
	if string(got.Body) != string(req.Body) {
		t.Errorf("Body mismatch: got=%s want=%s", got.Body, req.Body)
	}
	if len(got.Headers["X-Multi"]) != 2 || got.Headers["X-Multi"][0] != "a" || got.Headers["X-Multi"][1] != "b" {
		t.Errorf("multi-value header mismatch: got=%+v", got.Headers)
	}
}

func TestPluginHTTPRequestFromProtoNil(t *testing.T) {
	got := PluginHTTPRequestFromProto(nil)
	if got.Method != "" || got.Headers != nil && len(got.Headers) != 0 {
		t.Errorf("expected zero value, got %+v", got)
	}
}

func TestPluginHTTPResponseRoundTrip(t *testing.T) {
	resp := PluginHTTPResponse{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"text/plain"}},
		Body:       []byte("hello"),
	}
	got := PluginHTTPResponseFromProto(PluginHTTPResponseToProto(resp))
	if got.StatusCode != resp.StatusCode || string(got.Body) != string(resp.Body) {
		t.Fatalf("round-trip mismatch:\n got=%+v\nwant=%+v", got, resp)
	}
}

func TestAdminUIMenuItemsRoundTrip(t *testing.T) {
	items := []AdminUIMenuItem{
		{ID: "i1", Title: "Item 1", Source: "/s1", Visibility: "admin", Icon: "Cloud"},
		{ID: "i2", Title: "Item 2", Source: "/s2", Visibility: "spaceadmin", Icon: "Gift"},
	}
	got := AdminUIMenuItemsFromProto(AdminUIMenuItemsToProto(items))
	if len(got) != len(items) {
		t.Fatalf("length mismatch: got=%d want=%d", len(got), len(items))
	}
	for i := range items {
		if got[i] != items[i] {
			t.Errorf("item %d mismatch: got=%+v want=%+v", i, got[i], items[i])
		}
	}
}

func TestAdminWelcomeScreenRoundTrip(t *testing.T) {
	s := &AdminWelcomeScreen{Source: "/welcome.html", SkipOnSettingTrue: "skip_welcome"}
	got := AdminWelcomeScreenFromProto(AdminWelcomeScreenToProto(s))
	if got == nil || *got != *s {
		t.Fatalf("round-trip mismatch: got=%+v want=%+v", got, s)
	}
}

func TestAdminWelcomeScreenNil(t *testing.T) {
	got := AdminWelcomeScreenFromProto(AdminWelcomeScreenToProto(nil))
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestPluginSettingsRoundTrip(t *testing.T) {
	settings := []*PluginSetting{
		{Name: "cloud_hosted", Value: "1", SettingType: SettingTypeBool},
		{Name: "max_users", Value: "10", SettingType: SettingTypeInt},
	}
	got := PluginSettingsFromProto(PluginSettingsToProto(settings))
	if len(got) != len(settings) {
		t.Fatalf("length mismatch: got=%d want=%d", len(got), len(settings))
	}
	for i := range settings {
		if *got[i] != *settings[i] {
			t.Errorf("setting %d mismatch: got=%+v want=%+v", i, got[i], settings[i])
		}
	}
}
