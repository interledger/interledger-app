package cli

import (
	"github.com/interledger/interledger-app/go/backend/config"
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

	return &StartArgs{cfg}, nil
}
