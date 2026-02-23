package test

import (
	"encoding/json"
	"net/http"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestStatsGet(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(admin.ID)

	req := NewHTTPRequest("GET", "/stats/", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetStatsResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if resBody == nil {
		t.Fatal("Expected stats response body")
	}
	// All counts should be >= 0
	if resBody.NumUsers < 0 {
		t.Fatal("Expected numUsers >= 0")
	}
	if resBody.NumBookings < 0 {
		t.Fatal("Expected numBookings >= 0")
	}
	if resBody.NumLocations < 0 {
		t.Fatal("Expected numLocations >= 0")
	}
	if resBody.NumSpaces < 0 {
		t.Fatal("Expected numSpaces >= 0")
	}
}

func TestStatsForbiddenForUser(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	loginResponse := LoginTestUser(user.ID)

	req := NewHTTPRequest("GET", "/stats/", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestStatsWithData(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	user1 := CreateTestUserInOrg(org)
	user2 := CreateTestUserInOrg(org)

	l := &Location{
		Name:                  "Test Location",
		MaxConcurrentBookings: 10,
		OrganizationID:        org.ID,
	}
	GetLocationRepository().Create(l)
	s1 := &Space{Name: "Space 1", LocationID: l.ID}
	GetSpaceRepository().Create(s1)
	s2 := &Space{Name: "Space 2", LocationID: l.ID}
	GetSpaceRepository().Create(s2)

	// Create bookings for users
	CreateTestBooking9To5(user1, s1, 5)
	CreateTestBooking9To5(user2, s2, 5)

	loginResponse := LoginTestUser(admin.ID)
	req := NewHTTPRequest("GET", "/stats/", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetStatsResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if resBody == nil {
		t.Fatal("Expected stats response body")
	}
	// Should have 2 locations
	if resBody.NumLocations < 1 {
		t.Fatalf("Expected at least 1 location, got %d", resBody.NumLocations)
	}
	// Should have 2 spaces
	if resBody.NumSpaces < 2 {
		t.Fatalf("Expected at least 2 spaces, got %d", resBody.NumSpaces)
	}
}
