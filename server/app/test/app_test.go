package test

import (
	"testing"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/app"
	. "github.com/seatsurfing/seatsurfing/server/config"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

func TestUpdateInstallStats(t *testing.T) {
	ClearTestDB()
	GetConfig().DisableAnonymousUsageStats = false

	org := CreateTestOrg("test.com")
	user1 := CreateTestUserInOrg(org)
	user2 := CreateTestUserInOrg(org)

	location, space := CreateTestLocationAndSpace(org)
	_ = location

	now := time.Now()
	booking := &Booking{
		UserID:  user1.ID,
		SpaceID: space.ID,
		Enter:   time.Date(now.Year(), now.Month(), now.Day()+1, 9, 0, 0, 0, time.Local),
		Leave:   time.Date(now.Year(), now.Month(), now.Day()+1, 17, 0, 0, 0, time.Local),
	}
	GetBookingRepository().Create(booking)

	_ = user2

	GetApp().UpdateInstallStats()

	stats := GetInstallStats()
	if stats == nil {
		t.Fatal("Expected install stats to be set")
	}
	if stats.NumLocations < 1 {
		t.Errorf("Expected NumLocations >= 1, got %d", stats.NumLocations)
	}
	if stats.NumSpaces < 1 {
		t.Errorf("Expected NumSpaces >= 1, got %d", stats.NumSpaces)
	}
	if stats.NumUsers < 2 {
		t.Errorf("Expected NumUsers >= 2, got %d", stats.NumUsers)
	}
	if stats.NumBookings < 1 {
		t.Errorf("Expected NumBookings >= 1, got %d", stats.NumBookings)
	}
}

func TestUpdateInstallStatsDisabled(t *testing.T) {
	ClearTestDB()
	GetConfig().DisableAnonymousUsageStats = true
	defer func() { GetConfig().DisableAnonymousUsageStats = false }()

	// Reset stats to a known zero state before the call
	SetInstallStats(&InstallStats{})

	CreateTestOrg("test.com")
	GetApp().UpdateInstallStats()

	stats := GetInstallStats()
	if stats.NumLocations != 0 || stats.NumUsers != 0 || stats.NumBookings != 0 || stats.NumSpaces != 0 {
		t.Errorf("Expected stats to remain zero when disabled, got %+v", stats)
	}
}
