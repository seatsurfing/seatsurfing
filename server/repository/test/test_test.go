package test

import (
	"os"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestMain(m *testing.M) {
	// Set a 32-byte CRYPT_KEY so passkey encryption tests work
	os.Setenv("CRYPT_KEY", "12345678901234567890123456789012")
	TestRunner(m)
}
