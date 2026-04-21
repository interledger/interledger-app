package config

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"strconv"
	"strings"
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

// Load reads configuration from environment variables
func Load() *Config {
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

	// If no card-data token secret was supplied, generate a random one so
	// mock deployments never ship with a known signing key. If randomness
	// is unavailable (should never happen) fall back to a process-unique
	// value so mockgatehub still boots.
	if cfg.CardDataTokenSecret == "" {
		cfg.CardDataTokenSecret = randomSecret(32)
	}

	// Validate WebhookMinDelaySec: enforce minimum of 2 seconds to prevent
	// negative values or zero from bypassing intended minimum delay behavior.
	// Clamp to 2-second minimum if env var is misconfigured.
	if cfg.WebhookMinDelaySec < 2 {
		cfg.WebhookMinDelaySec = 2
	}

	// Use Redis if URL is provided
	cfg.UseRedis = cfg.RedisURL != ""

	return cfg
}

// getEnv gets environment variable with fallback
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// getEnvInt gets integer environment variable with fallback
func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

// getEnvFloat gets float64 environment variable with fallback
func getEnvFloat(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
			return floatVal
		}
	}
	return defaultVal
}

// getEnvBool gets boolean environment variable with fallback
func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		return val == "true" || val == "1" || val == "yes"
	}
	return defaultVal
}

// parseCredentials parses credentials from format: "appId1:secret1,appId2:secret2"
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

// randomSecret returns a base64-encoded cryptographically random byte string
// of the requested size. Falls back to a deterministic process-local value
// only if the system RNG fails, which should never happen on Linux.
func randomSecret(nBytes int) string {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "mockgatehub-rng-unavailable"
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// splitString splits a string by delimiter (helper for parsing)
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
