package config

import (
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	Port                  string
	LogLevel              string
	APIKey                string
	EnforceAuthentication bool
}

// Load reads configuration from environment variables
func Load() *Config {
	cfg := &Config{
		Port:                  getEnv("MOCKCHIMONEY_PORT", "41800"),
		LogLevel:              getEnv("LOG_LEVEL", "info"),
		APIKey:                getEnv("MOCKCHIMONEY_API_KEY", "local-test-api-key"),
		EnforceAuthentication: getEnvAsBool("MOCKCHIMONEY_ENFORCE_AUTHENTICATION", true),
	}

	return cfg
}

// getEnv gets environment variable with fallback
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}

	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return defaultVal
	}

	return parsed
}
