package config

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/interledger/interledger-app/go/configa"
)

// Config holds application configuration
type Config struct {
	Port                  string
	LogLevel              string
	RedisURL              string
	RedisDB               int
	WebhookURL            string
	WebhookSecret         string
	WebhookMinDelaySec    float64
	UseRedis              bool
	EnforceAuthentication bool
	ValidCredentials      map[string]string // appID -> secret
	DefaultOrganizationID string
	// PublicBaseURL is the externally reachable base URL of mockgatehub
	// (no trailing slash). It is used to build absolute URLs in API responses
	// that are followed directly by the browser (e.g. card-data tokenisation
	// links).
	PublicBaseURL string
	// CardDataTokenSecret is the HMAC secret used to sign card-data JWTs
	// returned by POST /cards/v1/token/card-data. It must not be a hard-coded
	// constant: when unset we generate a random value at startup so mock
	// deployments never share a signing key across processes.
	CardDataTokenSecret string
}

// yamlConfig is the YAML-tagged struct used when CONFIG is set.
type yamlConfig struct {
	Port                  string            `yaml:"port"`
	LogLevel              string            `yaml:"log_level"`
	RedisURL              string            `yaml:"redis_url"`
	RedisDB               int               `yaml:"redis_db"`
	WebhookURL            string            `yaml:"webhook_url"`
	WebhookSecret         string            `yaml:"webhook_secret"`
	WebhookMinDelaySec    float64           `yaml:"webhook_min_delay_sec"`
	EnforceAuthentication bool              `yaml:"enforce_authentication"`
	ValidCredentials      map[string]string `yaml:"valid_credentials"`
	DefaultOrganizationID string            `yaml:"default_organization_id"`
	PublicBaseURL         string            `yaml:"public_base_url"`
	CardDataTokenSecret   string            `yaml:"card_data_token_secret"`
}

// Load reads configuration. When CONFIG is set it is a comma-separated list
// of YAML files (later files overlay earlier ones); otherwise individual
// environment variables are read.
func Load() *Config {
	if v := os.Getenv("CONFIG"); v != "" {
		return loadFromFiles(splitFiles(v))
	}
	return loadFromEnv()
}

func loadFromFiles(files []string) *Config {
	parsed, err := configa.Parse[yamlConfig](files)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal: parse mockgatehub config:", err)
		os.Exit(1)
	}
	y, err := parsed.Resolve(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal: resolve mockgatehub config:", err)
		os.Exit(1)
	}

	cfg := &Config{
		Port:                  y.Port,
		LogLevel:              y.LogLevel,
		RedisURL:              y.RedisURL,
		RedisDB:               y.RedisDB,
		WebhookURL:            y.WebhookURL,
		WebhookSecret:         y.WebhookSecret,
		WebhookMinDelaySec:    y.WebhookMinDelaySec,
		EnforceAuthentication: y.EnforceAuthentication,
		ValidCredentials:      y.ValidCredentials,
		DefaultOrganizationID: y.DefaultOrganizationID,
		PublicBaseURL:         strings.TrimRight(y.PublicBaseURL, "/"),
		CardDataTokenSecret:   y.CardDataTokenSecret,
	}
	return applyDefaults(cfg)
}

func loadFromEnv() *Config {
	cfg := &Config{
		Port:                  getEnv("MOCKGATEHUB_PORT", "8080"),
		LogLevel:              getEnv("LOG_LEVEL", "info"),
		RedisURL:              getEnv("MOCKGATEHUB_REDIS_URL", ""),
		RedisDB:               getEnvInt("MOCKGATEHUB_REDIS_DB", 0),
		WebhookURL:            getEnv("WEBHOOK_URL", ""),
		WebhookSecret:         getEnv("WEBHOOK_SECRET", "mock-secret"),
		WebhookMinDelaySec:    getEnvFloat("WEBHOOK_MIN_DELAY_SEC", 0.05),
		EnforceAuthentication: getEnvBool("MOCKGATEHUB_ENFORCE_AUTHENTICATION", true),
		ValidCredentials:      parseCredentials(getEnv("MOCKGATEHUB_VALID_CREDENTIALS", "local-test-app-id:local-test-app-secret")),
		DefaultOrganizationID: getEnv("DEFAULT_ORGANIZATION_ID", "default-org"),
		PublicBaseURL:         strings.TrimRight(getEnv("MOCKGATEHUB_PUBLIC_BASE_URL", "https://mockgatehub.interledger.test"), "/"),
		CardDataTokenSecret:   getEnv("MOCKGATEHUB_CARD_DATA_TOKEN_SECRET", ""),
	}
	return applyDefaults(cfg)
}

// applyDefaults handles post-load derivations shared by both loading paths.
func applyDefaults(cfg *Config) *Config {
	if cfg.CardDataTokenSecret == "" {
		cfg.CardDataTokenSecret = randomSecret(32)
	}

	// Clamp to 2-second minimum to prevent near-zero delays.
	if cfg.WebhookMinDelaySec < 2 {
		cfg.WebhookMinDelaySec = 2
	}

	cfg.UseRedis = cfg.RedisURL != ""

	return cfg
}

// splitFiles splits the CONFIG env var value into file paths.
func splitFiles(v string) []string {
	var files []string
	for _, f := range strings.Split(v, ",") {
		if f = strings.TrimSpace(f); f != "" {
			files = append(files, f)
		}
	}
	return files
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvFloat(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
			return floatVal
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		return val == "true" || val == "1" || val == "yes"
	}
	return defaultVal
}

func parseCredentials(credStr string) map[string]string {
	creds := make(map[string]string)
	if credStr == "" {
		return creds
	}
	for _, pair := range splitString(credStr, ',') {
		parts := splitString(pair, ':')
		if len(parts) == 2 {
			creds[parts[0]] = parts[1]
		}
	}
	return creds
}

func randomSecret(nBytes int) string {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "mockgatehub-rng-unavailable"
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func splitString(s string, delim byte) []string {
	var result []string
	var current []byte
	for i := 0; i < len(s); i++ {
		if s[i] == delim {
			result = append(result, string(current))
			current = []byte{}
		} else {
			current = append(current, s[i])
		}
	}
	if len(current) > 0 {
		result = append(result, string(current))
	}
	return result
}
