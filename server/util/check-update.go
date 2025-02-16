package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type CheckVersionRequest struct {
	InstallID      string `json:"id" validate:"required,uuid"`
	CurrentVersion string `json:"version" validate:"required,semver"`
}

type CheckVersionResponse struct {
	UpdateAvailable bool   `json:"updateAvailable"`
	LatestVersion   string `json:"version"`
}

type UpdateChecker struct {
	Latest *CheckVersionResponse
}

var updateChecker *UpdateChecker
var updateCheckerOnce sync.Once

func GetUpdateChecker() *UpdateChecker {
	updateCheckerOnce.Do(func() {
		updateChecker = &UpdateChecker{}
	})
	return updateChecker
}

func (uc *UpdateChecker) pollLatestRelease(installID string) (*CheckVersionResponse, error) {
	const url = "https://uc.seatsurfing.io/"
	//installID, _ := GetSettingsRepository().GetGlobalString(SettingInstallID.Name)
	payload := CheckVersionRequest{
		InstallID:      installID,
		CurrentVersion: GetProductVersion(),
	}
	req, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(req))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Received status code " + strconv.Itoa(resp.StatusCode))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var details CheckVersionResponse
	if err := json.Unmarshal(body, &details); err != nil {
		return nil, err
	}
	return &details, nil
}

func (uc *UpdateChecker) updateLatestReleaseDetails(installID string) error {
	details, err := uc.pollLatestRelease(installID)
	if err != nil {
		return err
	}
	if (uc.Latest == nil && details.UpdateAvailable) || (uc.Latest != nil && uc.Latest.UpdateAvailable != details.UpdateAvailable) {
		log.Printf("Update available: %s\n", details.LatestVersion)
	}
	uc.Latest = details
	return nil
}

func (uc *UpdateChecker) onVersionUpdateTimerTick(installID string) error {
	if err := uc.updateLatestReleaseDetails(installID); err != nil {
		log.Printf("Could not update latest version: %s\n", err.Error())
		return err
	}
	return nil
}

func (uc *UpdateChecker) InitializeVersionUpdateTimer(installID string) {
	if err := uc.onVersionUpdateTimerTick(installID); err != nil {
		return
	}
	ticker := time.NewTicker(time.Minute * 60)
	go func() {
		for {
			<-ticker.C
			uc.onVersionUpdateTimerTick(installID)
		}
	}()
}
