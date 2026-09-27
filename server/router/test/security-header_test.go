package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestSecurityHeadersNonFramableUIPaths(t *testing.T) {
	ClearTestDB()
	paths := []string{
		"/ui/login/",
		"/ui/login",
		"/ui/resetpw/",
		"/ui/resetpw/2a4d7a4a-1b3c-4e5f-8a9b-0c1d2e3f4a5b/",
		"/ui/setpw/2a4d7a4a-1b3c-4e5f-8a9b-0c1d2e3f4a5b/",
		"/ui/book/",
		"/ui/book/details/abc/",
		"/ui/book/confirm/abc/",
		"/ui/admin",
		"/ui/admin/",
		"/ui/admin/dashboard/",
		"/ui/admin/users/2a4d7a4a-1b3c-4e5f-8a9b-0c1d2e3f4a5b/",
	}
	for _, p := range paths {
		req := httptest.NewRequest("GET", p, nil)
		res := ExecuteTestRequest(req)
		CheckTestString(t, "SAMEORIGIN", res.Header().Get("X-Frame-Options"))
		CheckTestBool(t, true, IsNonFramableUIPath(p))
		csp := res.Header().Get("Content-Security-Policy")
		if csp != "frame-ancestors 'self'" && csp != "upgrade-insecure-requests; frame-ancestors 'self'" {
			t.Fatalf("Expected frame-ancestors CSP for %s, got %q", p, csp)
		}
	}
}

func TestSecurityHeadersFramableUIPaths(t *testing.T) {
	ClearTestDB()
	// Pages loaded inside the MS Teams / Confluence iframe must stay embeddable.
	paths := []string{
		"/ui/",
		"/ui/search/",
		"/ui/bookings/",
		"/ui/preferences/",
		"/ui/login/success/2a4d7a4a-1b3c-4e5f-8a9b-0c1d2e3f4a5b/",
		"/ui/login/failed/",
		"/ui/loginx/",
		"/ui/booking/",
		"/ui/resetpwx/",
		"/ui/administration/",
		"/auth/login",
	}
	for _, p := range paths {
		req := httptest.NewRequest("GET", p, nil)
		res := ExecuteTestRequest(req)
		CheckTestString(t, "", res.Header().Get("X-Frame-Options"))
		CheckTestBool(t, false, IsNonFramableUIPath(p))
		csp := res.Header().Get("Content-Security-Policy")
		if csp != "" && csp != "upgrade-insecure-requests" {
			t.Fatalf("Expected no frame-ancestors CSP for %s, got %q", p, csp)
		}
	}
}

func TestSecurityHeadersKeepCommonHeaders(t *testing.T) {
	ClearTestDB()
	req := httptest.NewRequest("GET", "/ui/login/", nil)
	res := ExecuteTestRequest(req)
	CheckTestString(t, "nosniff", res.Header().Get("X-Content-Type-Options"))
	CheckTestString(t, "no-referrer", res.Header().Get("Referrer-Policy"))
	if res.Code == http.StatusInternalServerError {
		t.Fatalf("Unexpected status %d", res.Code)
	}
}
