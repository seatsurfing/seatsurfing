package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

const JsDateTimeFormat string = "2006-01-02T15:04:05"
const JsDateTimeFormatWithTimezone string = "2006-01-02T15:04:05-07:00"

type contextKey string

func (c contextKey) String() string {
	return "flexspace context key " + string(c)
}

var (
	contextKeyUserID     = contextKey("UserID")
	contextKeyAuthHeader = contextKey("AuthHeader")
)

var (
	ResponseCodeBookingSlotConflict              = 1001
	ResponseCodeBookingLocationMaxConcurrent     = 1002
	ResponseCodeBookingTooManyUpcomingBookings   = 1003
	ResponseCodeBookingTooManyDaysInAdvance      = 1004
	ResponseCodeBookingInvalidBookingDuration    = 1005
	ResponseCodeBookingMaxConcurrentForUser      = 1006
	ResponseCodeBookingInvalidMinBookingDuration = 1007
	ResponseCodeBookingMaxHoursBeforeDelete      = 1008
)

type Route interface {
	setupRoutes(s *mux.Router)
}

func SendTemporaryRedirect(w http.ResponseWriter, url string) {
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func SendNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

func SendForbidden(w http.ResponseWriter) {
	w.WriteHeader(http.StatusForbidden)
}

func SendForbiddenCode(w http.ResponseWriter, code int) {
	w.Header().Set("X-Error-Code", strconv.Itoa(code))
	w.WriteHeader(http.StatusForbidden)
}

func SendBadRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
}

func SendBadRequestCode(w http.ResponseWriter, code int) {
	w.Header().Set("X-Error-Code", strconv.Itoa(code))
	w.WriteHeader(http.StatusBadRequest)
}

func SendPaymentRequired(w http.ResponseWriter) {
	w.WriteHeader(http.StatusPaymentRequired)
}

func SendUnauthorized(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
}

func SendAleadyExists(w http.ResponseWriter) {
	w.WriteHeader(http.StatusConflict)
}

func SendCreated(w http.ResponseWriter, id string) {
	w.Header().Set("X-Object-ID", id)
	w.WriteHeader(http.StatusCreated)
}

func SendUpdated(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func SendInternalServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
}

func SendJSON(w http.ResponseWriter, v interface{}) {
	json, err := json.Marshal(v)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func SendTextNotFound(w http.ResponseWriter, contentType string, b []byte) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusNotFound)
	w.Write(b)
}

func UnmarshalBody(r *http.Request, o interface{}) error {
	if r.Body == nil {
		return errors.New("body is NIL")
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(body, &o); err != nil {
		return err
	}
	return nil
}

func UnmarshalValidateBody(r *http.Request, o interface{}) error {
	err := UnmarshalBody(r, &o)
	if err != nil {
		return err
	}
	err = GetValidator().Struct(o)
	if err != nil {
		return err
	}
	return nil
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SetCorsHeaders(w)
		next.ServeHTTP(w, r)
	})
}

func ExtractClaimsFromRequest(r *http.Request) (*Claims, string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, "", errors.New("JWT header verification failed: missing auth header")
	}
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, "", errors.New("JWT header verification failed: invalid auth header")
	}
	authHeader = strings.TrimPrefix(authHeader, "Bearer ")
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(authHeader, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return GetConfig().JwtPublicKey, nil
	})
	if err != nil {
		return nil, "", errors.New("JWT header verification failed: parsing JWT failed with: " + err.Error())
	}
	if !token.Valid {
		return nil, "", errors.New("JWT header verification failed: invalid JWT")
	}
	return claims, authHeader, nil
}

func VerifyAuthMiddleware(next http.Handler) http.Handler {
	var isWhitelistMatch = func(url string, whitelistedURL string) bool {
		whitelistedURL = strings.TrimSpace(whitelistedURL)
		whitelistedURL = strings.TrimSuffix(whitelistedURL, "/")
		if whitelistedURL != "" && (url == whitelistedURL || strings.HasPrefix(url, whitelistedURL+"/")) {
			return true
		}
		return false
	}

	var IsWhitelisted = func(r *http.Request) bool {
		url := r.URL.RequestURI()
		if url == "/" {
			return true
		}
		// Check for whitelisted public API paths
		for _, whitelistedURL := range unauthorizedRoutes {
			if isWhitelistMatch(url, whitelistedURL) {
				return true
			}
		}
		return false
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}
		if IsWhitelisted(r) {
			next.ServeHTTP(w, r)
			return
		}
		claims, authHeader, err := ExtractClaimsFromRequest(r)
		if err != nil {
			log.Println(err)
			SendUnauthorized(w)
			return
		}
		ctx := context.WithValue(r.Context(), contextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, contextKeyAuthHeader, authHeader)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func SetCorsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Expose-Headers", "X-Object-Id, X-Error-Code, Content-Length, Content-Type")
}

func CorsHandler(w http.ResponseWriter, r *http.Request) {
	SetCorsHeaders(w)
	w.WriteHeader(http.StatusNoContent)
}

func GetRequestUserID(r *http.Request) string {
	userID := r.Context().Value(contextKeyUserID)
	if userID == nil {
		return ""
	}
	return userID.(string)
}

func GetAuthHeaderFromContext(r *http.Request) string {
	authHeader := r.Context().Value(contextKeyAuthHeader)
	if authHeader == nil {
		return ""
	}
	return authHeader.(string)
}

func GetRequestUser(r *http.Request) *User {
	ID := GetRequestUserID(r)
	user, err := GetUserRepository().GetOne(ID)
	if err != nil {
		log.Println(err)
		return nil
	}
	return user
}

func CanAccessOrg(user *User, organizationID string) bool {
	if user.OrganizationID == organizationID {
		return true
	}
	if GetUserRepository().isSuperAdmin(user) {
		return true
	}
	return false
}

func CanSpaceAdminOrg(user *User, organizationID string) bool {
	if (user.OrganizationID == organizationID) && (GetUserRepository().isSpaceAdmin(user)) {
		return true
	}
	if GetUserRepository().isSuperAdmin(user) {
		return true
	}
	return false
}

func CanAdminOrg(user *User, organizationID string) bool {
	if (user.OrganizationID == organizationID) && (GetUserRepository().isOrgAdmin(user)) {
		return true
	}
	if GetUserRepository().isSuperAdmin(user) {
		return true
	}
	return false
}

func ParseJSDate(s string) (time.Time, error) {
	return time.Parse(JsDateTimeFormat, s)
}

func ToJSDate(date time.Time) string {
	return date.Format(JsDateTimeFormat)
}

func GetValidator() *validator.Validate {
	v := validator.New()
	v.RegisterValidation("jsDate", func(fl validator.FieldLevel) bool {
		_, err := ParseJSDate(fl.Field().String())
		return err == nil
	})
	return v
}

var unauthorizedRoutes = [...]string{
	"/auth/",
	"/organization/domain/",
	"/auth-provider/org/",
	"/signup/",
	"/admin/",
	"/ui/",
	"/confluence",
	"/booking/debugtimeissues/",
	"/robots.txt",
}
