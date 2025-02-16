package test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"testing"

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

	req = NewHTTPRequest("GET", "/setting/"+SettingConfluenceServerSharedSecret.Name, loginResponse.UserID, nil)
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
	}
	forbiddenSettings := []string{
		SettingDatabaseVersion.Name,
		SettingAllowAnyUser.Name,
		SettingConfluenceServerSharedSecret.Name,
		SettingConfluenceAnonymous.Name,
		SettingSubscriptionMaxUsers.Name,
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
		SettingConfluenceServerSharedSecret.Name,
		SettingConfluenceAnonymous.Name,
		SettingSubscriptionMaxUsers.Name,
		SettingDefaultTimezone.Name,
		SettingCustomLogoUrl.Name,
		SysSettingOrgSignupDelete,
		SysSettingVersion,
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
	CheckTestInt(t, 4, len(resBody))
	CheckTestString(t, SettingAllowAnyUser.Name, resBody[0].Name)
	CheckTestString(t, SettingMaxBookingsPerUser.Name, resBody[1].Name)
	CheckTestString(t, SysSettingOrgSignupDelete, resBody[2].Name)
	CheckTestString(t, SysSettingVersion, resBody[3].Name)
	CheckTestString(t, "1", resBody[0].Value)
	CheckTestString(t, "5", resBody[1].Value)
	CheckTestString(t, GetProductVersion(), resBody[3].Value)

	payload = `[{"name": "allow_any_user", "value": "0"}, {"name": "max_bookings_per_user", "value": "3"}]`
	req = NewHTTPRequest("PUT", "/setting/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	req = NewHTTPRequest("GET", "/setting/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody2 []GetSettingsResponse
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	CheckTestInt(t, 4, len(resBody2))
	CheckTestString(t, SettingAllowAnyUser.Name, resBody2[0].Name)
	CheckTestString(t, SettingMaxBookingsPerUser.Name, resBody2[1].Name)
	CheckTestString(t, SysSettingOrgSignupDelete, resBody2[2].Name)
	CheckTestString(t, SysSettingVersion, resBody2[3].Name)
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
	log.Println(resBody3)
	CheckTestInt(t, 3, len(resBody3))
	CheckTestString(t, SettingMaxHoursBeforeDelete.Name, resBody3[0].Name)
	CheckTestString(t, SysSettingOrgSignupDelete, resBody3[1].Name)
	CheckTestString(t, SysSettingVersion, resBody3[2].Name)
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
	log.Println(resBody3)
	CheckTestInt(t, 3, len(resBody3))
	CheckTestString(t, SettingMinBookingDurationHours.Name, resBody3[0].Name)
	CheckTestString(t, SysSettingOrgSignupDelete, resBody3[1].Name)
	CheckTestString(t, SysSettingVersion, resBody3[2].Name)
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
