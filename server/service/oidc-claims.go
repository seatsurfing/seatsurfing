package service

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// idTokenExpiryLeeway tolerates small clock differences between Seatsurfing
// and the identity provider when checking the expiry of an ID token.
const idTokenExpiryLeeway = 60 * time.Second

// FillMissingUserInfoFromIDToken copies the given fields from the claims of
// an OIDC ID token into a userinfo response if they are missing there. A
// field counts as missing if it is absent, null, an empty or whitespace-only
// string or an empty array. Values present in the userinfo response are never
// overwritten. It returns the names of the fields taken from the ID token.
//
// Some identity providers (e.g. AD FS) return only "sub" from the userinfo
// endpoint and put profile claims such as email and names into the ID token.
// The ID token is only returned if the provider's scopes include "openid",
// and the configured field names must match the claim names in the ID token
// (e.g. "upn" for AD FS).
//
// The ID token is only inspected if at least one of the given fields is
// missing. Per OIDC Core 1.0 section 5.3.2, it is only used if its "sub" claim
// matches the "sub" claim of the userinfo response, so that claims of
// different identities are never mixed. Only the given fields are copied, so
// that technical claims such as "aud" or "exp" never end up in the userinfo
// data.
func FillMissingUserInfoFromIDToken(userinfo map[string]interface{}, rawIDToken string, clientID string, now time.Time, fields ...string) ([]string, error) {
	missing := missingClaims(userinfo, fields)
	if len(missing) == 0 {
		return nil, nil
	}
	if rawIDToken == "" {
		return nil, errors.New("token response contains no ID token (do the provider's scopes include openid?)")
	}
	idClaims, err := claimsFromIDToken(rawIDToken, clientID, now)
	if err != nil {
		return nil, err
	}
	userinfoSub, _ := userinfo["sub"].(string)
	idTokenSub, _ := idClaims["sub"].(string)
	if userinfoSub == "" {
		return nil, errors.New("userinfo response has no subject (sub)")
	}
	if idTokenSub == "" {
		return nil, errors.New("ID token has no subject (sub)")
	}
	if userinfoSub != idTokenSub {
		return nil, errors.New("subject of ID token does not match subject of userinfo response")
	}
	return mergeMissingClaims(userinfo, idClaims, missing), nil
}

// claimsFromIDToken decodes the claims of an OIDC ID token that was received
// directly from the token endpoint in the authorization code flow. Per OIDC
// Core 1.0 section 3.1.3.7, the TLS server validation of the token endpoint
// may be used instead of checking the token signature in this case. Audience
// and expiry are still checked so that a token issued for another client or
// an outdated token is never used.
func claimsFromIDToken(rawIDToken string, clientID string, now time.Time) (map[string]interface{}, error) {
	claims := jwt.MapClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(rawIDToken, claims); err != nil {
		return nil, fmt.Errorf("failed parsing ID token: %w", err)
	}
	aud, err := claims.GetAudience()
	if err != nil || !slices.Contains(aud, clientID) {
		return nil, errors.New("ID token audience does not match client ID")
	}
	exp, err := claims.GetExpirationTime()
	if err != nil || exp == nil {
		return nil, errors.New("ID token has no valid expiry")
	}
	if !now.Before(exp.Time.Add(idTokenExpiryLeeway)) {
		return nil, fmt.Errorf("ID token expired at %s", exp.Time.UTC().Format(time.RFC3339))
	}
	return claims, nil
}

// missingClaims returns the non-empty field names whose values are missing in
// claims, without duplicates.
func missingClaims(claims map[string]interface{}, fields []string) []string {
	var missing []string
	for _, field := range fields {
		if field == "" || slices.Contains(missing, field) {
			continue
		}
		if isMissingClaim(claims[field]) {
			missing = append(missing, field)
		}
	}
	return missing
}

// mergeMissingClaims copies the given fields from claims into target if they
// are missing there and present in claims. It returns the copied fields.
func mergeMissingClaims(target map[string]interface{}, claims map[string]interface{}, fields []string) []string {
	var merged []string
	for _, field := range fields {
		if !isMissingClaim(target[field]) || isMissingClaim(claims[field]) {
			continue
		}
		target[field] = claims[field]
		merged = append(merged, field)
	}
	return merged
}

// isMissingClaim reports whether a claim value carries no information: nil
// (absent or null), an empty or whitespace-only string or an empty array.
func isMissingClaim(value interface{}) bool {
	switch v := value.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(v) == ""
	case []interface{}:
		return len(v) == 0
	case []string:
		return len(v) == 0
	}
	return false
}
