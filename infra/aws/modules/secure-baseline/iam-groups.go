package secure_baseline

import (
	"encoding/json"
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func NewCrossAccountGroup(ctx *pulumi.Context, provider *aws.Provider, name string, roleArn string) (*iam.Group, error) {
	group, err := iam.NewGroup(ctx, name, &iam.GroupArgs{
		Name: pulumi.String(name),
		Path: pulumi.String("/accounts/"),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, err
	}
	_, err = iam.NewGroupPolicy(ctx, fmt.Sprintf("%s-policy", name), &iam.GroupPolicyArgs{
		Group:  group.ID(),
		Policy: pulumi.String(createCrossAccountRolePolicy(roleArn)),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, err
	}

	return group, nil
}

func createCrossAccountRolePolicy(roleArn string) string {

	type Statement struct {
		Effect   string   `json:"Effect"`
		Action   string   `json:"Action"`
		Resource []string `json:"Resource"`
	}

	type Policy struct {
		Version   string    `json:"Version"`
		ID        string    `json:"Id"`
		Statement Statement `json:"Statement"`
	}

	rawPolicy := Policy{
		Version: "2012-10-17",
		Statement: Statement{
			Effect: "Allow",
			Action: "sts:AssumeRole",
			Resource: []string{
				roleArn,
			},
		},
	}

	policy, _ := json.Marshal(rawPolicy)

	return string(policy)
}
