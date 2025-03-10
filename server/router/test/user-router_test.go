package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/google/uuid"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestUserCRUD(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	// 1. Create
	username := uuid.New().String() + "@test.com"
	payload := "{\"email\": \"" + username + "\", \"password\": \"12345678\", \"role\": " + strconv.Itoa(int(UserRoleOrgAdmin)) + "}"
	req := NewHTTPRequest("POST", "/user/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	userID := res.Header().Get("X-Object-Id")

	// 2. Read
	req = NewHTTPRequest("GET", "/user/"+userID, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetUserResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestString(t, username, resBody.Email)
	CheckTestString(t, org.ID, resBody.OrganizationID)
	CheckTestString(t, "", resBody.AuthProviderID)
	CheckTestBool(t, true, resBody.RequirePassword)
	CheckTestInt(t, int(UserRoleOrgAdmin), resBody.Role)
	CheckTestBool(t, true, resBody.SpaceAdmin)
	CheckTestBool(t, true, resBody.OrgAdmin)
	CheckTestBool(t, false, resBody.SuperAdmin)

	// 3. Update
	username = uuid.New().String() + "@test.com"
	payload = "{\"email\": \"" + username + "\", \"password\": \"\", \"role\": " + strconv.Itoa(int(UserRoleSpaceAdmin)) + "}"
	req = NewHTTPRequest("PUT", "/user/"+userID, loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Read
	req = NewHTTPRequest("GET", "/user/"+userID, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody2 *GetUserResponse
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	CheckTestString(t, username, resBody2.Email)
	CheckTestString(t, org.ID, resBody2.OrganizationID)
	CheckTestString(t, "", resBody2.AuthProviderID)
	CheckTestBool(t, true, resBody2.RequirePassword)
	CheckTestInt(t, int(UserRoleSpaceAdmin), resBody2.Role)
	CheckTestBool(t, true, resBody2.SpaceAdmin)
	CheckTestBool(t, false, resBody2.OrgAdmin)
	CheckTestBool(t, false, resBody2.SuperAdmin)

	// 4. Delete
	req = NewHTTPRequest("DELETE", "/user/"+userID, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Read
	req = NewHTTPRequest("GET", "/user/"+userID, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestUserForbidden(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	loginResponse := LoginTestUser(user.ID)

	// 1. Create
	username := uuid.New().String() + "@test.com"
	payload := "{\"email\": \"" + username + "\", \"password\": \"12345678\"}"
	req := NewHTTPRequest("POST", "/user/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	// 2. Read
	req = NewHTTPRequest("GET", "/user/"+user.ID, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	// 3. Update
	req = NewHTTPRequest("PUT", "/user/"+user.ID, loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	// 4. Delete
	req = NewHTTPRequest("DELETE", "/user/"+user.ID, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestUserSetPassword(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	loginResponse := LoginTestUser(user.ID)

	payload := `{"password": "12345678"}`
	req := NewHTTPRequest("PUT", "/user/"+user.ID+"/password", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	user2, err := GetUserRepository().GetOne(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestBool(t, true, GetUserRepository().CheckPassword(string(user2.HashedPassword), "12345678"))
}

func TestUserSubscriptionExceeded(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	GetSettingsRepository().Set(org.ID, SettingSubscriptionMaxUsers.Name, "1")

	username := uuid.New().String() + "@test.com"
	payload := "{\"email\": \"" + username + "\", \"password\": \"12345678\"}"
	req := NewHTTPRequest("POST", "/user/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusPaymentRequired, res.Code)
}

func TestUserGetCount(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	req := NewHTTPRequest("GET", "/user/count", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetUserCountResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, 1, resBody.Count)
}

func TestUserMergeUsers(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	source := CreateTestUserInOrg(org)
	target := CreateTestUserInOrg(org)

	// Prepare source
	source.AtlassianID = NullString(source.Email)
	GetUserRepository().Update(source)

	// Init from source
	loginResponseSource := LoginTestUser(source.ID)
	payload := "{\"email\": \"" + target.Email + "\"}"
	req := NewHTTPRequest("POST", "/user/merge/init", loginResponseSource.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Get merge request list from target
	loginResponseTarget := LoginTestUser(target.ID)
	req = NewHTTPRequest("GET", "/user/merge", loginResponseTarget.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []GetMergeRequestResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, 1, len(resBody))
	CheckTestString(t, source.ID, resBody[0].UserID)
	CheckTestString(t, source.Email, resBody[0].Email)

	// Complete from target
	req = NewHTTPRequest("POST", "/user/merge/finish/"+resBody[0].ID, loginResponseTarget.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Check if source user is gone
	user, err := GetUserRepository().GetOne(source.ID)
	if err == nil || user != nil {
		t.Fatal("Expected source user to be deleted")
	}

	// Check if target user has inherited source user's properties
	user, err = GetUserRepository().GetOne(target.ID)
	if err != nil || user == nil {
		t.Fatal("Expected source user to be deleted")
	}
	CheckTestString(t, string(source.AtlassianID), string(user.AtlassianID))

	// Check if request is invalid now
	req = NewHTTPRequest("POST", "/user/merge/finish/"+resBody[0].ID, loginResponseTarget.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

// TODO test domain in org!

func TestUserCreateForeignOrgSuperAdmin(t *testing.T) {
	ClearTestDB()
	superAdmin := CreateTestUserSuperAdmin()
	org2 := CreateTestOrg("test2.com")
	loginResponse := LoginTestUser(superAdmin.ID)

	username := uuid.New().String() + "@test2.com"
	payload := "{\"email\": \"" + username + "\", \"password\": \"12345678\", \"organizationId\": \"" + org2.ID + "\"}"
	req := NewHTTPRequest("POST", "/user/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
}

func TestUserCreateForeignOrgOrgAdmin(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test1.com")
	admin := CreateTestUserOrgAdmin(org)
	org2 := CreateTestOrg("test2.com")
	loginResponse := LoginTestUser(admin.ID)

	username := uuid.New().String() + "@test.com"
	payload := "{\"email\": \"" + username + "\", \"password\": \"12345678\", \"organizationId\": \"" + org2.ID + "\"}"
	req := NewHTTPRequest("POST", "/user/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestUserForeignEmail(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	username := uuid.New().String() + "@gmail.com"
	payload := "{\"email\": \"" + username + "\", \"password\": \"12345678\", \"role\": " + strconv.Itoa(int(UserRoleOrgAdmin)) + "}"
	req := NewHTTPRequest("POST", "/user/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
}

func TestUserDuplicateSameOrg(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	username := uuid.New().String() + "@gmail.com"

	payload := "{\"email\": \"" + username + "\", \"password\": \"12345678\", \"role\": " + strconv.Itoa(int(UserRoleOrgAdmin)) + "}"
	req := NewHTTPRequest("POST", "/user/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	payload = "{\"email\": \"" + username + "\", \"password\": \"12345678\", \"role\": " + strconv.Itoa(int(UserRoleOrgAdmin)) + "}"
	req = NewHTTPRequest("POST", "/user/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusConflict, res.Code)
}

func TestUserDuplicateDifferentOrg(t *testing.T) {
	ClearTestDB()
	org1 := CreateTestOrg("test1.com")
	user1 := CreateTestUserOrgAdmin(org1)
	org2 := CreateTestOrg("test2.com")
	user2 := CreateTestUserOrgAdmin(org2)

	username := uuid.New().String() + "@gmail.com"

	loginResponse1 := LoginTestUser(user1.ID)
	payload := "{\"email\": \"" + username + "\", \"password\": \"12345678\", \"role\": " + strconv.Itoa(int(UserRoleOrgAdmin)) + "}"
	req := NewHTTPRequest("POST", "/user/", loginResponse1.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	loginResponse2 := LoginTestUser(user2.ID)
	payload = "{\"email\": \"" + username + "\", \"password\": \"12345678\", \"role\": " + strconv.Itoa(int(UserRoleOrgAdmin)) + "}"
	req = NewHTTPRequest("POST", "/user/", loginResponse2.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
}

func TestUserUpdateCreatesDuplicate(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	username1 := uuid.New().String() + "@gmail.com"
	username2 := uuid.New().String() + "@gmail.com"

	payload := "{\"email\": \"" + username1 + "\", \"password\": \"12345678\", \"role\": " + strconv.Itoa(int(UserRoleOrgAdmin)) + "}"
	req := NewHTTPRequest("POST", "/user/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	payload = "{\"email\": \"" + username2 + "\", \"password\": \"12345678\", \"role\": " + strconv.Itoa(int(UserRoleOrgAdmin)) + "}"
	req = NewHTTPRequest("POST", "/user/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	userID2 := res.Header().Get("X-Object-Id")

	payload = "{\"email\": \"" + username1 + "\", \"password\": \"\", \"role\": " + strconv.Itoa(int(UserRoleSpaceAdmin)) + "}"
	req = NewHTTPRequest("PUT", "/user/"+userID2, loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusConflict, res.Code)
}

func TestUserCreateInOwnOrgsVerifiedDomain(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	GetOrganizationRepository().AddDomain(org, "gmail.com", true)
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	username := uuid.New().String() + "@gmail.com"

	payload := "{\"email\": \"" + username + "\", \"password\": \"12345678\", \"role\": " + strconv.Itoa(int(UserRoleOrgAdmin)) + "}"
	req := NewHTTPRequest("POST", "/user/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
}
