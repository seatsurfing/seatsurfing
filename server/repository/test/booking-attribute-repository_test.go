package test

import (
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func createTestBookingAttribute(t *testing.T, orgID, label string, attrType BookingAttributeType, required bool) *BookingAttribute {
	e := &BookingAttribute{OrganizationID: orgID, Label: label, Type: attrType, Required: required}
	created, err := GetBookingAttributeRepository().CreateIfBelowLimit(e, 3)
	if err != nil || !created {
		t.Fatalf("could not create booking attribute: %v", err)
	}
	return e
}

func TestBookingAttributeRepositoryCRUD(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")

	a1 := createTestBookingAttribute(t, org.ID, "Company", BookingAttributeTypeText, true)
	a2 := createTestBookingAttribute(t, org.ID, "Notes", BookingAttributeTypeTextMultiline, false)
	CheckTestInt(t, 1, a1.Position)
	CheckTestInt(t, 2, a2.Position)

	a11, err := GetBookingAttributeRepository().GetOne(a1.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestString(t, "Company", a11.Label)
	CheckTestInt(t, int(BookingAttributeTypeText), int(a11.Type))
	CheckTestBool(t, true, a11.Required)

	a11.Label = "Company name"
	a11.Type = BookingAttributeTypeNumber
	a11.Required = false
	if err := GetBookingAttributeRepository().Update(a11); err != nil {
		t.Fatal(err)
	}
	a12, _ := GetBookingAttributeRepository().GetOne(a1.ID)
	CheckTestString(t, "Company name", a12.Label)
	CheckTestInt(t, int(BookingAttributeTypeNumber), int(a12.Type))
	CheckTestBool(t, false, a12.Required)

	list, err := GetBookingAttributeRepository().GetAll(org.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 2, len(list))
	CheckTestString(t, a1.ID, list[0].ID)
	CheckTestString(t, a2.ID, list[1].ID)

	if err := GetBookingAttributeRepository().Delete(a1); err != nil {
		t.Fatal(err)
	}
	list, _ = GetBookingAttributeRepository().GetAll(org.ID)
	CheckTestInt(t, 1, len(list))
}

func TestBookingAttributeRepositoryLimit(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	org2 := CreateTestOrg("test2.com")
	createTestBookingAttribute(t, org.ID, "A", BookingAttributeTypeText, false)
	createTestBookingAttribute(t, org.ID, "B", BookingAttributeTypeText, false)

	e := &BookingAttribute{OrganizationID: org.ID, Label: "C", Type: BookingAttributeTypeText}
	created, err := GetBookingAttributeRepository().CreateIfBelowLimit(e, 2)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestBool(t, false, created)
	CheckTestString(t, "", e.ID)

	// The limit is per organization
	createTestBookingAttribute(t, org2.ID, "A", BookingAttributeTypeText, false)
}

func TestBookingAttributeRepositoryUpdateDeleteForeignOrg(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	org2 := CreateTestOrg("test2.com")
	a := createTestBookingAttribute(t, org.ID, "A", BookingAttributeTypeText, false)

	foreign := &BookingAttribute{ID: a.ID, OrganizationID: org2.ID, Label: "Changed", Type: BookingAttributeTypeText}
	GetBookingAttributeRepository().Update(foreign)
	GetBookingAttributeRepository().Delete(foreign)

	a1, err := GetBookingAttributeRepository().GetOne(a.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestString(t, "A", a1.Label)
}

func TestBookingAttributeValueRepository(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	_, space := CreateTestLocationAndSpace(org)
	b1 := CreateTestBooking9To5(user, space, 1)
	b2 := CreateTestBooking9To5(user, space, 2)
	a1 := createTestBookingAttribute(t, org.ID, "Company", BookingAttributeTypeText, false)
	a2 := createTestBookingAttribute(t, org.ID, "Persons", BookingAttributeTypeNumber, false)

	CheckTestIsNil(t, GetBookingAttributeValueRepository().Set(b1.ID, a2.ID, "4"))
	CheckTestIsNil(t, GetBookingAttributeValueRepository().Set(b1.ID, a1.ID, "ACME"))
	CheckTestIsNil(t, GetBookingAttributeValueRepository().Set(b2.ID, a1.ID, "Foo"))
	CheckTestIsNil(t, GetBookingAttributeValueRepository().Set(b2.ID, a1.ID, "Bar"))

	list, err := GetBookingAttributeValueRepository().GetAllForBooking(b1.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 2, len(list))
	CheckTestString(t, "Company", list[0].Label)
	CheckTestString(t, "ACME", list[0].Value)
	CheckTestString(t, "Persons", list[1].Label)
	CheckTestInt(t, int(BookingAttributeTypeNumber), int(list[1].Type))
	CheckTestString(t, "4", list[1].Value)

	list, _ = GetBookingAttributeValueRepository().GetAllForBookingList([]string{b1.ID, b2.ID})
	CheckTestInt(t, 3, len(list))

	list, _ = GetBookingAttributeValueRepository().GetAllForBooking(b2.ID)
	CheckTestInt(t, 1, len(list))
	CheckTestString(t, "Bar", list[0].Value)

	// Deleting the attribute removes its values
	GetBookingAttributeRepository().Delete(a2)
	list, _ = GetBookingAttributeValueRepository().GetAllForBooking(b1.ID)
	CheckTestInt(t, 1, len(list))

	// Deleting the booking removes its values
	b1Details, _ := GetBookingRepository().GetOne(b1.ID)
	GetBookingRepository().Delete(b1Details)
	list, _ = GetBookingAttributeValueRepository().GetAllForBooking(b1.ID)
	CheckTestInt(t, 0, len(list))
}

func TestBookingAttributeValueRepositoryForeignOrgAttribute(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	org2 := CreateTestOrg("test2.com")
	user := CreateTestUserInOrg(org)
	_, space := CreateTestLocationAndSpace(org)
	b := CreateTestBooking9To5(user, space, 1)
	foreign := createTestBookingAttribute(t, org2.ID, "Company", BookingAttributeTypeText, false)

	CheckTestIsNil(t, GetBookingAttributeValueRepository().Set(b.ID, foreign.ID, "ACME"))
	list, _ := GetBookingAttributeValueRepository().GetAllForBooking(b.ID)
	CheckTestInt(t, 0, len(list))
}
