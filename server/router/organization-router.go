package router

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/config"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

type OrganizationRouter struct {
}

type CreateOrganizationRequest struct {
	Name      string `json:"name" validate:"required"`
	Firstname string `json:"firstname" validate:"required"`
	Lastname  string `json:"lastname" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Language  string `json:"language" validate:"required,len=2"`
}

type GetOrganizationResponse struct {
	ID string `json:"id"`
	CreateOrganizationRequest
}

type GetDomainResponse struct {
	DomainName  string     `json:"domain"`
	Active      bool       `json:"active"`
	VerifyToken string     `json:"verifyToken"`
	Primary     bool       `json:"primary"`
	Accessible  bool       `json:"accessible"`
	AccessCheck *time.Time `json:"accessCheck"`
}

type ChangeOrgEmailPayload struct {
	OrgID string `json:"orgId" validate:"required,uuid"`
	Email string `json:"email" validate:"required,email,max=254"`
	Code  int    `json:"code" validate:"required,numeric,len=6"`
}

type ChangeOrgEmailResponse struct {
	VerifyUUID string `json:"verifyUuid"`
}

type ChangeEmailAddressVerifyRequest struct {
	Code string `json:"code" validate:"required,numeric,len=6"`
}

type CompleteOrgDeletionRequest struct {
	Code string `json:"code" validate:"required,numeric,len=6"`
}

type DeleteOrgResponse struct {
	Code string `json:"code"`
}

type AuthStateOrgDeletionRequestPayload struct {
	OrganizationID string `json:"organizationId" validate:"required,uuid"`
	Code           string `json:"code" validate:"required,numeric,len=6"`
}

func (router *OrganizationRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/domain/verify/{domain}", router.getDomainAccessibilityToken).Methods("GET")
	s.HandleFunc("/domain/{domain}", router.getOrgForDomain).Methods("GET")
	s.HandleFunc("/{id}/domain/", router.getDomains).Methods("GET")
	s.HandleFunc("/{id}/domain/{domain}/verify", router.verifyDomain).Methods("POST")
	s.HandleFunc("/{id}/domain/{domain}/primary", router.setPrimaryDomain).Methods("POST")
	s.HandleFunc("/{id}/domain/{domain}", router.removeDomain).Methods("DELETE")
	s.HandleFunc("/{id}/domain/{domain}", router.addDomain).Methods("POST")
	s.HandleFunc("/{id}/verifyemail/{uuid}", router.verifyEmail).Methods("POST")
	s.HandleFunc("/{id}", router.getOne).Methods("GET")
	s.HandleFunc("/{id}", router.update).Methods("PUT")
	s.HandleFunc("/{id}", router.delete).Methods("DELETE")
	s.HandleFunc("/", router.create).Methods("POST")
	s.HandleFunc("/", router.getAll).Methods("GET")
	s.HandleFunc("/deleteorg/{id}", router.completeOrgDeletion).Methods("POST")
}

func (router *OrganizationRouter) getDomainAccessibilityToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	domain := vars["domain"]
	if domain == "" {
		SendBadRequest(w)
		return
	}
	// Check if domain exists in activated state in ANY org already
	org, err := GetOrganizationRepository().GetOneByDomain(domain)
	if err != nil || org == nil {
		SendNotFound(w)
		return
	}
	res := &DomainAccessibilityPayload{
		Domain: domain,
		OrgID:  org.ID,
		Status: "ok",
	}
	SendJSON(w, res)
}

func (router *OrganizationRouter) getOrgForDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetOrganizationRepository().GetOneByDomain(vars["domain"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	res := &GetOrganizationResponse{}
	res.ID = e.ID
	res.Name = e.Name
	SendJSON(w, res)
}

func (router *OrganizationRouter) getOne(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetOrganizationRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	user := GetRequestUser(r)
	if !(GetUserRepository().IsSuperAdmin(user) || CanAdminOrg(user, e.ID)) {
		SendForbidden(w)
		return
	}
	res := router.copyToRestModel(e)
	SendJSON(w, res)
}

func (router *OrganizationRouter) getAll(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !GetUserRepository().IsSuperAdmin(user) {
		SendForbidden(w)
		return
	}
	list, err := GetOrganizationRepository().GetAll()
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []*GetOrganizationResponse{}
	for _, e := range list {
		m := router.copyToRestModel(e)
		res = append(res, m)
	}
	SendJSON(w, res)
}

func (router *OrganizationRouter) getDomains(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetOrganizationRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	user := GetRequestUser(r)
	if !(GetUserRepository().IsSuperAdmin(user) || CanAdminOrg(user, e.ID)) {
		SendForbidden(w)
		return
	}
	list, err := GetOrganizationRepository().GetDomains(e)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := []*GetDomainResponse{}
	for _, domain := range list {
		item := &GetDomainResponse{
			DomainName:  domain.DomainName,
			Active:      domain.Active,
			VerifyToken: domain.VerifyToken,
			Primary:     domain.Primary,
			Accessible:  domain.Accessible,
			AccessCheck: domain.AccessCheck,
		}
		res = append(res, item)
	}
	SendJSON(w, res)
}

func (router *OrganizationRouter) addDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetOrganizationRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	user := GetRequestUser(r)
	if !(GetUserRepository().IsSuperAdmin(user) || CanAdminOrg(user, e.ID)) {
		SendForbidden(w)
		return
	}
	featureCustomDomains, _ := GetSettingsRepository().GetBool(e.ID, SettingFeatureCustomDomains.Name)
	if !featureCustomDomains {
		SendPaymentRequired(w)
		return
	}
	domainName := strings.TrimSpace(strings.ToLower(vars["domain"]))
	// Check if domain is special
	if strings.HasSuffix(domainName, ".seatsurfing.app") || strings.HasSuffix(domainName, ".seatsurfing.io") {
		SendBadRequest(w)
		return
	}
	// Check if domain exists in this org already
	domain, _ := GetOrganizationRepository().GetDomain(e, domainName)
	if domain != nil {
		SendAlreadyExists(w)
		return
	}
	// Check if domain exists in activated state in ANY org already
	someOrg, _ := GetOrganizationRepository().GetOneByDomain(domainName)
	if someOrg != nil {
		SendAlreadyExists(w)
		return
	}
	// Add domain
	err = GetOrganizationRepository().AddDomain(e, domainName, GetUserRepository().IsSuperAdmin(user))
	if err != nil {
		log.Println(err)
		SendAlreadyExists(w)
		return
	}
	router.ensureOrgHasPrimaryDomain(e, domainName)
	SendCreated(w, domainName)
}

func (router *OrganizationRouter) verifyEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetOrganizationRepository().GetOne(vars["id"])
	if err != nil || e == nil {
		SendNotFound(w)
		return
	}
	user := GetRequestUser(r)
	if !GetUserRepository().IsSuperAdmin(user) && !CanAdminOrg(user, e.ID) {
		SendForbidden(w)
		return
	}
	authState, err := GetAuthStateRepository().GetOne(vars["uuid"])
	if err != nil || authState == nil {
		SendNotFound(w)
		return
	}
	var m ChangeEmailAddressVerifyRequest
	if err := UnmarshalValidateBody(r, &m); err != nil {
		SendBadRequest(w)
		return
	}
	var authStatePayload ChangeOrgEmailPayload
	if err := json.Unmarshal([]byte(authState.Payload), &authStatePayload); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	if authStatePayload.OrgID != e.ID {
		log.Println("AuthState payload does not match organization ID")
		SendNotFound(w)
		return
	}
	if strconv.Itoa(authStatePayload.Code) != m.Code {
		log.Println("Invalid verification code")
		SendNotFound(w)
		return
	}
	e.ContactEmail = authStatePayload.Email
	if err := GetOrganizationRepository().Update(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *OrganizationRouter) verifyDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetOrganizationRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	user := GetRequestUser(r)
	if !(GetUserRepository().IsSuperAdmin(user) || CanAdminOrg(user, e.ID)) {
		SendForbidden(w)
		return
	}
	domain, err := GetOrganizationRepository().GetDomain(e, vars["domain"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	if domain.Active {
		return
	}
	// Check if domain exists in activated state in ANY org already
	someOrg, _ := GetOrganizationRepository().GetOneByDomain(vars["domain"])
	if someOrg != nil {
		SendAlreadyExists(w)
		return
	}
	if !IsValidTXTRecord(domain.DomainName, domain.VerifyToken) {
		SendBadRequest(w)
		return
	}
	err = GetOrganizationRepository().ActivateDomain(e, domain.DomainName)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}

func (router *OrganizationRouter) setPrimaryDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetOrganizationRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	user := GetRequestUser(r)
	if !(GetUserRepository().IsSuperAdmin(user) || CanAdminOrg(user, e.ID)) {
		SendForbidden(w)
		return
	}
	if _, err = GetOrganizationRepository().GetDomain(e, vars["domain"]); err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	GetOrganizationRepository().SetPrimaryDomain(e, vars["domain"])
	SendUpdated(w)
}

func (router *OrganizationRouter) removeDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetOrganizationRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	user := GetRequestUser(r)
	if !(GetUserRepository().IsSuperAdmin(user) || CanAdminOrg(user, e.ID)) {
		SendForbidden(w)
		return
	}
	// prevent removing signup domain
	if strings.HasSuffix(vars["domain"], ".seatsurfing.app") {
		SendForbidden(w)
		return
	}
	err = GetOrganizationRepository().RemoveDomain(e, vars["domain"])
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	router.ensureOrgHasPrimaryDomain(e, "")
	SendUpdated(w)
}

func (router *OrganizationRouter) update(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	vars := mux.Vars(r)
	if !GetUserRepository().IsSuperAdmin(user) && !CanAdminOrg(user, vars["id"]) {
		SendForbidden(w)
		return
	}
	var m CreateOrganizationRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}
	e, err := GetOrganizationRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	eIncoming := router.copyFromRestModel(&m)
	e.Name = eIncoming.Name
	e.Language = eIncoming.Language
	e.ContactFirstname = eIncoming.ContactFirstname
	e.ContactLastname = eIncoming.ContactLastname

	res := &ChangeOrgEmailResponse{
		VerifyUUID: "",
	}
	if !GetUserRepository().IsSuperAdmin(user) && CanAdminOrg(user, vars["id"]) && !strings.EqualFold(e.ContactEmail, eIncoming.ContactEmail) {
		payload := &ChangeOrgEmailPayload{
			OrgID: e.ID,
			Email: eIncoming.ContactEmail,
			Code:  GetRandomNumber(100000, 999999), // Random 6-digit code
		}
		json, _ := json.Marshal(payload)
		authState := &AuthState{
			AuthStateType:  AuthChangeOrgEmail,
			Payload:        string(json),
			Expiry:         time.Now().Add(time.Minute * 5),
			AuthProviderID: GetSettingsRepository().GetNullUUID(),
		}
		if err := GetAuthStateRepository().Create(authState); err != nil {
			log.Println(err)
			SendInternalServerError(w)
			return
		}
		res.VerifyUUID = authState.ID
		if err := router.sendVerifyEmailAddressEmail(e, eIncoming.ContactEmail, strconv.Itoa(payload.Code)); err != nil {
			log.Println(err)
			SendInternalServerError(w)
			return
		}
	} else {
		e.ContactEmail = eIncoming.ContactEmail
	}
	if err := GetOrganizationRepository().Update(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendJSON(w, res)
}

func (router *OrganizationRouter) sendVerifyEmailAddressEmail(org *Organization, newEmail string, code string) error {
	vars := map[string]string{
		"recipientName":  org.ContactFirstname + " " + org.ContactLastname,
		"recipientEmail": newEmail,
		"code":           code,
	}
	return SendEmailWithOrg(&MailAddress{Address: newEmail}, GetEmailTemplatePathChangeEmailAddress(), org.Language, vars, org.ID)
}

func (router *OrganizationRouter) delete(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)

	// user needs to be org admin or super user
	if !(GetUserRepository().IsSuperAdmin(user) || CanAdminOrg(user, user.OrganizationID)) {
		SendForbidden(w)
		return
	}

	// if no super user: check global "org delete" setting
	if !GetUserRepository().IsSuperAdmin(user) && CanAdminOrg(user, user.OrganizationID) {
		if !GetConfig().AllowOrgDelete {
			SendForbidden(w)
		}
	}

	vars := mux.Vars(r)
	e, err := GetOrganizationRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}

	// send confirmation mail
	code := strconv.Itoa(GetRandomNumber(100000, 999999)) // Random 6-digit code
	payload := &AuthStateOrgDeletionRequestPayload{
		OrganizationID: e.ID,
		Code:           code,
	}
	authState := &AuthState{
		AuthProviderID: GetSettingsRepository().GetNullUUID(),
		Expiry:         time.Now().Add(time.Hour * 1),
		AuthStateType:  AuthDeleteOrg,
		Payload:        marshalAuthStateOrgDeletionRequestPayload(payload),
	}
	GetAuthStateRepository().Create(authState)
	if err := router.SendOrgConfirmDeleteOrgEmail(user, authState.ID, e); err != nil {
		log.Printf("Sending confirm org delete email failed: %s\n", err)
		SendInternalServerError(w)
		return
	}

	res := DeleteOrgResponse{
		Code: code,
	}
	SendJSON(w, res)
}

func marshalAuthStateOrgDeletionRequestPayload(payload *AuthStateOrgDeletionRequestPayload) string {
	json, _ := json.Marshal(payload)
	return string(json)
}

func unmarshalAuthStateOrgDeletionRequestPayload(payload string) *AuthStateOrgDeletionRequestPayload {
	var o *AuthStateOrgDeletionRequestPayload
	json.Unmarshal([]byte(payload), &o)
	return o
}

func (router *OrganizationRouter) completeOrgDeletion(w http.ResponseWriter, r *http.Request) {
	if !GetConfig().AllowOrgDelete {
		SendNotFound(w)
		return
	}
	var m CompleteOrgDeletionRequest
	if UnmarshalValidateBody(r, &m) != nil {
		SendBadRequest(w)
		return
	}

	// test auth state
	vars := mux.Vars(r)
	authState, err := GetAuthStateRepository().GetOne(vars["id"])
	if err != nil {
		SendNotFound(w)
		return
	}
	if authState.AuthStateType != AuthDeleteOrg {
		SendNotFound(w)
		return
	}
	payload := unmarshalAuthStateOrgDeletionRequestPayload(authState.Payload)
	if payload.Code != m.Code {
		SendNotFound(w)
		return
	}

	// (finally) delete organization
	organization, err := GetOrganizationRepository().GetOne(payload.OrganizationID)
	if organization == nil || err != nil {
		SendNotFound(w)
		return
	}
	GetOrganizationRepository().Delete(organization)

	SendUpdated(w)
}

func (router *OrganizationRouter) SendOrgConfirmDeleteOrgEmail(user *User, ID string, org *Organization) error {
	domain, err := GetOrganizationRepository().GetPrimaryDomain(org)
	if err != nil {
		return err
	}
	vars := map[string]string{
		"recipientName":  user.GetSafeRecipientName(),
		"recipientEmail": user.Email,
		"confirmID":      ID,
		"orgDomain":      FormatURL(domain.DomainName) + "/",
		"orgName":        org.Name,
	}
	return SendEmailWithOrg(&MailAddress{Address: user.Email}, GetEmailTemplatePathConfirmDeleteOrg(), org.Language, vars, org.ID)
}

func (router *OrganizationRouter) create(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !GetUserRepository().IsSuperAdmin(user) {
		SendForbidden(w)
		return
	}
	var m CreateOrganizationRequest
	if err := UnmarshalValidateBody(r, &m); err != nil {
		SendBadRequest(w)
		return
	}
	e := router.copyFromRestModel(&m)
	e.SignupDate = time.Now()
	if err := GetOrganizationRepository().Create(e); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendCreated(w, e.ID)
}

func (router *OrganizationRouter) ensureOrgHasPrimaryDomain(e *Organization, favoritePrimaryDomain string) {
	domains, _ := GetOrganizationRepository().GetDomains(e)
	hasPrimary := false
	for _, domain := range domains {
		if domain.Primary {
			hasPrimary = true
			break
		}
	}
	if !hasPrimary {
		if favoritePrimaryDomain != "" {
			GetOrganizationRepository().SetPrimaryDomain(e, favoritePrimaryDomain)
		} else {
			domain, err := GetOrganizationRepository().GetPrimaryDomain(e)
			if err == nil && domain != nil {
				GetOrganizationRepository().SetPrimaryDomain(e, domain.DomainName)
			}
		}
	}
}

func (router *OrganizationRouter) copyFromRestModel(m *CreateOrganizationRequest) *Organization {
	e := &Organization{}
	e.Name = m.Name
	e.ContactFirstname = m.Firstname
	e.ContactLastname = m.Lastname
	e.ContactEmail = m.Email
	e.Language = m.Language
	return e
}

func (router *OrganizationRouter) copyToRestModel(e *Organization) *GetOrganizationResponse {
	m := &GetOrganizationResponse{}
	m.ID = e.ID
	m.Name = e.Name
	m.Firstname = e.ContactFirstname
	m.Lastname = e.ContactLastname
	m.Email = e.ContactEmail
	m.Language = e.Language
	return m
}
