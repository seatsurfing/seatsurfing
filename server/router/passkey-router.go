package router

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/config"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

const passkeyAuthStateZeroID = "00000000-0000-0000-0000-000000000000"
const passkeyAuthStateExpiry = 5 * time.Minute

// passkeyMinResponseTime is the minimum duration enforced for finishPasskeyLogin
// responses to mitigate credential-existence timing side-channels (Finding #12).
const passkeyMinResponseTime = 100 * time.Millisecond

// encryptPasskeyPayload encrypts a passkey state payload using AES-256-GCM so
// that challenge material is not stored in plaintext in the database (Finding #1).
func encryptPasskeyPayload(payload string) (string, error) {
	return EncryptString(payload)
}

// decryptPasskeyPayload decrypts a passkey state payload. If decryption fails
// it falls back to treating the value as plaintext, providing backward
// compatibility with states created before encryption was added (Finding #1).
func decryptPasskeyPayload(payload string) string {
	if decrypted, err := DecryptString(payload); err == nil {
		return decrypted
	}
	return payload
}

// isValidPasskeyName returns false if the name contains characters that could
// be used for HTML injection or log injection (Finding #6).
func isValidPasskeyName(name string) bool {
	return !strings.ContainsAny(name, "<>&\"'\x00")
}

// BeginPasskeyLoginRequest is the body accepted by beginPasskeyLogin.
type BeginPasskeyLoginRequest struct {
	OrganizationID string `json:"organizationId" validate:"required,uuid"`
}

// passkeyLoginStatePayload is stored in the AuthState for login flows.
type passkeyLoginStatePayload struct {
	OrgID       string               `json:"orgId"`
	SessionData webauthn.SessionData `json:"sessionData"`
}

// ---------------------------------------------------------------------------
// WebAuthnUser – implements the webauthn.User interface
// ---------------------------------------------------------------------------

type WebAuthnUser struct {
	user        *User
	credentials []webauthn.Credential
}

func (u *WebAuthnUser) WebAuthnID() []byte                         { return []byte(u.user.ID) }
func (u *WebAuthnUser) WebAuthnName() string                       { return u.user.Email }
func (u *WebAuthnUser) WebAuthnDisplayName() string                { return u.user.Firstname + " " + u.user.Lastname }
func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential { return u.credentials }

// loadWebAuthnUser loads all passkeys for the given user and converts them into
// webauthn.Credential values. Any passkey that cannot be decrypted is skipped
// with a warning (so that a corrupt credential doesn't lock the user out).
func loadWebAuthnUser(user *User) (*WebAuthnUser, error) {
	passkeys, err := GetPasskeyRepository().GetAllByUserID(user.ID)
	if err != nil {
		return nil, err
	}
	creds := make([]webauthn.Credential, 0, len(passkeys))
	for _, pk := range passkeys {
		cred, err := pk.ToWebAuthnCredential()
		if err != nil {
			log.Printf("Warning: could not decrypt passkey %s for user %s: %v\n", pk.ID, user.ID, err)
			continue
		}
		creds = append(creds, *cred)
	}
	return &WebAuthnUser{user: user, credentials: creds}, nil
}

// ---------------------------------------------------------------------------
// WebAuthn instance factory
// ---------------------------------------------------------------------------

// getWebAuthnInstance creates a configured webauthn.WebAuthn instance using
// the organisation's primary domain as the Relying Party ID and origin.
func getWebAuthnInstance(org *Organization) (*webauthn.WebAuthn, error) {
	domain, err := GetOrganizationRepository().GetPrimaryDomain(org)
	if err != nil {
		return nil, err
	}
	if domain == nil {
		return nil, fmt.Errorf("no primary domain configured for organisation %s", org.ID)
	}
	// Strip port for a bare RPID (WebAuthn spec §7.1)
	rpID := domain.DomainName
	if colonIdx := strings.LastIndex(rpID, ":"); colonIdx != -1 {
		rpID = rpID[:colonIdx]
	}
	rpDisplayName := GetConfig().WebAuthnRPDisplayName
	if rpDisplayName == "" {
		rpDisplayName = "Seatsurfing"
	}
	return webauthn.New(&webauthn.Config{
		RPID:          rpID,
		RPDisplayName: rpDisplayName,
		RPOrigins:     []string{FormatURL(domain.DomainName)},
	})
}

