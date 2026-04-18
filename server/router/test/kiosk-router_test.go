package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func newKioskRequest(spaceID, secret string) *http.Request {
	req, _ := http.NewRequest("GET", "/kiosk/"+spaceID+"/status", nil)
	if secret != "" {
		req.Header.Set("Authorization", "Bearer "+secret)
	}
	return req
}

func TestKioskNotFound(t *testing.T) {
	ClearTestDB()
	req := newKioskRequest("00000000-0000-0000-0000-000000000000", "anysecret")
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestKioskDisabledForSpace(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	location, space := CreateTestLocationAndSpace(org)
	_ = location

	// KioskEnabled is false by default
	req := newKioskRequest(space.ID, "anysecret")
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestKioskNoCredential(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	location, space := CreateTestLocationAndSpace(org)
	_ = location

	space.KioskEnabled = true
	GetSpaceRepository().Update(space)

	req := newKioskRequest(space.ID, "")
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusUnauthorized, res.Code)
}

func TestKioskNoSecretConfigured(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	location, space := CreateTestLocationAndSpace(org)
	_ = location

	space.KioskEnabled = true
	GetSpaceRepository().Update(space)

	req := newKioskRequest(space.ID, "anysecret")
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusUnauthorized, res.Code)
}

func TestKioskWrongSecret(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	location, space := CreateTestLocationAndSpace(org)
	_ = location

	space.KioskEnabled = true
	GetSpaceRepository().Update(space)

	// Configure a kiosk secret via the settings endpoint
	adminUser := CreateTestUserOrgAdmin(org)
	payload := `{"value": "correct-secret"}`
	req := NewHTTPRequest("PUT", "/setting/kiosk_access_secret", adminUser.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	req = newKioskRequest(space.ID, "wrong-secret")
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusUnauthorized, res.Code)
}

func TestKioskAvailable(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	location, space := CreateTestLocationAndSpace(org)
	_ = location

	space.KioskEnabled = true
	GetSpaceRepository().Update(space)

	adminUser := CreateTestUserOrgAdmin(org)
	payload := `{"value": "myKioskSecret"}`
	req := NewHTTPRequest("PUT", "/setting/kiosk_access_secret", adminUser.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	req = newKioskRequest(space.ID, "myKioskSecret")
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	var resBody KioskResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestString(t, "available", resBody.Status)
	CheckTestString(t, space.ID, resBody.SpaceID)
	CheckTestIsNil(t, resBody.CurrentBooking)
	CheckTestIsNil(t, resBody.NextBooking)
}

func TestKioskOccupied(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	location, space := CreateTestLocationAndSpace(org)
	location.Timezone = "UTC"
	GetLocationRepository().Update(location)

	space.KioskEnabled = true
	GetSpaceRepository().Update(space)

	adminUser := CreateTestUserOrgAdmin(org)
	payload := `{"value": "myKioskSecret"}`
	req := NewHTTPRequest("PUT", "/setting/kiosk_access_secret", adminUser.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Create a current booking
	user := CreateTestUserInOrg(org)
	now := time.Now().UTC()
	booking := &Booking{
		UserID:   user.ID,
		SpaceID:  space.ID,
		Enter:    now.Add(-1 * time.Hour),
		Leave:    now.Add(1 * time.Hour),
		Approved: true,
		Subject:  "Standup",
	}
	GetBookingRepository().Create(booking)

	req = newKioskRequest(space.ID, "myKioskSecret")
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	var resBody KioskResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestString(t, "occupied", resBody.Status)
	if resBody.CurrentBooking == nil {
		t.Fatalf("Expected CurrentBooking to be non-nil")
	}
	CheckTestString(t, "Standup", resBody.CurrentBooking.Subject)
}

func TestKioskNextBooking(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	location, space := CreateTestLocationAndSpace(org)
	location.Timezone = "UTC"
	GetLocationRepository().Update(location)

	space.KioskEnabled = true
	GetSpaceRepository().Update(space)

	adminUser := CreateTestUserOrgAdmin(org)
	payload := `{"value": "myKioskSecret"}`
	req := NewHTTPRequest("PUT", "/setting/kiosk_access_secret", adminUser.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Create a future booking (no current booking)
	user := CreateTestUserInOrg(org)
	now := time.Now().UTC()
	booking := &Booking{
		UserID:   user.ID,
		SpaceID:  space.ID,
		Enter:    now.Add(2 * time.Hour),
		Leave:    now.Add(4 * time.Hour),
		Approved: true,
		Subject:  "Team Sync",
	}
	GetBookingRepository().Create(booking)

	req = newKioskRequest(space.ID, "myKioskSecret")
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	var resBody KioskResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestString(t, "available", resBody.Status)
	CheckTestIsNil(t, resBody.CurrentBooking)
	if resBody.NextBooking == nil {
		t.Fatalf("Expected NextBooking to be non-nil")
	}
	CheckTestString(t, "Team Sync", resBody.NextBooking.Subject)
}

func TestKioskSecretMaskedInSettings(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	adminUser := CreateTestUserOrgAdmin(org)

	// Set the secret
	payload := `{"value": "supersecret"}`
	req := NewHTTPRequest("PUT", "/setting/kiosk_access_secret", adminUser.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Read back – must return "1", not the plaintext or the hash
	req = NewHTTPRequest("GET", "/setting/kiosk_access_secret", adminUser.ID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var value string
	json.Unmarshal(res.Body.Bytes(), &value)
	CheckTestString(t, "1", value)
}
