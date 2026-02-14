package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/seatsurfing/seatsurfing/server/config"
	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/util"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

type contextKey string

func (c contextKey) String() string {
	return "seatsurfing context key " + string(c)
}

var (
	contextKeyUserID    = contextKey("UserID")
	contextKeySessionID = contextKey("SessionID")
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
	ResponseCodeBookingNotAllowedBooker          = 1009
	ResponseCodeBookingSubjectRequired           = 1010
	ResponseCodeBookingInPast                    = 1011

	ResponseCodePresenceReportDateRangeTooLong = 2001
)

func sendErrorCode(w http.ResponseWriter, statusCode int, code int) {
	w.Header().Set("X-Error-Code", strconv.Itoa(code))
	w.WriteHeader(statusCode)
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
	sendErrorCode(w, http.StatusForbidden, code)
}

func SendBadRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
}

func SendBadRequestCode(w http.ResponseWriter, code int) {
	sendErrorCode(w, http.StatusBadRequest, code)
}

func SendPaymentRequired(w http.ResponseWriter) {
	w.WriteHeader(http.StatusPaymentRequired)
}

func SendUnauthorized(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
}

func SendTooManyRequests(w http.ResponseWriter) {
	w.WriteHeader(http.StatusTooManyRequests)
}

func SendAlreadyExists(w http.ResponseWriter) {
	w.WriteHeader(http.StatusConflict)
}

func SendAlreadyExistsCode(w http.ResponseWriter, code int) {
	sendErrorCode(w, http.StatusConflict, code)
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

func SecurityHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SetCorsHeaders(w, r)
		SetSecurityHeaders(w, r)
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
	installID, _ := GetSettingsRepository().GetGlobalString(SettingInstallID.Name)
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"RS512"}),
		jwt.WithIssuer(installID),
		jwt.WithAudience(installID),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	claims := &Claims{}
	token, err := parser.ParseWithClaims(authHeader, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return config.GetConfig().JwtPublicKey, nil
	})
	if err != nil {
		return nil, "", errors.New("JWT header verification failed: parsing JWT failed with: " + err.Error())
	}
	if !token.Valid {
		return nil, "", errors.New("JWT header verification failed: invalid JWT")
	}
	return claims, authHeader, nil
}

func GetRateLimiterMiddleware() mux.MiddlewareFunc {
	const anonymousUserID = "anonymous"
	periodSplit := strings.Split(config.GetConfig().RateLimitPeriod, "-")
	periodValue, _ := strconv.Atoi(periodSplit[0])
	var period time.Duration
	switch periodSplit[1] {
	case "S":
		period = time.Duration(periodValue) * time.Second
	case "M":
		period = time.Duration(periodValue) * time.Minute
	case "H":
		period = time.Duration(periodValue) * time.Hour
	case "D":
		period = time.Duration(periodValue*24) * time.Hour
	default:
		period = time.Duration(1) * time.Hour
	}
	rate := limiter.Rate{
		Period: period,
		Limit:  int64(config.GetConfig().RateLimit),
	}
	store := memory.NewStore()
	instance := limiter.New(store, rate, limiter.WithTrustForwardHeader(true))
	limiterMiddleware := &stdlib.Middleware{
		Limiter:        instance,
		OnError:        stdlib.DefaultErrorHandler,
		OnLimitReached: stdlib.DefaultLimitReachedHandler,
		KeyGetter: func(r *http.Request) string {
			userID := GetRequestUserID(r)
			if userID == "" {
				userID = anonymousUserID
			}
			return userID
		},
		ExcludedKey: func(key string) bool {
			if key == anonymousUserID {
				return true
			}
			return false
		},
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			url := r.URL.Path
			if strings.HasPrefix(url, "/ui/") {
				next.ServeHTTP(w, r)
				return
			}
			limiterHandler := limiterMiddleware.Handler(next)
			limiterHandler.ServeHTTP(w, r)
		})
	}
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
		url := r.URL.Path
		if url == "/" {
			return true
		}
		// Check for whitelisted public API paths
		for _, whitelistedURL := range getUnauthorizedRoutes() {
			if isWhitelistMatch(url, whitelistedURL) {
				return true
			}
		}
		return false
	}

	var handleServiceAccountAuth = func(w http.ResponseWriter, r *http.Request) bool {
		username, password, ok := r.BasicAuth()
		if !ok {
			return false
		}
		if len(username) < 36+2 && strings.Index(username, "_") != 36 {
			return false
		}
		organizationId := username[:36]
		email := username[37:]
		user, err := GetUserRepository().GetByEmail(organizationId, email)
		if err != nil || user == nil {
			return false
		}
		if user.Role != UserRoleServiceAccountRO && user.Role != UserRoleServiceAccountRW {
			return false
		}
		if user.HashedPassword == "" {
			return false
		}
		if user.Disabled {
			return false
		}
		if !GetUserRepository().CheckPassword(string(user.HashedPassword), password) {
			return false
		}
		if r.Method != "GET" && user.Role == UserRoleServiceAccountRO {
			return false
		}
		ctx := context.WithValue(r.Context(), contextKeyUserID, user.ID)
		// Note: service accounts do not have sessions, so we do not set session ID in context
		next.ServeHTTP(w, r.WithContext(ctx))
		return true
	}

	var handleTokenAuth = func(w http.ResponseWriter, r *http.Request) bool {
		claims, _, err := ExtractClaimsFromRequest(r)
		if err != nil {
			return false
		}
		session, err := GetSessionRepository().GetOne(claims.SessionID)
		if err != nil || session == nil {
			return false
		}
		if session.UserID != claims.UserID {
			return false
		}
		user, err := GetUserRepository().GetOne(claims.UserID)
		if err != nil || user == nil {
			return false
		}
		if user.Disabled {
			return false
		}
		// Update session activity on each authenticated request
		go GetSessionRepository().UpdateActivity(session.ID)
		ctx := context.WithValue(r.Context(), contextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, contextKeySessionID, session.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
		return true
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}
		if IsWhitelisted(r) {
			// even for whitelisted routes, we check if there is an auth header and if it is valid, so that we can have the user context available in whitelisted routes if the client provides a valid token
			processedWithAuth := false
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				processedWithAuth = handleTokenAuth(w, r) || handleServiceAccountAuth(w, r)
			}
			if !processedWithAuth {
				next.ServeHTTP(w, r)
			}
			return
		}
		success := handleTokenAuth(w, r) || handleServiceAccountAuth(w, r)
		if !success {
			SendUnauthorized(w)
			return
		}
	})
}

