package router

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/api"
	. "github.com/seatsurfing/seatsurfing/server/repository"
)

type AuthAttemptRouter struct {
}

type GetAuthAttemptResponse struct {
	ID               string    `json:"id"`
	UserID           string    `json:"userId"`
	Email            string    `json:"email"`
	Timestamp        time.Time `json:"timestamp"`
	Successful       bool      `json:"successful"`
	Method           string    `json:"method"`
	AuthProviderID   string    `json:"authProviderId"`
	AuthProviderName string    `json:"authProviderName"`
	ErrorCode        string    `json:"errorCode"`
	ErrorDetail      string    `json:"errorDetail"`
	Device           string    `json:"device"`
}

type GetAuthAttemptListResponse struct {
	Total int                       `json:"total"`
	Items []*GetAuthAttemptResponse `json:"items"`
}

const (
	authAttemptDefaultLimit = 50
	authAttemptMaxLimit     = 100
)

func (router *AuthAttemptRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/", router.getAll).Methods("GET")
}

func (router *AuthAttemptRouter) getAll(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !HasPermission(user, user.OrganizationID, PermissionAuditLog, PermissionLevelRead) {
		SendForbidden(w)
		return
	}
	end := time.Now()
	if param := r.URL.Query().Get("end"); param != "" {
		parsed, err := time.Parse(time.RFC3339Nano, param)
		if err != nil {
			SendBadRequest(w)
			return
		}
		end = parsed
	}
	start := end.Add(-7 * 24 * time.Hour)
	if param := r.URL.Query().Get("start"); param != "" {
		parsed, err := time.Parse(time.RFC3339Nano, param)
		if err != nil {
			SendBadRequest(w)
			return
		}
		start = parsed
	}
	filter := &AuthAttemptFilter{
		OrganizationID: user.OrganizationID,
		Start:          start,
		End:            end,
		EmailLike:      r.URL.Query().Get("user"),
	}
	if param := r.URL.Query().Get("success"); param != "" {
		successful := param == "true" || param == "1"
		filter.Successful = &successful
	}
	limit := authAttemptDefaultLimit
	if param := r.URL.Query().Get("limit"); param != "" {
		parsed, err := strconv.Atoi(param)
		if err != nil || parsed < 1 {
			SendBadRequest(w)
			return
		}
		limit = min(parsed, authAttemptMaxLimit)
	}
	offset := 0
	if param := r.URL.Query().Get("offset"); param != "" {
		parsed, err := strconv.Atoi(param)
		if err != nil || parsed < 0 {
			SendBadRequest(w)
			return
		}
		offset = parsed
	}
	total, err := GetAuthAttemptRepository().CountFiltered(filter)
	if err != nil {
		SendInternalServerError(w)
		return
	}
	list, err := GetAuthAttemptRepository().GetFiltered(filter, limit, offset)
	if err != nil {
		SendInternalServerError(w)
		return
	}
	res := &GetAuthAttemptListResponse{
		Total: total,
		Items: []*GetAuthAttemptResponse{},
	}
	for _, e := range list {
		m := &GetAuthAttemptResponse{
			ID:               e.ID,
			UserID:           e.UserID,
			Email:            e.Email,
			Timestamp:        e.Timestamp,
			Successful:       e.Successful,
			Method:           e.Method,
			AuthProviderID:   e.AuthProviderID,
			AuthProviderName: e.AuthProviderName,
			ErrorCode:        e.ErrorCode,
			ErrorDetail:      e.ErrorDetail,
			Device:           e.Device,
		}
		res.Items = append(res.Items, m)
	}
	SendJSON(w, res)
}
