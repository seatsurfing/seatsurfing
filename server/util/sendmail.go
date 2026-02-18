package util

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log"
	"mime/multipart"
	"net/smtp"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	. "github.com/seatsurfing/seatsurfing/server/config"
)

var (
	// ErrInvalidEmailAddress indicates an invalid email address
	ErrInvalidEmailAddress = errors.New("invalid email address")
	// ErrInvalidDisplayName indicates an invalid display name
	ErrInvalidDisplayName = errors.New("invalid display name")
	// ErrInvalidSubject indicates an invalid email subject
	ErrInvalidSubject = errors.New("invalid email subject")
)

// LoginAuth implements the LOGIN authentication mechanism
type LoginAuth struct {
	username, password string
}

func NewLoginAuth(username, password string) smtp.Auth {
	return &LoginAuth{username, password}
}

func (a *LoginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *LoginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch strings.ToLower(string(fromServer)) {
		case "username:":
			return []byte(a.username), nil
		case "password:":
			return []byte(a.password), nil
		default:
			// Try base64 decoded prompts (some servers send base64 encoded prompts)
			decoded, err := base64.StdEncoding.DecodeString(string(fromServer))
			if err == nil {
				switch strings.ToLower(strings.TrimSpace(string(decoded))) {
				case "username:", "user name:":
					return []byte(a.username), nil
				case "password:":
					return []byte(a.password), nil
				}
			}
			return nil, fmt.Errorf("unknown server challenge: %s", string(fromServer))
		}
	}
	return nil, nil
}

// isM365SMTPServer checks if the SMTP host is a Microsoft 365 server
func isM365SMTPServer(host string) bool {
	m365Hosts := []string{
		"smtp.office365.com",
		"smtp-mail.outlook.com",
		"outlook.office365.com",
	}

	host = strings.ToLower(host)
	for _, m365Host := range m365Hosts {
		if host == m365Host {
			return true
		}
	}
	return false
}

// getOptimalSMTPSettings returns optimal settings for the given SMTP host
func getOptimalSMTPSettings(config *Config) (port int, startTLS bool, authMethod string) {
	// Default settings
	port = config.SMTPPort
	startTLS = config.SMTPStartTLS
	authMethod = config.SMTPAuthMethod

	// M365 optimal settings
	if isM365SMTPServer(config.SMTPHost) {
		// Override with M365-specific settings if not explicitly configured
		if config.SMTPPort == 25 { // Default port, likely not configured for M365
			port = 587
		}
		if !config.SMTPStartTLS { // Force STARTTLS for M365
			startTLS = true
		}
		if config.SMTPAuthMethod == "PLAIN" || config.SMTPAuthMethod == "" {
			authMethod = "LOGIN" // M365 often works better with LOGIN
		}
	}

	return port, startTLS, authMethod
}

const EmailTemplateDefaultLanguage = "en"

// EmailLogCallback is a function that logs sent emails
type EmailLogCallback func(subject, recipient, organizationID string) error

var emailLogCallback EmailLogCallback

// SetEmailLogCallback sets the callback function for logging emails
func SetEmailLogCallback(callback EmailLogCallback) {
	emailLogCallback = callback
}

var SendMailMockContent = ""

type MailButton struct {
	Paragraph string `json:"paragraph"`
	URL       string `json:"url"`
	Label     string `json:"label"`
}

type MailTemplate struct {
	Subject         string       `json:"subject"`
	Headline        string       `json:"headline"`
	Paragraphs      []string     `json:"paragraphs"`
	Buttons         []MailButton `json:"buttons"`
	FinalParagraphs []string     `json:"finalParagraphs"`
}

type MailAddress struct {
	Address     string
	DisplayName string
}

type MailAttachment struct {
	Filename  string
	Data      []byte
	MimeType  string
	ContentID string
}

func GetEmailTemplatePathResetpassword() string {
	return filepath.Join(GetConfig().FilesystemBasePath, "./res/email-resetpw.json")
}

func GetEmailTemplatePathInviteUser() string {
	return filepath.Join(GetConfig().FilesystemBasePath, "./res/email-invite-user.json")
}

func GetEmailTemplatePathConfirmDeleteOrg() string {
	return filepath.Join(GetConfig().FilesystemBasePath, "./res/email-confirm-delete-org.json")
}

