package config

import "os"

// Config holds mockplaid configuration from environment variables.
type Config struct {
	Port     string
	LogLevel string
	RedisURL string
	RedisDB  string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Port:     getEnv("MOCKPLAID_PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		RedisURL: os.Getenv("MOCKPLAID_REDIS_URL"),
		RedisDB:  getEnv("MOCKPLAID_REDIS_DB", "6"),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
