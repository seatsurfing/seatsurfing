package router

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	. "github.com/seatsurfing/seatsurfing/server/repository"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

type ExchangeRouter struct {
}

type GetExchangeOrgSettingsResponse struct {
	Enabled      bool   `json:"enabled"`
	TenantID     string `json:"tenantId"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"` // always empty on GET
}

type SetExchangeOrgSettingsRequest struct {
	Enabled      bool   `json:"enabled"`
	TenantID     string `json:"tenantId"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"` // omitted means "keep existing"
}

type GetExchangeSyncErrorResponse struct {
	ID        string `json:"id"`
	BookingID string `json:"bookingId"`
	Operation string `json:"operation"`
	LastError string `json:"lastError"`
	CreatedAt string `json:"createdAt"`
}

func (router *ExchangeRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("/test", router.testConnection).Methods("POST")
	s.HandleFunc("/errors/{id}/retry", router.retryError).Methods("POST")
	s.HandleFunc("/errors/", router.getErrors).Methods("GET")
	s.HandleFunc("/", router.getSettings).Methods("GET")
	s.HandleFunc("/", router.setSettings).Methods("PUT")
}

func (router *ExchangeRouter) getSettings(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CanAdminOrg(user, user.OrganizationID) {
		SendForbidden(w)
		return
	}
	if enabled, _ := GetSettingsRepository().GetBool(user.OrganizationID, SettingFeatureExchangeIntegration.Name); !enabled {
		SendForbidden(w)
		return
	}
	settings, err := GetSettingsRepository().GetExchangeSettings(user.OrganizationID)
	if err != nil {
		// No settings yet, return defaults
		SendJSON(w, &GetExchangeOrgSettingsResponse{})
		return
	}
	res := &GetExchangeOrgSettingsResponse{
		Enabled:      settings.Enabled,
		TenantID:     settings.TenantID,
		ClientID:     settings.ClientID,
		ClientSecret: "", // never expose
	}
	SendJSON(w, res)
}

func (router *ExchangeRouter) setSettings(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CanAdminOrg(user, user.OrganizationID) {
		SendForbidden(w)
		return
	}
	if enabled, _ := GetSettingsRepository().GetBool(user.OrganizationID, SettingFeatureExchangeIntegration.Name); !enabled {
		SendForbidden(w)
		return
	}
	var m SetExchangeOrgSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		SendBadRequest(w)
		return
	}

	// Load existing to preserve secret if not provided
	existingSecret := ""
	if m.ClientSecret == "" {
		if existing, err := GetSettingsRepository().GetExchangeSettings(user.OrganizationID); err == nil {
			existingSecret = existing.ClientSecret
		}
	}

	settings := &ExchangeSettings{
		Enabled:      m.Enabled,
		TenantID:     m.TenantID,
		ClientID:     m.ClientID,
		ClientSecret: m.ClientSecret,
	}
	if settings.ClientSecret == "" {
		settings.ClientSecret = existingSecret
	}
	if err := GetSettingsRepository().SetExchangeSettings(user.OrganizationID, settings); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	// Invalidate cached token since credentials may have changed
	GetExchangeSyncWorker().InvalidateTokenCache(user.OrganizationID)
	SendUpdated(w)
}

func (router *ExchangeRouter) testConnection(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CanAdminOrg(user, user.OrganizationID) {
		SendForbidden(w)
		return
	}
	if enabled, _ := GetSettingsRepository().GetBool(user.OrganizationID, SettingFeatureExchangeIntegration.Name); !enabled {
		SendForbidden(w)
		return
	}
	settings, err := GetSettingsRepository().GetExchangeSettings(user.OrganizationID)
	if err != nil || !settings.Enabled {
		SendBadRequest(w)
		return
	}
	if err := GetExchangeSyncWorker().TestConnection(settings); err != nil {
		log.Println("Exchange test connection failed:", err)
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	SendUpdated(w)
}

func (router *ExchangeRouter) getErrors(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CanAdminOrg(user, user.OrganizationID) {
		SendForbidden(w)
		return
	}
	if enabled, _ := GetSettingsRepository().GetBool(user.OrganizationID, SettingFeatureExchangeIntegration.Name); !enabled {
		SendForbidden(w)
		return
	}
	items, err := GetExchangeSyncQueueRepository().GetFailed(user.OrganizationID)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	res := make([]*GetExchangeSyncErrorResponse, 0, len(items))
	for _, item := range items {
		res = append(res, &GetExchangeSyncErrorResponse{
			ID:        item.ID,
			BookingID: item.BookingID,
			Operation: item.Operation,
			LastError: item.LastError,
			CreatedAt: item.CreatedAt.String(),
		})
	}
	SendJSON(w, res)
}

func (router *ExchangeRouter) retryError(w http.ResponseWriter, r *http.Request) {
	user := GetRequestUser(r)
	if !CanAdminOrg(user, user.OrganizationID) {
		SendForbidden(w)
		return
	}
	if enabled, _ := GetSettingsRepository().GetBool(user.OrganizationID, SettingFeatureExchangeIntegration.Name); !enabled {
		SendForbidden(w)
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]
	if !ValidateGUID(id) {
		SendBadRequest(w)
		return
	}
	if err := GetExchangeSyncQueueRepository().ResetToPending(id); err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	SendUpdated(w)
}
