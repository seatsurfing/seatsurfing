package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func createRouterTestPasskey(user *User, name string) *Passkey {
	rawID := []byte(uuid.New().String())
	cred := &webauthn.Credential{
		ID:              rawID,
		PublicKey:       []byte{0x01, 0x02, 0x03},
		AttestationType: "none",
		Authenticator: webauthn.Authenticator{
			AAGUID:    make([]byte, 16),
			SignCount: 1,
		},
	}
	pk, err := NewPasskeyFromCredential(user.ID, cred, name)
	if err != nil {
		panic("NewPasskeyFromCredential: " + err.Error())
	}
	if err := GetPasskeyRepository().Create(pk); err != nil {
		panic("Create passkey: " + err.Error())
	}
	return pk
}

func newWebAuthnRequest(method, url, userID string, body *bytes.Buffer) *http.Request {
	var req *http.Request
	if body != nil {
		req = NewHTTPRequest(method, url, userID, body)
	} else {
		req = NewHTTPRequest(method, url, userID, nil)
	}
	req.Host = "localhost"
	return req
}

func TestPasskeyListEmpty(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	req := NewHTTPRequest("GET", "/user/passkey/", user.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	var list []*PasskeyListItemResponse
	json.Unmarshal(res.Body.Bytes(), &list)
	CheckTestInt(t, 0, len(list))
}

func TestPasskeyListNonEmpty(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	user.HashedPassword = NullString(GetUserRepository().GetHashedPassword("12345678"))
	GetUserRepository().Update(user)

	createRouterTestPasskey(user, "Key A")
	createRouterTestPasskey(user, "Key B")

	req := NewHTTPRequest("GET", "/user/passkey/", user.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	var list []*PasskeyListItemResponse
	json.Unmarshal(res.Body.Bytes(), &list)
	CheckTestInt(t, 2, len(list))
}

func TestPasskeyListUnauthorized(t *testing.T) {
	ClearTestDB()
	req := NewHTTPRequest("GET", "/user/passkey/", "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusUnauthorized, res.Code)
}

func TestPasskeyRenameSuccess(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	user.HashedPassword = NullString(GetUserRepository().GetHashedPassword("12345678"))
	GetUserRepository().Update(user)

	pk := createRouterTestPasskey(user, "Old Name")

	payload := `{"name": "New Name"}`
	req := NewHTTPRequest("PUT", "/user/passkey/"+pk.ID, user.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	updated, err := GetPasskeyRepository().GetOne(pk.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestString(t, "New Name", updated.Name)
}

func TestPasskeyRenameNotFound(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	payload := `{"name": "Whatever"}`
	req := NewHTTPRequest("PUT", "/user/passkey/"+uuid.New().String(), user.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestPasskeyRenameForbidden(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user1 := CreateTestUserInOrg(org)
	user2 := CreateTestUserInOrg(org)

	pk := createRouterTestPasskey(user1, "User1 Key")

	payload := `{"name": "Hacked"}`
	req := NewHTTPRequest("PUT", "/user/passkey/"+pk.ID, user2.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestPasskeyRenameBadRequest(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	pk := createRouterTestPasskey(user, "Some Key")

	payload := `{"name": ""}`
	req := NewHTTPRequest("PUT", "/user/passkey/"+pk.ID, user.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
}

func TestPasskeyRenameTooLong(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	pk := createRouterTestPasskey(user, "Some Key")

	longName := ""
	for i := 0; i < 256; i++ {
		longName += "a"
	}
	payload := `{"name": "` + longName + `"}`
	req := NewHTTPRequest("PUT", "/user/passkey/"+pk.ID, user.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
}

func TestPasskeyDeleteSuccess(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	pk := createRouterTestPasskey(user, "To Delete")

	req := NewHTTPRequest("DELETE", "/user/passkey/"+pk.ID, user.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)
	CheckTestInt(t, 0, GetPasskeyRepository().GetCountByUserID(user.ID))
}

func TestPasskeyDeleteNotFound(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	req := NewHTTPRequest("DELETE", "/user/passkey/"+uuid.New().String(), user.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestPasskeyDeleteForbidden(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user1 := CreateTestUserInOrg(org)
	user2 := CreateTestUserInOrg(org)

	pk := createRouterTestPasskey(user1, "User1 Key")

	req := NewHTTPRequest("DELETE", "/user/passkey/"+pk.ID, user2.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
	CheckTestInt(t, 1, GetPasskeyRepository().GetCountByUserID(user1.ID))
}

func TestPasskeyDeleteLastWithEnforceTOTP(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org.ID, SettingEnforceTOTP.Name, "1")

	user := CreateTestUserInOrg(org)
	pk := createRouterTestPasskey(user, "Only Key")

	req := NewHTTPRequest("DELETE", "/user/passkey/"+pk.ID, user.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
	CheckTestInt(t, 1, GetPasskeyRepository().GetCountByUserID(user.ID))
}

func TestPasskeyDeleteNotLastWithEnforceTOTP(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org.ID, SettingEnforceTOTP.Name, "1")

	user := CreateTestUserInOrg(org)
	pk1 := createRouterTestPasskey(user, "Key 1")
	createRouterTestPasskey(user, "Key 2")

	req := NewHTTPRequest("DELETE", "/user/passkey/"+pk1.ID, user.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)
	CheckTestInt(t, 1, GetPasskeyRepository().GetCountByUserID(user.ID))
}

func TestPasskeyDeleteLastWithEnforceTOTPButTOTPConfigured(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org.ID, SettingEnforceTOTP.Name, "1")

	user := CreateTestUserInOrg(org)
	user.TotpSecret = NullString("JBSWY3DPEHPK3PXP")
	GetUserRepository().Update(user)

	pk := createRouterTestPasskey(user, "Only Key")

	req := NewHTTPRequest("DELETE", "/user/passkey/"+pk.ID, user.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)
}

func TestPasskeyRegistrationBeginNonPrimaryDomain(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	user.HashedPassword = NullString(GetUserRepository().GetHashedPassword("12345678"))
	GetUserRepository().Update(user)

	// Host is "other.com", primary domain is "test.com" → must be rejected
	req := NewHTTPRequest("POST", "/user/passkey/registration/begin", user.ID, nil)
	req.Host = "other.com"
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestPasskeyRegistrationBeginNoPassword(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	// No password → 403 regardless of domain (check fires before domain check)
	req := NewHTTPRequest("POST", "/user/passkey/registration/begin", user.ID, nil)
	req.Host = "test.com"
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestPasskeyRegistrationBeginIdpUser(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	user.HashedPassword = NullString(GetUserRepository().GetHashedPassword("12345678"))
	user.AuthProviderID = NullUUID(uuid.New().String())
	GetUserRepository().Update(user)

	// IdP user → 403 regardless of domain (check fires before domain check)
	req := NewHTTPRequest("POST", "/user/passkey/registration/begin", user.ID, nil)
	req.Host = "test.com"
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestPasskeyRegistrationBeginSuccess(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	user.HashedPassword = NullString(GetUserRepository().GetHashedPassword("12345678"))
	GetUserRepository().Update(user)

	// Host must match the primary domain ("test.com") for registration to succeed
	req := newWebAuthnRequest("POST", "/user/passkey/registration/begin", user.ID, nil)
	req.Host = "test.com"
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	var regRes BeginPasskeyRegistrationResponse
	json.Unmarshal(res.Body.Bytes(), &regRes)
	CheckStringNotEmpty(t, regRes.StateID)
	CheckTestBool(t, true, regRes.Challenge != nil)
}

func TestPasskeyRegistrationBeginUnauthorized(t *testing.T) {
	ClearTestDB()
	req := NewHTTPRequest("POST", "/user/passkey/registration/begin", "", nil)
	req.Host = "localhost"
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusUnauthorized, res.Code)
}

func TestPasskeyRegistrationFinishExpiredState(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	user.HashedPassword = NullString(GetUserRepository().GetHashedPassword("12345678"))
	GetUserRepository().Update(user)

	payload := `{"stateId": "` + uuid.New().String() + `", "name": "My Key", "credential": {}}`
	req := newWebAuthnRequest("POST", "/user/passkey/registration/finish", user.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestPasskeyRegistrationFinishWrongUser(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user1 := CreateTestUserInOrg(org)
	user1.HashedPassword = NullString(GetUserRepository().GetHashedPassword("12345678"))
	GetUserRepository().Update(user1)
	user2 := CreateTestUserInOrg(org)
	user2.HashedPassword = NullString(GetUserRepository().GetHashedPassword("12345678"))
	GetUserRepository().Update(user2)

	// Begin from primary domain
	req := newWebAuthnRequest("POST", "/user/passkey/registration/begin", user1.ID, nil)
	req.Host = "test.com"
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var regRes BeginPasskeyRegistrationResponse
	json.Unmarshal(res.Body.Bytes(), &regRes)

	payload := `{"stateId": "` + regRes.StateID + `", "name": "Hijack", "credential": {}}`
	req = newWebAuthnRequest("POST", "/user/passkey/registration/finish", user2.ID, bytes.NewBufferString(payload))
	req.Host = "test.com"
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestPasskeyLoginBegin(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	payload := `{"organizationId": "` + org.ID + `"}`
	req := newWebAuthnRequest("POST", "/auth/passkey/login/begin", "", bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	var loginBeginRes BeginPasskeyLoginResponse
	json.Unmarshal(res.Body.Bytes(), &loginBeginRes)
	CheckStringNotEmpty(t, loginBeginRes.StateID)
	CheckTestBool(t, true, loginBeginRes.Challenge != nil)
}

func TestPasswordLoginPasskeyChallenge(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	user.HashedPassword = NullString(GetUserRepository().GetHashedPassword("12345678"))
	GetUserRepository().Update(user)
	createRouterTestPasskey(user, "Test Key")

	payload := `{"email": "` + user.Email + `", "password": "12345678", "organizationId": "` + org.ID + `"}`
	req := newWebAuthnRequest("POST", "/auth/login", "", bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusUnauthorized, res.Code)

	var challenge PasskeyChallengeResponse
	json.Unmarshal(res.Body.Bytes(), &challenge)
	CheckTestBool(t, true, challenge.RequirePasskey)
	CheckStringNotEmpty(t, challenge.StateID)
	CheckTestBool(t, true, challenge.PasskeyChallenge != nil)
}

func TestPasswordLoginPasskeyAllowTotpFallback(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	user.HashedPassword = NullString(GetUserRepository().GetHashedPassword("12345678"))
	user.TotpSecret = NullString("JBSWY3DPEHPK3PXP")
	GetUserRepository().Update(user)
	createRouterTestPasskey(user, "Test Key")

	payload := `{"email": "` + user.Email + `", "password": "12345678", "organizationId": "` + org.ID + `"}`
	req := newWebAuthnRequest("POST", "/auth/login", "", bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusUnauthorized, res.Code)

	var challenge PasskeyChallengeResponse
	json.Unmarshal(res.Body.Bytes(), &challenge)
	CheckTestBool(t, true, challenge.RequirePasskey)
	CheckTestBool(t, true, challenge.PasskeyChallenge != nil)
	CheckTestBool(t, true, challenge.AllowTotpFallback)
}

func TestPasswordLoginPasskeyExpiredState(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	user.HashedPassword = NullString(GetUserRepository().GetHashedPassword("12345678"))
	GetUserRepository().Update(user)
	createRouterTestPasskey(user, "Test Key")

	payload := `{"email": "` + user.Email + `", "password": "12345678", "organizationId": "` + org.ID + `", "passkeyStateId": "` + uuid.New().String() + `", "passkeyCredential": {}}`
	req := newWebAuthnRequest("POST", "/auth/login", "", bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestPasswordLoginPasskeyWrongUser(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user1 := CreateTestUserInOrg(org)
	user1.HashedPassword = NullString(GetUserRepository().GetHashedPassword("12345678"))
	GetUserRepository().Update(user1)
	user2 := CreateTestUserInOrg(org)
	user2.HashedPassword = NullString(GetUserRepository().GetHashedPassword("87654321"))
	GetUserRepository().Update(user2)
	createRouterTestPasskey(user1, "User1 Key")
	createRouterTestPasskey(user2, "User2 Key")

	state := &AuthState{
		AuthProviderID: user1.ID,
		AuthStateType:  AuthPasskey2FA,
		Expiry:         time.Now().Add(5 * time.Minute),
		Payload:        `{}`,
	}
	GetAuthStateRepository().Create(state)

	payload := `{"email": "` + user2.Email + `", "password": "87654321", "organizationId": "` + org.ID + `", "passkeyStateId": "` + state.ID + `", "passkeyCredential": {}}`
	req := newWebAuthnRequest("POST", "/auth/login", "", bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestGetUserMeHasPasskeysFalse(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	req := NewHTTPRequest("GET", "/user/me", user.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	var userRes GetUserResponse
	json.Unmarshal(res.Body.Bytes(), &userRes)
	CheckTestBool(t, false, userRes.HasPasskeys)
}

func TestGetUserMeHasPasskeysTrue(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	user.HashedPassword = NullString(GetUserRepository().GetHashedPassword("12345678"))
	GetUserRepository().Update(user)
	createRouterTestPasskey(user, "My Passkey")

	req := NewHTTPRequest("GET", "/user/me", user.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	var userRes GetUserResponse
	json.Unmarshal(res.Body.Bytes(), &userRes)
	CheckTestBool(t, true, userRes.HasPasskeys)
}

func TestGetUserMeIsPrimaryDomainTrue(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	// Host matches the primary domain → isPrimaryDomain must be true
	req := NewHTTPRequest("GET", "/user/me", user.ID, nil)
	req.Host = "test.com"
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	var userRes GetUserResponse
	json.Unmarshal(res.Body.Bytes(), &userRes)
	CheckTestBool(t, true, userRes.IsPrimaryDomain)
}

func TestGetUserMeIsPrimaryDomainFalse(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	// Host does not match the primary domain → isPrimaryDomain must be false
	req := NewHTTPRequest("GET", "/user/me", user.ID, nil)
	req.Host = "other.com"
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	var userRes GetUserResponse
	json.Unmarshal(res.Body.Bytes(), &userRes)
	CheckTestBool(t, false, userRes.IsPrimaryDomain)
}

func TestGetUserMeIsPrimaryDomainCaseInsensitive(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	// Host in different case → still matches (EqualFold)
	req := NewHTTPRequest("GET", "/user/me", user.ID, nil)
	req.Host = "TEST.COM"
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	var userRes GetUserResponse
	json.Unmarshal(res.Body.Bytes(), &userRes)
	CheckTestBool(t, true, userRes.IsPrimaryDomain)
}

func TestPasskeyLoginFinishExpiredState(t *testing.T) {
	ClearTestDB()
	payload := `{"stateId": "` + uuid.New().String() + `", "credential": {}}`
	req := newWebAuthnRequest("POST", "/auth/passkey/login/finish", "", bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}
