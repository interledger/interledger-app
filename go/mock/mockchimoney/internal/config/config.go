package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	Port     string
	LogLevel string
}

// Load reads configuration from environment variables
func Load() *Config {
	cfg := &Config{
		Port:     getEnv("MOCKCHIMONEY_PORT", "41800"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
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
