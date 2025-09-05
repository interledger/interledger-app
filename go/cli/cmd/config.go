package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var validConfigKeys = []string{"wallet", "clientKeyPath", "clientKeyID"}

func NewConfigCmd(b Backends) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config [command]",
		Short: "Manage config for the Fynbos CLI.",
		Example: `
		fynbos config set wallet https://ilp.link/protea
		fynbos config get wallet
		fynbos config list
		`,
	}

	cmd.AddCommand(NewSetConfigCmd(b), NewGetConfigCmd(b), NewListConfigCmd(b))

	return cmd
}

func NewSetConfigCmd(b Backends) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set config for the Fynbos CLI.",
		Example: `
		fynbos config set wallet https://ilp.link/protea
		`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			var validKey bool
			for _, k := range validConfigKeys {
				if k == key {
					validKey = true
					break
				}
			}

			if !validKey {
				return fmt.Errorf("Valid config variables are %+v", validConfigKeys)
			}

			b.Config().Set(key, value)
			err := b.Config().WriteConfig()
			if err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

func NewGetConfigCmd(b Backends) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [key]",
		Short: "Get specified config variable.",
		Example: `
		fynbos config get wallet
		`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			var validKey bool
			for _, k := range validConfigKeys {
				if k == key {
					validKey = true
					break
				}
			}

			if !validKey {
				return fmt.Errorf("Valid config variables are %+v", validConfigKeys)
			}

			fmt.Printf("%s=%s\n", key, b.Config().GetString(key))

			return nil
		},
	}

	return cmd
}

type config struct {
	ClientKeyPath string `json:"clientKeyPath"`
	ClientKeyID   string `json:"clientKeyID"`
	Wallet        string `json:"wallet"`
}

func NewListConfigCmd(b Backends) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get specified config variable.",
		Example: `
		fynbos config get wallet
		`,
		Args: cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := config{}
			err := b.Config().Unmarshal(&c)
			if err != nil {
				return err
			}

			jsonConfig, err := json.Marshal(c)
			if err != nil {
				return err
			}

			fmt.Println(string(jsonConfig))
			return nil
		},
	}

	return cmd
}
