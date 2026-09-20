package test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func createTestPublicBookingAuthState(t *testing.T, space *Space, enter, leave time.Time) string {
	payload := PublicBookingRequestPayload{
		SpaceID:  space.ID,
		Enter:    enter,
		Leave:    leave,
		Name:     "Jane Doe",
		Email:    "jane.doe@test.com",
		Subject:  "Test",
		Language: "en",
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	authState := &AuthState{
		Expiry:        time.Now().Add(30 * time.Minute),
		AuthStateType: AuthPublicBooking,
		Payload:       string(payloadJSON),
		Key:           "jane.doe@test.com",
	}
	if err := GetAuthStateRepository().Create(authState); err != nil {
		t.Fatal(err)
	}
	return authState.ID
}

func enablePublicBookingForOrgAndSpace(org *Organization, space *Space) *Space {
	GetSettingsRepository().Set(org.ID, SettingFeaturePublicBooking.Name, "1")
	GetSettingsRepository().Set(org.ID, SettingPublicBookingEnabled.Name, "1")
	space.PublicBookingEnabled = true
	if err := GetSpaceRepository().Update(space); err != nil {
		panic(err)
	}
	return space
}

func TestPublicBookingConfirmPendingReturnsBookingTimes(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	_, space := CreateTestLocationAndSpace(org)
	enablePublicBookingForOrgAndSpace(org, space)

	enter := time.Date(2030, 1, 2, 9, 0, 0, 0, time.UTC)
	leave := time.Date(2030, 1, 2, 17, 0, 0, 0, time.UTC)
	id := createTestPublicBookingAuthState(t, space, enter, leave)

	req := NewHTTPRequest("POST", "/public-booking/confirm/"+id, "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	var resBody ConfirmPublicBookingResponse
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		t.Fatal(err)
	}
	CheckTestString(t, "pending", resBody.Status)
	if !resBody.Enter.Equal(enter) {
		t.Fatalf("expected enter %v, got %v", enter, resBody.Enter)
	}
	if !resBody.Leave.Equal(leave) {
		t.Fatalf("expected leave %v, got %v", leave, resBody.Leave)
	}
}

func TestPublicBookingConfirmUnavailableReturnsBookingTimes(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	_, space := CreateTestLocationAndSpace(org)
	enablePublicBookingForOrgAndSpace(org, space)

	enter := time.Date(2030, 1, 2, 9, 0, 0, 0, time.UTC)
	leave := time.Date(2030, 1, 2, 17, 0, 0, 0, time.UTC)

	// Occupy the slot with a conflicting booking first, so confirm() finds it unavailable.
	conflictingBooking := &Booking{
		SpaceID: space.ID,
		Enter:   enter,
		Leave:   leave,
	}
	if err := GetBookingRepository().Create(conflictingBooking); err != nil {
		t.Fatal(err)
	}

	id := createTestPublicBookingAuthState(t, space, enter, leave)

	req := NewHTTPRequest("POST", "/public-booking/confirm/"+id, "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	var resBody ConfirmPublicBookingResponse
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		t.Fatal(err)
	}
	CheckTestString(t, "unavailable", resBody.Status)
	if !resBody.Enter.Equal(enter) {
		t.Fatalf("expected enter %v, got %v", enter, resBody.Enter)
	}
	if !resBody.Leave.Equal(leave) {
		t.Fatalf("expected leave %v, got %v", leave, resBody.Leave)
	}
}

func TestPublicBookingConfirmInvalidID(t *testing.T) {
	ClearTestDB()

	req := NewHTTPRequest("POST", "/public-booking/confirm/does-not-exist", "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}
