package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestUserPreferencesCRUD(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	loginResponse := LoginTestUser(user.ID)

	payload := `{"value": "1"}`
	req := NewHTTPRequest("PUT", "/preference/"+PreferenceEnterTime.Name, loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	req = NewHTTPRequest("GET", "/preference/"+PreferenceEnterTime.Name, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody string
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestString(t, "1", resBody)

	payload = `{"value": "2"}`
	req = NewHTTPRequest("PUT", "/preference/"+PreferenceEnterTime.Name, loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	req = NewHTTPRequest("GET", "/preference/"+PreferenceEnterTime.Name, loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody2 string
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	CheckTestString(t, "2", resBody2)
}

func TestUserPreferencesCRUDMany(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)
	loginResponse := LoginTestUser(user.ID)
	GetDatabase().DB().Exec("TRUNCATE users_preferences")

	payload := `[{"name": "enter_time", "value": "1"}, {"name": "workday_start", "value": "5"}]`
	req := NewHTTPRequest("PUT", "/preference/", loginResponse.UserID, bytes.NewBufferString(payload))
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	req = NewHTTPRequest("GET", "/preference/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody []GetSettingsResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestInt(t, 2, len(resBody))
	CheckTestString(t, PreferenceEnterTime.Name, resBody[0].Name)
	CheckTestString(t, PreferenceWorkdayStart.Name, resBody[1].Name)
	CheckTestString(t, "1", resBody[0].Value)
	CheckTestString(t, "5", resBody[1].Value)

	payload = `[{"name": "enter_time", "value": "2"}, {"name": "workday_start", "value": "3"}]`
	req = NewHTTPRequest("PUT", "/preference/", loginResponse.UserID, bytes.NewBufferString(payload))
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNoContent, res.Code)

	req = NewHTTPRequest("GET", "/preference/", loginResponse.UserID, nil)
	res = ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody2 []GetSettingsResponse
	json.Unmarshal(res.Body.Bytes(), &resBody2)
	CheckTestInt(t, 2, len(resBody2))
	CheckTestString(t, PreferenceEnterTime.Name, resBody2[0].Name)
	CheckTestString(t, PreferenceWorkdayStart.Name, resBody2[1].Name)
	CheckTestString(t, "2", resBody2[0].Value)
	CheckTestString(t, "3", resBody2[1].Value)
}
