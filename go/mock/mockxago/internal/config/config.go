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
type Config struct {
	Port                string
	LogLevel            string
	RedisURL            string
	RedisDB             int
	PublicKey           string
	Secret              string
	TestMode            bool
	WebhookURL          string
	WebhookSecret       string
	WebhookMinDelaySec  float64
	PersonaWebhookURL   string
	PersonaWebhookToken string
}

// yamlConfig is the YAML-tagged struct used when CONFIG is set.
type yamlConfig struct {
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
	parsed, err := configa.Parse[yamlConfig](files)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal: parse mockxago config:", err)
		os.Exit(1)
	}
	y, err := parsed.Resolve(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal: resolve mockxago config:", err)
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
	publicKey := y.PublicKey
	if publicKey == "" {
		publicKey = "test-public-key"
	}
	secret := y.Secret
	if secret == "" {
		secret = "test-secret"
	}
	personaWebhookURL := y.PersonaWebhookURL
	if personaWebhookURL == "" {
		personaWebhookURL = "http://backend:8080/webhooks/persona"
	}

	return &Config{
		Port:                port,
		LogLevel:            logLevel,
		RedisURL:            y.RedisURL,
		RedisDB:             y.RedisDB,
		PublicKey:           publicKey,
		Secret:              secret,
		TestMode:            y.TestMode,
		WebhookURL:          y.WebhookURL,
		WebhookSecret:       y.WebhookSecret,
		WebhookMinDelaySec:  y.WebhookMinDelaySec,
		PersonaWebhookURL:   personaWebhookURL,
		PersonaWebhookToken: y.PersonaWebhookToken,
	}
}

func loadFromEnv() *Config {
	port := os.Getenv("XAGO_MOCK_PORT")
	if port == "" {
		port = "8080"
	}
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	publicKey := os.Getenv("XAGO_API_PUBLIC_KEY")
	if publicKey == "" {
		publicKey = "test-public-key"
	}
	secret := os.Getenv("XAGO_API_SECRET")
	if secret == "" {
		secret = "test-secret"
	}
	redisDB := 0
	if s := os.Getenv("MOCKXAGO_REDIS_DB"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			redisDB = v
		}
	}
	personaWebhookURL := os.Getenv("PERSONA_WEBHOOK_URL")
	if personaWebhookURL == "" {
		personaWebhookURL = "http://backend:8080/webhooks/persona"
	}
	webhookMinDelaySec := 0.0
	if s := os.Getenv("WEBHOOK_MIN_DELAY_SEC"); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			webhookMinDelaySec = v
		}
	}

	return &Config{
		Port:                port,
		LogLevel:            logLevel,
		RedisURL:            os.Getenv("MOCKXAGO_REDIS_URL"),
		RedisDB:             redisDB,
		PublicKey:           publicKey,
		Secret:              secret,
		TestMode:            strings.EqualFold(os.Getenv("XAGO_MOCK_TEST_MODE"), "true"),
		WebhookURL:          os.Getenv("WEBHOOK_URL"),
		WebhookSecret:       os.Getenv("WEBHOOK_SECRET"),
		WebhookMinDelaySec:  webhookMinDelaySec,
		PersonaWebhookURL:   personaWebhookURL,
		PersonaWebhookToken: os.Getenv("PERSONA_WEBHOOK_TOKEN"),
	}
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
