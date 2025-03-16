package util

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type ACSSendMailContent struct {
	Subject   string `json:"subject"`
	Plaintext string `json:"plainText"`
}

type ACSRecipients struct {
	To []ACSAddress `json:"to"`
}

type ACSAddress struct {
	Address     string `json:"address"`
	DisplayName string `json:"displayName"`
}

type ACSSendMailRequest struct {
	SenderAddress string             `json:"senderAddress"`
	Recipients    ACSRecipients      `json:"recipients"`
	Content       ACSSendMailContent `json:"content"`
	ReplyTo       []ACSAddress       `json:"replyTo"`
}

func ACSSendEmail(host string, accessKey string, r *ACSSendMailRequest) error {
	url, err := url.Parse("https://" + host + "/emails:send?api-version=2023-03-31")
	if err != nil {
		return err
	}
	payload, err := json.Marshal(r)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(http.TimeFormat)
	contentHash := acsContentHash(payload)
	signature, err := GetACSSignature(accessKey, url, contentHash, now)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-ms-date", now)
	req.Header.Add("x-ms-content-sha256", contentHash)
	req.Header.Add("Authorization", "HMAC-SHA256 SignedHeaders=x-ms-date;host;x-ms-content-sha256&Signature="+signature)
	client := &http.Client{
		Timeout: time.Second * 15,
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("could not send mail via ACS, status code = %d, error: %s", res.StatusCode, string(body))
	}
	return nil
}

func GetACSSignature(accessKey string, requestURI *url.URL, contentHash string, date string) (string, error) {
	uriPathAndQuery := requestURI.Path + "?" + requestURI.RawQuery
	host := requestURI.Host
	stringToSign := "POST\n" + uriPathAndQuery + "\n" + date + ";" + host + ";" + contentHash
	return acsComputeSignature(accessKey, stringToSign)
}

func acsComputeSignature(accessKey string, data string) (string, error) {
	accessKeyDecoded, err := base64.StdEncoding.DecodeString(accessKey)
	if err != nil {
		return "", err
	}
	hmac := hmac.New(sha256.New, accessKeyDecoded)
	hmac.Write([]byte(data))
	dataHmac := hmac.Sum(nil)
	res := base64.StdEncoding.EncodeToString(dataHmac)
	return res, nil
}

func acsContentHash(data []byte) string {
	h := sha256.New()
	h.Write(data)
	bs := h.Sum(nil)
	res := base64.StdEncoding.EncodeToString(bs)
	return res
}
