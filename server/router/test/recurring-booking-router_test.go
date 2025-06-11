package test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestRecurringBookingsPrecheck(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
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

	log.Println(resBody)

	CheckTestBool(t, true, resBody[0].Success)  // 28
	CheckTestBool(t, true, resBody[1].Success)  // 29
	CheckTestBool(t, true, resBody[2].Success)  // 30
	CheckTestBool(t, true, resBody[3].Success)  // 31
	CheckTestBool(t, false, resBody[4].Success) // 01
	CheckTestBool(t, true, resBody[5].Success)  // 02
}
