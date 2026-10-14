package test

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/config"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

func TestNetutilHttpClientCustomResolver(t *testing.T) {
	os.Setenv("DNS_SERVER", "8.8.8.8")
	GetConfig().ReadConfig()
	client := GetHTTPClientWithCustomDNS(false)
	req, err := http.NewRequest(http.MethodGet, "https://www.google.com", nil)
	CheckTestBool(t, true, err == nil)
	res, err := client.Do(req)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, 200, res.StatusCode)
}

func TestNetutilIsBlockedOutboundIP(t *testing.T) {
	blocked := []string{"127.0.0.1", "10.1.2.3", "172.16.0.1", "192.168.1.1", "169.254.169.254", "100.64.0.1", "0.0.0.0", "0.1.2.3", "::1", "::", "fe80::1", "fd00::1", "::ffff:127.0.0.1", "::ffff:169.254.169.254", "64:ff9b::a00:1", "64:ff9b:1::a00:1", "224.0.0.1"}
	for _, s := range blocked {
		CheckTestBool(t, true, IsBlockedOutboundIP(net.ParseIP(s)))
	}
	allowed := []string{"8.8.8.8", "1.1.1.1", "2606:4700:4700::1111", "::ffff:8.8.8.8"}
	for _, s := range allowed {
		CheckTestBool(t, false, IsBlockedOutboundIP(net.ParseIP(s)))
	}
}

func TestNetutilOutboundHTTPClientBlocksPrivate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	GetConfig().BlockPrivateOutbound = true
	defer func() { GetConfig().BlockPrivateOutbound = false }()

	_, err := GetOutboundHTTPClient().Get(srv.URL)
	CheckTestBool(t, true, errors.Is(err, ErrBlockedOutboundAddress))
}

func TestNetutilOutboundHTTPClientAllowsPrivateWhenDisabled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	GetConfig().BlockPrivateOutbound = false

	res, err := GetOutboundHTTPClient().Get(srv.URL)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, http.StatusOK, res.StatusCode)
	res.Body.Close()
}

func TestNetutilInsecureSkipVerifyClientBlocksPrivate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	GetConfig().BlockPrivateOutbound = true
	defer func() { GetConfig().BlockPrivateOutbound = false }()

	_, err := GetHTTPClientInsecureSkipVerify().Get(srv.URL)
	CheckTestBool(t, true, errors.Is(err, ErrBlockedOutboundAddress))
}

func TestNetutilInsecureSkipVerifyClientAllowsPrivateWhenDisabled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	GetConfig().BlockPrivateOutbound = false

	res, err := GetHTTPClientInsecureSkipVerify().Get(srv.URL)
	CheckTestBool(t, true, err == nil)
	CheckTestInt(t, http.StatusOK, res.StatusCode)
	res.Body.Close()
}

func TestNetutilClientsLimitRedirects(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, srv.URL, http.StatusFound)
	}))
	defer srv.Close()
	GetConfig().BlockPrivateOutbound = false

	for _, client := range []*http.Client{GetHTTPClientInsecureSkipVerify(), GetOutboundHTTPClient()} {
		_, err := client.Get(srv.URL)
		CheckTestBool(t, true, err != nil && strings.Contains(err.Error(), "stopped after 5 redirects"))
	}
}

// TestNetutilNoUnguardedHTTPClients makes sure new outbound HTTP calls do not bypass
// the BLOCK_PRIVATE_OUTBOUND protection by using a bare http client. Requests to
// user-controlled destinations must use GetOutboundHTTPClient() or
// GetHTTPClientInsecureSkipVerify(). Add a file to the allowlist only if its
// destination is fixed or comes from server configuration.
func TestNetutilNoUnguardedHTTPClients(t *testing.T) {
	allowed := map[string]bool{
		"util/netutil.go":      true, // defines the guarded clients
		"util/check-update.go": true, // fixed update check URL
		"util/sendmail-acs.go": true, // host from server config
		"app/app.go":           true, // dev UI proxy target from server config
	}
	patterns := []string{"http.Client{", "http.DefaultClient", "http.Get(", "http.Post(", "http.PostForm(", "http.Head("}
	root := filepath.Join("..", "..")
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if d.Name() == "test" || d.Name() == "testutil" || d.Name() == "perftest" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(rel, ".go") || strings.HasSuffix(rel, "_test.go") || allowed[rel] {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, p := range patterns {
			if strings.Contains(string(content), p) {
				t.Errorf("%s uses %q; use GetOutboundHTTPClient() or GetHTTPClientInsecureSkipVerify() instead", rel, p)
			}
		}
		return nil
	})
	CheckTestBool(t, true, err == nil)
}