func GetEmailTemplatePathChangeEmailAddress() string {
	return filepath.Join(GetConfig().FilesystemBasePath, "./res/email-change-email.json")
}

func GetEmailTemplatePathRecurringBookingCreated() string {
	return filepath.Join(GetConfig().FilesystemBasePath, "./res/email-recurring-booking-created.json")
}

func GetEmailTemplatePathBookingCreated() string {
	return filepath.Join(GetConfig().FilesystemBasePath, "./res/email-booking-created.json")
}

func GetEmailTemplatePathBookingUpdated() string {
	return filepath.Join(GetConfig().FilesystemBasePath, "./res/email-booking-updated.json")
}

func GetEmailTemplatePathBookingDeclined() string {
	return filepath.Join(GetConfig().FilesystemBasePath, "./res/email-booking-declined.json")
}

func GetEmailTemplatePathBookingApproved() string {
	return filepath.Join(GetConfig().FilesystemBasePath, "./res/email-booking-approved.json")
}

func GetEmailTemplatePathBookingDeleted() string {
	return filepath.Join(GetConfig().FilesystemBasePath, "./res/email-booking-deleted.json")
}

func GetEmailTemplatePathBookingApprovalRequest() string {
	return filepath.Join(GetConfig().FilesystemBasePath, "./res/email-booking-approval-request.json")
}

func GetEmailTemplatePathFooter() string {
	return filepath.Join(GetConfig().FilesystemBasePath, "./res/email-footer.json")
}

func GetHTMLMailTemplate(jsonTemplate []byte) (*MailTemplate, string, error) {
	path := filepath.Join(GetConfig().FilesystemBasePath, "./res/email.html")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("error reading email template file: %v", err)
	}
	s := string(data)
	var jsonContent MailTemplate
	if err := json.Unmarshal(jsonTemplate, &jsonContent); err != nil {
		return nil, "", fmt.Errorf("error unmarshalling json template: %v", err)
	}
	s = strings.ReplaceAll(s, "{{headline}}", html.EscapeString(jsonContent.Headline))
	body := ""
	for _, paragraph := range jsonContent.Paragraphs {
		body += "<p>" + html.EscapeString(paragraph) + "</p>"
	}
	for _, button := range jsonContent.Buttons {
		body += "<p>" + html.EscapeString(button.Paragraph) + "</p>"
		body += "<p><a href=\"" + button.URL + "\" class=\"btn btn-primary\">" + html.EscapeString(button.Label) + "</a></p>"
	}
	for _, paragraph := range jsonContent.FinalParagraphs {
		body += "<p>" + html.EscapeString(paragraph) + "</p>"
	}
	s = strings.ReplaceAll(s, "{{body}}", body)
	return &jsonContent, s, nil
}

func SendEmail(recipient *MailAddress, templateFile, language string, vars map[string]string) error {
	return SendEmailWithOrg(recipient, templateFile, language, vars, "")
}

func SendEmailWithOrg(recipient *MailAddress, templateFile, language string, vars map[string]string, organizationID string) error {
	return SendEmailWithAttachmentsAndOrg(recipient, templateFile, language, vars, nil, organizationID)
}

func SendEmailWithAttachments(recipient *MailAddress, templateFile, language string, vars map[string]string, attachments []*MailAttachment) error {
	return SendEmailWithAttachmentsAndOrg(recipient, templateFile, language, vars, attachments, "")
}

func SendEmailWithAttachmentsAndOrg(recipient *MailAddress, templateFile, language string, vars map[string]string, attachments []*MailAttachment, organizationID string) error {
	actualTemplateFile, err := GetEmailTemplatePath(templateFile, language)
	if err != nil {
		return err
	}
	actualTemplateData, err := os.ReadFile(actualTemplateFile)
	if err != nil {
		return fmt.Errorf("error reading json template file: %v", err)
	}
	mailTemplate, body, err := GetHTMLMailTemplate(actualTemplateData)
	if err != nil {
		return err
	}
	body = ReplaceVarsInTemplate(body, vars)
	return SendEmailWithBodyAndAttachmentAndOrg(recipient, mailTemplate.Subject, body, language, attachments, organizationID)
}

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

func SendEmailWithBody(recipient *MailAddress, subject, body, language string) error {
	return SendEmailWithBodyAndAttachment(recipient, subject, body, language, nil)
}

