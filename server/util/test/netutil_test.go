package test

import (
	"net/http"
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
