package test

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
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
