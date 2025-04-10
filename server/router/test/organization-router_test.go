package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestOrganizationsEmptyResult(t *testing.T) {
	ClearTestDB()
	user := CreateTestUserSuperAdmin()
	loginResponse := LoginTestUser(user.ID)

	req := NewHTTPRequest("GET", "/organization/", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []string
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 1 {
		t.Fatalf("Expected array with one element (auto-created)")
	}
}

func TestOrganizationsForbidden(t *testing.T) {
	ClearTestDB()
	loginResponse := CreateLoginTestUser()
	org := CreateTestOrg("testing.com")

	req := NewHTTPRequest("GET", "/organization/", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	req = NewHTTPRequest("POST", "/organization/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	req = NewHTTPRequest("DELETE", "/organization/"+org.ID, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	req = NewHTTPRequest("PUT", "/organization/"+org.ID, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	req = NewHTTPRequest("GET", "/organization/"+org.ID, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestOrganizationsCRUD(t *testing.T) {
	ClearTestDB()
	user := CreateTestUserSuperAdmin()
	loginResponse := LoginTestUser(user.ID)

	// 1. Create
	payload := `{
		"name": "Some Company Ltd.",
		"firstname": "Foo",
		"lastname": "Bar",
		"email": "foo@seatsurfing.app",
		"language": "de"
	}`
	req := NewHTTPRequest("POST", "/organization/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-Id")

	// 2. Read
	req = NewHTTPRequest("GET", "/organization/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetOrganizationResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestString(t, "Some Company Ltd.", resBody.Name)
	CheckTestString(t, "Foo", resBody.Firstname)
	CheckTestString(t, "Bar", resBody.Lastname)
	CheckTestString(t, "foo@seatsurfing.app", resBody.Email)
	CheckTestString(t, "de", resBody.Language)

	// 3. Update
	payload = `{
		"name": "Some Company 2 Ltd.",
		"firstname": "Foo 2",
		"lastname": "Bar 2",
		"email": "foo2@seatsurfing.app",
		"language": "us"
	}`
	req = NewHTTPRequest("PUT", "/organization/"+id, loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Read
	req = NewHTTPRequest("GET", "/organization/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody2 *GetOrganizationResponse
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	CheckTestString(t, "Some Company 2 Ltd.", resBody2.Name)
	CheckTestString(t, "Foo 2", resBody2.Firstname)
	CheckTestString(t, "Bar 2", resBody2.Lastname)
	CheckTestString(t, "foo2@seatsurfing.app", resBody2.Email)
	CheckTestString(t, "us", resBody2.Language)

	// 4. Delete
	req = NewHTTPRequest("DELETE", "/organization/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Read
	req = NewHTTPRequest("GET", "/organization/"+id, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestOrganizationsGetByDomain(t *testing.T) {
	ClearTestDB()
	user := CreateTestUserSuperAdmin()
	loginResponse := LoginTestUser(user.ID)

	// Create organization
	payload := `{
		"name": "Some Company Ltd.",
		"firstname": "Foo",
		"lastname": "Bar",
		"email": "foo@seatsurfing.app",
		"language": "de"
	}`
	req := NewHTTPRequest("POST", "/organization/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-Id")
	GetSettingsRepository().Set(id, SettingFeatureCustomDomains.Name, "1")

	// Add domain 1
	req = NewHTTPRequest("POST", "/organization/"+id+"/domain/test1.com", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	// Add domain 2
	req = NewHTTPRequest("POST", "/organization/"+id+"/domain/test2.com", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	// Get by domain 1 (created by super admin, so it's verified from the start)
	req = NewHTTPRequest("GET", "/organization/domain/test1.com", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	// Verify both domains
	org, _ := GetOrganizationRepository().GetOne(id)
	GetOrganizationRepository().ActivateDomain(org, "test1.com")
	GetOrganizationRepository().ActivateDomain(org, "test2.com")

	// Get by domain 1
	req = NewHTTPRequest("GET", "/organization/domain/test1.com", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody *GetOrganizationResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestString(t, "Some Company Ltd.", resBody.Name)

	// Get by domain 2
	req = NewHTTPRequest("GET", "/organization/domain/test2.com", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody2 *GetOrganizationResponse
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	CheckTestString(t, "Some Company Ltd.", resBody.Name)

	// Get by unknown domain
	req = NewHTTPRequest("GET", "/organization/domain/test3.com", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestOrganizationsDomainsCRUD(t *testing.T) {
	ClearTestDB()
	user := CreateTestUserSuperAdmin()
	loginResponse := LoginTestUser(user.ID)

	// Create organization
	payload := `{
		"name": "Some Company Ltd.",
		"firstname": "Foo",
		"lastname": "Bar",
		"email": "foo@seatsurfing.app",
		"language": "de"
	}`
	req := NewHTTPRequest("POST", "/organization/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-Id")
	GetSettingsRepository().Set(id, SettingFeatureCustomDomains.Name, "1")

	// Add domain 1
	req = NewHTTPRequest("POST", "/organization/"+id+"/domain/test1.com", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	// Add domain 2
	req = NewHTTPRequest("POST", "/organization/"+id+"/domain/test2.com", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	// Add domain 3
	req = NewHTTPRequest("POST", "/organization/"+id+"/domain/abc.com", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	// Get domain list
	req = NewHTTPRequest("GET", "/organization/"+id+"/domain/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []*GetDomainResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 3 {
		t.Fatalf("Expected array with 3 elements, got %d", len(resBody))
	}
	CheckTestString(t, "abc.com", resBody[0].DomainName)
	CheckTestString(t, "test1.com", resBody[1].DomainName)
	CheckTestString(t, "test2.com", resBody[2].DomainName)
	CheckTestBool(t, true, resBody[0].Active)
	CheckTestBool(t, true, resBody[1].Active)
	CheckTestBool(t, true, resBody[2].Active)

	// Remove 2
	req = NewHTTPRequest("DELETE", "/organization/"+id+"/domain/test2.com", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Get domain list
	req = NewHTTPRequest("GET", "/organization/"+id+"/domain/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody2 []*GetDomainResponse
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	if len(resBody2) != 2 {
		t.Fatalf("Expected array with 2 elements")
	}
	CheckTestString(t, "abc.com", resBody[0].DomainName)
	CheckTestString(t, "test1.com", resBody[1].DomainName)
	CheckTestBool(t, true, resBody[0].Active)
	CheckTestBool(t, true, resBody[1].Active)
}

func TestOrganizationsVerifyDNS(t *testing.T) {
	ClearTestDB()
	user := CreateTestUserSuperAdmin()
	loginResponse := LoginTestUser(user.ID)

	// Create organization
	payload := `{
		"name": "Some Company Ltd.",
		"firstname": "Foo",
		"lastname": "Bar",
		"email": "foo@seatsurfing.app",
		"language": "de"
	}`
	req := NewHTTPRequest("POST", "/organization/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id := res.Header().Get("X-Object-Id")
	GetSettingsRepository().Set(id, SettingFeatureCustomDomains.Name, "1")

	org, _ := GetOrganizationRepository().GetOne(id)
	adminUser := CreateTestUserOrgAdmin(org)
	adminLoginResponse := LoginTestUser(adminUser.ID)

	// Add domain
	req = NewHTTPRequest("POST", "/organization/"+id+"/domain/seatsurfing-testcase.virtualzone.de", adminLoginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	// Fake verify token
	GetDatabase().DB().Exec("UPDATE organizations_domains "+
		"SET verify_token = '65e51a4b-339f-4b24-b376-f9d866057b38' "+
		"WHERE domain = LOWER($1) AND organization_id = $2",
		"seatsurfing-testcase.virtualzone.de", id)

	// Verify domain
	req = NewHTTPRequest("POST", "/organization/"+id+"/domain/seatsurfing-testcase.virtualzone.de/verify", adminLoginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Get domain list
	req = NewHTTPRequest("GET", "/organization/"+id+"/domain/", adminLoginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []*GetDomainResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 1 {
		t.Fatalf("Expected array with 1 elements, got %d", len(resBody))
	}
	CheckTestString(t, "seatsurfing-testcase.virtualzone.de", resBody[0].DomainName)
	CheckTestBool(t, true, resBody[0].Active)
}

func TestOrganizationsAddDomainConflict(t *testing.T) {
	ClearTestDB()
	user := CreateTestUserSuperAdmin()
	loginResponse := LoginTestUser(user.ID)

	// Create organization 1
	payload := `{
		"name": "Some Company Ltd.",
		"firstname": "Foo",
		"lastname": "Bar",
		"email": "foo@seatsurfing.app",
		"language": "de"
	}`
	req := NewHTTPRequest("POST", "/organization/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id1 := res.Header().Get("X-Object-Id")
	GetSettingsRepository().Set(id1, SettingFeatureCustomDomains.Name, "1")

	// Create organization 2
	payload = `{
		"name": "Some Company 2 Ltd.",
		"firstname": "Foo",
		"lastname": "Bar",
		"email": "foo@seatsurfing.app",
		"language": "de"
	}`
	req = NewHTTPRequest("POST", "/organization/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id2 := res.Header().Get("X-Object-Id")
	GetSettingsRepository().Set(id2, SettingFeatureCustomDomains.Name, "1")

	// Add domain to org 1 and activate it
	req = NewHTTPRequest("POST", "/organization/"+id1+"/domain/test1.com", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	org1, _ := GetOrganizationRepository().GetOne(id1)
	GetOrganizationRepository().ActivateDomain(org1, "test1.com")

	// Try to add same domain to org 2
	req = NewHTTPRequest("POST", "/organization/"+id2+"/domain/test1.com", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusConflict, res.Code)
}

func TestOrganizationsAddDomainNoConflictBecauseInactive(t *testing.T) {
	ClearTestDB()
	user := CreateTestUserSuperAdmin()
	loginResponse := LoginTestUser(user.ID)

	// Create organization 1
	payload := `{
		"name": "Some Company 1 Ltd.",
		"firstname": "Foo",
		"lastname": "Bar",
		"email": "foo@seatsurfing.app",
		"language": "de"
	}`
	req := NewHTTPRequest("POST", "/organization/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id1 := res.Header().Get("X-Object-Id")
	GetSettingsRepository().Set(id1, SettingFeatureCustomDomains.Name, "1")

	// Create organization 2
	payload = `{
		"name": "Some Company 2 Ltd.",
		"firstname": "Foo",
		"lastname": "Bar",
		"email": "foo@seatsurfing.app",
		"language": "de"
	}`
	req = NewHTTPRequest("POST", "/organization/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id2 := res.Header().Get("X-Object-Id")
	GetSettingsRepository().Set(id2, SettingFeatureCustomDomains.Name, "1")

	org1, _ := GetOrganizationRepository().GetOne(id1)
	adminUser1 := CreateTestUserOrgAdmin(org1)
	adminLoginResponse1 := LoginTestUser(adminUser1.ID)

	org2, _ := GetOrganizationRepository().GetOne(id2)
	adminUser2 := CreateTestUserOrgAdmin(org2)
	adminLoginResponse2 := LoginTestUser(adminUser2.ID)

	// Add domain to org 1
	req = NewHTTPRequest("POST", "/organization/"+id1+"/domain/test1.com", adminLoginResponse1.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	// Add same domain to org 2
	req = NewHTTPRequest("POST", "/organization/"+id2+"/domain/test1.com", adminLoginResponse2.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
}

func TestOrganizationsAddDomainActivateConflicting(t *testing.T) {
	ClearTestDB()
	user := CreateTestUserSuperAdmin()
	loginResponse := LoginTestUser(user.ID)

	// Create organization 1
	payload := `{
		"name": "Some Company 1 Ltd.",
		"firstname": "Foo",
		"lastname": "Bar",
		"email": "foo@seatsurfing.app",
		"language": "de"
	}`
	req := NewHTTPRequest("POST", "/organization/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id1 := res.Header().Get("X-Object-Id")
	GetSettingsRepository().Set(id1, SettingFeatureCustomDomains.Name, "1")

	// Create organization 2
	payload = `{
		"name": "Some Company 2 Ltd.",
		"firstname": "Foo",
		"lastname": "Bar",
		"email": "foo@seatsurfing.app",
		"language": "de"
	}`
	req = NewHTTPRequest("POST", "/organization/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)
	id2 := res.Header().Get("X-Object-Id")
	GetSettingsRepository().Set(id2, SettingFeatureCustomDomains.Name, "1")

	org1, _ := GetOrganizationRepository().GetOne(id1)
	adminUser1 := CreateTestUserOrgAdmin(org1)
	adminLoginResponse1 := LoginTestUser(adminUser1.ID)

	org2, _ := GetOrganizationRepository().GetOne(id2)
	adminUser2 := CreateTestUserOrgAdmin(org2)
	adminLoginResponse2 := LoginTestUser(adminUser2.ID)

	// Add domain to org 1
	req = NewHTTPRequest("POST", "/organization/"+id1+"/domain/seatsurfing-testcase.virtualzone.de", adminLoginResponse1.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	// Add same domain to org 2
	req = NewHTTPRequest("POST", "/organization/"+id2+"/domain/seatsurfing-testcase.virtualzone.de", adminLoginResponse2.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	// Fake verify tokens
	_, err := GetDatabase().DB().Exec("UPDATE organizations_domains "+
		"SET verify_token = '65e51a4b-339f-4b24-b376-f9d866057b38' "+
		"WHERE domain = LOWER($1) AND organization_id IN ($2, $3)",
		"seatsurfing-testcase.virtualzone.de", id1, id2)
	if err != nil {
		t.Fatal(err)
	}

	// Activate domain in org 1
	req = NewHTTPRequest("POST", "/organization/"+id1+"/domain/seatsurfing-testcase.virtualzone.de/verify", adminLoginResponse1.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Try to activate same domain in org 2
	req = NewHTTPRequest("POST", "/organization/"+id2+"/domain/seatsurfing-testcase.virtualzone.de/verify", adminLoginResponse2.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusConflict, res.Code)
}

func TestOrganizationsDelete(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	req := NewHTTPRequest("DELETE", "/organization/"+org.ID, loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Verify
	users, _ := GetUserRepository().GetAll(org.ID, 100, 0)
	CheckTestInt(t, 0, len(users))
}

func TestOrganizationsPrimaryDomain(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test1.com")
	GetSettingsRepository().Set(org.ID, SettingFeatureCustomDomains.Name, "1")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	// Add domain 2
	req := NewHTTPRequest("POST", "/organization/"+org.ID+"/domain/test2.com", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	// Add domain 3
	req = NewHTTPRequest("POST", "/organization/"+org.ID+"/domain/test3.com", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusCreated, res.Code)

	// Get domain list
	req = NewHTTPRequest("GET", "/organization/"+org.ID+"/domain/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []*GetDomainResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 3 {
		t.Fatalf("Expected array with 3 elements, got %d", len(resBody))
	}
	CheckTestString(t, "test1.com", resBody[0].DomainName)
	CheckTestString(t, "test2.com", resBody[1].DomainName)
	CheckTestString(t, "test3.com", resBody[2].DomainName)
	CheckTestBool(t, true, resBody[0].Primary)
	CheckTestBool(t, false, resBody[1].Primary)
	CheckTestBool(t, false, resBody[2].Primary)

	// Set domain 2 as primary
	req = NewHTTPRequest("POST", "/organization/"+org.ID+"/domain/test2.com/primary", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Get domain list
	req = NewHTTPRequest("GET", "/organization/"+org.ID+"/domain/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	resBody = nil
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 3 {
		t.Fatalf("Expected array with 3 elements, got %d", len(resBody))
	}
	CheckTestString(t, "test1.com", resBody[0].DomainName)
	CheckTestString(t, "test2.com", resBody[1].DomainName)
	CheckTestString(t, "test3.com", resBody[2].DomainName)
	CheckTestBool(t, false, resBody[0].Primary)
	CheckTestBool(t, true, resBody[1].Primary)
	CheckTestBool(t, false, resBody[2].Primary)

	// Delete domain 2
	req = NewHTTPRequest("DELETE", "/organization/"+org.ID+"/domain/test2.com", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	// Get domain list
	req = NewHTTPRequest("GET", "/organization/"+org.ID+"/domain/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	resBody = nil
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) != 2 {
		t.Fatalf("Expected array with 2 elements, got %d", len(resBody))
	}
	CheckTestString(t, "test1.com", resBody[0].DomainName)
	CheckTestString(t, "test3.com", resBody[1].DomainName)
	CheckTestBool(t, true, resBody[0].Primary)
	CheckTestBool(t, false, resBody[1].Primary)
}
