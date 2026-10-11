package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func createTestBookingAttributeViaAPI(t *testing.T, userID, payload string) string {
	req := NewHTTPRequest("POST", "/booking-attribute/", userID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	return res.Header().Get("X-Object-Id")
}

func getTestBookingAttributes(t *testing.T, userID string) []*GetBookingAttributeResponse {
	req := NewHTTPRequest("GET", "/booking-attribute/", userID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []*GetBookingAttributeResponse
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		t.Fatal(err)
	}
	return resBody
}

func TestBookingAttributesCRUD(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)

	id := createTestBookingAttributeViaAPI(t, admin.ID, `{"label": "Company", "type": 1, "required": true}`)
	createTestBookingAttributeViaAPI(t, admin.ID, `{"label": "Notes", "type": 2, "required": false}`)

	list := getTestBookingAttributes(t, admin.ID)
	CheckTestInt(t, 2, len(list))
	CheckTestString(t, id, list[0].ID)
	CheckTestString(t, "Company", list[0].Label)
	CheckTestInt(t, 1, list[0].Type)
	CheckTestBool(t, true, list[0].Required)
	CheckTestString(t, "Notes", list[1].Label)

	req := NewHTTPRequest("PUT", "/booking-attribute/"+id, admin.ID, bytes.NewBufferString(`{"label": "Persons", "type": 3, "required": false}`))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)
	list = getTestBookingAttributes(t, admin.ID)
	CheckTestString(t, "Persons", list[0].Label)
	CheckTestInt(t, 3, list[0].Type)
	CheckTestBool(t, false, list[0].Required)

	req = NewHTTPRequest("DELETE", "/booking-attribute/"+id, admin.ID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)
	list = getTestBookingAttributes(t, admin.ID)
	CheckTestInt(t, 1, len(list))
}

func TestBookingAttributesMove(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	user := CreateTestUserInOrg(org)
	id1 := createTestBookingAttributeViaAPI(t, admin.ID, `{"label": "Company", "type": 1}`)
	id2 := createTestBookingAttributeViaAPI(t, admin.ID, `{"label": "Notes", "type": 2}`)

	req := NewHTTPRequest("PUT", "/booking-attribute/"+id2+"/move", admin.ID, bytes.NewBufferString(`{"direction": "up"}`))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)
	list := getTestBookingAttributes(t, admin.ID)
	CheckTestString(t, id2, list[0].ID)
	CheckTestString(t, id1, list[1].ID)

	req = NewHTTPRequest("PUT", "/booking-attribute/"+id2+"/move", admin.ID, bytes.NewBufferString(`{"direction": "down"}`))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)
	list = getTestBookingAttributes(t, admin.ID)
	CheckTestString(t, id1, list[0].ID)
	CheckTestString(t, id2, list[1].ID)

	for _, payload := range []string{`{}`, `{"direction": "left"}`} {
		req = NewHTTPRequest("PUT", "/booking-attribute/"+id1+"/move", admin.ID, bytes.NewBufferString(payload))
		res = ExecuteTestRequest(req)
		CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
	}

	req = NewHTTPRequest("PUT", "/booking-attribute/"+id2+"/move", user.ID, bytes.NewBufferString(`{"direction": "up"}`))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
	list = getTestBookingAttributes(t, admin.ID)
	CheckTestString(t, id1, list[0].ID)
}

func TestBookingAttributesInvalidPayload(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)

	for _, payload := range []string{
		`{"label": "", "type": 1}`,
		`{"label": "Company", "type": 5}`,
		`{"label": "Company"}`,
		`{"label": "  ", "type": 1}`,
	} {
		req := NewHTTPRequest("POST", "/booking-attribute/", admin.ID, bytes.NewBufferString(payload))
		res := ExecuteTestRequest(req)
		CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
	}
}

