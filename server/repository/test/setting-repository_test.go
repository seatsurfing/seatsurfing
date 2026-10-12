package test

import (
	"sync"
	"testing"

	"github.com/google/uuid"

	. "github.com/seatsurfing/seatsurfing/server/api"
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

func TestSettingsSchemaUpgradeRemovesConfluenceSettings(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")

	GetSettingsRepository().Set(org.ID, "confluence_server_shared_secret", "secret")
	GetSettingsRepository().Set(org.ID, "confluence_anonymous", "1")
	GetSettingsRepository().Set(org.ID, "unrelated_setting", "1")

	GetSettingsRepository().RunSchemaUpgrade(59, 60)

	res, err := GetSettingsRepository().GetOrgIDsByValue("confluence_server_shared_secret", "secret")
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 0, len(res))
	res, err = GetSettingsRepository().GetOrgIDsByValue("confluence_anonymous", "1")
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 0, len(res))
	res, err = GetSettingsRepository().GetOrgIDsByValue("unrelated_setting", "1")
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, len(res))
}

func TestSetGlobalInstallIDCreatesIDIfMissing(t *testing.T) {
	ClearTestDB()
	GetCache().Delete(GetSettingsRepository().GetNullUUID() + "_" + SettingInstallID.Name)

	SetGlobalInstallID()

	id, err := GetSettingsRepository().GetGlobalString(SettingInstallID.Name)
	CheckTestBool(t, true, err == nil)
	_, err = uuid.Parse(id)
	CheckTestBool(t, true, err == nil)
}

func TestSetGlobalInstallIDKeepsExistingID(t *testing.T) {
	ClearTestDB()
	existing := "11111111-2222-3333-4444-555555555555"
	GetSettingsRepository().SetGlobal(SettingInstallID.Name, existing)

	SetGlobalInstallID()
	SetGlobalInstallID()

	id, err := GetSettingsRepository().GetGlobalString(SettingInstallID.Name)
	CheckTestBool(t, true, err == nil)
	CheckTestString(t, existing, id)
}

func TestSetGlobalInstallIDReplacesEmptyID(t *testing.T) {
	ClearTestDB()
	GetSettingsRepository().SetGlobal(SettingInstallID.Name, "")

	SetGlobalInstallID()

	id, err := GetSettingsRepository().GetGlobalString(SettingInstallID.Name)
	CheckTestBool(t, true, err == nil)
	_, err = uuid.Parse(id)
	CheckTestBool(t, true, err == nil)
}

func TestSetGlobalInstallIDConcurrent(t *testing.T) {
	ClearTestDB()
	GetCache().Delete(GetSettingsRepository().GetNullUUID() + "_" + SettingInstallID.Name)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			SetGlobalInstallID()
		}()
	}
	wg.Wait()

	first, err := GetSettingsRepository().GetGlobalString(SettingInstallID.Name)
	CheckTestBool(t, true, err == nil)
	SetGlobalInstallID()
	second, _ := GetSettingsRepository().GetGlobalString(SettingInstallID.Name)
	CheckTestString(t, first, second)
}