// isRequestFromPrimaryDomain reports whether the Host of r matches the
// organisation's primary domain (case-insensitive). Non-primary-domain
// requests must not be permitted to register new passkeys (spec §8.3).
func isRequestFromPrimaryDomain(r *http.Request, org *Organization) bool {
	primaryDomain, err := GetOrganizationRepository().GetPrimaryDomain(org)
	if err != nil || primaryDomain == nil {
		return false
	}
	return strings.EqualFold(r.Host, primaryDomain.DomainName)
}

// ---------------------------------------------------------------------------
// Request / response structs
// ---------------------------------------------------------------------------

type PasskeyListItemResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"createdAt"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
}

type BeginPasskeyRegistrationResponse struct {
	StateID   string      `json:"stateId"`
	Challenge interface{} `json:"challenge"`
}

type FinishPasskeyRegistrationRequest struct {
	StateID    string          `json:"stateId" validate:"required,uuid4"`
	Name       string          `json:"name" validate:"required,max=255"`
	Credential json.RawMessage `json:"credential" validate:"required"`
}

type RenamePasskeyRequest struct {
	Name string `json:"name" validate:"required,max=255"`
}

type BeginPasskeyLoginResponse struct {
	StateID   string      `json:"stateId"`
	Challenge interface{} `json:"challenge"`
}

type FinishPasskeyLoginRequest struct {
	StateID    string          `json:"stateId" validate:"required,uuid4"`
	Credential json.RawMessage `json:"credential" validate:"required"`
}

// PasskeyChallengeResponse is embedded in the 401 response from loginPassword
// when the user has passkeys and must prove ownership.
type PasskeyChallengeResponse struct {
	RequirePasskey    bool        `json:"requirePasskey"`
	StateID           string      `json:"stateId"`
	PasskeyChallenge  interface{} `json:"passkeyChallenge"`
	AllowTotpFallback bool        `json:"allowTotpFallback"`
}

// ---------------------------------------------------------------------------
// UserRouter – passkey management routes
// ---------------------------------------------------------------------------

