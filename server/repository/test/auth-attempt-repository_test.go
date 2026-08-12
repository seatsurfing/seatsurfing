package test

import (
	"testing"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestAuthAttemptRepositoryBanSimple(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrgWithName(org, "u1@test.com", UserRoleUser)

	CheckTestBool(t, false, AuthAttemptRepositoryIsUserDisabled(t, user.ID))

	// Attempt 1
	if err := GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{User: user, Method: AuthMethodPassword, ErrorCode: AuthErrorWrongPassword, BanCheck: true}); err != nil {
		t.Error(err)
	}
	CheckTestBool(t, false, AuthAttemptRepositoryIsUserDisabled(t, user.ID))

	// Attempt 2
	if err := GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{User: user, Method: AuthMethodPassword, ErrorCode: AuthErrorWrongPassword, BanCheck: true}); err != nil {
		t.Error(err)
	}
	CheckTestBool(t, false, AuthAttemptRepositoryIsUserDisabled(t, user.ID))

	// Attempt 3
	if err := GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{User: user, Method: AuthMethodPassword, ErrorCode: AuthErrorWrongPassword, BanCheck: true}); err != nil {
		t.Error(err)
	}
	CheckTestBool(t, true, AuthAttemptRepositoryIsUserDisabled(t, user.ID))
}

func TestAuthAttemptRepositoryBanWithSuccess(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrgWithName(org, "u1@test.com", UserRoleUser)

	CheckTestBool(t, false, AuthAttemptRepositoryIsUserDisabled(t, user.ID))

	// Attempt 1
	GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{User: user, Method: AuthMethodPassword, ErrorCode: AuthErrorWrongPassword, BanCheck: true})
	CheckTestBool(t, false, AuthAttemptRepositoryIsUserDisabled(t, user.ID))

	// Attempt 2
	GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{User: user, Method: AuthMethodPassword, ErrorCode: AuthErrorWrongPassword, BanCheck: true})
	CheckTestBool(t, false, AuthAttemptRepositoryIsUserDisabled(t, user.ID))

	// Successful Login
	GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{User: user, Successful: true, Method: AuthMethodPassword, BanCheck: true})
	CheckTestBool(t, false, AuthAttemptRepositoryIsUserDisabled(t, user.ID))

	// Attempt 1
	GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{User: user, Method: AuthMethodPassword, ErrorCode: AuthErrorWrongPassword, BanCheck: true})
	CheckTestBool(t, false, AuthAttemptRepositoryIsUserDisabled(t, user.ID))

	// Attempt 2
	GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{User: user, Method: AuthMethodPassword, ErrorCode: AuthErrorWrongPassword, BanCheck: true})
	CheckTestBool(t, false, AuthAttemptRepositoryIsUserDisabled(t, user.ID))

	// Attempt 3
	GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{User: user, Method: AuthMethodPassword, ErrorCode: AuthErrorWrongPassword, BanCheck: true})
	CheckTestBool(t, true, AuthAttemptRepositoryIsUserDisabled(t, user.ID))
}

func TestAuthAttemptRepositoryFilterAndOrgIsolation(t *testing.T) {
	ClearTestDB()
	org1 := CreateTestOrg("test1.com")
	org2 := CreateTestOrg("test2.com")
	user1 := CreateTestUserInOrgWithName(org1, "u1@test1.com", UserRoleUser)
	user2 := CreateTestUserInOrgWithName(org2, "u2@test2.com", UserRoleUser)

	GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{User: user1, Method: AuthMethodPassword, ErrorCode: AuthErrorWrongPassword, Device: "Chrome on Windows"})
	GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{User: user1, Successful: true, Method: AuthMethodPassword})
	GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{OrganizationID: org1.ID, Email: "unknown@test1.com", Method: AuthMethodOAuth, ErrorCode: AuthErrorUserNotFound})
	GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{User: user2, Method: AuthMethodPassword, ErrorCode: AuthErrorWrongPassword})

	filter := &AuthAttemptFilter{
		OrganizationID: org1.ID,
		Start:          time.Now().Add(-1 * time.Hour),
		End:            time.Now().Add(1 * time.Hour),
	}
	total, err := GetAuthAttemptRepository().CountFiltered(filter)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 3, total)
	list, err := GetAuthAttemptRepository().GetFiltered(filter, 100, 0)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 3, len(list))
	for _, e := range list {
		CheckTestString(t, org1.ID, e.OrganizationID)
	}

	// filter by success
	success := false
	filter.Successful = &success
	total, _ = GetAuthAttemptRepository().CountFiltered(filter)
	CheckTestInt(t, 2, total)
	filter.Successful = nil

	// filter by method
	filter.Method = AuthMethodOAuth
	list, _ = GetAuthAttemptRepository().GetFiltered(filter, 100, 0)
	CheckTestInt(t, 1, len(list))
	CheckTestString(t, AuthErrorUserNotFound, list[0].ErrorCode)
	CheckTestString(t, "unknown@test1.com", list[0].Email)
	CheckTestString(t, "", list[0].UserID)
	filter.Method = ""

	// filter by error code
	filter.ErrorCode = AuthErrorWrongPassword
	list, _ = GetAuthAttemptRepository().GetFiltered(filter, 100, 0)
	CheckTestInt(t, 1, len(list))
	CheckTestString(t, user1.ID, list[0].UserID)
	CheckTestString(t, "Chrome on Windows", list[0].Device)
	filter.ErrorCode = ""

	// filter by email substring
	filter.EmailLike = "UNKNOWN"
	list, _ = GetAuthAttemptRepository().GetFiltered(filter, 100, 0)
	CheckTestInt(t, 1, len(list))
	filter.EmailLike = ""

	// limit and offset
	list, _ = GetAuthAttemptRepository().GetFiltered(filter, 2, 0)
	CheckTestInt(t, 2, len(list))
	list, _ = GetAuthAttemptRepository().GetFiltered(filter, 2, 2)
	CheckTestInt(t, 1, len(list))
}

