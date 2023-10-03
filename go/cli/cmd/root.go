package cmd

import (
	"github.com/spf13/cobra"
)

func NewCmdRoot(b Backends) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fynbos",
		Short:   "Fynbos cli",
		Long:    "Access your wallet using the open-payments API.",
		Version: "0.0.1",
	}

	cmd.AddCommand(NewPayCmd(b), NewKeysCmd(b), NewConfigCmd(b), NewPaymentPointerCmd(b))

	return cmd
}
