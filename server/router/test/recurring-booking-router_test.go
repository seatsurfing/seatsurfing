package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestRecurringBookingsPrecheckFeatureDisabled(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org.ID, SettingFeatureRecurringBookings.Name, "0")
	GetSettingsRepository().Set(org.ID, SettingMaxDaysInAdvance.Name, strconv.Itoa(365*10))
	GetSettingsRepository().Set(org.ID, SettingMaxBookingsPerUser.Name, "1000")
	user1 := CreateTestUserInOrg(org)

	l := &Location{
		Name:                  "Test",
		MaxConcurrentBookings: 2,
		OrganizationID:        org.ID,
	}
	GetLocationRepository().Create(l)
	s1 := &Space{Name: "Test 1", LocationID: l.ID}
	GetSpaceRepository().Create(s1)

	payload := `{
	"spaceId": "` + s1.ID + `",
	"subject": "Test",
	"enter": "2030-08-28T09:00:00+02:00",
	"leave": "2030-08-28T15:00:00+02:00",
	"end": "2030-09-03T00:00:00+02:00",
	"cadence": 1,
	"cycle": 1
	}`
	req := NewHTTPRequest("POST", "/recurring-booking/precheck", user1.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusPaymentRequired, res.Code)
}

func TestRecurringBookingsCreateFeatureDisabled(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org.ID, SettingFeatureRecurringBookings.Name, "0")
	GetSettingsRepository().Set(org.ID, SettingMaxDaysInAdvance.Name, strconv.Itoa(365*10))
	GetSettingsRepository().Set(org.ID, SettingMaxBookingsPerUser.Name, "1000")
	user1 := CreateTestUserInOrg(org)

	l := &Location{
		Name:                  "Test",
		MaxConcurrentBookings: 2,
		OrganizationID:        org.ID,
	}
	GetLocationRepository().Create(l)
	s1 := &Space{Name: "Test 1", LocationID: l.ID}
	GetSpaceRepository().Create(s1)

	payload := `{
	"spaceId": "` + s1.ID + `",
	"subject": "Test",
	"enter": "2030-08-28T09:00:00+02:00",
	"leave": "2030-08-28T15:00:00+02:00",
	"end": "2030-09-03T00:00:00+02:00",
	"cadence": 1,
	"cycle": 1
	}`
	req := NewHTTPRequest("POST", "/recurring-booking/", user1.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusPaymentRequired, res.Code)
}

