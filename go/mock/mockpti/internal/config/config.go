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
// YAML tags are used directly so that configa can parse into this struct.
// WebhookSigningKey stores the PEM directly when loaded from YAML (no base64 wrapping).
type Config struct {
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
	parsed, err := configa.Parse[Config](files)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal: parse mockpti config:", err)
		os.Exit(1)
	}
	cfg, err := parsed.Resolve(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal: resolve mockpti config:", err)
		os.Exit(1)
	}
	return applyDefaults(&cfg)
}

func loadFromEnv() *Config {
	var signingKey string
	if b64 := os.Getenv("MOCKPTI_WEBHOOK_SIGNING_KEY_B64"); b64 != "" {
		if decoded, err := base64.StdEncoding.DecodeString(b64); err == nil {
			signingKey = string(decoded)
		}
	}

	return applyDefaults(&Config{
		Port:               os.Getenv("MOCKPTI_PORT"),
		LogLevel:           os.Getenv("LOG_LEVEL"),
		RedisURL:           os.Getenv("MOCKPTI_REDIS_URL"),
		RedisDB:            os.Getenv("MOCKPTI_REDIS_DB"),
		ClientID:           os.Getenv("MOCKPTI_CLIENT_ID"),
		WebhookURL:         os.Getenv("MOCKPTI_WEBHOOK_URL"),
		WebhookSigningKey:  signingKey,
		FormsMutationToken: os.Getenv("MOCKPTI_FORMS_MUTATION_TOKEN"),
	})
}

// applyDefaults handles post-load derivations shared by both loading paths.
func applyDefaults(cfg *Config) *Config {
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.RedisDB == "" {
		cfg.RedisDB = "0"
	}
	if cfg.ClientID == "" {
		cfg.ClientID = "test-client-id"
	}
	return cfg
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
