package router

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"image/png"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/pquerna/otp/totp"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

// TOTP validation rate limiter
type totpAttemptTracker struct {
	mu       sync.Mutex
	attempts map[string]int
}

var totpAttemptsTracker = &totpAttemptTracker{
	attempts: make(map[string]int),
}

const maxTotpAttempts = 5

func (t *totpAttemptTracker) recordAttempt(stateID string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	count := t.attempts[stateID]
	if count >= maxTotpAttempts {
		return false
	}
	t.attempts[stateID] = count + 1
	return true
}

func (t *totpAttemptTracker) clearAttempts(stateID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.attempts, stateID)
}

type UserRouter struct {
}

type GetSessionResponse struct {
	ID      string    `json:"id"`
	UserID  string    `json:"userId"`
	Device  string    `json:"device"`
	Created time.Time `json:"created"`
}

type CreateUserRequest struct {
	Email          string `json:"email" validate:"required,max=256"`
	Firstname      string `json:"firstname" validate:"required,max=128"`
	Lastname       string `json:"lastname" validate:"required,max=128"`
	AtlassianID    string `json:"atlassianId"`
	AccountType    int    `json:"accountType"`
	AuthProviderID string `json:"authProviderId"`
	Password       string `json:"password"`
	SendInvitation bool   `json:"sendInvitation"`
	OrganizationID string `json:"organizationId"`
}

type GetUserResponse struct {
	ID              string                  `json:"id"`
	Organization    GetOrganizationResponse `json:"organization"`
	RequirePassword bool                    `json:"requirePassword"`
	PasswordPending bool                    `json:"passwordPending"`
	AuthProviderID  string                  `json:"authProviderId"`
	RoleIDs         []string                `json:"roleIds"`
	TotpEnabled     bool                    `json:"totpEnabled"`
	HasPasskeys     bool                    `json:"hasPasskeys"`
	LastActivity    *time.Time              `json:"lastActivity"`
	CreateUserRequest
}

type GetUserSelfResponse struct {
	IsPrimaryDomain bool `json:"isPrimaryDomain"`
	// Permissions is the caller's own resolved access, keyed by permission
	// name. It is returned only for the authenticated user themselves, since
	// resolving it for every entry of a user list would cost a query each.
	Permissions map[string]int `json:"permissions"`
	GetUserResponse
}

type GetUserInfoSmall struct {
	UserID    string `json:"userId"`
	Email     string `json:"email"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

type GetMergeRequestResponse struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`
	Email  string `json:"email"`
}

type GetUserCountResponse struct {
	Count int `json:"count"`
}

type SetPasswordRequest struct {
	Password string `json:"password" validate:"required,min=8,max=64"`
}

type InitMergeUsersRequest struct {
	Email string `json:"email" validate:"required,email,max=256"`
}

type GenerateTotpResponse struct {
	Image   string `json:"image"`
	StateID string `json:"stateId"`
}

type GetTotpSecretResponse struct {
	Secret string `json:"secret"`
}

type ValidateTotpRequest struct {
	Code    string `json:"code" validate:"required,len=6,numeric"`
	StateID string `json:"stateId" validate:"required,uuid4"`
}

