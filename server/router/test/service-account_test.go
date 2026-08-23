package test

import (
	"net/http"
	"strconv"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

// A service account authenticates with a header on every request and never
// sees an interactive login, so a second factor cannot apply to it. Enforcing
// one would lock it out the moment this were used to gate authentication.
func TestServiceAccountExemptFromTotpEnforcement(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	person := CreateTestUserOrgAdmin(org)
	serviceAccount := CreateTestServiceAccountRW(org)

	for _, mode := range []int{SettingEnforceTOTPAllUsers, SettingEnforceTOTPAdminsOnly} {
		GetSettingsRepository().Set(org.ID, SettingEnforceTOTP.Name, strconv.Itoa(mode))

		if !IsTotpEnforcedForUser(person) {
			t.Fatalf("expected enforcement to apply to a person in mode %d", mode)
		}
		if IsTotpEnforcedForUser(serviceAccount) {
			t.Fatalf("expected a service account to be exempt in mode %d", mode)
		}
	}
}

// The behaviour that matters: a service account keeps working while the
// organization enforces two-factor authentication.
func TestServiceAccountUsableWhileTotpEnforced(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	CreateTestUserOrgAdmin(org)
	GetSettingsRepository().Set(org.ID, SettingEnforceTOTP.Name, strconv.Itoa(SettingEnforceTOTPAllUsers))

	readWrite := CreateTestServiceAccountRW(org)
	rawToken := GenerateTestApiToken(readWrite.ID)

	req := NewHTTPRequestBearer("GET", "/user/", rawToken, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	req = NewHTTPRequestBearer("POST", "/group/", rawToken, nil)
	res = ExecuteTestRequest(req)
	if res.Code == http.StatusUnauthorized {
		t.Fatal("read/write service account was refused while TOTP enforcement was on")
	}
}

// Read-only service accounts stay restricted to GET, whatever their roles say.
func TestServiceAccountReadOnlyStillRestrictedToGet(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	CreateTestUserOrgAdmin(org)
	readOnly := CreateTestServiceAccountWithPassword(org, "ro@test.com", TestPassword, UserRoleServiceAccountRO)
	rawToken := GenerateTestApiToken(readOnly.ID)

	req := NewHTTPRequestBearer("GET", "/user/", rawToken, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	req = NewHTTPRequestBearer("POST", "/group/", rawToken, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusUnauthorized, res.Code)
}
