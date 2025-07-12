package test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/seatsurfing/seatsurfing/server/router"
	"fmt"
	"encoding/json"
)

func TestHealthcheckHandler_Success(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	router.HealthcheckHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestHealthcheckHandler_DBError(t *testing.T) {
	original := router.CheckDatabase
	router.CheckDatabase = func() error {
		return fmt.Errorf("simulated DB error")
	}
	defer func() { router.CheckDatabase = original }()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	router.HealthcheckHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", resp.StatusCode)
	}

	var body struct {
		Status string   `json:"status"`
		Code   int      `json:"code"`
		Errors []string `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Errorf("invalid JSON response: %v", err)
	}
	if body.Status != "error" || body.Code != 500 || len(body.Errors) == 0 {
		t.Errorf("unexpected body: %+v", body)
	}
}
