package service

import (
	"sync"

	. "github.com/seatsurfing/seatsurfing/server/repository"
)

// GroupService holds group-related business rules shared by the REST API
// (package router) and the plugin host API.
type GroupService struct{}

var groupService *GroupService
var groupServiceOnce sync.Once

func GetGroupService() *GroupService {
	groupServiceOnce.Do(func() {
		groupService = &GroupService{}
	})
	return groupService
}

// ValidateGroupsInOrg reports whether every group ID references an existing
// group of the given organization. An empty list is valid. Duplicate IDs are
// rejected: callers persist the list as (entity, group) rows, so a duplicate
// would violate the primary key after the entity has already been written.
// The error is only set on unexpected (database) failures.
//
// Any group list that is persisted as a reference on an organization's entity
// (allowed bookers, approvers) must pass this check, otherwise another
// tenant's group could be attached and later resolved by name.
func (s *GroupService) ValidateGroupsInOrg(organizationID string, groupIDs []string) (bool, error) {
	if len(groupIDs) == 0 {
		return true, nil
	}
	seen := make(map[string]struct{}, len(groupIDs))
	for _, id := range groupIDs {
		if _, ok := seen[id]; ok {
			return false, nil
		}
		seen[id] = struct{}{}
	}
	return GetGroupRepository().GroupsExistAndBelongToOrg(organizationID, groupIDs)
}
