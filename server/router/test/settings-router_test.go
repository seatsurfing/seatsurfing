package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/config"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

func TestSettingsForbidden(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	loginResponse := LoginTestUser(user.ID)

	payload := `{"value": "1"}`
	req := NewHTTPRequest("PUT", "/setting/"+SettingAllowAnyUser.Name, loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	req = NewHTTPRequest("GET", "/setting/"+SettingAllowAnyUser.Name, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	req = NewHTTPRequest("GET", "/setting/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)

	payload = `[]`
	req = NewHTTPRequest("PUT", "/setting/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestSettingsReadPublic(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	loginResponse := LoginTestUser(user.ID)

	allowedSettings := []string{
		SettingDisableBuddies.Name,
		SettingMaxBookingsPerUser.Name,
		SettingMaxConcurrentBookingsPerUser.Name,
		SettingMaxDaysInAdvance.Name,
		SettingMaxBookingDurationHours.Name,
		SettingMaxHoursBeforeDelete.Name,
		SettingEnableMaxHourBeforeDelete.Name,
		SettingDailyBasisBooking.Name,
		SettingNoAdminRestrictions.Name,
		SettingShowNames.Name,
		SettingMaxHoursPartiallyBooked.Name,
		SettingMaxHoursPartiallyBookedEnabled.Name,
		SettingMinBookingDurationHours.Name,
		SettingAllowBookingsNonExistingUsers.Name,
		SettingDefaultTimezone.Name,
		SettingCustomLogoUrl.Name,
		SysSettingVersion,
		SysSettingOrgPrimaryDomain,
		SysSettingOrgLanguage,
		SysSettingDisablePasswordLogin,
		SettingFeatureNoUserLimit.Name,
		SettingFeatureGroups.Name,
		SettingFeatureCustomDomains.Name,
		SettingAllowRecurringBookings.Name,
		SettingSubjectDefault.Name,
		SettingEnforceTOTP.Name,
		SettingFeatureKioskMode.Name,
		SettingHideReports.Name,
		SettingHideStats.Name,
		SettingPublicBookingEnabled.Name,
		SettingFeaturePublicBooking.Name,
		SysSettingBookingUIIntegrations,
	}
	forbiddenSettings := []string{
		SettingDatabaseVersion.Name,
		SettingAllowAnyUser.Name,
		SettingHideDisallowedLocations.Name,
	}

	for _, name := range allowedSettings {
		req := NewHTTPRequest("GET", "/setting/"+name, loginResponse.UserID, nil)
		res := ExecuteTestRequest(req)
		CheckTestResponseCode(t, http.StatusOK, res.Code)
	}

	for _, name := range forbiddenSettings {
		req := NewHTTPRequest("GET", "/setting/"+name, loginResponse.UserID, nil)
		res := ExecuteTestRequest(req)
		CheckTestResponseCode(t, http.StatusForbidden, res.Code)
	}

	req := NewHTTPRequest("GET", "/setting/", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []GetSettingsResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, len(allowedSettings), len(resBody))
	found := 0
	for _, name := range allowedSettings {
		for _, cur := range resBody {
			if name == cur.Name {
				found++
			}
		}
	}
	CheckTestInt(t, len(allowedSettings), found)
}

func TestSettingsReadAdmin(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	allowedSettings := []string{
		SettingDisableBuddies.Name,
		SettingMaxBookingsPerUser.Name,
		SettingMaxConcurrentBookingsPerUser.Name,
		SettingMaxDaysInAdvance.Name,
		SettingMaxBookingDurationHours.Name,
		SettingMaxHoursBeforeDelete.Name,
		SettingDailyBasisBooking.Name,
		SettingMinBookingDurationHours.Name,
		SettingNoAdminRestrictions.Name,
		SettingShowNames.Name,
		SettingEnableMaxHourBeforeDelete.Name,
		SettingMaxHoursPartiallyBooked.Name,
		SettingMaxHoursPartiallyBookedEnabled.Name,
		SettingAllowBookingsNonExistingUsers.Name,
		SettingAllowAnyUser.Name,
		SettingFeatureNoUserLimit.Name,
		SettingFeatureGroups.Name,
		SettingFeatureCustomDomains.Name,
		SettingDefaultTimezone.Name,
		SettingCustomLogoUrl.Name,
		SysSettingOrgSignupDelete,
		SysSettingVersion,
		SysSettingAdminMenuItems,
		SysSettingAdminWelcomeScreens,
		SysSettingOrgPrimaryDomain,
		SysSettingOrgLanguage,
		SysSettingDisablePasswordLogin,
		SysSettingInstallID,
		SettingBookingRetentionEnabled.Name,
		SettingBookingRetentionDays.Name,
		SettingAllowRecurringBookings.Name,
		SettingNewUserDefaultMailNotification.Name,
		SettingSubjectDefault.Name,
		SettingEnforceTOTP.Name,
		SettingTargetUtilizationHoursPerWeek.Name,
		SettingFeatureKioskMode.Name,
		SettingKioskModeEnabled.Name,
		SettingHideReports.Name,
		SettingHideStats.Name,
		SettingPublicBookingEnabled.Name,
		SettingFeaturePublicBooking.Name,
		SettingPublicBookingShowMap.Name,
		SettingPublicBookingShowAvailability.Name,
		SettingPublicBookingMinDurationHours.Name,
		SettingPublicBookingMaxDurationHours.Name,
		SettingHideDisallowedLocations.Name,
		SysSettingBookingUIIntegrations,
	}
	forbiddenSettings := []string{
		SettingDatabaseVersion.Name,
	}

	for _, name := range allowedSettings {
		req := NewHTTPRequest("GET", "/setting/"+name, loginResponse.UserID, nil)
		res := ExecuteTestRequest(req)
		CheckTestResponseCode(t, http.StatusOK, res.Code)
	}

	for _, name := range forbiddenSettings {
		req := NewHTTPRequest("GET", "/setting/"+name, loginResponse.UserID, nil)
		res := ExecuteTestRequest(req)
		CheckTestResponseCode(t, http.StatusForbidden, res.Code)
	}

	req := NewHTTPRequest("GET", "/setting/", loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []GetSettingsResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, len(allowedSettings), len(resBody))
	found := 0
	for _, name := range allowedSettings {
		for _, cur := range resBody {
			if name == cur.Name {
				found++
			}
		}
	}
	CheckTestInt(t, len(allowedSettings), found)
}

func TestSettingsCRUD(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	payload := `{"value": "1"}`
	req := NewHTTPRequest("PUT", "/setting/"+SettingAllowAnyUser.Name, loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	req = NewHTTPRequest("GET", "/setting/"+SettingAllowAnyUser.Name, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody string
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestString(t, "1", resBody)

	payload = `{"value": "0"}`
	req = NewHTTPRequest("PUT", "/setting/"+SettingAllowAnyUser.Name, loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	req = NewHTTPRequest("GET", "/setting/"+SettingAllowAnyUser.Name, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody2 string
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	CheckTestString(t, "0", resBody2)
}

func TestSettingsCRUDMany(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)
	GetDatabase().DB().Exec("TRUNCATE settings")

	payload := `[{"name": "allow_any_user", "value": "1"}, {"name": "max_bookings_per_user", "value": "5"}]`
	req := NewHTTPRequest("PUT", "/setting/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	req = NewHTTPRequest("GET", "/setting/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []GetSettingsResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, 11, len(resBody))
	CheckTestString(t, SettingAllowAnyUser.Name, resBody[0].Name)
	CheckTestString(t, SettingMaxBookingsPerUser.Name, resBody[1].Name)
	CheckTestString(t, SysSettingOrgSignupDelete, resBody[2].Name)
	CheckTestString(t, SysSettingAdminWelcomeScreens, resBody[3].Name)
	CheckTestString(t, SysSettingAdminMenuItems, resBody[4].Name)
	CheckTestString(t, SysSettingBookingUIIntegrations, resBody[5].Name)
	CheckTestString(t, SysSettingVersion, resBody[6].Name)
	CheckTestString(t, "1", resBody[0].Value)
	CheckTestString(t, "5", resBody[1].Value)
	CheckTestString(t, GetProductVersion(), resBody[6].Value)

	payload = `[{"name": "allow_any_user", "value": "0"}, {"name": "max_bookings_per_user", "value": "3"}]`
	req = NewHTTPRequest("PUT", "/setting/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	req = NewHTTPRequest("GET", "/setting/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody2 []GetSettingsResponse
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	CheckTestInt(t, 11, len(resBody2))
	CheckTestString(t, SettingAllowAnyUser.Name, resBody2[0].Name)
	CheckTestString(t, SettingMaxBookingsPerUser.Name, resBody2[1].Name)
	CheckTestString(t, SysSettingOrgSignupDelete, resBody2[2].Name)
	CheckTestString(t, SysSettingVersion, resBody2[6].Name)
	CheckTestString(t, "0", resBody2[0].Value)
	CheckTestString(t, "3", resBody2[1].Value)

}

func TestSettingsMaxHoursBeforeDelete(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)
	GetDatabase().DB().Exec("TRUNCATE settings")

	payload := `[{"name": "max_hours_before_delete", "value": "2"}]`
	req := NewHTTPRequest("PUT", "/setting/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	req = NewHTTPRequest("GET", "/setting/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody3 []GetSettingsResponse
	json.Unmarshal(res.Body.Bytes(), &resBody3)
	CheckTestInt(t, 10, len(resBody3))
	CheckTestString(t, SettingMaxHoursBeforeDelete.Name, resBody3[0].Name)
	CheckTestString(t, SysSettingOrgSignupDelete, resBody3[1].Name)
	CheckTestString(t, SysSettingVersion, resBody3[5].Name)
	CheckTestString(t, "2", resBody3[0].Value)
}

func TestSettingsMinHoursBookingDuration(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)
	GetDatabase().DB().Exec("TRUNCATE settings")

	payload := `[{"name": "min_booking_duration_hours", "value": "2"}]`
	req := NewHTTPRequest("PUT", "/setting/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	req = NewHTTPRequest("GET", "/setting/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody3 []GetSettingsResponse
	json.Unmarshal(res.Body.Bytes(), &resBody3)
	CheckTestInt(t, 10, len(resBody3))
	CheckTestString(t, SettingMinBookingDurationHours.Name, resBody3[0].Name)
	CheckTestString(t, SysSettingOrgSignupDelete, resBody3[1].Name)
	CheckTestString(t, SysSettingVersion, resBody3[5].Name)
	CheckTestString(t, "2", resBody3[0].Value)
}

func TestSettingsInvalidName(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	payload := `{"value": "1"}`
	req := NewHTTPRequest("PUT", "/setting/test123", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestSettingsInvalidBool(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	payload := `{"value": "2"}`
	req := NewHTTPRequest("PUT", "/setting/"+SettingAllowAnyUser.Name, loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
}

func TestSettingsInvalidInt(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	payload := `{"value": "test"}`
	req := NewHTTPRequest("PUT", "/setting/"+SettingMaxBookingsPerUser.Name, loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
}

func TestSettingsInvalidTimezone(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	payload := `{"value": "Europe/Hamburg"}`
	req := NewHTTPRequest("PUT", "/setting/"+SettingDefaultTimezone.Name, loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)

	payload = `{"value": "Europe/Berlin"}`
	req = NewHTTPRequest("PUT", "/setting/"+SettingDefaultTimezone.Name, loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)
}

func TestSettingsInstallIDExposure(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	GetConfig().DisableInstallIDExposure = false
	defer func() { GetConfig().DisableInstallIDExposure = false }()

	req := NewHTTPRequest("GET", "/setting/"+SysSettingInstallID, loginResponse.UserID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var installID string
	json.Unmarshal(res.Body.Bytes(), &installID)
	if installID == "" {
		t.Fatal("Expected non-empty install ID when exposure is enabled")
	}

	req = NewHTTPRequest("GET", "/setting/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []GetSettingsResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	found := false
	for _, cur := range resBody {
		if cur.Name == SysSettingInstallID {
			found = true
		}
	}
	if !found {
		t.Fatal("Expected install ID setting to be present in list when exposure is enabled")
	}

	GetConfig().DisableInstallIDExposure = true

	req = NewHTTPRequest("GET", "/setting/"+SysSettingInstallID, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	json.Unmarshal(res.Body.Bytes(), &installID)
	CheckTestString(t, "", installID)

	req = NewHTTPRequest("GET", "/setting/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody2 []GetSettingsResponse
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	for _, cur := range resBody2 {
		if cur.Name == SysSettingInstallID {
			t.Fatal("Expected install ID setting to be absent from list when exposure is disabled")
		}
	}
}

func TestSettingsGetTimezones(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	// getTimezones has NO permission check - any authenticated user can access
	req := NewHTTPRequest("GET", "/setting/timezones", user.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []string
	json.Unmarshal(res.Body.Bytes(), &resBody)
	if len(resBody) == 0 {
		t.Fatal("Expected non-empty timezone list")
	}
}

func getBookingUIIntegrations(t *testing.T, userID string) []SettingsRouterBookingUIIntegration {
	req := NewHTTPRequest("GET", "/setting/"+SysSettingBookingUIIntegrations, userID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody GetSettingsResponse
	if err := json.Unmarshal(res.Body.Bytes(), &resBody); err != nil {
		t.Fatal(err)
	}
	var items []SettingsRouterBookingUIIntegration
	if err := json.Unmarshal([]byte(resBody.Value), &items); err != nil {
		t.Fatal(err)
	}
	return items
}

func findBookingUIIntegration(items []SettingsRouterBookingUIIntegration, id string) *SettingsRouterBookingUIIntegration {
	for i := range items {
		if items[i].ID == id {
			return &items[i]
		}
	}
	return nil
}

// fakeBookingUIPluginID identifies the plugin permissions registered for
// TestBookingUIIntegrationsPermissionFiltering, so they can be unregistered
// again afterwards (RegisterPermission has no per-test scoping otherwise).
const fakeBookingUIPluginID = "fake-booking-ui-test-plugin"

func TestBookingUIIntegrationsPermissionFiltering(t *testing.T) {
	ClearTestDB()
	ResetPluginsForTest()
	defer ResetPluginsForTest()
	// "plugin.chat"/"plugin.other" must be registered in the permission
	// catalogue for a role granting them to actually take effect:
	// GetEffectivePermissions drops any permission the catalogue doesn't
	// know, so a role referencing an unregistered plugin permission would
	// otherwise silently grant nothing.
	RegisterPermission(PermissionDefinition{
		Key:           "plugin.chat",
		AllowedLevels: []PermissionLevel{PermissionLevelNone, PermissionLevelRead, PermissionLevelAdmin},
		PluginID:      fakeBookingUIPluginID,
	})
	RegisterPermission(PermissionDefinition{
		Key:           "plugin.other",
		AllowedLevels: []PermissionLevel{PermissionLevelNone, PermissionLevelAdmin},
		PluginID:      fakeBookingUIPluginID,
	})
	defer UnregisterPluginPermissions(fakeBookingUIPluginID)

	org := CreateTestOrg("test.com")
	plainUser := CreateTestUserInOrg(org)
	chatUser := CreateTestUserWithPermissions(org, map[Permission]PermissionLevel{
		"plugin.chat": PermissionLevelRead,
	})

	RegisterPlugin(&fakeBookingUIPlugin{
		integrations: []BookingUIIntegration{
			{ID: "everyone", Title: "Everyone"},
			{
				ID:                 "chat",
				Title:              "Chat",
				RequiredPermission: "plugin.chat",
				RequiredLevel:      PermissionLevelRead,
			},
			{
				ID:                     "admin-any",
				Title:                  "Admin Any",
				RequiredPermissionsAny: []Permission{"plugin.chat", "plugin.other"},
				RequiredLevel:          PermissionLevelAdmin,
			},
		},
	})

	// A plain user with no permissions at all sees only the integration that
	// declares none.
	plainItems := getBookingUIIntegrations(t, plainUser.ID)
	if findBookingUIIntegration(plainItems, "everyone") == nil {
		t.Error("expected plain user to see the 'everyone' integration")
	}
	if findBookingUIIntegration(plainItems, "chat") != nil {
		t.Error("expected plain user not to see the 'chat' integration")
	}
	if findBookingUIIntegration(plainItems, "admin-any") != nil {
		t.Error("expected plain user not to see the 'admin-any' integration")
	}

	// A user holding "plugin.chat" at Read sees "everyone" and "chat", but
	// not "admin-any" (which requires Admin level).
	chatItems := getBookingUIIntegrations(t, chatUser.ID)
	if findBookingUIIntegration(chatItems, "everyone") == nil {
		t.Error("expected chat user to see the 'everyone' integration")
	}
	if findBookingUIIntegration(chatItems, "chat") == nil {
		t.Error("expected chat user to see the 'chat' integration")
	}
	if findBookingUIIntegration(chatItems, "admin-any") != nil {
		t.Error("expected chat user not to see the 'admin-any' integration (requires Admin level)")
	}
}

func TestBookingUIIntegrationsWidthClamping(t *testing.T) {
	ClearTestDB()
	ResetPluginsForTest()
	defer ResetPluginsForTest()

	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	RegisterPlugin(&fakeBookingUIPlugin{
		integrations: []BookingUIIntegration{
			{ID: "unset", Title: "Unset"},
			{ID: "too-small", Title: "Too Small", Width: 1},
			{ID: "too-big", Title: "Too Big", Width: 99999},
			{ID: "in-range", Title: "In Range", Width: 300},
		},
	})

	items := getBookingUIIntegrations(t, user.ID)

	tests := []struct {
		id    string
		width int
	}{
		{"unset", BookingUIIntegrationDefaultWidth},
		{"too-small", BookingUIIntegrationMinWidth},
		{"too-big", BookingUIIntegrationMaxWidth},
		{"in-range", 300},
	}
	for _, tc := range tests {
		item := findBookingUIIntegration(items, tc.id)
		if item == nil {
			t.Fatalf("expected integration %q to be present", tc.id)
		}
		CheckTestInt(t, tc.width, item.Width)
	}
}

func TestSettingsPublicBookingDurationHours(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	for _, name := range []string{SettingPublicBookingMinDurationHours.Name, SettingPublicBookingMaxDurationHours.Name} {
		req := NewHTTPRequest("PUT", "/setting/"+name, loginResponse.UserID, bytes.NewBufferString(`{"value": "3"}`))
		res := ExecuteTestRequest(req)
		CheckTestResponseCode(t, http.StatusNoContent, res.Code)

		req = NewHTTPRequest("GET", "/setting/"+name, loginResponse.UserID, nil)
		res = ExecuteTestRequest(req)
		CheckTestResponseCode(t, http.StatusOK, res.Code)
		var resBody string
		json.Unmarshal(res.Body.Bytes(), &resBody)
		CheckTestString(t, "3", resBody)

		for _, value := range []string{`{"value": "-1"}`, `{"value": "25"}`, `{"value": "abc"}`} {
			req = NewHTTPRequest("PUT", "/setting/"+name, loginResponse.UserID, bytes.NewBufferString(value))
			res = ExecuteTestRequest(req)
			CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
		}
	}
}

func TestSettingsPublicBookingDurationHoursForbiddenForUser(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	loginResponse := LoginTestUser(user.ID)

	req := NewHTTPRequest("PUT", "/setting/"+SettingPublicBookingMinDurationHours.Name, loginResponse.UserID, bytes.NewBufferString(`{"value": "3"}`))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)

	req = NewHTTPRequest("GET", "/setting/"+SettingPublicBookingMinDurationHours.Name, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusForbidden, res.Code)
}

func TestSettingsPublicBookingDurationHoursInvalidRange(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserOrgAdmin(org)
	loginResponse := LoginTestUser(user.ID)

	req := NewHTTPRequest("PUT", "/setting/"+SettingPublicBookingMaxDurationHours.Name, loginResponse.UserID, bytes.NewBufferString(`{"value": "2"}`))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	req = NewHTTPRequest("PUT", "/setting/"+SettingPublicBookingMinDurationHours.Name, loginResponse.UserID, bytes.NewBufferString(`{"value": "10"}`))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)

	req = NewHTTPRequest("PUT", "/setting/"+SettingMinBookingDurationHours.Name, loginResponse.UserID, bytes.NewBufferString(`{"value": "3"}`))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)

	payload := `[
		{"name": "` + SettingMaxDaysInAdvance.Name + `", "value": "7"},
		{"name": "` + SettingPublicBookingMinDurationHours.Name + `", "value": "10"},
		{"name": "` + SettingPublicBookingMaxDurationHours.Name + `", "value": "12"},
		{"name": "` + SettingMaxBookingDurationHours.Name + `", "value": "8"}
	]`
	req = NewHTTPRequest("PUT", "/setting/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)

	minHours, _ := GetSettingsRepository().GetInt(org.ID, SettingPublicBookingMinDurationHours.Name)
	CheckTestInt(t, 0, minHours)
	maxHours, _ := GetSettingsRepository().GetInt(org.ID, SettingPublicBookingMaxDurationHours.Name)
	CheckTestInt(t, 2, maxHours)
	maxDays, _ := GetSettingsRepository().GetInt(org.ID, SettingMaxDaysInAdvance.Name)
	if maxDays == 7 {
		t.Fatal("expected no setting to be written")
	}

	payload = `[
		{"name": "` + SettingPublicBookingMinDurationHours.Name + `", "value": "10"},
		{"name": "` + SettingPublicBookingMaxDurationHours.Name + `", "value": "12"}
	]`
	req = NewHTTPRequest("PUT", "/setting/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)
}
