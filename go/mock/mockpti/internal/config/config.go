package config

import "os"

// Config holds mockpti configuration from environment variables.
type Config struct {
	Port                 string
	LogLevel             string
	RedisURL             string
	RedisDB              string
	ClientID             string
	WebhookURL           string
	WebhookSigningJWK    string
	WebhookEncryptionJWK string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Port:                 getEnv("MOCKPTI_PORT", "8080"),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		RedisURL:             os.Getenv("MOCKPTI_REDIS_URL"),
		RedisDB:              getEnv("MOCKPTI_REDIS_DB", "0"),
		ClientID:             getEnv("MOCKPTI_CLIENT_ID", "test-client-id"),
		WebhookURL:           os.Getenv("MOCKPTI_WEBHOOK_URL"),
		WebhookSigningJWK:    os.Getenv("MOCKPTI_WEBHOOK_SIGNING_JWK"),
		WebhookEncryptionJWK: os.Getenv("MOCKPTI_WEBHOOK_ENCRYPTION_JWK"),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
