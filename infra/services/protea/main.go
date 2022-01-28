package protea

import (
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	rbacv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/services/ingress"
)

type DeployProteaArgs struct {
	ImageRepo string
	ImageTag  string
}

func DeployProtea(ctx *pulumi.Context, args DeployProteaArgs) error {

	err := deployService(ctx)
	if err != nil {
		return err
	}
	err = deployIngress(ctx)
	if err != nil {
		return err
	}
	err = deployRbac(ctx)
	if err != nil {
		return err
	}
	err = deployDeployment(ctx, args.ImageRepo, args.ImageTag)
	if err != nil {
		return err
	}

	return nil
}

func deployService(ctx *pulumi.Context) error {
	_, err := corev1.NewService(ctx, "protea-service", &corev1.ServiceArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Service"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("protea"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("protea"),
			},
		},
		Spec: &corev1.ServiceSpecArgs{
			Ports: corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(80),
					TargetPort: pulumi.Int(3000),
					Name:       pulumi.String("http"),
				},
			},
			Selector: pulumi.StringMap{
				"app": pulumi.String("protea"),
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func deployDeployment(ctx *pulumi.Context, imageRepo string, imageTag string) error {
	_, err := appsv1.NewDeployment(ctx, "protea-deployment", &appsv1.DeploymentArgs{
		ApiVersion: pulumi.String("apps/v1"),
		Kind:       pulumi.String("Deployment"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("protea"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("protea"),
			},
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"app": pulumi.String("protea"),
				},
			},
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: pulumi.StringMap{
						"app": pulumi.String("protea"),
					},
				},
				Spec: &corev1.PodSpecArgs{
					Containers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:  pulumi.String("protea"),
							Image: pulumi.Sprintf("%s/protea:%s", imageRepo, imageTag),
							Ports: corev1.ContainerPortArray{
								&corev1.ContainerPortArgs{
									ContainerPort: pulumi.Int(3000),
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
	return nil
}

func deployRbac(ctx *pulumi.Context) error {
	_, err := corev1.NewServiceAccount(ctx, "protea-sa", &corev1.ServiceAccountArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("ServiceAccount"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("protea"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("protea"),
			},
		},
		AutomountServiceAccountToken: pulumi.Bool(false),
	})
	if err != nil {
		return err
	}
	_, err = rbacv1.NewRole(ctx, "protea-role", &rbacv1.RoleArgs{
		ApiVersion: pulumi.String("rbac.authorization.k8s.io/v1"),
		Kind:       pulumi.String("Role"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("protea"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("protea"),
			},
		},
	})
	if err != nil {
		return err
	}
	_, err = rbacv1.NewRoleBinding(ctx, "protea-rb", &rbacv1.RoleBindingArgs{
		ApiVersion: pulumi.String("rbac.authorization.k8s.io/v1"),
		Kind:       pulumi.String("RoleBinding"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("protea"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("protea"),
			},
		},
		RoleRef: &rbacv1.RoleRefArgs{
			ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
			Kind:     pulumi.String("Role"),
			Name:     pulumi.String("protea"),
		},
		Subjects: rbacv1.SubjectArray{
			&rbacv1.SubjectArgs{
				Kind:      pulumi.String("ServiceAccount"),
				Name:      pulumi.String("protea"),
				Namespace: pulumi.String("default"),
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func deployIngress(ctx *pulumi.Context) error {
	err := ingress.DeployMapping(ctx, &ingress.MappingArgs{
		Name:            "protea-mapping",
		Hostname:        "*",
		Prefix:          "/",
		Service:         "protea",
		EnableWebsocket: true,
	})
	if err != nil {
		return err
	}

	return nil
}
