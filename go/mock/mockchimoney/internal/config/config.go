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
	WebhookURL            string
	WebhookSecret         string
	WebhookMinDelaySec    float64
	InteracFeeFlat        float64
	CADToUSDRate          float64
	PublicBaseURL         string
}

// Load reads configuration from environment variables
func Load() *Config {
	cfg := &Config{
		Port:                  getEnv("MOCKCHIMONEY_PORT", "41800"),
		LogLevel:              getEnv("LOG_LEVEL", "info"),
		APIKey:                getEnv("MOCKCHIMONEY_API_KEY", "local-test-api-key"),
		EnforceAuthentication: getEnvAsBool("MOCKCHIMONEY_ENFORCE_AUTHENTICATION", true),
		WebhookURL:            getEnv("WEBHOOK_URL", "http://backend:8080/webhooks/chimoney"),
		WebhookSecret:         getEnv("CHIMONEY_WEBHOOK_SECRET", "local_bG9jYWwtdGVzdC13ZWJob29rLXNlY3JldA=="),
		WebhookMinDelaySec:    getEnvAsFloat("WEBHOOK_MIN_DELAY_SEC", 0.5),
		InteracFeeFlat:        getEnvAsFloat("INTERAC_FEE_FLAT", 1.50),
		CADToUSDRate:          getEnvAsFloat("CAD_TO_USD_RATE", 0.735),
		PublicBaseURL:         getEnv("MOCKCHIMONEY_PUBLIC_BASE_URL", "https://mockchimoney.interledger.test"),
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

func getEnvAsFloat(key string, defaultVal float64) float64 {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}

	parsed, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return defaultVal
	}

	return parsed
}
