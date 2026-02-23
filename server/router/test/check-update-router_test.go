package test

import (
	"net/http"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestUpdateCheckNoUpdate(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	user := CreateTestUserInOrg(org)

	// In test environment GetUpdateChecker().Latest will be nil since the update checker
	// is not initialized, so the endpoint should return 404
	req := NewHTTPRequest("GET", "/uc/", user.ID, nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}
