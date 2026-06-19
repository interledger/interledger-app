package config

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/interledger/interledger-app/go/configa"
)

// Config holds mockpti configuration.
type Config struct {
	Port               string
	LogLevel           string
	RedisURL           string
	RedisDB            string
	ClientID           string
	WebhookURL         string
	WebhookSigningKey  string // decoded PEM, populated from MOCKPTI_WEBHOOK_SIGNING_KEY_B64 or yaml webhook_signing_key
	FormsMutationToken string
}

// yamlConfig is the YAML-tagged struct used when CONFIG is set.
// webhook_signing_key stores the PEM directly (no base64 wrapping).
type yamlConfig struct {
	Port               string `yaml:"port"`
	LogLevel           string `yaml:"log_level"`
	RedisURL           string `yaml:"redis_url"`
	RedisDB            string `yaml:"redis_db"`
	ClientID           string `yaml:"client_id"`
	WebhookURL         string `yaml:"webhook_url"`
	WebhookSigningKey  string `yaml:"webhook_signing_key" validate:"required"`
	FormsMutationToken string `yaml:"forms_mutation_token"`
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
		fmt.Fprintln(os.Stderr, "fatal: parse mockpti config:", err)
		os.Exit(1)
	}
	y, err := parsed.Resolve(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal: resolve mockpti config:", err)
		os.Exit(1)
	}

	port := y.Port
	if port == "" {
		port = "8080"
	}
	logLevel := y.LogLevel
	if logLevel == "" {
		logLevel = "info"
	}
	redisDB := y.RedisDB
	if redisDB == "" {
		redisDB = "0"
	}
	clientID := y.ClientID
	if clientID == "" {
		clientID = "test-client-id"
	}

	return &Config{
		Port:               port,
		LogLevel:           logLevel,
		RedisURL:           y.RedisURL,
		RedisDB:            redisDB,
		ClientID:           clientID,
		WebhookURL:         y.WebhookURL,
		WebhookSigningKey:  y.WebhookSigningKey,
		FormsMutationToken: y.FormsMutationToken,
	}
}

func loadFromEnv() *Config {
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

// IsFileMode reports whether the CONFIG env var is set, i.e. whether configa
// is being used. main.go uses this to skip the env-var-based key validation.
func IsFileMode() bool {
	return os.Getenv("CONFIG") != ""
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
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
