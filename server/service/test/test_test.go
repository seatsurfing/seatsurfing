package test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestMain(m *testing.M) {
	pwd, _ := os.Getwd()
	os.Setenv("FILESYSTEM_BASE_PATH", filepath.Join(pwd, "../../"))
	os.Setenv("CRYPT_KEY", "12345678901234567890123456789012")
	TestRunner(m)
}
