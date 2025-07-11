package router

import (
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	. "github.com/seatsurfing/seatsurfing/server/repository"
)

type HealthcheckRouter struct{}

func (router *HealthcheckRouter) SetupRoutes(s *mux.Router) {
	s.HandleFunc("", HealthcheckHandler).Methods("GET", "HEAD")
}

var CheckDatabase = func() error {
	db := GetDatabase().DB()
	return db.Ping()
}

func HealthcheckHandler(w http.ResponseWriter, req *http.Request) {
	type HealthcheckResponse struct {
		Status string        `json:"status"`
		Code   int           `json:"code"`
		Errors []string      `json:"errors"`
	}

	err := CheckDatabase()
	if err != nil {
		resp := HealthcheckResponse{
			Status: "error",
			Code:   500,
			Errors: []string{"Database connection error: " + err.Error()},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(resp)
		return
	}
	resp := HealthcheckResponse{
		Status: "success",
		Code:   200,
		Errors: nil,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
