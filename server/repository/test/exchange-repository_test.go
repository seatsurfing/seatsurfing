package test

import (
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestExchangeOrgSettingsUpsertAndGet(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("exchange-test.com")

	settings := &ExchangeSettings{
		Enabled:      true,
		TenantID:     "test-tenant",
		ClientID:     "test-client",
		ClientSecret: "super-secret",
	}

	err := GetSettingsRepository().SetExchangeSettings(org.ID, settings)
	CheckTestBool(t, true, err == nil)

	got, err := GetSettingsRepository().GetExchangeSettings(org.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestBool(t, true, got.Enabled)
	CheckTestString(t, "test-tenant", got.TenantID)
	CheckTestString(t, "test-client", got.ClientID)
	CheckTestString(t, "super-secret", got.ClientSecret)
}

func TestExchangeOrgSettingsUpsertPreservesSecretOnEmptyUpdate(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("exchange-test2.com")

	initial := &ExchangeSettings{
		Enabled:      true,
		TenantID:     "tid",
		ClientID:     "cid",
		ClientSecret: "original-secret",
	}
	GetSettingsRepository().SetExchangeSettings(org.ID, initial)

	// Update without secret (empty string -> preserve existing)
	update := &ExchangeSettings{
		Enabled:      false,
		TenantID:     "tid2",
		ClientID:     "cid2",
		ClientSecret: "", // empty -> keep existing
	}
	err := GetSettingsRepository().SetExchangeSettings(org.ID, update)
	CheckTestBool(t, true, err == nil)

	got, _ := GetSettingsRepository().GetExchangeSettings(org.ID)
	CheckTestString(t, "original-secret", got.ClientSecret)
	CheckTestBool(t, false, got.Enabled)
	CheckTestString(t, "tid2", got.TenantID)
}

func TestExchangeOrgSettingsDelete(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("exchange-delete.com")

	settings := &ExchangeSettings{
		Enabled:      true,
		TenantID:     "tid",
		ClientID:     "cid",
		ClientSecret: "secret",
	}
	GetSettingsRepository().SetExchangeSettings(org.ID, settings)

	// Clear by overwriting with empty/disabled values
	cleared := &ExchangeSettings{}
	err := GetSettingsRepository().SetExchangeSettings(org.ID, cleared)
	CheckTestBool(t, true, err == nil)

	got, _ := GetSettingsRepository().GetExchangeSettings(org.ID)
	CheckTestBool(t, false, got.Enabled)
	CheckTestString(t, "", got.TenantID)
	CheckTestString(t, "", got.ClientID)
}

func TestExchangeSpaceMappingUpsertAndGet(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("exchange-space.com")
	_, space := CreateTestLocationAndSpace(org)

	mapping := &ExchangeSpaceMapping{
		SpaceID:   space.ID,
		RoomEmail: "room@contoso.com",
	}
	err := GetExchangeSpaceMappingRepository().Upsert(mapping)
	CheckTestBool(t, true, err == nil)

	got, err := GetExchangeSpaceMappingRepository().GetBySpaceID(space.ID)
	CheckTestBool(t, true, err == nil)
	CheckTestString(t, space.ID, got.SpaceID)
	CheckTestString(t, "room@contoso.com", got.RoomEmail)
}

func TestExchangeSpaceMappingDelete(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("exchange-space-del.com")
	_, space := CreateTestLocationAndSpace(org)

	mapping := &ExchangeSpaceMapping{SpaceID: space.ID, RoomEmail: "room@contoso.com"}
	GetExchangeSpaceMappingRepository().Upsert(mapping)

	err := GetExchangeSpaceMappingRepository().Delete(space.ID)
	CheckTestBool(t, true, err == nil)

	_, err = GetExchangeSpaceMappingRepository().GetBySpaceID(space.ID)
	CheckTestBool(t, false, err == nil)
}

func TestExchangeBookingMappingCreateAndGet(t *testing.T) {
	ClearTestDB()

	bid := "550e8400-e29b-41d4-a716-446655440001"
	err := GetExchangeBookingMappingRepository().Create(bid, "exchange-event-abc", "room@contoso.com")
	CheckTestBool(t, true, err == nil)

	got, err := GetExchangeBookingMappingRepository().GetByBookingID(bid)
	CheckTestBool(t, true, err == nil)
	CheckTestString(t, bid, got.BookingID)
	CheckTestString(t, "exchange-event-abc", got.ExchangeEventID)
	CheckTestString(t, "room@contoso.com", got.RoomEmail)
}

func TestExchangeBookingMappingDelete(t *testing.T) {
	ClearTestDB()

	bid := "550e8400-e29b-41d4-a716-446655440002"
	GetExchangeBookingMappingRepository().Create(bid, "event-id-xyz", "room@test.com")
	err := GetExchangeBookingMappingRepository().Delete(bid)
	CheckTestBool(t, true, err == nil)

	_, err = GetExchangeBookingMappingRepository().GetByBookingID(bid)
	CheckTestBool(t, false, err == nil)
}

func TestExchangeSyncQueueEnqueueAndClaim(t *testing.T) {
	ClearTestDB()

	bid := "550e8400-e29b-41d4-a716-446655440010"
	err := GetExchangeSyncQueueRepository().Enqueue(bid, "CREATE", `{"orgID":"org1"}`)
	CheckTestBool(t, true, err == nil)

	items, err := GetExchangeSyncQueueRepository().ClaimBatch(10)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, len(items))
	CheckTestString(t, bid, items[0].BookingID)
	CheckTestString(t, "CREATE", items[0].Operation)
	CheckTestString(t, `{"orgID":"org1"}`, items[0].Payload)
}

func TestExchangeSyncQueueMarkFailed(t *testing.T) {
	ClearTestDB()

	GetExchangeSyncQueueRepository().Enqueue("550e8400-e29b-41d4-a716-446655440011", "UPDATE", `{"orgID":"org-failed"}`)
	items, _ := GetExchangeSyncQueueRepository().ClaimBatch(10)
	CheckTestInt(t, 1, len(items))

	err := GetExchangeSyncQueueRepository().MarkFailed(items[0].ID, "some error")
	CheckTestBool(t, true, err == nil)

	failed, err := GetExchangeSyncQueueRepository().GetFailed("org-failed")
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, len(failed))
	CheckTestString(t, "some error", failed[0].LastError)
}

func TestExchangeSyncQueueResetToPending(t *testing.T) {
	ClearTestDB()

	GetExchangeSyncQueueRepository().Enqueue("550e8400-e29b-41d4-a716-446655440012", "DELETE", `{"orgID":"org-reset"}`)
	items, _ := GetExchangeSyncQueueRepository().ClaimBatch(10)
	CheckTestInt(t, 1, len(items))
	GetExchangeSyncQueueRepository().MarkFailed(items[0].ID, "err")

	err := GetExchangeSyncQueueRepository().ResetToPending(items[0].ID)
	CheckTestBool(t, true, err == nil)

	// After reset, should be claimable again
	reclaimed, err := GetExchangeSyncQueueRepository().ClaimBatch(10)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 1, len(reclaimed))
}
