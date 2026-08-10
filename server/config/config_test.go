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
