package test

import (
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
	. "github.com/seatsurfing/seatsurfing/server/testutil"
)

func TestConfluenceLoginInvalidOrg(t *testing.T) {
	ClearTestDB()

	invalidOrgID := uuid.New().String()
	req := NewHTTPRequest("GET", "/confluence/"+invalidOrgID+"/some-jwt-token", "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusNotFound, res.Code)
}

func TestConfluenceLoginNoSharedSecret(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")

	req := NewHTTPRequest("GET", "/confluence/"+org.ID+"/some-jwt-token", "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusBadRequest, res.Code)
}

func TestConfluenceLoginInvalidJWT(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	sharedSecret := "test-shared-secret-12345"
	GetSettingsRepository().Set(org.ID, SettingConfluenceServerSharedSecret.Name, sharedSecret)

	invalidJWT := "invalid.jwt.token"
	req := NewHTTPRequest("GET", "/confluence/"+org.ID+"/"+invalidJWT, "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusTemporaryRedirect, res.Code)
}

func TestConfluenceLogin(t *testing.T) {
	ClearTestDB()
	org := CreateTestOrg("test.com")
	sharedSecret := "test-shared-secret-12345"
	GetSettingsRepository().Set(org.ID, SettingConfluenceServerSharedSecret.Name, sharedSecret)

	claims := &ConfluenceServerClaims{
		UserName: "testuser@test.com",
		UserKey:  "testuser-key",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(sharedSecret))
	if err != nil {
		t.Fatalf("Failed to create JWT: %v", err)
	}

	req := NewHTTPRequest("GET", "/confluence/"+org.ID+"/"+tokenString, "", nil)
	res := ExecuteTestRequest(req)
	CheckTestResponseCode(t, http.StatusTemporaryRedirect, res.Code)
	location := res.Header().Get("Location")
	if location == "" {
		t.Fatal("Expected Location header in redirect response")
	}
}
