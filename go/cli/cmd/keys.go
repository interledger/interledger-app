package cmd

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func NewKeysCmd(b Backends) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Create key pair to sign requests.",
		Example: `
		fynbos keys create
		`,
	}

	cmd.AddCommand(NewCreateKeysCmd(b))

	return cmd
}

func NewCreateKeysCmd(b Backends) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creates a ed25519 key pair.",
		Example: `
		fynbos keys create
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return err
			}

			keyID, err := cmd.Flags().GetString("keyID")
			if err != nil {
				return err
			}

			if keyID == "" {
				err = survey.AskOne(
					&survey.Input{
						Message: "Give this key an alias.",
						Suggest: func(toComplete string) []string {
							return []string{uuid.NewString()}
						},
					},
					&keyID,
				)
				if err != nil {
					return err
				}
			}

			return CreateKeyPair(context.Background(), b, keyID, force)
		},
	}

	_ = cmd.Flags().StringP("keyID", "k", "", "ID to identify key")
	_ = cmd.Flags().BoolP("force", "f", false, "Overwrite existing key")

	return cmd
}

func CreateKeyPair(ctx context.Context, b Backends, keyID string, force bool) error {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}

	privPem := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	})

	keyFilePath := b.Config().GetString("clientKeyPath")
	if keyFilePath == "" {
		return fmt.Errorf("`clientKeyPath` not set in config.")
	}

	f, err := os.Stat(keyFilePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if f != nil && !force {
		return fmt.Errorf("A key already exists. Use the `-f` flag to overwrite.")
	}

	err = os.WriteFile(keyFilePath, privPem, 0600)
	if err != nil {
		return nil
	}

	b.Config().Set("clientKeyID", keyID)
	err = b.Config().WriteConfig()
	if err != nil {
		return err
	}

	pkixEncodedPubKey, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return err
	}

	pemEncodedPubKey := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pkixEncodedPubKey,
	})

	fmt.Println("You can connect this cli to your wallet at https://interleddger.app/connections/add-a-public-key")
	fmt.Println(string(pemEncodedPubKey))

	fingerprint := sha256.Sum256(pemEncodedPubKey)
	fmt.Println("The key fingerprint is:")
	fmt.Println("SHA256:", string(hex.EncodeToString(fingerprint[:])))

	return nil
}
