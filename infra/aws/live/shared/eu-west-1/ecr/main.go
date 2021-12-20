package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"

	ecr "gitlab.com/fynbos/infra/aws/modules/ecr"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "fynbos")
		accountID := cfg.Require("accountID")
		backendEcrRepo, err := ecr.NewPrivateRepository(ctx, "backend", accountID)
		if err != nil {
			return err
		}

		ctx.Export("backendEcrRepoUri", backendEcrRepo.RepositoryUri)
		return nil
	})
}
