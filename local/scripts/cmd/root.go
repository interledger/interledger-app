package cmd

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "local-dev-tool",
		Short: "Local development environment management tool",
		Long: `A TUI tool to manage the Interledger local development environment.
<<<<<<< HEAD
Includes Rafiki setup/seeding and verification tools.`,
	}

	rootCmd.AddCommand(NewRafikiCmd())
	rootCmd.AddCommand(NewTotpCmd())
	rootCmd.AddCommand(NewVerifyCmd())
=======
Includes Rafiki setup/seeding and wallet user management.`,
	}

	rootCmd.AddCommand(NewRafikiCmd())
	rootCmd.AddCommand(NewUserCmd())
	rootCmd.AddCommand(NewTotpCmd())
>>>>>>> 2cb285c8b (cli tools for totp and users)

	return rootCmd
}
