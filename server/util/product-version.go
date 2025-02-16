package util

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/seatsurfing/seatsurfing/server/config"
)

var _productVersion = ""

func GetProductVersion() string {
	if _productVersion == "" {
		var path string
		if config.GetConfig().Development {
			path, _ = filepath.Abs("../version.txt")
		} else {
			path, _ = filepath.Abs("./version.txt")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return "UNKNOWN"
		}
		_productVersion = strings.TrimSpace(string(data))
	}
	return _productVersion
}
