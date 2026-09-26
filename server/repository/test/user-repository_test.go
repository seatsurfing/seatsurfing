package test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	. "github.com/seatsurfing/seatsurfing/server/api"
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
	CheckTestBool(t, true, err == nil)
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
	CheckTestBool(t, true, err == nil)
	CheckTestString(t, user.ID, user3.ID)
	CheckTestString(t, "61bf23af-0310-4d2b-b401-21c31d60c2c4", user3.OrganizationID)

	// Delete
	err = GetUserRepository().Delete(user)
	CheckTestBool(t, true, err == nil)

	_, err = GetUserRepository().GetOne(user.ID)
	CheckTestBool(t, true, err != nil)
}

func TestUsersCount(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	CreateTestUserInOrg(org)
	CreateTestUserInOrg(org)

	res, err := GetUserRepository().GetCount(org.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 2, res)
}

func TestUsersCountHuman(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	CreateTestUserInOrg(org)
	CreateTestUserInOrg(org)

	roUser := CreateTestUserInOrg(org)
	roUser.AccountType = AccountTypeServiceAccountRO
	GetUserRepository().Update(roUser)

	rwUser := CreateTestUserInOrg(org)
	rwUser.AccountType = AccountTypeServiceAccountRW
	GetUserRepository().Update(rwUser)

	res, err := GetUserRepository().GetCountHuman(org.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 2, res)

	res, err = GetUserRepository().GetCount(org.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 4, res)
}

func TestUsersCountHumanScopedByOrg(t *testing.T) {
	ClearTestDB()
	org1 := CreateTestOrg("test1.com")
	org2 := CreateTestOrg("test2.com")
	CreateTestUserInOrg(org1)
	CreateTestUserInOrg(org1)
	CreateTestUserInOrg(org2)

	res, err := GetUserRepository().GetCountHuman(org1.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 2, res)

	res, err = GetUserRepository().GetCountHuman(org2.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, res)
}

func TestUsersLastActivity(t *testing.T) {
	ClearTestDB()

	// Create
	user := &User{
		Email:             uuid.New().String() + "@test.com",
		OrganizationID:    "73980078-f4d7-40ff-9211-a7bcbf8d1981",
		LastActivityAtUTC: nil,
	}
	GetUserRepository().Create(user)
	CheckStringNotEmpty(t, user.ID)

	// Read
	user2, err := GetUserRepository().GetOne(user.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestBool(t, true, user2.LastActivityAtUTC == nil)

	// Update
	t1 := time.Now().Add(-10 * time.Minute).UTC().Truncate(time.Second)
	user2.LastActivityAtUTC = &t1
	err = GetUserRepository().Update(user2)
	CheckTestBool(t, true, err == nil)

	// Read
	user3, err := GetUserRepository().GetOne(user.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestBool(t, true, user3.LastActivityAtUTC != nil)
	if user3.LastActivityAtUTC != nil {
		CheckTestBool(t, true, user3.LastActivityAtUTC.Equal(t1))
	}
}

func TestUsersGetSafeRecipientName(t *testing.T) {
	u1 := &User{
		Email:     "foo@bar.com",
		Firstname: "",
		Lastname:  "",
	}
	u2 := &User{
		Email:     "fb@bar.com",
		Firstname: "",
		Lastname:  "",
	}
	u3 := &User{
		Email:     "f.bar@bar.com",
		Firstname: "",
		Lastname:  "",
	}
	u4 := &User{
		Email:     "f.bar@bar.com",
		Firstname: "Mr. Foo",
		Lastname:  "Bar",
	}
	u5 := &User{
		Email:     "nodomain",
		Firstname: "",
		Lastname:  "",
	}
	u6 := &User{
		Email:     "\"a@b\"@example.com",
		Firstname: "",
		Lastname:  "",
	}
	CheckTestString(t, "Foo", u1.GetSafeRecipientName())
	CheckTestString(t, "Fb", u2.GetSafeRecipientName())
	CheckTestString(t, "F.Bar", u3.GetSafeRecipientName())
	CheckTestString(t, "Mr. Foo", u4.GetSafeRecipientName())
	CheckTestString(t, "Nodomain", u5.GetSafeRecipientName())
	CheckTestString(t, "\"A@B\"", u6.GetSafeRecipientName())
}

func TestBookingDetailsGetSafeRecipientName(t *testing.T) {
	b1 := &BookingDetails{
		UserEmail:     "foo@bar.com",
		UserFirstname: "",
	}
	b2 := &BookingDetails{
		UserEmail:     "foo@bar.com",
		UserFirstname: "Mr. Foo",
	}
	CheckTestString(t, "Foo", b1.GetSafeRecipientName())
	CheckTestString(t, "Mr. Foo", b2.GetSafeRecipientName())
}

func TestSetUserPreferenceMailNotification(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	// test mail notification is enabled by default
	mailNotification, _ := GetUserPreferencesRepository().GetBool(user.ID, PreferenceMailNotifications.Name)
	CheckTestBool(t, true, mailNotification)

	// test mail notification is *not* by default if org setting was changed
	GetSettingsRepository().Set(org.ID, SettingNewUserDefaultMailNotification.Name, "0")
	user = CreateTestUserInOrg(org)
	mailNotification, _ = GetUserPreferencesRepository().GetBool(user.ID, PreferenceMailNotifications.Name)
	CheckTestBool(t, false, mailNotification)
}
