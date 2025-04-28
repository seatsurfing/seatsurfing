package util

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/smtp"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	. "github.com/seatsurfing/seatsurfing/server/config"
)

const EmailTemplateDefaultLanguage = "en"

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

func GetEmailTemplatePathResetpassword() string {
	return filepath.Join(GetConfig().FilesystemBasePath, "./res/email-resetpw.json")
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
	for key, val := range vars {
		body = strings.ReplaceAll(body, "{{"+key+"}}", val)
	}
	return SendEmailWithBody(recipient, mailTemplate.Subject, body, language)
}

func SendEmailWithBody(recipient *MailAddress, subject, body, language string) error {
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
		return acsDialAndSend(recipient, sender, subject, body)
	} else {
		to := []string{recipient.Address}
		body = "From: " + sender.DisplayName + " <" + sender.Address + ">\n" +
			"To: " + recipient.Address + "\n" +
			"Content-Type: text/html; charset=UTF-8\n" +
			"Subject: " + subject + "\n" +
			"\n" +
			body
		msg := []byte(body)
		return smtpDialAndSend(sender.Address, to, msg)
	}
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

func acsDialAndSend(recipient, sender *MailAddress, subject, body string) error {
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
			Plaintext: body,
		},
	}
	return ACSSendEmail(GetConfig().ACSHost, GetConfig().ACSAccessKey, mail)
}

func smtpDialAndSend(from string, to []string, msg []byte) error {
	config := GetConfig()
	addr := config.SMTPHost + ":" + strconv.Itoa(config.SMTPPort)
	c, err := smtp.Dial(addr)
	if err != nil {
		log.Println("Error dialing SMTP server:", err)
		return err
	}
	defer c.Close()
	if config.SMTPStartTLS {
		if ok, _ := c.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{
				ServerName:         config.SMTPHost,
				InsecureSkipVerify: config.SMTPInsecureSkipVerify,
			}
			if err = c.StartTLS(tlsConfig); err != nil {
				log.Println("Error starting TLS with SMTP server:", err)
				return err
			}
		}
	}
	if config.SMTPAuth {
		auth := smtp.PlainAuth("", config.SMTPAuthUser, config.SMTPAuthPass, config.SMTPHost)
		if err = c.Auth(auth); err != nil {
			log.Println("Error authenticating with SMTP server:", err)
			return err
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
