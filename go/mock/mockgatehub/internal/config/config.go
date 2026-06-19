package config

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/interledger/interledger-app/go/configa"
)

// Config holds application configuration.
// YAML tags are used directly so that configa can parse into this struct.
// UseRedis is derived from RedisURL after loading and is never read from YAML.
type Config struct {
	Port                  string            `yaml:"port"`
	LogLevel              string            `yaml:"log_level"`
	RedisURL              string            `yaml:"redis_url"`
	RedisDB               int               `yaml:"redis_db"`
	WebhookURL            string            `yaml:"webhook_url"`
	WebhookSecret         string            `yaml:"webhook_secret"`
	WebhookMinDelaySec    float64           `yaml:"webhook_min_delay_sec"`
	EnforceAuthentication bool              `yaml:"enforce_authentication"`
	ValidCredentials      map[string]string `yaml:"valid_credentials"` // appID -> secret
	DefaultOrganizationID string            `yaml:"default_organization_id"`
	// PublicBaseURL is the externally reachable base URL of mockgatehub
	// (no trailing slash). It is used to build absolute URLs in API responses
	// that are followed directly by the browser (e.g. card-data tokenisation
	// links).
	PublicBaseURL string `yaml:"public_base_url"`
	// CardDataTokenSecret is the HMAC secret used to sign card-data JWTs
	// returned by POST /cards/v1/token/card-data. It must not be a hard-coded
	// constant: when unset we generate a random value at startup so mock
	// deployments never share a signing key across processes.
	CardDataTokenSecret string `yaml:"card_data_token_secret"`
	UseRedis            bool   `yaml:"-"`
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
		fmt.Fprintln(os.Stderr, "fatal: parse mockgatehub config:", err)
		os.Exit(1)
	}
	cfg, err := parsed.Resolve(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal: resolve mockgatehub config:", err)
		os.Exit(1)
	}
	return applyDefaults(&cfg)
}

// applyDefaults handles post-load derivations shared by both loading paths.
func applyDefaults(cfg *Config) *Config {
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.CardDataTokenSecret == "" {
		cfg.CardDataTokenSecret = randomSecret(32)
	}
	// Clamp to 2-second minimum to prevent near-zero delays.
	if cfg.WebhookMinDelaySec < 2 {
		cfg.WebhookMinDelaySec = 2
	}
	cfg.PublicBaseURL = strings.TrimRight(cfg.PublicBaseURL, "/")
	cfg.UseRedis = cfg.RedisURL != ""
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

func randomSecret(nBytes int) string {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "mockgatehub-rng-unavailable"
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
