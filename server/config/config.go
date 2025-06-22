package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Config struct {
	PublicListenAddr                    string
	PostgresURL                         string
	JwtPrivateKey                       *rsa.PrivateKey
	JwtPublicKey                        *rsa.PublicKey
	StaticAdminUiPath                   string
	StaticBookingUiPath                 string
	MailService                         string
	MailSenderAddress                   string
	SMTPHost                            string
	SMTPPort                            int
	SMTPStartTLS                        bool
	SMTPInsecureSkipVerify              bool
	SMTPAuth                            bool
	SMTPAuthUser                        string
	SMTPAuthPass                        string
	ACSHost                             string
	ACSAccessKey                        string
	MockSendmail                        bool
	Development                         bool
	InitOrgName                         string
	InitOrgUser                         string
	InitOrgPass                         string
	InitOrgLanguage                     string
	AllowOrgDelete                      bool
	LoginProtectionMaxFails             int
	LoginProtectionSlidingWindowSeconds int
	LoginProtectionBanMinutes           int
	CryptKey                            string
	FilesystemBasePath                  string
	PluginsSubPath                      string
	PublicScheme                        string
	PublicPort                          int
}

var _configInstance *Config
var _configOnce sync.Once

func GetConfig() *Config {
	_configOnce.Do(func() {
		_configInstance = &Config{}
		_configInstance.ReadConfig()
	})
	return _configInstance
}