func (router *UserRouter) listPasskeys(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if user == nil {
		SendUnauthorized(w)
		return
	}
	passkeys, err := GetPasskeyRepository().GetAllByUserID(user.ID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	result := make([]*PasskeyListItemResponse, 0, len(passkeys))
	for _, pk := range passkeys {
		result = append(result, &PasskeyListItemResponse{
			ID:         pk.ID,
			Name:       pk.Name,
			CreatedAt:  pk.CreatedAt,
			LastUsedAt: pk.LastUsedAt,
		})
	}
	SendJSON(w, result)
}

func (router *UserRouter) beginPasskeyRegistration(w http.ResponseWriter, r *http.Request) {
	if !CanCrypt() {
		SendInternalServerError(w)
		return
	}
	user := GetRequestUser(r)
	if user == nil {
		SendUnauthorized(w)
		return
	}
	// Preconditions per spec §5.4: user must have a password set and not be an IdP user
	if user.HashedPassword == "" {
		SendForbidden(w)
		return
	}
	if string(user.AuthProviderID) != "" {
		SendForbidden(w)
		return
	}
	// Enforce per-user passkey limit (Finding #4)
	if GetPasskeyRepository().GetCountByUserID(user.ID) >= GetConfig().MaxPasskeysPerUser {
		SendForbidden(w)
		return
	}
	org, err := GetOrganizationRepository().GetOne(user.OrganizationID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	// Spec §8.3: passkey registration is only allowed from the primary domain.
	if !isRequestFromPrimaryDomain(r, org) {
		SendForbidden(w)
		return
	}
	wa, err := getWebAuthnInstance(org)
	if err != nil {
		log.Println("WebAuthn config error:", err)
		SendInternalServerError(w)
		return
	}
	webAuthnUser, err := loadWebAuthnUser(user)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	// Build exclusion list so the same authenticator can't be registered twice
	excludeList := make([]protocol.CredentialDescriptor, 0, len(webAuthnUser.credentials))
	for _, c := range webAuthnUser.credentials {
		excludeList = append(excludeList, c.Descriptor())
	}
	creation, sessionData, err := wa.BeginRegistration(webAuthnUser,
		webauthn.WithExclusions(excludeList),
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			ResidentKey:      protocol.ResidentKeyRequirementRequired,
			UserVerification: protocol.VerificationRequired,
		}),
	)
	if err != nil {
		log.Println("BeginRegistration error:", err)
		SendInternalServerError(w)
		return
	}
	sdJSON, err := json.Marshal(sessionData)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	encryptedSD, err := encryptPasskeyPayload(string(sdJSON))
	if err != nil {
		log.Println("encryptPasskeyPayload error:", err)
		SendInternalServerError(w)
		return
	}
	state := &AuthState{
		AuthProviderID: user.ID,
		Expiry:         time.Now().Add(passkeyAuthStateExpiry),
		AuthStateType:  AuthPasskeyRegistration,
		Payload:        encryptedSD,
	}
	if err := GetAuthStateRepository().Create(state); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendJSON(w, &BeginPasskeyRegistrationResponse{
		StateID:   state.ID,
		Challenge: creation,
	})
}

