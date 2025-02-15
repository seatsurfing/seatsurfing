package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
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
	PublicURL                           string
	FrontendURL                         string
	PostgresURL                         string
	JwtPrivateKey                       *rsa.PrivateKey
	JwtPublicKey                        *rsa.PublicKey
	DisableUiProxy                      bool
	AdminUiBackend                      string
	BookingUiBackend                    string
	SMTPHost                            string
	SMTPPort                            int
	SMTPSenderAddress                   string
	SMTPStartTLS                        bool
	SMTPInsecureSkipVerify              bool
	SMTPAuth                            bool
	SMTPAuthUser                        string
	SMTPAuthPass                        string
	MockSendmail                        bool
	PrintConfig                         bool
	Development                         bool
	InitOrgName                         string
	InitOrgDomain                       string
	InitOrgUser                         string
	InitOrgPass                         string
	InitOrgLanguage                     string
	OrgSignupEnabled                    bool
	OrgSignupDomain                     string
	OrgSignupPreviousDomains            []string
	OrgSignupAdmin                      string
	OrgSignupMaxUsers                   int
	OrgSignupDelete                     bool
	LoginProtectionMaxFails             int
	LoginProtectionSlidingWindowSeconds int
	LoginProtectionBanMinutes           int
	CryptKey                            string
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
	c.PublicURL = strings.TrimSuffix(c.getEnv("PUBLIC_URL", "http://localhost:8080"), "/") + "/"
	if c.Development {
		c.FrontendURL = c.getEnv("FRONTEND_URL", "http://localhost:3000")
	} else {
		c.FrontendURL = c.getEnv("FRONTEND_URL", "http://localhost:8080")
	}
	c.FrontendURL = strings.TrimSuffix(c.FrontendURL, "/") + "/"
	c.DisableUiProxy = (c.getEnv("DISABLE_UI_PROXY", "0") == "1")
	c.AdminUiBackend = c.getEnv("ADMIN_UI_BACKEND", "localhost:3000")
	c.BookingUiBackend = c.getEnv("BOOKING_UI_BACKEND", "localhost:3001")
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
	c.SMTPSenderAddress = c.getEnv("SMTP_SENDER_ADDRESS", "no-reply@seatsurfing.local")
	c.MockSendmail = (c.getEnv("MOCK_SENDMAIL", "0") == "1")
	c.PrintConfig = (c.getEnv("PRINT_CONFIG", "0") == "1")
	c.InitOrgName = c.getEnv("INIT_ORG_NAME", "Sample Company")
	c.InitOrgDomain = c.getEnv("INIT_ORG_DOMAIN", "seatsurfing.local")
	c.InitOrgUser = c.getEnv("INIT_ORG_USER", "admin")
	c.InitOrgPass = c.getEnv("INIT_ORG_PASS", "12345678")
	c.InitOrgLanguage = c.getEnv("INIT_ORG_LANGUAGE", "en")
	c.OrgSignupEnabled = (c.getEnv("ORG_SIGNUP_ENABLED", "0") == "1")
	c.OrgSignupDomain = c.getEnv("ORG_SIGNUP_DOMAIN", ".on.seatsurfing.local")
	c.OrgSignupPreviousDomains = strings.Split(c.getEnv("ORG_SIGNUP_PREVIOUS_DOMAINS", ""), ",")
	c.OrgSignupAdmin = c.getEnv("ORG_SIGNUP_ADMIN", "admin")
	c.OrgSignupMaxUsers = c.getEnvInt("ORG_SIGNUP_MAX_USERS", 10)
	c.OrgSignupDelete = (c.getEnv("ORG_SIGNUP_DELETE", "0") == "1")
	c.LoginProtectionMaxFails = c.getEnvInt("LOGIN_PROTECTION_MAX_FAILS", 10)
	c.LoginProtectionSlidingWindowSeconds = c.getEnvInt("LOGIN_PROTECTION_SLIDING_WINDOW_SECONDS", 600)
	c.LoginProtectionBanMinutes = c.getEnvInt("LOGIN_PROTECTION_BAN_MINUTES", 5)
	c.CryptKey = c.getEnv("CRYPT_KEY", "")
	if c.CryptKey == "" || len(c.CryptKey) != 32 {
		log.Println("Warning: No valid CRYPT_KEY set. Set it to a 32 bytes long string in order to use features such as CalDAV integration.")
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

func (c *Config) isValidLanguageCode(isoLanguageCode string) bool {
	validLanguageCodes := []string{"de", "en"}
	lc := strings.ToLower(isoLanguageCode)
	for _, s := range validLanguageCodes {
		if lc == s {
			return true
		}
	}
	return false
}

func (c *Config) Print() {
	s, _ := json.MarshalIndent(c, "", "\t")
	log.Println("Using config:\n" + string(s))
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
