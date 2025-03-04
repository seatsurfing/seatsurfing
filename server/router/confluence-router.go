package router

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/config"
	. "github.com/seatsurfing/seatsurfing/server/repository"
)

type ConfluenceServerClaims struct {
	UserName string `json:"user"`
	UserKey  string `json:"key"`
	jwt.StandardClaims
}

type ConfluenceRouter struct {
}

func (router *ConfluenceRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/{orgID}/{jwt}", router.serverLogin).Methods("GET")
}

func (router *ConfluenceRouter) serverLogin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	org, err := GetOrganizationRepository().GetOne(vars["orgID"])
	if err != nil || org == nil {
		SendTextNotFound(w, "text/plain", router.getOrgNotFoundBody())
		return
	}
	sharedSecret, err := GetSettingsRepository().Get(org.ID, SettingConfluenceServerSharedSecret.Name)
	if err != nil || sharedSecret == "" {
		SendBadRequest(w)
		return
	}
	claims := &ConfluenceServerClaims{}
	token, err := jwt.ParseWithClaims(vars["jwt"], claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(sharedSecret), nil
	})
	if err != nil {
		log.Println("JWT header verification failed: parsing JWT failed with: " + err.Error())
		SendTemporaryRedirect(w, GetConfig().FrontendURL+"ui/login/failed")
		return
	}
	if !token.Valid {
		log.Println("JWT header verification failed: invalid JWT")
		SendTemporaryRedirect(w, GetConfig().FrontendURL+"ui/login/failed")
		return
	}
	allowAnonymous, _ := GetSettingsRepository().GetBool(org.ID, SettingConfluenceAnonymous.Name)
	userID := router.getUserEmailServer(org, claims, allowAnonymous)
	if userID == "" {
		SendTemporaryRedirect(w, GetConfig().FrontendURL+"ui/login/confluence/anonymous")
		return
	}
	_, err = GetUserRepository().GetByAtlassianID(userID)
	if err != nil {
		// user not found using atlassianID, try by mail
		u, err := GetUserRepository().GetByEmail(org.ID, userID)
		if err == nil {
			// got it, update it now
			GetUserRepository().UpdateAtlassianClientIDForUser(u.OrganizationID, u.ID, userID)
		}
		// and load again
		GetUserRepository().GetByAtlassianID(userID)
	}
	if err != nil {
		if !GetUserRepository().CanCreateUser(org) {
			SendTemporaryRedirect(w, GetConfig().FrontendURL+"ui/login/failed")
			return
		}
		user := &User{
			Email:          userID,
			AtlassianID:    NullString(userID),
			OrganizationID: org.ID,
			Role:           UserRoleUser,
		}
		GetUserRepository().Create(user)
	}
	payload := &AuthStateLoginPayload{
		LoginType: "",
		UserID:    userID,
		LongLived: false,
	}
	authState := &AuthState{
		AuthProviderID: GetSettingsRepository().GetNullUUID(),
		Expiry:         time.Now().Add(time.Minute * 5),
		AuthStateType:  AuthAtlassian,
		Payload:        marshalAuthStateLoginPayload(payload),
	}
	if err := GetAuthStateRepository().Create(authState); err != nil {
		SendInternalServerError(w)
		return
	}
	SendTemporaryRedirect(w, GetConfig().FrontendURL+"ui/login/success/"+authState.ID)
}

func (router *ConfluenceRouter) getOrgNotFoundBody() []byte {
	var sb strings.Builder
	sb.WriteString("Instance ID could not be found at this instance: ")
	sb.WriteString(GetConfig().PublicURL)
	sb.WriteString("\n\n")
	sb.WriteString("To get your Intance ID, log in to the Admin interface (i.e. " + GetConfig().FrontendURL + "admin/), go to Settings and copy the Instance ID.")
	sb.WriteString("\n\n")
	sb.WriteString("Make sure the Instance ID is set under 'Seatsurfing Configuration' in your Confluence installation.")
	sb.WriteString("\n\n")
	sb.WriteString("For more information, please read the documentation at: https://docs.seatsurfing.io/")
	sb.WriteString("\n")
	return []byte(sb.String())
}

func (router *ConfluenceRouter) getUserEmailServer(org *Organization, claims *ConfluenceServerClaims, allowAnonymous bool) string {
	userAccountID := ""
	desiredDomain := ""
	if claims.UserName != "" {
		mailparts := strings.Split(claims.UserName, "@")
		if len(mailparts) == 2 {
			userAccountID = mailparts[0]
			desiredDomain = mailparts[1]
		}
	}
	if userAccountID == "" {
		if claims.UserName != "" {
			userAccountID = "confluence-" + claims.UserName
		}
		if claims.UserName == "" {
			if !allowAnonymous {
				return ""
			}
			userAccountID = "confluence-anonymous-" + uuid.New().String()
		}
	}
	domains, err := GetOrganizationRepository().GetDomains(org)
	if err != nil {
		return ""
	}
	domain := ""
	otherDomain := ""
	for _, curDomain := range domains {
		if curDomain.Active {
			otherDomain = curDomain.DomainName
			if desiredDomain != "" && desiredDomain == curDomain.DomainName {
				domain = curDomain.DomainName
			}
		}
	}
	if domain == "" {
		domain = otherDomain
	}
	return userAccountID + "@" + domain
}
