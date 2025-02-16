package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestAuthProvidersEmptyResult(t *testing.T) {
	ClearTestDB()
	user := CreateTestUserSuperAdmin()
	loginResponse := LoginTestUser(user.ID)

	req := NewHTTPRequest("GET", "/auth-provider/", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []string
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 0 {
		t.Fatalf("Expected empty array")
	}
}

func TestAuthProvidersForbidden(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	userAdmin := CreateTestUserOrgAdmin(org)
	loginResponseAdmin := LoginTestUser(userAdmin.ID)
	user := CreateTestUserInOrg(org)
	loginResponse := LoginTestUser(user.ID)

	payload := `{"name": "Test", "providerType": 1, "clientId": "test1", "clientSecret": "test2", "authUrl": "http://test.com/1", "tokenUrl": "http://test.com/2", "authStyle": 0, "scopes": "http://test.com/3", "userInfoUrl": "http://test.com/userinfo", "userInfoEmailField": "email"}`
	req := NewHTTPRequest("POST", "/auth-provider/", loginResponseAdmin.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-Id")

	req = NewHTTPRequest("GET", "/auth-provider/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	req = NewHTTPRequest("POST", "/auth-provider/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	req = NewHTTPRequest("DELETE", "/auth-provider/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	req = NewHTTPRequest("PUT", "/auth-provider/"+id, loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	req = NewHTTPRequest("GET", "/auth-provider/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestAuthProvidersCRUD(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	userAdmin := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(userAdmin.ID)

	// 1. Create
	payload := `{"name": "Test", "providerType": 1, "clientId": "test1", "clientSecret": "test2", "authUrl": "http://test.com/1", "tokenUrl": "http://test.com/2", "authStyle": 0, "scopes": "http://test.com/3", "userInfoUrl": "http://test.com/userinfo", "userInfoEmailField": "email"}`
	req := NewHTTPRequest("POST", "/auth-provider/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-Id")

	// 2. Read
	req = NewHTTPRequest("GET", "/auth-provider/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetAuthProviderResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestString(t, "Test", resBody.Name)
	CheckTestString(t, "test1", resBody.ClientID)
	CheckTestString(t, "test2", resBody.ClientSecret)
	CheckTestString(t, "http://test.com/1", resBody.AuthURL)
	CheckTestString(t, "http://test.com/2", resBody.TokenURL)
	CheckTestInt(t, 0, resBody.AuthStyle)
	CheckTestString(t, "http://test.com/3", resBody.Scopes)
	CheckTestString(t, org.ID, resBody.OrganizationID)
	CheckTestString(t, "http://test.com/userinfo", resBody.UserInfoURL)
	CheckTestString(t, "email", resBody.UserInfoEmailField)
	CheckTestInt(t, int(OAuth2), resBody.ProviderType)

	// 3. Update
	payload = `{"name": "Test_2", "providerType": 1, "clientId": "test1_2", "clientSecret": "test2_2", "authUrl": "http://test.com/1_2", "tokenUrl": "http://test.com/2_2", "authStyle": 1, "scopes": "http://test.com/3_2", "userInfoUrl": "http://test.com/userinfo_2", "userInfoEmailField": "email_2"}`
	req = NewHTTPRequest("PUT", "/auth-provider/"+id, loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Read
	req = NewHTTPRequest("GET", "/auth-provider/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody2 *GetAuthProviderResponse
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	CheckTestString(t, "Test_2", resBody2.Name)
	CheckTestString(t, "test1_2", resBody2.ClientID)
	CheckTestString(t, "test2_2", resBody2.ClientSecret)
	CheckTestString(t, "http://test.com/1_2", resBody2.AuthURL)
	CheckTestString(t, "http://test.com/2_2", resBody2.TokenURL)
	CheckTestInt(t, 1, resBody2.AuthStyle)
	CheckTestString(t, "http://test.com/3_2", resBody2.Scopes)
	CheckTestString(t, org.ID, resBody2.OrganizationID)
	CheckTestString(t, "http://test.com/userinfo_2", resBody2.UserInfoURL)
	CheckTestString(t, "email_2", resBody2.UserInfoEmailField)
	CheckTestInt(t, int(OAuth2), resBody2.ProviderType)

	// 4. Delete
	req = NewHTTPRequest("DELETE", "/auth-provider/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Read
	req = NewHTTPRequest("GET", "/auth-provider/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestAuthProvidersGetPublicForOrg(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	userAdmin := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(userAdmin.ID)

	// Create 1
	payload := `{"name": "Test", "providerType": 1, "clientId": "test1", "clientSecret": "test2", "authUrl": "http://test.com/1", "tokenUrl": "http://test.com/2", "authStyle": 0, "scopes": "http://test.com/3", "userInfoUrl": "http://test.com/userinfo", "userInfoEmailField": "email"}`
	req := NewHTTPRequest("POST", "/auth-provider/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id1 := res.Header().Get("X-Object-Id")

	// Create 2
	payload = `{"name": "Test2", "providerType": 2, "clientId": "test2", "clientSecret": "test3", "authUrl": "http://test.com/7", "tokenUrl": "http://test.com/8", "authStyle": 0, "scopes": "http://test.com/9", "userInfoUrl": "http://test.com/userinfo", "userInfoEmailField": "email"}`
	req = NewHTTPRequest("POST", "/auth-provider/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id2 := res.Header().Get("X-Object-Id")

	// Get Public List
	req = NewHTTPRequest("GET", "/auth-provider/org/"+org.ID, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []*GetAuthProviderPublicResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 2 {
		t.Fatalf("Expected array with 2 elements")
	}
	CheckTestString(t, id1, resBody[0].ID)
	CheckTestString(t, "Test", resBody[0].Name)
	CheckTestString(t, id2, resBody[1].ID)
	CheckTestString(t, "Test2", resBody[1].Name)
}
