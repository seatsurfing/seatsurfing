package test

import (
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestGetOrgIDsByValue(t *testing.T) {
	ClearTestDB()
	org1 := CreateTestOrg("test1.com")
	org2 := CreateTestOrg("test2.com")
	org3 := CreateTestOrg("test3.com")

	GetSettingsRepository().Set(org1.ID, "key1", "match-me")
	GetSettingsRepository().Set(org1.ID, "key2", "something")
	GetSettingsRepository().Set(org2.ID, "key1", "dont-match-me")
	GetSettingsRepository().Set(org3.ID, "key1", "match-me")

	res, err := GetSettingsRepository().GetOrgIDsByValue("key1", "match-me")
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 2, len(res))
	CheckTestBool(t, true, Contains(res, org1.ID))
	CheckTestBool(t, true, Contains(res, org3.ID))
}
