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
		crossAccountIds := []string{
			"634848879735",
			"806897333161",
		}

		backendEcrRepo, err := ecr.NewPrivateRepository(ctx, "backend", accountID, crossAccountIds)
		if err != nil {
			return err
		}
		proteaRepo, err := ecr.NewPrivateRepository(ctx, "protea", accountID, crossAccountIds)
		if err != nil {
			return err
		}
		pacioliRepo, err := ecr.NewPrivateRepository(ctx, "pacioli", accountID, crossAccountIds)
		if err != nil {
			return err
		}

		// used to store our custom docker image that has the aws credential helper.
		dockerRepo, err := ecr.NewPrivateRepository(ctx, "docker", accountID, crossAccountIds)
		if err != nil {
			return err
		}

		dockerImage, err := docker.NewImage(ctx, "docker-ecr", &docker.ImageArgs{
			Build: &docker.DockerBuildArgs{
				Context:    pulumi.String("./"),
				Dockerfile: pulumi.String("./DockerfileEcr"),
			},
			ImageName: pulumi.String(accountID + ".dkr.ecr.eu-west-1.amazonaws.com/docker:22.04"),
			Registry:  docker.ImageRegistryArgs{}, // use ECR credential helper
		})
		if err != nil {
			return err
		}

		// used to store the custom image for deploying to eks
		eksRepo, err := ecr.NewPrivateRepository(ctx, "eks", accountID, crossAccountIds)
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
		if err != nil {
			return err
		}

		// Registry for dependent images
		_, err = ecr.NewPrivateRepository(ctx, "cockroach", accountID, crossAccountIds)
		if err != nil {
			return err
		}
		_, err = ecr.NewPrivateRepository(ctx, "kratos", accountID, crossAccountIds)
		if err != nil {
			return err
		}
		_, err = ecr.NewPrivateRepository(ctx, "tigerbeetle", accountID, crossAccountIds)
		if err != nil {
			return err
		}
		_, err = ecr.NewPrivateRepository(ctx, "temporalite", accountID, crossAccountIds)
		if err != nil {
			return err
		}
		if _, err = ecr.NewPrivateRepository(ctx, "rafiki-backend", accountID, crossAccountIds); err != nil {
			return err
		}
		if _, err = ecr.NewPrivateRepository(ctx, "botanist", accountID, crossAccountIds); err != nil {
			return err
		}

		//Temporal
		if _, err = ecr.NewPrivateRepository(ctx, "temporalio/auto-setup", accountID, crossAccountIds); err != nil {
			return err
		}
		if _, err = ecr.NewPrivateRepository(ctx, "temporalio/ui", accountID, crossAccountIds); err != nil {
			return err
		}
		if _, err = ecr.NewPrivateRepository(ctx, "temporalio/server", accountID, crossAccountIds); err != nil {
			return err
		}
		if _, err = ecr.NewPrivateRepository(ctx, "temporalio/admin-tools", accountID, crossAccountIds); err != nil {
			return err
		}

		certWatcherRepo, err := ecr.NewPrivateRepository(ctx, "certwatcher", accountID, crossAccountIds)
		if err != nil {
			return err
		}

		certWatcherImage, err := docker.NewImage(ctx, "docker-certwatcher", &docker.ImageArgs{
			Build: &docker.DockerBuildArgs{
				Context:    pulumi.String("./certwatcher"),
				Dockerfile: pulumi.String("./certwatcher/Dockerfile"),
			},
			ImageName: pulumi.String(accountID + ".dkr.ecr.eu-west-1.amazonaws.com/certwatcher:3.17.0"),
			Registry:  docker.ImageRegistryArgs{}, // use ECR credential helper
		})
		if err != nil {
			return err
		}

		if _, err = ecr.NewPrivateRepository(ctx, "discordbot", accountID, crossAccountIds); err != nil {
			return err
		}

		ctx.Export("certWatcherRepoUri", certWatcherRepo.RepositoryUri)
		ctx.Export("certWatcherDockerImage", certWatcherImage.ImageName)
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
