package test

import (
	"os"
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
