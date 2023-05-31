package cmd

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
	"gitlab.com/fynbos/cli/identities"
)

type VerifyBackends interface {
	Validator() *validator.Validate
	Identities() identities.Client
}

func NewVerifyCmd(b VerifyBackends) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "verify [platform] [identifier] [wallet address]",
		Short:   "Verify an identity",
		Example: "fynbos verify twitter @fynbosdev https://fynbos.me/money",
		RunE: func(cmd *cobra.Command, args []string) error {
			commandArgs := &VerifyCommandArgs{
				Type:          identities.Platform(args[0]),
				Identifier:    args[1],
				WalletAddress: args[2],
			}

			err := b.Validator().Struct(commandArgs)
			if err != nil {
				return fmt.Errorf("invalid arguments: %w", err)
			}

			err = b.Identities().Verify(context.Background(), &identities.VerifyArgs{
				Type:          commandArgs.Type,
				Identifier:    commandArgs.Identifier,
				WalletAddress: commandArgs.WalletAddress,
			})
			if err != nil {
				return fmt.Errorf("failed to verify identity: %w", err)
			}

			return nil
		},
	}

	return cmd
}
