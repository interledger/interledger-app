package temporal

import (
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func DeployTemporalDev(ctx *pulumi.Context, imageRepo string, imageTag string) error {
	err := deployment(ctx, imageRepo, imageTag)
	if err != nil {
		return err
	}

	err = service(ctx)
	if err != nil {
		return err
	}

	return nil
}

func deployment(ctx *pulumi.Context, imageRepo string, imageTag string) error {
	_, err := appsv1.NewDeployment(ctx, "temporal-dev", &appsv1.DeploymentArgs{
		ApiVersion: pulumi.String("apps/v1"),
		Kind:       pulumi.String("Deployment"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("temporal"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("temporal"),
			},
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"app": pulumi.String("temporal"),
				},
			},
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: pulumi.StringMap{
						"app": pulumi.String("temporal"),
					},
				},
				Spec: &corev1.PodSpecArgs{
					Containers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:            pulumi.String("temporal"),
							Image:           pulumi.Sprintf("%s/temporalite:%s", imageRepo, imageTag),
							ImagePullPolicy: pulumi.String("Always"),
							Ports: corev1.ContainerPortArray{
								&corev1.ContainerPortArgs{
									ContainerPort: pulumi.Int(7233),
									Name:          pulumi.String("grpc"),
								},
								&corev1.ContainerPortArgs{
									ContainerPort: pulumi.Int(8233),
									Name:          pulumi.String("http"),
								},
							},
						},
					},
					TerminationGracePeriodSeconds: pulumi.Int(30),
				},
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func service(ctx *pulumi.Context) error {
	_, err := corev1.NewService(ctx, "temporal-service", &corev1.ServiceArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Service"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("temporal"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("temporal"),
			},
		},
		Spec: &corev1.ServiceSpecArgs{
			Ports: corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(7233),
					TargetPort: pulumi.Int(7233),
					Name:       pulumi.String("grpc"),
				},
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(80),
					TargetPort: pulumi.Int(8233),
					Name:       pulumi.String("http"),
				},
			},
			Selector: pulumi.StringMap{
				"app": pulumi.String("temporal"),
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}
