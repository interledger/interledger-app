package main

import (
	"encoding/json"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CreateRole
// creates the required role for the gitlab runner manager
// ensures ec2 instances can assume the role
// adds the necessary policies for the role for ec2 management
func CreateRole(ctx *pulumi.Context) (*iam.Role, error) {
	instanceAssumeRolePolicy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
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

	role, err := iam.NewRole(ctx, "gl-runner-manager", &iam.RoleArgs{
		Path:             pulumi.String("/"),
		AssumeRolePolicy: pulumi.String(instanceAssumeRolePolicy.Json),
	})
	if err != nil {
		return nil, err
	}

	_, err = iam.NewRolePolicy(ctx, "gl-runner-manager", &iam.RolePolicyArgs{
		Role:   role.ID(),
		Policy: policy(),
	})
	return role, nil
}

func policy() pulumi.String {

	type Statement struct {
		Effect   string   `json:"Effect"`
		Action   []string `json:"Action,omitempty"`
		Resource string   `json:"Resource"`
	}

	type Policy struct {
		Version    string      `json:"Version"`
		ID         string      `json:"Id"`
		Statements []Statement `json:"Statement"`
	}

	rawPolicy := &Policy{
		Version: "2012-10-17",
		ID:      "Gitlab runner management instance policy",
		Statements: []Statement{
			{
				Effect: "Allow",
				Action: []string{
					"ec2:DescribeSubnets",
					"ec2:DescribeSecurityGroups",
					"ec2:DescribeKeyPairs",
					"ec2:ImportKeyPair",
					"ec2:DeleteKeyPair",
					"ec2:RequestSpotInstances",
					"ec2:DescribeSpotInstanceRequests",
					"ec2:CancelSpotInstanceRequests",
					"ec2:RunInstances",
					"ec2:DescribeInstances",
					"ec2:CreateTags",
					"ec2:TerminateInstances",
				},
				Resource: "*",
			},
		},
	}

	policy, _ := json.Marshal(rawPolicy)

	return pulumi.String(policy)
}
