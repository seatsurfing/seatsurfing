package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestLocationsEmptyResult(t *testing.T) {
	clearTestDB()
	loginResponse := createLoginTestUser()

	req := newHTTPRequest("GET", "/location/", loginResponse.UserID, nil)
	res := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []string
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 0 {
		t.Fatalf("Expected empty array")
	}
}

func TestLocationsForbidden(t *testing.T) {
	clearTestDB()
	org := createTestOrg("test.com")
	user := createTestUserOrgAdmin(org)
	loginResponse := loginTestUser(user.ID)

	// Create
	payload := `{"name": "Location 1"}`
	req := newHTTPRequest("POST", "/location/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-Id")

	org2 := createTestOrg("test2.com")
	user2 := createTestUserOrgAdmin(org2)
	loginResponse2 := loginTestUser(user2.ID)

	// Get from other org
	req = newHTTPRequest("GET", "/location/"+id, loginResponse2.UserID, nil)
	res = executeTestRequest(req)
	checkTestResponseCode(t, http.StatusForbidden, res.Code)

	// Update location from other org
	payload = `{"name": "Location 2"}`
	req = newHTTPRequest("PUT", "/location/"+id, loginResponse2.UserID, bytes.NewBufferString(payload))
	res = executeTestRequest(req)
	checkTestResponseCode(t, http.StatusForbidden, res.Code)

	// Delete location from other org
	req = newHTTPRequest("DELETE", "/location/"+id, loginResponse2.UserID, nil)
	res = executeTestRequest(req)
	checkTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestLocationsCRUD(t *testing.T) {
	clearTestDB()
	org := createTestOrg("test.com")
	user := createTestUserOrgAdmin(org)
	loginResponse := loginTestUser(user.ID)

	// 1. Create
	payload := `{"name": "Location 1"}`
	req := newHTTPRequest("POST", "/location/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-Id")

	// 2. Read
	req = newHTTPRequest("GET", "/location/"+id, loginResponse.UserID, nil)
	res = executeTestRequest(req)
	checkTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetLocationResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	checkTestString(t, "Location 1", resBody.Name)
	checkTestString(t, "", resBody.Description)
	checkTestString(t, "", resBody.Timezone)
	checkTestInt(t, 0, int(resBody.MaxConcurrentBookings))

	// 3. Update
	payload = `{"name": "Location 2", "description": "Test 123", "maxConcurrentBookings": 20, "timezone": "Africa/Cairo"}`
	req = newHTTPRequest("PUT", "/location/"+id, loginResponse.UserID, bytes.NewBufferString(payload))
	res = executeTestRequest(req)
	checkTestResponseCode(t, http.StatusNoContent, res.Code)

	// Read
	req = newHTTPRequest("GET", "/location/"+id, loginResponse.UserID, nil)
	res = executeTestRequest(req)
	checkTestResponseCode(t, http.StatusOK, res.Code)
	var resBody2 *GetLocationResponse
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	checkTestString(t, "Location 2", resBody2.Name)
	checkTestString(t, "Test 123", resBody2.Description)
	checkTestString(t, "Africa/Cairo", resBody2.Timezone)
	checkTestInt(t, 20, int(resBody2.MaxConcurrentBookings))

	// 4. Delete
	req = newHTTPRequest("DELETE", "/location/"+id, loginResponse.UserID, nil)
	res = executeTestRequest(req)
	checkTestResponseCode(t, http.StatusNoContent, res.Code)

	// Read
	req = newHTTPRequest("GET", "/location/"+id, loginResponse.UserID, nil)
	res = executeTestRequest(req)
	checkTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestLocationsList(t *testing.T) {
	clearTestDB()
	org := createTestOrg("test.com")
	user := createTestUserOrgAdmin(org)
	loginResponse := loginTestUser(user.ID)

	payload := `{"name": "Location 1"}`
	req := newHTTPRequest("POST", "/location/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusCreated, res.Code)

	payload = `{"name": "Location 2"}`
	req = newHTTPRequest("POST", "/location/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = executeTestRequest(req)
	checkTestResponseCode(t, http.StatusCreated, res.Code)

	payload = `{"name": "Location 0"}`
	req = newHTTPRequest("POST", "/location/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = executeTestRequest(req)
	checkTestResponseCode(t, http.StatusCreated, res.Code)

	req = newHTTPRequest("GET", "/location/", loginResponse.UserID, nil)
	res = executeTestRequest(req)
	checkTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []*GetLocationResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 3 {
		t.Fatalf("Expected array with 3 elements")
	}
	checkTestString(t, "Location 0", resBody[0].Name)
	checkTestString(t, "Location 1", resBody[1].Name)
	checkTestString(t, "Location 2", resBody[2].Name)
}

func TestLocationsUpload(t *testing.T) {
	resp, err := http.Get("https://upload.wikimedia.org/wikipedia/commons/7/70/Claybury_Asylum%2C_first_floor_plan._Wellcome_L0023316.jpg")
	if err != nil {
		t.Fatal("Could not load example image")
	}
	checkTestResponseCode(t, http.StatusOK, resp.StatusCode)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Could not read body from example image")
	}

	clearTestDB()
	org := createTestOrg("test.com")
	user := createTestUserOrgAdmin(org)
	loginResponse := loginTestUser(user.ID)

	// Create location
	payload := `{"name": "Location 1"}`
	req := newHTTPRequest("POST", "/location/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-Id")

	// Upload
	req = newHTTPRequest("POST", "/location/"+id+"/map", loginResponse.UserID, bytes.NewBuffer(data))
	res = executeTestRequest(req)
	checkTestResponseCode(t, http.StatusNoContent, res.Code)

	// Get metadata
	req = newHTTPRequest("GET", "/location/"+id, loginResponse.UserID, nil)
	res = executeTestRequest(req)
	checkTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetLocationResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	checkTestString(t, "jpeg", resBody.MapMimeType)
	checkTestUint(t, 4895, resBody.MapWidth)
	checkTestUint(t, 3504, resBody.MapHeight)

	// Retrieve
	req = newHTTPRequest("GET", "/location/"+id+"/map", loginResponse.UserID, nil)
	res = executeTestRequest(req)
	checkTestResponseCode(t, http.StatusOK, res.Code)
	var resBodyMap *GetMapResponse
	json.Unmarshal(res.Body.Bytes(), &resBodyMap)
	data2, err := base64.StdEncoding.DecodeString(resBodyMap.Data)
	if err != nil {
		t.Fatal(err)
	}
	checkTestUint(t, uint(len(data)), uint(len(data2)))
	checkTestUint(t, 0, uint(bytes.Compare(data, data2)))
}

func TestLocationsInvalidTimezone(t *testing.T) {
	clearTestDB()
	org := createTestOrg("test.com")
	user := createTestUserOrgAdmin(org)
	loginResponse := loginTestUser(user.ID)

	// Create with invalid
	payload := `{"name": "Location 1", "timezone": "Europe/Hamburg"}`
	req := newHTTPRequest("POST", "/location/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := executeTestRequest(req)
	checkTestResponseCode(t, http.StatusBadRequest, res.Code)

	// Create with valid
	payload = `{"name": "Location 1", "timezone": "Europe/Berlin"}`
	req = newHTTPRequest("POST", "/location/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = executeTestRequest(req)
	checkTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-Id")

	// Update with invalid
	payload = `{"name": "Location 2", "description": "Test 123", "maxConcurrentBookings": 20, "timezone": "Africa/Dubai"}`
	req = newHTTPRequest("PUT", "/location/"+id, loginResponse.UserID, bytes.NewBufferString(payload))
	res = executeTestRequest(req)
	checkTestResponseCode(t, http.StatusBadRequest, res.Code)

	// Update with valid
	payload = `{"name": "Location 2", "description": "Test 123", "maxConcurrentBookings": 20, "timezone": "Africa/Cairo"}`
	req = newHTTPRequest("PUT", "/location/"+id, loginResponse.UserID, bytes.NewBufferString(payload))
	res = executeTestRequest(req)
	checkTestResponseCode(t, http.StatusNoContent, res.Code)
}

func TestLocationsMatchesSearchAttributesSuccess(t *testing.T) {
	router := &LocationRouter{}
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "eq", Value: "value1"},
		{AttributeID: "2", Comparator: "neq", Value: "value2"},
		{AttributeID: "3", Comparator: "contains", Value: "value3"},
		{AttributeID: "4", Comparator: "ncontains", Value: "value4"},
		{AttributeID: "5", Comparator: "lt", Value: "5"},
		{AttributeID: "6", Comparator: "gt", Value: "5"},
		{AttributeID: "7", Comparator: "contains", Value: "foo"},
		{AttributeID: "7", Comparator: "contains", Value: "bar"},
		{AttributeID: "7", Comparator: "contains", Value: "*"},
		{AttributeID: "7", Comparator: "ncontains", Value: "test2"},
		{AttributeID: "8", Comparator: "ncontains", Value: "*"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "value1"},
		{AttributeID: "2", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "value2.2"},
		{AttributeID: "3", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "test-value3-test"},
		{AttributeID: "4", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "test-valuefour-test"},
		{AttributeID: "5", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "4"},
		{AttributeID: "6", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "7"},
		{AttributeID: "7", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: `["foo", "bar", "test"]`},
		{AttributeID: "8", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: `[]`},
	}
	checkTestBool(t, true, router.matchesSearchAttributes("1", &searchAttributes, attributeValues))
}

func TestLocationsMatchesSearchAttributesSuccess2(t *testing.T) {
	router := &LocationRouter{}
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "eq", Value: "value1"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "value1"},
		{AttributeID: "2", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "value2.2"},
		{AttributeID: "3", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "test-value3-test"},
		{AttributeID: "4", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "test-valuefour-test"},
		{AttributeID: "5", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "4"},
		{AttributeID: "6", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "7"},
	}
	checkTestBool(t, true, router.matchesSearchAttributes("1", &searchAttributes, attributeValues))
}

func TestLocationsMatchesSearchAttributesMultipleEntities(t *testing.T) {
	router := &LocationRouter{}
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "eq", Value: "value1"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "2", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "value1111"},
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "value1"},
	}
	checkTestBool(t, true, router.matchesSearchAttributes("1", &searchAttributes, attributeValues))
}

