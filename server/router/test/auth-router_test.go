package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

func TestAuthPasswordLogin(t *testing.T) {
	ClearTestDB()

	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	user.HashedPassword = NullString(GetUserRepository().GetHashedPassword("12345678"))
	GetUserRepository().Update(user)

	// Log in
	payload := "{ \"email\": \"" + user.Email + "\", \"password\": \"12345678\", \"organizationId\": \"" + org.ID + "\" }"
	req := NewHTTPRequest("POST", "/auth/login", "", bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *JWTResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestBool(t, true, len(resBody.AccessToken) > 32)
	CheckTestBool(t, true, len(resBody.RefreshToken) == 36)

	// Test access token
	req = NewHTTPRequestWithAccessToken("GET", "/user/me", resBody.AccessToken, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody2 *GetUserResponse
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	CheckTestString(t, user.Email, resBody2.Email)
}

func TestAuthPasswordLoginBan(t *testing.T) {
	ClearTestDB()

	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	user.HashedPassword = NullString(GetUserRepository().GetHashedPassword("12345678"))
	GetUserRepository().Update(user)

	// Attempt 1
	payload := "{ \"email\": \"" + user.Email + "\", \"password\": \"12345670\", \"organizationId\": \"" + org.ID + "\" }"
	req := NewHTTPRequest("POST", "/auth/login", "", bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
	CheckTestBool(t, false, AuthAttemptRepositoryIsUserDisabled(t, user.ID))

	// Attempt 2
	payload = "{ \"email\": \"" + user.Email + "\", \"password\": \"12345679\", \"organizationId\": \"" + org.ID + "\" }"
	req = NewHTTPRequest("POST", "/auth/login", "", bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
	CheckTestBool(t, false, AuthAttemptRepositoryIsUserDisabled(t, user.ID))

	// Attempt 3
	payload = "{ \"email\": \"" + user.Email + "\", \"password\": \"12345671\", \"organizationId\": \"" + org.ID + "\" }"
	req = NewHTTPRequest("POST", "/auth/login", "", bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
	CheckTestBool(t, true, AuthAttemptRepositoryIsUserDisabled(t, user.ID))

	// Would be successful, but fails cause banned
	payload = "{ \"email\": \"" + user.Email + "\", \"password\": \"12345678\", \"organizationId\": \"" + org.ID + "\" }"
	req = NewHTTPRequest("POST", "/auth/login", "", bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
	CheckTestBool(t, true, AuthAttemptRepositoryIsUserDisabled(t, user.ID))
}

func TestAuthRefresh(t *testing.T) {
	ClearTestDB()

	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	user.HashedPassword = NullString(GetUserRepository().GetHashedPassword("12345678"))
	GetUserRepository().Update(user)

	// Log in
	payload := "{ \"email\": \"" + user.Email + "\", \"password\": \"12345678\", \"organizationId\": \"" + org.ID + "\" }"
	req := NewHTTPRequest("POST", "/auth/login", "", bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *JWTResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestBool(t, true, len(resBody.AccessToken) > 32)
	CheckTestBool(t, true, len(resBody.RefreshToken) == 36)

	// Sleep to ensure new access token
	time.Sleep(time.Second * 2)

	// Refresh access token
	payload = "{ \"refreshToken\": \"" + resBody.RefreshToken + "\" }"
	req = NewHTTPRequest("POST", "/auth/refresh", "", bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody3 *JWTResponse
	json.Unmarshal(res.Body.Bytes(), &resBody3)
	CheckTestBool(t, true, len(resBody3.AccessToken) > 32)
	CheckTestBool(t, true, len(resBody3.RefreshToken) == 36)
	CheckTestBool(t, false, resBody3.AccessToken == resBody.AccessToken)
	CheckTestBool(t, false, resBody3.RefreshToken == resBody.RefreshToken)

	// Test refreshed access token
	req = NewHTTPRequestWithAccessToken("GET", "/user/me", resBody3.AccessToken, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody2 *GetUserResponse
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	CheckTestString(t, user.Email, resBody2.Email)
}

func TestAuthRefreshNonExistent(t *testing.T) {
	ClearTestDB()

	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	user.HashedPassword = NullString(GetUserRepository().GetHashedPassword("12345678"))
	GetUserRepository().Update(user)

	// Refresh access token
	payload := "{ \"refreshToken\": \"" + uuid.New().String() + "\" }"
	req := NewHTTPRequest("POST", "/auth/refresh", "", bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestAuthPasswordReset(t *testing.T) {
	ClearTestDB()

	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	user.HashedPassword = NullString(GetUserRepository().GetHashedPassword("12345678"))
	GetUserRepository().Update(user)

	// Init password reset
	payload := "{ \"email\": \"" + user.Email + "\", \"organizationId\": \"" + org.ID + "\" }"
	req := NewHTTPRequest("POST", "/auth/initpwreset", "", bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)
	CheckTestBool(t, true, strings.Contains(SendMailMockContent, "Hallo "+user.Email+","))

	// Extract Confirm ID from email
	rx := regexp.MustCompile(`/resetpw/([0-9a-fA-F]{8}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{12})?/"`)
	confirmTokens := rx.FindStringSubmatch(SendMailMockContent)
	CheckTestInt(t, 2, len(confirmTokens))
	confirmID := confirmTokens[1]

	// Complete password reset
	payload = `{
			"password": "abcd1234"
		}`
	req = NewHTTPRequest("POST", "/auth/pwreset/"+confirmID, "", bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Test login with old password
	payload = "{ \"email\": \"" + user.Email + "\", \"password\": \"12345678\", \"organizationId\": \"" + org.ID + "\" }"
	req = NewHTTPRequest("POST", "/auth/login", "", bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)

	// Test login with new password
	payload = "{ \"email\": \"" + user.Email + "\", \"password\": \"abcd1234\", \"organizationId\": \"" + org.ID + "\" }"
	req = NewHTTPRequest("POST", "/auth/login", "", bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
}

func TestAuthSingleOrg(t *testing.T) {
	ClearTestDB()
	CreateTestOrg("test.com")

	req := NewHTTPRequestWithAccessToken("GET", "/auth/singleorg", "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	var resBody *AuthPreflightResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestBool(t, false, resBody.RequirePassword)
	CheckTestBool(t, false, resBody.Organization == nil)
	CheckTestString(t, "Test Org", resBody.Organization.Name)
}

func TestAuthSingleOrgWithMultipleOrgs(t *testing.T) {
	ClearTestDB()
	CreateTestOrg("test1.com")
	CreateTestOrg("test2.com")

	req := NewHTTPRequestWithAccessToken("GET", "/auth/singleorg", "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestAuthOrgDetails(t *testing.T) {
	ClearTestDB()
	org1 := CreateTestOrg("test1.com")
	CreateTestUserInOrg(org1)

	org2 := CreateTestOrg("test2.com")
	user2 := CreateTestUserInOrg(org2)
	user2.HashedPassword = NullString(GetUserRepository().GetHashedPassword("12345678"))
	GetUserRepository().Update(user2)

	req := NewHTTPRequestWithAccessToken("GET", "/auth/org/test1.com", "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *AuthPreflightResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestString(t, "Test Org", resBody.Organization.Name)
	CheckTestBool(t, false, resBody.RequirePassword)

	req = NewHTTPRequestWithAccessToken("GET", "/auth/org/test2.com", "", nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody2 *AuthPreflightResponse
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	CheckTestString(t, "Test Org", resBody2.Organization.Name)
	CheckTestBool(t, true, resBody2.RequirePassword)

	req = NewHTTPRequestWithAccessToken("GET", "/auth/org/test3.com", "", nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestAuthVerify(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test1.com")
	authProvider := &AuthProvider{
		OrganizationID: org.ID,
		Name:           "test",
		ProviderType:   int(OAuth2),
	}
	GetAuthProviderRepository().Create(authProvider)

	payloadAuthState := &AuthStateLoginPayload{
		UserID:    "test@foo.bar",
		LoginType: "ui",
	}
	payloadAuthStateJson, _ := json.Marshal(payloadAuthState)
	authState := &AuthState{
		AuthProviderID: authProvider.ID,
		Expiry:         time.Now().Add(time.Minute * 5),
		AuthStateType:  AuthResponseCache,
		Payload:        string(payloadAuthStateJson),
	}
	GetAuthStateRepository().Create(authState)

	req := NewHTTPRequest("GET", "/auth/verify/"+authState.ID, "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var loginRes JWTResponse
	json.Unmarshal(res.Body.Bytes(), &loginRes)
	CheckTestBool(t, true, len(loginRes.AccessToken) > 0)
	CheckTestBool(t, true, len(loginRes.RefreshToken) > 0)

	user, _ := GetUserRepository().GetByEmail(org.ID, "test@foo.bar")
	CheckTestBool(t, true, user != nil)
	CheckTestString(t, "test@foo.bar", user.Email)
}

func TestAuthServiceAccountNoLogin(t *testing.T) {
	ClearTestDB()

	org := CreateTestOrg("test.com")
	user := &User{
		Email:          uuid.New().String() + "@test.com",
		OrganizationID: org.ID,
		Role:           UserRoleServiceAccountRO,
		HashedPassword: NullString(GetUserRepository().GetHashedPassword("12345678")),
	}
	if err := GetUserRepository().Create(user); err != nil {
		t.Fatal(err)
	}

	// Log in
	payload := "{ \"email\": \"" + user.Email + "\", \"password\": \"12345678\", \"organizationId\": \"" + org.ID + "\" }"
	req := NewHTTPRequest("POST", "/auth/login", "", bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}
