package vault

import (
	"context"
	"fmt"
	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/kubernetes"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

func NewVaultClient() (*vault.Client, error) {

	config := vault.DefaultConfig()
	config.Address = "https://vault1.fynbos.cloud:8200"
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %w", err)
	}

	k8sAuth, err := auth.NewKubernetesAuth(
		"k8s-app",
		auth.WithMountPath("k8s-dev-euw1"),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Kubernetes auth method: %w", err)
	}

	authInfo, err := client.Auth().Login(context.TODO(), k8sAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to log in with Kubernetes auth: %w", err)
	}
	if authInfo == nil {
		return nil, fmt.Errorf("no auth info was returned after login")
	}

	log.Info("Successfully authed", zap.Any("authInfo", authInfo))

	return client, nil
}
