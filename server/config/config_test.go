package config_test

import (
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/config"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestIsValidLanguageCodeLowercase(t *testing.T) {
	c := &Config{}
	CheckTestBool(t, true, c.IsValidLanguageCode("en"))
}

func TestIsValidLanguageCodeUppercase(t *testing.T) {
	c := &Config{}
	CheckTestBool(t, false, c.IsValidLanguageCode("EN"))
}

func TestPrivateOutboundBlockedByDefault(t *testing.T) {
	t.Setenv("CRYPT_KEY", "12345678901234567890123456789012")
	t.Setenv("FEATURE_FLAGS", "")
	c := &Config{}
	c.ReadConfig()
	CheckTestBool(t, false, c.AllowPrivateOutbound)
}

func TestPrivateOutboundAllowedWithFeatureFlag(t *testing.T) {
	t.Setenv("CRYPT_KEY", "12345678901234567890123456789012")
	t.Setenv("FEATURE_FLAGS", "ALLOW_PRIVATE_OUTBOUND")
	c := &Config{}
	c.ReadConfig()
	CheckTestBool(t, true, c.AllowPrivateOutbound)
}

func TestAllowOrgDeleteEnvVarIgnored(t *testing.T) {
	t.Setenv("CRYPT_KEY", "12345678901234567890123456789012")
	t.Setenv("FEATURE_FLAGS", "")
	t.Setenv("ALLOW_ORG_DELETE", "1")
	c := &Config{}
	c.ReadConfig()
	CheckTestBool(t, false, c.AllowOrgDelete)
}
