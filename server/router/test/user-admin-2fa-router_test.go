package test

import (
	"encoding/json"
	"net/http"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func setTotpSecretForTest(t *testing.T, user *User, plainSecret string) {
	t.Helper()
	encrypted, err := EncryptString(plainSecret)
	if err != nil {
		t.Fatalf("EncryptString: %v", err)
	}
	user.TotpSecret = NullString(encrypted)
	if err := GetUserRepository().Update(user); err != nil {
		t.Fatalf("Update user TOTP secret: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Admin reset passkeys
// ---------------------------------------------------------------------------

func TestAdminResetPasskeysSuccess(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	target := CreateTestUserInOrg(org)

	createRouterTestPasskey(target, "Key A")
	createRouterTestPasskey(target, "Key B")

	req := NewHTTPRequest("DELETE", "/user/"+target.ID+"/passkeys", admin.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	passkeys, err := GetPasskeyRepository().GetAllByUserID(target.ID)
	if err != nil {
		t.Fatalf("GetAllByUserID: %v", err)
	}
	CheckTestInt(t, 0, len(passkeys))
}

func TestAdminResetPasskeysNoPasskeysFine(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	target := CreateTestUserInOrg(org)

	req := NewHTTPRequest("DELETE", "/user/"+target.ID+"/passkeys", admin.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)
}

func TestAdminResetPasskeysNotFound(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)

	req := NewHTTPRequest("DELETE", "/user/00000000-0000-0000-0000-000000000000/passkeys", admin.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestAdminResetPasskeysForbiddenForRegularUser(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	regularUser := CreateTestUserInOrg(org)
	target := CreateTestUserInOrg(org)

	createRouterTestPasskey(target, "Key A")

	req := NewHTTPRequest("DELETE", "/user/"+target.ID+"/passkeys", regularUser.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	passkeys, err := GetPasskeyRepository().GetAllByUserID(target.ID)
	if err != nil {
		t.Fatalf("GetAllByUserID: %v", err)
	}
	CheckTestInt(t, 1, len(passkeys))
}

func TestAdminResetPasskeysForbiddenForForeignOrg(t *testing.T) {
	ClearTestDB()
	org1 := CreateTestOrg("org1.com")
	org2 := CreateTestOrg("org2.com")
	admin1 := CreateTestUserOrgAdmin(org1)
	target := CreateTestUserInOrg(org2)

	createRouterTestPasskey(target, "Key A")

	req := NewHTTPRequest("DELETE", "/user/"+target.ID+"/passkeys", admin1.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	passkeys, err := GetPasskeyRepository().GetAllByUserID(target.ID)
	if err != nil {
		t.Fatalf("GetAllByUserID: %v", err)
	}
	CheckTestInt(t, 1, len(passkeys))
}

func TestAdminResetPasskeysUnauthorized(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	target := CreateTestUserInOrg(org)

	req := NewHTTPRequest("DELETE", "/user/"+target.ID+"/passkeys", "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusUnauthorized, res.Code)
}

func TestAdminResetPasskeysResponseFieldUpdated(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	target := CreateTestUserInOrg(org)

	createRouterTestPasskey(target, "Key A")

	req := NewHTTPRequest("DELETE", "/user/"+target.ID+"/passkeys", admin.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	req = NewHTTPRequest("GET", "/user/"+target.ID, admin.ID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody GetUserResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestBool(t, false, resBody.HasPasskeys)
}

// ---------------------------------------------------------------------------
// Admin reset TOTP
// ---------------------------------------------------------------------------

func TestAdminResetTotpSuccess(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	target := CreateTestUserInOrg(org)

	setTotpSecretForTest(t, target, "JBSWY3DPEHPK3PXP")

	req := NewHTTPRequest("DELETE", "/user/"+target.ID+"/totp", admin.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	updated, err := GetUserRepository().GetOne(target.ID)
	if err != nil {
		t.Fatalf("GetOne: %v", err)
	}
	CheckTestBool(t, true, updated.TotpSecret == "")
}

func TestAdminResetTotpNoTotpFine(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	target := CreateTestUserInOrg(org)

	req := NewHTTPRequest("DELETE", "/user/"+target.ID+"/totp", admin.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)
}

func TestAdminResetTotpNotFound(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)

	req := NewHTTPRequest("DELETE", "/user/00000000-0000-0000-0000-000000000000/totp", admin.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestAdminResetTotpForbiddenForRegularUser(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	regularUser := CreateTestUserInOrg(org)
	target := CreateTestUserInOrg(org)

	setTotpSecretForTest(t, target, "JBSWY3DPEHPK3PXP")

	req := NewHTTPRequest("DELETE", "/user/"+target.ID+"/totp", regularUser.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	updated, err := GetUserRepository().GetOne(target.ID)
	if err != nil {
		t.Fatalf("GetOne: %v", err)
	}
	if updated.TotpSecret == "" {
		t.Error("expected TotpSecret to still be set after forbidden request")
	}
}

func TestAdminResetTotpForbiddenForForeignOrg(t *testing.T) {
	ClearTestDB()
	org1 := CreateTestOrg("org1.com")
	org2 := CreateTestOrg("org2.com")
	admin1 := CreateTestUserOrgAdmin(org1)
	target := CreateTestUserInOrg(org2)

	setTotpSecretForTest(t, target, "JBSWY3DPEHPK3PXP")

	req := NewHTTPRequest("DELETE", "/user/"+target.ID+"/totp", admin1.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	updated, err := GetUserRepository().GetOne(target.ID)
	if err != nil {
		t.Fatalf("GetOne: %v", err)
	}
	if updated.TotpSecret == "" {
		t.Error("expected TotpSecret to still be set after forbidden request")
	}
}

func TestAdminResetTotpUnauthorized(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	target := CreateTestUserInOrg(org)

	req := NewHTTPRequest("DELETE", "/user/"+target.ID+"/totp", "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusUnauthorized, res.Code)
}

func TestAdminResetTotpResponseFieldUpdated(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	target := CreateTestUserInOrg(org)

	setTotpSecretForTest(t, target, "JBSWY3DPEHPK3PXP")

	req := NewHTTPRequest("DELETE", "/user/"+target.ID+"/totp", admin.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	req = NewHTTPRequest("GET", "/user/"+target.ID, admin.ID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody GetUserResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestBool(t, false, resBody.TotpEnabled)
}
