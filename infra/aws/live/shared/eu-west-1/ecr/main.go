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

		dockerImage, err := docker.NewImage(ctx, "docker-ecr", &docker.ImageArgs{
			Build: &docker.DockerBuildArgs{
				Context:    pulumi.String("./"),
				Dockerfile: pulumi.String("./DockerfileEcr"),
			},
			ImageName: pulumi.String(accountID + ".dkr.ecr.eu-west-1.amazonaws.com/docker:19.03.12"),
			Registry:  docker.ImageRegistryArgs{}, // use ECR credential helper
		})

		// used to store the custom image for deploying to eks
		eksRepo, err := ecr.NewPrivateRepository(ctx, "eks", accountID)
		if err != nil {
			return err
		}

		eksImage, err := docker.NewImage(ctx, "eks", &docker.ImageArgs{
			Build: &docker.DockerBuildArgs{
				Context:    pulumi.String("./"),
				Dockerfile: pulumi.String("./DockerfileEks"),
			},
			ImageName: pulumi.String(accountID + ".dkr.ecr.eu-west-1.amazonaws.com/eks"),
			Registry:  docker.ImageRegistryArgs{}, // use ECR credential helper
		})

		ctx.Export("eksRepoUri", eksRepo.RepositoryUri)
		ctx.Export("eksImage", eksImage.ImageName)
		ctx.Export("dockerRepoUri", dockerRepo.RepositoryUri)
		ctx.Export("dockerImage", dockerImage.ImageName)
		ctx.Export("backendEcrRepoUri", backendEcrRepo.RepositoryUri)
		return nil
	})
}
