package util

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
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

type MailAttachment struct {
	Filename  string
	Data      []byte
	MimeType  string
	ContentID string
}

func GetEmailTemplatePathResetpassword() string {
	return filepath.Join(GetConfig().FilesystemBasePath, "./res/email-resetpw.json")
}

func GetEmailTemplatePathChangeEmailAddress() string {
	return filepath.Join(GetConfig().FilesystemBasePath, "./res/email-change-email.json")
}

func GetEmailTemplatePathBookingCreated() string {
	return filepath.Join(GetConfig().FilesystemBasePath, "./res/email-booking-created.json")
}

func GetEmailTemplatePathBookingDeclined() string {
	return filepath.Join(GetConfig().FilesystemBasePath, "./res/email-booking-declined.json")
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
	return SendEmailWithAttachments(recipient, templateFile, language, vars, nil)
}

func SendEmailWithAttachments(recipient *MailAddress, templateFile, language string, vars map[string]string, attachments []*MailAttachment) error {
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
	return SendEmailWithBodyAndAttachment(recipient, mailTemplate.Subject, body, language, attachments)
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
		body = strings.ReplaceAll(body, "{{"+key+"}}", val)
	}
	return body
}

func SendEmailWithBody(recipient *MailAddress, subject, body, language string) error {
	return SendEmailWithBodyAndAttachment(recipient, subject, body, language, nil)
}

func SendEmailWithBodyAndAttachment(recipient *MailAddress, subject, body, language string, attachments []*MailAttachment) error {
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
		return acsDialAndSend(recipient, sender, subject, "", body, attachments)
	} else {
		buf := bytes.NewBuffer(nil)
		buf.WriteString(fmt.Sprintf("From: %s\n", sender.DisplayName+" <"+sender.Address+">"))
		buf.WriteString(fmt.Sprintf("To: %s\n", recipient.Address))
		buf.WriteString(fmt.Sprintf("Subject: %s\n", subject))
		buf.WriteString("MIME-Version: 1.0\n")

		writer := multipart.NewWriter(buf)
		boundary := writer.Boundary()
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\n", boundary))

		// Write body
		buf.WriteString(fmt.Sprintf("\n--%s\n", boundary))
		buf.WriteString("Content-Type: text/html; charset=utf-8\n")
		buf.WriteString("Content-Transfer-Encoding: base64\n")
		buf.WriteString(fmt.Sprintf("\n%s\n", base64.StdEncoding.EncodeToString([]byte(body))))

		// Write attachments
		for _, attachment := range attachments {
			buf.WriteString(fmt.Sprintf("--%s\n", boundary))
			buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\n", attachment.Filename))
			buf.WriteString(fmt.Sprintf("Content-Type: %s\n", attachment.MimeType))
			if attachment.ContentID != "" {
				buf.WriteString(fmt.Sprintf("Content-ID: <%s>\n", attachment.ContentID))
			}
			buf.WriteString("Content-Transfer-Encoding: base64\n")
			buf.WriteString(fmt.Sprintf("\n%s\n", base64.StdEncoding.EncodeToString(attachment.Data)))
		}
		buf.WriteString(fmt.Sprintf("--%s--\n", boundary))

		to := []string{recipient.Address}
		return smtpDialAndSend(sender.Address, to, buf.Bytes())
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
