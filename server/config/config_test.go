package config_test

import (
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/config"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestIsValidLanguageCode_lowercase(t *testing.T) {
	c := &Config{}
	CheckTestBool(t, true, c.IsValidLanguageCode("en"))
}

func TestIsValidLanguageCode_uppercase(t *testing.T) {
	c := &Config{}
	CheckTestBool(t, false, c.IsValidLanguageCode("EN"))
}
