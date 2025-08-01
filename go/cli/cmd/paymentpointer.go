package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func NewPaymentPointerCmd(b Backends) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [payment-pointer]",
		Short: "View details about the payment pointer",
		Example: `
fynbos get https://ilp.link/protea
fynbos get https://ilp.link/791f09c0-c6a2-4a27-8e05-6f7ae37a8a28
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return getPaymentPointer(cmd.Context(), b, args[0])
		},
	}

	return cmd
}

func getPaymentPointer(ctx context.Context, b Backends, url string) error {
	resp, err := b.HttpClient().Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var prettyOP bytes.Buffer
	err = json.Indent(&prettyOP, body, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(prettyOP.String())

	return nil
}
