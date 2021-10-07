package secure_baseline

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/ebs"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/kms"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func EbsBaseline(ctx *pulumi.Context, name string, accountId string) (*kms.Key, error) {

	key, err := kms.NewKey(ctx, "ebs", &kms.KeyArgs{
		Description: pulumi.String("KMS key ebs baseline"),
		Policy:      CreateEbsPolicy(accountId), // Need to still add policies for the IAM roles
	})
	if err != nil {
		return nil, err
	}

	// Configure new default kms key
	_, err = ebs.NewDefaultKmsKey(ctx, name, &ebs.DefaultKmsKeyArgs{
		KeyArn: key.Arn,
	})
	if err != nil {
		return nil, err
	}

	// Ensure that the default kms key is used for all EBS volumes
	_, err = ebs.NewEncryptionByDefault(ctx, name, &ebs.EncryptionByDefaultArgs{
		Enabled: pulumi.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	return key, nil
}

func CreateEbsPolicy(accountId string) pulumi.String {

	type Condition struct {
		BoolIfExists map[string]string `json:"BoolIfExists,omitempty"`
	}

	type Principal struct {
		AWS     []string `json:"AWS,omitempty"`
		Service string   `json:"Service,omitempty"`
	}

	type Statement struct {
		Sid       string     `json:"Sid"`
		Effect    string     `json:"Effect"`
		Principal Principal  `json:"Principal"`
		Action    []string   `json:"Action,omitempty"`
		NotAction []string   `json:"NotAction,omitempty"`
		Resource  string     `json:"Resource"`
		Condition *Condition `json:"Condition,omitempty"`
	}

	type Policy struct {
		Version    string      `json:"Version"`
		ID         string      `json:"Id"`
		Statements []Statement `json:"Statement"`
	}

	rawPolicy := &Policy{
		Version: "2012-10-17",
		ID:      "EBS encryption policy",
		Statements: []Statement{
			{
				Sid:    "Enable IAM policies",
				Effect: "Allow",
				Principal: Principal{
					AWS: []string{
						fmt.Sprintf("arn:aws:iam::%s:root", accountId),
					},
				},
				Action: []string{
					"kms:*",
				},
				Resource: "*",
			},
		},
	}

	policy, _ := json.Marshal(rawPolicy)

	return pulumi.String(policy)
}
