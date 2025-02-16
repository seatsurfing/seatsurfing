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

func TestSpacesSameOrgForbidden(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user2 := CreateTestUserOrgAdmin(org)
	loginResponse2 := LoginTestUser(user2.ID)

	// Create location
	payload := `{"name": "Location 1"}`
	req := NewHTTPRequest("POST", "/location/", loginResponse2.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-Id")

	// Create space
	payload = `{"name": "H234", "x": 50, "y": 100, "width": 200, "height": 300, "rotation": 90}`
	req = NewHTTPRequest("POST", "/location/"+id+"/space/", loginResponse2.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	spaceID := res.Header().Get("X-Object-Id")

	// Switch to non-admin user
	user := CreateTestUserInOrg(org)
	loginResponse := LoginTestUser(user.ID)

	payload = `{"name": "Location 1"}`
	req = NewHTTPRequest("POST", "/location/"+id+"/space/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	payload = `{"name": "Location 1"}`
	req = NewHTTPRequest("PUT", "/location/"+id+"/space/"+spaceID, loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	req = NewHTTPRequest("DELETE", "/location/"+id+"/space/"+spaceID, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestSpacesEmptyResult(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	// Create location
	payload := `{"name": "Location 1"}`
	req := NewHTTPRequest("POST", "/location/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-Id")

	// Get spaces
	req = NewHTTPRequest("GET", "/location/"+id+"/space/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []string
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 0 {
		t.Fatalf("Expected empty array")
	}
}

func TestSpacesCRUD(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	// Create location
	payload := `{"name": "Location 1"}`
	req := NewHTTPRequest("POST", "/location/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	locationID := res.Header().Get("X-Object-Id")

	// 1. Create
	payload = `{"name": "H234", "x": 50, "y": 100, "width": 200, "height": 300, "rotation": 90}`
	req = NewHTTPRequest("POST", "/location/"+locationID+"/space/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-Id")

	// 2. Read
	req = NewHTTPRequest("GET", "/location/"+locationID+"/space/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetSpaceResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestString(t, "H234", resBody.Name)
	CheckTestUint(t, 50, resBody.X)
	CheckTestUint(t, 100, resBody.Y)
	CheckTestUint(t, 200, resBody.Width)
	CheckTestUint(t, 300, resBody.Height)
	CheckTestUint(t, 90, resBody.Rotation)

	// 3. Update
	payload = `{"name": "H235", "x": 51, "y": 101, "width": 201, "height": 301, "rotation": 91}`
	req = NewHTTPRequest("PUT", "/location/"+locationID+"/space/"+id, loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Read
	req = NewHTTPRequest("GET", "/location/"+locationID+"/space/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody2 *GetSpaceResponse
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	CheckTestString(t, "H235", resBody2.Name)
	CheckTestUint(t, 51, resBody2.X)
	CheckTestUint(t, 101, resBody2.Y)
	CheckTestUint(t, 201, resBody2.Width)
	CheckTestUint(t, 301, resBody2.Height)
	CheckTestUint(t, 91, resBody2.Rotation)

	// 4. Delete
	req = NewHTTPRequest("DELETE", "/location/"+locationID+"/space/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Read
	req = NewHTTPRequest("GET", "/location/"+locationID+"/space/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestSpacesBulkUpdate(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	// Create location
	payload := `{"name": "Location 1"}`
	req := NewHTTPRequest("POST", "/location/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	locationID := res.Header().Get("X-Object-Id")

	// 1. Create 3 spaces
	payload = `{
		"creates": [
			{"name": "H1", "x": 50, "y": 110, "width": 210, "height": 310, "rotation": 90},
			{"name": "H2", "x": 60, "y": 120, "width": 220, "height": 320, "rotation": 91},
			{"name": "H3", "x": 70, "y": 130, "width": 230, "height": 330, "rotation": 92}
		]
	}`
	req = NewHTTPRequest("POST", "/location/"+locationID+"/space/bulk", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *BulkUpdateResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, 3, len(resBody.Creates))
	CheckTestInt(t, 0, len(resBody.Updates))
	CheckTestInt(t, 0, len(resBody.Deletes))
	CheckTestBool(t, true, resBody.Creates[0].Success)
	CheckTestBool(t, true, resBody.Creates[1].Success)
	CheckTestBool(t, true, resBody.Creates[2].Success)

	// 2. Create, Update, Delete
	payload = `{
		"creates": [
			{"name": "H4", "x": 80, "y": 140, "width": 240, "height": 340, "rotation": 93}
		],
		"updates": [
			{"id": "` + resBody.Creates[1].ID + `", "name": "H2.2", "x": 69, "y": 129, "width": 229, "height": 329, "rotation": 99}
		],
		"deleteIds": [
			"` + resBody.Creates[2].ID + `"
		]
	}`
	req = NewHTTPRequest("POST", "/location/"+locationID+"/space/bulk", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, 1, len(resBody.Creates))
	CheckTestInt(t, 1, len(resBody.Updates))
	CheckTestInt(t, 1, len(resBody.Deletes))
	CheckTestBool(t, true, resBody.Creates[0].Success)
	CheckTestBool(t, true, resBody.Updates[0].Success)
	CheckTestBool(t, true, resBody.Deletes[0].Success)

	// 3. List
	req = NewHTTPRequest("GET", "/location/"+locationID+"/space/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody2 []*GetSpaceResponse
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	if len(resBody2) != 3 {
		t.Fatalf("Expected array with 3 elements")
	}
	CheckTestString(t, "H1", resBody2[0].Name)
	CheckTestString(t, "H2.2", resBody2[1].Name)
	CheckTestString(t, "H4", resBody2[2].Name)
}

func TestSpacesList(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	locationID, _, _, _ := createTestSpaces(t, loginResponse)

	req := NewHTTPRequest("GET", "/location/"+locationID+"/space/", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []*GetSpaceResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 3 {
		t.Fatalf("Expected array with 3 elements")
	}
	CheckTestString(t, "H234", resBody[0].Name)
	CheckTestString(t, "H235", resBody[1].Name)
	CheckTestString(t, "H236", resBody[2].Name)
}

func TestSpacesAvailabilityOuter(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)
	GetSettingsRepository().Set(org.ID, SettingMaxDaysInAdvance.Name, "5000")

	locationID, spaceID, _, _ := createTestSpaces(t, loginResponse)

	// Create booking
	payload := "{\"spaceId\": \"" + spaceID + "\", \"enter\": \"2030-09-01T06:00:00+02:00\", \"leave\": \"2030-09-01T18:00:00+02:00\"}"
	req := NewHTTPRequest("POST", "/booking/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	// Check
	payload = `{"enter": "2030-09-01T08:30:00+02:00", "leave": "2030-09-01T17:00:00+02:00"}`
	req = NewHTTPRequest("POST", "/location/"+locationID+"/space/availability", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []*GetSpaceResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 3 {
		t.Fatalf("Expected array with 3 elements")
	}
	CheckTestString(t, "H234", resBody[0].Name)
	CheckTestString(t, "H235", resBody[1].Name)
	CheckTestString(t, "H236", resBody[2].Name)
	CheckTestBool(t, false, resBody[0].Available)
	CheckTestBool(t, true, resBody[1].Available)
	CheckTestBool(t, true, resBody[2].Available)
}

func TestSpacesAvailabilityInner(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)
	GetSettingsRepository().Set(org.ID, SettingMaxDaysInAdvance.Name, "5000")

	locationID, spaceID, _, _ := createTestSpaces(t, loginResponse)

	// Create booking
	payload := "{\"spaceId\": \"" + spaceID + "\", \"enter\": \"2030-09-01T09:00:00+02:00\", \"leave\": \"2030-09-01T11:00:00+02:00\"}"
	req := NewHTTPRequest("POST", "/booking/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	// Check
	payload = `{"enter": "2020-09-01T08:30:00+02:00", "leave": "2030-09-01T17:00:00+02:00"}`
	req = NewHTTPRequest("POST", "/location/"+locationID+"/space/availability", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []*GetSpaceResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 3 {
		t.Fatalf("Expected array with 3 elements")
	}
	CheckTestString(t, "H234", resBody[0].Name)
	CheckTestString(t, "H235", resBody[1].Name)
	CheckTestString(t, "H236", resBody[2].Name)
	CheckTestBool(t, false, resBody[0].Available)
	CheckTestBool(t, true, resBody[1].Available)
	CheckTestBool(t, true, resBody[2].Available)
}

func TestSpacesAvailabilityStart(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)
	GetSettingsRepository().Set(org.ID, SettingMaxDaysInAdvance.Name, "5000")

	locationID, spaceID, _, _ := createTestSpaces(t, loginResponse)

	// Create booking
	payload := "{\"spaceId\": \"" + spaceID + "\", \"enter\": \"2030-09-01T07:00:00Z\", \"leave\": \"2030-09-01T09:00:00Z\"}"
	req := NewHTTPRequest("POST", "/booking/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	// Check
	payload = `{"enter": "2030-09-01T08:30:00Z", "leave": "2030-09-01T17:00:00Z"}`
	req = NewHTTPRequest("POST", "/location/"+locationID+"/space/availability", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []*GetSpaceResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 3 {
		t.Fatalf("Expected array with 3 elements")
	}
	CheckTestString(t, "H234", resBody[0].Name)
	CheckTestString(t, "H235", resBody[1].Name)
	CheckTestString(t, "H236", resBody[2].Name)
	CheckTestBool(t, false, resBody[0].Available)
	CheckTestBool(t, true, resBody[1].Available)
	CheckTestBool(t, true, resBody[2].Available)
}

func TestSpacesAvailabilityEnd(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)
	GetSettingsRepository().Set(org.ID, SettingMaxDaysInAdvance.Name, "5000")

	locationID, spaceID, _, _ := createTestSpaces(t, loginResponse)

	// Create booking
	payload := "{\"spaceId\": \"" + spaceID + "\", \"enter\": \"2030-09-01T16:30:00+02:00\", \"leave\": \"2030-09-01T17:30:00+02:00\"}"
	req := NewHTTPRequest("POST", "/booking/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	// Check
	payload = `{"enter": "2030-09-01T08:30:00+02:00", "leave": "2030-09-01T17:00:00+02:00"}`
	req = NewHTTPRequest("POST", "/location/"+locationID+"/space/availability", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []*GetSpaceResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 3 {
		t.Fatalf("Expected array with 3 elements")
	}
	CheckTestString(t, "H234", resBody[0].Name)
	CheckTestString(t, "H235", resBody[1].Name)
	CheckTestString(t, "H236", resBody[2].Name)
	CheckTestBool(t, false, resBody[0].Available)
	CheckTestBool(t, true, resBody[1].Available)
	CheckTestBool(t, true, resBody[2].Available)
}

func TestSpacesAvailabilityNoBookings(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	locationID, _, _, _ := createTestSpaces(t, loginResponse)

	payload := `{"enter": "2020-09-01T08:30:00+02:00", "leave": "2020-09-01T17:00:00+02:00"}`
	req := NewHTTPRequest("POST", "/location/"+locationID+"/space/availability", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []*GetSpaceResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 3 {
		t.Fatalf("Expected array with 3 elements")
	}
	CheckTestString(t, "H234", resBody[0].Name)
	CheckTestString(t, "H235", resBody[1].Name)
	CheckTestString(t, "H236", resBody[2].Name)
	CheckTestBool(t, true, resBody[0].Available)
	CheckTestBool(t, true, resBody[1].Available)
	CheckTestBool(t, true, resBody[2].Available)
}

func createTestSpaces(t *testing.T, loginResponse *LoginResponse) (lID, s1ID, s2ID, s3ID string) {
	// Create location
	payload := `{"name": "Location 1"}`
	req := NewHTTPRequest("POST", "/location/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	locationID := res.Header().Get("X-Object-Id")

	// Create #1
	payload = `{"name": "H234", "x": 50, "y": 100, "width": 200, "height": 300, "rotation": 90}`
	req = NewHTTPRequest("POST", "/location/"+locationID+"/space/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	space1ID := res.Header().Get("X-Object-Id")

	// Create #2
	payload = `{"name": "H236", "x": 50, "y": 100, "width": 200, "height": 300, "rotation": 90}`
	req = NewHTTPRequest("POST", "/location/"+locationID+"/space/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	space2ID := res.Header().Get("X-Object-Id")

	// Create #3
	payload = `{"name": "H235", "x": 50, "y": 100, "width": 200, "height": 300, "rotation": 90}`
	req = NewHTTPRequest("POST", "/location/"+locationID+"/space/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	space3ID := res.Header().Get("X-Object-Id")

	return locationID, space1ID, space2ID, space3ID
}
