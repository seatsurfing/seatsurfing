package router

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pquerna/otp/totp"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

type UserRouter struct {
}

type GetSessionResponse struct {
	ID      string    `json:"id"`
	UserID  string    `json:"userId"`
	Device  string    `json:"device"`
	Created time.Time `json:"created"`
}

type CreateUserRequest struct {
	Email          string `json:"email" validate:"required,email,max=254"`
	Firstname      string `json:"firstname" validate:"required,max=128"`
	Lastname       string `json:"lastname" validate:"required,max=128"`
	AtlassianID    string `json:"atlassianId"`
	Role           int    `json:"role"`
	AuthProviderID string `json:"authProviderId"`
	Password       string `json:"password"`
	OrganizationID string `json:"organizationId"`
}

type GetUserResponse struct {
	ID              string                  `json:"id"`
	Organization    GetOrganizationResponse `json:"organization"`
	RequirePassword bool                    `json:"requirePassword"`
	SpaceAdmin      bool                    `json:"spaceAdmin"`
	OrgAdmin        bool                    `json:"admin"`
	SuperAdmin      bool                    `json:"superAdmin"`
	TotpEnabled     bool                    `json:"totpEnabled"`
	CreateUserRequest
}

type GetUserInfoSmall struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
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
	Email string `json:"email" validate:"required,email,max=254"`
}

type GenerateTotpResponse struct {
	Secret  string `json:"secret"`
	Image   string `json:"image"`
	StateID string `json:"stateId"`
}

type ValidateTotpRequest struct {
	Code    string `json:"code" validate:"required,len=6,numeric"`
	StateID string `json:"stateId" validate:"required,uuid4"`
}

func (router *UserRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/totp/generate", router.generateTotp).Methods("GET")
	s.HandleFunc("/totp/validate", router.validateTotp).Methods("POST")
	s.HandleFunc("/totp/disable", router.disableTotp).Methods("POST")
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

func (router *UserRouter) disableTotp(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if user == nil {
		SendUnauthorized(w)
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
		SendNotFound(w)
		return
	}
	valid := totp.Validate(m.Code, authState.Payload)
	GetAuthStateRepository().Delete(authState)
	if !valid {
		SendBadRequest(w)
		return
	}
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
		Secret:  key.Secret(),
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
	if !CanAdminOrg(user, user.OrganizationID) {
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
	if !CanAdminOrg(user, e.OrganizationID) && (user.ID != e.ID) {
		SendForbidden(w)
		return
	}
	e.HashedPassword = NullString(GetUserRepository().GetHashedPassword(m.Password))
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
	res := router.copyToRestModel(e, false)
	res.Organization = GetOrganizationResponse{
		ID: org.ID,
		CreateOrganizationRequest: CreateOrganizationRequest{
			Name: org.Name,
		},
	}
	SendJSON(w, res)
}

func (router *UserRouter) getOneByEmail(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	var showNames bool = false
	if CanSpaceAdminOrg(user, user.OrganizationID) {
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
	res := router.copyToRestModel(e, true)
	SendJSON(w, res)
}

func (router *UserRouter) getOne(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CanAdminOrg(user, user.OrganizationID) {
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
	res := router.copyToRestModel(e, true)
	SendJSON(w, res)
}

func (router *UserRouter) getAll(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("q")
	user := GetRequestUser(r)
	if !CanSpaceAdminOrg(user, user.OrganizationID) {
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
	res := []*GetUserResponse{}
	for _, e := range list {
		m := router.copyToRestModel(e, true)
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
	vars := mux.Vars(r)
	e, err := GetUserRepository().GetOne(vars["id"])
	if err != nil {
		SendBadRequest(w)
		return
	}
	user := GetRequestUser(r)
	if !CanAdminOrg(user, e.OrganizationID) {
		SendForbidden(w)
		return
	}
	eNew := router.copyFromRestModel(&m)
	eNew.ID = e.ID
	if eNew.Role > user.Role && eNew.Role != UserRoleServiceAccountRO && eNew.Role != UserRoleServiceAccountRW {
		eNew.Role = e.Role
	}
	eNew.OrganizationID = e.OrganizationID
	eNew.HashedPassword = e.HashedPassword
	existingUser, err := GetUserRepository().GetByEmail(e.OrganizationID, eNew.Email)
	if err == nil && existingUser != nil {
		if existingUser.ID != e.ID {
			SendAlreadyExists(w)
			return
		}
	}
	if err := GetUserRepository().Update(eNew); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
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
	if !CanAdminOrg(user, e.OrganizationID) || e.ID == user.ID {
		SendForbidden(w)
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
	if !CanAdminOrg(user, user.OrganizationID) {
		SendForbidden(w)
		return
	}
	var m CreateUserRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	if m.OrganizationID != "" && m.OrganizationID != user.OrganizationID && !GetUserRepository().IsSuperAdmin(user) {
		SendForbidden(w)
		return
	}
	e := router.copyFromRestModel(&m)
	if e.OrganizationID == "" || !GetUserRepository().IsSuperAdmin(user) {
		e.OrganizationID = user.OrganizationID
	}
	if e.Role > user.Role && e.Role != UserRoleServiceAccountRO && e.Role != UserRoleServiceAccountRW {
		e.Role = UserRoleUser
	}
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
		SendAlreadyExists(w)
		return
	}
	if err := GetUserRepository().Create(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendCreated(w, e.ID)
}

func (router *UserRouter) copyFromRestModel(m *CreateUserRequest) *User {
	e := &User{}
	e.Email = m.Email
	e.Firstname = m.Firstname
	e.Lastname = m.Lastname
	e.Role = UserRole(m.Role)
	if m.Password != "" {
		e.HashedPassword = NullString(GetUserRepository().GetHashedPassword(m.Password))
		e.AuthProviderID = NullString("")
	} else {
		e.AuthProviderID = NullString(m.AuthProviderID)
	}
	e.OrganizationID = m.OrganizationID
	return e
}

func (router *UserRouter) copyToRestModel(e *User, admin bool) *GetUserResponse {
	m := &GetUserResponse{}
	m.ID = e.ID
	m.OrganizationID = e.OrganizationID
	m.Email = e.Email
	m.Firstname = e.Firstname
	m.Lastname = e.Lastname
	m.AtlassianID = string(e.AtlassianID)
	m.Role = int(e.Role)
	m.SpaceAdmin = GetUserRepository().IsSpaceAdmin(e)
	m.OrgAdmin = GetUserRepository().IsOrgAdmin(e)
	m.SuperAdmin = GetUserRepository().IsSuperAdmin(e)
	m.RequirePassword = (e.HashedPassword != "")
	m.TotpEnabled = (e.TotpSecret != "")
	if admin {
		m.AuthProviderID = string(e.AuthProviderID)
	}
	return m
}
