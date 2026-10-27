package test

import (
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/service"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestOrgAnnouncementServiceDefaultPermission(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	svc := GetOrgAnnouncementService()
	CheckTestString(t, string(PermissionOrgSettings), string(svc.RequiredPermission(org.ID)))
	CheckTestBool(t, true, svc.CanView(CreateTestUserOrgAdmin(org)))
	CheckTestBool(t, false, svc.CanView(CreateTestUserInOrg(org)))
	CheckTestBool(t, false, svc.CanView(CreateTestUserWithPermissions(org, map[Permission]PermissionLevel{PermissionBookings: PermissionLevelAdmin})))
	CheckTestBool(t, false, svc.CanView(nil))
}

func TestOrgAnnouncementServiceCustomPermission(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org.ID, SettingOrgAnnouncementPermission.Name, string(PermissionBookings))
	svc := GetOrgAnnouncementService()
	CheckTestString(t, string(PermissionBookings), string(svc.RequiredPermission(org.ID)))
	CheckTestBool(t, true, svc.CanView(CreateTestUserWithPermissions(org, map[Permission]PermissionLevel{PermissionBookings: PermissionLevelAdmin})))
	CheckTestBool(t, false, svc.CanView(CreateTestUserWithPermissions(org, map[Permission]PermissionLevel{PermissionBookings: PermissionLevelRead})))
	CheckTestBool(t, false, svc.CanView(CreateTestUserWithPermissions(org, map[Permission]PermissionLevel{PermissionOrgSettings: PermissionLevelAdmin})))
	CheckTestBool(t, false, svc.CanView(CreateTestUserInOrg(org)))
}

func TestOrgAnnouncementServiceUnknownPermission(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	GetSettingsRepository().Set(org.ID, SettingOrgAnnouncementPermission.Name, "does_not_exist")
	CheckTestBool(t, false, GetOrgAnnouncementService().CanView(CreateTestUserOrgAdmin(org)))
}