func TestLocationsMatchesSearchAttributesMissingAttribute(t *testing.T) {
	router := &LocationRouter{}
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "eq", Value: "value1"},
		{AttributeID: "2", Comparator: "neq", Value: "value2"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "value1"},
	}
	checkTestBool(t, false, router.matchesSearchAttributes("1", &searchAttributes, attributeValues))
}

func TestLocationsMatchesSearchAttributesEqWrong(t *testing.T) {
	router := &LocationRouter{}
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "eq", Value: "value1"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "value11"},
	}
	checkTestBool(t, false, router.matchesSearchAttributes("1", &searchAttributes, attributeValues))
}

func TestLocationsMatchesSearchAttributesNeqWrong(t *testing.T) {
	router := &LocationRouter{}
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "neq", Value: "value1"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "value1"},
	}
	checkTestBool(t, false, router.matchesSearchAttributes("1", &searchAttributes, attributeValues))
}

func TestLocationsMatchesSearchAttributesContainsWrong(t *testing.T) {
	router := &LocationRouter{}
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "contains", Value: "value1"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "test-value2-test"},
	}
	checkTestBool(t, false, router.matchesSearchAttributes("1", &searchAttributes, attributeValues))
}

func TestLocationsMatchesSearchAttributesNcontainsWrong(t *testing.T) {
	router := &LocationRouter{}
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "ncontains", Value: "value1"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "test-value1-test"},
	}
	checkTestBool(t, false, router.matchesSearchAttributes("1", &searchAttributes, attributeValues))
}

