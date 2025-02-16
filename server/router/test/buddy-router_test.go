package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestBuddiesEmptyResult(t *testing.T) {
	ClearTestDB()
	loginResponse := CreateLoginTestUser()

	req := NewHTTPRequest("GET", "/buddy/", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []string
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 0 {
		t.Fatalf("Expected empty array")
	}
}

func TestBuddiesCRUD(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org.ID, SettingMaxDaysInAdvance.Name, "5000")

	// Create buddy users
	buddyUser1 := CreateTestUserInOrg(org)

	// Switch to non-admin user
	user := CreateTestUserInOrg(org)
	loginResponse := LoginTestUser(user.ID)

	// 1. Create
	payload := "{\"buddyId\": \"" + buddyUser1.ID + "\"}"
	req := NewHTTPRequest("POST", "/buddy/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-Id")

	// 2. Read all buddies and ensure buddy was created correctly
	req = NewHTTPRequest("GET", "/buddy/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []*GetBuddyResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 1 {
		t.Fatalf("Expected array with 1 element")
	}
	CheckTestString(t, buddyUser1.ID, resBody[0].BuddyID)
	CheckTestString(t, id, resBody[0].ID)

	// 3. Delete
	req = NewHTTPRequest("DELETE", "/buddy/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// 4. Read all buddies and ensure buddy was removed correctly
	req = NewHTTPRequest("GET", "/buddy/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody2 []*GetBuddyResponse
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	if len(resBody2) != 0 {
		t.Fatalf("Expected empty array")
	}
}

func TestDeleteBuddyOfAnotherUser(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org.ID, SettingMaxDaysInAdvance.Name, "5000")

	// Create buddy users
	buddyUser1 := CreateTestUserInOrg(org)

	// Switch to non-admin user
	user := CreateTestUserInOrg(org)
	user2 := CreateTestUserInOrg(org)
	loginResponse := LoginTestUser(user.ID)

	// Create
	payload := "{\"buddyId\": \"" + buddyUser1.ID + "\"}"
	req := NewHTTPRequest("POST", "/buddy/", user2.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-Id")

	// Delete
	req = NewHTTPRequest("DELETE", "/buddy/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestBuddiesCreateWithMissingUser(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user2 := CreateTestUserOrgAdmin(org)
	loginResponse2 := LoginTestUser(user2.ID)
	GetSettingsRepository().Set(org.ID, SettingMaxDaysInAdvance.Name, "5000")
	GetSettingsRepository().Set(org.ID, SettingAllowBookingsNonExistingUsers.Name, "1")

	// Create
	payload := "{\"buddyId\": \"" + uuid.New().String() + "\"}"
	req := NewHTTPRequest("POST", "/buddy/", loginResponse2.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
}

func TestBuddiesList(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org.ID, SettingMaxDaysInAdvance.Name, "5000")

	// Create buddy users
	buddyUser1 := CreateTestUserInOrg(org)
	buddyUser2 := CreateTestUserInOrg(org)
	buddyUser3 := CreateTestUserInOrg(org)

	// Switch to non-admin user
	user := CreateTestUserInOrg(org)
	loginResponse := LoginTestUser(user.ID)

	// Create #1
	payload := "{\"buddyId\": \"" + buddyUser1.ID + "\"}"
	req := NewHTTPRequest("POST", "/buddy/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	// Create #2
	payload = "{\"buddyId\": \"" + buddyUser2.ID + "\"}"
	req = NewHTTPRequest("POST", "/buddy/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	// Create #3(for a different user)
	payload = "{\"buddyId\": \"" + buddyUser3.ID + "\"}"
	req = NewHTTPRequest("POST", "/buddy/", buddyUser2.ID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	// Read all buddies for user 1
	req = NewHTTPRequest("GET", "/buddy/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []*GetBuddyResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 2 {
		t.Fatalf("Expected array with 2 elements")
	}
	acceptedBuddyIDs := []string{buddyUser1.ID, buddyUser2.ID}
	if !Contains(acceptedBuddyIDs, resBody[0].BuddyID) {
		t.Fatalf("Expected %s to one of %#v", resBody[0].BuddyID, acceptedBuddyIDs)
	}
	if !Contains(acceptedBuddyIDs, resBody[1].BuddyID) {
		t.Fatalf("Expected %s to one of %#v", resBody[1].BuddyID, acceptedBuddyIDs)
	}

	// Read all buddies for user 2
	req = NewHTTPRequest("GET", "/buddy/", buddyUser2.ID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody2 []*GetBuddyResponse
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	if len(resBody2) != 1 {
		t.Fatalf("Expected array with 1 elements")
	}
	CheckTestString(t, buddyUser3.ID, resBody2[0].BuddyID)
}
