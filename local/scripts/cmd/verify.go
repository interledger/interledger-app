package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"local-dev-tool/internal/verify"

	"github.com/spf13/cobra"
)

func NewVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "verify [email]",
		Short:   "Manually verify a user's email address",
		Long:    "Directly updates the Kratos database to mark the user's email as verified. Useful for local development to bypass email verification.",
		Args:    cobra.ExactArgs(1),
		Example: "  local-dev-tool verify user@example.com",
		RunE: func(cmd *cobra.Command, args []string) error {
			email := strings.TrimSpace(args[0])
			if email == "" {
				return fmt.Errorf("email is required")
			}

			client := verify.NewClient()
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			if err := client.VerifyEmail(ctx, email); err != nil {
				return err
			}

			fmt.Printf("✅ Email verified successfully: %s\n", email)
			fmt.Println("\nThe user can now log in without email verification.")

			return nil
		},
	}

	return cmd
}
