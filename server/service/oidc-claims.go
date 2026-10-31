package service

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// FillMissingUserInfoFromIDToken copies the given fields from the claims of
// an OIDC ID token into a userinfo response if they are missing there.
// Existing userinfo values are never overwritten.
//
// Some identity providers (e.g. AD FS) return only "sub" from the userinfo
// endpoint and put profile claims such as email and names into the ID token.
//
// Per OIDC Core 1.0 section 5.3.2, the ID token is only used if its "sub"
// claim matches the "sub" claim of the userinfo response, so that claims of
// different identities are never mixed. Only the given fields are copied, so
// that technical claims such as "aud" or "exp" never end up in the userinfo
// data.
func FillMissingUserInfoFromIDToken(userinfo map[string]interface{}, rawIDToken string, clientID string, now time.Time, fields ...string) error {
	idClaims, err := ClaimsFromIDToken(rawIDToken, clientID, now)
	if err != nil {
		return err
	}
	userinfoSub, _ := userinfo["sub"].(string)
	idTokenSub, _ := idClaims["sub"].(string)
	if userinfoSub == "" || idTokenSub == "" {
		return errors.New("userinfo response or ID token has no subject")
	}
	if userinfoSub != idTokenSub {
		return errors.New("subject of ID token does not match subject of userinfo response")
	}
	MergeMissingClaims(userinfo, idClaims, fields...)
	return nil
}

// ClaimsFromIDToken decodes the claims of an OIDC ID token that was received
// directly from the token endpoint in the authorization code flow. Per OIDC
// Core 1.0 section 3.1.3.7, the TLS server validation of the token endpoint
// may be used instead of checking the token signature in this case. Audience
// and expiry are still checked so that a token issued for another client or
// an outdated token is never used.
func ClaimsFromIDToken(rawIDToken string, clientID string, now time.Time) (map[string]interface{}, error) {
	claims := jwt.MapClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(rawIDToken, claims); err != nil {
		return nil, fmt.Errorf("failed parsing ID token: %w", err)
	}
	aud, err := claims.GetAudience()
	if err != nil || !slices.Contains(aud, clientID) {
		return nil, errors.New("ID token audience does not match client ID")
	}
	exp, err := claims.GetExpirationTime()
	if err != nil || exp == nil || !now.Before(exp.Time) {
		return nil, errors.New("ID token is expired or has no expiry")
	}
	return claims, nil
}

// MergeMissingClaims copies the given fields from claims into target if
// target does not have them yet. Existing values in target are never
// overwritten. Empty field names are ignored.
func MergeMissingClaims(target map[string]interface{}, claims map[string]interface{}, fields ...string) {
	for _, field := range fields {
		if field == "" {
			continue
		}
		if _, exists := target[field]; exists {
			continue
		}
		if v, ok := claims[field]; ok {
			target[field] = v
		}
	}
}