func TestAuthAttemptRepositoryPurgeOld(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrgWithName(org, "u1@test.com", UserRoleUser)

	old := &AuthAttempt{OrganizationID: org.ID, UserID: user.ID, Email: user.Email, Timestamp: time.Now().Add(-100 * 24 * time.Hour), Successful: false}
	GetAuthAttemptRepository().Create(old)
	old2 := &AuthAttempt{OrganizationID: org.ID, UserID: user.ID, Email: user.Email, Timestamp: time.Now().Add(-95 * 24 * time.Hour), Successful: false}
	GetAuthAttemptRepository().Create(old2)
	recent := &AuthAttempt{OrganizationID: org.ID, UserID: user.ID, Email: user.Email, Timestamp: time.Now().Add(-1 * time.Hour), Successful: true}
	GetAuthAttemptRepository().Create(recent)

	// batch size 1 deletes only the oldest row
	num, err := GetAuthAttemptRepository().PurgeOld(90*24*time.Hour, 1)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, num)

	num, err = GetAuthAttemptRepository().PurgeOld(90*24*time.Hour, 100)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, num)

	filter := &AuthAttemptFilter{OrganizationID: org.ID, Start: time.Now().Add(-101 * 24 * time.Hour), End: time.Now()}
	total, _ := GetAuthAttemptRepository().CountFiltered(filter)
	CheckTestInt(t, 1, total)
}

func TestAuthAttemptRepositoryDeleteAll(t *testing.T) {
	ClearTestDB()
	org1 := CreateTestOrg("test1.com")
	org2 := CreateTestOrg("test2.com")
	user1 := CreateTestUserInOrgWithName(org1, "u1@test1.com", UserRoleUser)
	user2 := CreateTestUserInOrgWithName(org2, "u2@test2.com", UserRoleUser)
	GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{User: user1, Method: AuthMethodPassword, ErrorCode: AuthErrorWrongPassword})
	GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{User: user2, Method: AuthMethodPassword, ErrorCode: AuthErrorWrongPassword})

	if err := GetAuthAttemptRepository().DeleteAll(org1.ID); err != nil {
		t.Error(err)
	}

	filter := &AuthAttemptFilter{OrganizationID: org1.ID, Start: time.Now().Add(-1 * time.Hour), End: time.Now().Add(1 * time.Hour)}
	total, _ := GetAuthAttemptRepository().CountFiltered(filter)
	CheckTestInt(t, 0, total)
	filter.OrganizationID = org2.ID
	total, _ = GetAuthAttemptRepository().CountFiltered(filter)
	CheckTestInt(t, 1, total)
}

// TestAuthAttemptRepositoryBanIgnoresNonPasswordMethods verifies that OAuth/Confluence
// failures (persisted for the audit log but not ban-relevant) never contribute to
// banning a user out of password/TOTP/passkey login.
func TestAuthAttemptRepositoryBanIgnoresNonPasswordMethods(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrgWithName(org, "u1@test.com", UserRoleUser)

	// Plenty of OAuth/Confluence failures for the same user, none of which
	// call checkBanUser themselves (BanCheck: false, as in the real code paths).
	for i := 0; i < 5; i++ {
		GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{User: user, Method: AuthMethodOAuth, ErrorCode: AuthErrorIdpProviderMismatch})
	}
	for i := 0; i < 5; i++ {
		GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{User: user, Method: AuthMethodConfluence, ErrorCode: AuthErrorConfluenceJwtInvalid})
	}
	CheckTestBool(t, false, AuthAttemptRepositoryIsUserDisabled(t, user.ID))

	// Two real password failures (ban threshold in tests is 3, see LOGIN_PROTECTION_MAX_FAILS).
	GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{User: user, Method: AuthMethodPassword, ErrorCode: AuthErrorWrongPassword, BanCheck: true})
	CheckTestBool(t, false, AuthAttemptRepositoryIsUserDisabled(t, user.ID))
	GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{User: user, Method: AuthMethodPassword, ErrorCode: AuthErrorWrongPassword, BanCheck: true})
	// If the OAuth/Confluence rows above were (incorrectly) counted, this would already be banned.
	CheckTestBool(t, false, AuthAttemptRepositoryIsUserDisabled(t, user.ID))

	// Third real password failure crosses the threshold.
	GetAuthAttemptRepository().RecordAuthEvent(&AuthEvent{User: user, Method: AuthMethodPassword, ErrorCode: AuthErrorWrongPassword, BanCheck: true})
	CheckTestBool(t, true, AuthAttemptRepositoryIsUserDisabled(t, user.ID))
}
