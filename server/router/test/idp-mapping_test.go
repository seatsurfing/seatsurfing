package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"slices"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func createTestAuthProvider(org *Organization, groupsField string) *AuthProvider {
	p := &AuthProvider{
		OrganizationID:      org.ID,
		Name:                "Test IdP",
		ProviderType:        int(OAuth2),
		AuthURL:             "https://idp.example.com/auth",
		TokenURL:            "https://idp.example.com/token",
		UserInfoURL:         "https://idp.example.com/userinfo",
		UserInfoEmailField:  "email",
		UserInfoGroupsField: groupsField,
		ClientID:            "client",
		ClientSecret:        "secret",
	}
	if err := GetAuthProviderRepository().Create(p); err != nil {
		panic(err)
	}
	return p
}

func addMapping(provider *AuthProvider, claimValue, targetType, targetID string) {
	m := &AuthProviderMapping{
		AuthProviderID: provider.ID,
		ClaimValue:     claimValue,
		TargetType:     targetType,
		TargetID:       targetID,
	}
	if err := GetAuthProviderMappingRepository().Create(m); err != nil {
		panic(err)
	}
}

// Providers differ in how they report groups, and a Keycloak path must be
// matched in full: two groups can share a leaf name under different parents.
func TestExtractGroupsClaim(t *testing.T) {
	cases := []struct {
		name     string
		payload  string
		field    string
		expected []string
	}{
		{"array of strings", `{"email":"a@b.c","groups":["eng","ops"]}`, "groups", []string{"eng", "ops"}},
		{"single string", `{"email":"a@b.c","groups":"eng"}`, "groups", []string{"eng"}},
		{"nested paths kept whole", `{"email":"a@b.c","groups":["/parent/child","/other/child"]}`, "groups",
			[]string{"/parent/child", "/other/child"}},
		{"absent claim", `{"email":"a@b.c"}`, "groups", nil},
		{"no field configured", `{"email":"a@b.c","groups":["eng"]}`, "", nil},
		{"non-string entries ignored", `{"email":"a@b.c","groups":["eng",42,null]}`, "groups", []string{"eng"}},
		{"blank entries dropped", `{"email":"a@b.c","groups":["eng","  "]}`, "groups", []string{"eng"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(c.payload), &result); err != nil {
				t.Fatal(err)
			}
			info, err := ExtractUserInfoFields(result, "email", "", "", c.field)
			if err != nil {
				t.Fatal(err)
			}
			if len(info.Groups) != len(c.expected) {
				t.Fatalf("expected %v, got %v", c.expected, info.Groups)
			}
			for i := range c.expected {
				CheckTestString(t, c.expected[i], info.Groups[i])
			}
		})
	}
}

func TestIdPReconciliationAssignsRolesAndGroups(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	CreateTestUserOrgAdmin(org)
	user := CreateTestUserInOrg(org)
	provider := createTestAuthProvider(org, "groups")

	role := CreateTestRole(org, "Group Manager", map[Permission]PermissionLevel{PermissionGroups: PermissionLevelAdmin})
	group := &Group{OrganizationID: org.ID, Name: "Engineering"}
	if err := GetGroupRepository().Create(group); err != nil {
		t.Fatal(err)
	}
	addMapping(provider, "/eng", AuthProviderMappingTargetRole, role.ID)
	addMapping(provider, "/eng", AuthProviderMappingTargetGroup, group.ID)

	ReconcileUserFromIdP(user, provider, []string{"/eng"})

	roleIDs, _ := GetUserRoleRepository().GetRoleIDsForUser(user.ID)
	CheckTestInt(t, 1, len(roleIDs))
	CheckTestString(t, role.ID, roleIDs[0])
	groups, _ := GetGroupRepository().GetAllWhereUserIsMember(user.ID)
	CheckTestInt(t, 1, len(groups))
	CheckTestString(t, group.ID, groups[0].ID)

	// A claim value with no mapping grants nothing.
	other := CreateTestUserInOrg(org)
	ReconcileUserFromIdP(other, provider, []string{"/unmapped"})
	otherRoles, _ := GetUserRoleRepository().GetRoleIDsForUser(other.ID)
	CheckTestInt(t, 0, len(otherRoles))
}

