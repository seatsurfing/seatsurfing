package test

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	. "github.com/seatsurfing/seatsurfing/server/service"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func signTestIDToken(t *testing.T, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	raw, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func validTestIDTokenClaims(now time.Time) jwt.MapClaims {
	return jwt.MapClaims{
		"sub":         "subject-1",
		"aud":         "client-1",
		"exp":         now.Add(time.Hour).Unix(),
		"email":       "user@example.com",
		"given_name":  "Jane",
		"family_name": "Doe",
		"groups":      []string{"staff"},
	}
}

func checkTestIDTokenError(t *testing.T, err error, contains string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected an error containing %q", contains)
	}
	if !strings.Contains(err.Error(), contains) {
		t.Fatalf("expected an error containing %q, got %q", contains, err.Error())
	}
}

func TestOIDCClaimsFillsMissingFields(t *testing.T) {
	now := time.Now()
	userinfo := map[string]interface{}{"sub": "subject-1"}
	raw := signTestIDToken(t, validTestIDTokenClaims(now))
	filled, err := FillMissingUserInfoFromIDToken(userinfo, raw, "client-1", now, "email", "given_name", "family_name", "groups")
	if err != nil {
		t.Fatal(err)
	}
	CheckTestString(t, "email,given_name,family_name,groups", strings.Join(filled, ","))
	CheckTestString(t, "user@example.com", userinfo["email"].(string))
	CheckTestString(t, "Jane", userinfo["given_name"].(string))
	CheckTestString(t, "Doe", userinfo["family_name"].(string))
	CheckTestInt(t, 1, len(userinfo["groups"].([]interface{})))
}

func TestOIDCClaimsAudienceArray(t *testing.T) {
	now := time.Now()
	claims := validTestIDTokenClaims(now)
	claims["aud"] = []string{"other", "client-1"}
	userinfo := map[string]interface{}{"sub": "subject-1"}
	if _, err := FillMissingUserInfoFromIDToken(userinfo, signTestIDToken(t, claims), "client-1", now, "email"); err != nil {
		t.Fatal(err)
	}
	CheckTestString(t, "user@example.com", userinfo["email"].(string))
}

func TestOIDCClaimsWrongAudience(t *testing.T) {
	now := time.Now()
	claims := validTestIDTokenClaims(now)
	claims["aud"] = "other-client"
	userinfo := map[string]interface{}{"sub": "subject-1"}
	_, err := FillMissingUserInfoFromIDToken(userinfo, signTestIDToken(t, claims), "client-1", now, "email")
	checkTestIDTokenError(t, err, "audience")
	CheckTestBool(t, false, userinfo["email"] != nil)
}

func TestOIDCClaimsExpiryLeeway(t *testing.T) {
	now := time.Now()
	claims := validTestIDTokenClaims(now)
	claims["exp"] = now.Add(-30 * time.Second).Unix()
	userinfo := map[string]interface{}{"sub": "subject-1"}
	if _, err := FillMissingUserInfoFromIDToken(userinfo, signTestIDToken(t, claims), "client-1", now, "email"); err != nil {
		t.Fatalf("expected small clock skew to be tolerated: %v", err)
	}
	CheckTestString(t, "user@example.com", userinfo["email"].(string))
}

func TestOIDCClaimsExpired(t *testing.T) {
	now := time.Now()
	claims := validTestIDTokenClaims(now)
	claims["exp"] = now.Add(-2 * time.Minute).Unix()
	userinfo := map[string]interface{}{"sub": "subject-1"}
	_, err := FillMissingUserInfoFromIDToken(userinfo, signTestIDToken(t, claims), "client-1", now, "email")
	checkTestIDTokenError(t, err, "expired")
	CheckTestBool(t, false, userinfo["email"] != nil)
}

func TestOIDCClaimsMissingExpiry(t *testing.T) {
	now := time.Now()
	claims := validTestIDTokenClaims(now)
	delete(claims, "exp")
	userinfo := map[string]interface{}{"sub": "subject-1"}
	_, err := FillMissingUserInfoFromIDToken(userinfo, signTestIDToken(t, claims), "client-1", now, "email")
	checkTestIDTokenError(t, err, "expiry")
}

func TestOIDCClaimsMalformedToken(t *testing.T) {
	userinfo := map[string]interface{}{"sub": "subject-1"}
	_, err := FillMissingUserInfoFromIDToken(userinfo, "not-a-jwt", "client-1", time.Now(), "email")
	checkTestIDTokenError(t, err, "failed parsing ID token")
}

func TestOIDCClaimsNoIDToken(t *testing.T) {
	userinfo := map[string]interface{}{"sub": "subject-1"}
	_, err := FillMissingUserInfoFromIDToken(userinfo, "", "client-1", time.Now(), "email")
	checkTestIDTokenError(t, err, "openid")
}

