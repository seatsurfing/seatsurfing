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
	GetSettingsRepository().Set(org.ID, SettingFeatureExchangeIntegration.Name, "1")
	loginResponse := LoginTestUser(admin.ID)

	req := NewHTTPRequest("GET", "/setting/exchange/", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	var result map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &result)
	// Should return empty defaults
	if result["clientSecret"] != "" && result["clientSecret"] != nil {
		t.Fatal("clientSecret should be empty or missing in GET response")
	}
}

func TestExchangeSettingsForbiddenForUser(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("exchange-router2.com")
	user := CreateTestUserInOrg(org)
	GetSettingsRepository().Set(org.ID, SettingFeatureExchangeIntegration.Name, "1")
	loginResponse := LoginTestUser(user.ID)

	req := NewHTTPRequest("GET", "/setting/exchange/", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	payload := `{"enabled":true,"tenantId":"tid","clientId":"cid","clientSecret":"secret"}`
	req = NewHTTPRequest("PUT", "/setting/exchange/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestExchangeSettingsForbiddenWhenFeatureDisabled(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("exchange-router-disabled.com")
	admin := CreateTestUserOrgAdmin(org)
	// Feature flag NOT set -> should be forbidden even for org admin
	loginResponse := LoginTestUser(admin.ID)

	req := NewHTTPRequest("GET", "/setting/exchange/", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	payload := `{"enabled":true,"tenantId":"tid","clientId":"cid","clientSecret":"s"}`
	req = NewHTTPRequest("PUT", "/setting/exchange/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	req = NewHTTPRequest("GET", "/setting/exchange/errors/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestExchangeSettingsSetAndGet(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("exchange-router3.com")
	admin := CreateTestUserOrgAdmin(org)
	GetSettingsRepository().Set(org.ID, SettingFeatureExchangeIntegration.Name, "1")
	loginResponse := LoginTestUser(admin.ID)

	payload := `{"enabled":true,"tenantId":"my-tenant","clientId":"my-client","clientSecret":"my-secret"}`
	req := NewHTTPRequest("PUT", "/setting/exchange/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	req = NewHTTPRequest("GET", "/setting/exchange/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	var result map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &result)
	if result["tenantId"] != "my-tenant" {
		t.Fatalf("Expected tenantId to be 'my-tenant', got '%v'", result["tenantId"])
	}
	if result["clientId"] != "my-client" {
		t.Fatalf("Expected clientId to be 'my-client', got '%v'", result["clientId"])
	}
	// clientSecret must NOT be returned
	if result["clientSecret"] != "" && result["clientSecret"] != nil {
		t.Fatal("clientSecret should be empty in GET response")
	}
	if result["enabled"] != true {
		t.Fatal("enabled should be true")
	}
}

func TestExchangeSettingsPreservesSecretOnUpdate(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("exchange-router4.com")
	admin := CreateTestUserOrgAdmin(org)
	GetSettingsRepository().Set(org.ID, SettingFeatureExchangeIntegration.Name, "1")
	loginResponse := LoginTestUser(admin.ID)

	// Initial save with secret
	payload := `{"enabled":true,"tenantId":"tid","clientId":"cid","clientSecret":"original"}`
	req := NewHTTPRequest("PUT", "/setting/exchange/", loginResponse.UserID, bytes.NewBufferString(payload))
	ExecuteTestRequest(req)

	// Update without providing clientSecret -> should preserve original
	payload2 := `{"enabled":false,"tenantId":"tid2","clientId":"cid2"}`
	req = NewHTTPRequest("PUT", "/setting/exchange/", loginResponse.UserID, bytes.NewBufferString(payload2))
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

func TestExchangeSpaceMappingEndpoints(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("exchange-space-ep.com")
	admin := CreateTestUserOrgAdmin(org)
	GetSettingsRepository().Set(org.ID, SettingFeatureExchangeIntegration.Name, "1")
	loginResponse := LoginTestUser(admin.ID)
	_, space := CreateTestLocationAndSpace(org)

	// GET should return empty
	req := NewHTTPRequest("GET", "/location/"+space.LocationID+"/space/"+space.ID+"/exchangemapping", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	// PUT to set mapping
	payload := `{"roomEmail":"room@contoso.com"}`
	req = NewHTTPRequest("PUT", "/location/"+space.LocationID+"/space/"+space.ID+"/exchangemapping", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// GET should return the room email
	req = NewHTTPRequest("GET", "/location/"+space.LocationID+"/space/"+space.ID+"/exchangemapping", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var result map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &result)
	if result["roomEmail"] != "room@contoso.com" {
		t.Fatalf("Expected roomEmail to be 'room@contoso.com', got '%v'", result["roomEmail"])
	}
}