func (router *UserRouter) finishPasskeyRegistration(w http.ResponseWriter, r *http.Request) {
	if !CanCrypt() {
		SendInternalServerError(w)
		return
	}
	user := GetRequestUser(r)
	if user == nil {
		SendUnauthorized(w)
		return
	}
	var m FinishPasskeyRegistrationRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	// Reject names containing characters that could be used for injection (Finding #6)
	if !isValidPasskeyName(m.Name) {
		SendBadRequest(w)
		return
	}
	state, err := GetAuthStateRepository().GetOne(m.StateID)
	if err != nil || state == nil {
		SendNotFound(w)
		return
	}
	if state.AuthStateType != AuthPasskeyRegistration {
		SendNotFound(w)
		return
	}
	if state.Expiry.Before(time.Now()) {
		GetAuthStateRepository().Delete(state)
		SendNotFound(w)
		return
	}
	if state.AuthProviderID != user.ID {
		SendForbidden(w)
		return
	}
	GetAuthStateRepository().Delete(state)

	org2, err := GetOrganizationRepository().GetOne(user.OrganizationID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	var sessionData webauthn.SessionData
	if err := json.Unmarshal([]byte(decryptPasskeyPayload(state.Payload)), &sessionData); err != nil {
		log.Println("unmarshal session data error:", err)
		SendInternalServerError(w)
		return
	}
	wa, err := getWebAuthnInstance(org2)
	if err != nil {
		log.Println("WebAuthn config error:", err)
		SendInternalServerError(w)
		return
	}
	webAuthnUser, err := loadWebAuthnUser(user)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	parsedCred, err := protocol.ParseCredentialCreationResponseBytes(m.Credential)
	if err != nil {
		log.Println("parse credential creation error:", err)
		SendBadRequest(w)
		return
	}
	credential, err := wa.CreateCredential(webAuthnUser, sessionData, parsedCred)
	if err != nil {
		log.Println("CreateCredential error:", err)
		SendBadRequest(w)
		return
	}
	passkey, err := NewPasskeyFromCredential(user.ID, credential, m.Name)
	if err != nil {
		log.Println("NewPasskeyFromCredential error:", err)
		SendInternalServerError(w)
		return
	}
	if err := GetPasskeyRepository().Create(passkey); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendJSON(w, &PasskeyListItemResponse{
		ID:        passkey.ID,
		Name:      passkey.Name,
		CreatedAt: passkey.CreatedAt,
	})
}

func (router *UserRouter) renamePasskey(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if user == nil {
		SendUnauthorized(w)
		return
	}
	vars := mux.Vars(r)
	pk, err := GetPasskeyRepository().GetOne(vars["id"])
	if err != nil || pk == nil {
		SendNotFound(w)
		return
	}
	if pk.UserID != user.ID {
		SendForbidden(w)
		return
	}
	var m RenamePasskeyRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	// Reject names containing characters that could be used for injection (Finding #6)
	if !isValidPasskeyName(m.Name) {
		SendBadRequest(w)
		return
	}
	pk.Name = m.Name
	if err := GetPasskeyRepository().UpdateName(pk); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *UserRouter) deletePasskey(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if user == nil {
		SendUnauthorized(w)
		return
	}
	vars := mux.Vars(r)
	pk, err := GetPasskeyRepository().GetOne(vars["id"])
	if err != nil || pk == nil {
		SendNotFound(w)
		return
	}
	if pk.UserID != user.ID {
		SendForbidden(w)
		return
	}
	// Spec §5.4: if enforce_totp is enabled and this is the user's last passkey
	// and TOTP is not configured, refuse deletion (would violate 2FA enforcement).
	enforceTotp, _ := GetSettingsRepository().GetBool(user.OrganizationID, SettingEnforceTOTP.Name)
	if enforceTotp && user.TotpSecret == "" {
		passkeyCount := GetPasskeyRepository().GetCountByUserID(user.ID)
		if passkeyCount <= 1 {
			SendForbidden(w)
			return
		}
	}
	if err := GetPasskeyRepository().Delete(pk); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

// ---------------------------------------------------------------------------
// AuthRouter – passkey login routes (discoverable / usernameless)
// ---------------------------------------------------------------------------

func (router *AuthRouter) beginPasskeyLogin(w http.ResponseWriter, r *http.Request) {
	// Encryption must be available to securely store the challenge (Finding #1)
	if !CanCrypt() {
		SendInternalServerError(w)
		return
	}
	var m BeginPasskeyLoginRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	org, err := GetOrganizationRepository().GetOne(m.OrganizationID)
	if err != nil {
		SendNotFound(w)
		return
	}
	wa, err := getWebAuthnInstance(org)
	if err != nil {
		log.Println("WebAuthn config error:", err)
		SendInternalServerError(w)
		return
	}
	assertion, sessionData, err := wa.BeginDiscoverableLogin(
		webauthn.WithUserVerification(protocol.VerificationRequired),
	)
	if err != nil {
		log.Println("BeginDiscoverableLogin error:", err)
		SendInternalServerError(w)
		return
	}
	statePayload, err := json.Marshal(&passkeyLoginStatePayload{
		OrgID:       org.ID,
		SessionData: *sessionData,
	})
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	encryptedPayload, err := encryptPasskeyPayload(string(statePayload))
	if err != nil {
		log.Println("encryptPasskeyPayload error:", err)
		SendInternalServerError(w)
		return
	}
	state := &AuthState{
		AuthProviderID: passkeyAuthStateZeroID,
		Expiry:         time.Now().Add(passkeyAuthStateExpiry),
		AuthStateType:  AuthPasskeyLogin,
		Payload:        encryptedPayload,
	}
	if err := GetAuthStateRepository().Create(state); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendJSON(w, &BeginPasskeyLoginResponse{
		StateID:   state.ID,
		Challenge: assertion,
	})
}

func (router *AuthRouter) finishPasskeyLogin(w http.ResponseWriter, r *http.Request) {
	// Enforce a minimum response time to mitigate credential-existence
	// timing side-channels in the DiscoverableUserHandler (Finding #12).
	start := time.Now()
	defer func() {
		if elapsed := time.Since(start); elapsed < passkeyMinResponseTime {
			time.Sleep(passkeyMinResponseTime - elapsed)
		}
	}()
	var m FinishPasskeyLoginRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	state, err := GetAuthStateRepository().GetOne(m.StateID)
	if err != nil || state == nil {
		SendNotFound(w)
		return
	}
	if state.AuthStateType != AuthPasskeyLogin {
		SendNotFound(w)
		return
	}
	if state.Expiry.Before(time.Now()) {
		GetAuthStateRepository().Delete(state)
		SendNotFound(w)
		return
	}
	GetAuthStateRepository().Delete(state)

	var sp passkeyLoginStatePayload
	if err := json.Unmarshal([]byte(decryptPasskeyPayload(state.Payload)), &sp); err != nil {
		log.Println("unmarshal state payload error:", err)
		SendInternalServerError(w)
		return
	}
	sessionData := sp.SessionData
	loginOrg, err := GetOrganizationRepository().GetOne(sp.OrgID)
	if err != nil {
		log.Println("Error loading org:", err)
		SendInternalServerError(w)
		return
	}
	wa, err := getWebAuthnInstance(loginOrg)
	if err != nil {
		log.Println("WebAuthn config error:", err)
		SendInternalServerError(w)
		return
	}
	parsedAssertion, err := protocol.ParseCredentialRequestResponseBytes(m.Credential)
	if err != nil {
		log.Println("parse assertion error:", err)
		SendBadRequest(w)
		return
	}

	var matchedPasskey *Passkey
	var matchedUser *User

	handler := webauthn.DiscoverableUserHandler(func(rawID, userHandle []byte) (webauthn.User, error) {
		pk, err := GetPasskeyRepository().GetByCredentialIDRaw(rawID)
		if err != nil {
			return nil, err
		}
		u, err := GetUserRepository().GetOne(pk.UserID)
		if err != nil {
			return nil, err
		}
		matchedPasskey = pk
		matchedUser = u
		return loadWebAuthnUser(u)
	})

	credential, err := wa.ValidateDiscoverableLogin(handler, sessionData, parsedAssertion)
	if err != nil {
		log.Println("ValidateDiscoverableLogin error:", err)
		// Record failed attempt if we managed to identify the user
		if matchedUser != nil {
			GetAuthAttemptRepository().RecordLoginAttempt(matchedUser, false)
		}
		SendNotFound(w)
		return
	}
	if matchedPasskey == nil || matchedUser == nil {
		SendNotFound(w)
		return
	}
	if matchedUser.Disabled {
		SendNotFound(w)
		return
	}
	// Must not be a service account
	if matchedUser.Role == UserRoleServiceAccountRO || matchedUser.Role == UserRoleServiceAccountRW {
		SendNotFound(w)
		return
	}
	// Must have a password set (not IdP-only)
	if matchedUser.HashedPassword == "" {
		SendNotFound(w)
		return
	}
	// Must not be pending password set
	if matchedUser.PasswordPending {
		SendNotFound(w)
		return
	}
	// Check ban status
	if matchedUser.BanExpiry != nil && matchedUser.BanExpiry.After(time.Now()) {
		SendUnauthorized(w)
		return
	}

	// Clone detection (spec §7.5): if stored signCount > 0 and the returned
	// signCount is not greater, the credential may have been cloned.
	if matchedPasskey.SignCount > 0 && credential.Authenticator.SignCount <= matchedPasskey.SignCount {
		log.Printf("Warning: clone detection triggered for passkey %s (user %s): stored=%d, returned=%d\n",
			matchedPasskey.ID, matchedUser.ID, matchedPasskey.SignCount, credential.Authenticator.SignCount)
		GetAuthAttemptRepository().RecordLoginAttempt(matchedUser, false)
		SendNotFound(w)
		return
	}

	// Update the stored sign count
	matchedPasskey.SignCount = credential.Authenticator.SignCount
	if err := GetPasskeyRepository().UpdateSignCount(matchedPasskey); err != nil {
		log.Println("UpdateSignCount error:", err)
	}
	if err := GetPasskeyRepository().UpdateLastUsedAt(matchedPasskey); err != nil {
		log.Println("UpdateLastUsedAt error:", err)
	}

	GetAuthAttemptRepository().RecordLoginAttempt(matchedUser, true)
	now := time.Now().UTC()
	matchedUser.LastActivityAtUTC = &now
	GetUserRepository().Update(matchedUser)

	session := router.CreateSession(r, matchedUser)
	if session == nil {
		log.Println("Error: Failed to create session during passkey login")
		SendInternalServerError(w)
		return
	}
	claims := router.CreateClaims(matchedUser, session)
	accessToken := router.CreateAccessToken(claims)
	refreshToken := router.createRefreshToken(claims)
	res := &JWTResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	SendJSON(w, res)
}

// ---------------------------------------------------------------------------
// Passkey 2FA helper – called from AuthRouter.loginPassword
// ---------------------------------------------------------------------------

// passkey2FAResult represents the outcome of handlePasskey2FA.
type passkey2FAResult int

const (
	// passkey2FANotApplicable means the user has no passkeys – caller should
	// continue with TOTP or session creation as before.
	passkey2FANotApplicable passkey2FAResult = iota
	// passkey2FAHandled means the response has been written (challenge issued
	// or error sent) – caller must return immediately.
	passkey2FAHandled
	// passkey2FAVerified means a passkey assertion was successfully verified –
	// caller should skip TOTP and proceed to session creation.
	passkey2FAVerified
)

// handlePasskey2FA checks if the user has passkeys and handles the 2FA flow.
//
// Decision logic (from spec §5.6):
//
//	Has passkeys?
//	├── YES → Passkey assertion provided?
//	│   ├── YES → Verify passkey → passkey2FAVerified / passkey2FAHandled (error)
//	│   └── NO → TOTP code provided AND TOTP configured?
//	│       ├── YES → passkey2FANotApplicable (let caller verify TOTP)
//	│       └── NO → Issue passkey challenge (401) → passkey2FAHandled
//	└── NO  → passkey2FANotApplicable
func (router *AuthRouter) handlePasskey2FA(w http.ResponseWriter, r *http.Request, user *User, m *AuthPasswordRequest) passkey2FAResult {
	count := GetPasskeyRepository().GetCountByUserID(user.ID)
	if count == 0 {
		return passkey2FANotApplicable
	}

	// Load organisation for WebAuthn instance (needed in both the validate and challenge paths)
	org, err := GetOrganizationRepository().GetOne(user.OrganizationID)
	if err != nil {
		log.Println("Error loading org for passkey 2FA:", err)
		SendInternalServerError(w)
		return passkey2FAHandled
	}

	// Passkey credential provided → validate it
	if m.PasskeyStateID != "" && len(m.PasskeyCredential) > 0 {
		state, err := GetAuthStateRepository().GetOne(m.PasskeyStateID)
		if err != nil || state == nil {
			SendNotFound(w)
			return passkey2FAHandled
		}
		if state.AuthStateType != AuthPasskey2FA {
			SendNotFound(w)
			return passkey2FAHandled
		}
		if state.Expiry.Before(time.Now()) {
			GetAuthStateRepository().Delete(state)
			SendNotFound(w)
			return passkey2FAHandled
		}
		// Ensure this challenge was issued for this specific user
		if state.AuthProviderID != user.ID {
			SendForbidden(w)
			return passkey2FAHandled
		}
		GetAuthStateRepository().Delete(state)

		var sp2fa passkeyLoginStatePayload
		if err := json.Unmarshal([]byte(decryptPasskeyPayload(state.Payload)), &sp2fa); err != nil {
			log.Println("unmarshal state payload error:", err)
			SendInternalServerError(w)
			return passkey2FAHandled
		}
		sessionData := sp2fa.SessionData
		wa, err := getWebAuthnInstance(org)
		if err != nil {
			log.Println("WebAuthn config error:", err)
			SendInternalServerError(w)
			return passkey2FAHandled
		}
		webAuthnUser, err := loadWebAuthnUser(user)
		if err != nil {
			log.Println(err)
			SendInternalServerError(w)
			return passkey2FAHandled
		}
		parsedAssertion, err := protocol.ParseCredentialRequestResponseBytes(m.PasskeyCredential)
		if err != nil {
			log.Println("parse assertion error:", err)
			GetAuthAttemptRepository().RecordLoginAttempt(user, false)
			SendNotFound(w)
			return passkey2FAHandled
		}
		credential, err := wa.ValidateLogin(webAuthnUser, sessionData, parsedAssertion)
		if err != nil {
			log.Println("ValidateLogin error:", err)
			GetAuthAttemptRepository().RecordLoginAttempt(user, false)
			SendNotFound(w)
			return passkey2FAHandled
		}
		// Clone detection – same check as the discoverable login path (Finding #3)
		pk, err := GetPasskeyRepository().GetByCredentialIDRaw(credential.ID)
		if err == nil && pk != nil {
			if pk.SignCount > 0 && credential.Authenticator.SignCount <= pk.SignCount {
				log.Printf("Warning: clone detection triggered for passkey %s (user %s): stored=%d, returned=%d\n",
					pk.ID, user.ID, pk.SignCount, credential.Authenticator.SignCount)
				GetAuthAttemptRepository().RecordLoginAttempt(user, false)
				SendNotFound(w)
				return passkey2FAHandled
			}
			pk.SignCount = credential.Authenticator.SignCount
			GetPasskeyRepository().UpdateSignCount(pk)
			GetPasskeyRepository().UpdateLastUsedAt(pk)
		}
		// Passkey is valid – skip TOTP and proceed to session creation
		return passkey2FAVerified
	}

	// TOTP fallback: if user has TOTP configured and a TOTP code was provided,
	// let the caller's existing TOTP verification logic handle it (spec §5.6).
	if m.Code != "" && user.TotpSecret != "" {
		return passkey2FANotApplicable
	}

	// No credential and no TOTP fallback → issue passkey challenge
	wa, err := getWebAuthnInstance(org)
	if err != nil {
		log.Println("WebAuthn config error:", err)
		SendInternalServerError(w)
		return passkey2FAHandled
	}
	webAuthnUser, err := loadWebAuthnUser(user)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return passkey2FAHandled
	}
	allowedCreds := make([]protocol.CredentialDescriptor, 0, len(webAuthnUser.credentials))
	for _, c := range webAuthnUser.credentials {
		allowedCreds = append(allowedCreds, c.Descriptor())
	}
	assertion, sessionData, err := wa.BeginLogin(webAuthnUser,
		webauthn.WithAllowedCredentials(allowedCreds),
		webauthn.WithUserVerification(protocol.VerificationRequired),
	)
	if err != nil {
		log.Println("BeginLogin error:", err)
		SendInternalServerError(w)
		return passkey2FAHandled
	}
	challengePayload, err := json.Marshal(&passkeyLoginStatePayload{
		OrgID:       org.ID,
		SessionData: *sessionData,
	})
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return passkey2FAHandled
	}
	encryptedChallenge, err := encryptPasskeyPayload(string(challengePayload))
	if err != nil {
		log.Println("encryptPasskeyPayload error:", err)
		SendInternalServerError(w)
		return passkey2FAHandled
	}
	state := &AuthState{
		AuthProviderID: user.ID,
		Expiry:         time.Now().Add(passkeyAuthStateExpiry),
		AuthStateType:  AuthPasskey2FA,
		Payload:        encryptedChallenge,
	}
	if err := GetAuthStateRepository().Create(state); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return passkey2FAHandled
	}
	// Return 401 with challenge payload
	resp := &PasskeyChallengeResponse{
		RequirePasskey:    true,
		StateID:           state.ID,
		PasskeyChallenge:  assertion,
		AllowTotpFallback: user.TotpSecret != "",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(resp)
	return passkey2FAHandled
}
