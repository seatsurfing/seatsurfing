package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2/clientcredentials"

	. "github.com/seatsurfing/seatsurfing/server/repository"
)

const (
	ExchangeOpCreate = "CREATE"
	ExchangeOpUpdate = "UPDATE"
	ExchangeOpDelete = "DELETE"

	exchangeMaxRetries = 5
	exchangeBatchSize  = 20
	graphBaseURL       = "https://graph.microsoft.com/v1.0"
)

// ExchangeSyncPayload is stored as JSON in exchange_sync_queue.payload.
type ExchangeSyncPayload struct {
	OrgID           string    `json:"orgID"`
	BookingID       string    `json:"bookingID"`
	Operation       string    `json:"operation"`
	RoomEmail       string    `json:"roomEmail"`
	ExchangeEventID string    `json:"exchangeEventID"`
	Enter           time.Time `json:"enter"`
	Leave           time.Time `json:"leave"`
	LocationTZ      string    `json:"locationTZ"`
	UserFirstname   string    `json:"userFirstname"`
	UserLastname    string    `json:"userLastname"`
	SpaceName       string    `json:"spaceName"`
	LocationName    string    `json:"locationName"`
}

// ExchangeSyncWorker processes pending Exchange sync jobs.
type ExchangeSyncWorker struct {
	mu      sync.Mutex
	running bool

	// token cache: orgID -> *cachedToken
	tokenCache   map[string]*cachedToken
	tokenCacheMu sync.Mutex
}

type cachedToken struct {
	accessToken string
	expiresAt   time.Time
}

var exchangeSyncWorker *ExchangeSyncWorker
var exchangeSyncWorkerOnce sync.Once

func GetExchangeSyncWorker() *ExchangeSyncWorker {
	exchangeSyncWorkerOnce.Do(func() {
		exchangeSyncWorker = &ExchangeSyncWorker{
			tokenCache: make(map[string]*cachedToken),
		}
	})
	return exchangeSyncWorker
}

func (w *ExchangeSyncWorker) ProcessPendingJobs() {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}
	w.running = true
	w.mu.Unlock()
	defer func() {
		w.mu.Lock()
		w.running = false
		w.mu.Unlock()
	}()

	jobs, err := GetExchangeSyncQueueRepository().ClaimBatch(exchangeBatchSize)
	if err != nil {
		log.Println("ExchangeSyncWorker: error claiming batch:", err)
		return
	}
	for _, job := range jobs {
		w.processJob(job)
	}
}

func (w *ExchangeSyncWorker) processJob(job *ExchangeSyncQueueItem) {
	var payload ExchangeSyncPayload
	if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
		log.Printf("ExchangeSyncWorker: failed to parse payload for job %s: %v", job.ID, err)
		GetExchangeSyncQueueRepository().MarkFailed(job.ID, "payload parse error: "+err.Error())
		return
	}

	settings, err := GetSettingsRepository().GetExchangeSettings(payload.OrgID)
	if err != nil || !settings.Enabled {
		GetExchangeSyncQueueRepository().MarkFailed(job.ID, "exchange not configured or disabled for org")
		return
	}

	token, err := w.getToken(payload.OrgID, settings)
	if err != nil {
		log.Printf("ExchangeSyncWorker: token fetch failed for job %s: %v", job.ID, err)
		GetExchangeSyncQueueRepository().MarkFailed(job.ID, "token fetch failed: "+err.Error())
		return
	}

	var execErr error
	switch payload.Operation {
	case ExchangeOpCreate:
		execErr = w.executeCreate(job, &payload, token)
	case ExchangeOpUpdate:
		execErr = w.executeUpdate(job, &payload, token)
	case ExchangeOpDelete:
		execErr = w.executeDelete(job, &payload, token)
	default:
		GetExchangeSyncQueueRepository().MarkFailed(job.ID, "unknown operation: "+payload.Operation)
		return
	}

	if execErr != nil {
		w.handleJobError(job, execErr.Error())
	}
}

func (w *ExchangeSyncWorker) executeCreate(job *ExchangeSyncQueueItem, payload *ExchangeSyncPayload, token string) error {
	body, err := w.buildEventBody(payload)
	if err != nil {
		GetExchangeSyncQueueRepository().MarkFailed(job.ID, "build event body failed: "+err.Error())
		return nil
	}
	url := fmt.Sprintf("%s/users/%s/calendar/events", graphBaseURL, payload.RoomEmail)
	resp, respBody, err := w.doGraphRequest(http.MethodPost, url, token, body)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusCreated {
		var result struct {
			ID string `json:"id"`
		}
		if jsonErr := json.Unmarshal(respBody, &result); jsonErr != nil || result.ID == "" {
			GetExchangeSyncQueueRepository().MarkFailed(job.ID, "failed to parse event id from create response")
			return nil
		}
		if dbErr := GetExchangeBookingMappingRepository().Create(payload.BookingID, result.ID, payload.RoomEmail); dbErr != nil {
			log.Printf("ExchangeSyncWorker: failed to store booking mapping for job %s: %v", job.ID, dbErr)
		}
		GetExchangeSyncQueueRepository().Delete(job.ID)
		return nil
	}
	return w.httpStatusError(resp.StatusCode, respBody)
}

