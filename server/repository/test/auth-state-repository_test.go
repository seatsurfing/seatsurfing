package test

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestAuthStateRepositoryClaimActiveSimple(t *testing.T) {
	ClearTestDB()
	authState := &AuthState{
		Expiry:        time.Now().Add(30 * time.Minute),
		AuthStateType: AuthPublicBooking,
		Payload:       `{"foo":"bar"}`,
		Key:           "jane.doe@test.com",
	}
	if err := GetAuthStateRepository().Create(authState); err != nil {
		t.Fatal(err)
	}

	claimed, err := GetAuthStateRepository().ClaimActive(authState.ID, AuthPublicBooking)
	CheckTestBool(t, true, err == nil)
	CheckTestString(t, `{"foo":"bar"}`, claimed.Payload)

	// Second claim must fail: the row was deleted by the first claim.
	_, err = GetAuthStateRepository().ClaimActive(authState.ID, AuthPublicBooking)
	CheckTestBool(t, true, err != nil)

	// The state is also gone for a plain lookup.
	_, err = GetAuthStateRepository().GetOne(authState.ID)
	CheckTestBool(t, true, err != nil)
}

func TestAuthStateRepositoryClaimActiveWrongType(t *testing.T) {
	ClearTestDB()
	authState := &AuthState{
		Expiry:        time.Now().Add(30 * time.Minute),
		AuthStateType: AuthTotpSetup,
		Payload:       `{"foo":"bar"}`,
		Key:           "jane.doe@test.com",
	}
	if err := GetAuthStateRepository().Create(authState); err != nil {
		t.Fatal(err)
	}

	// Claiming with a mismatched type must fail and must not delete the row,
	// so an ID from an unrelated flow (e.g. TOTP setup) cannot be invalidated
	// by posting it to an endpoint that claims a different type (e.g. public
	// booking confirmation).
	_, err := GetAuthStateRepository().ClaimActive(authState.ID, AuthPublicBooking)
	CheckTestBool(t, true, err != nil)

	stillThere, err := GetAuthStateRepository().GetOne(authState.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestString(t, `{"foo":"bar"}`, stillThere.Payload)
}

func TestAuthStateRepositoryClaimActiveExpired(t *testing.T) {
	ClearTestDB()
	authState := &AuthState{
		Expiry:        time.Now().Add(-1 * time.Minute),
		AuthStateType: AuthPublicBooking,
		Payload:       `{"foo":"bar"}`,
		Key:           "jane.doe@test.com",
	}
	if err := GetAuthStateRepository().Create(authState); err != nil {
		t.Fatal(err)
	}

	_, err := GetAuthStateRepository().ClaimActive(authState.ID, AuthPublicBooking)
	CheckTestBool(t, true, err != nil)
}

// TestAuthStateRepositoryClaimActiveConcurrent guards against the double-opt-in
// public booking regression where GetOneActive (a plain SELECT) followed by a
// deferred Delete let two concurrent confirmations of the same one-time link
// both observe the state as active and each create a booking. ClaimActive
// folds the read and the invalidation into a single atomic DELETE ... RETURNING,
// so Postgres serializes concurrent claims on the same row and only one caller
// ever gets the payload back.
func TestAuthStateRepositoryClaimActiveConcurrent(t *testing.T) {
	ClearTestDB()
	payload, err := json.Marshal(map[string]string{"foo": "bar"})
	if err != nil {
		t.Fatal(err)
	}
	authState := &AuthState{
		Expiry:        time.Now().Add(30 * time.Minute),
		AuthStateType: AuthPublicBooking,
		Payload:       string(payload),
		Key:           "jane.doe@test.com",
	}
	if err := GetAuthStateRepository().Create(authState); err != nil {
		t.Fatal(err)
	}

	numAttempts := 20
	var successCount int32
	var wg sync.WaitGroup
	wg.Add(numAttempts)
	for i := 0; i < numAttempts; i++ {
		go func() {
			defer wg.Done()
			if _, err := GetAuthStateRepository().ClaimActive(authState.ID, AuthPublicBooking); err == nil {
				atomic.AddInt32(&successCount, 1)
			}
		}()
	}
	wg.Wait()

	CheckTestInt(t, 1, int(successCount))
}
