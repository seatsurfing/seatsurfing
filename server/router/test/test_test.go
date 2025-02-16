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
	TestRunner(m)
}
