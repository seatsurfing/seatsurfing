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

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/config"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

type JWTResponse struct {
	AccessToken    string `json:"accessToken"`
	RefreshToken   string `json:"refreshToken"`
	LogoutURL      string `json:"logoutUrl"`
	ProfilePageURL string `json:"profilePageUrl"`
}

type Claims struct {
	Email      string `json:"email"`
	UserID     string `json:"userID"`
	SpaceAdmin bool   `json:"spaceAdmin"`
	OrgAdmin   bool   `json:"admin"`
	Role       int    `json:"role"`
	jwt.RegisteredClaims
}

type IdPUserInfo struct {
	Email     string
	Firstname string
	Lastname  string
}

type InitPasswordResetRequest struct {
	OrganizationID string `json:"organizationId" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
}

type CompletePasswordResetRequest struct {
	Password string `json:"password" validate:"required,min=8"`
}

type AuthPreflightResponse struct {
	Organization         *GetOrganizationResponse         `json:"organization"`
	AuthProviders        []*GetAuthProviderPublicResponse `json:"authProviders"`
	RequirePassword      bool                             `json:"requirePassword"`
	DisablePasswordLogin bool                             `json:"disablePasswordLogin"`
	Domain               string                           `json:"domain"`
}

type AuthPasswordRequest struct {
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=8"`
	OrganizationID string `json:"organizationId" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type AuthStateLoginPayload struct {
	UserID    string `json:"userId"`
	LoginType string `json:"type"`
	Redirect  string `json:"redirect,omitempty"`
}

type AuthRouter struct {
}

func (router *AuthRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/verify/{id}", router.verify).Methods("GET")
	s.HandleFunc("/{id}/login/{type}/", router.login).Methods("GET")
	s.HandleFunc("/{id}/callback", router.callback).Methods("GET")
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
	if requirePassword && GetConfig().DisablePasswordLogin {
		requirePassword = false
	}
	res.RequirePassword = requirePassword
	res.DisablePasswordLogin = GetConfig().DisablePasswordLogin
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
	if requirePassword && GetConfig().DisablePasswordLogin {
		requirePassword = false
	}
	res.RequirePassword = requirePassword
	res.DisablePasswordLogin = GetConfig().DisablePasswordLogin
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
	now := time.Now().UTC()
	user.LastActivityAtUTC = &now
	GetUserRepository().Update(user)
	claims := router.createClaims(user)
	accessToken := router.CreateAccessToken(claims)
	newRefreshToken := router.createRefreshToken(claims)
	res := &JWTResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}
	GetRefreshTokenRepository().Delete(refreshToken)
	SendJSON(w, res)
}

func (router *AuthRouter) initPasswordReset(w http.ResponseWriter, r *http.Request) {
	if GetConfig().DisablePasswordLogin {
		SendNotFound(w)
		return
	}
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
	if user.Role == UserRoleServiceAccountRO || user.Role == UserRoleServiceAccountRW {
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
	if err := router.SendPasswordResetEmail(user, authState.ID, org); err != nil {
		log.Printf("Password reset email failed: %s\n", err)
	}
	SendUpdated(w)
}

func (router *AuthRouter) completePasswordReset(w http.ResponseWriter, r *http.Request) {
	if GetConfig().DisablePasswordLogin {
		SendNotFound(w)
		return
	}
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
	if user.Role == UserRoleServiceAccountRO || user.Role == UserRoleServiceAccountRW {
		SendNotFound(w)
		return
	}
	user.HashedPassword = NullString(GetUserRepository().GetHashedPassword(m.Password))
	GetUserRepository().Update(user)
	GetAuthStateRepository().Delete(authState)
	SendUpdated(w)
}

func (router *AuthRouter) loginPassword(w http.ResponseWriter, r *http.Request) {
	if GetConfig().DisablePasswordLogin {
		SendNotFound(w)
		return
	}
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
	if user.Role == UserRoleServiceAccountRO || user.Role == UserRoleServiceAccountRW {
		SendNotFound(w)
		return
	}
	if !GetUserRepository().CheckPassword(string(user.HashedPassword), m.Password) {
		GetAuthAttemptRepository().RecordLoginAttempt(user, false)
		SendNotFound(w)
		return
	}
	GetAuthAttemptRepository().RecordLoginAttempt(user, true)
	now := time.Now().UTC()
	user.LastActivityAtUTC = &now
	GetUserRepository().Update(user)
	claims := router.createClaims(user)
	accessToken := router.CreateAccessToken(claims)
	refreshToken := router.createRefreshToken(claims)
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
	if user.Role == UserRoleServiceAccountRO || user.Role == UserRoleServiceAccountRW {
		SendNotFound(w)
		return
	}
	GetAuthStateRepository().Delete(authState)
	GetAuthAttemptRepository().RecordLoginAttempt(user, true)
	claims := router.createClaims(user)
	accessToken := router.CreateAccessToken(claims)
	refreshToken := router.createRefreshToken(claims)
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
	payload := unmarshalAuthStateLoginPayload(authState.Payload)
	var user *User
	var provider *AuthProvider
	if authState.AuthProviderID != "" && authState.AuthProviderID != GetSettingsRepository().GetNullUUID() {
		provider, _ = GetAuthProviderRepository().GetOne(authState.AuthProviderID)
		if provider == nil {
			SendNotFound(w)
			return
		}
		user, _ = GetUserRepository().GetByEmail(provider.OrganizationID, payload.UserID)
		if user == nil {
			org, err := GetOrganizationRepository().GetOne(provider.OrganizationID)
			if err != nil {
				SendInternalServerError(w)
				return
			}
			allowAnyUser, _ := GetSettingsRepository().GetBool(provider.OrganizationID, SettingAllowAnyUser.Name)
			if !allowAnyUser {
				SendNotFound(w)
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
		} else {
			if user.OrganizationID != provider.OrganizationID {
				SendBadRequest(w)
				return
			}
		}
	} else if err := uuid.Validate(payload.UserID); err == nil {
		user, _ = GetUserRepository().GetOne(payload.UserID)
	}
	if user == nil {
		SendNotFound(w)
		return
	}
	if user.Disabled {
		SendNotFound(w)
		return
	}
	if user.Role == UserRoleServiceAccountRO || user.Role == UserRoleServiceAccountRW {
		SendNotFound(w)
		return
	}
	GetAuthStateRepository().Delete(authState)
	GetAuthAttemptRepository().RecordLoginAttempt(user, true)
	now := time.Now().UTC()
	user.LastActivityAtUTC = &now
	GetUserRepository().Update(user)
	claims := router.createClaims(user)
	accessToken := router.CreateAccessToken(claims)
	refreshToken := router.createRefreshToken(claims)
	res := &JWTResponse{
		AccessToken:    accessToken,
		RefreshToken:   refreshToken,
		LogoutURL:      router.getLogoutUrl(provider),
		ProfilePageURL: router.getProfilePageURL(provider),
	}
	SendJSON(w, res)
}

func (router *AuthRouter) getLogoutUrl(provider *AuthProvider) string {
	if provider == nil || provider.LogoutURL == "" {
		return ""
	}
	org, _ := GetOrganizationRepository().GetOne(provider.OrganizationID)
	primaryDomain, _ := GetOrganizationRepository().GetPrimaryDomain(org)
	redirectUrl := FormatURL(primaryDomain.DomainName) + "/ui/login"
	logoutUrl := strings.ReplaceAll(provider.LogoutURL, "{logoutRedirectUri}", redirectUrl)
	return logoutUrl
}

func (router *AuthRouter) getProfilePageURL(provider *AuthProvider) string {
	if provider == nil || provider.ProfilePageURL == "" {
		return ""
	}
	return provider.ProfilePageURL
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
		SendTemporaryRedirect(w, router.getRedirectFailedUrl(loginType, provider, "provider"))
		return
	}
	redir := r.URL.Query().Get("redir")
	config := router.getConfig(provider)
	if config == nil {
		SendTemporaryRedirect(w, router.getRedirectFailedUrl(loginType, provider, "config"))
		return
	}
	payload := &AuthStateLoginPayload{
		LoginType: loginType,
		UserID:    "",
		Redirect:  redir,
	}
	authState := &AuthState{
		AuthProviderID: provider.ID,
		Expiry:         time.Now().Add(time.Minute * 5),
		AuthStateType:  AuthRequestState,
		Payload:        marshalAuthStateLoginPayload(payload),
	}
	if err := GetAuthStateRepository().Create(authState); err != nil {
		SendTemporaryRedirect(w, router.getRedirectFailedUrl(loginType, provider, "authState"))
		return
	}
	url := config.AuthCodeURL(authState.ID)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (router *AuthRouter) callback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider, err := GetAuthProviderRepository().GetOne(vars["id"])
	if err != nil {
		log.Printf("Error getting auth provider %s: %s\n", vars["id"], err)
		SendTemporaryRedirect(w, router.getRedirectFailedUrl("ui", provider, "provider"))
		return
	}
	userInfo, payload, err := router.getUserInfo(provider, r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		log.Printf("Error getting user info for provider %s: %s\n", vars["id"], err)

		// forward error and error_description from callback URL to our login failed page
		redirectUrl := router.getRedirectFailedUrl("ui", provider, "userinfo")
		if error := r.FormValue("error"); error != "" {
			redirectUrl += "&error=" + url.QueryEscape(error)
		}
		if errorDesc := r.FormValue("error_description"); errorDesc != "" {
			redirectUrl += "&error_description=" + url.QueryEscape(errorDesc)
		}
		SendTemporaryRedirect(w, redirectUrl)

		return
	}
	allowAnyUser, _ := GetSettingsRepository().GetBool(provider.OrganizationID, SettingAllowAnyUser.Name)
	user, err := GetUserRepository().GetByEmail(provider.OrganizationID, userInfo.Email)
	if !allowAnyUser {
		if err != nil || user == nil {
			SendTemporaryRedirect(w, router.getRedirectFailedUrl(payload.LoginType, provider, "login"))
			return
		}
	}
	if user == nil {
		org, err := GetOrganizationRepository().GetOne(provider.OrganizationID)
		if org == nil || err != nil {
			SendNotFound(w)
			return
		}
		if !GetUserRepository().CanCreateUser(org) {
			SendPaymentRequired(w)
			return
		}
		user = &User{
			Email:          userInfo.Email,
			OrganizationID: org.ID,
			Role:           UserRoleUser,
		}
		GetUserRepository().Create(user)
	}
	needUserUpdate := false
	if userInfo.Firstname != "" && user.Firstname != userInfo.Firstname {
		user.Firstname = userInfo.Firstname
		needUserUpdate = true
	}
	if userInfo.Lastname != "" && user.Lastname != userInfo.Lastname {
		user.Lastname = userInfo.Lastname
		needUserUpdate = true
	}
	if needUserUpdate {
		GetUserRepository().Update(user)
	}
	payloadNew := &AuthStateLoginPayload{
		UserID:    userInfo.Email,
		LoginType: payload.LoginType,
	}
	authState := &AuthState{
		AuthProviderID: provider.ID,
		Expiry:         time.Now().Add(time.Minute * 5),
		AuthStateType:  AuthResponseCache,
		Payload:        marshalAuthStateLoginPayload(payloadNew),
	}
	if err := GetAuthStateRepository().Create(authState); err != nil {
		log.Println(err)
		SendTemporaryRedirect(w, router.getRedirectFailedUrl(payload.LoginType, provider, "authState"))
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
		return FormatURL(primaryDomain.DomainName) + "/ui/login/success/" + authState.ID + "/"
	} else {
		return FormatURL(primaryDomain.DomainName) + "/ui/login/success/" + authState.ID + "/"
	}
}

func (router *AuthRouter) getRedirectFailedUrl(loginType string, provider *AuthProvider, reason string) string {
	org, _ := GetOrganizationRepository().GetOne(provider.OrganizationID)
	primaryDomain, _ := GetOrganizationRepository().GetPrimaryDomain(org)

	queryString := "?reason=" + url.QueryEscape(reason)

	if primaryDomain == nil {
		log.Println("Error compiling redirect failed URL for auth provider " + provider.Name + ": No primary domain found for organization")
		if loginType == "ui" {
			return "/ui/login/failed/" + queryString
		} else {
			return "/ui/login/failed/" + queryString
		}
	}
	if loginType == "ui" {
		return FormatURL(primaryDomain.DomainName) + "/ui/login/failed/" + queryString
	} else {
		return FormatURL(primaryDomain.DomainName) + "/ui/login/failed/" + queryString
	}
}

func (router *AuthRouter) getUserInfo(provider *AuthProvider, state string, code string) (*IdPUserInfo, *AuthStateLoginPayload, error) {
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
	email := strings.TrimSpace(result[provider.UserInfoEmailField].(string))
	firstname := ""
	lastname := ""
	if provider.UserInfoFirstnameField != "" {
		if result[provider.UserInfoFirstnameField] != nil {
			firstname = strings.TrimSpace(result[provider.UserInfoFirstnameField].(string))
		}
	}
	if provider.UserInfoLastnameField != "" {
		if result[provider.UserInfoLastnameField] != nil {
			lastname = strings.TrimSpace(result[provider.UserInfoLastnameField].(string))
		}
	}
	info := &IdPUserInfo{
		Email:     email,
		Firstname: firstname,
		Lastname:  lastname,
	}
	payload := unmarshalAuthStateLoginPayload(authState.Payload)
	return info, payload, nil
}

func (router *AuthRouter) SendPasswordResetEmail(user *User, ID string, org *Organization) error {
	domain, err := GetOrganizationRepository().GetPrimaryDomain(org)
	if err != nil {
		return err
	}
	vars := map[string]string{
		"recipientName":  user.GetSafeRecipientName(),
		"recipientEmail": user.Email,
		"confirmID":      ID,
		"orgDomain":      FormatURL(domain.DomainName) + "/",
	}
	return SendEmailWithOrg(&MailAddress{Address: user.Email}, GetEmailTemplatePathResetpassword(), org.Language, vars, org.ID)
}

func (router *AuthRouter) getConfig(provider *AuthProvider) *oauth2.Config {
	org, _ := GetOrganizationRepository().GetOne(provider.OrganizationID)
	primaryDomain, _ := GetOrganizationRepository().GetPrimaryDomain(org)
	if primaryDomain == nil {
		log.Println("Error compiling config for auth provider " + provider.Name + ": No primary domain found for organization")
		return nil
	}
	config := &oauth2.Config{
		RedirectURL:  FormatURL(primaryDomain.DomainName) + "/auth/" + provider.ID + "/callback",
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
		log.Println(err)
		return ""
	}
	return jwtString
}

func (router *AuthRouter) createRefreshToken(claims *Claims) string {
	var expiry time.Time
	expiry = time.Now().Add(60 * 24 * 28 * time.Minute)
	refreshToken := &RefreshToken{
		UserID:  claims.UserID,
		Expiry:  expiry,
		Created: time.Now(),
	}
	GetRefreshTokenRepository().Create(refreshToken)
	return refreshToken.ID
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
		RequirePassword:      false,
		DisablePasswordLogin: GetConfig().DisablePasswordLogin,
		AuthProviders:        []*GetAuthProviderPublicResponse{},
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
