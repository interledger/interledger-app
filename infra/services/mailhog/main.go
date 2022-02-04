package mailhog

import (
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	rbacv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/services/ingress"
)

func DeployMailHog(ctx *pulumi.Context, domain string) error {
	_, err := appsv1.NewDeployment(ctx, "mailhog-deployment", &appsv1.DeploymentArgs{
		ApiVersion: pulumi.String("apps/v1"),
		Kind:       pulumi.String("Deployment"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("mailhog"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("mailhog"),
			},
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"app": pulumi.String("mailhog"),
				},
			},
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: pulumi.StringMap{
						"app": pulumi.String("mailhog"),
					},
				},
				Spec: &corev1.PodSpecArgs{
					Containers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:  pulumi.String("mailhog"),
							Image: pulumi.String("mailhog/mailhog"),
							Ports: &corev1.ContainerPortArray{
								corev1.ContainerPortArgs{
									ContainerPort: pulumi.Int(1025),
								},
								corev1.ContainerPortArgs{
									ContainerPort: pulumi.Int(8025),
								},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	_, err = corev1.NewServiceAccount(ctx, "mailhog-service-account", &corev1.ServiceAccountArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("ServiceAccount"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("mailhog"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("mailhog"),
			},
		},
		AutomountServiceAccountToken: pulumi.Bool(false),
	})
	if err != nil {
		return err
	}
	_, err = rbacv1.NewRole(ctx, "mailhog-role", &rbacv1.RoleArgs{
		ApiVersion: pulumi.String("rbac.authorization.k8s.io/v1"),
		Kind:       pulumi.String("Role"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("mailhog"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("mailhog"),
			},
		},
	})
	if err != nil {
		return err
	}
	_, err = rbacv1.NewRoleBinding(ctx, "mailhog-role-binding", &rbacv1.RoleBindingArgs{
		ApiVersion: pulumi.String("rbac.authorization.k8s.io/v1"),
		Kind:       pulumi.String("RoleBinding"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("mailhog"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("mailhog"),
			},
		},
		RoleRef: &rbacv1.RoleRefArgs{
			ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
			Kind:     pulumi.String("Role"),
			Name:     pulumi.String("mailhog"),
		},
		Subjects: rbacv1.SubjectArray{
			&rbacv1.SubjectArgs{
				Kind:      pulumi.String("ServiceAccount"),
				Name:      pulumi.String("mailhog"),
				Namespace: pulumi.String("default"),
			},
		},
	})
	if err != nil {
		return err
	}

	_, err = corev1.NewService(ctx, "mailhog-service", &corev1.ServiceArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Service"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("mailhog"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("mailhog"),
			},
		},
		Spec: &corev1.ServiceSpecArgs{
			Ports: corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(1025),
					TargetPort: pulumi.Int(1025),
					Name:       pulumi.String("smtp"),
				},
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(80),
					TargetPort: pulumi.Int(8025),
					Name:       pulumi.String("http"),
				},
			},
			Selector: pulumi.StringMap{
				"app": pulumi.String("mailhog"),
			},
		},
	})
	if err != nil {
		return err
	}

	err = ingress.DeployMapping(ctx, &ingress.MappingArgs{
		Name:            "mailhog-mapping",
		Hostname:        domain,
		Prefix:          "/",
		Rewrite:         "/",
		Service:         "mailhog",
		EnableWebsocket: true,
	})
	if err != nil {
		return err
	}

	return nil
}
