package pluginapi

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// ─── Mail template types ─────────────────────────────────────────────────────

type MailButton struct {
	Paragraph string `json:"paragraph"`
	URL       string `json:"url"`
	Label     string `json:"label"`
}

type FinalInfo struct {
	Text  string `json:"text"`
	Label string `json:"label"`
	URL   string `json:"url"`
}

type MailTemplate struct {
	Subject         string       `json:"subject"`
	Headline        string       `json:"headline"`
	Paragraphs      []string     `json:"paragraphs"`
	Buttons         []MailButton `json:"buttons"`
	FinalParagraphs []string     `json:"finalParagraphs"`
	FinalInfo       *FinalInfo   `json:"finalInfo"`
}

// GetHTMLMailTemplate parses jsonTemplate into a MailTemplate and renders the
// HTML body using the email.html wrapper from FILESYSTEM_BASE_PATH/res/email.html.
func GetHTMLMailTemplate(jsonTemplate []byte) (*MailTemplate, string, error) {
	basePath := os.Getenv("FILESYSTEM_BASE_PATH")
	path := filepath.Join(basePath, "./res/email.html")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("error reading email template: %v", err)
	}
	s := string(data)
	var t MailTemplate
	if err := json.Unmarshal(jsonTemplate, &t); err != nil {
		return nil, "", fmt.Errorf("error parsing mail template JSON: %v", err)
	}
	s = strings.ReplaceAll(s, "{{headline}}", html.EscapeString(t.Headline))
	body := ""
	for _, p := range t.Paragraphs {
		body += "<p>" + html.EscapeString(p) + "</p>"
	}
	for _, btn := range t.Buttons {
		body += "<p>" + html.EscapeString(btn.Paragraph) + "</p>"
		body += "<p><a href=\"" + btn.URL + "\" class=\"btn btn-primary\">" + html.EscapeString(btn.Label) + "</a></p>"
	}
	for _, p := range t.FinalParagraphs {
		body += "<p>" + html.EscapeString(p) + "</p>"
	}
	if t.FinalInfo != nil {
		anchor := "<a href=\"" + t.FinalInfo.URL + "\">" + html.EscapeString(t.FinalInfo.Label) + "</a>"
		text := strings.ReplaceAll(html.EscapeString(t.FinalInfo.Text), "{{link}}", anchor)
		body += "<p class=\"small\">" + text + "</p>"
	}
	s = strings.ReplaceAll(s, "{{body}}", body)
	return &t, s, nil
}

// ReplaceVarsInTemplate substitutes {{key}} placeholders and {{if key}}…{{end}}
// conditionals in body using the provided vars map.
func ReplaceVarsInTemplate(body string, vars map[string]string) string {
	for key, val := range vars {
		rx := regexp.MustCompile(`{{if ` + key + `}}(.*?){{end}}`)
		if val == "1" {
			body = rx.ReplaceAllString(body, "$1")
		} else {
			body = rx.ReplaceAllString(body, "")
		}
	}
	for key, val := range vars {
		rx := regexp.MustCompile(`{{if \!` + key + `}}(.*?){{end}}`)
		if val != "1" {
			body = rx.ReplaceAllString(body, "$1")
		} else {
			body = rx.ReplaceAllString(body, "")
		}
	}
	for key, val := range vars {
		body = strings.ReplaceAll(body, "{{"+key+"}}", html.EscapeString(val))
	}
	return body
}

// ─── Validation helpers ──────────────────────────────────────────────────────

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
var guidRegex  = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func ValidateEmail(s string) bool {
	if len([]rune(s)) > 254 {
		return false
	}
	return emailRegex.MatchString(s)
}

func ValidatePassword(s string) bool {
	l := len([]rune(s))
	if l < 8 || l > 64 {
		return false
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range s {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case !unicode.IsLetter(r) && !unicode.IsDigit(r):
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasDigit && hasSpecial
}

func ValidateGUID(s string) bool {
	return guidRegex.MatchString(s)
}
