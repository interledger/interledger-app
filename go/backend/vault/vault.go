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
			auth.WithServiceAccountToken("eyJhbGciOiJSUzI1NiIsImtpZCI6ImQyMGZmZTJkZTM3YzliMWMxZjk2ODU5ZDRjZmFlODUyMmYxNWQ0YWEifQ.eyJh\ndWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjIl0sImV4cCI6MTcxNDYzNTcxNiwiaWF0IjoxNjgzMDk5NzE2LCJpc3MiOiJodHRwczovL29pZGMuZWtzLmV1LXdlc3QtMS5hbWF6b25hd3MuY29tL2lkL0FGODNGRkNDOEEzMUQxNkQyRTFDOUMxNzg4MjA1ODYzIiwia3ViZXJuZXRlcy5pbyI6eyJuYW1lc3BhY2UiOiJiYWNrZW5kIiwicG9kIjp7Im5hbWUiOiJiYWNrZW5kLXdvcmtlci02YzRmYjQ0ODktejdxOW0iLCJ1aWQiOiJiMWIyNzVmNi1mMzJkLTQzMDQtYmMyOC1jMmQ2ZDgzMjQzYTgifSwic2VydmljZWFjY291bnQiOnsibmFtZSI6ImJhY2tlbmQiLCJ1aWQiOiIxMTgxZmY5Ni0wOTI3LTQzN2MtYTJhNS1lMTZiMWM3MTJjYjkifSwid2FybmFmdGVyIjoxNjgzMTAzMzIzfSwibmJmIjoxNjgzMDk5NzE2LCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6YmFja2VuZDpiYWNrZW5kIn0.KfuFc_GmPnlzZ3y4k_PNfKBkSKLojdluIJmKCbnNyUXSKaU4joGl8fRMLP_u9xecych_5IZQnwnYsHa3FCg9C-biY7SeJMV4xfN9kYZo29GeBEuH5EiqdgFANUwpJ03s9lO3AkcYLi1ttg5XSnuZIAFX0yTh_whtyu4kknc_eYsGzjxUBE0f1eMSnsn1it1GFeIZEJLIiSD3cXYN5oU7zNEyx09_hmgw8oFwH4fDSUVxfPCxlDiWYAd6aNk8V-78PveKodoFW87NRb9ckNPUf9uPwYxzcH2vfNK2cLmMLkeuFf876DlgwyJiwlvVGfZeFMzG5gX4_VqRPVaW6NEluQ"),
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

func (c client) Sign(keyName string, input string) (string, error) {
	encodedInput := base64.StdEncoding.EncodeToString([]byte(input))
	keyPath := fmt.Sprintf("%s/sign/%s", c.transitEnginePath, keyName)

	data := map[string]interface{}{
		"input": encodedInput,
	}

	resp, err := c.vc.Logical().Write(keyPath, data)
	if err != nil {
		return "", err
	}

	return resp.Data["signature"].(string), nil
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
