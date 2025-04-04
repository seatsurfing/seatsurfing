package testutil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	. "github.com/seatsurfing/seatsurfing/server/app"
	. "github.com/seatsurfing/seatsurfing/server/config"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/router"
)

type LoginResponse struct {
	RequireOTP   bool   `json:"otpRequired"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	UserID       string `json:"userId"`
}

func GetTestJWT(userID string) string {
	claims := &Claims{
		Email:  userID,
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(60 * 24 * 14 * time.Minute)),
		},
	}
	router := &AuthRouter{}
	accessToken := router.CreateAccessToken(claims)
	return accessToken
}

func NewHTTPRequest(method, url, userID string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, url, body)
	if userID != "" {
		req.Header.Set("Authorization", "Bearer "+GetTestJWT(userID))
	}
	return req
}

func NewHTTPRequestWithAccessToken(method, url, accessToken string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, url, body)
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	return req
}

func CreateTestUser(orgDomain string) *User {
	return CreateTestUserParams(orgDomain)
}

func CreateTestUserParams(orgDomain string) *User {
	org := CreateTestOrg(orgDomain)
	user := &User{
		Email:          uuid.New().String() + "@" + orgDomain,
		OrganizationID: org.ID,
		Role:           UserRoleUser,
	}
	if err := GetUserRepository().Create(user); err != nil {
		panic(err)
	}
	return user
}

func CreateTestUserSuperAdmin() *User {
	org := CreateTestOrg("test.com")
	user := &User{
		Email:          uuid.New().String() + "@test.com",
		OrganizationID: org.ID,
		Role:           UserRoleSuperAdmin,
	}
	if err := GetUserRepository().Create(user); err != nil {
		panic(err)
	}
	return user
}

func CreateTestOrg(orgDomain string) *Organization {
	org := &Organization{
		Name:             "Test Org",
		ContactEmail:     "foo@seatsurfing.app",
		ContactFirstname: "Foo",
		ContactLastname:  "Bar",
		Language:         "de",
		SignupDate:       time.Now(),
	}
	if err := GetOrganizationRepository().Create(org); err != nil {
		panic(err)
	}
	if err := GetOrganizationRepository().AddDomain(org, orgDomain, true); err != nil {
		panic(err)
	}
	if err := GetOrganizationRepository().SetPrimaryDomain(org, orgDomain); err != nil {
		panic(err)
	}
	return org
}

func CreateTestUserInOrgWithName(org *Organization, email string, role UserRole) *User {
	user := &User{
		Email:          email,
		OrganizationID: org.ID,
		Role:           role,
	}
	if err := GetUserRepository().Create(user); err != nil {
		panic(err)
	}
	return user
}

func CreateTestUserInOrgDomain(org *Organization, domain string) *User {
	return CreateTestUserInOrgWithName(org, uuid.New().String()+"@"+domain, UserRoleUser)
}

func CreateTestUserInOrg(org *Organization) *User {
	return CreateTestUserInOrgDomain(org, "test.com")
}

func CreateTestUserOrgAdminDomain(org *Organization, domain string) *User {
	user := &User{
		Email:          uuid.New().String() + "@" + domain,
		OrganizationID: org.ID,
		Role:           UserRoleOrgAdmin,
	}
	if err := GetUserRepository().Create(user); err != nil {
		panic(err)
	}
	return user
}

func CreateTestUserOrgAdmin(org *Organization) *User {
	return CreateTestUserOrgAdminDomain(org, "test.com")
}

func LoginTestUserParams(userID string) *LoginResponse {
	// TODO
	res := &LoginResponse{
		AccessToken:  "abc",
		RefreshToken: "def",
		RequireOTP:   false,
		UserID:       userID,
	}
	return res
}

func LoginTestUser(userID string) *LoginResponse {
	return LoginTestUserParams(userID)
}

func CreateLoginTestUser() *LoginResponse {
	user := CreateTestUser("test.com")
	return LoginTestUser(user.ID)
}

func CreateLoginTestUserParams() *LoginResponse {
	user := CreateTestUserParams("test.com")
	return LoginTestUserParams(user.ID)
}

func DropTestDB() {
	tables := []string{"auth_providers", "auth_states", "bookings", "spaces", "locations", "organizations_domains", "organizations", "users", "signups", "settings", "space_attributes"}
	for _, s := range tables {
		GetDatabase().DB().Exec("DROP TABLE IF EXISTS " + s)
	}
}

func ClearTestDB() {
	tables := []string{"auth_providers", "auth_states", "auth_attempts", "bookings", "spaces", "locations", "organizations_domains", "organizations", "users", "users_preferences", "signups", "settings", "space_attributes"}
	for _, s := range tables {
		GetDatabase().DB().Exec("TRUNCATE " + s)
	}
}

func ExecuteTestRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	GetApp().Router.ServeHTTP(rr, req)
	return rr
}

func CheckTestResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Fatalf("Expected HTTP Status %d, but got %d at:\n%s", expected, actual, debug.Stack())
	}
}

func CheckTestString(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Fatalf("Expected '%s', but got '%s' at:\n%s", expected, actual, debug.Stack())
	}
}

func CheckTestBool(t *testing.T, expected, actual bool) {
	if expected != actual {
		t.Fatalf("Expected '%t', but got '%t' at:\n%s", expected, actual, debug.Stack())
	}
}

func CheckTestUint(t *testing.T, expected, actual uint) {
	if expected != actual {
		t.Fatalf("Expected '%d', but got '%d' at:\n%s", expected, actual, debug.Stack())
	}
}

func CheckTestInt(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Fatalf("Expected '%d', but got '%d' at:\n%s", expected, actual, debug.Stack())
	}
}

func CheckStringNotEmpty(t *testing.T, s string) {
	if strings.TrimSpace(s) == "" {
		t.Fatalf("Expected non-empty string at:\n%s", debug.Stack())
	}
}

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func AuthAttemptRepositoryIsUserDisabled(t *testing.T, userID string) bool {
	user, err := GetUserRepository().GetOne(userID)
	if err != nil {
		t.Error(err)
	}
	return user.Disabled
}

func TestRunner(m *testing.M) {
	if os.Getenv("POSTGRES_URL") == "" {
		os.Setenv("POSTGRES_URL", "postgres://postgres:root@localhost/seatsurfing_test?sslmode=disable")
	}
	os.Setenv("MOCK_SENDMAIL", "1")
	os.Setenv("ALLOW_ORG_DELETE", "1")
	os.Setenv("LOGIN_PROTECTION_MAX_FAILS", "3")
	GetConfig().ReadConfig()
	db := GetDatabase()
	DropTestDB()
	a := GetApp()
	a.InitializeDatabases()
	a.InitializeRouter()
	code := m.Run()
	DropTestDB()
	db.Close()
	os.Exit(code)
}
