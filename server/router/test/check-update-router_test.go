package test

import (
	"encoding/json"
	"net/http"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/testutil"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

func TestUpdateCheckNoUpdate(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	// In test environment GetUpdateChecker().Latest will be nil since the update checker
	// is not initialized, so the endpoint should return an empty version
	req := NewHTTPRequest("GET", "/uc/", user.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusOK, res.Code)
	var resBody CheckVersionResponse
	json.Unmarshal(res.Body.Bytes(), &resBody)
	CheckTestBool(t, false, resBody.UpdateAvailable)
	CheckTestString(t, "", resBody.LatestVersion)
}
