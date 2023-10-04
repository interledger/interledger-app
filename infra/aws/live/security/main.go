package main

import (
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/cloudtrail"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	secure_baseline "gitlab.com/fynbos/infra/aws/modules/secure-baseline"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		conf := config.New(ctx, "fynbos")
		accountId := conf.Require("accountId")
		provider, err := aws.NewProvider(ctx, "privileged", &aws.ProviderArgs{
			AssumeRole: &aws.ProviderAssumeRoleArgs{
				RoleArn:     pulumi.Sprintf("arn:aws:iam::%s:role/OrganizationAccountAccessRole", accountId),
				SessionName: pulumi.String("PulumiSession"),
				ExternalId:  pulumi.String("PulumiApplication"),
			},
			Region: aws.RegionEUWest1,
		})
		if err != nil {
			return err
		}
		rootStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-root/baseline", nil)
		if err != nil {
			return err
		}
		ctKMSKeyArn := rootStack.GetOutput(pulumi.String("cloudtrailKMSKeyArn"))
		ctS3BucketName := rootStack.GetOutput(pulumi.String("cloudtrailS3BucketName"))

		// Configure cloudtrail
		_, err = cloudtrail.NewTrail(ctx, "security-ct", &cloudtrail.TrailArgs{
			EnableLogFileValidation:    pulumi.Bool(true),
			IncludeGlobalServiceEvents: pulumi.Bool(true),
			IsMultiRegionTrail:         pulumi.Bool(true),
			KmsKeyId:                   pulumi.Sprintf("%s", ctKMSKeyArn),
			S3BucketName:               pulumi.Sprintf("%s", ctS3BucketName),
		}, pulumi.Provider(provider))
		if err != nil {
			return err
		}

		/* Setup strong password policy */
		_, err = iam.NewAccountPasswordPolicy(ctx, "password-policy", &iam.AccountPasswordPolicyArgs{
			AllowUsersToChangePassword: pulumi.Bool(true),
			MinimumPasswordLength:      pulumi.Int(14),
			RequireLowercaseCharacters: pulumi.Bool(true),
			RequireNumbers:             pulumi.Bool(true),
			RequireSymbols:             pulumi.Bool(true),
			RequireUppercaseCharacters: pulumi.Bool(true),
		}, pulumi.Provider(provider))
		if err != nil {
			return err
		}

		/**
		Create Groups
		*/
		fullAccessGroup, err := iam.NewGroup(ctx, "full-access", &iam.GroupArgs{
			Path: pulumi.String("/users/"),
		}, pulumi.Provider(provider))
		if err != nil {
			return err
		}

		userSelfMgmtGroup, err := iam.NewGroup(ctx, "iam-user-self-mgmt", &iam.GroupArgs{
			Path: pulumi.String("/users/"),
		}, pulumi.Provider(provider))
		if err != nil {
			return err
		}
		_, err = iam.NewGroupPolicy(ctx, "self-mgmt-policy", &iam.GroupPolicyArgs{
			Group:  userSelfMgmtGroup.ID(),
			Policy: CreateSlfMgmtPolicy(),
		}, pulumi.Provider(provider))
		if err != nil {
			return err
		}

		/** Cross Account Groups
		 */
		accSharedFullAccessGroup, err := secure_baseline.NewCrossAccountGroup(ctx, provider, "_account.shared-full-access", "arn:aws:iam::823058932981:role/allow-full-access-from-other-accounts")
		if err != nil {
			return err
		}
		accSharedReadAccessGroup, err := secure_baseline.NewCrossAccountGroup(ctx, provider, "_account.shared-read-access", "arn:aws:iam::823058932981:role/allow-read-access-from-other-accounts")
		if err != nil {
			return err
		}
		accSharedEcrFullAccessGroup, err := secure_baseline.NewCrossAccountGroup(ctx, provider, "_account.ecr-full-access", "arn:aws:iam::823058932981:role/allow-full-ecr-access-from-other-accounts")
		if err != nil {
			return err
		}

		accDevFullAccessGroup, err := secure_baseline.NewCrossAccountGroup(ctx, provider, "_account.dev-full-access", "arn:aws:iam::634848879735:role/allow-full-access-from-other-accounts")
		if err != nil {
			return err
		}

		accSandboxFullAccessGroup, err := secure_baseline.NewCrossAccountGroup(ctx, provider, "_account.sandbox-full-access", "arn:aws:iam::282137670145:role/allow-full-access-from-other-accounts")
		if err != nil {
			return err
		}

		accProdFullAccessGroup, err := secure_baseline.NewCrossAccountGroup(ctx, provider, "_account.prod-full-access", "arn:aws:iam::806897333161:role/allow-full-access-from-other-accounts")
		if err != nil {
			return err
		}

		/**
		Create Users
		*/
		err = CreateUser(ctx, "matt", pulumi.StringArray{
			fullAccessGroup.Name,
			userSelfMgmtGroup.Name,
			accSharedFullAccessGroup.Name,
			accDevFullAccessGroup.Name,
			accSandboxFullAccessGroup.Name,
			accProdFullAccessGroup.Name,
		}, "keybase:matdehaast", provider)
		if err != nil {
			return err
		}
		err = CreateUser(ctx, "don", pulumi.StringArray{
			userSelfMgmtGroup.Name,
			accSharedFullAccessGroup.Name,
			accDevFullAccessGroup.Name,
			accSandboxFullAccessGroup.Name,
			accProdFullAccessGroup.Name,
		}, "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", provider)
		if err != nil {
			return err
		}
		err = CreateUser(ctx, "cairin", pulumi.StringArray{
			userSelfMgmtGroup.Name,
			accSharedFullAccessGroup.Name,
			accDevFullAccessGroup.Name,
		}, "mQINBGEWazkBEACn9GBYQWGXKaOgZSPBljHNjl/UNs0YqiucT/HT54OzujjjO723pc9SCqMe333WmMobarrJPrD8usXHOoLrZYcZvegJJFsBHLeHe9/AXpwvY09jgHIVHV+IyGg2VqUA61pr9uk1z2QI9IrIs0CPcoVX/hqsLL83Sd/eg/5bmx0qxxEd611Pixbebi+MMXnKq6fKFdOnWled+HV5lvt9RTt/lVGbhukYgGS96UO0sli1a3iZomHKVCl7KkPa7d5l7kLlKT/yIUKgQ2L3+RX7q+3NRQzo/J0ZQqLgrUQA/XdbsDe9+7YgdRMgBAWNzB13QReB9uT2VOlMoXj3dbWWoA5r7+hCrKkkMpfHCvJBhvdwKYN9UeazZsOq2YhbbMk7DycZ7EI22my6boZUIUH55I5vjkly42om1DnE4932KQRIKzX91v1vZ8NyyCPxvMSkStxfj8FAisJi0cDP074QeCveihpoEbhrK+O8fAas7bqcfZ9/f6l8iar9rxHIEdFZdMPisK9JsxzqxLj5lNFe0ifeQShmSFJmZPDmLK4w5od7VfWKuWTllyVJ2t5FYm4lZUtpbF9i2Qy58vW38AwCDkIrBO4sDawlxwlfcI8bdAKsLVYdeiSlr8BIwyedSpfiNs2WrYQprK+4/CwPEzhOyDCYuVkBdCLRCebk/YUtBAlS3QARAQABtCZDYWlyaW4gTWljaGllIDxjYWlyaW5taWNoaWVAZ21haWwuY29tPokCTgQTAQgAOBYhBLgzREoGhXkK3f57Jkl6EuNU+zp0BQJhFms5AhsDBQsJCAcCBhUKCQgLAgQWAgMBAh4BAheAAAoJEEl6EuNU+zp0JO8QAJSAQF1bDHWv1SbIjqh8P0kn3T0oHmT4l2g259bGNlIWGgTr3iRtBfdvbeVTKaW6DNyqZVboJwazYBRNDgNQ/a2Q7nSF4Kg8t0re1IwcCrK1S03xPupC19YbU+ph52HPl7/NQx9Z1bi3zTVkKNojB/WuBn7oIQypk3xuFxtU6oaR2YveESOaKzl1rfqge336lHFqAildxg159bAE1BPZft6EkXo4scBBZahc55tnRl6KYzNt5ywH7zNH20QNaHBEiY914whJuOceIRBJRHSc1BOd07AIW+9tCwMNPkmBHTKC+1R1wfkSMTfccnblqyVEXXrvruu5smNihTBCng4zB/yp2FMJg9+MHF8Z3dBiQvA9RHEKHTPZ8AdpQCjxlMwjQR9AU3Y0SrMPMut1DRGzwyUpfXH/hB4Id3e2lY59X8lV4gB9AKqypPO12fMSdBq23mh9PCOcEwmCT05MyqCVXXj3vjX8v6feVdLp/vvrwGgPR8YvuntCODegGHLV7ftZAUnEHT2CyG0ze61E7dL9JNKntrxKubuFXFUrdrpIu87XBlxmoiSFcWUuA3K5IGX8v3z2fyPNsL0JQIxhoDfgYtveOptVX1E8c+QQJqJsJ+zUvHdoy0gdJEnivQ+hjd/waLFXUTgy5Z5SBj+cWmkvWasLrdGCpwbfWobL3k+p8PfwuQINBGEWazkBEACmZxJN4vDkjwzMvS1PVm5IKsKY/cLgcLU0zRnTRxhCRVlaKQYDL2OtM3tgd7DpDgJP5de+4eSGDZOf/AgmJVtxVzgdIdIZbiV1bsdz6mR/+gbD+SEpVSGWUToNWRY07vALLQEQT/TtlSam/xURknMp0zkfGtS+XBYEJCBHDlstJFOMK32i4PTs/4JxR/a/Ible/Sgt0XAd1y7GzU1NwrwSIzPot36jwAq6ZA6IBhZCXlza4ifdjRNtATScZApJKMDvXMC2Muj9xC60LcjXG1P9QzBMBnI6rhpePvneVC1+CyaMGv1V/g8k7onPXqvE1b0xX9MHbj3iSkFxZMeK273DzA8YhIuo21iKzyuTf7ZUUnQhhj1TcNJBCateDg8KAzC86liTeJU/vrnelfnU58yvGnABD1EKjMhEcNNMTNeQ+1NOAn5M7m9dcMIieXNCY36UDrDZj12Ys+HN+BRLDexffozrz3D8Z5uaHkRMTnjnwa4yEtRq3MQnFF0Z2JEaGoH6PhogscByvit5yzXWomp8LibLH+IltC1a6VBzXWvbGsr5Y2Smi5GQA0Vs30nMSHRTdDkZRRxNLgrkpS71nFJGSLv77kuk2ks0xmkAaexPHazAo+IOKKlj1Zi/iWilDGsHiA8tN26bLW1N1fDFtzc2t7NUm7TsnA3p2SEnSqnPbwARAQABiQI2BBgBCAAgFiEEuDNESgaFeQrd/nsmSXoS41T7OnQFAmEWazkCGwwACgkQSXoS41T7OnQQCA//TidYeH+LSCX8VFjVM+ikT2drFx6Bzb2YpgVfnT+H2f41eIBFEqW/4F6vBVxEhejd4BlL/+1Nd6XlFtCpPC70uXB4bWY8BuVKMlEeEJlsvDFWQL2mKBOpFFihlWH7OMtQiGPZJHqnx7m2ytZGfw1HKAyKD1AzpaPC0B13KoUBv46qcwhpPprsx5OG8fIat0Jf3+A0f+PZEFFoSz12VoKCVIzXNdJvlBidIrNYq9G7StkDWRv2kPBtr3LUJdfk9cB+xBZGiMcbAG+JjZTxmhwAMQf4+01PnuYG9F7qfczDW49Nk4DkK0u8o1MFBhKTjQXmwF5CtTIvVd7WxVIJpijjIsu96oZoKK6XDMwzvrCisvMmlnZPcCsjJVrAhBV9pwDixRT+zJg7G2DxWWDWzA7VPhD+SaKdnKFZkMJU0nrL3jCBnkXeU6TUNVMD6PiRulxzInJe8gE05FdObiNCzDyeNWEl+nBHCAK03YxnyBvUyZDio6e76wQ2kpprTBb1i9bq4jAAZ4ipv1FkjeXp6v+x2lnBLJyx3D693NUb6ySHcc0gzCxzHCvgrDShPv+gx4EAjJp8hB4QaYI+zwKvhoY0Xd3vyFLJGQPeI90y3znBBNsp0CJlCtnQFYfsQ51BTABu/1vDoXRj7wkLU8N5dzisS6xEoOwwjcqlHzX3Zj1/B44=", provider)
		if err != nil {
			return err
		}
		err = CreateUser(ctx, "omer", pulumi.StringArray{
			userSelfMgmtGroup.Name,
			accSharedReadAccessGroup.Name,
			accDevFullAccessGroup.Name,
		}, "keybase:omertoast", provider)
		if err != nil {
			return err
		}

		err = CreateUser(ctx, "barnard", pulumi.StringArray{
			userSelfMgmtGroup.Name,
			accSharedReadAccessGroup.Name,
			accDevFullAccessGroup.Name,
		}, "keybase:barnarddt", provider)
		if err != nil {
			return err
		}

		err = createECRUser(ctx, accSharedEcrFullAccessGroup.Name, provider)
		if err != nil {
			return err
		}

		err = createTabapayUser(ctx, provider)
		if err != nil {
			return err
		}

		err = createGMTUser(ctx, provider)
		if err != nil {
			return err
		}

		return nil
	})
}

