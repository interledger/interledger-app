package kubernetes

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/aws/modules/utils"
)

// This role is able to manage all eks resources and is intended
// to be mapped to the k8s system:masters group which would
// allow full access to all k8s resources.
func NewEksClusterAdminPolicy(ctx *pulumi.Context) string {
	policy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"eks:*",
					"ec2:DescribeImages",
				},
				Resources: []string{"*"},
			},
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{"iam:PassRole"},
				Resources: []string{"*"},
			},
		},
	})
	if err != nil { return "" }

	return policy.Json
}

func NewEksClusterAdminRoleTrustPolicy(ctx *pulumi.Context, accountId string) string {
	policy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"sts:AssumeRole",
				},
				Principals: []iam.GetPolicyDocumentStatementPrincipal{
					{
						Type: "AWS",
						Identifiers: []string{
							//The suffix root in the policy’s Principal attribute equates to 
							// “authenticated and authorized principals in the account,” 
							// not the special and all-powerful root user principal that is created 
							// when an AWS account is created.
							// TODO: this needs to be more tightly scoped once we have more users.  e.g. arn:aws:iam::%s:user/xxx
							fmt.Sprintf("arn:aws:iam::%s:root", accountId),
						},
					},
				},
				Conditions: []iam.GetPolicyDocumentStatementCondition{
					{
						Test: "Bool",
						Values: []string{"true"},
						Variable: "aws:MultiFactorAuthPresent",
					},
				},
			},
		},
	})
	if err != nil { return "" }

	return policy.Json
}

func NewFluentdCloudwatchPolicy(ctx *pulumi.Context) string {
	policy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"logs:*",
				},
				Resources: []string{"arn:aws:logs:*:*:*"},
			},
		},
	})
	if err != nil { return "" }

	return policy.Json
}

func NewEksAutomationRoleTrustPolicy(ctx *pulumi.Context, accountId string) string {
	policy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"sts:AssumeRole",
				},
				Principals: []iam.GetPolicyDocumentStatementPrincipal{ // our Gitlab runners are currently run on ec2 instances.
					{
						Type: "Service",
						Identifiers: []string{"ec2.amazonaws.com"},
					},
				},
			},
		},
	})
	if err != nil { return "" }

	return policy.Json
}

func NewEksNodeGroupRoleTrustPolicy(ctx *pulumi.Context) string {
	policy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"sts:AssumeRole",
				},
				Principals: []iam.GetPolicyDocumentStatementPrincipal{
					{
						Type: "Service",
						Identifiers: []string{"ec2.amazonaws.com"},
					},
				},
			},
		},
	})
	if err != nil { return "" }

	return policy.Json
}

type EksIamRoles struct {
	Admin *iam.Role
	Automation *iam.Role
	NodeGroup *iam.Role
}
func NewEksRoles(ctx *pulumi.Context, accountId string) (EksIamRoles, error) {
	var ret EksIamRoles
	adminRole, err := iam.NewRole(ctx, "eks-cluster-admin-role", &iam.RoleArgs{
		Name: pulumi.String("eksClusterAdminRole"),
		Description: pulumi.String("Admin role for eks cluster"),
		AssumeRolePolicy: pulumi.String(NewEksClusterAdminRoleTrustPolicy(ctx, accountId)),
		InlinePolicies: iam.RoleInlinePolicyArray{
			iam.RoleInlinePolicyArgs{
				Name:   pulumi.String("eks-admin-policy"),
				Policy: pulumi.String(NewEksClusterAdminPolicy(ctx)),
			},
		},
	})
	if err != nil { return ret, err }

	automationRole, err := iam.NewRole(ctx, "eks-automation-role", &iam.RoleArgs{
		Name: pulumi.String("eksAutomationRole"),
		Description: pulumi.String("Automation role for eks cluster"),
		AssumeRolePolicy: pulumi.String(NewEksAutomationRoleTrustPolicy(ctx, accountId)),
	})
	if err != nil { return ret, err }

	nodeGroupRole, err := iam.NewRole(ctx, "eks-cluster-node-group-role", &iam.RoleArgs{
		Name: pulumi.String("eksNodeGroupRole"),
		Description: pulumi.String("Node group role for eks cluster"),
		AssumeRolePolicy: pulumi.String(NewEksNodeGroupRoleTrustPolicy(ctx)),
		ManagedPolicyArns: pulumi.StringArray{
			pulumi.String("arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"),
			pulumi.String("arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"),
			pulumi.String("arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"),
		},
		InlinePolicies: iam.RoleInlinePolicyArray{
			iam.RoleInlinePolicyArgs{
				Name:   pulumi.String("fluentd-cloudwatch-policy"),
				Policy: pulumi.String(NewFluentdCloudwatchPolicy(ctx)),
			},
		},
	})
	if err != nil { return ret, err }

	ret.Admin = adminRole
	ret.Automation = automationRole
	ret.NodeGroup = nodeGroupRole
	return ret, nil
}