package main

import (
	"github.com/pulumi/pulumi-docker/sdk/v3/go/docker"
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

		// used to store our custom docker image that has the aws credential helper.
		dockerRepo, err := ecr.NewPrivateRepository(ctx, "docker", accountID)
		if err != nil {
			return err
		}

		image, err := docker.NewImage(ctx, "docker-ecr", &docker.ImageArgs{
			Build:     &docker.DockerBuildArgs{Context: pulumi.String("./")},
			ImageName: pulumi.String(accountID + ".dkr.ecr.eu-west-1.amazonaws.com/docker:19.03.12"),
			Registry:  docker.ImageRegistryArgs{}, // use ECR credential helper
		})

		ctx.Export("dockerRepoUri", dockerRepo.RepositoryUri)
		ctx.Export("dockerImage", image.ImageName)
		ctx.Export("backendEcrRepoUri", backendEcrRepo.RepositoryUri)
		return nil
	})
}
