package test

import (
	"bytes"
	"net/http"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

// The "limited" level on groups exists for issue #2470: a room administrator
// who maintains who is in a group, without being able to create, rename or
// delete the groups themselves.
func TestGroupsWriteLevelManagesMembershipOnly(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	CreateTestUserOrgAdmin(org)
	member := CreateTestUserInOrg(org)

	group := &Group{OrganizationID: org.ID, Name: "Lab Access"}
	if err := GetGroupRepository().Create(group); err != nil {
		t.Fatal(err)
	}

	roomAdmin := CreateTestUserWithPermissions(org, map[Permission]PermissionLevel{
		PermissionGroups: PermissionLevelWrite,
	})
	login := LoginTestUser(roomAdmin.ID)

	// Reading is included in the level.
	req := NewHTTPRequest("GET", "/group/", login.UserID, nil)
	CheckTestResponseCode(t, http.StatusOK, ExecuteTestRequest(req).Code)
	req = NewHTTPRequest("GET", "/group/"+group.ID, login.UserID, nil)
	CheckTestResponseCode(t, http.StatusOK, ExecuteTestRequest(req).Code)
	req = NewHTTPRequest("GET", "/group/"+group.ID+"/member", login.UserID, nil)
	CheckTestResponseCode(t, http.StatusOK, ExecuteTestRequest(req).Code)

	// Membership is what the level is for.
	payload := `["` + member.ID + `"]`
	req = NewHTTPRequest("PUT", "/group/"+group.ID+"/member", login.UserID, bytes.NewBufferString(payload))
	CheckTestResponseCode(t, http.StatusNoContent, ExecuteTestRequest(req).Code)

	members, err := GetGroupRepository().GetMemberUserIDs(group)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 1, len(members))

	req = NewHTTPRequest("POST", "/group/"+group.ID+"/member/remove", login.UserID, bytes.NewBufferString(payload))
	CheckTestResponseCode(t, http.StatusNoContent, ExecuteTestRequest(req).Code)
	members, _ = GetGroupRepository().GetMemberUserIDs(group)
	CheckTestInt(t, 0, len(members))

	// The groups themselves are not theirs to change.
	req = NewHTTPRequest("POST", "/group/", login.UserID, bytes.NewBufferString(`{"name": "Invented"}`))
	CheckTestResponseCode(t, http.StatusForbidden, ExecuteTestRequest(req).Code)
	req = NewHTTPRequest("PUT", "/group/"+group.ID, login.UserID, bytes.NewBufferString(`{"name": "Renamed"}`))
	CheckTestResponseCode(t, http.StatusForbidden, ExecuteTestRequest(req).Code)
	req = NewHTTPRequest("DELETE", "/group/"+group.ID, login.UserID, nil)
	CheckTestResponseCode(t, http.StatusForbidden, ExecuteTestRequest(req).Code)

	// The group survived all three attempts, under its original name.
	reloaded, err := GetGroupRepository().GetOne(group.ID)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestString(t, "Lab Access", reloaded.Name)

	// Nothing outside groups came along with it.
	req = NewHTTPRequest("GET", "/user/", login.UserID, nil)
	CheckTestResponseCode(t, http.StatusForbidden, ExecuteTestRequest(req).Code)
	req = NewHTTPRequest("GET", "/setting/", login.UserID, nil)
	CheckTestResponseCode(t, http.StatusOK, ExecuteTestRequest(req).Code) // public settings are baseline
}

// Read level stops short of membership changes.
func TestGroupsReadLevelCannotChangeMembership(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	CreateTestUserOrgAdmin(org)
	member := CreateTestUserInOrg(org)
	group := &Group{OrganizationID: org.ID, Name: "Lab Access"}
	if err := GetGroupRepository().Create(group); err != nil {
		t.Fatal(err)
	}
	reader := CreateTestUserWithPermissions(org, map[Permission]PermissionLevel{
		PermissionGroups: PermissionLevelRead,
	})
	login := LoginTestUser(reader.ID)

	req := NewHTTPRequest("GET", "/group/"+group.ID+"/member", login.UserID, nil)
	CheckTestResponseCode(t, http.StatusOK, ExecuteTestRequest(req).Code)

	payload := `["` + member.ID + `"]`
	req = NewHTTPRequest("PUT", "/group/"+group.ID+"/member", login.UserID, bytes.NewBufferString(payload))
	CheckTestResponseCode(t, http.StatusForbidden, ExecuteTestRequest(req).Code)
	req = NewHTTPRequest("POST", "/group/"+group.ID+"/member/remove", login.UserID, bytes.NewBufferString(payload))
	CheckTestResponseCode(t, http.StatusForbidden, ExecuteTestRequest(req).Code)
}
