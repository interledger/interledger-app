package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"local-dev-tool/internal/totp"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func NewTotpCmd() *cobra.Command {
	var secret string

	cmd := &cobra.Command{
		Use:   "totp [email]",
		Short: "Generate a TOTP code for a user",
		Long:  `Generate a Time-based One-Time Password (TOTP) code for a user by email address.`,
		Args:  cobra.ExactArgs(1),
		Example: `  local-dev-tool totp user@example.com
  local-dev-tool totp alice@test.com
  local-dev-tool totp alice@test.com --secret JBSWY3DPEHPK3PXP`,
		RunE: func(cmd *cobra.Command, args []string) error {
			email := args[0]
			client := totp.NewClient()
			ctx := context.Background()

			// If secret flag is provided, use it directly
			if secret != "" {
				// Store the secret
				if err := client.StoreSecret(ctx, email, secret); err != nil {
					return fmt.Errorf("failed to store secret: %w", err)
				}

				// Generate code with the secret
				code, err := client.GenerateCode(ctx, email)
				if err != nil {
					return fmt.Errorf("failed to generate TOTP code: %w", err)
				}

				fmt.Println(code)
				return nil
			}

			code, err := client.GenerateCode(ctx, email)
			if err != nil {
				if errors.Is(err, totp.ErrTOTPNotConfigured) {
					// TOTP not configured - guide user to set it up
					fmt.Println("\n⚠️  TOTP is not configured for this user.")
					fmt.Println("\n📱 To set up TOTP:")
					fmt.Println("   1. Open the wallet application and log in")
					fmt.Println("   2. Go to Security Settings")
					fmt.Println("   3. Enable Two-Factor Authentication (TOTP)")
					fmt.Println("   4. Copy the secret shown on screen")
					fmt.Println("\n🔐 Enter the TOTP secret from the wallet application:")
					fmt.Print("Secret: ")

					// Read secret from stdin
					reader := bufio.NewReader(os.Stdin)
					secret, err := reader.ReadString('\n')
					if err != nil {
						return fmt.Errorf("failed to read secret: %w", err)
					}
					secret = strings.TrimSpace(secret)

					if secret == "" {
						return fmt.Errorf("secret cannot be empty")
					}

					// Store the secret
					if err := client.StoreSecret(ctx, email, secret); err != nil {
						return fmt.Errorf("failed to store secret: %w", err)
					}

					// Generate code with the new secret
					code, err = client.GenerateCode(ctx, email)
					if err != nil {
						return fmt.Errorf("failed to generate TOTP code: %w", err)
					}

					fmt.Println("\n✅ TOTP secret stored successfully!")
					fmt.Println("\n📋 Current TOTP Code:")
					fmt.Printf("   %s\n", code)
					fmt.Println("   (This code expires in 30 seconds)")
					fmt.Println("\n💡 Next time, just run: ./local-dev-tool totp", email)
					return nil
				}
				return fmt.Errorf("failed to generate TOTP code: %w", err)
			}

			// TOTP configured, just show the code
			fmt.Println(code)
			return nil
		},
	}

	cmd.Flags().StringVar(&secret, "secret", "", "TOTP secret (base32 encoded)")

	return cmd
}
