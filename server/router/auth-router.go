package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/config"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

type JWTResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	LongLived    bool   `json:"longLived"`
	LogoutURL    string `json:"logoutUrl"`
}

type Claims struct {
	Email      string `json:"email"`
	UserID     string `json:"userID"`
	SpaceAdmin bool   `json:"spaceAdmin"`
	OrgAdmin   bool   `json:"admin"`
	Role       int    `json:"role"`
	jwt.RegisteredClaims
}

type AuthPreflightRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type InitPasswordResetRequest struct {
	OrganizationID string `json:"organizationId" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
}

type CompletePasswordResetRequest struct {
	Password string `json:"password" validate:"required,min=8"`
}

type AuthPreflightResponse struct {
	Organization    *GetOrganizationResponse         `json:"organization"`
	AuthProviders   []*GetAuthProviderPublicResponse `json:"authProviders"`
	RequirePassword bool                             `json:"requirePassword"`
	BackendVersion  string                           `json:"backendVersion"`
	Domain          string                           `json:"domain"`
}

type AuthPasswordRequest struct {
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=8"`
	OrganizationID string `json:"organizationId" validate:"required"`
	LongLived      bool   `json:"longLived"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type AuthStateLoginPayload struct {
	UserID    string `json:"userId"`
	LoginType string `json:"type"`
	LongLived bool   `json:"longLived"`
	Redirect  string `json:"redirect,omitempty"`
}

type AuthRouter struct {
}

func (router *AuthRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/verify/{id}", router.verify).Methods("GET")
	s.HandleFunc("/{id}/login/{type}/{longLived}", router.login).Methods("GET")
	s.HandleFunc("/{id}/login/{type}", router.login).Methods("GET")
	s.HandleFunc("/{id}/callback", router.callback).Methods("GET")
	s.HandleFunc("/preflight", router.preflight).Methods("POST")
	s.HandleFunc("/login", router.loginPassword).Methods("POST")
	s.HandleFunc("/initpwreset", router.initPasswordReset).Methods("POST")
	s.HandleFunc("/pwreset/{id}", router.completePasswordReset).Methods("POST")
	s.HandleFunc("/refresh", router.refreshAccessToken).Methods("POST")
	s.HandleFunc("/singleorg", router.singleOrg).Methods("GET")
	s.HandleFunc("/org/{domain}", router.getOrgDetails).Methods("GET")
}

func (router *AuthRouter) getOrgDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if vars["domain"] == "" {
		SendBadRequest(w)
		return
	}
	org, err := GetOrganizationRepository().GetOneByDomain(vars["domain"])
	if err != nil || org == nil {
		SendNotFound(w)
		return
	}
	res := router.getPreflightResponseForOrg(org)
	if res == nil {
		SendInternalServerError(w)
		return
	}
	requirePassword, err := GetUserRepository().HasAnyUserInOrgPasswordSet(org.ID)
	if err != nil {
		SendInternalServerError(w)
		return
	}
	res.RequirePassword = requirePassword
	SendJSON(w, res)
}

func (router *AuthRouter) singleOrg(w http.ResponseWriter, r *http.Request) {
	numOrgs, err := GetOrganizationRepository().GetNumOrgs()
	if err != nil {
		SendInternalServerError(w)
		return
	}
	if numOrgs != 1 {
		SendNotFound(w)
		return
	}
	list, err := GetOrganizationRepository().GetAll()
	if err != nil {
		SendInternalServerError(w)
		return
	}
	if len(list) != 1 {
		SendInternalServerError(w)
		return
	}
	org := list[0]
	res := router.getPreflightResponseForOrg(org)
	if res == nil {
		SendInternalServerError(w)
		return
	}
	requirePassword, err := GetUserRepository().HasAnyUserInOrgPasswordSet(org.ID)
	if err != nil {
		SendInternalServerError(w)
		return
	}
	res.RequirePassword = requirePassword
	SendJSON(w, res)
}

func (router *AuthRouter) refreshAccessToken(w http.ResponseWriter, r *http.Request) {
	var m RefreshRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	refreshToken, err := GetRefreshTokenRepository().GetOne(m.RefreshToken)
	if err != nil || refreshToken == nil {
		SendNotFound(w)
		return
	}
	if refreshToken.Expiry.Before(time.Now()) {
		SendBadRequest(w)
		return
	}
	user, err := GetUserRepository().GetOne(refreshToken.UserID)
	if err != nil {
		SendNotFound(w)
		return
	}
	if user.Disabled {
		SendNotFound(w)
		return
	}
	claims := router.createClaims(user)
	longLived := refreshToken.Expiry.Sub(refreshToken.Created) > time.Duration(time.Minute*60*25)
	accessToken := router.CreateAccessToken(claims)
	newRefreshToken := router.createRefreshToken(claims, longLived)
	res := &JWTResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}
	GetRefreshTokenRepository().Delete(refreshToken)
	SendJSON(w, res)
}

func (router *AuthRouter) initPasswordReset(w http.ResponseWriter, r *http.Request) {
	var m InitPasswordResetRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	user, err := GetUserRepository().GetByEmail(m.OrganizationID, m.Email)
	if user == nil || err != nil {
		log.Printf("Password reset failed: user %s not found in org %s\n", m.Email, m.OrganizationID)
		SendUpdated(w)
		return
	}
	if user.HashedPassword == "" {
		SendUpdated(w)
		return
	}
	if user.Disabled {
		SendUpdated(w)
		return
	}
	org, err := GetOrganizationRepository().GetOne(user.OrganizationID)
	if org == nil || err != nil {
		SendUpdated(w)
		return
	}
	authState := &AuthState{
		AuthProviderID: GetSettingsRepository().GetNullUUID(),
		Expiry:         time.Now().Add(time.Hour * 1),
		AuthStateType:  AuthResetPasswordRequest,
		Payload:        user.ID,
	}
	GetAuthStateRepository().Create(authState)
	router.SendPasswordResetEmail(user, authState.ID, org)
	SendUpdated(w)
}

func (router *AuthRouter) completePasswordReset(w http.ResponseWriter, r *http.Request) {
	var m CompletePasswordResetRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	vars := mux.Vars(r)
	authState, err := GetAuthStateRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	if authState.AuthStateType != AuthResetPasswordRequest {
		SendNotFound(w)
		return
	}
	user, err := GetUserRepository().GetOne(authState.Payload)
	if user == nil || err != nil {
		SendNotFound(w)
		return
	}
	if user.HashedPassword == "" {
		SendNotFound(w)
		return
	}
	if user.Disabled {
		SendNotFound(w)
		return
	}
	user.HashedPassword = NullString(GetUserRepository().GetHashedPassword(m.Password))
	GetUserRepository().Update(user)
	GetAuthStateRepository().Delete(authState)
	SendUpdated(w)
}

func (router *AuthRouter) preflight(w http.ResponseWriter, r *http.Request) {
	var m AuthPreflightRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}

	// Check if user exists.
	// If so, return preflight response with requirePassword set to true if user has a password set.
	users, err := GetUserRepository().GetUsersWithEmail(m.Email)
	if err != nil {
		SendInternalServerError(w)
		return
	}
	var user *User = nil
	if len(users) > 0 {
		user = users[0]
	}
	if user != nil {
		org, _ := GetOrganizationRepository().GetOne(user.OrganizationID)
		res := router.getPreflightResponseForOrg(org)
		res.RequirePassword = (user.HashedPassword != "")
		SendJSON(w, res)
		return
	}

	// If the user doesn't exist, check if the email domain is associated with an organization.
	org := router.getOrgForEmail(m.Email)
	if org != nil {
		res := router.getPreflightResponseForOrg(org)
		SendJSON(w, res)
		return
	}

	// If neither user nor org for domain exists, return 404.
	SendNotFound(w)
}

func (router *AuthRouter) loginPassword(w http.ResponseWriter, r *http.Request) {
	var m AuthPasswordRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	user, err := GetUserRepository().GetByEmail(m.OrganizationID, m.Email)
	if err != nil {
		SendNotFound(w)
		return
	}
	if user.HashedPassword == "" {
		SendNotFound(w)
		return
	}
	if user.Disabled {
		SendNotFound(w)
		return
	}
	if !GetUserRepository().CheckPassword(string(user.HashedPassword), m.Password) {
		GetAuthAttemptRepository().RecordLoginAttempt(user, false)
		SendNotFound(w)
		return
	}
	GetAuthAttemptRepository().RecordLoginAttempt(user, true)
	claims := router.createClaims(user)
	accessToken := router.CreateAccessToken(claims)
	refreshToken := router.createRefreshToken(claims, m.LongLived)
	res := &JWTResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	SendJSON(w, res)
}

func (router *AuthRouter) handleAtlassianVerify(authState *AuthState, w http.ResponseWriter) {
	payload := unmarshalAuthStateLoginPayload(authState.Payload)
	user, err := GetUserRepository().GetByAtlassianID(payload.UserID)
	if err != nil {
		SendNotFound(w)
		return
	}
	if user.Disabled {
		SendNotFound(w)
		return
	}
	GetAuthStateRepository().Delete(authState)
	GetAuthAttemptRepository().RecordLoginAttempt(user, true)
	claims := router.createClaims(user)
	accessToken := router.CreateAccessToken(claims)
	refreshToken := router.createRefreshToken(claims, payload.LongLived)
	res := &JWTResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	SendJSON(w, res)
}

func (router *AuthRouter) verify(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	authState, err := GetAuthStateRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	if authState.AuthStateType == AuthAtlassian {
		router.handleAtlassianVerify(authState, w)
		return
	}
	if authState.AuthStateType != AuthResponseCache {
		SendNotFound(w)
		return
	}
	provider, err := GetAuthProviderRepository().GetOne(authState.AuthProviderID)
	if err != nil {
		SendNotFound(w)
		return
	}
	payload := unmarshalAuthStateLoginPayload(authState.Payload)
	user, err := GetUserRepository().GetByEmail(provider.OrganizationID, payload.UserID)
	// TODO Change email to auth server ID???
	if err != nil {
		org, err := GetOrganizationRepository().GetOne(provider.OrganizationID)
		if err != nil {
			SendInternalServerError(w)
			return
		}
		if !GetUserRepository().CanCreateUser(org) {
			SendPaymentRequired(w)
			return
		}
		user = &User{
			Email:          payload.UserID,
			OrganizationID: org.ID,
			Role:           UserRoleUser,
		}
		GetUserRepository().Create(user)
	}
	if user.OrganizationID != provider.OrganizationID {
		SendBadRequest(w)
		return
	}
	if user.Disabled {
		SendNotFound(w)
		return
	}
	GetAuthStateRepository().Delete(authState)
	GetAuthAttemptRepository().RecordLoginAttempt(user, true)
	claims := router.createClaims(user)
	accessToken := router.CreateAccessToken(claims)
	refreshToken := router.createRefreshToken(claims, payload.LongLived)
	res := &JWTResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		LongLived:    payload.LongLived,
		LogoutURL:    router.getLogoutUrl(provider),
	}
	SendJSON(w, res)
}

func (router *AuthRouter) getLogoutUrl(provider *AuthProvider) string {
	if provider.LogoutURL == "" {
		return ""
	}
	org, _ := GetOrganizationRepository().GetOne(provider.OrganizationID)
	primaryDomain, _ := GetOrganizationRepository().GetPrimaryDomain(org)
	redirectUrl := "https://" + primaryDomain.DomainName + "/ui/login"
	logoutUrl := strings.ReplaceAll(provider.LogoutURL, "{logoutRedirectUri}", redirectUrl)
	return logoutUrl
}

func (router *AuthRouter) login(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	loginType := vars["type"]
	if loginType != "web" && loginType != "app" && loginType != "ui" {
		SendBadRequest(w)
		return
	}
	provider, err := GetAuthProviderRepository().GetOne(vars["id"])
	if err != nil {
		SendTemporaryRedirect(w, router.getRedirectFailedUrl(loginType, provider))
		return
	}
	longLived := false
	if vars["longLived"] == "1" {
		longLived = true
	}
	redir := r.URL.Query().Get("redir")
	config := router.getConfig(provider)
	payload := &AuthStateLoginPayload{
		LoginType: loginType,
		UserID:    "",
		LongLived: longLived, // TODO
		Redirect:  redir,
	}
	authState := &AuthState{
		AuthProviderID: provider.ID,
		Expiry:         time.Now().Add(time.Minute * 5),
		AuthStateType:  AuthRequestState,
		Payload:        marshalAuthStateLoginPayload(payload),
	}
	if err := GetAuthStateRepository().Create(authState); err != nil {
		SendTemporaryRedirect(w, router.getRedirectFailedUrl(loginType, provider))
		return
	}
	url := config.AuthCodeURL(authState.ID)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (router *AuthRouter) callback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider, err := GetAuthProviderRepository().GetOne(vars["id"])
	if err != nil {
		SendTemporaryRedirect(w, router.getRedirectFailedUrl("ui", provider))
		return
	}
	claims, payload, err := router.getUserInfo(provider, r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		log.Println(err)
		SendTemporaryRedirect(w, router.getRedirectFailedUrl("ui", provider))
		return
	}
	if !router.isValidEmailForOrg(provider, claims.Email) {
		SendTemporaryRedirect(w, router.getRedirectFailedUrl(payload.LoginType, provider))
		return
	}
	allowAnyUser, _ := GetSettingsRepository().GetBool(provider.OrganizationID, SettingAllowAnyUser.Name)
	if !allowAnyUser {
		_, err := GetUserRepository().GetByEmail(provider.OrganizationID, claims.Email)
		if err != nil {
			SendTemporaryRedirect(w, router.getRedirectFailedUrl(payload.LoginType, provider))
			return
		}
	}
	payloadNew := &AuthStateLoginPayload{
		UserID:    claims.Email,
		LoginType: payload.LoginType,
		LongLived: payload.LongLived,
	}
	authState := &AuthState{
		AuthProviderID: provider.ID,
		Expiry:         time.Now().Add(time.Minute * 5),
		AuthStateType:  AuthResponseCache,
		Payload:        marshalAuthStateLoginPayload(payloadNew),
	}
	if err := GetAuthStateRepository().Create(authState); err != nil {
		log.Println(err)
		SendTemporaryRedirect(w, router.getRedirectFailedUrl(payload.LoginType, provider))
		return
	}
	redirectUrl := router.getRedirectSuccessUrl(payload.LoginType, authState, provider)
	if payload.Redirect != "" {
		redirectUrl = redirectUrl + "?redir=" + url.QueryEscape(payload.Redirect)
	}
	SendTemporaryRedirect(w, redirectUrl)
}

func (router *AuthRouter) getRedirectSuccessUrl(loginType string, authState *AuthState, provider *AuthProvider) string {
	org, _ := GetOrganizationRepository().GetOne(provider.OrganizationID)
	primaryDomain, _ := GetOrganizationRepository().GetPrimaryDomain(org)
	if loginType == "ui" {
		return "https://" + primaryDomain.DomainName + "/ui/login/success/" + authState.ID
	} else {
		return "https://" + primaryDomain.DomainName + "/admin/login/success/" + authState.ID
	}
}

func (router *AuthRouter) getRedirectFailedUrl(loginType string, provider *AuthProvider) string {
	org, _ := GetOrganizationRepository().GetOne(provider.OrganizationID)
	primaryDomain, _ := GetOrganizationRepository().GetPrimaryDomain(org)
	if loginType == "ui" {
		return "https://" + primaryDomain.DomainName + "/ui/login/failed"
	} else {
		return "https://" + primaryDomain.DomainName + "/admin/login/failed"
	}
}

func (router *AuthRouter) isValidEmailForOrg(provider *AuthProvider, email string) bool {
	org, err := GetOrganizationRepository().GetOne(provider.OrganizationID)
	if err != nil {
		return false
	}
	return GetOrganizationRepository().IsValidCustomDomainForOrg(email, org)
}

func (router *AuthRouter) getUserInfo(provider *AuthProvider, state string, code string) (*Claims, *AuthStateLoginPayload, error) {
	// Verify state string
	authState, err := GetAuthStateRepository().GetOne(state)
	if err != nil {
		return nil, nil, fmt.Errorf("state not found for id %s", strings.Replace(strings.Replace(state, "\r", "", -1), "\n", "", -1))
	}
	if authState.AuthProviderID != provider.ID {
		return nil, nil, fmt.Errorf("auth providers don't match")
	}
	defer GetAuthStateRepository().Delete(authState)
	// Exchange authorization code for an access token
	config := router.getConfig(provider)
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}
	// Get user info from resource server
	client := &http.Client{}
	req, err := http.NewRequest("GET", provider.UserInfoURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed creating http request: %s", err.Error())
	}
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)
	response, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed reading response body: %s", err.Error())
	}
	// Extract email address from JSON response
	var result map[string]interface{}
	json.Unmarshal([]byte(contents), &result)
	if (result[provider.UserInfoEmailField] == nil) || (strings.TrimSpace(result[provider.UserInfoEmailField].(string)) == "") {
		return nil, nil, fmt.Errorf("could not read email address from field: %s", provider.UserInfoEmailField)
	}
	claims := &Claims{
		Email: result[provider.UserInfoEmailField].(string),
	}
	payload := unmarshalAuthStateLoginPayload(authState.Payload)
	return claims, payload, nil
}

func (router *AuthRouter) SendPasswordResetEmail(user *User, ID string, org *Organization) error {
	domain, err := GetOrganizationRepository().GetPrimaryDomain(org)
	if err != nil {
		return err
	}
	vars := map[string]string{
		"recipientName":  user.Email,
		"recipientEmail": user.Email,
		"confirmID":      ID,
		"orgDomain":      "https://" + domain.DomainName + "/",
	}
	return SendEmail(&MailAddress{Address: user.Email}, GetEmailSubjectResetPassword(org.Language), GetEmailTemplatePathResetpassword(), org.Language, vars)
}

func (router *AuthRouter) getConfig(provider *AuthProvider) *oauth2.Config {
	org, _ := GetOrganizationRepository().GetOne(provider.OrganizationID)
	primaryDomain, _ := GetOrganizationRepository().GetPrimaryDomain(org)
	config := &oauth2.Config{
		RedirectURL:  "https://" + primaryDomain.DomainName + "/auth/" + provider.ID + "/callback",
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		Scopes:       strings.Split(provider.Scopes, ","),
		Endpoint: oauth2.Endpoint{
			AuthURL:   provider.AuthURL,
			TokenURL:  provider.TokenURL,
			AuthStyle: oauth2.AuthStyle(provider.AuthStyle),
		},
	}
	return config
}

func (router *AuthRouter) createClaims(user *User) *Claims {
	claims := &Claims{
		UserID:     user.ID,
		Email:      user.Email,
		SpaceAdmin: GetUserRepository().IsSpaceAdmin(user),
		OrgAdmin:   GetUserRepository().IsOrgAdmin(user),
		Role:       int(user.Role),
	}
	return claims
}

func (router *AuthRouter) CreateAccessToken(claims *Claims) string {
	claims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	jwtString, err := accessToken.SignedString(GetConfig().JwtPrivateKey)
	if err != nil {
		return ""
	}
	return jwtString
}

func (router *AuthRouter) createRefreshToken(claims *Claims, longLived bool) string {
	var expiry time.Time
	if longLived {
		expiry = time.Now().Add(60 * 24 * 28 * time.Minute)
	} else {
		expiry = time.Now().Add(60 * 24 * time.Minute)
	}
	refreshToken := &RefreshToken{
		UserID:  claims.UserID,
		Expiry:  expiry,
		Created: time.Now(),
	}
	GetRefreshTokenRepository().Create(refreshToken)
	return refreshToken.ID
}

func (router *AuthRouter) getOrgForEmail(email string) *Organization {
	mailParts := strings.Split(email, "@")
	if len(mailParts) != 2 {
		return nil
	}
	domain := strings.ToLower(mailParts[1])
	org, err := GetOrganizationRepository().GetOneByDomain(domain)
	if err != nil {
		log.Println(err)
		return nil
	}
	return org
}

func (router *AuthRouter) getPreflightResponseForOrg(org *Organization) *AuthPreflightResponse {
	list, err := GetAuthProviderRepository().GetAll(org.ID)
	if err != nil {
		return nil
	}
	res := &AuthPreflightResponse{
		Organization: &GetOrganizationResponse{
			ID: org.ID,
			CreateOrganizationRequest: CreateOrganizationRequest{
				Name: org.Name,
			},
		},
		RequirePassword: false,
		AuthProviders:   []*GetAuthProviderPublicResponse{},
		BackendVersion:  GetProductVersion(),
	}
	domain, err := GetOrganizationRepository().GetPrimaryDomain(org)
	if domain != nil && err == nil {
		res.Domain = domain.DomainName
	}
	for _, e := range list {
		m := &GetAuthProviderPublicResponse{}
		m.ID = e.ID
		m.Name = e.Name
		res.AuthProviders = append(res.AuthProviders, m)
	}
	return res
}

func marshalAuthStateLoginPayload(payload *AuthStateLoginPayload) string {
	json, _ := json.Marshal(payload)
	return string(json)
}

func unmarshalAuthStateLoginPayload(payload string) *AuthStateLoginPayload {
	var o *AuthStateLoginPayload
	json.Unmarshal([]byte(payload), &o)
	return o
}
