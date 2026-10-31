package test

import (
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

func TestClaimsFromIDTokenHappyPath(t *testing.T) {
	now := time.Now()
	raw := signTestIDToken(t, jwt.MapClaims{
		"aud":   "client-1",
		"exp":   now.Add(time.Hour).Unix(),
		"email": "user@example.com",
	})
	claims, err := ClaimsFromIDToken(raw, "client-1", now)
	if err != nil {
		t.Fatal(err)
	}
	CheckTestString(t, "user@example.com", claims["email"].(string))
}

func TestClaimsFromIDTokenAudienceArray(t *testing.T) {
	now := time.Now()
	raw := signTestIDToken(t, jwt.MapClaims{
		"aud": []string{"other", "client-1"},
		"exp": now.Add(time.Hour).Unix(),
	})
	if _, err := ClaimsFromIDToken(raw, "client-1", now); err != nil {
		t.Fatal(err)
	}
}

func TestClaimsFromIDTokenWrongAudience(t *testing.T) {
	now := time.Now()
	raw := signTestIDToken(t, jwt.MapClaims{
		"aud": "other-client",
		"exp": now.Add(time.Hour).Unix(),
	})
	if _, err := ClaimsFromIDToken(raw, "client-1", now); err == nil {
		t.Fatal("expected error for wrong audience")
	}
}

func TestClaimsFromIDTokenExpired(t *testing.T) {
	now := time.Now()
	raw := signTestIDToken(t, jwt.MapClaims{
		"aud": "client-1",
		"exp": now.Add(-time.Minute).Unix(),
	})
	if _, err := ClaimsFromIDToken(raw, "client-1", now); err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestClaimsFromIDTokenMissingExpiry(t *testing.T) {
	raw := signTestIDToken(t, jwt.MapClaims{"aud": "client-1"})
	if _, err := ClaimsFromIDToken(raw, "client-1", time.Now()); err == nil {
		t.Fatal("expected error for missing expiry")
	}
}

func TestClaimsFromIDTokenMalformed(t *testing.T) {
	if _, err := ClaimsFromIDToken("not-a-jwt", "client-1", time.Now()); err == nil {
		t.Fatal("expected error for malformed token")
	}
}

func TestMergeMissingClaimsKeepsUserinfoValues(t *testing.T) {
	target := map[string]interface{}{
		"sub":   "abc",
		"email": "userinfo@example.com",
	}
	MergeMissingClaims(target, map[string]interface{}{
		"email":       "idtoken@example.com",
		"given_name":  "Jane",
		"family_name": "Doe",
	}, "email", "given_name", "family_name")
	CheckTestString(t, "userinfo@example.com", target["email"].(string))
	CheckTestString(t, "Jane", target["given_name"].(string))
	CheckTestString(t, "Doe", target["family_name"].(string))
	CheckTestString(t, "abc", target["sub"].(string))
}

func TestMergeMissingClaimsOnlyCopiesGivenFields(t *testing.T) {
	target := map[string]interface{}{"sub": "abc"}
	MergeMissingClaims(target, map[string]interface{}{
		"aud":        "client-1",
		"exp":        int64(123),
		"nonce":      "n",
		"given_name": "Jane",
	}, "given_name", "", "missing_field")
	CheckTestInt(t, 2, len(target))
	CheckTestString(t, "Jane", target["given_name"].(string))
	if _, ok := target["missing_field"]; ok {
		t.Fatal("field missing in claims must not be added")
	}
}

func TestFillMissingUserInfoFromIDTokenSubjectMismatch(t *testing.T) {
	now := time.Now()
	userinfo := map[string]interface{}{"sub": "user-a"}
	raw := signTestIDToken(t, jwt.MapClaims{
		"sub":   "user-b",
		"aud":   "client-1",
		"exp":   now.Add(time.Hour).Unix(),
		"email": "b@example.com",
	})
	if err := FillMissingUserInfoFromIDToken(userinfo, raw, "client-1", now, "email"); err == nil {
		t.Fatal("expected error for subject mismatch")
	}
	CheckTestInt(t, 1, len(userinfo))
}

func TestFillMissingUserInfoFromIDTokenMissingSubject(t *testing.T) {
	now := time.Now()
	raw := signTestIDToken(t, jwt.MapClaims{
		"aud":   "client-1",
		"exp":   now.Add(time.Hour).Unix(),
		"email": "a@example.com",
	})
	userinfo := map[string]interface{}{"sub": "user-a"}
	if err := FillMissingUserInfoFromIDToken(userinfo, raw, "client-1", now, "email"); err == nil {
		t.Fatal("expected error for ID token without subject")
	}
	userinfo = map[string]interface{}{}
	rawWithSub := signTestIDToken(t, jwt.MapClaims{
		"sub":   "user-a",
		"aud":   "client-1",
		"exp":   now.Add(time.Hour).Unix(),
		"email": "a@example.com",
	})
	if err := FillMissingUserInfoFromIDToken(userinfo, rawWithSub, "client-1", now, "email"); err == nil {
		t.Fatal("expected error for userinfo without subject")
	}
	CheckTestInt(t, 0, len(userinfo))
}

func TestFillMissingUserInfoFromIDTokenInvalidToken(t *testing.T) {
	now := time.Now()
	userinfo := map[string]interface{}{"sub": "user-a"}
	raw := signTestIDToken(t, jwt.MapClaims{
		"sub":   "user-a",
		"aud":   "other-client",
		"exp":   now.Add(time.Hour).Unix(),
		"email": "a@example.com",
	})
	if err := FillMissingUserInfoFromIDToken(userinfo, raw, "client-1", now, "email"); err == nil {
		t.Fatal("expected error for wrong audience")
	}
	CheckTestInt(t, 1, len(userinfo))
}
