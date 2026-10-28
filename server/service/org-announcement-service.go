package service

import (
	"sync"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
)

// OrgAnnouncementService decides who may see an organization's announcement.
// Text and required permission are maintained by an external process.
type OrgAnnouncementService struct{}

var orgAnnouncementService *OrgAnnouncementService
var orgAnnouncementServiceOnce sync.Once

func GetOrgAnnouncementService() *OrgAnnouncementService {
	orgAnnouncementServiceOnce.Do(func() {
		orgAnnouncementService = &OrgAnnouncementService{}
	})
	return orgAnnouncementService
}

// RequiredPermission returns the permission a user must hold at admin level to
// see the announcement, falling back to PermissionOrgSettings if none is set.
func (s *OrgAnnouncementService) RequiredPermission(organizationID string) Permission {
	value, _ := GetSettingsRepository().Get(organizationID, SettingOrgAnnouncementPermission.Name)
	if value == "" {
		return PermissionOrgSettings
	}
	return Permission(value)
}

func (s *OrgAnnouncementService) CanView(user *User) bool {
	if user == nil {
		return false
	}
	return HasPermission(user, user.OrganizationID, s.RequiredPermission(user.OrganizationID), PermissionLevelAdmin)
}