func TestBookingAttributesLimit(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	for i := 0; i < 3; i++ {
		createTestBookingAttributeViaAPI(t, admin.ID, `{"label": "Attr", "type": 1}`)
	}
	req := NewHTTPRequest("POST", "/booking-attribute/", admin.ID, bytes.NewBufferString(`{"label": "Attr", "type": 1}`))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
	CheckTestString(t, strconv.Itoa(ResponseCodeBookingAttributeLimitReached), res.Header().Get("X-Error-Code"))
}

func TestBookingAttributesForbiddenForRegularUser(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	user := CreateTestUserInOrg(org)
	id := createTestBookingAttributeViaAPI(t, admin.ID, `{"label": "Company", "type": 1}`)

	for _, r := range []struct{ method, url, payload string }{
		{"GET", "/booking-attribute/", ""},
		{"POST", "/booking-attribute/", `{"label": "X", "type": 1}`},
		{"PUT", "/booking-attribute/" + id, `{"label": "X", "type": 1}`},
		{"DELETE", "/booking-attribute/" + id, ""},
	} {
		req := NewHTTPRequest(r.method, r.url, user.ID, bytes.NewBufferString(r.payload))
		res := ExecuteTestRequest(req)
		CheckTestResponseCode(t, http.StatusForbidden, res.Code)
	}
}

func TestBookingAttributesForeignOrg(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	org2 := CreateTestOrg("test2.com")
	admin := CreateTestUserOrgAdmin(org)
	admin2 := CreateTestUserOrgAdmin(org2)
	id := createTestBookingAttributeViaAPI(t, admin.ID, `{"label": "Company", "type": 1}`)

	req := NewHTTPRequest("PUT", "/booking-attribute/"+id, admin2.ID, bytes.NewBufferString(`{"label": "X", "type": 1}`))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
	req = NewHTTPRequest("PUT", "/booking-attribute/"+id+"/move", admin2.ID, bytes.NewBufferString(`{"direction": "down"}`))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
	req = NewHTTPRequest("DELETE", "/booking-attribute/"+id, admin2.ID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
	CheckTestInt(t, 0, len(getTestBookingAttributes(t, admin2.ID)))

	list := getTestBookingAttributes(t, admin.ID)
	CheckTestInt(t, 1, len(list))
	CheckTestString(t, "Company", list[0].Label)
}

func TestPublicBookingGetSpacesReturnsAttributes(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	_, space := CreateTestLocationAndSpace(org)
	enablePublicBookingForOrgAndSpace(org, space)
	id := createTestBookingAttributeViaAPI(t, admin.ID, `{"label": "Company", "type": 1, "required": true}`)

	req := NewHTTPRequest("GET", "/public-booking/"+org.ID+"/spaces", "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody GetPublicBookableSpacesResponse
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(resBody.Attributes))
	CheckTestString(t, id, resBody.Attributes[0].ID)
	CheckTestString(t, "Company", resBody.Attributes[0].Label)
	CheckTestBool(t, true, resBody.Attributes[0].Required)
}

func publicBookingRequestPayload(spaceID, attributes string) string {
	return `{"spaceId": "` + spaceID + `", "enter": "2030-01-02T09:00:00Z", "leave": "2030-01-02T17:00:00Z", "name": "Jane Doe", "email": "jane.doe@test.com", "attributes": ` + attributes + `}`
}

func TestPublicBookingRequestInvalidAttributes(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org.ID, SettingMaxDaysInAdvance.Name, "5000")
	admin := CreateTestUserOrgAdmin(org)
	_, space := CreateTestLocationAndSpace(org)
	enablePublicBookingForOrgAndSpace(org, space)
	company := createTestBookingAttributeViaAPI(t, admin.ID, `{"label": "Company", "type": 1, "required": true}`)
	persons := createTestBookingAttributeViaAPI(t, admin.ID, `{"label": "Persons", "type": 3}`)
	terms := createTestBookingAttributeViaAPI(t, admin.ID, `{"label": "Terms", "type": 4, "required": true}`)

	for _, attributes := range []string{
		`{}`,
		`{"` + company + `": "", "` + terms + `": "1"}`,
		`{"` + company + `": "ACME", "` + terms + `": "1", "` + persons + `": "many"}`,
		`{"` + company + `": "ACME", "` + terms + `": "1", "` + persons + `": "-1"}`,
		`{"` + company + `": "ACME", "` + terms + `": "0"}`,
		`{"` + company + `": "ACME"}`,
		`{"` + company + `": "ACME", "00000000-0000-0000-0000-000000000000": "x"}`,
	} {
		req := NewHTTPRequest("POST", "/public-booking/"+org.ID+"/request", "", bytes.NewBufferString(publicBookingRequestPayload(space.ID, attributes)))
		res := ExecuteTestRequest(req)
		CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
		CheckTestString(t, strconv.Itoa(ResponseCodeBookingInvalidAttributes), res.Header().Get("X-Error-Code"))
	}
	pending, _ := GetAuthStateRepository().GetActiveByKeyAndType("jane.doe@test.com", AuthPublicBooking)
	CheckTestInt(t, 0, len(pending))
}

