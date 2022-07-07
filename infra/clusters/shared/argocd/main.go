package main

import (
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"gitlab.com/fynbos/infra/aws/modules/utils"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		k8sStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-shared-eu-west-1-shared-k8s/main", nil)
		if err != nil {
			return err
		}
		kubeConfig := k8sStack.GetStringOutput(pulumi.String("kubeconfig"))
		kubeProvider, err := kubernetes.NewProvider(ctx, "kubernetes-provider", &kubernetes.ProviderArgs{
			Kubeconfig: kubeConfig,
			Namespace:  pulumi.String("argocd"),
		})
		if err != nil {
			return err
		}
		oidcProvider := k8sStack.GetStringOutput(pulumi.String("oidcProvider"))
		cfg := config.New(ctx, "fynbos")
		accountID := cfg.Require("accountId")

		namespace, err := corev1.NewNamespace(ctx, "namespace", &corev1.NamespaceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name: pulumi.String("argocd"),
			},
		}, pulumi.Provider(kubeProvider))
		if err != nil {
			return err
		}

		role, err := newArgoCdRole(ctx, pulumi.String(accountID), oidcProvider, namespace.Metadata.Name().Elem())
		if err != nil {
			return err
		}

		_, err = yaml.NewConfigFile(ctx, "argo", &yaml.ConfigFileArgs{
			File: "./argocd.yaml",
			Transformations: []yaml.Transformation{
				// Set the aws role arn on the correct service accounts
				func(state map[string]interface{}, opts ...pulumi.ResourceOption) {
					name := state["metadata"].(map[string]interface{})["name"]
					if state["kind"] == "ServiceAccount" && (name == "argocd-server" || name == "argocd-application-controller") {
						metadata := state["metadata"].(map[string]interface{})
						annotations := make(map[string]interface{})
						annotations["eks.amazonaws.com/role-arn"] = role.Arn
						metadata["annotations"] = annotations
					}
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{role, namespace}))
		if err != nil {
			return err
		}

		ctx.Export("argoRoleArn", role.Arn)
		return nil
	})
}

func newArgoCdRole(ctx *pulumi.Context, accountId pulumi.StringPtrInput, oidcProvider pulumi.StringPtrInput, namespace pulumi.StringPtrInput) (*iam.Role, error) {

	policy, err := newRolePolicy(ctx)
	if err != nil {
		return nil, err
	}
	trustPolicy := newTrustPolicy(ctx, accountId, oidcProvider, namespace)

	role, err := iam.NewRole(ctx, "argocd-role", &iam.RoleArgs{
		Name:             pulumi.String("eksArgoRole"),
		Description:      pulumi.String(""),
		AssumeRolePolicy: trustPolicy,
		InlinePolicies: iam.RoleInlinePolicyArray{
			iam.RoleInlinePolicyArgs{
				Name:   pulumi.String("assumePolicy"),
				Policy: pulumi.String(policy),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return role, err
}

func newRolePolicy(ctx *pulumi.Context) (string, error) {
	policy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"sts:AssumeRole",
				},
				Resources: []string{
					"*",
				},
			},
		},
	})
	if err != nil {
		return "", err
	}

	return policy.Json, nil
}

// Need a custom one as it has different requirements for the conditions
func newTrustPolicy(ctx *pulumi.Context, accountId pulumi.StringPtrInput, oidcProvider pulumi.StringPtrInput, namespace pulumi.StringPtrInput) pulumi.StringOutput {
	return pulumi.All(accountId, oidcProvider, namespace).ApplyT(func(args []interface{}) (string, error) {
		aid := args[0].(string)
		oidc := args[1].(string)
		ns := args[2].(string)

		conditions := []iam.GetPolicyDocumentStatementCondition{
			{
				Test:     "StringEquals",
				Values:   []string{"sts.amazonaws.com"},
				Variable: fmt.Sprintf("%s:aud", oidc),
			},
		}

		conditions = append(conditions, iam.GetPolicyDocumentStatementCondition{
			Test: "StringLike",
			Values: []string{
				fmt.Sprintf("system:serviceaccount:%s:*", ns),
			},
			Variable: fmt.Sprintf("%s:sub", oidc),
		})

		policy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
			Statements: []iam.GetPolicyDocumentStatement{
				{
					Effect: utils.StringPtr("Allow"),
					Actions: []string{
						"sts:AssumeRoleWithWebIdentity",
					},
					Principals: []iam.GetPolicyDocumentStatementPrincipal{
						{
							Type: "Federated",
							Identifiers: []string{
								fmt.Sprintf("arn:aws:iam::%s:oidc-provider/%s", aid, oidc),
							},
						},
					},
					Conditions: conditions,
				},
			},
		})
		if err != nil {
			return "", err
		}

		return policy.Json, nil
	}).(pulumi.StringOutput)
}
