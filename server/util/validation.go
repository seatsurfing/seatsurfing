package util

import (
	"net/url"
	"regexp"
	"strconv"
	"unicode"
)

var colorHexRegex = regexp.MustCompile(`^#([0-9A-Fa-f]{6}|[0-9A-Fa-f]{3})$`)
var guidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
var domainRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9\-]{0,61}[a-z0-9])?(\.[a-z0-9]([a-z0-9\-]{0,61}[a-z0-9])?)+$`)
var humanNameRegex = regexp.MustCompile(`^[\p{L}\p{N} \-'.]+$`)
var orgNameRegex = regexp.MustCompile(`^[\p{L}\p{N} \-'.&,+()/#@!_<>]+$`)
var validOrgLanguages = map[string]bool{"de": true, "en": true}

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

func ValidateColorHex(s string) bool {
	return colorHexRegex.MatchString(s)
}

func ValidateGUID(s string) bool {
	return guidRegex.MatchString(s)
}

func ValidateDomain(s string) bool {
	if len(s) > 253 {
		return false
	}
	return domainRegex.MatchString(s)
}

func IsValidOrgLanguage(s string) bool {
	return validOrgLanguages[s]
}

func IsValidHumanName(s string) bool {
	l := len([]rune(s))
	if l < 2 || l > 64 {
		return false
	}
	return humanNameRegex.MatchString(s)
}

func IsValidOrgName(s string) bool {
	l := len([]rune(s))
	if l < 2 || l > 64 {
		return false
	}
	return orgNameRegex.MatchString(s)
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
