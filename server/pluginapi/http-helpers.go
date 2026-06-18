package pluginapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-playground/validator"
)

type pluginContextKey string

const pluginContextKeyUserID pluginContextKey = "userID"

// SetRequestUserID injects a user ID into the request context so that
// GetRequestUserID can retrieve it from downstream handlers.
func SetRequestUserID(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), pluginContextKeyUserID, userID)
	return r.WithContext(ctx)
}

// GetRequestUserID extracts the user ID previously set by SetRequestUserID.
func GetRequestUserID(r *http.Request) string {
	v := r.Context().Value(pluginContextKeyUserID)
	if v == nil {
		return ""
	}
	return v.(string)
}

// GetValidator returns a new go-playground validator instance.
func GetValidator() *validator.Validate {
	return validator.New()
}

func UnmarshalBody(r *http.Request, o interface{}) error {
	if r.Body == nil {
		return errors.New("body is nil")
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, &o)
}

func UnmarshalValidateBody(r *http.Request, o interface{}) error {
	if err := UnmarshalBody(r, &o); err != nil {
		return err
	}
	return GetValidator().Struct(o)
}

func SendJSON(w http.ResponseWriter, v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		log.Println(err)
		SendInternalServerError(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func SendNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

func SendForbidden(w http.ResponseWriter) {
	w.WriteHeader(http.StatusForbidden)
}

func SendBadRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
}

func sendErrorCode(w http.ResponseWriter, status, code int) {
	w.Header().Set("X-Error-Code", strconv.Itoa(code))
	w.WriteHeader(status)
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

func SendGone(w http.ResponseWriter) {
	w.WriteHeader(http.StatusGone)
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

func SendTemporaryRedirect(w http.ResponseWriter, url string) {
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
