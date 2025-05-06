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

func TestGroupsEmptyResult(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	req := NewHTTPRequest("GET", "/group/", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []string
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 0 {
		t.Fatalf("Expected empty array")
	}
}

func TestGroupsForbidden(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	loginResponse := LoginTestUser(user.ID)

	// Create
	payload := `{"name": "G1"}`
	req := NewHTTPRequest("POST", "/location/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	// List
	req = NewHTTPRequest("GET", "/group/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestGroupsCRUD(t *testing.T) {
	ClearTestDB()

	org := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org.ID, SettingFeatureGroups.Name, "1")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	// 1. Create
	payload := `{"name": "G1"}`
	req := NewHTTPRequest("POST", "/group/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-Id")

	// 2. Read
	req = NewHTTPRequest("GET", "/group/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetGroupResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestString(t, "G1", resBody.Name)

	// 3. Update
	payload = `{"name": "G1.2"}`
	req = NewHTTPRequest("PUT", "/group/"+id, loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Read
	req = NewHTTPRequest("GET", "/group/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody2 *GetGroupResponse
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	CheckTestString(t, "G1.2", resBody2.Name)

	// 4. Delete
	req = NewHTTPRequest("DELETE", "/group/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Read
	req = NewHTTPRequest("GET", "/group/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestGroupsList(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org.ID, SettingFeatureGroups.Name, "1")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	payload := `{"name": "G 1"}`
	req := NewHTTPRequest("POST", "/group/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	payload = `{"name": "G 2"}`
	req = NewHTTPRequest("POST", "/group/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	payload = `{"name": "G 0"}`
	req = NewHTTPRequest("POST", "/group/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	req = NewHTTPRequest("GET", "/group/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []*GetGroupResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 3 {
		t.Fatalf("Expected array with 3 elements")
	}
	CheckTestString(t, "G 0", resBody[0].Name)
	CheckTestString(t, "G 1", resBody[1].Name)
	CheckTestString(t, "G 2", resBody[2].Name)
}

func TestGroupsMembersCRUD(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org.ID, SettingFeatureGroups.Name, "1")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	payload := `{"name": "G 1"}`
	req := NewHTTPRequest("POST", "/group/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-Id")

	u1 := CreateTestUserInOrg(org)
	u2 := CreateTestUserInOrg(org)
	u3 := CreateTestUserInOrg(org)
	u4 := CreateTestUserInOrg(org)

	userIDs := []string{u1.ID, u2.ID, u3.ID}
	userIDsJson, _ := json.Marshal(userIDs)
	req = NewHTTPRequest("PUT", "/group/"+id+"/member", loginResponse.UserID, bytes.NewBuffer(userIDsJson))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	userIDs = []string{u4.ID}
	userIDsJson, _ = json.Marshal(userIDs)
	req = NewHTTPRequest("PUT", "/group/"+id+"/member", loginResponse.UserID, bytes.NewBuffer(userIDsJson))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	userIDs = []string{u2.ID}
	userIDsJson, _ = json.Marshal(userIDs)
	req = NewHTTPRequest("POST", "/group/"+id+"/member/remove", loginResponse.UserID, bytes.NewBuffer(userIDsJson))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	req = NewHTTPRequest("GET", "/group/"+id+"/member", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []*GetUserResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 3 {
		t.Fatalf("Expected array with 3 elements")
	}
}

func TestGroupsMembersAddForeignOrg(t *testing.T) {
	ClearTestDB()
	org1 := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org1.ID, SettingFeatureGroups.Name, "1")
	org2 := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org2.ID, SettingFeatureGroups.Name, "1")
	user := CreateTestUserOrgAdmin(org1)
	loginResponse := LoginTestUser(user.ID)

	payload := `{"name": "G 1"}`
	req := NewHTTPRequest("POST", "/group/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-Id")

	u1 := CreateTestUserInOrg(org1)
	u2 := CreateTestUserInOrg(org2)

	userIDs := []string{u1.ID, u2.ID}
	userIDsJson, _ := json.Marshal(userIDs)
	req = NewHTTPRequest("PUT", "/group/"+id+"/member", loginResponse.UserID, bytes.NewBuffer(userIDsJson))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
}
