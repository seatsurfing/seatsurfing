package test

import (
	"testing"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestOrganizationsCRUD(t *testing.T) {
	ClearTestDB()
	org := &Organization{
		Name:             "Test Org",
		ContactFirstname: "Alice",
		ContactLastname:  "Bob",
		ContactEmail:     "alice@bob.test",
		Language:         "en",
		SignupDate:       time.Now(),
	}
	err := GetOrganizationRepository().Create(org)
	CheckTestBool(t, true, err == nil)

	org2, err := GetOrganizationRepository().GetOne(org.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestBool(t, true, org != nil)
	CheckTestString(t, org.ID, org2.ID)
	CheckTestString(t, "Test Org", org2.Name)

	org2.Name = "New Name"
	err = GetOrganizationRepository().Update(org2)
	CheckTestBool(t, true, err == nil)

	org3, err := GetOrganizationRepository().GetOne(org.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestBool(t, true, org3 != nil)
	CheckTestString(t, "New Name", org3.Name)

	err = GetOrganizationRepository().DeleteSoft(org)
	CheckTestBool(t, true, err == nil)

	err = GetOrganizationRepository().DeleteHard(org)
	CheckTestBool(t, true, err == nil)
}

func TestOrganizationsGetAllDaysPassedSinceSignupPositive(t *testing.T) {
	ClearTestDB()
	CreateTestOrg("test1.com")
	org2 := CreateTestOrg("test2.com")
	org2.SignupDate = org2.SignupDate.AddDate(0, 0, -1)
	org2.SignupDate = time.Date(org2.SignupDate.Year(), org2.SignupDate.Month(), org2.SignupDate.Day(), 0, 0, 0, 0, org2.SignupDate.Location())
	GetOrganizationRepository().Update(org2)

	list, err := GetOrganizationRepository().GetAllDaysPassedSinceSignup(1, "")
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, len(list))
	CheckTestString(t, org2.ID, list[0].ID)
}

func TestOrganizationsGetAllDaysPassedSinceSignupNegative(t *testing.T) {
	ClearTestDB()
	CreateTestOrg("test1.com")
	org2 := CreateTestOrg("test2.com")
	org2.SignupDate = org2.SignupDate.AddDate(0, 0, -2)
	org2.SignupDate = time.Date(org2.SignupDate.Year(), org2.SignupDate.Month(), org2.SignupDate.Day(), 0, 0, 0, 0, org2.SignupDate.Location())
	GetOrganizationRepository().Update(org2)

	list, err := GetOrganizationRepository().GetAllDaysPassedSinceSignup(1, "")
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 0, len(list))
}

func TestOrganizationsGetAllDaysPassedSinceSignupWithSettingExists(t *testing.T) {
	ClearTestDB()
	org1 := CreateTestOrg("test1.com")
	org1.SignupDate = org1.SignupDate.AddDate(0, 0, -1)
	org1.SignupDate = time.Date(org1.SignupDate.Year(), org1.SignupDate.Month(), org1.SignupDate.Day(), 0, 0, 0, 0, org1.SignupDate.Location())
	GetSettingsRepository().Set(org1.ID, "test_setting", "value")
	org2 := CreateTestOrg("test2.com")
	org2.SignupDate = org2.SignupDate.AddDate(0, 0, -1)
	org2.SignupDate = time.Date(org2.SignupDate.Year(), org2.SignupDate.Month(), org2.SignupDate.Day(), 0, 0, 0, 0, org2.SignupDate.Location())
	GetOrganizationRepository().Update(org2)

	list, err := GetOrganizationRepository().GetAllDaysPassedSinceSignup(1, "test_setting")
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, len(list))
	CheckTestString(t, org2.ID, list[0].ID)
}
