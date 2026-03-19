package util

import (
	"net/url"
	"strconv"
)

func ValidateURL(s string) bool {
	if len([]rune(s)) > 256 {
		return false
	}
	u, err := url.ParseRequestURI(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func ValidateNumber(s string, min, max int) bool {
	n, err := strconv.Atoi(s)
	if err != nil {
		return false
	}
	if strconv.Itoa(n) != s {
		return false
	}
	return n >= min && n <= max
}
