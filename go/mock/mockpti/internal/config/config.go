package config

import (
	"encoding/base64"
	"os"
)

// Config holds mockpti configuration from environment variables.
type Config struct {
	Port               string
	LogLevel           string
	RedisURL           string
	RedisDB            string
	ClientID           string
	WebhookURL         string
	WebhookSigningKey  string // decoded PEM, populated from MOCKPTI_WEBHOOK_SIGNING_KEY_B64
	FormsMutationToken string
}

// Load reads configuration from environment variables with sensible defaults.
// MOCKPTI_WEBHOOK_SIGNING_KEY_B64 must be a valid base64-encoded Ed25519 private key PEM;
// startup validation (refusing to boot when absent or invalid) is enforced by runServer.
func Load() *Config {
	var signingKey string
	if b64 := os.Getenv("MOCKPTI_WEBHOOK_SIGNING_KEY_B64"); b64 != "" {
		if decoded, err := base64.StdEncoding.DecodeString(b64); err == nil {
			signingKey = string(decoded)
		}
	}

	return &Config{
		Port:               getEnv("MOCKPTI_PORT", "8080"),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		RedisURL:           os.Getenv("MOCKPTI_REDIS_URL"),
		RedisDB:            getEnv("MOCKPTI_REDIS_DB", "0"),
		ClientID:           getEnv("MOCKPTI_CLIENT_ID", "test-client-id"),
		WebhookURL:         os.Getenv("MOCKPTI_WEBHOOK_URL"),
		WebhookSigningKey:  signingKey,
		FormsMutationToken: os.Getenv("MOCKPTI_FORMS_MUTATION_TOKEN"),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