func TestLocationsMatchesSearchAttributesLtWrong(t *testing.T) {
	router := &LocationRouter{}
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "lt", Value: "5"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "5"},
	}
	checkTestBool(t, false, router.matchesSearchAttributes("1", &searchAttributes, attributeValues))
}

func TestLocationsMatchesSearchAttributesGtWrong(t *testing.T) {
	router := &LocationRouter{}
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "gt", Value: "5"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "5"},
	}
	checkTestBool(t, false, router.matchesSearchAttributes("1", &searchAttributes, attributeValues))
}

func TestLocationsMatchesSearchAttributesGteWrong(t *testing.T) {
	router := &LocationRouter{}
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "gte", Value: "5"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "4"},
	}
	checkTestBool(t, false, router.matchesSearchAttributes("1", &searchAttributes, attributeValues))
}

func TestLocationsMatchesSearchAttributesLteWrong(t *testing.T) {
	router := &LocationRouter{}
	searchAttributes := []SearchAttribute{
		{AttributeID: "1", Comparator: "lte", Value: "5"},
	}
	attributeValues := []*SpaceAttributeValue{
		{AttributeID: "1", EntityID: "1", EntityType: SpaceAttributeValueEntityTypeLocation, Value: "6"},
	}
	checkTestBool(t, false, router.matchesSearchAttributes("1", &searchAttributes, attributeValues))
}
