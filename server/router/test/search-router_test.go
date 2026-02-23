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

func TestSearchForbidden(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	loginResponse := LoginTestUser(user.ID)

	req := NewHTTPRequest("GET", "/search/?query=test", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestSearchUsers(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	u1 := &User{
		Email:          "this.is.max@test.com",
		OrganizationID: org.ID,
	}
	GetUserRepository().Create(u1)
	u2 := &User{
		Email:          "max.it.is@test.com",
		OrganizationID: org.ID,
	}
	GetUserRepository().Create(u2)
	u3 := &User{
		Email:          "other.name@test.com",
		OrganizationID: org.ID,
	}
	GetUserRepository().Create(u3)

	req := NewHTTPRequest("GET", "/search/?query=max&includeUsers=1", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetSearchResultsResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)

	CheckTestInt(t, 2, len(resBody.Users))
	CheckTestString(t, u2.Email, resBody.Users[0].Email)
	CheckTestString(t, u1.Email, resBody.Users[1].Email)
}

func TestSearchLocations(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	l1 := &Location{
		Name:           "Frankfurt 1",
		OrganizationID: org.ID,
	}
	GetLocationRepository().Create(l1)
	l2 := &Location{
		Name:           "Frankfurt 2",
		OrganizationID: org.ID,
	}
	GetLocationRepository().Create(l2)
	l3 := &Location{
		Name:           "Berlin 1",
		OrganizationID: org.ID,
	}
	GetLocationRepository().Create(l3)

	req := NewHTTPRequest("GET", "/search/?query=frank&includeLocations=1", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetSearchResultsResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)

	CheckTestInt(t, 2, len(resBody.Locations))
	CheckTestString(t, l1.Name, resBody.Locations[0].Name)
	CheckTestString(t, l2.Name, resBody.Locations[1].Name)
}

func TestSearchSpaces(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	l1 := &Location{
		Name:           "Frankfurt 1",
		OrganizationID: org.ID,
	}
	GetLocationRepository().Create(l1)

	s1 := &Space{
		Name:       "H123",
		LocationID: l1.ID,
	}
	GetSpaceRepository().Create(s1)
	s2 := &Space{
		Name:       "H234",
		LocationID: l1.ID,
	}
	GetSpaceRepository().Create(s2)
	s3 := &Space{
		Name:       "G123",
		LocationID: l1.ID,
	}
	GetSpaceRepository().Create(s3)

	req := NewHTTPRequest("GET", "/search/?query=123&includeSpaces=1", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetSearchResultsResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)

	CheckTestInt(t, 2, len(resBody.Spaces))
	CheckTestString(t, s3.Name, resBody.Spaces[0].Name)
	CheckTestString(t, s1.Name, resBody.Spaces[1].Name)
}
func TestSearchGroups(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(admin.ID)

	// Enable groups feature
	GetSettingsRepository().Set(org.ID, SettingFeatureGroups.Name, "1")

	// Create groups
	payload := `{"name": "Group Alpha"}`
	req := NewHTTPRequest("POST", "/group/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	payload = `{"name": "Group Beta"}`
	req = NewHTTPRequest("POST", "/group/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	// Search for "Alpha"
	req = NewHTTPRequest("GET", "/search/?query=Alpha&includeGroups=1", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetSearchResultsResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, 1, len(resBody.Groups))
	CheckTestString(t, "Group Alpha", resBody.Groups[0].Name)
}

func TestSearchEmptyQuery(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(admin.ID)

	// Empty query → 200 with empty results (no results match empty string)
	req := NewHTTPRequest("GET", "/search/?query=&includeSpaces=1&includeLocations=1", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetSearchResultsResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if resBody == nil {
		t.Fatal("Expected non-nil response body")
	}
}
