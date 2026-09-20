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
	id := createTestPublicBookingAuthState(t, space, enter, leave)

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
