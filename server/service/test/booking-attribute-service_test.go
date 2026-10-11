package test

import (
	"strings"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/service"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func createTestServiceBookingAttribute(t *testing.T, admin *User, label string, attrType BookingAttributeType, required bool) *BookingAttribute {
	e, bErr := GetBookingAttributeService().Create(admin, &BookingAttributeInput{Label: label, Type: attrType, Required: required})
	if bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}
	return e
}

func TestBookingAttributeServiceCRUD(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)

	e := createTestServiceBookingAttribute(t, admin, "  Company  ", BookingAttributeTypeText, true)
	CheckTestString(t, "Company", e.Label)

	if bErr := GetBookingAttributeService().Update(admin, e.ID, &BookingAttributeInput{Label: "Persons", Type: BookingAttributeTypeNumber}); bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}
	list, bErr := GetBookingAttributeService().List(admin)
	if bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}
	CheckTestInt(t, 1, len(list))
	CheckTestString(t, "Persons", list[0].Label)
	CheckTestInt(t, int(BookingAttributeTypeNumber), int(list[0].Type))
	CheckTestBool(t, false, list[0].Required)

	if bErr := GetBookingAttributeService().Delete(admin, e.ID); bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}
	list, _ = GetBookingAttributeService().List(admin)
	CheckTestInt(t, 0, len(list))
}

func TestBookingAttributeServiceInvalidInput(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)

	for _, in := range []*BookingAttributeInput{
		{Label: "   ", Type: BookingAttributeTypeText},
		{Label: strings.Repeat("a", 257), Type: BookingAttributeTypeText},
		{Label: "Foo", Type: 0},
		{Label: "Foo", Type: 4},
	} {
		_, bErr := GetBookingAttributeService().Create(admin, in)
		if bErr == nil || bErr.Kind != BookingErrorInvalid {
			t.Fatalf("expected invalid error for %+v, got %v", in, bErr)
		}
	}
}

func TestBookingAttributeServiceLimit(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	for i := 0; i < MaxBookingAttributes; i++ {
		createTestServiceBookingAttribute(t, admin, "Attr", BookingAttributeTypeText, false)
	}
	_, bErr := GetBookingAttributeService().Create(admin, &BookingAttributeInput{Label: "Too many", Type: BookingAttributeTypeText})
	if bErr == nil || bErr.Kind != BookingErrorInvalid || bErr.Code != BookingCodeAttributeLimitReached {
		t.Fatalf("expected limit error, got %v", bErr)
	}
}

func TestBookingAttributeServiceForbidden(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	user := CreateTestUserInOrg(org)
	e := createTestServiceBookingAttribute(t, admin, "Company", BookingAttributeTypeText, false)

	if _, bErr := GetBookingAttributeService().List(user); bErr == nil || bErr.Kind != BookingErrorForbidden {
		t.Fatalf("expected forbidden, got %v", bErr)
	}
	if _, bErr := GetBookingAttributeService().Create(user, &BookingAttributeInput{Label: "X", Type: BookingAttributeTypeText}); bErr == nil || bErr.Kind != BookingErrorForbidden {
		t.Fatalf("expected forbidden, got %v", bErr)
	}
	if bErr := GetBookingAttributeService().Update(user, e.ID, &BookingAttributeInput{Label: "X", Type: BookingAttributeTypeText}); bErr == nil || bErr.Kind != BookingErrorForbidden {
		t.Fatalf("expected forbidden, got %v", bErr)
	}
	if bErr := GetBookingAttributeService().Delete(user, e.ID); bErr == nil || bErr.Kind != BookingErrorForbidden {
		t.Fatalf("expected forbidden, got %v", bErr)
	}
}

func TestBookingAttributeServiceForeignOrg(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	org2 := CreateTestOrg("test2.com")
	admin := CreateTestUserOrgAdmin(org)
	admin2 := CreateTestUserOrgAdmin(org2)
	e := createTestServiceBookingAttribute(t, admin, "Company", BookingAttributeTypeText, false)

	if bErr := GetBookingAttributeService().Update(admin2, e.ID, &BookingAttributeInput{Label: "X", Type: BookingAttributeTypeText}); bErr == nil || bErr.Kind != BookingErrorNotFound {
		t.Fatalf("expected not found, got %v", bErr)
	}
	if bErr := GetBookingAttributeService().Delete(admin2, e.ID); bErr == nil || bErr.Kind != BookingErrorNotFound {
		t.Fatalf("expected not found, got %v", bErr)
	}
	list, _ := GetBookingAttributeService().List(admin2)
	CheckTestInt(t, 0, len(list))
	list, _ = GetBookingAttributeService().List(admin)
	CheckTestString(t, "Company", list[0].Label)
}

func TestBookingAttributeServiceValidateValues(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	org2 := CreateTestOrg("test2.com")
	admin := CreateTestUserOrgAdmin(org)
	admin2 := CreateTestUserOrgAdmin(org2)
	company := createTestServiceBookingAttribute(t, admin, "Company", BookingAttributeTypeText, true)
	notes := createTestServiceBookingAttribute(t, admin, "Notes", BookingAttributeTypeTextMultiline, false)
	persons := createTestServiceBookingAttribute(t, admin, "Persons", BookingAttributeTypeNumber, false)
	foreign := createTestServiceBookingAttribute(t, admin2, "Foreign", BookingAttributeTypeText, false)

	res, bErr := GetBookingAttributeService().ValidateValues(org.ID, map[string]string{
		company.ID: " ACME ",
		notes.ID:   "Line 1\r\nLine 2\n",
		persons.ID: "-2.5",
	})
	if bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}
	CheckTestString(t, "ACME", res[company.ID])
	CheckTestString(t, "Line 1\nLine 2", res[notes.ID])
	CheckTestString(t, "-2.5", res[persons.ID])

	res, bErr = GetBookingAttributeService().ValidateValues(org.ID, map[string]string{
		company.ID: "ACME",
		notes.ID:   "  ",
	})
	if bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}
	CheckTestInt(t, 1, len(res))

	for _, values := range []map[string]string{
		nil,
		{company.ID: "   "},
		{company.ID: "ACME", foreign.ID: "x"},
		{company.ID: "ACME", "00000000-0000-0000-0000-000000000000": "x"},
		{company.ID: "AC\nME"},
		{company.ID: strings.Repeat("a", MaxBookingAttributeTextLength+1)},
		{company.ID: "ACME", notes.ID: strings.Repeat("a", MaxBookingAttributeMultilineTextLength+1)},
		{company.ID: "ACME", persons.ID: "abc"},
		{company.ID: "ACME", persons.ID: "1e5"},
		{company.ID: "ACME", persons.ID: "0x10"},
		{company.ID: "ACME", persons.ID: "Inf"},
		{company.ID: "ACME", persons.ID: "1,5"},
	} {
		_, bErr := GetBookingAttributeService().ValidateValues(org.ID, values)
		if bErr == nil || bErr.Code != BookingCodeInvalidAttributes {
			t.Fatalf("expected invalid attributes error for %v, got %v", values, bErr)
		}
	}
}

func TestBookingAttributeServiceValidateValuesNoAttributes(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	res, bErr := GetBookingAttributeService().ValidateValues(org.ID, nil)
	if bErr != nil {
		t.Fatalf("unexpected error %v", bErr)
	}
	CheckTestInt(t, 0, len(res))
}
