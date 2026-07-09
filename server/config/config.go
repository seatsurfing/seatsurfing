package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"log"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RemotePluginConfig describes one remote plugin the host should connect to over
// gRPC. Replaces the old directory-scan discovery (FilesystemBasePath /
// PluginsSubPath) since there is no local binary to discover once plugins
// are separate processes/containers - see PLUGINS_CONFIG below.
type RemotePluginConfig struct {
	Name          string   `json:"name"`
	Address       string   `json:"address"`       // e.g. "subscription-plugin:50051"
	RoutePrefixes []string `json:"routePrefixes"` // e.g. ["/subscription/"] - the host's router is built from this at startup, independent of plugin liveness
	Token         string   `json:"token"`         // shared secret the host presents when dialing this plugin
	TLS           bool     `json:"tls"`
}

type Config struct {
	PublicListenAddr                    string
	PostgresURL                         string
	JwtPrivateKey                       *rsa.PrivateKey
	JwtPublicKey                        *rsa.PublicKey
	StaticUiPath                        string
	MailService                         string
	MailSenderAddress                   string
	SMTPHost                            string
	SMTPPort                            int
	SMTPStartTLS                        bool
	SMTPInsecureSkipVerify              bool
	SMTPAuth                            bool
	SMTPAuthUser                        string
	SMTPAuthPass                        string
	SMTPAuthMethod                      string
	ACSHost                             string
	ACSAccessKey                        string
	MockSendmail                        bool
	Development                         bool
	InitOrgName                         string
	InitOrgUser                         string
	InitOrgPass                         string
	InitOrgLanguage                     string
	InitOrgDomain                       string
	AllowOrgDelete                      bool
	LoginProtectionMaxFails             int
	LoginProtectionSlidingWindowSeconds int
	LoginProtectionBanMinutes           int
	CryptKey                            string
	FilesystemBasePath                  string
	Plugins                             []RemotePluginConfig
	HostAPIListenAddr                   string
	HostAPIToken                        string
	PluginCallTimeout                   time.Duration
	PluginMaxMsgSize                    int
	PublicScheme                        string
	PublicPort                          int
	CacheType                           string // "valkey" or "default"
	ValkeyHosts                         []string
	ValkeyUsername                      string
	ValkeyPassword                      string
	DNSServer                           string // DNS server address for custom resolver
	DisablePasswordLogin                bool   // Disable password login for all users (only allow OAuth2 and SSO)
	CORSOrigins                         []string
	RateLimit                           int
	RateLimitPeriod                     string // e.g., "1-M" for 1 minute
	MaxSessionsPerUser                  int    // Maximum number of concurrent sessions per user
	WebAuthnRPDisplayName               string
	MaxPasskeysPerUser                  int  // Maximum number of passkeys a single user may register
	DisableVersionCheck                 bool // Disable polling seatsurfing.io for latest version information
	DisableAnonymousUsageStats          bool // Disable sending anonymous usage statistics for this installation
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
	log.Println("Reading config …")
	c.Development = (c.getEnv("DEV", "0") == "1")
	if c.Development {
		log.Println("ℹ️  Development mode is enabled, do not use this in production environments!")
	}
	c.PublicListenAddr = c.getEnv("PUBLIC_LISTEN_ADDR", "0.0.0.0:8080")
	c.StaticUiPath = strings.TrimSuffix(c.getEnv("STATIC_UI_PATH", "/app/ui"), "/") + "/"
	c.PostgresURL = c.getEnv("POSTGRES_URL", "postgres://postgres:root@localhost/seatsurfing?sslmode=disable")

	privateKeyFile := c.getEnv("JWT_PRIVATE_KEY", "")
	privateKey, err := c.loadPrivateKey(privateKeyFile)
	if privateKeyFile != "" && err != nil {
		log.Println("⚠️  Warning: Loading private key failed", err)
	}
	publicKeyFile := c.getEnv("JWT_PUBLIC_KEY", "")
	publicKey, err := c.loadPublicKey(publicKeyFile)
	if publicKeyFile != "" && err != nil {
		log.Println("⚠️  Warning: Loading public key failed", err)
	}
	if publicKey == nil || privateKey == nil {
		log.Println("⚠️  Warning: No valid JWT_PRIVATE_KEY or JWT_PUBLIC_KEY set. Generating a temporary random private/public key pair …")
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
	c.SMTPAuthMethod = strings.ToUpper(c.getEnv("SMTP_AUTH_METHOD", "PLAIN"))
	if c.SMTPAuthMethod != "PLAIN" && c.SMTPAuthMethod != "LOGIN" {
		log.Println("⚠️  Warning: Invalid SMTP_AUTH_METHOD set. Only 'PLAIN' and 'LOGIN' are allowed. Defaulting to 'PLAIN'.")
		c.SMTPAuthMethod = "PLAIN"
	}
	c.MailSenderAddress = c.getEnv("MAIL_SENDER_ADDRESS", "no-reply@localhost")
	if c.MailSenderAddress == "" {
		// Deprecated
		c.MailSenderAddress = c.getEnv("SMTP_SENDER_ADDRESS", "no-reply@localhost")
	}
	c.MailService = strings.ToLower(c.getEnv("MAIL_SERVICE", "smtp"))
	if c.MailService != "smtp" && c.MailService != "acs" {
		log.Println("⚠️  Warning: Invalid MAIL_SERVICE set. Only 'smtp' and 'acs' are allowed. Defaulting to 'smtp'.")
		c.MailService = "smtp"
	}
	c.ACSHost = c.getEnv("ACS_HOST", "")
	c.ACSAccessKey = c.getEnv("ACS_ACCESS_KEY", "")
	c.MockSendmail = (c.getEnv("MOCK_SENDMAIL", "0") == "1")
	c.InitOrgName = c.getEnv("INIT_ORG_NAME", "Sample Company")
	c.InitOrgUser = c.getEnv("INIT_ORG_USER", "admin")
	c.InitOrgPass = c.getEnv("INIT_ORG_PASS", "Sea!surf1ng")
	c.InitOrgLanguage = c.getEnv("INIT_ORG_LANGUAGE", "en")
	if !c.IsValidLanguageCode(c.InitOrgLanguage) {
		log.Println("⚠️  Warning: Invalid INIT_ORG_LANGUAGE set. Defaulting to 'en'.")
		c.InitOrgLanguage = "en"
	}
	c.InitOrgDomain = c.getEnv("INIT_ORG_DOMAIN", "localhost")
	c.AllowOrgDelete = (c.getEnv("ALLOW_ORG_DELETE", "0") == "1")
	c.LoginProtectionMaxFails = c.getEnvInt("LOGIN_PROTECTION_MAX_FAILS", 10)
	c.LoginProtectionSlidingWindowSeconds = c.getEnvInt("LOGIN_PROTECTION_SLIDING_WINDOW_SECONDS", 600)
	c.LoginProtectionBanMinutes = c.getEnvInt("LOGIN_PROTECTION_BAN_MINUTES", 5)
	c.CryptKey = c.getEnv("CRYPT_KEY", "")
	if c.CryptKey == "" || len(c.CryptKey) != 32 {
		log.Fatalln("Error: No valid CRYPT_KEY set. CRYPT_KEY needs to be set to a 32 bytes long random string.")
	}
	pwd, _ := os.Getwd()
	c.FilesystemBasePath = c.getEnv("FILESYSTEM_BASE_PATH", pwd)
	c.Plugins = c.parsePluginsConfig(c.getEnv("PLUGINS_CONFIG", ""))
	c.HostAPIListenAddr = c.getEnv("HOSTAPI_LISTEN_ADDR", "0.0.0.0:50052")
	c.HostAPIToken = c.getEnv("HOSTAPI_TOKEN", "")
	if len(c.Plugins) > 0 && c.HostAPIToken == "" {
		log.Println("⚠️  Warning: PLUGINS_CONFIG is set but HOSTAPI_TOKEN is empty - the HostAPI gRPC listener will accept an empty token from any caller that can reach it. Set HOSTAPI_TOKEN to a random secret shared with your plugin(s).")
	}
	c.PluginCallTimeout = time.Duration(c.getEnvInt("PLUGIN_CALL_TIMEOUT_SECONDS", 30)) * time.Second
	c.PluginMaxMsgSize = c.getEnvInt("PLUGIN_MAX_MSG_SIZE", 16<<20)
	c.PublicScheme = c.getEnv("PUBLIC_SCHEME", "https")
	c.PublicPort = c.getEnvInt("PUBLIC_PORT", 443)
	c.CacheType = c.getEnv("CACHE_TYPE", "default")
	if c.CacheType != "valkey" && c.CacheType != "default" {
		log.Println("⚠️  Warning: Invalid CACHE_TYPE set. Only 'valkey' and 'default' are allowed. Defaulting to 'default'.")
		c.CacheType = "default"
	}
	c.ValkeyHosts = strings.Split(c.getEnv("VALKEY_HOSTS", "127.0.0.1:6379"), ",")
	c.ValkeyUsername = c.getEnv("VALKEY_USERNAME", "default")
	c.ValkeyPassword = c.getEnv("VALKEY_PASSWORD", "")
	c.DNSServer = c.getEnv("DNS_SERVER", "")
	if c.DNSServer != "" {
		if !strings.ContainsRune(c.DNSServer, ':') {
			c.DNSServer += ":53" // Default DNS port
		}
	}
	c.DisablePasswordLogin = (c.getEnv("DISABLE_PASSWORD_LOGIN", "0") == "1")
	c.CORSOrigins = strings.Split(c.getEnv("CORS_ORIGINS", ""), ",")
	if len(c.CORSOrigins) == 1 && c.CORSOrigins[0] == "" {
		c.CORSOrigins = []string{}
	}
	if c.Development && !slices.Contains(c.CORSOrigins, "http://localhost:3000") {
		c.CORSOrigins = append(c.CORSOrigins, "http://localhost:3000")
	}
	c.RateLimit = c.getEnvInt("RATE_LIMIT", 250)
	c.RateLimitPeriod = c.getEnv("RATE_LIMIT_PERIOD", "1-M")
	rxRateLimitPeriod := regexp.MustCompile(`^[0-9]+\-[SMHD]$`)
	if !rxRateLimitPeriod.MatchString(c.RateLimitPeriod) {
		log.Println("⚠️  Warning: Invalid RATE_LIMIT_PERIOD set. Must be in format '<number>-<S|M|H|D>'. Defaulting to '1-M'.")
		c.RateLimitPeriod = "1-M"
	}
	c.MaxSessionsPerUser = c.getEnvInt("MAX_SESSIONS_PER_USER", 10)
	if c.MaxSessionsPerUser < 1 {
		log.Println("⚠️  Warning: MAX_SESSIONS_PER_USER must be at least 1. Defaulting to 10.")
		c.MaxSessionsPerUser = 10
	}
	c.WebAuthnRPDisplayName = c.getEnv("WEBAUTHN_RP_DISPLAY_NAME", "Seatsurfing")
	c.MaxPasskeysPerUser = c.getEnvInt("MAX_PASSKEYS_PER_USER", 10)
	if c.MaxPasskeysPerUser < 1 {
		log.Println("⚠️  Warning: MAX_PASSKEYS_PER_USER must be at least 1. Defaulting to 10.")
		c.MaxPasskeysPerUser = 10
	}
	c.DisableVersionCheck = (c.getEnv("DISABLE_VERSION_CHECK", "0") == "1")
	c.DisableAnonymousUsageStats = (c.getEnv("DISABLE_ANONYMOUS_USAGE_STATS", "0") == "1")

	// Check deprecated environment variables
	if c.getEnv("ADMIN_UI_BACKEND", "") != "" {
		log.Println("⚠️  Warning: ADMIN_UI_BACKEND is deprecated. The Admin UI now uses the same backend as the booking UI. Please remove this environment variable.")
	}
	if c.getEnv("BOOKING_UI_BACKEND", "") != "" {
		log.Println("⚠️  Warning: BOOKING_UI_BACKEND is deprecated. The Booking UI now uses the same backend as the Admin UI. Please remove this environment variable.")
	}
	if c.getEnv("DISABLE_UI_PROXY", "") != "" {
		log.Println("⚠️  Warning: DISABLE_UI_PROXY is deprecated. Admin and Booking UI assets are now part of the backend. Please adjust your proxy configuration accordingly.")
	}
}

// parsePluginsConfig parses the PLUGINS_CONFIG env var, a JSON array of
// RemotePluginConfig entries, e.g.:
//
//	[{"name":"subscription","address":"subscription-plugin:50051","routePrefixes":["/subscription/"],"token":"<secret>","tls":false}]
//
// A single env var supports an arbitrary-length list of plugins - unset or
// empty means zero plugins loaded, matching the old behavior of a missing
// plugins directory being a non-fatal, valid state.
func (c *Config) parsePluginsConfig(raw string) []RemotePluginConfig {
	if raw == "" {
		return nil
	}
	var plugins []RemotePluginConfig
	if err := json.Unmarshal([]byte(raw), &plugins); err != nil {
		log.Fatalln("Error: Could not parse PLUGINS_CONFIG as JSON:", err)
	}
	return plugins
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
	for _, s := range validLanguageCodes {
		if isoLanguageCode == s {
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
