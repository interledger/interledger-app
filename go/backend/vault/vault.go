package vault

import (
	"context"
	"encoding/base64"
	"fmt"
	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/kubernetes"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
	"os"
	"strings"
)

type client struct {
	vc                *vault.Client
	transitEnginePath string
}

func NewClient() (Client, error) {

	// Only login to vault on non-local environments.
	if !env.IsLocal() {
		addr := os.Getenv("VAULT_ADDR")
		mountPath := os.Getenv("VAULT_AUTH_PATH")
		transitEnginePath := os.Getenv("VAULT_TRANSIT_ENGINE_PATH")

		config := vault.DefaultConfig()
		config.Address = addr
		vc, err := vault.NewClient(config)
		if err != nil {
			return nil, fmt.Errorf("unable to initialize Vault client: %w", err)
		}

		k8sAuth, err := auth.NewKubernetesAuth(
			"k8s-app",
			auth.WithMountPath(mountPath),
		)
		if err != nil {
			return nil, fmt.Errorf("unable to initialize Kubernetes auth method: %w", err)
		}

		log.Info("Attempting to auth vault")
		authInfo, err := vc.Auth().Login(context.Background(), k8sAuth)
		if err != nil {
			return nil, fmt.Errorf("unable to log in with Kubernetes auth: %w", err)
		}
		if authInfo == nil {
			return nil, fmt.Errorf("no auth info was returned after login")
		}

		log.Info("Successfully authed vault")
		return client{vc: vc, transitEnginePath: transitEnginePath}, nil
	}

	return client{vc: nil}, nil
}

// CreateKey Currently only supports ED25119 keys
func (c client) CreateKey(keyName string) error {
	keyPath := fmt.Sprintf("%s/keys/%s", c.transitEnginePath, keyName)

	data := map[string]interface{}{
		"type":                   "ed25519",
		"convergent_encryption":  false,
		"derived":                false,
		"exportable":             true,
		"allow_plaintext_backup": false,
		"deletion_allowed":       true,
	}

	_, err := c.vc.Logical().Write(keyPath, data)
	return err
}

func (c client) Sign(keyName string, input string) ([]byte, error) {
	encodedInput := base64.StdEncoding.EncodeToString([]byte(input))
	keyPath := fmt.Sprintf("%s/sign/%s", c.transitEnginePath, keyName)

	data := map[string]interface{}{
		"input": encodedInput,
	}

	resp, err := c.vc.Logical().Write(keyPath, data)
	if err != nil {
		return nil, err
	}

	signatureBase64Url := strings.TrimPrefix(resp.Data["signature"].(string), "vault:v1:") // remove "vault:v1:" prefix

	signature, err := base64.StdEncoding.DecodeString(signatureBase64Url)
	if err != nil {
		return nil, fmt.Errorf("unable to decode signature: %w", err)
	}

	return signature, nil
}

func (c client) Verify(keyName string, input VerifyInput) (bool, error) {
	keyPath := fmt.Sprintf("%s/verify/%s", c.transitEnginePath, keyName)
	encodedInput := base64.StdEncoding.EncodeToString([]byte(input.Input))

	data := map[string]interface{}{
		"input":     encodedInput,
		"signature": input.Signature,
	}

	resp, err := c.vc.Logical().Write(keyPath, data)
	if err != nil {
		return false, err
	}

	return resp.Data["valid"].(bool), nil
}

func (c client) GetPublicKey(keyName string) (string, error) {
	keyPath := fmt.Sprintf("%s/keys/%s", c.transitEnginePath, keyName)
	secret, err := c.vc.Logical().Read(keyPath)
	if err != nil {
		return "", err
	}

	if secret == nil {
		return "", fmt.Errorf("transit key %q not found", keyName)
	}

	keys, ok := secret.Data["keys"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("failed to extract public key from transit key %q", keyName)
	}

	pubKey := ""
	for _, keyData := range keys {
		keyDataMap := keyData.(map[string]interface{})
		key, ok := keyDataMap["public_key"].(string)
		if ok {
			pubKey = key
			break
		}
	}

	return pubKey, nil
}
