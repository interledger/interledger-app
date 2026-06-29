package config

import (
	"context"
	"fmt"

	"github.com/interledger/interledger-app/go/configa"
)

// MigrationConfig is the typed configuration for the backend migrate command.
// It is loaded from YAML files listed in the CONFIG environment variable.
type MigrationConfig struct {
	DBUrl               string       `yaml:"db_url"                validate:"required"`
	PacioliDBUrl        string       `yaml:"pacioli_db_url"        validate:"required"`
	OpenPaymentsBaseURL string       `yaml:"open_payments_base_url"`
	KratosUrl           string       `yaml:"kratos_url"`
	LogLevel            string       `yaml:"log_level"`
	LogOutputPath       string       `yaml:"log_output_path"`
	// Label is the telemetry tag attached to monitoring signals (Sentry, etc.).
	Label  string       `yaml:"label"`
	Sentry SentryConfig `yaml:"sentry"`
}

// LoadMigration parses the YAML files listed in the CONFIG environment variable
// and returns a validated MigrationConfig.
func LoadMigration() (*MigrationConfig, error) {
	files, err := configFiles()
	if err != nil {
		return nil, err
	}

	secretClient := configa.NewInClusterSecretClient()
	parsed, err := configa.Parse[MigrationConfig](files, configa.WithSecretClient(secretClient))
	if err != nil {
		return nil, fmt.Errorf("parse migration config: %w", err)
	}

	cfg, err := parsed.Resolve(context.Background())
	if err != nil {
		return nil, fmt.Errorf("resolve migration config: %w", err)
	}

	applyMigrationDefaults(&cfg)
	return &cfg, nil
}

func applyMigrationDefaults(cfg *MigrationConfig) {
	if cfg.KratosUrl == "" {
		cfg.KratosUrl = "http://localhost:4433"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.LogOutputPath == "" {
		cfg.LogOutputPath = "stderr"
	}
}
