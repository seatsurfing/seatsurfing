package test

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/config"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

func TestDomainVerificationDefaultResolver(t *testing.T) {
	os.Setenv("DNS_SERVER", "")
	GetConfig().ReadConfig()
	CheckTestBool(t, true, IsValidTXTRecord("seatsurfing-testcase.virtualzone.de", "65e51a4b-339f-4b24-b376-f9d866057b38"))
	CheckTestBool(t, false, IsValidTXTRecord("seatsurfing-testcase.virtualzone.de", "invalid-uuid"))
}

func TestDomainVerificationCustomResolver(t *testing.T) {
	os.Setenv("DNS_SERVER", "8.8.8.8")
	GetConfig().ReadConfig()
	CheckTestBool(t, true, IsValidTXTRecord("seatsurfing-testcase.virtualzone.de", "65e51a4b-339f-4b24-b376-f9d866057b38"))
	CheckTestBool(t, false, IsValidTXTRecord("seatsurfing-testcase.virtualzone.de", "invalid-uuid"))
}

// setAccessibilityConfig configures PUBLIC_SCHEME/PUBLIC_PORT and registers a cleanup
// that restores the previous config after the test.
func setAccessibilityConfig(t *testing.T, scheme, port string) {
	t.Helper()
	prevScheme := os.Getenv("PUBLIC_SCHEME")
	prevPort := os.Getenv("PUBLIC_PORT")
	os.Setenv("PUBLIC_SCHEME", scheme)
	os.Setenv("PUBLIC_PORT", port)
	GetConfig().ReadConfig()
	t.Cleanup(func() {
		os.Setenv("PUBLIC_SCHEME", prevScheme)
		os.Setenv("PUBLIC_PORT", prevPort)
		GetConfig().ReadConfig()
	})
}

// accessibilityHandler returns an HTTP handler that writes a DomainAccessibilityPayload
// JSON response with the given domain, orgID and status.
func accessibilityHandler(domain, orgID, status string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"domain":%q,"orgID":%q,"status":%q}`, domain, orgID, status)
	}
}

// TestIsDomainAccessibleSuccess verifies that IsDomainAccessible returns true when
// the server at the configured PUBLIC_SCHEME/PUBLIC_PORT returns a valid response.
func TestIsDomainAccessibleSuccess(t *testing.T) {
	const orgID = "test-org-id"
	const domain = "127.0.0.1"

	srv := httptest.NewServer(accessibilityHandler(domain, orgID, "OK"))
	defer srv.Close()

	port := srv.Listener.Addr().(*net.TCPAddr).Port
	setAccessibilityConfig(t, "http", strconv.Itoa(port))

	ok, err := IsDomainAccessible(domain, orgID)
	CheckTestBool(t, true, ok)
	CheckTestBool(t, true, err == nil)
}

// TestIsDomainAccessibleWrongOrgID verifies that IsDomainAccessible returns false
// when the server returns a payload with a mismatched orgID.
func TestIsDomainAccessibleWrongOrgID(t *testing.T) {
	const domain = "127.0.0.1"

	srv := httptest.NewServer(accessibilityHandler(domain, "other-org-id", "OK"))
	defer srv.Close()

	port := srv.Listener.Addr().(*net.TCPAddr).Port
	setAccessibilityConfig(t, "http", strconv.Itoa(port))

	ok, err := IsDomainAccessible(domain, "expected-org-id")
	CheckTestBool(t, false, ok)
	CheckTestBool(t, false, err == nil)
}

// TestIsDomainAccessibleWrongDomain verifies that IsDomainAccessible returns false
// when the server returns a payload with a mismatched domain.
func TestIsDomainAccessibleWrongDomain(t *testing.T) {
	const orgID = "test-org-id"
	const domain = "127.0.0.1"

	srv := httptest.NewServer(accessibilityHandler("other.domain", orgID, "OK"))
	defer srv.Close()

	port := srv.Listener.Addr().(*net.TCPAddr).Port
	setAccessibilityConfig(t, "http", strconv.Itoa(port))

	ok, err := IsDomainAccessible(domain, orgID)
	CheckTestBool(t, false, ok)
	CheckTestBool(t, false, err == nil)
}

// TestIsDomainAccessibleBadStatus verifies that IsDomainAccessible returns false
// when the server returns a non-OK status in the payload.
func TestIsDomainAccessibleBadStatus(t *testing.T) {
	const orgID = "test-org-id"
	const domain = "127.0.0.1"

	srv := httptest.NewServer(accessibilityHandler(domain, orgID, "FAIL"))
	defer srv.Close()

	port := srv.Listener.Addr().(*net.TCPAddr).Port
	setAccessibilityConfig(t, "http", strconv.Itoa(port))

	ok, err := IsDomainAccessible(domain, orgID)
	CheckTestBool(t, false, ok)
	CheckTestBool(t, false, err == nil)
}

// TestIsDomainAccessibleUnreachable verifies that IsDomainAccessible returns false
// when no server is reachable on any of the attempted scheme/port combinations.
func TestIsDomainAccessibleUnreachable(t *testing.T) {
	// Grab a free port then immediately release it so nothing is listening there.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()

	setAccessibilityConfig(t, "http", strconv.Itoa(port))

	ok, err := IsDomainAccessible("127.0.0.1", "test-org-id")
	CheckTestBool(t, false, ok)
	CheckTestBool(t, false, err == nil)
}
