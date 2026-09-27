package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

func createTestPublicBookingAuthState(t *testing.T, space *Space, enter, leave time.Time) string {
	return createTestPublicBookingAuthStateWithLanguage(t, space, enter, leave, "en")
}

func createTestPublicBookingAuthStateWithLanguage(t *testing.T, space *Space, enter, leave time.Time, language string) string {
	payload := PublicBookingRequestPayload{
		SpaceID:  space.ID,
		Enter:    enter,
		Leave:    leave,
		Name:     "Jane Doe",
		Email:    "jane.doe@test.com",
		Subject:  "Test",
		Language: language,
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

func TestPublicBookingRequestInvalidEmail(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	_, space := CreateTestLocationAndSpace(org)
	space = enablePublicBookingForOrgAndSpace(org, space)

	for _, email := range []string{"test@test", "test@test.c", "ä@test.com"} {
		payload := `{"spaceId": "` + space.ID + `", "enter": "2030-09-01T08:30:00Z", "leave": "2030-09-01T17:00:00Z", "name": "Jane Doe", "email": "` + email + `"}`
		req := NewHTTPRequest("POST", "/public-booking/"+org.ID+"/request", "", bytes.NewBufferString(payload))
		res := ExecuteTestRequest(req)
		CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
	}
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

// TestPublicBookingConfirmConcurrentOnlyCreatesOneBooking guards against the
// double opt-in confirmation link being posted twice at once (e.g. a
// double-mounted client tab) and creating two bookings from a single one-time
// link. The handler must consume the auth state atomically so only one of the
// concurrent requests succeeds.
func TestPublicBookingConfirmConcurrentOnlyCreatesOneBooking(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	_, space := CreateTestLocationAndSpace(org)
	enablePublicBookingForOrgAndSpace(org, space)

	enter := time.Date(2030, 1, 2, 9, 0, 0, 0, time.UTC)
	leave := time.Date(2030, 1, 2, 17, 0, 0, 0, time.UTC)
	id := createTestPublicBookingAuthState(t, space, enter, leave)

	numAttempts := 10
	var pendingCount int32
	var wg sync.WaitGroup
	wg.Add(numAttempts)
	for i := 0; i < numAttempts; i++ {
		go func() {
			defer wg.Done()
			req := NewHTTPRequest("POST", "/public-booking/confirm/"+id, "", nil)
			res := ExecuteTestRequest(req)
			if res.Code != http.StatusOK {
				return
			}
			var resBody ConfirmPublicBookingResponse
			if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
				return
			}
			if resBody.Status == "pending" {
				atomic.AddInt32(&pendingCount, 1)
			}
		}()
	}
	wg.Wait()

	CheckTestInt(t, 1, int(pendingCount))

	conflicts, err := GetBookingRepository().GetConflicts(space.ID, enter, leave, "")
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(conflicts))
}

// setUpPublicBookingRequiringApproval confirms a public booking request for
// a space whose sole approver is adminUser, so the approve endpoint can
// subsequently approve/decline it and trigger the (async) notification mail.
func setUpPublicBookingRequiringApproval(t *testing.T, org *Organization, adminUser *User) *BookingDetails {
	return setUpPublicBookingRequiringApprovalWithLanguage(t, org, adminUser, "en")
}

func setUpPublicBookingRequiringApprovalWithLanguage(t *testing.T, org *Organization, adminUser *User, language string) *BookingDetails {
	GetSettingsRepository().Set(org.ID, SettingFeatureGroups.Name, "1")
	_, space := CreateTestLocationAndSpace(org)
	enablePublicBookingForOrgAndSpace(org, space)

	group := &Group{Name: "Approvers", OrganizationID: org.ID}
	if err := GetGroupRepository().Create(group); err != nil {
		t.Fatal(err)
	}
	if err := GetGroupRepository().AddMembers(group, []string{adminUser.ID}); err != nil {
		t.Fatal(err)
	}
	if err := GetSpaceRepository().AddApprovers(space, []string{group.ID}); err != nil {
		t.Fatal(err)
	}

	enter := time.Date(2030, 1, 2, 9, 0, 0, 0, time.UTC)
	leave := time.Date(2030, 1, 2, 17, 0, 0, 0, time.UTC)
	id := createTestPublicBookingAuthStateWithLanguage(t, space, enter, leave, language)

	req := NewHTTPRequest("POST", "/public-booking/confirm/"+id, "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	pending, err := GetBookingRepository().GetBookingsRequiringApproval(adminUser.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending booking, got %d", len(pending))
	}
	return pending[0]
}

