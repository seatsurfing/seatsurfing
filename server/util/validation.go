package util

import (
	"net/url"
	"regexp"
	"strconv"
	"unicode"
)

func ValidatePassword(s string) bool {
	l := len([]rune(s))
	if l < 8 || l > 64 {
		return false
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range s {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case !unicode.IsLetter(r) && !unicode.IsDigit(r):
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasDigit && hasSpecial
}

func ValidateURL(s string) bool {
	if len([]rune(s)) > 256 {
		return false
	}
	u, err := url.ParseRequestURI(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}

var colorHexRegex = regexp.MustCompile(`^#([0-9A-Fa-f]{6}|[0-9A-Fa-f]{3})$`)

func ValidateColorHex(s string) bool {
	return colorHexRegex.MatchString(s)
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
