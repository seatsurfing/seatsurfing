package test

import (
	"net/http"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

// The presence report covers personal data - who was in the office and when -
// while the statistics endpoints are aggregate. A role must be able to grant
// occupancy planning without handing over the former.
func TestAnalyticsAndPresenceReportAreSeparable(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	CreateTestUserOrgAdmin(org)

	statsOnly := CreateTestUserWithPermissions(org, map[Permission]PermissionLevel{
		PermissionAnalytics: PermissionLevelRead,
	})
	presenceOnly := CreateTestUserWithPermissions(org, map[Permission]PermissionLevel{
		PermissionPresenceReport: PermissionLevelRead,
	})

	statsLogin := LoginTestUser(statsOnly.ID)
	presenceLogin := LoginTestUser(presenceOnly.ID)

	// Utilization statistics: granted by analytics, withheld from the other.
	req := NewHTTPRequest("GET", "/stats/", statsLogin.UserID, nil)
	CheckTestResponseCode(t, http.StatusOK, ExecuteTestRequest(req).Code)
	req = NewHTTPRequest("GET", "/stats/", presenceLogin.UserID, nil)
	CheckTestResponseCode(t, http.StatusForbidden, ExecuteTestRequest(req).Code)

	// The presence report: the reverse.
	url := "/booking/report/presence/?start=2030-01-01T00:00:00Z&end=2030-01-02T00:00:00Z"
	req = NewHTTPRequest("GET", url, presenceLogin.UserID, nil)
	CheckTestResponseCode(t, http.StatusOK, ExecuteTestRequest(req).Code)
	req = NewHTTPRequest("GET", url, statsLogin.UserID, nil)
	CheckTestResponseCode(t, http.StatusForbidden, ExecuteTestRequest(req).Code)
}

// The organization-wide settings are a ceiling above the permission: they
// apply to everyone, including a user who holds the permission. That is what
// makes them expressible at all - an administrator can always grant themselves
// a permission, but not defeat the setting without deliberately changing it.
func TestOrgWideSettingsOverrideThePermission(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	login := LoginTestUser(admin.ID)

	GetSettingsRepository().Set(org.ID, SettingHideReports.Name, "1")
	GetSettingsRepository().Set(org.ID, SettingHideStats.Name, "1")

	req := NewHTTPRequest("GET",
		"/booking/report/presence/?start=2030-01-01T00:00:00Z&end=2030-01-02T00:00:00Z",
		login.UserID, nil)
	CheckTestResponseCode(t, http.StatusNotFound, ExecuteTestRequest(req).Code)

	req = NewHTTPRequest("GET", "/stats/", login.UserID, nil)
	CheckTestResponseCode(t, http.StatusNotFound, ExecuteTestRequest(req).Code)

	// With the settings off, the same administrator gets through.
	GetSettingsRepository().Set(org.ID, SettingHideReports.Name, "0")
	GetSettingsRepository().Set(org.ID, SettingHideStats.Name, "0")
	req = NewHTTPRequest("GET", "/stats/", login.UserID, nil)
	CheckTestResponseCode(t, http.StatusOK, ExecuteTestRequest(req).Code)
}

// A date range with no bookings must return an empty report, not a 500. The
// handler guarded the date count against an empty result but then indexed
// items[0] regardless - a pre-existing panic, unrelated to the permission
// model, found while verifying the split.
func TestPresenceReportEmptyRange(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	admin := CreateTestUserOrgAdmin(org)
	login := LoginTestUser(admin.ID)

	req := NewHTTPRequest("GET",
		"/booking/report/presence/?start=2035-01-01T00:00:00Z&end=2035-01-02T00:00:00Z",
		login.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
}
