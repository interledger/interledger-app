package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
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

func loadFromEnv() *Config {
	redisDB := 0
	if s := os.Getenv("MOCKXAGO_REDIS_DB"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			redisDB = v
		}
	}
	webhookMinDelaySec := 0.0
	if s := os.Getenv("WEBHOOK_MIN_DELAY_SEC"); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			webhookMinDelaySec = v
		}
	}

	return applyDefaults(&Config{
		Port:                os.Getenv("XAGO_MOCK_PORT"),
		LogLevel:            os.Getenv("LOG_LEVEL"),
		RedisURL:            os.Getenv("MOCKXAGO_REDIS_URL"),
		RedisDB:             redisDB,
		PublicKey:           os.Getenv("XAGO_API_PUBLIC_KEY"),
		Secret:              os.Getenv("XAGO_API_SECRET"),
		TestMode:            strings.EqualFold(os.Getenv("XAGO_MOCK_TEST_MODE"), "true"),
		WebhookURL:          os.Getenv("WEBHOOK_URL"),
		WebhookSecret:       os.Getenv("WEBHOOK_SECRET"),
		WebhookMinDelaySec:  webhookMinDelaySec,
		PersonaWebhookURL:   os.Getenv("PERSONA_WEBHOOK_URL"),
		PersonaWebhookToken: os.Getenv("PERSONA_WEBHOOK_TOKEN"),
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
