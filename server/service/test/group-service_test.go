package test

import (
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/service"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestGroupServiceValidateGroupsInOrg(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	foreignOrg := CreateTestOrg("foreign.com")
	ownGroup1 := CreateTestGroup(org, nil)
	ownGroup2 := CreateTestGroup(org, nil)
	foreignGroup := CreateTestGroup(foreignOrg, nil)
	s := GetGroupService()

	check := func(expected bool, groupIDs []string) {
		t.Helper()
		ok, err := s.ValidateGroupsInOrg(org.ID, groupIDs)
		if err != nil {
			t.Fatal(err)
		}
		CheckTestBool(t, expected, ok)
	}

	// Positive: empty list, own groups, duplicates of own groups
	check(true, nil)
	check(true, []string{ownGroup1.ID, ownGroup2.ID})
	check(true, []string{ownGroup1.ID, ownGroup1.ID})

	// Negative: foreign group, mixed list, unknown group
	check(false, []string{foreignGroup.ID})
	check(false, []string{ownGroup1.ID, foreignGroup.ID})
	check(false, []string{"00000000-0000-4000-8000-000000000000"})
}