func SendEmailWithBodyAndOrg(recipient *MailAddress, subject, body, language string, organizationID string) error {
	return SendEmailWithBodyAndAttachmentAndOrg(recipient, subject, body, language, nil, organizationID)
}

func SendEmailWithBodyAndAttachment(recipient *MailAddress, subject, body, language string, attachments []*MailAttachment) error {
	return SendEmailWithBodyAndAttachmentAndOrg(recipient, subject, body, language, attachments, "")
}

func SendEmailWithBodyAndAttachmentAndOrg(recipient *MailAddress, subject, body, language string, attachments []*MailAttachment, organizationID string) error {
	// Validate and sanitize recipient address
	recipient.Address = SanitizeEmailAddress(recipient.Address)
	if err := ValidateEmailAddress(recipient.Address); err != nil {
		return fmt.Errorf("invalid recipient email address: %w", err)
	}

	// Validate and sanitize display name
	recipient.DisplayName = SanitizeDisplayName(recipient.DisplayName)
	if err := ValidateDisplayName(recipient.DisplayName); err != nil {
		return fmt.Errorf("invalid recipient display name: %w", err)
	}

	// Validate and sanitize subject
	subject = SanitizeEmailSubject(subject)
	if err := ValidateEmailSubject(subject); err != nil {
		return fmt.Errorf("invalid email subject: %w", err)
	}

	logoData, err := os.ReadFile(filepath.Join(GetConfig().FilesystemBasePath, "./res/seatsurfing.png"))
	if err != nil {
		return fmt.Errorf("error reading logo file: %v", err)
	}
	attachments = append(attachments, &MailAttachment{
		Filename:  "seatsurfing.png",
		Data:      logoData,
		MimeType:  "image/png",
		ContentID: "seatsurfing-logo",
	})
	footerFile, err := GetEmailTemplatePath(GetEmailTemplatePathFooter(), language)
	if err != nil {
		return fmt.Errorf("error getting footer template path: %v", err)
	}
	footerData, err := os.ReadFile(footerFile)
	if err != nil {
		return fmt.Errorf("error reading footer template file: %v", err)
	}
	var jsonFooter []string
	if err := json.Unmarshal(footerData, &jsonFooter); err != nil {
		return fmt.Errorf("error unmarshalling footer json: %v", err)
	}
	footer := ""
	for _, paragraph := range jsonFooter {
		footer += "<p>" + html.EscapeString(paragraph) + "</p>"
	}
	body = strings.ReplaceAll(body, "{{footer}}", footer)
	if GetConfig().MockSendmail {
		SendMailMockContent = body
		return nil
	}
	sender := &MailAddress{
		Address:     GetConfig().MailSenderAddress,
		DisplayName: "Seatsurfing",
	}
	if GetConfig().MailService == "acs" {
		err = acsDialAndSend(recipient, sender, subject, "", body, attachments)
	} else {
		buf := bytes.NewBuffer(nil)
		fmt.Fprintf(buf, "From: %s\n", sender.DisplayName+" <"+sender.Address+">")
		fmt.Fprintf(buf, "To: %s\n", recipient.Address)
		fmt.Fprintf(buf, "Subject: %s\n", subject)
		buf.WriteString("MIME-Version: 1.0\n")

		writer := multipart.NewWriter(buf)
		boundary := writer.Boundary()
		fmt.Fprintf(buf, "Content-Type: multipart/mixed; boundary=\"%s\"\n", boundary)

		// Write body
		fmt.Fprintf(buf, "\n--%s\n", boundary)
		buf.WriteString("Content-Type: text/html; charset=utf-8\n")
		buf.WriteString("Content-Transfer-Encoding: base64\n")
		fmt.Fprintf(buf, "\n%s\n", base64.StdEncoding.EncodeToString([]byte(body)))

		// Write attachments
		for _, attachment := range attachments {
			fmt.Fprintf(buf, "--%s\n", boundary)
			fmt.Fprintf(buf, "Content-Disposition: attachment; filename=\"%s\"\n", attachment.Filename)
			fmt.Fprintf(buf, "Content-Type: %s\n", attachment.MimeType)
			if attachment.ContentID != "" {
				fmt.Fprintf(buf, "Content-ID: <%s>\n", attachment.ContentID)
			}
			buf.WriteString("Content-Transfer-Encoding: base64\n")
			fmt.Fprintf(buf, "\n%s\n", base64.StdEncoding.EncodeToString(attachment.Data))
		}
		fmt.Fprintf(buf, "--%s--\n", boundary)

		to := []string{recipient.Address}
		err = smtpDialAndSend(sender.Address, to, buf.Bytes())
	}

	// Log email to database if sending was successful
	if err == nil && emailLogCallback != nil {
		if logErr := emailLogCallback(subject, recipient.Address, organizationID); logErr != nil {
			log.Printf("Failed to log email to database: %v\n", logErr)
		}
	}

	return err
}

