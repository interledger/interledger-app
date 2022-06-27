package crdb

import (
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	policyv1beta1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/policy/v1beta1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func DeployPodDistributionBudget(ctx *pulumi.Context, opts ...pulumi.ResourceOption) error {
	_, err := policyv1beta1.NewPodDisruptionBudget(ctx, "crd-pdb", &policyv1beta1.PodDisruptionBudgetArgs{
		ApiVersion: pulumi.String("policy/v1beta1"),
		Kind:       pulumi.String("PodDisruptionBudget"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("cockroachdb-budget"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("cockroachdb"),
			},
		},
		Spec: &policyv1beta1.PodDisruptionBudgetSpecArgs{
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"app": pulumi.String("cockroachdb"),
				},
			},
			MaxUnavailable: pulumi.Int(1),
		},
	}, opts...)
	if err != nil {
		return err
	}
	return nil
}
