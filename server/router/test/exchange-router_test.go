package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestExchangeSettingsGetEmpty(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("exchange-router.com")
	admin := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(admin.ID)

	// Exchange settings not yet set → 404 from settings endpoint
	req := NewHTTPRequest("GET", "/setting/"+SettingExchangeEnabled.Name, loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)

	// Secret not yet set → masking returns ""
	req = NewHTTPRequest("GET", "/setting/"+SettingExchangeClientSecret.Name, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var secretVal string
	json.Unmarshal(res.Body.Bytes(), &secretVal)
	if secretVal != "" {
		t.Fatalf("Expected empty secret indicator, got '%s'", secretVal)
	}
}

func TestExchangeSettingsForbiddenForUser(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("exchange-router2.com")
	user := CreateTestUserInOrg(org)
	loginResponse := LoginTestUser(user.ID)

	// Regular user cannot read admin-only exchange settings
	req := NewHTTPRequest("GET", "/setting/"+SettingExchangeEnabled.Name, loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	// Regular user cannot write exchange settings
	payload := `{"value":"1"}`
	req = NewHTTPRequest("PUT", "/setting/"+SettingExchangeEnabled.Name, loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestExchangeSettingsForbiddenWhenFeatureDisabled(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("exchange-router-disabled.com")
	admin := CreateTestUserOrgAdmin(org)
	// Feature flag NOT set → errors endpoint should be forbidden for org admin
	loginResponse := LoginTestUser(admin.ID)

	req := NewHTTPRequest("GET", "/setting/exchange/errors/", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestExchangeSettingsSetAndGet(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("exchange-router3.com")
	admin := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(admin.ID)

	// Write individual settings via standard settings endpoint
	for name, value := range map[string]string{
		SettingExchangeEnabled.Name:  "1",
		SettingExchangeTenantID.Name: "my-tenant",
		SettingExchangeClientID.Name: "my-client",
	} {
		payload := `{"value":"` + value + `"}`
		req := NewHTTPRequest("PUT", "/setting/"+name, loginResponse.UserID, bytes.NewBufferString(payload))
		res := ExecuteTestRequest(req)
		CheckTestResponseCode(t, http.StatusNoContent, res.Code)
	}
	secretPayload := `{"value":"my-secret"}`
	req := NewHTTPRequest("PUT", "/setting/"+SettingExchangeClientSecret.Name, loginResponse.UserID, bytes.NewBufferString(secretPayload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Read back via standard settings endpoint
	req = NewHTTPRequest("GET", "/setting/"+SettingExchangeTenantID.Name, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var tenantVal string
	json.Unmarshal(res.Body.Bytes(), &tenantVal)
	if tenantVal != "my-tenant" {
		t.Fatalf("Expected tenantId 'my-tenant', got '%s'", tenantVal)
	}

	// Secret must be masked: returns "1" when set, never the real value
	req = NewHTTPRequest("GET", "/setting/"+SettingExchangeClientSecret.Name, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var secretVal string
	json.Unmarshal(res.Body.Bytes(), &secretVal)
	if secretVal != "1" {
		t.Fatalf("Expected secret indicator '1', got '%s'", secretVal)
	}
}

func TestExchangeSettingsPreservesSecretOnUpdate(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("exchange-router4.com")
	admin := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(admin.ID)

	// Write initial secret via standard settings endpoint
	payload := `{"value":"original"}`
	req := NewHTTPRequest("PUT", "/setting/"+SettingExchangeClientSecret.Name, loginResponse.UserID, bytes.NewBufferString(payload))
	ExecuteTestRequest(req)

	// Write empty secret → should preserve existing
	payload2 := `{"value":""}`
	req = NewHTTPRequest("PUT", "/setting/"+SettingExchangeClientSecret.Name, loginResponse.UserID, bytes.NewBufferString(payload2))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Verify secret is preserved in DB
	stored, err := GetSettingsRepository().GetExchangeSettings(org.ID)
	if err != nil {
		t.Fatal("Failed to get settings from repo:", err)
	}
	if stored.ClientSecret != "original" {
		t.Fatalf("Expected preserved secret 'original', got '%s'", stored.ClientSecret)
	}
}

func TestExchangeGetErrorsEmpty(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("exchange-errors.com")
	admin := CreateTestUserOrgAdmin(org)
	GetSettingsRepository().Set(org.ID, SettingFeatureExchangeIntegration.Name, "1")
	loginResponse := LoginTestUser(admin.ID)

	req := NewHTTPRequest("GET", "/setting/exchange/errors/", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
}

func TestExchangeSpaceRoomEmailCRUD(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("exchange-space-ep.com")
	admin := CreateTestUserOrgAdmin(org)
	GetSettingsRepository().Set(org.ID, SettingFeatureExchangeIntegration.Name, "1")
	loginResponse := LoginTestUser(admin.ID)
	_, space := CreateTestLocationAndSpace(org)

	// GET should return empty roomEmail initially
	req := NewHTTPRequest("GET", "/location/"+space.LocationID+"/space/"+space.ID, loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var result map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &result)
	if result["roomEmail"] != "" {
		t.Fatalf("Expected roomEmail to be empty, got '%v'", result["roomEmail"])
	}

	// PUT to update space with room email
	payload := `{"name":"Test Space","x":10,"y":10,"width":100,"height":100,"rotation":0,"requireSubject":false,"enabled":true,"kioskEnabled":false,"roomEmail":"room@contoso.com"}`
	req = NewHTTPRequest("PUT", "/location/"+space.LocationID+"/space/"+space.ID, loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// GET should now return the room email
	req = NewHTTPRequest("GET", "/location/"+space.LocationID+"/space/"+space.ID, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	json.Unmarshal(res.Body.Bytes(), &result)
	if result["roomEmail"] != "room@contoso.com" {
		t.Fatalf("Expected roomEmail to be 'room@contoso.com', got '%v'", result["roomEmail"])
	}
}