// ValidateEmailAddress checks if an email address is valid and safe from header injection
func ValidateEmailAddress(email string) error {
	if email == "" {
		return ErrInvalidEmailAddress
	}

	if len(email) > 254 {
		return ErrInvalidEmailAddress
	}

	// Check for newline characters that could be used for header injection
	if strings.ContainsAny(email, "\r\n") {
		return ErrInvalidEmailAddress
	}

	// Basic email format validation
	if !strings.Contains(email, "@") {
		return ErrInvalidEmailAddress
	}

	// Check for potentially dangerous characters
	dangerousChars := []string{"\x00", "\x1f", "<script", "javascript:", "data:"}
	emailLower := strings.ToLower(email)
	for _, char := range dangerousChars {
		if strings.Contains(emailLower, char) {
			return ErrInvalidEmailAddress
		}
	}

	return nil
}

// ValidateDisplayName checks if a display name is safe from header injection
func ValidateDisplayName(displayName string) error {
	if len(displayName) > 78 {
		return ErrInvalidDisplayName
	}

	// Check for newline characters that could be used for header injection
	if strings.ContainsAny(displayName, "\r\n") {
		return ErrInvalidDisplayName
	}

	// Check for null bytes and control characters
	if strings.ContainsAny(displayName, "\x00\x1f") {
		return ErrInvalidDisplayName
	}

	// Check for potentially dangerous patterns
	dangerousPatterns := []string{"<script", "javascript:", "data:", "vbscript:"}
	displayNameLower := strings.ToLower(displayName)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(displayNameLower, pattern) {
			return ErrInvalidDisplayName
		}
	}

	return nil
}

// ValidateEmailSubject checks if an email subject is safe from header injection
func ValidateEmailSubject(subject string) error {
	if len(subject) > 256 {
		return ErrInvalidSubject
	}

	// Check for newline characters that could be used for header injection
	if strings.ContainsAny(subject, "\r\n") {
		return ErrInvalidSubject
	}

	// Check for null bytes and certain control characters
	if strings.ContainsAny(subject, "\x00\x1f") {
		return ErrInvalidSubject
	}

	// Check for MIME boundary markers that could break email structure
	if strings.Contains(subject, "boundary=") || strings.Contains(subject, "--") {
		return ErrInvalidSubject
	}

	return nil
}

// SanitizeEmailAddress removes dangerous characters from email address
func SanitizeEmailAddress(email string) string {
	// Remove newlines and null bytes
	email = strings.ReplaceAll(email, "\r", "")
	email = strings.ReplaceAll(email, "\n", "")
	email = strings.ReplaceAll(email, "\x00", "")
	return strings.TrimSpace(email)
}

// SanitizeDisplayName removes dangerous characters from display name
func SanitizeDisplayName(displayName string) string {
	// Remove newlines and null bytes
	displayName = strings.ReplaceAll(displayName, "\r", "")
	displayName = strings.ReplaceAll(displayName, "\n", "")
	displayName = strings.ReplaceAll(displayName, "\x00", "")
	return strings.TrimSpace(displayName)
}

// SanitizeEmailSubject removes dangerous characters from email subject
func SanitizeEmailSubject(subject string) string {
	// Remove newlines and null bytes
	subject = strings.ReplaceAll(subject, "\r", "")
	subject = strings.ReplaceAll(subject, "\n", "")
	subject = strings.ReplaceAll(subject, "\x00", "")
	return strings.TrimSpace(subject)
}

