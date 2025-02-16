package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestSpaceAttributesEmptyResult(t *testing.T) {
	ClearTestDB()
	loginResponse := CreateLoginTestUser()

	req := NewHTTPRequest("GET", "/space-attribute/", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []string
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 0 {
		t.Fatalf("Expected empty array")
	}
}

func TestSpaceAttributesCRUD(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	// 1. Create
	payload := `{"label": "Test 123", "type": 3, "spaceApplicable": true, "locationApplicable": false}`
	req := NewHTTPRequest("POST", "/space-attribute/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-Id")

	// 2. Read
	req = NewHTTPRequest("GET", "/space-attribute/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetSpaceAttributeResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestString(t, "Test 123", resBody.Label)
	CheckTestBool(t, true, resBody.SpaceApplicable)
	CheckTestBool(t, false, resBody.LocationApplicable)

	// 3. Update
	payload = `{"label": "Test 456", "type": 2, "spaceApplicable": false, "locationApplicable": true}`
	req = NewHTTPRequest("PUT", "/space-attribute/"+id, loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Read
	req = NewHTTPRequest("GET", "/space-attribute/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody2 *GetSpaceAttributeResponse
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	CheckTestString(t, "Test 456", resBody2.Label)
	CheckTestBool(t, false, resBody2.SpaceApplicable)
	CheckTestBool(t, true, resBody2.LocationApplicable)

	// 4. Delete
	req = NewHTTPRequest("DELETE", "/space-attribute/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Read
	req = NewHTTPRequest("GET", "/space-attribute/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}