// waitForSendMailMockContent polls SendMailMockContent, which is filled
// asynchronously by the approve/decline notification goroutine.
func waitForSendMailMockContent(t *testing.T, timeout time.Duration) string {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if SendMailMockContent != "" {
			return SendMailMockContent
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timed out waiting for notification mail")
	return ""
}

// waitForSendMailMockContentContaining is like waitForSendMailMockContent,
// but for tests that trigger more than one notification mail in sequence
// (e.g. approve, then delete): it waits for the specific mail the test cares
// about instead of just "any" non-empty content, so it can't be fooled by
// an earlier notification's content still sitting in the mock var.
func waitForSendMailMockContentContaining(t *testing.T, substr string, timeout time.Duration) string {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if strings.Contains(SendMailMockContent, substr) {
			return SendMailMockContent
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timed out waiting for notification mail containing " + substr)
	return ""
}

func TestPublicBookingApprovedMailOmitsYourBookingsButton(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	adminUser := CreateTestUserOrgAdmin(org)
	pending := setUpPublicBookingRequiringApproval(t, org, adminUser)

	SendMailMockContent = ""
	payload := `{"approved": true}`
	req := NewHTTPRequest("POST", "/booking/"+pending.ID+"/approve", adminUser.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	mail := waitForSendMailMockContent(t, 2*time.Second)
	CheckTestBool(t, true, strings.Contains(mail, "approved"))
	CheckTestBool(t, false, strings.Contains(mail, "Your bookings"))
	CheckTestBool(t, false, strings.Contains(mail, "ui/bookings/"))
}

func TestPublicBookingDeclinedMailHasNewBookingLinkNotYourBookings(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	adminUser := CreateTestUserOrgAdmin(org)
	pending := setUpPublicBookingRequiringApproval(t, org, adminUser)

	SendMailMockContent = ""
	payload := `{"approved": false}`
	req := NewHTTPRequest("POST", "/booking/"+pending.ID+"/approve", adminUser.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	mail := waitForSendMailMockContent(t, 2*time.Second)
	CheckTestBool(t, true, strings.Contains(mail, "declined"))
	CheckTestBool(t, true, strings.Contains(mail, "New booking"))
	CheckTestBool(t, true, strings.Contains(mail, "ui/book/"))
	CheckTestBool(t, false, strings.Contains(mail, "Your bookings"))
	CheckTestBool(t, false, strings.Contains(mail, "ui/bookings/"))
}

// TestPublicBookingApprovedMailContainsDetailsLink asserts the approved-mail
// now links to the public details page keyed by the booking's external_id,
// not its internal id.
func TestPublicBookingApprovedMailContainsDetailsLink(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	adminUser := CreateTestUserOrgAdmin(org)
	pending := setUpPublicBookingRequiringApproval(t, org, adminUser)

	SendMailMockContent = ""
	payload := `{"approved": true}`
	req := NewHTTPRequest("POST", "/booking/"+pending.ID+"/approve", adminUser.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	mail := waitForSendMailMockContent(t, 2*time.Second)

	approved, err := GetBookingRepository().GetOne(pending.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckStringNotEmpty(t, approved.PublicExternalID)
	CheckTestBool(t, true, strings.Contains(mail, "ui/book/details/"+approved.PublicExternalID+"/"))
	CheckTestBool(t, false, strings.Contains(mail, approved.ID))
}

func TestPublicBookingApprovedMailGermanDetailsLinkHasLangParam(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	adminUser := CreateTestUserOrgAdmin(org)
	pending := setUpPublicBookingRequiringApprovalWithLanguage(t, org, adminUser, "de")

	SendMailMockContent = ""
	payload := `{"approved": true}`
	req := NewHTTPRequest("POST", "/booking/"+pending.ID+"/approve", adminUser.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	mail := waitForSendMailMockContent(t, 2*time.Second)
	approved, err := GetBookingRepository().GetOne(pending.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestBool(t, true, strings.Contains(mail, "ui/book/details/"+approved.PublicExternalID+"/?lang=de"))
}

func TestPublicBookingApprovedMailEnglishDetailsLinkHasNoLangParam(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	adminUser := CreateTestUserOrgAdmin(org)
	pending := setUpPublicBookingRequiringApproval(t, org, adminUser)

	SendMailMockContent = ""
	payload := `{"approved": true}`
	req := NewHTTPRequest("POST", "/booking/"+pending.ID+"/approve", adminUser.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	mail := waitForSendMailMockContent(t, 2*time.Second)
	CheckTestBool(t, false, strings.Contains(mail, "lang=de"))
}

func TestPublicBookingDeclinedMailGermanNewBookingLinkHasLangParam(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	adminUser := CreateTestUserOrgAdmin(org)
	pending := setUpPublicBookingRequiringApprovalWithLanguage(t, org, adminUser, "de")

	SendMailMockContent = ""
	payload := `{"approved": false}`
	req := NewHTTPRequest("POST", "/booking/"+pending.ID+"/approve", adminUser.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	mail := waitForSendMailMockContent(t, 2*time.Second)
	CheckTestBool(t, true, strings.Contains(mail, "ui/book/?lang=de"))
}

func TestPublicBookingDetailsReturnsBookingForApprovedFutureBooking(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	adminUser := CreateTestUserOrgAdmin(org)
	pending := setUpPublicBookingRequiringApproval(t, org, adminUser)

	payload := `{"approved": true}`
	req := NewHTTPRequest("POST", "/booking/"+pending.ID+"/approve", adminUser.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	approved, err := GetBookingRepository().GetOne(pending.ID)
	if err != nil {
		t.Fatal(err)
	}

	req = NewHTTPRequest("GET", "/public-booking/details/"+approved.PublicExternalID, "", nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	var resBody PublicBookingDetailsResponse
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		t.Fatal(err)
	}
	CheckTestString(t, approved.Space.Name, resBody.SpaceName)
	CheckTestString(t, approved.Space.Location.Name, resBody.LocationName)
	if !resBody.Enter.Equal(approved.Enter) {
		t.Fatalf("expected enter %v, got %v", approved.Enter, resBody.Enter)
	}

	// Drain the async "approved" notification mail so its goroutine can't
	// still be running (and racing SendMailMockContent) once this test hands
	// off to the next one.
	waitForSendMailMockContent(t, 2*time.Second)
}

func TestPublicBookingDetailsReturnsNotFoundForUnknownExternalID(t *testing.T) {
	ClearTestDB()

	req := NewHTTPRequest("GET", "/public-booking/details/does-not-exist", "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

// TestPublicBookingDetailsReturnsNotFoundForPendingBooking guards the "must
// be approved" rule: a booking still awaiting approval must not be
// disclosable via its external_id yet.
func TestPublicBookingDetailsReturnsNotFoundForPendingBooking(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	adminUser := CreateTestUserOrgAdmin(org)
	pending := setUpPublicBookingRequiringApproval(t, org, adminUser)

	req := NewHTTPRequest("GET", "/public-booking/details/"+pending.PublicExternalID, "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

// TestPublicBookingDetailsReturnsNotFoundForPastBooking guards the
// "current or future only" rule.
func TestPublicBookingDetailsReturnsNotFoundForPastBooking(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	_, space := CreateTestLocationAndSpace(org)
	enablePublicBookingForOrgAndSpace(org, space)

	publicBooking := &PublicBooking{Name: "Jane Doe", Email: "jane.doe@test.com", Language: "en"}
	if err := GetPublicBookingRepository().Create(publicBooking); err != nil {
		t.Fatal(err)
	}
	booking := &Booking{
		SpaceID:  space.ID,
		Enter:    time.Now().Add(-48 * time.Hour),
		Leave:    time.Now().Add(-24 * time.Hour),
		Approved: true,
		PublicID: NullUUID(publicBooking.ID),
	}
	if err := GetBookingRepository().Create(booking); err != nil {
		t.Fatal(err)
	}

	req := NewHTTPRequest("GET", "/public-booking/details/"+publicBooking.ExternalID, "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestPublicBookingDeleteRemovesApprovedFutureBooking(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	adminUser := CreateTestUserOrgAdmin(org)
	pending := setUpPublicBookingRequiringApproval(t, org, adminUser)

	payload := `{"approved": true}`
	req := NewHTTPRequest("POST", "/booking/"+pending.ID+"/approve", adminUser.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	approved, err := GetBookingRepository().GetOne(pending.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Drain the async "approved" notification mail so its goroutine can't
	// still be running (and racing SendMailMockContent) once this test hands
	// off to the next one.
	waitForSendMailMockContent(t, 2*time.Second)
	SendMailMockContent = ""

	req = NewHTTPRequest("DELETE", "/public-booking/details/"+approved.PublicExternalID, "", nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	if _, err := GetBookingRepository().GetOne(approved.ID); err == nil {
		t.Fatal("expected booking to be deleted")
	}

	// Same for the "cancelled" notification mail triggered by the delete.
	waitForSendMailMockContent(t, 2*time.Second)
}

// TestPublicBookingDeleteMailOmitsYourBookingsButton guards against the
// self-service cancel mail reusing the authenticated "deleted" template,
// which links to a bookings page public bookers (no account) can't sign
// into.
func TestPublicBookingDeleteMailOmitsYourBookingsButton(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	adminUser := CreateTestUserOrgAdmin(org)
	pending := setUpPublicBookingRequiringApproval(t, org, adminUser)

	payload := `{"approved": true}`
	req := NewHTTPRequest("POST", "/booking/"+pending.ID+"/approve", adminUser.ID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	approved, err := GetBookingRepository().GetOne(pending.ID)
	if err != nil {
		t.Fatal(err)
	}

	req = NewHTTPRequest("DELETE", "/public-booking/details/"+approved.PublicExternalID, "", nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// The approve call above also sends a notification mail asynchronously;
	// wait for the "cancelled" one specifically rather than "any" mail, so a
	// still-in-flight "approved" mail can't be mistaken for it.
	mail := waitForSendMailMockContentContaining(t, "cancelled", 2*time.Second)
	CheckTestBool(t, true, strings.Contains(mail, "New booking"))
	CheckTestBool(t, true, strings.Contains(mail, "ui/book/"))
	CheckTestBool(t, false, strings.Contains(mail, "Your bookings"))
	CheckTestBool(t, false, strings.Contains(mail, "ui/bookings/"))
}

func TestPublicBookingDeleteReturnsNotFoundForUnknownExternalID(t *testing.T) {
	ClearTestDB()

	req := NewHTTPRequest("DELETE", "/public-booking/details/does-not-exist", "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

// TestPublicBookingDeleteReturnsNotFoundForPendingBooking guards against
// cancelling a booking that has not been approved yet through this
// unauthenticated endpoint (its external_id is not supposed to be
// disclosed/usable before approval).
func TestPublicBookingDeleteReturnsNotFoundForPendingBooking(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	adminUser := CreateTestUserOrgAdmin(org)
	pending := setUpPublicBookingRequiringApproval(t, org, adminUser)

	req := NewHTTPRequest("DELETE", "/public-booking/details/"+pending.PublicExternalID, "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)

	if _, err := GetBookingRepository().GetOne(pending.ID); err != nil {
		t.Fatal("expected booking to still exist")
	}
}

func TestPublicBookingGetSpacesReturnsBookableDays(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	location, space := CreateTestLocationAndSpace(org)
	enablePublicBookingForOrgAndSpace(org, space)
	location.BookableDays = "1,3,5"
	if err := GetLocationRepository().Update(location); err != nil {
		t.Fatal(err)
	}

	req := NewHTTPRequest("GET", "/public-booking/"+org.ID+"/spaces", "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetPublicBookableSpacesResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, 1, len(resBody.Spaces))
	CheckTestInt(t, 3, len(resBody.Spaces[0].BookableDays))
	CheckTestInt(t, 1, resBody.Spaces[0].BookableDays[0])
	CheckTestInt(t, 3, resBody.Spaces[0].BookableDays[1])
	CheckTestInt(t, 5, resBody.Spaces[0].BookableDays[2])
}

func TestPublicBookingGetSpacesReturnsEmptyBookableDaysWhenUnrestricted(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	_, space := CreateTestLocationAndSpace(org)
	enablePublicBookingForOrgAndSpace(org, space)

	req := NewHTTPRequest("GET", "/public-booking/"+org.ID+"/spaces", "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetPublicBookableSpacesResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, 1, len(resBody.Spaces))
	CheckTestInt(t, 0, len(resBody.Spaces[0].BookableDays))
}

func setTestLocationMap(t *testing.T, location *Location) {
	locationMap := &LocationMap{
		MimeType: "png",
		Width:    100,
		Height:   50,
		Scale:    1.0,
		Data:     []byte{1, 2, 3},
	}
	if err := GetLocationRepository().SetMap(location, locationMap); err != nil {
		t.Fatal(err)
	}
}

func TestPublicBookingGetSpacesShowMapDisabledByDefault(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	_, space := CreateTestLocationAndSpace(org)
	space.X = 10
	space.Y = 20
	space.Width = 30
	space.Height = 40
	space.Rotation = 90
	space.Shape = "circle"
	enablePublicBookingForOrgAndSpace(org, space)

	req := NewHTTPRequest("GET", "/public-booking/"+org.ID+"/spaces", "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetPublicBookableSpacesResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestBool(t, false, resBody.ShowMap)
	// Space geometry must not be disclosed if the map is disabled
	CheckTestInt(t, 1, len(resBody.Spaces))
	CheckTestUint(t, 0, resBody.Spaces[0].X)
	CheckTestUint(t, 0, resBody.Spaces[0].Y)
	CheckTestUint(t, 0, resBody.Spaces[0].Width)
	CheckTestUint(t, 0, resBody.Spaces[0].Height)
	CheckTestUint(t, 0, resBody.Spaces[0].Rotation)
	CheckTestString(t, "", resBody.Spaces[0].Shape)
	CheckTestString(t, "", resBody.Spaces[0].FontSize)
}

func TestPublicBookingGetSpacesReturnsShowMapAndCoordinates(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	_, space := CreateTestLocationAndSpace(org)
	space.X = 10
	space.Y = 20
	space.Width = 30
	space.Height = 40
	space.Rotation = 90
	space.Shape = "circle"
	enablePublicBookingForOrgAndSpace(org, space)
	GetSettingsRepository().Set(org.ID, SettingPublicBookingShowMap.Name, "1")

	req := NewHTTPRequest("GET", "/public-booking/"+org.ID+"/spaces", "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetPublicBookableSpacesResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestBool(t, true, resBody.ShowMap)
	CheckTestInt(t, 1, len(resBody.Spaces))
	CheckTestUint(t, 10, resBody.Spaces[0].X)
	CheckTestUint(t, 20, resBody.Spaces[0].Y)
	CheckTestUint(t, 30, resBody.Spaces[0].Width)
	CheckTestUint(t, 40, resBody.Spaces[0].Height)
	CheckTestUint(t, 90, resBody.Spaces[0].Rotation)
	CheckTestString(t, "circle", resBody.Spaces[0].Shape)
}

func TestPublicBookingGetMap(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	location, space := CreateTestLocationAndSpace(org)
	enablePublicBookingForOrgAndSpace(org, space)
	GetSettingsRepository().Set(org.ID, SettingPublicBookingShowMap.Name, "1")
	setTestLocationMap(t, location)

	req := NewHTTPRequest("GET", "/public-booking/"+org.ID+"/location/"+location.ID+"/map", "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetMapResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestUint(t, 100, resBody.Width)
	CheckTestUint(t, 50, resBody.Height)
	CheckTestString(t, "png", resBody.MimeType)
	CheckTestString(t, "AQID", resBody.Data)
}

func TestPublicBookingGetMapShowMapDisabled(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	location, space := CreateTestLocationAndSpace(org)
	enablePublicBookingForOrgAndSpace(org, space)
	setTestLocationMap(t, location)

	req := NewHTTPRequest("GET", "/public-booking/"+org.ID+"/location/"+location.ID+"/map", "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestPublicBookingGetMapPublicBookingDisabled(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	location, space := CreateTestLocationAndSpace(org)
	enablePublicBookingForOrgAndSpace(org, space)
	GetSettingsRepository().Set(org.ID, SettingPublicBookingShowMap.Name, "1")
	GetSettingsRepository().Set(org.ID, SettingPublicBookingEnabled.Name, "0")
	setTestLocationMap(t, location)

	req := NewHTTPRequest("GET", "/public-booking/"+org.ID+"/location/"+location.ID+"/map", "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestPublicBookingGetMapLocationWithoutPublicSpace(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	_, space := CreateTestLocationAndSpace(org)
	enablePublicBookingForOrgAndSpace(org, space)
	GetSettingsRepository().Set(org.ID, SettingPublicBookingShowMap.Name, "1")
	otherLocation, _ := CreateTestLocationAndSpace(org)
	setTestLocationMap(t, otherLocation)

	req := NewHTTPRequest("GET", "/public-booking/"+org.ID+"/location/"+otherLocation.ID+"/map", "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestPublicBookingGetMapForeignOrgLocation(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	_, space := CreateTestLocationAndSpace(org)
	enablePublicBookingForOrgAndSpace(org, space)
	GetSettingsRepository().Set(org.ID, SettingPublicBookingShowMap.Name, "1")

	org2 := CreateTestOrg("test2.com")
	location2, space2 := CreateTestLocationAndSpace(org2)
	enablePublicBookingForOrgAndSpace(org2, space2)
	setTestLocationMap(t, location2)

	req := NewHTTPRequest("GET", "/public-booking/"+org.ID+"/location/"+location2.ID+"/map", "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}
