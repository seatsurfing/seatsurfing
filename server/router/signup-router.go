package router

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/config"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

type SignupRouter struct {
}

type SignupRequest struct {
	HoneypotFirstname string `json:"firstname"`
	HoneypotLastname  string `json:"lastname"`
	Email             string `json:"email" validate:"required,email"`
	Organization      string `json:"organization" validate:"required"`
	Domain            string `json:"domain" validate:"required"`
	Firstname         string `json:"contactFirstname" validate:"required"`
	Lastname          string `json:"contactLastname" validate:"required"`
	Password          string `json:"password" validate:"required,min=8"`
	Language          string `json:"language" validate:"required,len=2"`
	AcceptTerms       bool   `json:"acceptTerms" validate:"required"`
}

func (router *SignupRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/confirm/{id}", router.confirm).Methods("POST")
	s.HandleFunc("/", router.signup).Methods("POST")
}

func (router *SignupRouter) signup(w http.ResponseWriter, r *http.Request) {
	var m SignupRequest
	if UnmarshalValidateBody(r, &m) != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if m.HoneypotFirstname != "" || m.HoneypotLastname != "" {
		// Honeypot, act as if everything was fine
		w.WriteHeader(http.StatusNoContent)
		return
	}
	domain := strings.ToLower(m.Domain) + GetConfig().OrgSignupDomain
	if !router.isDomainAvailable(domain) {
		w.WriteHeader(http.StatusConflict)
		return
	}
	if !router.isEmailAvailable(m.Email) {
		w.WriteHeader(http.StatusConflict)
		return
	}
	if !GetConfig().IsValidLanguageCode(m.Language) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	signup := &Signup{
		Date:         time.Now(),
		Email:        m.Email,
		Password:     GetUserRepository().GetHashedPassword(m.Password),
		Firstname:    m.Firstname,
		Lastname:     m.Lastname,
		Organization: m.Organization,
		Language:     m.Language,
		Domain:       domain,
	}
	if err := GetSignupRepository().Create(signup); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := router.sendDoubleOptInMail(signup, router.getLanguage(signup.Language)); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (router *SignupRouter) confirm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	e, err := GetSignupRepository().GetOne(vars["id"])
	if err != nil {
		log.Println(err)
		SendNotFound(w)
		return
	}
	if !router.isDomainAvailable(e.Domain) {
		GetSignupRepository().Delete(e)
		w.WriteHeader(http.StatusConflict)
		return
	}
	org := &Organization{
		Name:             e.Organization,
		ContactFirstname: e.Firstname,
		ContactLastname:  e.Lastname,
		ContactEmail:     e.Email,
		Language:         e.Language,
		SignupDate:       e.Date,
	}
	if err := GetOrganizationRepository().Create(org); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	if err := GetOrganizationRepository().AddDomain(org, e.Domain, true); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	user := &User{
		Email:          GetConfig().OrgSignupAdmin + "@" + e.Domain,
		HashedPassword: NullString(e.Password),
		OrganizationID: org.ID,
		Role:           UserRoleOrgAdmin,
	}
	if err := GetUserRepository().Create(user); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	if err := GetOrganizationRepository().CreateSampleData(org); err != nil {
		log.Println(err)
	}
	router.sendConfirmMail(e, router.getLanguage(e.Language))
	GetSignupRepository().Delete(e)
	w.WriteHeader(http.StatusNoContent)
}

func (router *SignupRouter) sendDoubleOptInMail(signup *Signup, language string) error {
	vars := map[string]string{
		"recipientName":  signup.Firstname + " " + signup.Lastname,
		"recipientEmail": signup.Email,
		"confirmID":      signup.ID,
	}
	return SendEmail(signup.Email, GetConfig().SMTPSenderAddress, GetEmailTemplatePathSignup(), language, vars)
}

func (router *SignupRouter) sendConfirmMail(signup *Signup, language string) error {
	vars := map[string]string{
		"recipientName":  signup.Firstname + " " + signup.Lastname,
		"recipientEmail": signup.Email,
		"username":       GetConfig().OrgSignupAdmin + "@" + signup.Domain,
	}
	return SendEmail(signup.Email, GetConfig().SMTPSenderAddress, GetEmailTemplatePathConfirm(), language, vars)
}

func (router *SignupRouter) getLanguage(language string) string {
	lng := strings.ToLower(language)
	switch lng {
	case "de":
		return lng
	default:
		return "en"
	}
}

func (router *SignupRouter) isEmailAvailable(email string) bool {
	org, err := GetOrganizationRepository().GetByEmail(email)
	if (err == nil) && (org != nil) {
		return false
	}
	signup, err := GetSignupRepository().GetByEmail(email)
	if (err == nil) && (signup != nil) {
		return false
	}
	return true
}

func (router *SignupRouter) isDomainAvailable(domain string) bool {
	org, err := GetOrganizationRepository().GetOneByDomain(domain)
	if (err == nil) && (org != nil) {
		return false
	}
	return true
}
