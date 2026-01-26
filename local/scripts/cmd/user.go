package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"local-dev-tool/internal/user"

	"github.com/spf13/cobra"
)

func NewUserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Wallet user management",
		Long:  `Create and manage wallet users for local development.`,
	}

	cmd.AddCommand(NewUserCreateCmd())
	cmd.AddCommand(NewUserListCmd())

	return cmd
}

func NewUserCreateCmd() *cobra.Command {
	var opts user.CreateUserOptions

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new wallet user",
		Long: `Create a new wallet user with email verification, KYC completion, and Gatehub activation.
The user will be fully activated and ready to use.`,
		Example: `  local-dev-tool user create --email user@example.com --first-name John --last-name Doe --country DE
  local-dev-tool user create -e alice@test.com -f Alice -l Smith -c FR --password MyPassword123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate required fields
			if opts.Email == "" {
				return fmt.Errorf("email is required")
			}
			if opts.FirstName == "" {
				return fmt.Errorf("first name is required")
			}
			if opts.LastName == "" {
				return fmt.Errorf("last name is required")
			}
			if opts.CountryCode == "" {
				return fmt.Errorf("country code is required")
			}

			// Set default password if not provided
			if opts.Password == "" {
				opts.Password = "Test@123456?"
				fmt.Println("Using default password: Test@123456?")
			}

			// Validate country code (basic check - 2 letters)
			if len(opts.CountryCode) != 2 {
				return fmt.Errorf("country code must be 2 letters (ISO 3166-1 alpha-2)")
			}

			opts.CountryCode = strings.ToUpper(opts.CountryCode)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			fmt.Printf("\n📝 Creating user: %s %s (%s)\n", opts.FirstName, opts.LastName, opts.Email)
			fmt.Printf("🌍 Country: %s\n", opts.CountryCode)
			fmt.Println()

			userClient := user.NewClient()
			result, err := userClient.CreateUser(ctx, opts)
			if err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}

			fmt.Println("\n✅ User created successfully!")
			fmt.Println("\n═══════════════════════════════════════")
			fmt.Printf("Identity ID:  %s\n", result.IdentityID)
			fmt.Printf("Email:        %s\n", result.Email)
			fmt.Printf("Password:     %s\n", opts.Password)
			fmt.Printf("Wallet ID:    %s\n", result.WalletID)
			fmt.Printf("KYC Status:   %s\n", result.KYCStatus)
			fmt.Println("═══════════════════════════════════════")
			fmt.Println("\nYou can now log in with the email and password above.")

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.Email, "email", "e", "", "User email address (required)")
	cmd.Flags().StringVarP(&opts.FirstName, "first-name", "f", "", "First name (required)")
	cmd.Flags().StringVarP(&opts.LastName, "last-name", "l", "", "Last name (required)")
	cmd.Flags().StringVarP(&opts.CountryCode, "country", "c", "", "Country code (2 letters, e.g., US, DE, FR) (required)")
	cmd.Flags().StringVarP(&opts.Password, "password", "p", "", "Password (default: Test@123456?)")

	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("first-name")
	cmd.MarkFlagRequired("last-name")
	cmd.MarkFlagRequired("country")

	return cmd
}

func NewUserListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all wallet users",
		Long:    `List all users from Kratos identity system with their details.`,
		Example: `  local-dev-tool user list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := user.NewClient()
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			users, err := client.ListUsers(ctx)
			if err != nil {
				return fmt.Errorf("failed to list users: %w", err)
			}

			if len(users) == 0 {
				fmt.Println("No users found.")
				return nil
			}

			fmt.Printf("Found %d user(s):\n\n", len(users))
			fmt.Printf("%-36s  %-30s  %-20s  %-10s\n", "ID", "EMAIL", "NAME", "STATE")
			fmt.Println(strings.Repeat("-", 100))
			for _, u := range users {
				name := fmt.Sprintf("%s %s", u.FirstName, u.LastName)
				fmt.Printf("%-36s  %-30s  %-20s  %-10s\n", u.ID, u.Email, name, u.State)
			}

			return nil
		},
	}

	return cmd
}
