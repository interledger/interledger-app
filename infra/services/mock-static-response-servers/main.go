package mockstaticresponseservers

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"runtime"

	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func DeployMockGnapServer(ctx *pulumi.Context, name string) error {
	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok {
		return errors.New("could not get directory path for kratos module")
	}

	dat, err := ioutil.ReadFile(filepath.Join(filepath.Dir(moduleDir), "Caddyfile"))
	if err != nil {
		return err
	}

	caddyConfig, err := corev1.NewConfigMap(ctx, "caddy-config", &corev1.ConfigMapArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("caddy-config"),
		},
		Data: pulumi.StringMap{"Caddyfile": pulumi.String(dat)},
	})
	if err != nil {
		return err
	}

	_, err = corev1.NewService(ctx, name+"-service", &corev1.ServiceArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Service"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(name),
			Labels: pulumi.StringMap{
				"app": pulumi.String(name),
			},
		},
		Spec: &corev1.ServiceSpecArgs{
			Ports: corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(80),
					TargetPort: pulumi.Int(8080),
					Name:       pulumi.String("http"),
				},
			},
			Selector: pulumi.StringMap{
				"app": pulumi.String(name),
			},
		},
	})
	if err != nil {
		return err
	}

	_, err = appsv1.NewDeployment(ctx, name+"-deployment", &appsv1.DeploymentArgs{
		ApiVersion: pulumi.String("apps/v1"),
		Kind:       pulumi.String("Deployment"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(name),
			Labels: pulumi.StringMap{
				"app": pulumi.String(name),
			},
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"app": pulumi.String(name),
				},
			},
			Strategy: &appsv1.DeploymentStrategyArgs{
				Type: pulumi.String("RollingUpdate"),
				RollingUpdate: &appsv1.RollingUpdateDeploymentArgs{
					MaxSurge:       pulumi.Int(2),
					MaxUnavailable: pulumi.Int(1),
				},
			},
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: pulumi.StringMap{
						"app": pulumi.String(name),
					},
				},
				Spec: &corev1.PodSpecArgs{
					SecurityContext: &corev1.PodSecurityContextArgs{
						RunAsUser: pulumi.Int(65532),
					},
					Volumes: corev1.VolumeArray{
						&corev1.VolumeArgs{
							Name: pulumi.String("caddy"),
							ConfigMap: &corev1.ConfigMapVolumeSourceArgs{
								Name: caddyConfig.Metadata.Name(),
								Items: corev1.KeyToPathArray{
									&corev1.KeyToPathArgs{
										Key:  pulumi.String("Caddyfile"),
										Path: pulumi.String("Caddyfile"),
									},
								},
							},
						},
					},
					Containers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:            pulumi.String("backend"),
							Image:           pulumi.Sprintf("caddy:latest"),
							ImagePullPolicy: pulumi.String("Always"),
							Ports: corev1.ContainerPortArray{
								&corev1.ContainerPortArgs{
									ContainerPort: pulumi.Int(8080),
									Name:          pulumi.String("http"),
								},
							},
							LivenessProbe: &corev1.ProbeArgs{
								HttpGet: &corev1.HTTPGetActionArgs{
									Path:   pulumi.String("/healthz"),
									Port:   pulumi.String("http"),
									Scheme: pulumi.String("HTTP"),
								},
								PeriodSeconds: pulumi.Int(5),
							},
							ReadinessProbe: &corev1.ProbeArgs{
								HttpGet: &corev1.HTTPGetActionArgs{
									Path:   pulumi.String("/healthz"),
									Port:   pulumi.String("http"),
									Scheme: pulumi.String("HTTP"),
								},
								PeriodSeconds: pulumi.Int(5),
							},
							Env: corev1.EnvVarArray{},
							VolumeMounts: corev1.VolumeMountArray{
								&corev1.VolumeMountArgs{
									Name:      pulumi.String("caddy"),
									MountPath: pulumi.String("/etc/caddy"),
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
