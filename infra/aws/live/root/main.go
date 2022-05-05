package main

import (
	"encoding/json"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/cloudtrail"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/kms"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/organizations"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"io/ioutil"
	"os"
)

type Accounts struct {
	Accounts []Account `json:"accounts"`
}

type Account struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	IsLogs bool   `json:"is_logs"`
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		/**
		CREATE THE ORG
		*/
		organisation, err := organizations.NewOrganization(ctx, "fynbos", &organizations.OrganizationArgs{
			AwsServiceAccessPrincipals: pulumi.StringArray{
				pulumi.String("cloudtrail.amazonaws.com"),
				pulumi.String("config.amazonaws.com"),
			},
			FeatureSet: pulumi.String("ALL"),
		})

		if err != nil {
			return err
		}

		accounts, err := getConfigAccounts()
		if err != nil {
			return err
		}

		/**
		Create the accounts
		*/
		var logAccount *organizations.Account
		var accountIds []pulumi.IDOutput
		for _, a := range accounts.Accounts {
			if a.IsLogs {
				logAccount, err = createOrgAccount(ctx, a.Name, a.Email, organisation)
				if err != nil {
					return err
				}
				accountIds = append(accountIds, logAccount.ID())
			} else {
				account, err := createOrgAccount(ctx, a.Name, a.Email, organisation)
				if err != nil {
					return err
				}
				accountIds = append(accountIds, account.ID())
			}
		}

		/**
		Create the logs AWS Cloudtrail S3 Bucket and KMS key fynbos-cloudtrail-logs
		*/
		provider, err := aws.NewProvider(ctx, "privileged", &aws.ProviderArgs{
			AssumeRole: &aws.ProviderAssumeRoleArgs{
				RoleArn:     pulumi.Sprintf("arn:aws:iam::%s:role/OrganizationAccountAccessRole", logAccount.ID()),
				SessionName: pulumi.String("PulumiSession"),
				ExternalId:  pulumi.String("PulumiApplication"),
			},
			Region: aws.RegionEUWest1,
		})
		if err != nil {
			return err
		}

		ctAccessLogBucket, err := s3.NewBucket(ctx, "ct-access-log-bucket", &s3.BucketArgs{
			Acl: pulumi.String("log-delivery-write"),
		}, pulumi.Provider(provider))
		if err != nil {
			return err
		}

		ctBucket, err := s3.NewBucket(ctx, "ct-bucket", &s3.BucketArgs{
			Acl: pulumi.String("log-delivery-write"),
			Loggings: s3.BucketLoggingArray{
				&s3.BucketLoggingArgs{
					TargetBucket: ctAccessLogBucket.ID(),
					TargetPrefix: pulumi.String("log/"),
				},
			},
			Versioning: &s3.BucketVersioningArgs{
				Enabled: pulumi.Bool(true),
			},
		}, pulumi.Provider(provider))
		_, err = s3.NewBucketPolicy(ctx, "ct-bucket-policy", &s3.BucketPolicyArgs{
			Bucket: ctBucket.ID(),
			Policy: CreateS3Policy(ctBucket, organisation, accountIds),
		}, pulumi.Provider(provider))
		if err != nil {
			return err
		}

		ctx.Export("cloudtrailS3BucketName", ctBucket.ID())

		ctKms, err := kms.NewKey(ctx, "ct-kms", &kms.KeyArgs{
			Description:       pulumi.String("A KMS key to encrypt CloudTrail events"),
			EnableKeyRotation: pulumi.Bool(true),
			Policy:            CreateKMSPolicy(logAccount, organisation, accountIds),
		}, pulumi.Provider(provider))
		ctx.Export("cloudtrailKMSKeyArn", ctKms.Arn)

		/**
		Create the logs AWS Config S3 Bucket fynbos-config-logs
		*/

		/**
		Configure root account AWS Cloudtrail
		*/
		_, err = cloudtrail.NewTrail(ctx, "root-ct", &cloudtrail.TrailArgs{
			EnableLogFileValidation:    pulumi.Bool(true),
			IncludeGlobalServiceEvents: pulumi.Bool(true),
			IsMultiRegionTrail:         pulumi.Bool(true),
			KmsKeyId:                   ctKms.Arn,
			S3BucketName:               ctBucket.ID(),
		})
		if err != nil {
			return err
		}

		/**
		Setup IAM Groups for root
		*/
		fullAccessGroup, err := iam.NewGroup(ctx, "full-access", &iam.GroupArgs{
			Path: pulumi.String("/users/"),
		})
		if err != nil {
			return err
		}
		_, err = iam.NewGroupPolicyAttachment(ctx, "full-access-attach", &iam.GroupPolicyAttachmentArgs{
			Group:     fullAccessGroup.Name,
			PolicyArn: pulumi.String("arn:aws:iam::aws:policy/AdministratorAccess"),
		})
		if err != nil {
			return err
		}

		billingAccessGroup, err := iam.NewGroup(ctx, "billing-access", &iam.GroupArgs{
			Path: pulumi.String("/users/"),
		})

		if err != nil {
			return err
		}
		_, err = iam.NewGroupPolicyAttachment(ctx, "billing-access-attach", &iam.GroupPolicyAttachmentArgs{
			Group:     billingAccessGroup.Name,
			PolicyArn: pulumi.String("arn:aws:iam::aws:policy/AWSBillingReadOnlyAccess"),
		})
		if err != nil {
			return err
		}

		/**
		Setup Users
		*/
		// Imported user that was initially created
		userMatt, err := iam.NewUser(ctx, "user-matt", &iam.UserArgs{
			ForceDestroy: pulumi.Bool(false),
			Name:         pulumi.String("matt"),
			Path:         pulumi.String("/"),
		}, pulumi.Protect(true))
		if err != nil {
			return err
		}
		_, err = iam.NewUserGroupMembership(ctx, "matt-group", &iam.UserGroupMembershipArgs{
			User: userMatt.Name,
			Groups: pulumi.StringArray{
				fullAccessGroup.Name,
				billingAccessGroup.Name,
			},
		})
		if err != nil {
			return err
		}
		userAdrian, err := iam.NewUser(ctx, "user-adrian", &iam.UserArgs{
			Name: pulumi.String("adrian"),
			Path: pulumi.String("/"),
		})
		if err != nil {
			return err
		}
		_, err = iam.NewUserGroupMembership(ctx, "adrian-group", &iam.UserGroupMembershipArgs{
			User: userAdrian.Name,
			Groups: pulumi.StringArray{
				billingAccessGroup.Name,
			},
		})
		if err != nil {
			return err
		}

		return nil
	})
}

func getConfigAccounts() (*Accounts, error) {
	var accounts Accounts

	jsonFile, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &accounts)
	if err != nil {
		return nil, err
	}

	return &accounts, nil
}

func createOrgAccount(ctx *pulumi.Context, name string, email string, org *organizations.Organization) (*organizations.Account, error) {
	return organizations.NewAccount(ctx, name, &organizations.AccountArgs{
		Email: pulumi.String(email),
	}, pulumi.DependsOn([]pulumi.Resource{org}))
}
