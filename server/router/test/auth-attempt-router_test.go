package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestAuthAttemptRouterForbiddenForNonAdmin(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	req := NewHTTPRequest("GET", "/auth-attempt/", user.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestAuthAttemptRouterOrgIsolation(t *testing.T) {
	ClearTestDB()
	org1 := CreateTestOrg("test1.com")
	org2 := CreateTestOrg("test2.com")
	admin1 := CreateTestUserOrgAdmin(org1)
	user1 := CreateTestUserInOrgWithName(org1, "u1@test1.com", UserRoleUser)
	user2 := CreateTestUserInOrgWithName(org2, "u2@test2.com", UserRoleUser)

	GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{User: user1, Method: AuthMethodPassword, ErrorCode: AuthErrorWrongPassword})
	GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{User: user2, Method: AuthMethodPassword, ErrorCode: AuthErrorWrongPassword})

	req := NewHTTPRequest("GET", "/auth-attempt/", admin1.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetAuthAttemptListResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, 1, resBody.Total)
	CheckTestInt(t, 1, len(resBody.Items))
	CheckTestString(t, user1.Email, resBody.Items[0].Email)
	CheckTestString(t, AuthErrorWrongPassword, resBody.Items[0].ErrorCode)
}

func TestAuthAttemptRouterFiltersAndPaging(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	user := CreateTestUserInOrgWithName(org, "u1@test.com", UserRoleUser)

	GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{User: user, Method: AuthMethodPassword, ErrorCode: AuthErrorWrongPassword})
	GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{User: user, Successful: true, Method: AuthMethodPassword})
	GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{OrganizationID: org.ID, Email: "ghost@test.com", Method: AuthMethodOAuth, ErrorCode: AuthErrorUserNotFound})

	// success filter
	req := NewHTTPRequest("GET", "/auth-attempt/?success=false", admin.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetAuthAttemptListResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, 2, resBody.Total)

	// method filter
	req = NewHTTPRequest("GET", "/auth-attempt/?method=oauth", admin.ID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, 1, resBody.Total)
	CheckTestString(t, "ghost@test.com", resBody.Items[0].Email)

	// error code filter
	req = NewHTTPRequest("GET", "/auth-attempt/?errorCode=wrong_password", admin.ID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, 1, resBody.Total)

	// email filter
	req = NewHTTPRequest("GET", "/auth-attempt/?user=ghost", admin.ID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, 1, resBody.Total)

	// paging: total reflects all matches, items are limited
	req = NewHTTPRequest("GET", "/auth-attempt/?limit=2", admin.ID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, 3, resBody.Total)
	CheckTestInt(t, 2, len(resBody.Items))
	req = NewHTTPRequest("GET", "/auth-attempt/?limit=2&offset=2", admin.ID, nil)
	res = ExecuteTestRequest(req)
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, 1, len(resBody.Items))

	// limit is capped at 100 (still a valid request)
	req = NewHTTPRequest("GET", "/auth-attempt/?limit=1000", admin.ID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	// invalid params
	req = NewHTTPRequest("GET", "/auth-attempt/?start=notadate", admin.ID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
	req = NewHTTPRequest("GET", "/auth-attempt/?limit=0", admin.ID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
}

func TestAuthAttemptRouterLoginInstrumentation(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	user := CreateTestUserInOrg(org)
	user.HashedPassword = NullString(GetUserRepository().GetHashedPassword(TestPassword))
	GetUserRepository().Update(user)

	// failed login with wrong password
	payload := "{ \"email\": \"" + user.Email + "\", \"password\": \"wrongpassword\", \"organizationId\": \"" + org.ID + "\" }"
	req := NewHTTPRequest("POST", "/auth/login", "", bytes.NewBufferString(payload))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/126.0.0.0")
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)

	// failed login with unknown email
	payload = "{ \"email\": \"nobody@test.com\", \"password\": \"whatever123\", \"organizationId\": \"" + org.ID + "\" }"
	req = NewHTTPRequest("POST", "/auth/login", "", bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)

	// successful login
	payload = "{ \"email\": \"" + user.Email + "\", \"password\": \"" + TestPassword + "\", \"organizationId\": \"" + org.ID + "\" }"
	req = NewHTTPRequest("POST", "/auth/login", "", bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	req = NewHTTPRequest("GET", "/auth-attempt/", admin.ID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetAuthAttemptListResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, 3, resBody.Total)

	numSuccess := 0
	for _, item := range resBody.Items {
		if item.Successful {
			numSuccess++
			CheckTestString(t, "password", item.Method)
			CheckTestString(t, "", item.ErrorCode)
			CheckTestString(t, user.ID, item.UserID)
		}
		if item.ErrorCode == AuthErrorWrongPassword {
			CheckTestString(t, user.Email, item.Email)
			CheckTestString(t, "Chrome on Windows 10/11", item.Device)
		}
		if item.ErrorCode == AuthErrorUserNotFound {
			CheckTestString(t, "nobody@test.com", item.Email)
			CheckTestString(t, "", item.UserID)
		}
	}
	CheckTestInt(t, 1, numSuccess)
}

func TestAuthAttemptRouterOAuthStateInvalid(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	provider := &AuthProvider{
		OrganizationID:     org.ID,
		Name:               "Test IdP",
		ProviderType:       int(OAuth2),
		AuthURL:            "https://idp.example.com/auth",
		TokenURL:           "https://idp.example.com/token",
		AuthStyle:          1,
		Scopes:             "openid,email",
		UserInfoURL:        "https://idp.example.com/userinfo",
		UserInfoEmailField: "email",
		ClientID:           "client",
		ClientSecret:       "secret",
	}
	if err := GetAuthProviderRepository().Create(provider); err != nil {
		t.Fatal(err)
	}

	req := NewHTTPRequest("GET", "/auth/"+provider.ID+"/callback?state=bogus&code=bogus", "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusTemporaryRedirect, res.Code)

	req = NewHTTPRequest("GET", "/auth-attempt/", admin.ID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetAuthAttemptListResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, 1, resBody.Total)
	CheckTestString(t, AuthErrorIdpStateInvalid, resBody.Items[0].ErrorCode)
	CheckTestString(t, "oauth", resBody.Items[0].Method)
	CheckTestString(t, provider.ID, resBody.Items[0].AuthProviderID)
	CheckTestString(t, "Test IdP", resBody.Items[0].AuthProviderName)
}
