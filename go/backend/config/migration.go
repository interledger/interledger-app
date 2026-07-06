package config

import (
	"context"
	"fmt"

	"github.com/interledger/interledger-app/go/configa"
)

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