func isServiceAccountType(accountType int) bool {
	return AccountType(accountType).IsServiceAccount()
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func (router *UserRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/passkey/", router.listPasskeys).Methods("GET")
	s.HandleFunc("/passkey/registration/begin", router.beginPasskeyRegistration).Methods("POST")
	s.HandleFunc("/passkey/registration/finish", router.finishPasskeyRegistration).Methods("POST")
	s.HandleFunc("/passkey/{id}", router.renamePasskey).Methods("PUT")
	s.HandleFunc("/passkey/{id}", router.deletePasskey).Methods("DELETE")
	s.HandleFunc("/totp/generate", router.generateTotp).Methods("GET")
	s.HandleFunc("/totp/{stateId}/secret", router.getTotpSecret).Methods("GET")
	s.HandleFunc("/totp/validate", router.validateTotp).Methods("POST")
	s.HandleFunc("/totp/disable", router.disableTotp).Methods("POST")
	s.HandleFunc("/{id}/passkeys", router.adminResetPasskeys).Methods("DELETE")
	s.HandleFunc("/{id}/totp", router.adminResetTotp).Methods("DELETE")
	s.HandleFunc("/{id}/api-token", router.getApiToken).Methods("GET")
	s.HandleFunc("/{id}/api-token", router.generateApiToken).Methods("POST")
	s.HandleFunc("/{id}/api-token", router.revokeApiToken).Methods("DELETE")
	s.HandleFunc("/{id}/roles", router.getRoles).Methods("GET")
	s.HandleFunc("/{id}/roles", router.setRoles).Methods("PUT")
	s.HandleFunc("/{id}/permissions", router.getPermissions).Methods("GET")
	s.HandleFunc("/merge/init", router.mergeInit).Methods("POST")
	s.HandleFunc("/merge/finish/{id}", router.mergeFinish).Methods("POST")
	s.HandleFunc("/merge", router.getMergeRequests).Methods("GET")
	s.HandleFunc("/count", router.getCount).Methods("GET")
	s.HandleFunc("/session", router.getActiveSessions).Methods("GET")
	s.HandleFunc("/me", router.getSelf).Methods("GET")
	s.HandleFunc("/{id}", router.getOne).Methods("GET")
	s.HandleFunc("/byEmail/{email}", router.getOneByEmail).Methods("GET")
	s.HandleFunc("/{id}/password", router.setPassword).Methods("PUT")
	s.HandleFunc("/{id}", router.update).Methods("PUT")
	s.HandleFunc("/{id}", router.delete).Methods("DELETE")
	s.HandleFunc("/", router.create).Methods("POST")
	s.HandleFunc("/", router.getAll).Methods("GET")
}

func (router *UserRouter) adminResetPasskeys(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !HasPermission(user, user.OrganizationID, PermissionUsers, PermissionLevelAdmin) {
		SendForbidden(w)
		return
	}
	vars := mux.Vars(r)
	e, err := GetUserRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	if e.OrganizationID != user.OrganizationID {
		SendForbidden(w)
		return
	}
	if err := GetPasskeyRepository().DeleteAllByUserID(e.ID); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *UserRouter) adminResetTotp(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !HasPermission(user, user.OrganizationID, PermissionUsers, PermissionLevelAdmin) {
		SendForbidden(w)
		return
	}
	vars := mux.Vars(r)
	e, err := GetUserRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	if e.OrganizationID != user.OrganizationID {
		SendForbidden(w)
		return
	}
	e.TotpSecret = NullString("")
	if err := GetUserRepository().Update(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *UserRouter) disableTotp(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if user == nil {
		SendUnauthorized(w)
		return
	}
	if IsTotpEnforcedForUser(user) {
		SendForbidden(w)
		return
	}
	user.TotpSecret = NullString("")
	if err := GetUserRepository().Update(user); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *UserRouter) getTotpSecret(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if user == nil {
		SendUnauthorized(w)
		return
	}

	vars := mux.Vars(r)
	stateID := vars["stateId"]

	authState, err := GetAuthStateRepository().GetOne(stateID)
	if err != nil || authState == nil || authState.AuthStateType != AuthTotpSetup || authState.AuthProviderID != user.ID {
		SendNotFound(w)
		return
	}

	if time.Now().After(authState.Expiry) {
		GetAuthStateRepository().Delete(authState)
		totpAttemptsTracker.clearAttempts(stateID)
		SendNotFound(w)
		return
	}

	// Rate limiting to prevent abuse
	if !totpAttemptsTracker.recordAttempt(stateID + ":secret") {
		SendTooManyRequests(w)
		return
	}

	res := &GetTotpSecretResponse{
		Secret: authState.Payload,
	}
	SendJSON(w, res)
}

func (router *UserRouter) validateTotp(w http.ResponseWriter, r *http.Request) {
	var m ValidateTotpRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	user := GetRequestUser(r)
	if user == nil {
		SendUnauthorized(w)
		return
	}
	authState, err := GetAuthStateRepository().GetOne(m.StateID)
	if err != nil || authState == nil || authState.AuthStateType != AuthTotpSetup || authState.AuthProviderID != user.ID {
		SendNotFound(w)
		return
	}
	if time.Now().After(authState.Expiry) {
		GetAuthStateRepository().Delete(authState)
		totpAttemptsTracker.clearAttempts(m.StateID)
		SendNotFound(w)
		return
	}

	// Check rate limiting before validation
	if !totpAttemptsTracker.recordAttempt(m.StateID) {
		GetAuthStateRepository().Delete(authState)
		totpAttemptsTracker.clearAttempts(m.StateID)
		SendTooManyRequests(w)
		return
	}

	valid, err := totp.ValidateCustom(m.Code, authState.Payload, time.Now(), *TotpOptions)
	if err != nil || !valid {
		SendBadRequest(w)
		return
	}

	// Clear attempts on success
	GetAuthStateRepository().Delete(authState)
	totpAttemptsTracker.clearAttempts(m.StateID)

	encryptedTotpSecret, err := EncryptString(authState.Payload)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	user.TotpSecret = NullString(encryptedTotpSecret)
	if err := GetUserRepository().Update(user); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *UserRouter) generateTotp(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if user == nil {
		SendUnauthorized(w)
		return
	}
	org, err := GetOrganizationRepository().GetOne(user.OrganizationID)
	if org == nil || err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	opts := totp.GenerateOpts{
		Issuer:      "Seatsurfing for " + org.Name,
		AccountName: GetRequestUser(r).Email,
	}
	key, err := totp.Generate(opts)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	img, err := key.Image(256, 256)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	authState := &AuthState{
		AuthProviderID: user.ID,
		Expiry:         time.Now().Add(time.Minute * 5),
		AuthStateType:  AuthTotpSetup,
		Payload:        key.Secret(),
	}
	if err := GetAuthStateRepository().Create(authState); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	imageBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	res := &GenerateTotpResponse{
		Image:   imageBase64,
		StateID: authState.ID,
	}
	SendJSON(w, res)
}

func (router *UserRouter) getMergeRequests(w http.ResponseWriter, r *http.Request) {
	target := GetRequestUser(r)
	list, err := GetAuthStateRepository().GetByAuthProviderID(target.ID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []*GetMergeRequestResponse{}
	for _, e := range list {
		source, err := GetUserRepository().GetOne(e.Payload)
		if err == nil && source != nil {
			m := &GetMergeRequestResponse{
				ID:     e.ID,
				UserID: source.ID,
				Email:  source.Email,
			}
			res = append(res, m)
		}
	}
	SendJSON(w, res)
}

func (router *UserRouter) mergeInit(w http.ResponseWriter, r *http.Request) {
	var m InitMergeUsersRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	source := GetRequestUser(r)
	target, err := GetUserRepository().GetByEmail(source.OrganizationID, m.Email)
	if err != nil || target == nil {
		SendNotFound(w)
		return
	}
	authState := &AuthState{
		AuthProviderID: target.ID,
		Expiry:         time.Now().Add(time.Minute * 60),
		AuthStateType:  AuthMergeRequest,
		Payload:        source.ID,
	}
	if err := GetAuthStateRepository().Create(authState); err != nil {
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *UserRouter) mergeFinish(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	target := GetRequestUser(r)
	authState, err := GetAuthStateRepository().GetOne(vars["id"])
	if err != nil || authState == nil || authState.AuthStateType != AuthMergeRequest || authState.AuthProviderID != target.ID {
		SendNotFound(w)
		return
	}
	source, err := GetUserRepository().GetOne(authState.Payload)
	if err != nil || source == nil {
		SendBadRequest(w)
		return
	}
	if err := GetUserRepository().MergeUsers(source, target); err != nil {
		SendInternalServerError(w)
		return
	}
	GetAuthStateRepository().Delete(authState)
	SendUpdated(w)
}

func (router *UserRouter) getCount(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !HasPermission(user, user.OrganizationID, PermissionUsers, PermissionLevelRead) {
		SendForbidden(w)
		return
	}
	num, _ := GetUserRepository().GetCount(user.OrganizationID)
	m := &GetUserCountResponse{
		Count: num,
	}
	SendJSON(w, m)
}

func (router *UserRouter) setPassword(w http.ResponseWriter, r *http.Request) {
	var m SetPasswordRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}

	if !ValidatePassword(m.Password) {
		SendBadRequest(w)
		return
	}

	vars := mux.Vars(r)
	user := GetRequestUser(r)
	e := user
	if vars["id"] != "me" {
		eUser, err := GetUserRepository().GetOne(vars["id"])
		if err != nil {
			SendBadRequest(w)
			return
		}
		e = eUser
	}
	if !HasPermission(user, e.OrganizationID, PermissionUsers, PermissionLevelAdmin) && (user.ID != e.ID) {
		SendForbidden(w)
		return
	}
	e.HashedPassword = NullString(GetUserRepository().GetHashedPassword(m.Password))
	e.PasswordUpdateRequired = user.ID != e.ID
	if err := GetUserRepository().Update(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	GetSessionRepository().DeleteOfUser(e)
	SendUpdated(w)
}

func (router *UserRouter) getActiveSessions(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if user == nil {
		SendNotFound(w)
		return
	}
	sessions, err := GetSessionRepository().GetOfUser(user)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []*GetSessionResponse{}
	for _, e := range sessions {
		m := &GetSessionResponse{
			ID:      e.ID,
			UserID:  e.UserID,
			Device:  e.Device,
			Created: e.Created,
		}
		res = append(res, m)
	}
	SendJSON(w, res)
}

func (router *UserRouter) getSelf(w http.ResponseWriter, r *http.Request) {
	e := GetRequestUser(r)
	if e == nil {
		SendNotFound(w)
		return
	}
	org, err := GetOrganizationRepository().GetOne(e.OrganizationID)
	if err != nil {
		SendInternalServerError(w)
		return
	}
	passkeyCount, err := GetPasskeyRepository().GetCountByUserID(e.ID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := &GetUserSelfResponse{
		GetUserResponse: *router.copyToRestModel(e, false, passkeyCount > 0),
		Permissions:     PermissionsToRestModel(GetEffectivePermissions(e, e.OrganizationID)),
	}
	res.Organization = GetOrganizationResponse{
		ID: org.ID,
		CreateOrganizationRequest: CreateOrganizationRequest{
			Name: org.Name,
		},
	}
	primaryDomain, err := GetOrganizationRepository().GetPrimaryDomain(org)
	if err == nil && primaryDomain != nil {
		res.IsPrimaryDomain = strings.EqualFold(r.Host, primaryDomain.DomainName)
	}
	SendJSON(w, res)
}

func (router *UserRouter) getOneByEmail(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	var showNames bool = false
	if HasPermission(user, user.OrganizationID, PermissionUsers, PermissionLevelRead) {
		showNames = true
	} else {
		showNames, _ = GetSettingsRepository().GetBool(user.OrganizationID, SettingShowNames.Name)
	}

	if !showNames {
		SendForbidden(w)
		return
	}

	vars := mux.Vars(r)
	e, err := GetUserRepository().GetByEmail(user.OrganizationID, vars["email"])

	if err != nil || e.ID == user.ID {
		log.Println(err)
		SendNotFound(w)
		return
	}
	if e.OrganizationID != user.OrganizationID {
		SendForbidden(w)
		return
	}
	passkeyCount, err := GetPasskeyRepository().GetCountByUserID(e.ID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := router.copyToRestModel(e, true, passkeyCount > 0)
	SendJSON(w, res)
}

func (router *UserRouter) getOne(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !HasPermission(user, user.OrganizationID, PermissionUsers, PermissionLevelRead) {
		SendForbidden(w)
		return
	}
	vars := mux.Vars(r)
	e, err := GetUserRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	if e.OrganizationID != user.OrganizationID {
		SendForbidden(w)
		return
	}
	passkeyCount, err := GetPasskeyRepository().GetCountByUserID(e.ID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := router.copyToRestModel(e, true, passkeyCount > 0)
	SendJSON(w, res)
}

func (router *UserRouter) getAll(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("q")
	user := GetRequestUser(r)
	if !HasPermission(user, user.OrganizationID, PermissionUsers, PermissionLevelRead) {
		SendForbidden(w)
		return
	}
	var list []*User
	var err error
	if strings.TrimSpace(search) != "" {
		list, err = GetUserRepository().GetByKeyword(user.OrganizationID, strings.TrimSpace(search))
	} else {
		list, err = GetUserRepository().GetAll(user.OrganizationID, 1000, 0)
	}
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	userIDs := make([]string, len(list))
	for i, e := range list {
		userIDs[i] = e.ID
	}
	hasPasskeysByUserID, err := GetPasskeyRepository().GetUserIDsWithPasskeys(userIDs)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []*GetUserResponse{}
	for _, e := range list {
		m := router.copyToRestModel(e, true, hasPasskeysByUserID[e.ID])
		res = append(res, m)
	}
	SendJSON(w, res)
}

func (router *UserRouter) update(w http.ResponseWriter, r *http.Request) {
	var m CreateUserRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	if !isServiceAccountType(m.AccountType) && !isValidEmail(m.Email) {
		SendBadRequest(w)
		return
	}

	if !IsValidHumanName(m.Firstname) || !IsValidHumanName(m.Lastname) {
		SendBadRequest(w)
		return
	}

	if m.Password != "" && !ValidatePassword(m.Password) {
		SendBadRequest(w)
		return
	}

	vars := mux.Vars(r)
	e, err := GetUserRepository().GetOne(vars["id"])
	if err != nil {
		SendBadRequest(w)
		return
	}
	user := GetRequestUser(r)
	if !HasPermission(user, e.OrganizationID, PermissionUsers, PermissionLevelAdmin) {
		SendForbidden(w)
		return
	}

	if m.AuthProviderID != "" {
		if !ValidateGUID(m.AuthProviderID) {
			SendBadRequest(w)
			return
		}
		authProvider, _ := GetAuthProviderRepository().GetOneByOrgId(m.AuthProviderID, user.OrganizationID)
		if authProvider == nil {
			SendBadRequest(w)
			return
		}
	}

	eNew := router.copyFromRestModel(&m)
	eNew.ID = e.ID
	if user.ID == e.ID {
		// Nobody turns their own account into a service account, which would
		// lock them out of the web interface.
		eNew.AccountType = e.AccountType
	}
	eNew.OrganizationID = e.OrganizationID

	// Handle auth method updates
	if m.SendInvitation {
		// Admin wants to send invitation - reset auth to pending state
		eNew.HashedPassword = NullString("")
		eNew.AuthProviderID = NullUUID("")
		eNew.PasswordPending = true
		GetSessionRepository().DeleteOfUser(e)
	} else if m.Password != "" {
		// Admin provided a new password - update it
		eNew.HashedPassword = NullString(GetUserRepository().GetHashedPassword(m.Password))
		eNew.AuthProviderID = NullUUID("")
		eNew.PasswordPending = false
		GetSessionRepository().DeleteOfUser(e)
	} else if m.AuthProviderID != "" {
		// Admin set an auth provider - update it
		eNew.HashedPassword = NullString("")
		eNew.AuthProviderID = NullUUID(m.AuthProviderID)
		eNew.PasswordPending = false
		if m.AuthProviderID != string(e.AuthProviderID) {
			GetSessionRepository().DeleteOfUser(e)
		}
	} else {
		// No auth method change - preserve existing values
		eNew.HashedPassword = e.HashedPassword
		eNew.AuthProviderID = e.AuthProviderID
		eNew.PasswordPending = e.PasswordPending
		eNew.PasswordUpdateRequired = e.PasswordUpdateRequired
	}

	eNew.TotpSecret = e.TotpSecret
	eNew.AtlassianID = e.AtlassianID

	existingUser, err := GetUserRepository().GetByEmail(e.OrganizationID, eNew.Email)
	if err == nil && existingUser != nil {
		if existingUser.ID != e.ID {
			SendAlreadyExistsCode(w, ResponseCodeUserAlreadyExists)
			return
		}
	}
	if err := GetUserRepository().Update(eNew); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}

	// Send invitation email if requested
	if m.SendInvitation {
		org, err := GetOrganizationRepository().GetOne(e.OrganizationID)
		if err != nil {
			log.Println("Failed to get organization for user invitation:", err)
			SendInternalServerError(w)
			return
		}
		authState := &AuthState{
			AuthProviderID: GetSettingsRepository().GetNullUUID(),
			Expiry:         time.Now().Add(time.Hour * 72), // 3 days
			AuthStateType:  AuthInviteUser,
			Payload:        eNew.ID,
		}
		if err := GetAuthStateRepository().Create(authState); err != nil {
			log.Println("Failed to create auth state for user invitation:", err)
			SendInternalServerError(w)
			return
		}
		authRouter := &AuthRouter{}
		if err := authRouter.SendUserInvitationEmail(eNew, authState.ID, org); err != nil {
			log.Printf("User invitation email failed: %s\n", err)
			SendInternalServerError(w)
			return
		}
	}

	SendUpdated(w)
}

func (router *UserRouter) delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetUserRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	user := GetRequestUser(r)
	if !HasPermission(user, e.OrganizationID, PermissionUsers, PermissionLevelAdmin) || e.ID == user.ID {
		SendForbidden(w)
		return
	}
	// Refuse before deleting rather than repairing afterwards: this is the
	// organization's last administrator.
	if !CheckOrgRetainsAdmin(w, e.OrganizationID, e.ID) {
		return
	}
	if err := GetUserRepository().Delete(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *UserRouter) create(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !HasPermission(user, user.OrganizationID, PermissionUsers, PermissionLevelAdmin) {
		SendForbidden(w)
		return
	}
	var m CreateUserRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	if !isServiceAccountType(m.AccountType) && !isValidEmail(m.Email) {
		SendBadRequest(w)
		return
	}

	if !IsValidHumanName(m.Firstname) || !IsValidHumanName(m.Lastname) {
		SendBadRequest(w)
		return
	}

	if m.Password != "" && !ValidatePassword(m.Password) {
		SendBadRequest(w)
		return
	}

	if m.OrganizationID != "" && m.OrganizationID != user.OrganizationID {
		SendForbidden(w)
		return
	}

	if m.AuthProviderID != "" {
		if !ValidateGUID(m.AuthProviderID) {
			SendBadRequest(w)
			return
		}
		authProvider, _ := GetAuthProviderRepository().GetOneByOrgId(m.AuthProviderID, user.OrganizationID)
		if authProvider == nil {
			SendBadRequest(w)
			return
		}
	}

	e := router.copyFromRestModel(&m)
	e.OrganizationID = user.OrganizationID
	org, err := GetOrganizationRepository().GetOne(e.OrganizationID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	if !GetUserRepository().CanCreateUser(org) {
		SendPaymentRequired(w)
		return
	}
	existingUser, err := GetUserRepository().GetByEmail(e.OrganizationID, e.Email)
	if err == nil && existingUser != nil {
		SendAlreadyExistsCode(w, ResponseCodeUserAlreadyExists)
		return
	}
	if err := GetUserRepository().Create(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}

	// Send invitation email if requested
	if m.SendInvitation {
		authState := &AuthState{
			AuthProviderID: GetSettingsRepository().GetNullUUID(),
			Expiry:         time.Now().Add(time.Hour * 72), // 3 days
			AuthStateType:  AuthInviteUser,
			Payload:        e.ID,
		}
		if err := GetAuthStateRepository().Create(authState); err != nil {
			log.Println("Failed to create auth state for user invitation:", err)
			SendInternalServerError(w)
			return
		}
		authRouter := &AuthRouter{}
		if err := authRouter.SendUserInvitationEmail(e, authState.ID, org); err != nil {
			log.Printf("User invitation email failed: %s\n", err)
			SendInternalServerError(w)
			return
		}
	}

	SendCreated(w, e.ID)
}

func (router *UserRouter) copyFromRestModel(m *CreateUserRequest) *User {
	e := &User{}
	e.Email = m.Email
	e.Firstname = m.Firstname
	e.Lastname = m.Lastname
	e.AccountType = AccountType(m.AccountType)

	if m.SendInvitation {
		// Invitation mode: user needs to set password via email link
		e.HashedPassword = NullString("")
		e.AuthProviderID = NullUUID("")
		e.PasswordUpdateRequired = false
		e.PasswordPending = true
	} else if m.Password != "" {
		// Password mode: password provided by admin
		e.HashedPassword = NullString(GetUserRepository().GetHashedPassword(m.Password))
		e.AuthProviderID = NullUUID("")
		e.PasswordUpdateRequired = true
		e.PasswordPending = false
	} else {
		// IdP mode: user logs in via external auth provider
		e.HashedPassword = NullString("")
		e.AuthProviderID = NullUUID(m.AuthProviderID)
		e.PasswordUpdateRequired = false
		e.PasswordPending = false
	}

	e.OrganizationID = m.OrganizationID
	return e
}

func (router *UserRouter) copyToRestModel(e *User, admin bool, hasPasskeys bool) *GetUserResponse {
	m := &GetUserResponse{}
	m.ID = e.ID
	m.OrganizationID = e.OrganizationID
	m.Email = e.Email
	m.Firstname = e.Firstname
	m.Lastname = e.Lastname
	m.AtlassianID = string(e.AtlassianID)
	m.AccountType = int(e.AccountType)
	if roleIDs, err := GetUserRoleRepository().GetRoleIDsForUser(e.ID); err == nil {
		m.RoleIDs = roleIDs
	}
	if m.RoleIDs == nil {
		m.RoleIDs = []string{}
	}
	m.RequirePassword = (e.HashedPassword != "")
	m.PasswordPending = e.PasswordPending
	m.TotpEnabled = (e.TotpSecret != "")
	m.HasPasskeys = hasPasskeys
	m.LastActivity = e.LastActivityAtUTC
	if admin {
		m.AuthProviderID = string(e.AuthProviderID)
	}
	return m
}

type GenerateApiTokenResponse struct {
	Token string `json:"token"`
}

type GetApiTokenStatusResponse struct {
	Configured bool `json:"configured"`
}

func (router *UserRouter) getApiToken(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !HasPermission(user, user.OrganizationID, PermissionServiceAccounts, PermissionLevelAdmin) {
		SendForbidden(w)
		return
	}
	vars := mux.Vars(r)
	e, err := GetUserRepository().GetOne(vars["id"])
	if err != nil || e == nil {
		SendNotFound(w)
		return
	}
	if e.OrganizationID != user.OrganizationID {
		SendNotFound(w)
		return
	}
	if !e.AccountType.IsServiceAccount() {
		SendBadRequest(w)
		return
	}
	res := &GetApiTokenStatusResponse{
		Configured: e.ApiToken != "",
	}
	SendJSON(w, res)
}

func (router *UserRouter) generateApiToken(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !HasPermission(user, user.OrganizationID, PermissionServiceAccounts, PermissionLevelAdmin) {
		SendForbidden(w)
		return
	}
	vars := mux.Vars(r)
	e, err := GetUserRepository().GetOne(vars["id"])
	if err != nil || e == nil {
		SendNotFound(w)
		return
	}
	if e.OrganizationID != user.OrganizationID {
		SendNotFound(w)
		return
	}
	if !e.AccountType.IsServiceAccount() {
		SendBadRequest(w)
		return
	}
	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	rawToken := hex.EncodeToString(rawBytes)
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])
	if err := GetUserRepository().SetApiToken(e.ID, NullString(tokenHash)); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := &GenerateApiTokenResponse{
		Token: rawToken,
	}
	SendJSON(w, res)
}

func (router *UserRouter) revokeApiToken(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !HasPermission(user, user.OrganizationID, PermissionServiceAccounts, PermissionLevelAdmin) {
		SendForbidden(w)
		return
	}
	vars := mux.Vars(r)
	e, err := GetUserRepository().GetOne(vars["id"])
	if err != nil || e == nil {
		SendNotFound(w)
		return
	}
	if e.OrganizationID != user.OrganizationID {
		SendNotFound(w)
		return
	}
	if !e.AccountType.IsServiceAccount() {
		SendBadRequest(w)
		return
	}
	if err := GetUserRepository().SetApiToken(e.ID, NullString("")); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// getRoles lists the role IDs assigned to a user.
func (router *UserRouter) getRoles(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CheckPermission(w, user, user.OrganizationID, PermissionRoles, PermissionLevelRead) {
		return
	}
	e := router.loadUserInOwnOrg(w, r, user)
	if e == nil {
		return
	}
	roleIDs, err := GetUserRoleRepository().GetRoleIDsForUser(e.ID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	if roleIDs == nil {
		roleIDs = []string{}
	}
	SendJSON(w, roleIDs)
}

// setRoles replaces a user's manually assigned roles. Assignments made by an
// identity provider are left alone, so that reconciliation on the user's next
// login still governs those.
func (router *UserRouter) setRoles(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CheckPermission(w, user, user.OrganizationID, PermissionRoles, PermissionLevelAdmin) {
		return
	}
	e := router.loadUserInOwnOrg(w, r, user)
	if e == nil {
		return
	}
	var m SetUserRolesRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	// Every role involved must be one the caller could grant outright, both
	// the ones being added and the ones the user already holds. Otherwise a
	// limited administrator could hand out access they do not have themselves,
	// or strip a role they can not see the full weight of.
	current, err := GetUserRoleRepository().GetRolesForUser(e.ID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	for _, roleID := range m.RoleIDs {
		role, err := GetRoleRepository().GetOne(roleID)
		if err != nil || role.OrganizationID != user.OrganizationID {
			SendBadRequest(w)
			return
		}
		if !CanGrantPermissions(user, user.OrganizationID, role.Permissions) {
			SendBadRequestCode(w, ResponseCodeRoleEscalationNotAllowed)
			return
		}
	}
	for _, role := range current {
		full, err := GetRoleRepository().GetOne(role.ID)
		if err != nil {
			continue
		}
		if !CanGrantPermissions(user, user.OrganizationID, full.Permissions) {
			SendBadRequestCode(w, ResponseCodeRoleEscalationNotAllowed)
			return
		}
	}
	// Evaluate the outcome before writing: a refused request must leave the
	// assignments untouched.
	resulting := ResultingPermissions(e, m.RoleIDs)

	// A user must not be able to sign away their own ability to manage roles.
	if e.ID == user.ID && resulting[PermissionRoles] < PermissionLevelAdmin {
		SendBadRequestCode(w, ResponseCodeRoleCannotRemoveOwnAccess)
		return
	}
	// The organization must keep at least one administrator. If this user
	// still qualifies afterwards there is nothing to check; otherwise somebody
	// else must.
	stillAdmin := true
	for p, level := range AdminRetentionPermissions {
		if resulting[p] < level {
			stillAdmin = false
			break
		}
	}
	if !stillAdmin && !CheckOrgRetainsAdmin(w, user.OrganizationID, e.ID) {
		return
	}

	if err := GetUserRoleRepository().SetRolesForUser(e.ID, m.RoleIDs, RoleAssignmentSourceManual); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

// getPermissions returns a user's resolved access. Administrators use it to
// see the effect of an assignment before relying on it; every user may read
// their own.
func (router *UserRouter) getPermissions(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	vars := mux.Vars(r)
	if vars["id"] != user.ID && !CheckPermission(w, user, user.OrganizationID, PermissionRoles, PermissionLevelRead) {
		return
	}
	e := router.loadUserInOwnOrg(w, r, user)
	if e == nil {
		return
	}
	SendJSON(w, &GetUserPermissionsResponse{
		Permissions: PermissionsToRestModel(GetEffectivePermissions(e, e.OrganizationID)),
	})
}

// loadUserInOwnOrg resolves the user named in the path, refusing anything
// outside the caller's organization.
func (router *UserRouter) loadUserInOwnOrg(w http.ResponseWriter, r *http.Request, requestUser *User) *User {
	vars := mux.Vars(r)
	e, err := GetUserRepository().GetOne(vars["id"])
	if err != nil || e == nil || e.OrganizationID != requestUser.OrganizationID {
		SendNotFound(w)
		return nil
	}
	return e
}