func TestPublicBookingAttributesFlow(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org.ID, SettingMaxDaysInAdvance.Name, "5000")
	GetSettingsRepository().Set(org.ID, SettingFeatureGroups.Name, "1")
	admin := CreateTestUserOrgAdmin(org)
	_, space := CreateTestLocationAndSpace(org)
	enablePublicBookingForOrgAndSpace(org, space)
	group := CreateTestGroup(org, admin)
	if err := GetSpaceRepository().AddApprovers(space, []string{group.ID}); err != nil {
		t.Fatal(err)
	}
	company := createTestBookingAttributeViaAPI(t, admin.ID, `{"label": "Company", "type": 1, "required": true}`)
	notes := createTestBookingAttributeViaAPI(t, admin.ID, `{"label": "Notes", "type": 2}`)
	createTestBookingAttributeViaAPI(t, admin.ID, `{"label": "Persons", "type": 3}`)

	attributes := `{"` + company + `": " ACME ", "` + notes + `": "Line 1\nLine 2"}`
	req := NewHTTPRequest("POST", "/public-booking/"+org.ID+"/request", "", bytes.NewBufferString(publicBookingRequestPayload(space.ID, attributes)))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	pending, err := GetAuthStateRepository().GetActiveByKeyAndType("jane.doe@test.com", AuthPublicBooking)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(pending))
	var payload PublicBookingRequestPayload
	if err := json.Unmarshal([]byte(pending[0].Payload), &payload); err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 2, len(payload.Attributes))
	CheckTestString(t, "ACME", payload.Attributes[company])

	CheckTestString(t, "pending", confirmTestPublicBooking(t, pending[0].ID))

	req = NewHTTPRequest("GET", "/booking/pendingapprovals/", admin.ID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var approvals []*GetBookingResponse
	if err := json.Unmarshal(res.Body.Bytes(), &approvals); err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(approvals))
	CheckTestInt(t, 2, len(approvals[0].Attributes))
	CheckTestString(t, "Company", approvals[0].Attributes[0].Label)
	CheckTestString(t, "ACME", approvals[0].Attributes[0].Value)
	CheckTestString(t, "Notes", approvals[0].Attributes[1].Label)
	CheckTestInt(t, 2, approvals[0].Attributes[1].Type)
	CheckTestString(t, "Line 1\nLine 2", approvals[0].Attributes[1].Value)

	req = NewHTTPRequest("GET", "/booking/"+approvals[0].ID, admin.ID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var booking GetBookingResponse
	if err := json.Unmarshal(res.Body.Bytes(), &booking); err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 2, len(booking.Attributes))
	CheckTestString(t, company, booking.Attributes[0].AttributeID)

	// Drain the async approval request notification mail.
	waitForSendMailMockContent(t, 2*time.Second)
}

func TestBookingWithoutAttributesOmitsAttributes(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	_, space := CreateTestLocationAndSpace(org)
	b := CreateTestBooking9To5(admin, space, 1)

	req := NewHTTPRequest("GET", "/booking/"+b.ID, admin.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		t.Fatal(err)
	}
	if _, ok := resBody["attributes"]; ok {
		t.Fatal("expected no attributes")
	}
}