func (w *ExchangeSyncWorker) executeUpdate(job *ExchangeSyncQueueItem, payload *ExchangeSyncPayload, token string) error {
	// Check if we have the exchange event ID
	eventID := payload.ExchangeEventID
	if eventID == "" {
		// Try to look it up from the mapping table
		mapping, err := GetExchangeBookingMappingRepository().GetByBookingID(payload.BookingID)
		if err != nil || mapping == nil {
			// CREATE may have never succeeded; skip UPDATE
			GetExchangeSyncQueueRepository().Delete(job.ID)
			return nil
		}
		eventID = mapping.ExchangeEventID
	}

	body, err := w.buildEventBody(payload)
	if err != nil {
		GetExchangeSyncQueueRepository().MarkFailed(job.ID, "build event body failed: "+err.Error())
		return nil
	}
	url := fmt.Sprintf("%s/users/%s/calendar/events/%s", graphBaseURL, payload.RoomEmail, eventID)
	resp, respBody, err := w.doGraphRequest(http.MethodPatch, url, token, body)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		GetExchangeSyncQueueRepository().Delete(job.ID)
		return nil
	}
	return w.httpStatusError(resp.StatusCode, respBody)
}

func (w *ExchangeSyncWorker) executeDelete(job *ExchangeSyncQueueItem, payload *ExchangeSyncPayload, token string) error {
	eventID := payload.ExchangeEventID
	if eventID == "" {
		// CREATE never succeeded; nothing to delete
		GetExchangeSyncQueueRepository().Delete(job.ID)
		return nil
	}

	url := fmt.Sprintf("%s/users/%s/calendar/events/%s", graphBaseURL, payload.RoomEmail, eventID)
	resp, respBody, err := w.doGraphRequest(http.MethodDelete, url, token, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusNotFound {
		GetExchangeBookingMappingRepository().Delete(payload.BookingID)
		GetExchangeSyncQueueRepository().Delete(job.ID)
		return nil
	}
	return w.httpStatusError(resp.StatusCode, respBody)
}

func (w *ExchangeSyncWorker) buildEventBody(payload *ExchangeSyncPayload) ([]byte, error) {
	tz := payload.LocationTZ
	if tz == "" {
		tz = "UTC"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		tz = "UTC"
		loc = time.UTC
	}
	localEnter := payload.Enter.In(loc)
	localLeave := payload.Leave.In(loc)

	displayName := payload.UserFirstname + " " + payload.UserLastname
	startStr := localEnter.Format("2006-01-02T15:04:05")
	endStr := localLeave.Format("2006-01-02T15:04:05")
	timeRangeStr := localEnter.Format("15:04") + "–" + localLeave.Format("15:04") + " (" + tz + ")"

	event := map[string]interface{}{
		"subject": "[Seatsurfing] Booking: " + displayName,
		"body": map[string]string{
			"contentType": "HTML",
			"content": fmt.Sprintf("<p>Booked by: <b>%s</b><br>Space: %s, %s<br>Time: %s</p>",
				displayName, payload.SpaceName, payload.LocationName, timeRangeStr),
		},
		"start": map[string]string{
			"dateTime": startStr,
			"timeZone": tz,
		},
		"end": map[string]string{
			"dateTime": endStr,
			"timeZone": tz,
		},
		"showAs":            "busy",
		"isReminderOn":      false,
		"responseRequested": false,
	}
	return json.Marshal(event)
}

func (w *ExchangeSyncWorker) doGraphRequest(method, url, token string, body []byte) (*http.Response, []byte, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(context.Background(), method, url, bodyReader)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return resp, respBody, nil
}

type graphHTTPError struct {
	statusCode int
	retriable  bool
	message    string
}

func (e *graphHTTPError) Error() string {
	return e.message
}

func (w *ExchangeSyncWorker) httpStatusError(statusCode int, body []byte) error {
	retriable := statusCode == http.StatusTooManyRequests || (statusCode >= 500 && statusCode <= 599)
	return &graphHTTPError{
		statusCode: statusCode,
		retriable:  retriable,
		message:    fmt.Sprintf("HTTP %d: %s", statusCode, string(body)),
	}
}

func (w *ExchangeSyncWorker) handleJobError(job *ExchangeSyncQueueItem, errMsg string) {
	newRetryCount := job.RetryCount + 1
	if newRetryCount >= exchangeMaxRetries {
		GetExchangeSyncQueueRepository().MarkFailed(job.ID, errMsg)
		return
	}
	GetExchangeSyncQueueRepository().ScheduleRetry(job.ID, newRetryCount, errMsg)
}

func (w *ExchangeSyncWorker) getToken(orgID string, settings *ExchangeSettings) (string, error) {
	w.tokenCacheMu.Lock()
	defer w.tokenCacheMu.Unlock()

	cached, ok := w.tokenCache[orgID]
	if ok && time.Now().Before(cached.expiresAt) {
		return cached.accessToken, nil
	}

	cfg := clientcredentials.Config{
		ClientID:     settings.ClientID,
		ClientSecret: settings.ClientSecret,
		TokenURL:     fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", settings.TenantID),
		Scopes:       []string{"https://graph.microsoft.com/.default"},
	}
	token, err := cfg.Token(context.Background())
	if err != nil {
		return "", err
	}
	w.tokenCache[orgID] = &cachedToken{
		accessToken: token.AccessToken,
		expiresAt:   token.Expiry.Add(-60 * time.Second),
	}
	return token.AccessToken, nil
}

// InvalidateTokenCache removes the cached token for an org (e.g., after credential update).
func (w *ExchangeSyncWorker) InvalidateTokenCache(orgID string) {
	w.tokenCacheMu.Lock()
	defer w.tokenCacheMu.Unlock()
	delete(w.tokenCache, orgID)
}

// TestConnection verifies the Exchange credentials by acquiring a token.
func (w *ExchangeSyncWorker) TestConnection(settings *ExchangeSettings) error {
	cfg := clientcredentials.Config{
		ClientID:     settings.ClientID,
		ClientSecret: settings.ClientSecret,
		TokenURL:     fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", settings.TenantID),
		Scopes:       []string{"https://graph.microsoft.com/.default"},
	}
	_, err := cfg.Token(context.Background())
	return err
}
