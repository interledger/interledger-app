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
<<<<<<< HEAD
Includes Rafiki setup/seeding and verification tools.`,
	}

	rootCmd.AddCommand(NewRafikiCmd())
	rootCmd.AddCommand(NewTotpCmd())
	rootCmd.AddCommand(NewVerifyCmd())
=======
Includes Rafiki setup/seeding and wallet user management.`,
=======
Includes Rafiki setup/seeding and verification tools.`,
>>>>>>> 3d5d696bf (checkpoint)
	}

	rootCmd.AddCommand(NewRafikiCmd())
	rootCmd.AddCommand(NewTotpCmd())
<<<<<<< HEAD
>>>>>>> 2cb285c8b (cli tools for totp and users)
=======
	rootCmd.AddCommand(NewVerifyCmd())
>>>>>>> 3d5d696bf (checkpoint)

	return rootCmd
}