func createECRUser(ctx *pulumi.Context, group pulumi.StringInput, provider *aws.Provider) error {

	user, err := iam.NewUser(ctx, fmt.Sprintf("ecrUser"), &iam.UserArgs{
		Name: pulumi.String("ecrUser"),
		Path: pulumi.String("/"),
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	accessKey, err := iam.NewAccessKey(ctx, "ecrUser", &iam.AccessKeyArgs{
		User:   user.Name,
		PgpKey: pulumi.String("keybase:matdehaast"),
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	_, err = iam.NewUserGroupMembership(ctx, "ecrUserGroup", &iam.UserGroupMembershipArgs{
		User: user.Name,
		Groups: pulumi.StringArray{
			group,
		},
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	ctx.Export(fmt.Sprintf("ecrUserAccessKey"), pulumi.Map{
		"user":      accessKey.User,
		"keyID":     accessKey.ID(),
		"keySecret": accessKey.EncryptedSecret,
	})
	return nil
}

func createTabapayUser(ctx *pulumi.Context, provider *aws.Provider) error {

	user, err := iam.NewUser(ctx, fmt.Sprintf("tabapay"), &iam.UserArgs{
		Name: pulumi.String("tabapay"),
		Path: pulumi.String("/"),
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	accessKey, err := iam.NewAccessKey(ctx, "tabapay", &iam.AccessKeyArgs{
		User:   user.Name,
		PgpKey: pulumi.String("keybase:matdehaast"),
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	ctx.Export(fmt.Sprintf("tabapayUser"), pulumi.Map{
		"arn":       user.Arn,
		"user":      accessKey.User,
		"keyID":     accessKey.ID(),
		"keySecret": accessKey.EncryptedSecret,
	})
	return nil
}

func createGMTUser(ctx *pulumi.Context, provider *aws.Provider) error {

	user, err := iam.NewUser(ctx, fmt.Sprintf("gmt"), &iam.UserArgs{
		Name: pulumi.String("gmt"),
		Path: pulumi.String("/"),
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	accessKey, err := iam.NewAccessKey(ctx, "gmt", &iam.AccessKeyArgs{
		User:   user.Name,
		PgpKey: pulumi.String("keybase:matdehaast"),
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	ctx.Export(fmt.Sprintf("gmtUser"), pulumi.Map{
		"arn":       user.Arn,
		"user":      accessKey.User,
		"keyID":     accessKey.ID(),
		"keySecret": accessKey.EncryptedSecret,
	})
	return nil
}