// The provider is authoritative for what it granted, and only for that.
func TestIdPReconciliationRevokesOnlyItsOwnGrants(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	CreateTestUserOrgAdmin(org)
	user := CreateTestUserInOrg(org)
	provider := createTestAuthProvider(org, "groups")

	manualRole := CreateTestRole(org, "Manual", map[Permission]PermissionLevel{PermissionAuditLog: PermissionLevelRead})
	idpRole := CreateTestRole(org, "FromIdp", map[Permission]PermissionLevel{PermissionGroups: PermissionLevelRead})
	AssignTestRole(user, manualRole)
	addMapping(provider, "/eng", AuthProviderMappingTargetRole, idpRole.ID)

	manualGroup := &Group{OrganizationID: org.ID, Name: "Manual Group"}
	GetGroupRepository().Create(manualGroup)
	GetGroupRepository().AddMembers(manualGroup, []string{user.ID})
	idpGroup := &Group{OrganizationID: org.ID, Name: "Idp Group"}
	GetGroupRepository().Create(idpGroup)
	addMapping(provider, "/eng", AuthProviderMappingTargetGroup, idpGroup.ID)

	ReconcileUserFromIdP(user, provider, []string{"/eng"})
	roleIDs, _ := GetUserRoleRepository().GetRoleIDsForUser(user.ID)
	CheckTestInt(t, 2, len(roleIDs))

	// The user leaves the group at the provider. Their manual assignments must
	// survive; the provider's must go.
	ReconcileUserFromIdP(user, provider, []string{})

	roleIDs, _ = GetUserRoleRepository().GetRoleIDsForUser(user.ID)
	CheckTestInt(t, 1, len(roleIDs))
	CheckTestString(t, manualRole.ID, roleIDs[0])

	groups, _ := GetGroupRepository().GetAllWhereUserIsMember(user.ID)
	var names []string
	for _, g := range groups {
		names = append(names, g.Name)
	}
	if !slices.Contains(names, "Manual Group") {
		t.Fatalf("expected the manual group membership to survive, got %v", names)
	}
	if slices.Contains(names, "Idp Group") {
		t.Fatalf("expected the provider group membership to be revoked, got %v", names)
	}
}

// An organization must not lose its last administrator because somebody edited
// a group in the identity provider.
func TestIdPReconciliationWillNotStripLastAdmin(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	provider := createTestAuthProvider(org, "groups")

	adminRole, err := GetRoleRepository().GetByName(org.ID, RoleNameOrgAdmin)
	if err != nil {
		t.Fatal(err)
	}
	// The organization's only administrator holds the role from the provider.
	admin := CreateTestUserInOrg(org)
	addMapping(provider, "/admins", AuthProviderMappingTargetRole, adminRole.ID)
	ReconcileUserFromIdP(admin, provider, []string{"/admins"})
	CheckTestBool(t, true, OrgRetainsAdminWithout(org.ID))

	// Removing them from the group at the provider would empty the
	// organization, so the roles are left alone.
	ReconcileUserFromIdP(admin, provider, []string{})
	CheckTestBool(t, true, OrgRetainsAdminWithout(org.ID))
	roleIDs, _ := GetUserRoleRepository().GetRoleIDsForUser(admin.ID)
	CheckTestInt(t, 1, len(roleIDs))

	// With a second administrator present, the demotion goes through.
	CreateTestUserOrgAdmin(org)
	ReconcileUserFromIdP(admin, provider, []string{})
	roleIDs, _ = GetUserRoleRepository().GetRoleIDsForUser(admin.ID)
	CheckTestInt(t, 0, len(roleIDs))
	CheckTestBool(t, true, OrgRetainsAdminWithout(org.ID))
}

func TestAuthProviderMappingEndpoints(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(admin.ID)
	provider := createTestAuthProvider(org, "groups")
	role := CreateTestRole(org, "Mapped", map[Permission]PermissionLevel{PermissionGroups: PermissionLevelRead})

	payload := `{"mappings":[{"claimValue":"/eng","targetType":"role","targetId":"` + role.ID + `"}]}`
	req := NewHTTPRequest("PUT", "/auth-provider/"+provider.ID+"/mapping", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	req = NewHTTPRequest("GET", "/auth-provider/"+provider.ID+"/mapping", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []*AuthProviderMappingModel
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, 1, len(resBody))
	CheckTestString(t, "/eng", resBody[0].ClaimValue)

	// A role belonging to another organization is refused.
	otherOrg := CreateTestOrg("other.com")
	foreignRole := CreateTestRole(otherOrg, "Foreign", nil)
	payload = `{"mappings":[{"claimValue":"/eng","targetType":"role","targetId":"` + foreignRole.ID + `"}]}`
	req = NewHTTPRequest("PUT", "/auth-provider/"+provider.ID+"/mapping", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)

	// Deleting the role removes the mapping that pointed at it.
	if err := GetRoleRepository().Delete(role); err != nil {
		t.Fatal(err)
	}
	list, _ := GetAuthProviderMappingRepository().GetAll(provider.ID)
	CheckTestInt(t, 0, len(list))
}

// Nobody can use a mapping to hand out access they do not hold themselves.
func TestAuthProviderMappingNoEscalation(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	CreateTestUserOrgAdmin(org)
	limited := CreateTestUserWithPermissions(org, map[Permission]PermissionLevel{
		PermissionRoles:         PermissionLevelAdmin,
		PermissionAuthProviders: PermissionLevelAdmin,
	})
	loginResponse := LoginTestUser(limited.ID)
	provider := createTestAuthProvider(org, "groups")
	powerful := CreateTestRole(org, "Powerful", map[Permission]PermissionLevel{PermissionOrgSettings: PermissionLevelAdmin})

	payload := `{"mappings":[{"claimValue":"/eng","targetType":"role","targetId":"` + powerful.ID + `"}]}`
	req := NewHTTPRequest("PUT", "/auth-provider/"+provider.ID+"/mapping", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
}
