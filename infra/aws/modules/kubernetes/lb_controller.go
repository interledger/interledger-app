package kubernetes

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"io"
	"net/http"
)

type CreateLbControllerRoleArgs struct {
	AccountId      pulumi.StringPtrInput
	OidcProvider   pulumi.StringPtrInput
	Namespace      pulumi.StringPtrInput
	ServiceAccount pulumi.StringPtrInput
}

func CreateLbControllerRole(ctx *pulumi.Context, args CreateLbControllerRoleArgs) (*iam.Role, error) {

	policy, err := lbRolePolicy()
	if err != nil {
		return nil, err
	}

	trustPolicy := NewIamTrustPolicyDocumentV2(ctx, args.AccountId, args.OidcProvider, args.Namespace, args.ServiceAccount)

	role, err := iam.NewRole(ctx, "lb-controller-role", &iam.RoleArgs{
		Name:        pulumi.String("eks-shared-lb-controller-role"),
		Description: pulumi.String("AWS LB Controller role for the shared EKS cluster"),
		InlinePolicies: iam.RoleInlinePolicyArray{
			iam.RoleInlinePolicyArgs{
				Name:   pulumi.String("lbRole"),
				Policy: pulumi.String(policy),
			},
		},
		AssumeRolePolicy: trustPolicy,
	})

	return role, err
}

func lbRolePolicy() (string, error) {
	resp, err := http.Get("https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/v2.4.1/docs/install/iam_policy.json")
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
