package config

import (
	"context"
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

// Load reads configuration from YAML files specified in the CONFIG environment variable.
// CONFIG should be a comma-separated list of file paths; later files overlay earlier ones.
func Load() *Config {
	filesStr := os.Getenv("CONFIG")
	if filesStr == "" {
		fmt.Fprintln(os.Stderr, "fatal: CONFIG environment variable is required")
		os.Exit(1)
	}
	return loadFromFiles(splitFiles(filesStr))
}

func loadFromFiles(files []string) *Config {
	secretClient := configa.NewInClusterSecretClient()
	parsed, err := configa.Parse[Config](files, configa.WithSecretClient(secretClient))
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
