package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/interledger/interledger-app/go/configa"
)

// Config holds mockxago configuration.
// YAML tags are used directly so that configa can parse into this struct.
type Config struct {
	Port                string  `yaml:"port"`
	LogLevel            string  `yaml:"log_level"`
	RedisURL            string  `yaml:"redis_url"`
	RedisDB             int     `yaml:"redis_db"`
	PublicKey           string  `yaml:"public_key"`
	Secret              string  `yaml:"secret"`
	TestMode            bool    `yaml:"test_mode"`
	WebhookURL          string  `yaml:"webhook_url"`
	WebhookSecret       string  `yaml:"webhook_secret"`
	WebhookMinDelaySec  float64 `yaml:"webhook_min_delay_sec"`
	PersonaWebhookURL   string  `yaml:"persona_webhook_url"`
	PersonaWebhookToken string  `yaml:"persona_webhook_token"`
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
		fmt.Fprintln(os.Stderr, "fatal: parse mockxago config:", err)
		os.Exit(1)
	}
	cfg, err := parsed.Resolve(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal: resolve mockxago config:", err)
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
	if cfg.PublicKey == "" {
		cfg.PublicKey = "test-public-key"
	}
	if cfg.Secret == "" {
		cfg.Secret = "test-secret"
	}
	if cfg.PersonaWebhookURL == "" {
		cfg.PersonaWebhookURL = "http://backend:8080/webhooks/persona"
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
