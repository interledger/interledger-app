package crdb

import (
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	rbacv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func DeployRbac(ctx *pulumi.Context, namespace pulumi.StringPtrInput, opts ...pulumi.ResourceOption) (*corev1.ServiceAccount, error) {
	serviceAccount, err := corev1.NewServiceAccount(ctx, "crdb-sa", &corev1.ServiceAccountArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("ServiceAccount"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("cockroachdb"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("cockroachdb"),
			},
			Namespace: namespace,
		},
		AutomountServiceAccountToken: pulumi.Bool(false),
	}, opts...)
	if err != nil {
		return nil, err
	}

	role, err := rbacv1.NewRole(ctx, "crdb-role", &rbacv1.RoleArgs{
		ApiVersion: pulumi.String("rbac.authorization.k8s.io/v1"),
		Kind:       pulumi.String("Role"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("cockroachdb"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("cockroachdb"),
			},
			Namespace: namespace,
		},
		Rules: rbacv1.PolicyRuleArray{
			&rbacv1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{
					pulumi.String(""),
				},
				Resources: pulumi.StringArray{
					pulumi.String("secrets"),
				},
				Verbs: pulumi.StringArray{
					pulumi.String("get"),
				},
			},
		},
	}, opts...)
	if err != nil {
		return nil, err
	}

	opts = append(opts, pulumi.DependsOn([]pulumi.Resource{role, serviceAccount}))
	_, err = rbacv1.NewRoleBinding(ctx, "crdb-role-binding", &rbacv1.RoleBindingArgs{
		ApiVersion: pulumi.String("rbac.authorization.k8s.io/v1"),
		Kind:       pulumi.String("RoleBinding"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("cockroachdb"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("cockroachdb"),
			},
			Namespace: namespace,
		},
		RoleRef: &rbacv1.RoleRefArgs{
			ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
			Kind:     pulumi.String("Role"),
			Name:     role.Metadata.Name().Elem(),
		},
		Subjects: rbacv1.SubjectArray{
			&rbacv1.SubjectArgs{
				Kind:      pulumi.String("ServiceAccount"),
				Name:      serviceAccount.Metadata.Name().Elem(),
				Namespace: serviceAccount.Metadata.Namespace(),
			},
		},
	}, opts...)
	if err != nil {
		return nil, err
	}

	return serviceAccount, nil
}
