package cmd

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "local-dev-tool",
		Short: "Local development environment management tool",
		Long: `A TUI tool to manage the Interledger local development environment.
		Long: `A TUI tool to manage the Interledger local development environment.
Includes Rafiki setup/seeding and verification tools.

Defaults:
  KRATOS_DATABASE_URL=postgres://postgres:postgres@localhost:5432/kratos?sslmode=disable
  KRATOS_ADMIN_URL=http://localhost:4434`,
	}

	rootCmd.AddCommand(NewRafikiCmd())
	rootCmd.AddCommand(NewTotpCmd())
	rootCmd.AddCommand(NewVerifyCmd())

	return rootCmd
}