func TestRecurringBookingsPrecheck(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org.ID, SettingFeatureRecurringBookings.Name, "1")
	GetSettingsRepository().Set(org.ID, SettingMaxDaysInAdvance.Name, strconv.Itoa(365*10))
	GetSettingsRepository().Set(org.ID, SettingMaxBookingsPerUser.Name, "1000")
	user1 := CreateTestUserInOrg(org)
	user2 := CreateTestUserInOrg(org)
	user3 := CreateTestUserInOrg(org)

	l := &Location{
		Name:                  "Test",
		MaxConcurrentBookings: 2,
		OrganizationID:        org.ID,
	}
	GetLocationRepository().Create(l)
	s1 := &Space{Name: "Test 1", LocationID: l.ID}
	GetSpaceRepository().Create(s1)
	s2 := &Space{Name: "Test 2", LocationID: l.ID}
	GetSpaceRepository().Create(s2)

	// Create booking 1
	payload := "{\"spaceId\": \"" + s1.ID + "\", \"enter\": \"2030-09-01T08:30:00+02:00\", \"leave\": \"2030-09-01T17:00:00+02:00\"}"
	req := NewHTTPRequest("POST", "/booking/", user1.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	// Create booking 2
	payload = "{\"spaceId\": \"" + s2.ID + "\", \"enter\": \"2030-09-02T07:30:00+02:00\", \"leave\": \"2030-09-02T12:00:00+02:00\"}"
	req = NewHTTPRequest("POST", "/booking/", user2.ID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	payload = `{
	"spaceId": "` + s1.ID + `",
	"subject": "Test",
	"enter": "2030-08-28T09:00:00+02:00",
	"leave": "2030-08-28T15:00:00+02:00",
	"end": "2030-09-03T00:00:00+02:00",
	"cadence": 1,
	"cycle": 1
	}`
	req = NewHTTPRequest("POST", "/recurring-booking/precheck", user3.ID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []CreateRecurringBookingResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, 6, len(resBody))

	CheckTestBool(t, true, resBody[0].Success)  // 28
	CheckTestBool(t, true, resBody[1].Success)  // 29
	CheckTestBool(t, true, resBody[2].Success)  // 30
	CheckTestBool(t, true, resBody[3].Success)  // 31
	CheckTestBool(t, false, resBody[4].Success) // 01
	CheckTestBool(t, true, resBody[5].Success)  // 02
}

func TestRecurringBookingsCreateDelete(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org.ID, SettingFeatureRecurringBookings.Name, "1")
	GetSettingsRepository().Set(org.ID, SettingMaxDaysInAdvance.Name, strconv.Itoa(365*10))
	GetSettingsRepository().Set(org.ID, SettingMaxBookingsPerUser.Name, "1000")
	user1 := CreateTestUserInOrg(org)
	user2 := CreateTestUserInOrg(org)
	user3 := CreateTestUserInOrg(org)

	l := &Location{
		Name:                  "Test",
		MaxConcurrentBookings: 2,
		OrganizationID:        org.ID,
	}
	GetLocationRepository().Create(l)
	s1 := &Space{Name: "Test 1", LocationID: l.ID}
	GetSpaceRepository().Create(s1)
	s2 := &Space{Name: "Test 2", LocationID: l.ID}
	GetSpaceRepository().Create(s2)

	// Create booking 1
	payload := "{\"spaceId\": \"" + s1.ID + "\", \"enter\": \"2030-09-01T08:30:00+02:00\", \"leave\": \"2030-09-01T17:00:00+02:00\"}"
	req := NewHTTPRequest("POST", "/booking/", user1.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	// Create booking 2
	payload = "{\"spaceId\": \"" + s2.ID + "\", \"enter\": \"2030-09-02T07:30:00+02:00\", \"leave\": \"2030-09-02T12:00:00+02:00\"}"
	req = NewHTTPRequest("POST", "/booking/", user2.ID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	payload = `{
	"spaceId": "` + s1.ID + `",
	"subject": "Test",
	"enter": "2030-08-28T09:00:00+02:00",
	"leave": "2030-08-28T15:00:00+02:00",
	"end": "2030-09-03T00:00:00+02:00",
	"cadence": 1,
	"cycle": 1
	}`
	req = NewHTTPRequest("POST", "/recurring-booking/", user3.ID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	var resBody []CreateRecurringBookingResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, 6, len(resBody))

	CheckTestBool(t, true, resBody[0].Success)  // 28
	CheckTestBool(t, true, resBody[1].Success)  // 29
	CheckTestBool(t, true, resBody[2].Success)  // 30
	CheckTestBool(t, true, resBody[3].Success)  // 31
	CheckTestBool(t, false, resBody[4].Success) // 01
	CheckTestBool(t, true, resBody[5].Success)  // 02

	for idx, b := range resBody {
		if idx != 4 { // 01 is not created
			CheckTestBool(t, true, b.ID != "")
			booking, err := GetBookingRepository().GetOne(b.ID)
			CheckTestBool(t, true, err == nil)
			CheckTestBool(t, true, booking != nil)
		}
	}

	booking, _ := GetBookingRepository().GetOne(resBody[0].ID)
	req = NewHTTPRequest("DELETE", "/recurring-booking/"+string(booking.RecurringID), user3.ID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)
}

func TestRecurringBookingsGet(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org.ID, SettingFeatureRecurringBookings.Name, "1")
	GetSettingsRepository().Set(org.ID, SettingMaxDaysInAdvance.Name, strconv.Itoa(365*10))
	GetSettingsRepository().Set(org.ID, SettingMaxBookingsPerUser.Name, "1000")
	user1 := CreateTestUserInOrg(org)

	l := &Location{
		Name:                  "Test",
		MaxConcurrentBookings: 2,
		OrganizationID:        org.ID,
	}
	GetLocationRepository().Create(l)
	s1 := &Space{Name: "Test 1", LocationID: l.ID}
	GetSpaceRepository().Create(s1)

	payload := `{
	"spaceId": "` + s1.ID + `",
	"subject": "Test",
	"enter": "2030-08-28T09:00:00+02:00",
	"leave": "2030-08-28T15:00:00+02:00",
	"end": "2030-08-30T00:00:00+02:00",
	"cadence": 1,
	"cycle": 1
	}`
	req := NewHTTPRequest("POST", "/recurring-booking/", user1.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-ID")

	// Get the recurring booking by ID
	req = NewHTTPRequest("GET", "/recurring-booking/"+id, user1.ID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetRecurringBookingResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if resBody == nil {
		t.Fatal("Expected non-nil recurring booking response")
	}
	CheckTestString(t, s1.ID, resBody.SpaceID)
}

func TestRecurringBookingsGetNotFound(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user1 := CreateTestUserInOrg(org)

	req := NewHTTPRequest("GET", "/recurring-booking/nonexistent-id", user1.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestRecurringBookingsGetForeign(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org.ID, SettingFeatureRecurringBookings.Name, "1")
	GetSettingsRepository().Set(org.ID, SettingMaxDaysInAdvance.Name, strconv.Itoa(365*10))
	GetSettingsRepository().Set(org.ID, SettingMaxBookingsPerUser.Name, "1000")
	user1 := CreateTestUserInOrg(org)
	user2 := CreateTestUserInOrg(org)

	l := &Location{
		Name:                  "Test",
		MaxConcurrentBookings: 2,
		OrganizationID:        org.ID,
	}
	GetLocationRepository().Create(l)
	s1 := &Space{Name: "Test 1", LocationID: l.ID}
	GetSpaceRepository().Create(s1)

	// user1 creates a recurring booking
	payload := `{
	"spaceId": "` + s1.ID + `",
	"subject": "Test",
	"enter": "2030-08-28T09:00:00+02:00",
	"leave": "2030-08-28T15:00:00+02:00",
	"end": "2030-08-30T00:00:00+02:00",
	"cadence": 1,
	"cycle": 1
	}`
	req := NewHTTPRequest("POST", "/recurring-booking/", user1.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-ID")

	// user2 tries to get user1's recurring booking → should be forbidden
	req = NewHTTPRequest("GET", "/recurring-booking/"+id, user2.ID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestRecurringBookingsGetICal(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org.ID, SettingFeatureRecurringBookings.Name, "1")
	GetSettingsRepository().Set(org.ID, SettingMaxDaysInAdvance.Name, strconv.Itoa(365*10))
	GetSettingsRepository().Set(org.ID, SettingMaxBookingsPerUser.Name, "1000")
	user1 := CreateTestUserInOrg(org)

	l := &Location{
		Name:                  "Test",
		MaxConcurrentBookings: 2,
		OrganizationID:        org.ID,
	}
	GetLocationRepository().Create(l)
	s1 := &Space{Name: "Test 1", LocationID: l.ID}
	GetSpaceRepository().Create(s1)

	payload := `{
	"spaceId": "` + s1.ID + `",
	"subject": "Test",
	"enter": "2030-08-28T09:00:00+02:00",
	"leave": "2030-08-28T15:00:00+02:00",
	"end": "2030-08-30T00:00:00+02:00",
	"cadence": 1,
	"cycle": 1
	}`
	req := NewHTTPRequest("POST", "/recurring-booking/", user1.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-ID")

	// Get iCal for the recurring booking
	req = NewHTTPRequest("GET", "/recurring-booking/"+id+"/ical", user1.ID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	contentType := res.Header().Get("Content-Type")
	if contentType != "text/calendar" {
		t.Fatalf("Expected Content-Type text/calendar, got %s", contentType)
	}
}

func TestRecurringBookingsGetICalNotFound(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user1 := CreateTestUserInOrg(org)

	req := NewHTTPRequest("GET", "/recurring-booking/nonexistent-id/ical", user1.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestRecurringBookingsDeleteNotFound(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user1 := CreateTestUserInOrg(org)

	req := NewHTTPRequest("DELETE", "/recurring-booking/nonexistent-id", user1.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestRecurringBookingsDeleteForeign(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org.ID, SettingFeatureRecurringBookings.Name, "1")
	GetSettingsRepository().Set(org.ID, SettingMaxDaysInAdvance.Name, strconv.Itoa(365*10))
	GetSettingsRepository().Set(org.ID, SettingMaxBookingsPerUser.Name, "1000")
	user1 := CreateTestUserInOrg(org)
	user2 := CreateTestUserInOrg(org)

	l := &Location{
		Name:                  "Test",
		MaxConcurrentBookings: 2,
		OrganizationID:        org.ID,
	}
	GetLocationRepository().Create(l)
	s1 := &Space{Name: "Test 1", LocationID: l.ID}
	GetSpaceRepository().Create(s1)

	// user1 creates a recurring booking
	payload := `{
	"spaceId": "` + s1.ID + `",
	"subject": "Test",
	"enter": "2030-08-28T09:00:00+02:00",
	"leave": "2030-08-28T15:00:00+02:00",
	"end": "2030-08-30T00:00:00+02:00",
	"cadence": 1,
	"cycle": 1
	}`
	req := NewHTTPRequest("POST", "/recurring-booking/", user1.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-ID")

	// user2 tries to delete user1's recurring booking → should be forbidden
	req = NewHTTPRequest("DELETE", "/recurring-booking/"+id, user2.ID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}
