package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"slices"
	"strconv"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestRoleRouterCRUD(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(admin.ID)

	// Create
	payload := `{"name": "Group Manager", "description": "Manages groups", "permissions": {"groups": 30}}`
	req := NewHTTPRequest("POST", "/role/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-Id")

	// Read
	req = NewHTTPRequest("GET", "/role/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetRoleResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestString(t, "Group Manager", resBody.Name)
	CheckTestBool(t, false, resBody.System)
	CheckTestInt(t, int(PermissionLevelAdmin), resBody.Permissions["groups"])

	// Update
	payload = `{"name": "Group Reader", "permissions": {"groups": 10}}`
	req = NewHTTPRequest("PUT", "/role/"+id, loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	req = NewHTTPRequest("GET", "/role/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	var resBody2 *GetRoleResponse
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	CheckTestString(t, "Group Reader", resBody2.Name)
	CheckTestInt(t, int(PermissionLevelRead), resBody2.Permissions["groups"])

	// Delete
	req = NewHTTPRequest("DELETE", "/role/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	req = NewHTTPRequest("GET", "/role/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestRoleRouterForbiddenWithoutPermission(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	loginResponse := LoginTestUser(user.ID)

	for _, c := range []struct {
		method string
		url    string
	}{
		{"GET", "/role/"},
		{"POST", "/role/"},
	} {
		req := NewHTTPRequest(c.method, c.url, loginResponse.UserID, bytes.NewBufferString(`{"name": "X"}`))
		res := ExecuteTestRequest(req)
		CheckTestResponseCode(t, http.StatusForbidden, res.Code)
	}
}

// The permission catalogue describes the model rather than the caller's own
// access, so any authenticated user may read it.
func TestRoleRouterPermissionCatalogue(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	loginResponse := LoginTestUser(user.ID)

	req := NewHTTPRequest("GET", "/role/permissions", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []*GetPermissionDefinitionResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, len(GetPermissionDefinitions()), len(resBody))

	byKey := map[string][]int{}
	for _, d := range resBody {
		byKey[d.Key] = d.AllowedLevels
	}
	// Only the levels that mean something for a permission are offered.
	CheckTestInt(t, 2, len(byKey[string(PermissionAreas)]))
	CheckTestInt(t, 4, len(byKey[string(PermissionGroups)]))
	CheckTestInt(t, 2, len(byKey[string(PermissionAnalytics)]))
}

// A level a permission does not declare must be refused rather than silently
// dropped: a typo should not quietly produce a weaker role than intended.
func TestRoleRouterRejectsUndeclaredLevel(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(admin.ID)

	for _, payload := range []string{
		`{"name": "Bad", "permissions": {"analytics": 30}}`,
		`{"name": "Bad", "permissions": {"areas": 10}}`,
		`{"name": "Bad", "permissions": {"nonexistent": 30}}`,
	} {
		req := NewHTTPRequest("POST", "/role/", loginResponse.UserID, bytes.NewBufferString(payload))
		res := ExecuteTestRequest(req)
		CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
	}
}

func TestRoleRouterSystemRoleIsProtected(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(admin.ID)

	role, err := GetRoleRepository().GetByName(org.ID, RoleNameOrgAdmin)
	if err != nil {
		t.Fatal(err)
	}

	payload := `{"name": "Renamed", "permissions": {"groups": 30}}`
	req := NewHTTPRequest("PUT", "/role/"+role.ID, loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
	CheckTestString(t, strconv.Itoa(ResponseCodeRoleIsSystemRole), res.Header().Get("X-Error-Code"))

	req = NewHTTPRequest("DELETE", "/role/"+role.ID, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
	CheckTestString(t, strconv.Itoa(ResponseCodeRoleIsSystemRole), res.Header().Get("X-Error-Code"))
}

func TestRoleRouterOrgIsolation(t *testing.T) {
	ClearTestDB()
	org1 := CreateTestOrg("test1.com")
	org2 := CreateTestOrg("test2.com")
	admin1 := CreateTestUserOrgAdmin(org1)
	loginResponse := LoginTestUser(admin1.ID)
	foreignRole := CreateTestRole(org2, "Foreign", map[Permission]PermissionLevel{PermissionGroups: PermissionLevelAdmin})

	req := NewHTTPRequest("GET", "/role/"+foreignRole.ID, loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)

	req = NewHTTPRequest("DELETE", "/role/"+foreignRole.ID, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

// Nobody may hand out access they do not hold themselves.
func TestRoleRouterNoEscalation(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	CreateTestUserOrgAdmin(org) // keeps the organization's admin invariant satisfied

	// This user may manage roles, but has no access to settings at all.
	limited := CreateTestUserWithPermissions(org, map[Permission]PermissionLevel{
		PermissionRoles:  PermissionLevelAdmin,
		PermissionGroups: PermissionLevelRead,
	})
	loginResponse := LoginTestUser(limited.ID)

	payload := `{"name": "Too Powerful", "permissions": {"org_settings": 30}}`
	req := NewHTTPRequest("POST", "/role/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
	CheckTestString(t, strconv.Itoa(ResponseCodeRoleEscalationNotAllowed), res.Header().Get("X-Error-Code"))

	// Granting a higher level of a permission they do hold is refused too.
	payload = `{"name": "Group Boss", "permissions": {"groups": 30}}`
	req = NewHTTPRequest("POST", "/role/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
	CheckTestString(t, strconv.Itoa(ResponseCodeRoleEscalationNotAllowed), res.Header().Get("X-Error-Code"))

	// What they do hold, at their own level, is fine.
	payload = `{"name": "Group Viewer", "permissions": {"groups": 10}}`
	req = NewHTTPRequest("POST", "/role/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
}

func TestUserRolesAssignment(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(admin.ID)
	target := CreateTestUserInOrg(org)
	role := CreateTestRole(org, "Group Manager", map[Permission]PermissionLevel{PermissionGroups: PermissionLevelAdmin})

	payload := `{"roleIds": ["` + role.ID + `"]}`
	req := NewHTTPRequest("PUT", "/user/"+target.ID+"/roles", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	req = NewHTTPRequest("GET", "/user/"+target.ID+"/roles", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var roleIDs []string
	json.Unmarshal(res.Body.Bytes(), &roleIDs)
	CheckTestInt(t, 1, len(roleIDs))
	CheckTestString(t, role.ID, roleIDs[0])

	// The resolved effect is visible before relying on it.
	req = NewHTTPRequest("GET", "/user/"+target.ID+"/permissions", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var perms *GetUserPermissionsResponse
	json.Unmarshal(res.Body.Bytes(), &perms)
	CheckTestInt(t, int(PermissionLevelAdmin), perms.Permissions["groups"])

	// The assignment actually grants access: this endpoint was forbidden before.
	targetLogin := LoginTestUser(target.ID)
	req = NewHTTPRequest("GET", "/group/", targetLogin.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	// Clearing the assignment revokes it again.
	req = NewHTTPRequest("PUT", "/user/"+target.ID+"/roles", loginResponse.UserID, bytes.NewBufferString(`{"roleIds": []}`))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)
	req = NewHTTPRequest("GET", "/group/", targetLogin.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

// Every user reads their own resolved access from /user/me.
func TestUserSelfExposesPermissions(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(admin.ID)

	req := NewHTTPRequest("GET", "/user/me", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetUserSelfResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, int(PermissionLevelAdmin), resBody.Permissions[string(PermissionOrgSettings)])
	CheckTestInt(t, len(GetPermissionDefinitions()), len(resBody.Permissions))
	CheckTestInt(t, 1, len(resBody.RoleIDs))

	// A user with no roles has no administrative access at all.
	plain := CreateTestUserInOrg(org)
	plainLogin := LoginTestUser(plain.ID)
	req = NewHTTPRequest("GET", "/user/me", plainLogin.UserID, nil)
	res = ExecuteTestRequest(req)
	var resBody2 *GetUserSelfResponse
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	CheckTestInt(t, 0, len(resBody2.Permissions))
}

// ─── Lock-out prevention ─────────────────────────────────────────────────────

func TestLockoutDeletingLastAdminRefused(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin1 := CreateTestUserOrgAdmin(org)
	admin2 := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(admin1.ID)

	// Two administrators: removing one is fine.
	req := NewHTTPRequest("DELETE", "/user/"+admin2.ID, loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Only one left. Deleting yourself is already refused, so promote another
	// user and have them try to delete the last remaining administrator.
	other := CreateTestUserWithPermissions(org, map[Permission]PermissionLevel{
		PermissionUsers: PermissionLevelAdmin,
	})
	otherLogin := LoginTestUser(other.ID)
	req = NewHTTPRequest("DELETE", "/user/"+admin1.ID, otherLogin.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
	CheckTestString(t, strconv.Itoa(ResponseCodeRoleWouldLeaveOrgWithoutAdmin), res.Header().Get("X-Error-Code"))

	// The user is still there.
	if _, err := GetUserRepository().GetOne(admin1.ID); err != nil {
		t.Fatal("expected the last administrator to survive")
	}
}

func TestLockoutRemovingOwnRoleManagementRefused(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(admin.ID)

	req := NewHTTPRequest("PUT", "/user/"+admin.ID+"/roles", loginResponse.UserID, bytes.NewBufferString(`{"roleIds": []}`))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)

	// The assignment survived, so the administrator can still act.
	roleIDs, _ := GetUserRoleRepository().GetRoleIDsForUser(admin.ID)
	CheckTestInt(t, 1, len(roleIDs))
}

func TestLockoutStrippingLastAdminRoleRefused(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(admin.ID)

	// A second administrator holding a custom, editable role.
	custom := CreateTestRole(org, "Custom Admin", map[Permission]PermissionLevel{
		PermissionRoles: PermissionLevelAdmin,
		PermissionUsers: PermissionLevelAdmin,
	})
	second := CreateTestUserInOrg(org)
	AssignTestRole(second, custom)

	// Removing the built-in administrator leaves the custom one, so this is
	// allowed; the organization still has an administrator.
	req := NewHTTPRequest("PUT", "/user/"+admin.ID+"/roles", loginResponse.UserID,
		bytes.NewBufferString(`{"roleIds": ["`+custom.ID+`"]}`))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Now weaken the only role granting administration. Both remaining
	// administrators hold it, so this would empty the organization.
	secondLogin := LoginTestUser(second.ID)
	payload := `{"name": "Custom Admin", "permissions": {"users": 30}}`
	req = NewHTTPRequest("PUT", "/role/"+custom.ID, secondLogin.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
	CheckTestString(t, strconv.Itoa(ResponseCodeRoleWouldLeaveOrgWithoutAdmin), res.Header().Get("X-Error-Code"))
}

// Disabled users and service accounts do not count towards the invariant: an
// organization whose only administrator is disabled is locked out just the
// same.
func TestLockoutDisabledAdminDoesNotCount(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)

	if OrgRetainsAdminWithout(org.ID) != true {
		t.Fatal("expected the organization to have an administrator")
	}
	admin.Disabled = true
	if err := GetUserRepository().Update(admin); err != nil {
		t.Fatal(err)
	}
	if OrgRetainsAdminWithout(org.ID) != false {
		t.Fatal("expected a disabled administrator not to count")
	}
}

// The start-up repair is the backstop for a database that reached a
// locked-out state outside the API.
func TestEnsureEveryOrgHasAdminRepairs(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	// Simulate a database in which nobody holds administrative access.
	if _, err := GetDatabase().DB().Exec("DELETE FROM user_roles"); err != nil {
		t.Fatal(err)
	}
	CheckTestBool(t, false, OrgRetainsAdminWithout(org.ID))

	EnsureEveryOrgHasAdmin()

	CheckTestBool(t, true, OrgRetainsAdminWithout(org.ID))
	roles, err := GetUserRoleRepository().GetRolesForUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, r := range roles {
		names = append(names, r.Name)
	}
	if !slices.Contains(names, RoleNameOrgAdmin) {
		t.Fatalf("expected the organization administrator role to be restored, got %v", names)
	}
}
