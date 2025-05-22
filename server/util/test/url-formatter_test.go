package test

import (
	"os"
	"testing"

	"github.com/seatsurfing/seatsurfing/server/config"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

func TestURLFormatter(t *testing.T) {
	CheckTestString(t, "https://example.com", FormatURL("example.com"))

	os.Setenv("PUBLIC_SCHEME", "http")
	os.Setenv("PUBLIC_PORT", "80")
	config.GetConfig().ReadConfig()
	CheckTestString(t, "http://example.com", FormatURL("example.com"))

	os.Setenv("PUBLIC_SCHEME", "http")
	os.Setenv("PUBLIC_PORT", "8080")
	config.GetConfig().ReadConfig()
	CheckTestString(t, "http://example.com:8080", FormatURL("example.com"))

	os.Setenv("PUBLIC_SCHEME", "https")
	os.Setenv("PUBLIC_PORT", "8443")
	config.GetConfig().ReadConfig()
	CheckTestString(t, "https://example.com:8443", FormatURL("example.com"))
}
