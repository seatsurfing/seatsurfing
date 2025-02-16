package test

import (
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestAuthAttemptRepositoryBanSimple(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrgWithName(org, "u1@test.com", UserRoleUser)

	CheckTestBool(t, false, AuthAttemptRepositoryIsUserDisabled(t, user.ID))

	// Attempt 1
	if err := GetAuthAttemptRepository().RecordLoginAttempt(user, false); err != nil {
		t.Error(err)
	}
	CheckTestBool(t, false, AuthAttemptRepositoryIsUserDisabled(t, user.ID))

	// Attempt 2
	if err := GetAuthAttemptRepository().RecordLoginAttempt(user, false); err != nil {
		t.Error(err)
	}
	CheckTestBool(t, false, AuthAttemptRepositoryIsUserDisabled(t, user.ID))

	// Attempt 3
	if err := GetAuthAttemptRepository().RecordLoginAttempt(user, false); err != nil {
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
	GetAuthAttemptRepository().RecordLoginAttempt(user, false)
	CheckTestBool(t, false, AuthAttemptRepositoryIsUserDisabled(t, user.ID))

	// Attempt 2
	GetAuthAttemptRepository().RecordLoginAttempt(user, false)
	CheckTestBool(t, false, AuthAttemptRepositoryIsUserDisabled(t, user.ID))

	// Successful Login
	GetAuthAttemptRepository().RecordLoginAttempt(user, true)
	CheckTestBool(t, false, AuthAttemptRepositoryIsUserDisabled(t, user.ID))

	// Attempt 1
	GetAuthAttemptRepository().RecordLoginAttempt(user, false)
	CheckTestBool(t, false, AuthAttemptRepositoryIsUserDisabled(t, user.ID))

	// Attempt 2
	GetAuthAttemptRepository().RecordLoginAttempt(user, false)
	CheckTestBool(t, false, AuthAttemptRepositoryIsUserDisabled(t, user.ID))

	// Attempt 3
	GetAuthAttemptRepository().RecordLoginAttempt(user, false)
	CheckTestBool(t, true, AuthAttemptRepositoryIsUserDisabled(t, user.ID))
}
