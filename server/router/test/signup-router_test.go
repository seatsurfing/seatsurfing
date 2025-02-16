package test

import (
	"bytes"
	"net/http"
	"regexp"
	"strings"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

func TestSignup(t *testing.T) {
	ClearTestDB()

	// Perform Signup
	payload := `{
		"firstname": "",
		"lastname": "",
		"email": "foo@bar.com", 
		"organization": "Test Org", 
		"domain": "testorg", 
		"contactFirstname": "Foo", 
		"contactLastname": "Bar", 
		"password": "12345678", 
		"language": "de",
		"acceptTerms": true
		}`
	req := NewHTTPRequest("POST", "/signup/", "", bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)
	CheckTestBool(t, true, strings.Contains(SendMailMockContent, "Hallo Foo Bar,"))
	CheckTestBool(t, true, strings.Contains(SendMailMockContent, "To: Foo Bar <foo@bar.com>"))

	// Extract Confirm ID from email
	rx := regexp.MustCompile(`/confirm/(.*)?\n`)
	confirmTokens := rx.FindStringSubmatch(SendMailMockContent)
	CheckTestInt(t, 2, len(confirmTokens))
	confirmID := confirmTokens[1]

	// Confirm signup (Double Opt In)
	req = NewHTTPRequest("POST", "/signup/confirm/"+confirmID, "", nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)
	CheckTestBool(t, true, strings.Contains(SendMailMockContent, "Hallo Foo Bar,"))
	CheckTestBool(t, true, strings.Contains(SendMailMockContent, "To: foo@bar.com"))
	CheckTestBool(t, true, strings.Contains(SendMailMockContent, "admin@testorg.on.seatsurfing.local"))

	// Check if login works
	payload = `{"email": "admin@testorg.on.seatsurfing.local", "password": "12345678"}`
	req = NewHTTPRequest("POST", "/auth/login", "", bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	// Verify signup confirm is not possible anymore
	req = NewHTTPRequest("POST", "/signup/confirm/"+confirmID, "", nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestSignupLanguageEN(t *testing.T) {
	ClearTestDB()

	// Perform Signup
	payload := `{
		"firstname": "",
		"lastname": "",
		"email": "foo@bar.com", 
		"organization": "Test Org", 
		"domain": "testorg", 
		"contactFirstname": "Foo", 
		"contactLastname": "Bar", 
		"password": "12345678", 
		"language": "en",
		"acceptTerms": true
		}`
	req := NewHTTPRequest("POST", "/signup/", "", bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)
	CheckTestBool(t, true, strings.Contains(SendMailMockContent, "Hello Foo Bar,"))
	CheckTestBool(t, true, strings.Contains(SendMailMockContent, "To: Foo Bar <foo@bar.com>"))
}

func TestSignupNotAcceptTerms(t *testing.T) {
	ClearTestDB()

	// Perform Signup
	payload := `{
		"firstname": "",
		"lastname": "",
		"email": "foo@bar.com", 
		"organization": "Test Org", 
		"domain": "testorg", 
		"contactFirstname": "Foo", 
		"contactLastname": "Bar", 
		"password": "12345678", 
		"language": "de",
		"acceptTerms": false
		}`
	req := NewHTTPRequest("POST", "/signup/", "", bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
}

func TestSignupInvalidEmail(t *testing.T) {
	ClearTestDB()

	// Perform Signup
	payload := `{
		"firstname": "",
		"lastname": "",
		"email": "foobar.com", 
		"organization": "Test Org", 
		"domain": "testorg", 
		"contactFirstname": "Foo", 
		"contactLastname": "Bar", 
		"password": "12345678", 
		"language": "de",
		"acceptTerms": true
		}`
	req := NewHTTPRequest("POST", "/signup/", "", bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
}

func TestSignupShortPassword(t *testing.T) {
	ClearTestDB()

	// Perform Signup
	payload := `{
		"firstname": "",
		"lastname": "",
		"email": "foo@bar.com", 
		"organization": "Test Org", 
		"domain": "testorg", 
		"contactFirstname": "Foo", 
		"contactLastname": "Bar", 
		"password": "123456", 
		"language": "de",
		"acceptTerms": true
		}`
	req := NewHTTPRequest("POST", "/signup/", "", bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
}

func TestSignupDomainConflict(t *testing.T) {
	ClearTestDB()

	CreateTestOrg("testorg.on.seatsurfing.local")

	// Perform Signup
	payload := `{
		"firstname": "",
		"lastname": "",
		"email": "foo@bar.com", 
		"organization": "Test Org", 
		"domain": "testorg", 
		"contactFirstname": "Foo", 
		"contactLastname": "Bar", 
		"password": "12345678", 
		"language": "de",
		"acceptTerms": true
		}`
	req := NewHTTPRequest("POST", "/signup/", "", bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusConflict, res.Code)
}

func TestSignupEmailConflictSignup(t *testing.T) {
	ClearTestDB()

	// Perform Signup
	payload := `{
			"firstname": "",
			"lastname": "",
			"email": "foo@bar.com",
			"organization": "Test Org",
			"domain": "testorg",
			"contactFirstname": "Foo",
			"contactLastname": "Bar",
			"password": "12345678",
			"language": "de",
			"acceptTerms": true
			}`
	req := NewHTTPRequest("POST", "/signup/", "", bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	payload = `{
			"firstname": "",
			"lastname": "",
			"email": "foo@bar.com",
			"organization": "Test Org",
			"domain": "testorg2",
			"contactFirstname": "Foo",
			"contactLastname": "Bar",
			"password": "12345678",
			"language": "de",
			"acceptTerms": true
			}`
	req = NewHTTPRequest("POST", "/signup/", "", bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusConflict, res.Code)
}

func TestSignupEmailConflictExistingOrg(t *testing.T) {
	ClearTestDB()

	org := CreateTestOrg("testorg.on.seatsurfing.app")
	org.ContactEmail = "foo@bar.com"
	GetOrganizationRepository().Update(org)

	payload := `{
			"firstname": "",
			"lastname": "",
			"email": "foo@bar.com",
			"organization": "Test Org",
			"domain": "testorg2",
			"contactFirstname": "Foo",
			"contactLastname": "Bar",
			"password": "12345678",
			"language": "de",
			"acceptTerms": true
			}`
	req := NewHTTPRequest("POST", "/signup/", "", bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusConflict, res.Code)
}

func TestSignupInvalidLanguage(t *testing.T) {
	ClearTestDB()

	payload := `{
			"firstname": "",
			"lastname": "",
			"email": "foo@bar.com",
			"organization": "Test Org",
			"domain": "testorg2",
			"contactFirstname": "Foo",
			"contactLastname": "Bar",
			"password": "12345678",
			"language": "tr",
			"acceptTerms": true
			}`
	req := NewHTTPRequest("POST", "/signup/", "", bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
}
