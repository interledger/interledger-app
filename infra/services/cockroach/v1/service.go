package crdb

import (
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func DeployPublicService(ctx *pulumi.Context, namespace pulumi.StringPtrInput, opts ...pulumi.ResourceOption) error {
	_, err := corev1.NewService(ctx, "crdb-public-service", &corev1.ServiceArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Service"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("cockroachdb-public"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("cockroachdb"),
			},
			Namespace: namespace,
		},
		Spec: &corev1.ServiceSpecArgs{
			Ports: corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(26257),
					TargetPort: pulumi.Int(26257),
					Name:       pulumi.String("grpc"),
				},
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(8080),
					TargetPort: pulumi.Int(8080),
					Name:       pulumi.String("http"),
				},
			},
			Selector: pulumi.StringMap{
				"app": pulumi.String("cockroachdb"),
			},
		},
	}, opts...)
	if err != nil {
		return err
	}
	return nil
}

func DeployPrivateService(ctx *pulumi.Context, namespace pulumi.StringPtrInput, opts ...pulumi.ResourceOption) error {
	_, err := corev1.NewService(ctx, "crdb-private-service", &corev1.ServiceArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Service"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("cockroachdb"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("cockroachdb"),
			},
			Namespace: namespace,
			Annotations: pulumi.StringMap{
				"service.alpha.kubernetes.io/tolerate-unready-endpoints": pulumi.String("true"),
				"prometheus.io/scrape": pulumi.String("true"),
				"prometheus.io/path":   pulumi.String("_status/vars"),
				"prometheus.io/port":   pulumi.String("8080"),
			},
		},
		Spec: &corev1.ServiceSpecArgs{
			Ports: corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(26257),
					TargetPort: pulumi.Int(26257),
					Name:       pulumi.String("grpc"),
				},
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(8080),
					TargetPort: pulumi.Int(8080),
					Name:       pulumi.String("http"),
				},
			},
			PublishNotReadyAddresses: pulumi.Bool(true),
			ClusterIP:                pulumi.String("None"),
			Selector: pulumi.StringMap{
				"app": pulumi.String("cockroachdb"),
			},
		},
	}, opts...)
	if err != nil {
		return err
	}
	return nil
}
