package test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestUsersCRUD(t *testing.T) {
	ClearTestDB()

	// Create
	user := &User{
		Email:          uuid.New().String() + "@test.com",
		OrganizationID: "73980078-f4d7-40ff-9211-a7bcbf8d1981",
	}
	GetUserRepository().Create(user)
	CheckStringNotEmpty(t, user.ID)

	// Read
	user2, err := GetUserRepository().GetOne(user.ID)
	if err != nil {
		t.Fatalf("Expected non-nil user")
	}
	CheckTestString(t, user.ID, user2.ID)
	CheckTestString(t, "73980078-f4d7-40ff-9211-a7bcbf8d1981", user.OrganizationID)

	// Update
	user2 = &User{
		ID:             user.ID,
		OrganizationID: "61bf23af-0310-4d2b-b401-21c31d60c2c4",
	}
	GetUserRepository().Update(user2)

	// Read
	user3, err := GetUserRepository().GetOne(user.ID)
	if err != nil {
		t.Fatalf("Expected non-nil user")
	}
	CheckTestString(t, user.ID, user3.ID)
	CheckTestString(t, "61bf23af-0310-4d2b-b401-21c31d60c2c4", user3.OrganizationID)

	// Delete
	GetUserRepository().Delete(user)
	_, err = GetUserRepository().GetOne(user.ID)
	if err == nil {
		t.Fatalf("Expected nil user")
	}
}

func TestUsersCount(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	CreateTestUserInOrg(org)
	CreateTestUserInOrg(org)

	res, err := GetUserRepository().GetCount(org.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 2, res)
}

func TestDeleteObsoleteConfluenceAnonymousUsers(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	u1 := CreateTestUserInOrg(org) // Regular user 1
	u2 := CreateTestUserInOrg(org) // Regular user 2

	// Confluence User 1 with recent login (not to be deleted)
	cu1 := CreateTestUserInOrgWithName(org, "confluence-anonymous-"+uuid.New().String()+"@test.com", UserRoleUser)
	GetAuthAttemptRepository().RecordLoginAttempt(cu1, true)

	// Confluence User 2 without login (to be deleted)
	CreateTestUserInOrgWithName(org, "confluence-anonymous-"+uuid.New().String()+"@test.com", UserRoleUser)

	// Confluence User 3 with old login (to be deleted)
	cu3 := CreateTestUserInOrgWithName(org, "confluence-anonymous-"+uuid.New().String()+"@test.com", UserRoleUser)
	la := &AuthAttempt{
		UserID:     cu3.ID,
		Email:      cu3.Email,
		Timestamp:  time.Now().Add(-26 * time.Hour),
		Successful: true,
	}
	GetAuthAttemptRepository().Create(la)

	// Confluence User 4 with recent failed login (to be deleted)
	cu4 := CreateTestUserInOrgWithName(org, "confluence-anonymous-"+uuid.New().String()+"@test.com", UserRoleUser)
	la = &AuthAttempt{
		UserID:     cu4.ID,
		Email:      cu4.Email,
		Timestamp:  time.Now().Add(-5 * time.Hour),
		Successful: false,
	}
	GetAuthAttemptRepository().Create(la)

	num, err := GetUserRepository().DeleteObsoleteConfluenceAnonymousUsers()
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 3, num)

	users, _ := GetUserRepository().GetAll(org.ID, 10000, 0)
	CheckTestInt(t, 3, len(users))

	invalid := false
	for _, user := range users {
		if !((user.ID == u1.ID) ||
			(user.ID == u2.ID) ||
			(user.ID == cu1.ID)) {
			invalid = true
		}
	}
	CheckTestBool(t, false, invalid)
}
