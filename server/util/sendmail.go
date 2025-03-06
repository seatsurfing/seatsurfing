package util

import (
	"crypto/tls"
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
	vars["frontendUrl"] = c.FrontendURL
	vars["publicUrl"] = c.PublicURL
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
				return err
			}
		}
	}
	if config.SMTPAuth {
		auth := smtp.PlainAuth("", config.SMTPAuthUser, config.SMTPAuthPass, config.SMTPHost)
		if err = c.Auth(auth); err != nil {
			return err
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}