func GetEmailTemplatePath(templateFile, language string) (string, error) {
	if !GetConfig().IsValidLanguageCode(language) {
		language = EmailTemplateDefaultLanguage
	}
	res := strings.ReplaceAll(templateFile, ".json", "_"+language+".json")
	if _, err := os.Stat(res); err == nil {
		return res, nil
	}
	if language == EmailTemplateDefaultLanguage {
		return "", os.ErrNotExist
	}

	res = strings.ReplaceAll(templateFile, ".json", "_"+EmailTemplateDefaultLanguage+".json")
	if _, err := os.Stat(res); err == nil {
		return res, nil
	}
	return "", os.ErrNotExist
}

func GetLocalPartFromEmailAddress(email string) string {
	idx := strings.LastIndex(email, "@")
	if idx == -1 {
		return email
	}
	return email[:idx]
}

func acsDialAndSend(recipient, sender *MailAddress, subject, bodyPlainText, bodyHTML string, attachments []*MailAttachment) error {
	attachmentsList := []ACSAttachment{}
	for _, attachment := range attachments {
		attachmentsList = append(attachmentsList, ACSAttachment{
			Name:            attachment.Filename,
			ContentType:     attachment.MimeType,
			ContentInBase64: base64.StdEncoding.EncodeToString(attachment.Data),
			ContentID:       attachment.ContentID,
		})
	}
	mail := &ACSSendMailRequest{
		SenderAddress: sender.Address,
		Recipients: ACSRecipients{
			To: []ACSAddress{
				{
					Address:     recipient.Address,
					DisplayName: recipient.DisplayName,
				},
			},
		},
		Content: ACSSendMailContent{
			Subject:   subject,
			Plaintext: bodyPlainText,
			HTML:      bodyHTML,
		},
		Attachments: attachmentsList,
	}
	return ACSSendEmail(GetConfig().ACSHost, GetConfig().ACSAccessKey, mail)
}

func smtpDialAndSend(from string, to []string, msg []byte) error {
	config := GetConfig()

	// Get optimal settings for the SMTP server
	port, startTLS, authMethod := getOptimalSMTPSettings(config)
	addr := config.SMTPHost + ":" + strconv.Itoa(port)
	c, err := smtp.Dial(addr)
	if err != nil {
		log.Println("Error dialing SMTP server:", err)
		return err
	}
	defer c.Close()

	// Always check and use STARTTLS if available, especially important for M365
	if ok, _ := c.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			ServerName:         config.SMTPHost,
			InsecureSkipVerify: config.SMTPInsecureSkipVerify,
		}
		if err = c.StartTLS(tlsConfig); err != nil {
			log.Println("Error starting TLS with SMTP server:", err)
			return err
		}
	} else if startTLS {
		// If STARTTLS is required but not available, fail
		log.Println("STARTTLS required but not supported by server")
		return fmt.Errorf("STARTTLS required but not supported by server")
	}

	if config.SMTPAuth {
		var auth smtp.Auth
		actualAuthMethod := strings.ToUpper(authMethod)

		switch actualAuthMethod {
		case "LOGIN":
			auth = NewLoginAuth(config.SMTPAuthUser, config.SMTPAuthPass)
		case "PLAIN", "":
			auth = smtp.PlainAuth("", config.SMTPAuthUser, config.SMTPAuthPass, config.SMTPHost)
		default:
			log.Printf("Warning: Unknown SMTP auth method '%s', falling back to PLAIN", config.SMTPAuthMethod)
			auth = smtp.PlainAuth("", config.SMTPAuthUser, config.SMTPAuthPass, config.SMTPHost)
		}

		if err = c.Auth(auth); err != nil {
			log.Printf("Error authenticating with SMTP server using %s method: %v", actualAuthMethod, err)

			// For M365 compatibility, try LOGIN method if PLAIN fails
			if actualAuthMethod == "PLAIN" {
				log.Println("Retrying with LOGIN authentication method for M365 compatibility...")
				loginAuth := NewLoginAuth(config.SMTPAuthUser, config.SMTPAuthPass)
				if err = c.Auth(loginAuth); err != nil {
					log.Println("Error authenticating with SMTP server using LOGIN method:", err)
					return err
				}
			} else {
				return err
			}
		}
	}
	if err = c.Mail(from); err != nil {
		log.Println("Error sending 'Mail From' to SMTP server:", err)
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			log.Println("Error sending 'Rcpt To' to SMTP server:", err)
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		log.Println("Error sending 'Data' to SMTP server:", err)
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		log.Println("Error writing message to SMTP server:", err)
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}
