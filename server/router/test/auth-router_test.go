package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/config"
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
	CheckTestBool(t, true, strings.Contains(SendMailMockContent, "Hallo "+user.GetSafeRecipientName()+","))

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

func TestTokenValid(t *testing.T) {
	ClearTestDB()

	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	router := &AuthRouter{}
	session := router.CreateSession(nil, user)
	claims := router.CreateClaims(user, session)
	token := router.CreateAccessToken(claims)

	req, _ := http.NewRequest("GET", "/user/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
}

func TestTokenNoAudience(t *testing.T) {
	ClearTestDB()

	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	router := &AuthRouter{}
	session := router.CreateSession(nil, user)
	claims := router.CreateClaims(user, session)
	token := router.CreateAccessToken(claims, WithoutAudience)

	req, _ := http.NewRequest("GET", "/user/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusUnauthorized, res.Code)
}

func TestTokenNoExpiry(t *testing.T) {
	ClearTestDB()

	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	router := &AuthRouter{}
	session := router.CreateSession(nil, user)
	claims := router.CreateClaims(user, session)
	token := router.CreateAccessToken(claims, WithoutExpiry)

	req, _ := http.NewRequest("GET", "/user/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusUnauthorized, res.Code)
}

func TestTokenNoIssuer(t *testing.T) {
	ClearTestDB()

	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	router := &AuthRouter{}
	session := router.CreateSession(nil, user)
	claims := router.CreateClaims(user, session)
	token := router.CreateAccessToken(claims, WithoutIssuer)

	req, _ := http.NewRequest("GET", "/user/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusUnauthorized, res.Code)
}

func TestTokenExpired(t *testing.T) {
	ClearTestDB()

	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	router := &AuthRouter{}
	session := router.CreateSession(nil, user)
	claims := router.CreateClaims(user, session)
	installID, _ := GetSettingsRepository().GetGlobalString(SettingInstallID.Name)
	jti := uuid.New().String()
	claims.RegisteredClaims = jwt.RegisteredClaims{
		ID: jti,
	}
	claims.RegisteredClaims.Issuer = installID
	claims.RegisteredClaims.Audience = jwt.ClaimStrings{installID}
	claims.RegisteredClaims.IssuedAt = jwt.NewNumericDate(time.Now())
	claims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-1 * time.Minute))
	claims.RegisteredClaims.NotBefore = jwt.NewNumericDate(time.Now())
	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	token, err := accessToken.SignedString(GetConfig().JwtPrivateKey)
	CheckTestBool(t, true, err == nil)

	req, _ := http.NewRequest("GET", "/user/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusUnauthorized, res.Code)
}

func TestTokenNotValidYet(t *testing.T) {
	ClearTestDB()

	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	router := &AuthRouter{}
	session := router.CreateSession(nil, user)
	claims := router.CreateClaims(user, session)
	installID, _ := GetSettingsRepository().GetGlobalString(SettingInstallID.Name)
	jti := uuid.New().String()
	claims.RegisteredClaims = jwt.RegisteredClaims{
		ID: jti,
	}
	claims.RegisteredClaims.Issuer = installID
	claims.RegisteredClaims.Audience = jwt.ClaimStrings{installID}
	claims.RegisteredClaims.IssuedAt = jwt.NewNumericDate(time.Now())
	claims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(15 * time.Minute))
	claims.RegisteredClaims.NotBefore = jwt.NewNumericDate(time.Now().Add(5 * time.Minute))
	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	token, err := accessToken.SignedString(GetConfig().JwtPrivateKey)
	CheckTestBool(t, true, err == nil)

	req, _ := http.NewRequest("GET", "/user/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusUnauthorized, res.Code)
}

func TestLogoutCurrent(t *testing.T) {
	ClearTestDB()

	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	token1 := GetTestJWT(user.ID)
	token2 := GetTestJWT(user.ID)

	// Test both tokens are valid
	req := NewHTTPRequestWithAccessToken("GET", "/user/me", token1, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	req = NewHTTPRequestWithAccessToken("GET", "/user/me", token2, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	// Logout of first token
	req = NewHTTPRequestWithAccessToken("GET", "/auth/logout/current", token1, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Test first token is invalid
	req = NewHTTPRequestWithAccessToken("GET", "/user/me", token1, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusUnauthorized, res.Code)

	// Test second token is still valid
	req = NewHTTPRequestWithAccessToken("GET", "/user/me", token2, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
}

func TestLogoutAll(t *testing.T) {
	ClearTestDB()

	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	token1 := GetTestJWT(user.ID)
	token2 := GetTestJWT(user.ID)

	// Test both tokens are valid
	req := NewHTTPRequestWithAccessToken("GET", "/user/me", token1, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	req = NewHTTPRequestWithAccessToken("GET", "/user/me", token2, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	// Logout of all tokens
	req = NewHTTPRequestWithAccessToken("GET", "/auth/logout/all", token1, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Test both tokens are invalid
	req = NewHTTPRequestWithAccessToken("GET", "/user/me", token1, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusUnauthorized, res.Code)
	req = NewHTTPRequestWithAccessToken("GET", "/user/me", token2, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusUnauthorized, res.Code)
}

func TestLogoutSpecific(t *testing.T) {
	ClearTestDB()

	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	token1 := GetTestJWT(user.ID)
	token2 := GetTestJWT(user.ID)

	claims := &Claims{}
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"RS512"}),
	)
	_, _, err := parser.ParseUnverified(token1, claims)
	CheckTestBool(t, true, err == nil)

	// Logout of token1
	req := NewHTTPRequestWithAccessToken("GET", "/auth/logout/"+claims.SessionID, token2, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)
}

func TestLogoutSpecificForeign(t *testing.T) {
	ClearTestDB()

	org := CreateTestOrg("test.com")
	user1 := CreateTestUserInOrg(org)
	token1 := GetTestJWT(user1.ID)
	user2 := CreateTestUserInOrg(org)
	token2 := GetTestJWT(user2.ID)

	claims := &Claims{}
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"RS512"}),
	)
	_, _, err := parser.ParseUnverified(token1, claims)
	CheckTestBool(t, true, err == nil)

	// Logout of token1
	req := NewHTTPRequestWithAccessToken("GET", "/auth/logout/"+claims.SessionID, token2, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestAuthPasswordLoginWithPasswordPending(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	user.HashedPassword = NullString(GetUserRepository().GetHashedPassword("12345678"))
	user.PasswordPending = true
	GetUserRepository().Update(user)

	// Attempt to log in should fail
	payload := "{ \"email\": \"" + user.Email + "\", \"password\": \"12345678\", \"organizationId\": \"" + org.ID + "\" }"
	req := NewHTTPRequest("POST", "/auth/login", "", bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestAuthProviderBinding(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")

	// Create two auth providers
	authProvider1 := &AuthProvider{
		OrganizationID: org.ID,
		Name:           "Provider1",
		ProviderType:   int(OAuth2),
	}
	GetAuthProviderRepository().Create(authProvider1)

	authProvider2 := &AuthProvider{
		OrganizationID: org.ID,
		Name:           "Provider2",
		ProviderType:   int(OAuth2),
	}
	GetAuthProviderRepository().Create(authProvider2)

	// Create user via OAuth (simulated with AuthResponseCache)
	email := "test@test.com"
	payloadAuthState := &AuthStateLoginPayload{
		UserID:    email,
		LoginType: "ui",
	}
	payloadAuthStateJson, _ := json.Marshal(payloadAuthState)
	authState := &AuthState{
		AuthProviderID: authProvider1.ID,
		Expiry:         time.Now().Add(time.Minute * 5),
		AuthStateType:  AuthResponseCache,
		Payload:        string(payloadAuthStateJson),
	}
	GetAuthStateRepository().Create(authState)

	// User logs in via provider1 - should create user bound to provider1
	req := NewHTTPRequest("GET", "/auth/verify/"+authState.ID, "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	// Check user was created and bound to provider1
	user, _ := GetUserRepository().GetByEmail(org.ID, email)
	CheckTestBool(t, true, user != nil)
	CheckTestString(t, authProvider1.ID, string(user.AuthProviderID))

	// Try to login with provider2 - should fail
	payloadAuthState2 := &AuthStateLoginPayload{
		UserID:    email,
		LoginType: "ui",
	}
	payloadAuthStateJson2, _ := json.Marshal(payloadAuthState2)
	authState2 := &AuthState{
		AuthProviderID: authProvider2.ID,
		Expiry:         time.Now().Add(time.Minute * 5),
		AuthStateType:  AuthResponseCache,
		Payload:        string(payloadAuthStateJson2),
	}
	GetAuthStateRepository().Create(authState2)

	req = NewHTTPRequest("GET", "/auth/verify/"+authState2.ID, "", nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestAuthProviderBindingBackwardsCompatibility(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")

	// Create auth provider
	authProvider := &AuthProvider{
		OrganizationID: org.ID,
		Name:           "Provider1",
		ProviderType:   int(OAuth2),
	}
	GetAuthProviderRepository().Create(authProvider)

	// Create user without auth provider binding (existing user scenario)
	email := "olduser@test.com"
	user := &User{
		Email:          email,
		OrganizationID: org.ID,
		Role:           UserRoleUser,
		AuthProviderID: NullUUID(""),
	}
	GetUserRepository().Create(user)

	// User logs in via OAuth - should bind to provider
	payloadAuthState := &AuthStateLoginPayload{
		UserID:    email,
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

	// Check user is now bound to provider
	updatedUser, _ := GetUserRepository().GetByEmail(org.ID, email)
	CheckTestString(t, authProvider.ID, string(updatedUser.AuthProviderID))
}
