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
		Country:          "US",
		AddressLine1:     "123 Main St",
		AddressLine2:     "Suite 100",
		PostalCode:       "12345",
		City:             "New York",
		VATID:            "US123456789",
	}
	err := GetOrganizationRepository().Create(org)
	CheckTestBool(t, true, err == nil)

	org2, err := GetOrganizationRepository().GetOne(org.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestBool(t, true, org != nil)
	CheckTestString(t, org.ID, org2.ID)
	CheckTestString(t, "Test Org", org2.Name)
	CheckTestString(t, "US", org2.Country)
	CheckTestString(t, "123 Main St", org2.AddressLine1)
	CheckTestString(t, "Suite 100", org2.AddressLine2)
	CheckTestString(t, "12345", org2.PostalCode)
	CheckTestString(t, "New York", org2.City)
	CheckTestString(t, "US123456789", org2.VATID)

	org2.Name = "New Name"
	org2.Country = "DE"
	org2.AddressLine1 = "456 Other St"
	org2.AddressLine2 = ""
	org2.PostalCode = "54321"
	org2.City = "Berlin"
	org2.VATID = "DE987654321"
	err = GetOrganizationRepository().Update(org2)
	CheckTestBool(t, true, err == nil)

	org3, err := GetOrganizationRepository().GetOne(org.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestBool(t, true, org3 != nil)
	CheckTestString(t, "New Name", org3.Name)
	CheckTestString(t, "DE", org3.Country)
	CheckTestString(t, "456 Other St", org3.AddressLine1)
	CheckTestString(t, "", org3.AddressLine2)
	CheckTestString(t, "54321", org3.PostalCode)
	CheckTestString(t, "Berlin", org3.City)
	CheckTestString(t, "DE987654321", org3.VATID)

	err = GetOrganizationRepository().Delete(org)
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

func TestOrganizationsAddressFields(t *testing.T) {
	ClearTestDB()
	org := &Organization{
		Name:             "Address Test Org",
		ContactFirstname: "John",
		ContactLastname:  "Doe",
		ContactEmail:     "john@doe.test",
		Language:         "en",
		SignupDate:       time.Now(),
		Country:          "GB",
		AddressLine1:     "10 Downing Street",
		AddressLine2:     "Westminster",
		PostalCode:       "SW1A 2AA",
		City:             "London",
		VATID:            "GB123456789",
	}
	err := GetOrganizationRepository().Create(org)
	CheckTestBool(t, true, err == nil)

	// Test GetOne
	retrieved, err := GetOrganizationRepository().GetOne(org.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestString(t, "GB", retrieved.Country)
	CheckTestString(t, "10 Downing Street", retrieved.AddressLine1)
	CheckTestString(t, "Westminster", retrieved.AddressLine2)
	CheckTestString(t, "SW1A 2AA", retrieved.PostalCode)
	CheckTestString(t, "London", retrieved.City)
	CheckTestString(t, "GB123456789", retrieved.VATID)

	// Test GetAll
	allOrgs, err := GetOrganizationRepository().GetAll()
	CheckTestBool(t, true, err == nil)
	CheckTestBool(t, true, len(allOrgs) > 0)
	found := false
	for _, o := range allOrgs {
		if o.ID == org.ID {
			found = true
			CheckTestString(t, "GB", o.Country)
			CheckTestString(t, "10 Downing Street", o.AddressLine1)
			CheckTestString(t, "Westminster", o.AddressLine2)
			CheckTestString(t, "SW1A 2AA", o.PostalCode)
			CheckTestString(t, "London", o.City)
			CheckTestString(t, "GB123456789", o.VATID)
		}
	}
	CheckTestBool(t, true, found)

	// Test GetByEmail
	byEmail, err := GetOrganizationRepository().GetByEmail("john@doe.test")
	CheckTestBool(t, true, err == nil)
	CheckTestString(t, "GB", byEmail.Country)
	CheckTestString(t, "10 Downing Street", byEmail.AddressLine1)
	CheckTestString(t, "Westminster", byEmail.AddressLine2)
	CheckTestString(t, "SW1A 2AA", byEmail.PostalCode)
	CheckTestString(t, "London", byEmail.City)
	CheckTestString(t, "GB123456789", byEmail.VATID)

	err = GetOrganizationRepository().Delete(org)
	CheckTestBool(t, true, err == nil)
}

func TestOrganizationsEmptyAddressFields(t *testing.T) {
	ClearTestDB()
	org := &Organization{
		Name:             "Minimal Org",
		ContactFirstname: "Jane",
		ContactLastname:  "Smith",
		ContactEmail:     "jane@smith.test",
		Language:         "en",
		SignupDate:       time.Now(),
		Country:          "",
		AddressLine1:     "",
		AddressLine2:     "",
		PostalCode:       "",
		City:             "",
		VATID:            "",
	}
	err := GetOrganizationRepository().Create(org)
	CheckTestBool(t, true, err == nil)

	retrieved, err := GetOrganizationRepository().GetOne(org.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestString(t, "", retrieved.Country)
	CheckTestString(t, "", retrieved.AddressLine1)
	CheckTestString(t, "", retrieved.AddressLine2)
	CheckTestString(t, "", retrieved.PostalCode)
	CheckTestString(t, "", retrieved.City)
	CheckTestString(t, "", retrieved.VATID)

	err = GetOrganizationRepository().Delete(org)
	CheckTestBool(t, true, err == nil)
}
