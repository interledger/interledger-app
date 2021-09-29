package main

import (
	"bytes"
	"text/template"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/imagebuilder"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func newVaultAmi(ctx *pulumi.Context,
	infrastructureConfigurationArn pulumi.StringOutput,
	distributionConfigurationArn pulumi.StringOutput,
	version string,
) (*imagebuilder.Image, error) {

	installVault, err := imagebuilder.NewComponent(ctx, "install-vault", &imagebuilder.ComponentArgs{
		Platform: pulumi.String("Linux"),
		Version:  pulumi.String("1.0.0"),
		Data:     pulumi.String(newInstallComponent("1.8.1", "/opt/vault")),
	})

	if err != nil {
		return nil, err
	}

	mostRecent := true
	linux2Ami, err := ec2.LookupAmi(ctx, &ec2.LookupAmiArgs{
		Filters: []ec2.GetAmiFilter{
			{
				Name:   "name",
				Values: []string{"amzn2-ami-hvm-*-x86_64-ebs"},
			},
		},
		Owners:     []string{"137112412989"},
		MostRecent: &mostRecent,
	})
	if err != nil {	
		return nil, err
	}

	recipe, err := imagebuilder.NewImageRecipe(ctx, "vault-image-recipe", &imagebuilder.ImageRecipeArgs{
		Components: imagebuilder.ImageRecipeComponentArray{
			&imagebuilder.ImageRecipeComponentArgs{
				ComponentArn: installVault.Arn,
			},
		},
		ParentImage: pulumi.String(linux2Ami.Id),
		Version:     pulumi.String(version),
		Name:        pulumi.String("fynbos-vault-linux-ebs"),
	})
	if err != nil {
		return nil, err
	}

	image, err := imagebuilder.NewImage(ctx, "vault-ami", &imagebuilder.ImageArgs{
		ImageRecipeArn:                 recipe.Arn,
		DistributionConfigurationArn:   distributionConfigurationArn,
		InfrastructureConfigurationArn: infrastructureConfigurationArn,
	})
	if err != nil {
		return nil, err
	}

	return image, nil
}

func newInstallComponent(
	version string,
	installPath string,
) string {
	type Data struct {
		InstallPath       string
		Version           string
	}
	data := Data{
		InstallPath: "/opt/vault",
		Version: "1.8.1",
	}

	template, err := template.ParseFiles("./install_component.tmpl")
	if err != nil {
		return ""
	}

	document := &bytes.Buffer{}
	template.Execute(document, data)

	return document.String()
}


type ImageBuilderConfig struct {
	ctx *pulumi.Context
	subnet pulumi.StringOutput
	region string
	secGroup *ec2.SecurityGroup
	enableLogging bool
}
func newImageBuilderConfiguration(args ImageBuilderConfig) (*imagebuilder.InfrastructureConfiguration, *imagebuilder.DistributionConfiguration, *s3.Bucket, error) {
	var infrastructure * imagebuilder.InfrastructureConfiguration
	var err error
	var bucket *s3.Bucket

	if args.enableLogging {
		bucket, err = newLogsBucket(args.ctx)
		if err != nil {
			return nil, nil, nil, err
		}

		profile, err := newBuilderThatLogsProfile(args.ctx, bucket)
		if err != nil {
			return nil, nil, nil, err
		}

		infrastructure, err = imagebuilder.NewInfrastructureConfiguration(args.ctx, "vault-image-infrastructure-config", &imagebuilder.InfrastructureConfigurationArgs{
			Description:         pulumi.String("Vault image builder configuration"),
			InstanceProfileName: profile.Name,
			InstanceTypes: pulumi.StringArray{
				pulumi.String("m5.large"),
			},
			TerminateInstanceOnFailure: pulumi.Bool(true),
			SubnetId:                   args.subnet,
			SecurityGroupIds: pulumi.StringArray{
				args.secGroup.ID(),
			},
			Logging: &imagebuilder.InfrastructureConfigurationLoggingArgs{
				S3Logs: &imagebuilder.InfrastructureConfigurationLoggingS3LogsArgs{
					S3BucketName: bucket.ID(),
					S3KeyPrefix:  pulumi.String("logs"),
				},
			},
			Tags: pulumi.StringMap{
				"service": pulumi.String("vault-image-builder"),
			},
		})
		if err != nil {
			return nil, nil, nil, err
		}	
	} else {
		profile, err := newBuilderProfile(args.ctx)
		if err != nil {
			return nil, nil, nil, err
		}

		infrastructure, err = imagebuilder.NewInfrastructureConfiguration(args.ctx, "vault-image-infrastructure-config", &imagebuilder.InfrastructureConfigurationArgs{
			Description:         pulumi.String("Vault image builder configuration"),
			InstanceProfileName: profile.Name,
			InstanceTypes: pulumi.StringArray{
				pulumi.String("m5.large"),
			},
			TerminateInstanceOnFailure: pulumi.Bool(true),
			SubnetId:                   args.subnet,
			SecurityGroupIds: pulumi.StringArray{
				args.secGroup.ID(),
			},
			Tags: pulumi.StringMap{
				"service": pulumi.String("vault-image-builder"),
			},
		})
		if err != nil {
			return nil, nil, nil, err
		}	
	}	

	distribution, err := imagebuilder.NewDistributionConfiguration(args.ctx, "vault-image-distribution-config", &imagebuilder.DistributionConfigurationArgs{
		Distributions: imagebuilder.DistributionConfigurationDistributionArray{
			&imagebuilder.DistributionConfigurationDistributionArgs{
				AmiDistributionConfiguration: &imagebuilder.DistributionConfigurationDistributionAmiDistributionConfigurationArgs{
					Name: pulumi.String("fynbos-vault-{{ imagebuilder:buildDate }}"),
				},
				Region: pulumi.String(args.region),
			},
		},
		Tags: pulumi.StringMap{
			"service": pulumi.String("vault-image-builder"),
		},
	})
	if err != nil {
		return nil, nil, nil, err
	}

	return infrastructure, distribution, bucket, nil
}



func newLogsBucket(ctx *pulumi.Context) (*s3.Bucket, error) {
	bucket, err := s3.NewBucket(ctx, "vault-bucket", &s3.BucketArgs{
		Tags: pulumi.StringMap{
			"service": pulumi.String("vault-image-builder"),
		},
	})
	if err != nil {
		return nil, err
	}

	return bucket,nil
}
