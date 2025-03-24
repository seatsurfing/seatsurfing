package test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestLocationsEmptyResult(t *testing.T) {
	ClearTestDB()
	loginResponse := CreateLoginTestUser()

	req := NewHTTPRequest("GET", "/location/", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []string
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 0 {
		t.Fatalf("Expected empty array")
	}
}

func TestLocationsForbidden(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	// Create
	payload := `{"name": "Location 1"}`
	req := NewHTTPRequest("POST", "/location/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-Id")

	org2 := CreateTestOrg("test2.com")
	user2 := CreateTestUserOrgAdmin(org2)
	loginResponse2 := LoginTestUser(user2.ID)

	// Get from other org
	req = NewHTTPRequest("GET", "/location/"+id, loginResponse2.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	// Update location from other org
	payload = `{"name": "Location 2"}`
	req = NewHTTPRequest("PUT", "/location/"+id, loginResponse2.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	// Delete location from other org
	req = NewHTTPRequest("DELETE", "/location/"+id, loginResponse2.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestLocationsCRUD(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	// 1. Create
	payload := `{"name": "Location 1"}`
	req := NewHTTPRequest("POST", "/location/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-Id")

	// 2. Read
	req = NewHTTPRequest("GET", "/location/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetLocationResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestString(t, "Location 1", resBody.Name)
	CheckTestString(t, "", resBody.Description)
	CheckTestString(t, "", resBody.Timezone)
	CheckTestInt(t, 0, int(resBody.MaxConcurrentBookings))

	// 3. Update
	payload = `{"name": "Location 2", "description": "Test 123", "maxConcurrentBookings": 20, "timezone": "Africa/Cairo"}`
	req = NewHTTPRequest("PUT", "/location/"+id, loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Read
	req = NewHTTPRequest("GET", "/location/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody2 *GetLocationResponse
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	CheckTestString(t, "Location 2", resBody2.Name)
	CheckTestString(t, "Test 123", resBody2.Description)
	CheckTestString(t, "Africa/Cairo", resBody2.Timezone)
	CheckTestInt(t, 20, int(resBody2.MaxConcurrentBookings))

	// 4. Delete
	req = NewHTTPRequest("DELETE", "/location/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Read
	req = NewHTTPRequest("GET", "/location/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestLocationsList(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	payload := `{"name": "Location 1"}`
	req := NewHTTPRequest("POST", "/location/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	payload = `{"name": "Location 2"}`
	req = NewHTTPRequest("POST", "/location/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	payload = `{"name": "Location 0"}`
	req = NewHTTPRequest("POST", "/location/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	req = NewHTTPRequest("GET", "/location/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []*GetLocationResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 3 {
		t.Fatalf("Expected array with 3 elements")
	}
	CheckTestString(t, "Location 0", resBody[0].Name)
	CheckTestString(t, "Location 1", resBody[1].Name)
	CheckTestString(t, "Location 2", resBody[2].Name)
}

func TestLocationsUpload(t *testing.T) {
	resp, err := http.Get("https://upload.wikimedia.org/wikipedia/commons/7/70/Claybury_Asylum%2C_first_floor_plan._Wellcome_L0023316.jpg")
	if err != nil {
		t.Fatal("Could not load example image")
	}
	CheckTestResponseCode(t, http.StatusOK, resp.StatusCode)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Could not read body from example image")
	}

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

	// Upload
	req = NewHTTPRequest("POST", "/location/"+id+"/map", loginResponse.UserID, bytes.NewBuffer(data))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Get metadata
	req = NewHTTPRequest("GET", "/location/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetLocationResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestString(t, "jpeg", resBody.MapMimeType)
	CheckTestUint(t, 4895, resBody.MapWidth)
	CheckTestUint(t, 3504, resBody.MapHeight)

	// Retrieve
	req = NewHTTPRequest("GET", "/location/"+id+"/map", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBodyMap *GetMapResponse
	json.Unmarshal(res.Body.Bytes(), &resBodyMap)
	data2, err := base64.StdEncoding.DecodeString(resBodyMap.Data)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestUint(t, uint(len(data)), uint(len(data2)))
	CheckTestUint(t, 0, uint(bytes.Compare(data, data2)))
}

func TestLocationsInvalidTimezone(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	// Create with invalid
	payload := `{"name": "Location 1", "timezone": "Europe/Hamburg"}`
	req := NewHTTPRequest("POST", "/location/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)

	// Create with valid
	payload = `{"name": "Location 1", "timezone": "Europe/Berlin"}`
	req = NewHTTPRequest("POST", "/location/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-Id")

	// Update with invalid
	payload = `{"name": "Location 2", "description": "Test 123", "maxConcurrentBookings": 20, "timezone": "Africa/Dubai"}`
	req = NewHTTPRequest("PUT", "/location/"+id, loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)

	// Update with valid
	payload = `{"name": "Location 2", "description": "Test 123", "maxConcurrentBookings": 20, "timezone": "Africa/Cairo"}`
	req = NewHTTPRequest("PUT", "/location/"+id, loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)
}
