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
		proteaRepo, err := ecr.NewPrivateRepository(ctx, "protea", accountID)
		if err != nil {
			return err
		}
		pacioliRepo, err := ecr.NewPrivateRepository(ctx, "pacioli", accountID)
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
			ImageName: pulumi.String(accountID + ".dkr.ecr.eu-west-1.amazonaws.com/docker:20.10"),
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

		// Registry for dependent images
		_, err = ecr.NewPrivateRepository(ctx, "cockroach", accountID)
		if err != nil {
			return err
		}
		_, err = ecr.NewPrivateRepository(ctx, "kratos", accountID)
		if err != nil {
			return err
		}
		_, err = ecr.NewPrivateRepository(ctx, "tigerbeetle", accountID)
		if err != nil {
			return err
		}
		_, err = ecr.NewPrivateRepository(ctx, "temporalite", accountID)
		if err != nil {
			return err
		}
		if _, err = ecr.NewPrivateRepository(ctx, "rafiki-backend", accountID); err != nil {
			return err
		}

		ctx.Export("eksRepoUri", eksRepo.RepositoryUri)
		ctx.Export("eksImage", eksImage.ImageName)
		ctx.Export("dockerRepoUri", dockerRepo.RepositoryUri)
		ctx.Export("dockerImage", dockerImage.ImageName)
		ctx.Export("backendEcrRepoUri", backendEcrRepo.RepositoryUri)
		ctx.Export("proteaEcrRepoUri", proteaRepo.RepositoryUri)
		ctx.Export("pacioliRepoEcrRepoUri", pacioliRepo.RepositoryUri)
		return nil
	})
}