func (c *Config) ReadConfig() {
	log.Println("Reading config...")
	c.Development = (c.getEnv("DEV", "0") == "1")
	c.PublicListenAddr = c.getEnv("PUBLIC_LISTEN_ADDR", "0.0.0.0:8080")
	c.StaticAdminUiPath = strings.TrimSuffix(c.getEnv("STATIC_ADMIN_UI_PATH", "/app/admin-ui"), "/") + "/"
	c.StaticBookingUiPath = strings.TrimSuffix(c.getEnv("STATIC_BOOKING_UI_PATH", "/app/booking-ui"), "/") + "/"
	c.PostgresURL = c.getEnv("POSTGRES_URL", "postgres://postgres:root@localhost/seatsurfing?sslmode=disable")
	privateKey, _ := c.loadPrivateKey(c.getEnv("JWT_PRIVATE_KEY", ""))
	publicKey, _ := c.loadPublicKey(c.getEnv("JWT_PUBLIC_KEY", ""))
	if publicKey == nil || privateKey == nil {
		log.Println("Warning: No valid JWT_PRIVATE_KEY or JWT_PUBLIC_KEY set. Generating a temporary random private/public key pair...")
		privkey, _ := rsa.GenerateKey(rand.Reader, 2048)
		c.JwtPrivateKey = privkey
		c.JwtPublicKey = &privkey.PublicKey
	} else {
		c.JwtPrivateKey = privateKey
		c.JwtPublicKey = publicKey
	}
	c.SMTPHost = c.getEnv("SMTP_HOST", "127.0.0.1")
	c.SMTPPort = c.getEnvInt("SMTP_PORT", 25)
	c.SMTPStartTLS = (c.getEnv("SMTP_START_TLS", "0") == "1")
	c.SMTPInsecureSkipVerify = (c.getEnv("SMTP_INSECURE_SKIP_VERIFY", "0") == "1")
	c.SMTPAuth = (c.getEnv("SMTP_AUTH", "0") == "1")
	c.SMTPAuthUser = c.getEnv("SMTP_AUTH_USER", "")
	c.SMTPAuthPass = c.getEnv("SMTP_AUTH_PASS", "")
	c.MailSenderAddress = c.getEnv("MAIL_SENDER_ADDRESS", "no-reply@seatsurfing.local")
	if c.MailSenderAddress == "" {
		// Deprecated
		c.MailSenderAddress = c.getEnv("SMTP_SENDER_ADDRESS", "no-reply@seatsurfing.local")
	}
	c.MailService = c.getEnv("MAIL_SERVICE", "smtp")
	if c.MailService != "smtp" && c.MailService != "acs" {
		log.Println("Warning: Invalid MAIL_SERVICE set. Only 'smtp' and 'acs' are allowed. Defaulting to 'smtp'.")
	}
	c.ACSHost = c.getEnv("ACS_HOST", "")
	c.ACSAccessKey = c.getEnv("ACS_ACCESS_KEY", "")
	c.MockSendmail = (c.getEnv("MOCK_SENDMAIL", "0") == "1")
	c.InitOrgName = c.getEnv("INIT_ORG_NAME", "Sample Company")
	c.InitOrgUser = c.getEnv("INIT_ORG_USER", "admin")
	c.InitOrgPass = c.getEnv("INIT_ORG_PASS", "12345678")
	c.InitOrgLanguage = c.getEnv("INIT_ORG_LANGUAGE", "en")
	c.AllowOrgDelete = (c.getEnv("ALLOW_ORG_DELETE", "0") == "1")
	c.LoginProtectionMaxFails = c.getEnvInt("LOGIN_PROTECTION_MAX_FAILS", 10)
	c.LoginProtectionSlidingWindowSeconds = c.getEnvInt("LOGIN_PROTECTION_SLIDING_WINDOW_SECONDS", 600)
	c.LoginProtectionBanMinutes = c.getEnvInt("LOGIN_PROTECTION_BAN_MINUTES", 5)
	c.CryptKey = c.getEnv("CRYPT_KEY", "")
	if c.CryptKey == "" || len(c.CryptKey) != 32 {
		log.Println("Warning: No valid CRYPT_KEY set. Set it to a 32 bytes long string in order to use features such as CalDAV integration.")
	}
	pwd, _ := os.Getwd()
	c.FilesystemBasePath = c.getEnv("FILESYSTEM_BASE_PATH", pwd)
	c.PluginsSubPath = c.getEnv("PLUGINS_SUB_PATH", "plugins")
	c.PublicScheme = c.getEnv("PUBLIC_SCHEME", "https")
	c.PublicPort = c.getEnvInt("PUBLIC_PORT", 443)

	// Check deprecated environment variables
	if c.getEnv("ADMIN_UI_BACKEND", "") != "" {
		log.Println("Warning: ADMIN_UI_BACKEND is deprecated. The Admin UI now uses the same backend as the booking UI. Please remove this environment variable.")
	}
	if c.getEnv("BOOKING_UI_BACKEND", "") != "" {
		log.Println("Warning: BOOKING_UI_BACKEND is deprecated. The Booking UI now uses the same backend as the Admin UI. Please remove this environment variable.")
	}
	if c.getEnv("DISABLE_UI_PROXY", "") != "" {
		log.Println("Warning: DISABLE_UI_PROXY is deprecated. Admin and Booking UI assets are now part of the backend. Please adjust your proxy configuration accordingly.")
	}
}

func (c *Config) loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode([]byte(pemBytes))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func (c *Config) loadPublicKey(path string) (*rsa.PublicKey, error) {
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode([]byte(pemBytes))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		return nil, errors.New("key type is not RSA")
	}
}

func (c *Config) IsValidLanguageCode(isoLanguageCode string) bool {
	validLanguageCodes := []string{"de", "en"}
	lc := strings.ToLower(isoLanguageCode)
	for _, s := range validLanguageCodes {
		if lc == s {
			return true
		}
	}
	return false
}

func (c *Config) getEnv(key, defaultValue string) string {
	res := os.Getenv(key)
	if res == "" {
		return defaultValue
	}
	return res
}

func (c *Config) getEnvInt(key string, defaultValue int) int {
	val, err := strconv.Atoi(c.getEnv(key, strconv.Itoa(defaultValue)))
	if err != nil {
		log.Fatal("Could not parse " + key + " to int")
	}
	return val
}
