package util

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/config"
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
	httpPort := 80
	if GetConfig().Development {
		httpPort, _ = strconv.Atoi(GetConfig().PublicListenAddr[strings.Index(GetConfig().PublicListenAddr, ":")+1:])
	}
	if err := isDomainAccessible("http", domain, httpPort, orgID); err != nil {
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
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: tr,
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
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal response body: %v", err)
	}
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