func TestOIDCClaimsKeepsUserinfoValues(t *testing.T) {
	now := time.Now()
	userinfo := map[string]interface{}{"sub": "subject-1", "email": "userinfo@example.com"}
	raw := signTestIDToken(t, validTestIDTokenClaims(now))
	filled, err := FillMissingUserInfoFromIDToken(userinfo, raw, "client-1", now, "email", "given_name")
	if err != nil {
		t.Fatal(err)
	}
	CheckTestString(t, "given_name", strings.Join(filled, ","))
	CheckTestString(t, "userinfo@example.com", userinfo["email"].(string))
	CheckTestString(t, "Jane", userinfo["given_name"].(string))
}

func TestOIDCClaimsEmptyUserinfoValuesAreMissing(t *testing.T) {
	now := time.Now()
	userinfo := map[string]interface{}{
		"sub":         "subject-1",
		"email":       "",
		"given_name":  "   ",
		"family_name": nil,
		"groups":      []interface{}{},
	}
	raw := signTestIDToken(t, validTestIDTokenClaims(now))
	filled, err := FillMissingUserInfoFromIDToken(userinfo, raw, "client-1", now, "email", "given_name", "family_name", "groups")
	if err != nil {
		t.Fatal(err)
	}
	CheckTestString(t, "email,given_name,family_name,groups", strings.Join(filled, ","))
	CheckTestString(t, "user@example.com", userinfo["email"].(string))
	CheckTestString(t, "Jane", userinfo["given_name"].(string))
	CheckTestString(t, "Doe", userinfo["family_name"].(string))
	CheckTestInt(t, 1, len(userinfo["groups"].([]interface{})))
}

func TestOIDCClaimsEmptyIDTokenValuesAreNotCopied(t *testing.T) {
	now := time.Now()
	claims := validTestIDTokenClaims(now)
	claims["given_name"] = " "
	claims["groups"] = []string{}
	delete(claims, "family_name")
	userinfo := map[string]interface{}{"sub": "subject-1", "given_name": ""}
	filled, err := FillMissingUserInfoFromIDToken(userinfo, signTestIDToken(t, claims), "client-1", now, "email", "given_name", "family_name", "groups")
	if err != nil {
		t.Fatal(err)
	}
	CheckTestString(t, "email", strings.Join(filled, ","))
	CheckTestString(t, "", userinfo["given_name"].(string))
	_, hasFamilyName := userinfo["family_name"]
	CheckTestBool(t, false, hasFamilyName)
	_, hasGroups := userinfo["groups"]
	CheckTestBool(t, false, hasGroups)
}

func TestOIDCClaimsOnlyCopiesGivenFields(t *testing.T) {
	now := time.Now()
	userinfo := map[string]interface{}{"sub": "subject-1"}
	raw := signTestIDToken(t, validTestIDTokenClaims(now))
	if _, err := FillMissingUserInfoFromIDToken(userinfo, raw, "client-1", now, "email", ""); err != nil {
		t.Fatal(err)
	}
	CheckTestInt(t, 2, len(userinfo))
	for _, claim := range []string{"aud", "exp", "given_name", ""} {
		if _, ok := userinfo[claim]; ok {
			t.Fatalf("claim %q must not be merged", claim)
		}
	}
}

func TestOIDCClaimsNothingMissingSkipsIDToken(t *testing.T) {
	userinfo := map[string]interface{}{"sub": "subject-1", "email": "userinfo@example.com"}
	filled, err := FillMissingUserInfoFromIDToken(userinfo, "not-a-jwt", "client-1", time.Now(), "email", "")
	if err != nil {
		t.Fatalf("expected the ID token to be ignored if no field is missing: %v", err)
	}
	CheckTestInt(t, 0, len(filled))
}

func TestOIDCClaimsSubjectMismatch(t *testing.T) {
	now := time.Now()
	claims := validTestIDTokenClaims(now)
	claims["sub"] = "subject-2"
	userinfo := map[string]interface{}{"sub": "subject-1"}
	_, err := FillMissingUserInfoFromIDToken(userinfo, signTestIDToken(t, claims), "client-1", now, "email")
	checkTestIDTokenError(t, err, "subject")
	CheckTestBool(t, false, userinfo["email"] != nil)
}

func TestOIDCClaimsMissingSubject(t *testing.T) {
	now := time.Now()
	withoutSub := validTestIDTokenClaims(now)
	delete(withoutSub, "sub")
	for name, testCase := range map[string]struct {
		userinfo map[string]interface{}
		claims   jwt.MapClaims
	}{
		"userinfo": {map[string]interface{}{}, validTestIDTokenClaims(now)},
		"id-token": {map[string]interface{}{"sub": "subject-1"}, withoutSub},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := FillMissingUserInfoFromIDToken(testCase.userinfo, signTestIDToken(t, testCase.claims), "client-1", now, "email")
			checkTestIDTokenError(t, err, "no subject")
			CheckTestBool(t, false, testCase.userinfo["email"] != nil)
		})
	}
}
