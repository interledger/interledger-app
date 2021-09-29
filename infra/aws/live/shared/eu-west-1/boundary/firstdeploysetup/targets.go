package main

import (
	"context"
	"log"
	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/api/hostcatalogs"
)

func createVaultApiTarget(ctx context.Context, client *api.Client, catalog *hostcatalogs.HostCatalog) error {
	vaultHostSet, err := createHostSet(ctx, client, catalog.Id, "vault")
	if err != nil {
		log.Fatalln("Failed to create Vault host set.")
		return nil
	}
	log.Println(vaultHostSet)

	vaultHost, err := createHost(ctx, client, catalog.Id, "vault", "vault.fynbos.cloud") // adding the private dns as the host
	if err != nil {
		log.Fatalln("Failed to create Vault host.")
		return nil
	}
	log.Println(vaultHost)

	_, err = addHostToHostSet(ctx, client, vaultHostSet, vaultHost.Id)
	if err != nil {
		log.Fatalln("Failed to add Vault host to host set.")
		return err
	}

	vaultTarget, err := createTarget(ctx, client, catalog.ScopeId, "vault-api", "tcp", 8200, 2) // create api access to vault. We limit to 2 connections as that's the minimum number of unseal keys needed.
	if err != nil {
		log.Fatalln("Failed to create target for Vault.")
		return nil
	}
	log.Println(vaultTarget)

	_, err = addHostSetToTarget(ctx, client, vaultTarget, vaultHostSet.Id)
	if err != nil {
		log.Fatalln("Failed to add host set to Vault target.")
		return nil
	}

	return nil
}