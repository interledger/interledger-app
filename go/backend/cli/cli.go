package cli

import (
	"github.com/interledger/interledger-app/go/backend/config"
	"github.com/interledger/interledger-app/go/env"
)

// StartArgs wraps the parsed and validated backend start/worker configuration.
type StartArgs struct {
	*config.StartConfig
}

// MigrationArgs wraps the parsed and validated migration configuration.
type MigrationArgs struct {
	*config.MigrationConfig
}

func ParseMigrationArgs() (*MigrationArgs, error) {
	cfg, err := config.LoadMigration()
	if err != nil {
		return nil, err
	}
	return &MigrationArgs{cfg}, nil
}

func ParseStartArgs() (*StartArgs, error) {
	cfg, err := config.LoadStart()
	if err != nil {
		return nil, err
	}

	// Feed global env package state. LoadStart() has already validated the config.
	env.SetIlwEnv(cfg.Environment.Mode)
	if len(cfg.AllowedWalletIDs) > 0 {
		env.SetAllowedWalletIDs(cfg.AllowedWalletIDs)
	}
	if len(cfg.BlockedRegions) > 0 {
		env.SetBlockedRegions(cfg.BlockedRegions)
	}
	if cfg.OpenPaymentsBaseURL != "" {
		env.SetOpenPaymentsURL(cfg.OpenPaymentsBaseURL)
	}
	if cfg.AuthBaseURL != "" {
		env.SetAuthURL(cfg.AuthBaseURL)
	}
	env.SetApplicationURL(cfg.ApplicationURL)
	env.SetRafikiNodeEnabled(cfg.Rafiki.NodeEnabled)

	return &StartArgs{cfg}, nil
}