func SetCorsHeaders(w http.ResponseWriter, r *http.Request) {
	allowedOrigins := config.GetConfig().CORSOrigins
	origin := r.Header.Get("Origin")
	setOrigin := ""
	if slices.Contains(allowedOrigins, origin) {
		setOrigin = origin
	} else if slices.Contains(allowedOrigins, "*") {
		setOrigin = "*"
	}
	if setOrigin != "" {
		w.Header().Set("Access-Control-Allow-Origin", setOrigin)
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Expose-Headers", "X-Object-Id, X-Error-Code, Content-Length, Content-Type")
	}
}

func SetSecurityHeaders(w http.ResponseWriter, r *http.Request) {
	if strings.ToLower(config.GetConfig().PublicScheme) == "https" {
		w.Header().Set("Content-Security-Policy", "upgrade-insecure-requests")
	}
	w.Header().Set("Permissions-Policy", "accelerometer=(), ambient-light-sensor=(), autoplay=(), battery=(), camera=(), cross-origin-isolated=(), display-capture=(), document-domain=(), encrypted-media=(), execution-while-not-rendered=(), execution-while-out-of-viewport=(), fullscreen=(), geolocation=(), gyroscope=(), keyboard-map=(), magnetometer=(), microphone=(), midi=(), navigation-override=(), payment=(), picture-in-picture=(), publickey-credentials-get=(), screen-wake-lock=(), sync-xhr=(), usb=(), web-share=(), xr-spatial-tracking=()")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
}

func CorsHandler(w http.ResponseWriter, r *http.Request) {
	SetCorsHeaders(w, r)
	w.WriteHeader(http.StatusNoContent)
}

func GetRequestSessionID(r *http.Request) string {
	sessionID := r.Context().Value(contextKeySessionID)
	if sessionID == nil {
		return ""
	}
	return sessionID.(string)
}

func GetRequestUserID(r *http.Request) string {
	userID := r.Context().Value(contextKeyUserID)
	if userID == nil {
		return ""
	}
	return userID.(string)
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
	if GetUserRepository().IsSuperAdmin(user) {
		return true
	}
	return false
}

func CanSpaceAdminOrg(user *User, organizationID string) bool {
	if (user.OrganizationID == organizationID) && (GetUserRepository().IsSpaceAdmin(user)) {
		return true
	}
	if GetUserRepository().IsSuperAdmin(user) {
		return true
	}
	return false
}

func CanAdminOrg(user *User, organizationID string) bool {
	if (user.OrganizationID == organizationID) && (GetUserRepository().IsOrgAdmin(user)) {
		return true
	}
	if GetUserRepository().IsSuperAdmin(user) {
		return true
	}
	return false
}

func GetValidator() *validator.Validate {
	v := validator.New()
	v.RegisterValidation("jsDate", func(fl validator.FieldLevel) bool {
		_, err := ParseJSDate(fl.Field().String())
		return err == nil
	})
	return v
}
