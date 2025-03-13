package util

import (
	"crypto/tls"
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

func GetEmailTemplatePathResetpassword() string {
	return filepath.Join(GetConfig().FilesystemBasePath, "./res/email-resetpw.txt")
}

func SendEmail(recipient, sender, templateFile, language string, vars map[string]string) error {
	actualTemplateFile, err := GetEmailTemplatePath(templateFile, language)
	if err != nil {
		return err
	}
	body, err := compileEmailTemplateFromFile(actualTemplateFile, vars)
	if err != nil {
		return err
	}
	return SendEmailWithBody(recipient, sender, body)
}

func SendEmailWithBody(recipient, sender, body string) error {
	if GetConfig().MockSendmail {
		SendMailMockContent = body
		return nil
	}
	to := []string{recipient}
	msg := []byte(body)
	err := smtpDialAndSend(sender, to, msg)
	return err
}

func GetEmailTemplatePath(templateFile, language string) (string, error) {
	if !GetConfig().IsValidLanguageCode(language) {
		language = EmailTemplateDefaultLanguage
	}
	res := strings.ReplaceAll(templateFile, ".txt", "_"+language+".txt")
	if _, err := os.Stat(res); err == nil {
		return res, nil
	}
	if language == EmailTemplateDefaultLanguage {
		return "", os.ErrNotExist
	}

	res = strings.ReplaceAll(templateFile, ".txt", "_"+EmailTemplateDefaultLanguage+".txt")
	if _, err := os.Stat(res); err == nil {
		return res, nil
	}
	return "", os.ErrNotExist
}

func CompileEmailTemplate(template string, vars map[string]string) string {
	c := GetConfig()
	vars["senderAddress"] = c.SMTPSenderAddress
	for key, val := range vars {
		template = strings.ReplaceAll(template, "{{"+key+"}}", val)
	}
	return template
}

func compileEmailTemplateFromFile(templateFile string, vars map[string]string) (string, error) {
	data, err := os.ReadFile(templateFile)
	if err != nil {
		return "", err
	}
	s := string(data)
	return CompileEmailTemplate(s, vars), nil
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
