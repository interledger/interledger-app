package secure_baseline

import (
	"encoding/json"
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func NewCrossAccountIamRoles(ctx *pulumi.Context, securityAccountId string) (*iam.Role, error) {

	role, err := iam.NewRole(ctx, "fullAccessRole", &iam.RoleArgs{
		Name:             pulumi.String("allow-full-access-from-other-accounts"),
		AssumeRolePolicy: pulumi.String(newAssumeRolePolicy(securityAccountId, true)),
		ManagedPolicyArns: pulumi.StringArray{
			pulumi.String("arn:aws:iam::aws:policy/AdministratorAccess"),
		},
	})
	if err != nil {
		return nil, err
	}

	return role, nil
}

func NewDevCrossAccountIamRole(ctx *pulumi.Context, securityAccountId string) (*iam.Role, error) {

	role, err := iam.NewRole(ctx, "readAccessRole", &iam.RoleArgs{
		Name:             pulumi.String("allow-read-access-from-other-accounts"),
		AssumeRolePolicy: pulumi.String(newAssumeRolePolicy(securityAccountId, true)),
		ManagedPolicyArns: pulumi.StringArray{
			pulumi.String("arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"),
		},
	})
	if err != nil {
		return nil, err
	}

	return role, nil
}

func NewECRCrossAccountIamRole(ctx *pulumi.Context, securityAccountId string) (*iam.Role, error) {

	role, err := iam.NewRole(ctx, "fullECRAccessRole", &iam.RoleArgs{
		Name:             pulumi.String("allow-full-ecr-access-from-other-accounts"),
		AssumeRolePolicy: pulumi.String(newAssumeRolePolicy(securityAccountId, false)),
		ManagedPolicyArns: pulumi.StringArray{
			pulumi.String("arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryPowerUser"),
		},
	})
	if err != nil {
		return nil, err
	}

	return role, nil
}

func newAssumeRolePolicy(securityAccountId string, enforceMFA bool) string {

	type Principal struct {
		AWS string `json:"AWS,omitempty"`
	}

	type Condition struct {
		Bool map[string]string `json:"Bool,omitempty"`
	}

	type Statement struct {
		Effect    string     `json:"Effect"`
		Principal Principal  `json:"Principal"`
		Action    string     `json:"Action"`
		Condition *Condition `json:"Condition,omitempty"`
	}

	type Policy struct {
		Version   string    `json:"Version"`
		ID        string    `json:"Id"`
		Statement Statement `json:"Statement"`
	}

	condition := &Condition{}
	if enforceMFA {
		condition.Bool = map[string]string{
			"aws:MultiFactorAuthPresent": "true",
		}
	}

	rawPolicy := Policy{
		Version: "2012-10-17",
		Statement: Statement{
			Effect: "Allow",
			Principal: Principal{
				AWS: fmt.Sprintf("arn:aws:iam::%s:root", securityAccountId),
			},
			Action:    "sts:AssumeRole",
			Condition: condition,
		},
	}

	policy, _ := json.Marshal(rawPolicy)

	return string(policy)
}
