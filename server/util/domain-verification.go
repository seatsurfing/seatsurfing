package util

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type DomainAccessibilityPayload struct {
	Domain string `json:"domain"`
	OrgID  string `json:"orgID"`
	Status string `json:"status"`
}

func IsValidTXTRecord(domain, uuid string) bool {
	records, err := net.LookupTXT(domain)
	if err != nil {
		log.Println(err)
		return false
	}
	checkString := "seatsurfing-verification=" + uuid
	for _, record := range records {
		if record == checkString {
			return true
		}
	}
	return false
}

func IsDomainAccessible(domain, orgID string) (bool, error) {
	if err := isDomainAccessible("https", domain, 443, orgID); err == nil {
		return true, nil
	}
	if err := isDomainAccessible("http", domain, 80, orgID); err != nil {
		return false, err
	}
	return true, nil
}

func isDomainAccessible(scheme, domain string, port int, orgID string) error {
	u := new(url.URL)
	u.Scheme = scheme
	u.Host = domain + ":" + fmt.Sprint(port)
	u.Path = "/organization/domain/verify/" + url.PathEscape(domain)
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("could not verify domain accessibility, status code = %d, error: %s", res.StatusCode, string(body))
	}
	var payload DomainAccessibilityPayload
	json.Unmarshal(body, &payload)
	if strings.ToUpper(payload.Status) != "OK" {
		return fmt.Errorf("domain verification failed, status = %s", payload.Status)
	}
	if payload.Domain != domain {
		return fmt.Errorf("domain verification failed, expected domain %s, but got %s", domain, payload.Domain)
	}
	if payload.OrgID != orgID {
		return fmt.Errorf("domain verification failed, expected orgID %s, but got %s", orgID, payload.OrgID)
	}
	return nil
}
