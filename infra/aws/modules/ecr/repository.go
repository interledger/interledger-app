package ecr

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/ecr"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// This creates a private repository that:
// - removes untagged images after 2 months
// - scans images on push
// - has immutable tags
// - relies on default s3 encryption settings
func NewPrivateRepository(ctx *pulumi.Context, name string, accountID string) (*ecr.Repository, error) {
	// Defaulting to remove untagged images after ~2 months for now.
	lifecyclePolicy, err := RemoveUntaggedImagesAfterDaysPolicy(60)
	if err != nil {
		return nil, err
	}

	repo, err := ecr.NewRepository(ctx, name, &ecr.RepositoryArgs{
		LifecyclePolicy: &ecr.RepositoryLifecyclePolicyArgs{
			LifecyclePolicyText: pulumi.String(lifecyclePolicy),
			RegistryId:          pulumi.String(accountID),
		},
		ImageScanningConfiguration: &ecr.RepositoryImageScanningConfigurationArgs{
			ScanOnPush: pulumi.Bool(true),
		},
		// allow mutable tags as builds with the same hash may need to be overwritten.
		// e.g. base builds.
		ImageTagMutability: ecr.RepositoryImageTagMutabilityMutable,
		RepositoryName:     pulumi.String(name),
	})
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func AllowAccountPushPullAccessToRepo(ctx *pulumi.Context) (*iam.GetPolicyDocumentResult, error) {
	policy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Actions: []string{
					"sts:AssumeRole",
				},
				Principals: []iam.GetPolicyDocumentStatementPrincipal{
					{
						Type: "Service",
						Identifiers: []string{
							"ec2.amazonaws.com",
						},
					},
				},
			},
		},
	}, nil)
	if err != nil {
		return nil, err
	}

	return policy, nil
}

func RemoveUntaggedImagesAfterDaysPolicy(days uint) ([]byte, error) {
	type Selection struct {
		TagStatus   string `json:"tagStatus"`
		CountType   string `json:"countType"`
		CountUnit   string `json:"countUnit"`
		CountNumber int    `json:"countNumber"`
	}

	type Action struct {
		Type string `json:"type"`
	}

	type Rule struct {
		Priority    int       `json:"rulePriority"`
		Description string    `json:"description"`
		Selection   Selection `json:"selection"`
		Action      Action    `json:"action"`
	}
	type LifecyclePolicy struct {
		Rules []Rule `json:"rules"`
	}

	lifecyclePolicy := LifecyclePolicy{
		Rules: []Rule{
			{
				Priority:    1,
				Description: fmt.Sprintf("Expire untagged images older than %d days", days),
				Selection: Selection{
					TagStatus:   "untagged",
					CountType:   "sinceImagePushed",
					CountUnit:   "days",
					CountNumber: int(days),
				},
				Action: Action{
					Type: "expire",
				},
			},
		},
	}

	policyBytes, err := json.Marshal(lifecyclePolicy)
	if err != nil {
		return nil, err
	}

	return policyBytes, nil
}
