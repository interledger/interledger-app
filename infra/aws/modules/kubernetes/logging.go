package kubernetes

import (
	"errors"
	"path/filepath"
	"runtime"

	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	rbacv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/aws/modules/utils"
)

// We use Fluentbit as a log aggregrator as it is more performant that Fluentd and is recommended by AWS.
// It will send the logs to Cloudwatch. The deployment will create the log groups if it doesn't exist.
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Container-Insights-setup-logs-FluentBit.html
func DeployFluentbit(ctx *pulumi.Context, clusterName string, region string, namespace string) error {
	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok { return errors.New("Could not get directory path for kubernetes module.") }

	type FluentBitConfig struct {
		HTTP_SERVER 	string
		HTTP_PORT   	string
		READ_FROM_HEAD 	string
		READ_FROM_TAIL 	string
		AWS_REGION 		string
		CLUSTER_NAME 	string
	}
	config := FluentBitConfig{
		HTTP_SERVER: "on",
		HTTP_PORT: "2020",
		READ_FROM_HEAD: "off",
		READ_FROM_TAIL: "on",
		AWS_REGION: region,
		CLUSTER_NAME: clusterName,
	}

	_, err := corev1.NewConfigMap(ctx, "fluent-bit-cluster-info-configmap", &corev1.ConfigMapArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("ConfigMap"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("fluent-bit-cluster-info"),
			Namespace: pulumi.String(namespace),
		},
		Data: pulumi.StringMap{
			"cluster.name": pulumi.String(clusterName),
			"http.server":  pulumi.String(config.HTTP_SERVER),
			"http.port":    pulumi.String(config.HTTP_PORT),
			"read.head":    pulumi.String(config.READ_FROM_HEAD),
			"read.tail":    pulumi.String(config.READ_FROM_TAIL),
			"logs.region":  pulumi.String(region),
		},
	})
	if err != nil { return err }

	// https://raw.githubusercontent.com/aws-samples/amazon-cloudwatch-container-insights/latest/k8s-deployment-manifest-templates/deployment-mode/daemonset/container-insights-monitoring/fluent-bit/fluent-bit.yaml
	_, err = corev1.NewServiceAccount(ctx, "fluent-bit-service-account", &corev1.ServiceAccountArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("ServiceAccount"),
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("fluent-bit"),
			Namespace: pulumi.String(namespace),
		},
	})
	if err != nil { return err }

	_, err = rbacv1.NewClusterRole(ctx, "fluent-bit-role-cluster-role", &rbacv1.ClusterRoleArgs{
		ApiVersion: pulumi.String("rbac.authorization.k8s.io/v1"),
		Kind:       pulumi.String("ClusterRole"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("fluent-bit-role"),
		},
		Rules: rbacv1.PolicyRuleArray{
			&rbacv1.PolicyRuleArgs{
				NonResourceURLs: pulumi.StringArray{
					pulumi.String("/metrics"),
				},
				Verbs: pulumi.StringArray{
					pulumi.String("get"),
				},
			},
			&rbacv1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{
					pulumi.String(""),
				},
				Resources: pulumi.StringArray{
					pulumi.String("namespaces"),
					pulumi.String("pods"),
					pulumi.String("pods/logs"),
				},
				Verbs: pulumi.StringArray{
					pulumi.String("get"),
					pulumi.String("list"),
					pulumi.String("watch"),
				},
			},
		},
	})
	if err != nil { return err }

	_, err = rbacv1.NewClusterRoleBinding(ctx, "fluent-bit-cluster-role-binding", &rbacv1.ClusterRoleBindingArgs{
		ApiVersion: pulumi.String("rbac.authorization.k8s.io/v1"),
		Kind:       pulumi.String("ClusterRoleBinding"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("fluent-bit-role-binding"),
		},
		RoleRef: &rbacv1.RoleRefArgs{
			ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
			Kind:     pulumi.String("ClusterRole"),
			Name:     pulumi.String("fluent-bit-role"),
		},
		Subjects: rbacv1.SubjectArray{
			&rbacv1.SubjectArgs{
				Kind:      pulumi.String("ServiceAccount"),
				Name:      pulumi.String("fluent-bit"),
				Namespace: pulumi.String(namespace),
			},
		},
	})
	if err != nil { return err }


	fluentbitConf, err := utils.ParseTemplateAsBytes(config, filepath.Join(filepath.Dir(moduleDir), "fluentbit/fluent-bit.conf"))
	if err != nil { return err }
	applogConf, err := utils.ParseTemplateAsBytes(config, filepath.Join(filepath.Dir(moduleDir), "fluentbit/application-log.conf"))
	if err != nil { return err }
	dataplaneConf, err := utils.ParseTemplateAsBytes(config, filepath.Join(filepath.Dir(moduleDir), "fluentbit/dataplane-log.conf"))
	if err != nil { return err }
	hostlogConf, err := utils.ParseTemplateAsBytes(config, filepath.Join(filepath.Dir(moduleDir), "fluentbit/host-log.conf"))
	if err != nil { return err }
	parserConf, err := utils.ParseTemplateAsBytes(config, filepath.Join(filepath.Dir(moduleDir), "fluentbit/parsers.conf"))
	if err != nil { return err }
	_, err = corev1.NewConfigMap(ctx, "fluent-bit-config-map", &corev1.ConfigMapArgs{
			ApiVersion: pulumi.String("v1"),
			Kind:       pulumi.String("ConfigMap"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("fluent-bit-config"),
				Namespace: pulumi.String(namespace),
				Labels: pulumi.StringMap{
					"k8s-app": pulumi.String("fluent-bit"),
				},
			},
			Data: pulumi.StringMap{
				"fluent-bit.conf":      pulumi.String(fluentbitConf.String()),
				"application-log.conf": pulumi.String(applogConf.String()),
				"dataplane-log.conf":   pulumi.String(dataplaneConf.String()),
				"host-log.conf":   		pulumi.String(hostlogConf.String()),
				"parsers.conf":   		pulumi.String(parserConf.String()),
			},
		})

	if err != nil { return err }

	_, err = appsv1.NewDaemonSet(ctx, "fluent-bit-daemon-set", &appsv1.DaemonSetArgs{
			ApiVersion: pulumi.String("apps/v1"),
			Kind:       pulumi.String("DaemonSet"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("fluent-bit"),
				Namespace: pulumi.String(namespace),
				Labels: pulumi.StringMap{
					"k8s-app":                       pulumi.String("fluent-bit"),
					"version":                       pulumi.String("v1"),
					"kubernetes.io/cluster-service": pulumi.String("true"),
				},
			},
			Spec: &appsv1.DaemonSetSpecArgs{
				Selector: &metav1.LabelSelectorArgs{
					MatchLabels: pulumi.StringMap{
						"k8s-app": pulumi.String("fluent-bit"),
					},
				},
				Template: &corev1.PodTemplateSpecArgs{
					Metadata: &metav1.ObjectMetaArgs{
						Labels: pulumi.StringMap{
							"k8s-app":                       pulumi.String("fluent-bit"),
							"version":                       pulumi.String("v1"),
							"kubernetes.io/cluster-service": pulumi.String("true"),
						},
					},
					Spec: &corev1.PodSpecArgs{
						Containers: corev1.ContainerArray{
							&corev1.ContainerArgs{
								Name:            pulumi.String("fluent-bit"),
								Image:           pulumi.String("amazon/aws-for-fluent-bit:2.10.0"),
								ImagePullPolicy: pulumi.String("Always"),
								Env: corev1.EnvVarArray{
									&corev1.EnvVarArgs{
										Name: pulumi.String("AWS_REGION"),
										ValueFrom: &corev1.EnvVarSourceArgs{
											ConfigMapKeyRef: &corev1.ConfigMapKeySelectorArgs{
												Name: pulumi.String("fluent-bit-cluster-info"),
												Key:  pulumi.String("logs.region"),
											},
										},
									},
									&corev1.EnvVarArgs{
										Name: pulumi.String("CLUSTER_NAME"),
										ValueFrom: &corev1.EnvVarSourceArgs{
											ConfigMapKeyRef: &corev1.ConfigMapKeySelectorArgs{
												Name: pulumi.String("fluent-bit-cluster-info"),
												Key:  pulumi.String("cluster.name"),
											},
										},
									},
									&corev1.EnvVarArgs{
										Name: pulumi.String("HTTP_SERVER"),
										ValueFrom: &corev1.EnvVarSourceArgs{
											ConfigMapKeyRef: &corev1.ConfigMapKeySelectorArgs{
												Name: pulumi.String("fluent-bit-cluster-info"),
												Key:  pulumi.String("http.server"),
											},
										},
									},
									&corev1.EnvVarArgs{
										Name: pulumi.String("HTTP_PORT"),
										ValueFrom: &corev1.EnvVarSourceArgs{
											ConfigMapKeyRef: &corev1.ConfigMapKeySelectorArgs{
												Name: pulumi.String("fluent-bit-cluster-info"),
												Key:  pulumi.String("http.port"),
											},
										},
									},
									&corev1.EnvVarArgs{
										Name: pulumi.String("READ_FROM_HEAD"),
										ValueFrom: &corev1.EnvVarSourceArgs{
											ConfigMapKeyRef: &corev1.ConfigMapKeySelectorArgs{
												Name: pulumi.String("fluent-bit-cluster-info"),
												Key:  pulumi.String("read.head"),
											},
										},
									},
									&corev1.EnvVarArgs{
										Name: pulumi.String("READ_FROM_TAIL"),
										ValueFrom: &corev1.EnvVarSourceArgs{
											ConfigMapKeyRef: &corev1.ConfigMapKeySelectorArgs{
												Name: pulumi.String("fluent-bit-cluster-info"),
												Key:  pulumi.String("read.tail"),
											},
										},
									},
									&corev1.EnvVarArgs{
										Name: pulumi.String("HOST_NAME"),
										ValueFrom: &corev1.EnvVarSourceArgs{
											FieldRef: &corev1.ObjectFieldSelectorArgs{
												FieldPath: pulumi.String("spec.nodeName"),
											},
										},
									},
									&corev1.EnvVarArgs{
										Name:  pulumi.String("CI_VERSION"),
										Value: pulumi.String("k8s/1.3.8"),
									},
								},
								Resources: &corev1.ResourceRequirementsArgs{
									Limits: pulumi.StringMap{
										"memory": pulumi.String("100Mi"),
										"cpu": pulumi.String("100m"),
									},
									Requests: pulumi.StringMap{
										"cpu":    pulumi.String("50m"),
										"memory": pulumi.String("20Mi"),
									},
								},
								VolumeMounts: corev1.VolumeMountArray{
									&corev1.VolumeMountArgs{
										Name:      pulumi.String("fluentbitstate"),
										MountPath: pulumi.String("/var/fluent-bit/state"),
									},
									&corev1.VolumeMountArgs{
										Name:      pulumi.String("varlog"),
										MountPath: pulumi.String("/var/log"),
										ReadOnly:  pulumi.Bool(true),
									},
									&corev1.VolumeMountArgs{
										Name:      pulumi.String("varlibdockercontainers"),
										MountPath: pulumi.String("/var/lib/docker/containers"),
										ReadOnly:  pulumi.Bool(true),
									},
									&corev1.VolumeMountArgs{
										Name:      pulumi.String("fluent-bit-config"),
										MountPath: pulumi.String("/fluent-bit/etc/"),
									},
									&corev1.VolumeMountArgs{
										Name:      pulumi.String("runlogjournal"),
										MountPath: pulumi.String("/run/log/journal"),
										ReadOnly:  pulumi.Bool(true),
									},
									&corev1.VolumeMountArgs{
										Name:      pulumi.String("dmesg"),
										MountPath: pulumi.String("/var/log/dmesg"),
										ReadOnly:  pulumi.Bool(true),
									},
								},
							},
						},
						TerminationGracePeriodSeconds: pulumi.Int(10),
						Volumes: corev1.VolumeArray{
							&corev1.VolumeArgs{
								Name: pulumi.String("fluentbitstate"),
								HostPath: &corev1.HostPathVolumeSourceArgs{
									Path: pulumi.String("/var/fluent-bit/state"),
								},
							},
							&corev1.VolumeArgs{
								Name: pulumi.String("varlog"),
								HostPath: &corev1.HostPathVolumeSourceArgs{
									Path: pulumi.String("/var/log"),
								},
							},
							&corev1.VolumeArgs{
								Name: pulumi.String("varlibdockercontainers"),
								HostPath: &corev1.HostPathVolumeSourceArgs{
									Path: pulumi.String("/var/lib/docker/containers"),
								},
							},
							&corev1.VolumeArgs{
								Name: pulumi.String("fluent-bit-config"),
								ConfigMap: &corev1.ConfigMapVolumeSourceArgs{
									Name: pulumi.String("fluent-bit-config"),
								},
							},
							&corev1.VolumeArgs{
								Name: pulumi.String("runlogjournal"),
								HostPath: &corev1.HostPathVolumeSourceArgs{
									Path: pulumi.String("/run/log/journal"),
								},
							},
							&corev1.VolumeArgs{
								Name: pulumi.String("dmesg"),
								HostPath: &corev1.HostPathVolumeSourceArgs{
									Path: pulumi.String("/var/log/dmesg"),
								},
							},
						},
						ServiceAccountName: pulumi.String("fluent-bit"),
						Tolerations: corev1.TolerationArray{
							&corev1.TolerationArgs{
								Key:      pulumi.String("node-role.kubernetes.io/master"),
								Operator: pulumi.String("Exists"),
								Effect:   pulumi.String("NoSchedule"),
							},
							&corev1.TolerationArgs{
								Operator: pulumi.String("Exists"),
								Effect:   pulumi.String("NoExecute"),
							},
							&corev1.TolerationArgs{
								Operator: pulumi.String("Exists"),
								Effect:   pulumi.String("NoSchedule"),
							},
						},
					},
				},
			},
		})
	if err != nil { return err }

	return nil
}

// This will collect merics such as cpu usage and send them to Cloudwatch.
// For full list of metrics see https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Container-Insights-metrics-EKS.html
func DeployCloudwatchAgent(ctx *pulumi.Context, clusterName string, region string, namespace string) error {
	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok { return errors.New("Could not get directory path for kubernetes module.") }

	_, err := corev1.NewServiceAccount(ctx, "cloudwatch-agent-service-account", &corev1.ServiceAccountArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("ServiceAccount"),
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("cloudwatch-agent"),
			Namespace: pulumi.String(namespace),
		},
	})
	if err != nil { return nil }

	_, err = rbacv1.NewClusterRole(ctx, "cloudwatch-agent-cluster-role", &rbacv1.ClusterRoleArgs{
		Kind:       pulumi.String("ClusterRole"),
		ApiVersion: pulumi.String("rbac.authorization.k8s.io/v1"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("cloudwatch-agent-role"),
		},
		Rules: rbacv1.PolicyRuleArray{
			&rbacv1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{
					pulumi.String(""),
				},
				Resources: pulumi.StringArray{
					pulumi.String("pods"),
					pulumi.String("nodes"),
					pulumi.String("endpoints"),
				},
				Verbs: pulumi.StringArray{
					pulumi.String("list"),
					pulumi.String("watch"),
				},
			},
			&rbacv1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{
					pulumi.String("apps"),
				},
				Resources: pulumi.StringArray{
					pulumi.String("replicasets"),
				},
				Verbs: pulumi.StringArray{
					pulumi.String("list"),
					pulumi.String("watch"),
				},
			},
			&rbacv1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{
					pulumi.String("batch"),
				},
				Resources: pulumi.StringArray{
					pulumi.String("jobs"),
				},
				Verbs: pulumi.StringArray{
					pulumi.String("list"),
					pulumi.String("watch"),
				},
			},
			&rbacv1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{
					pulumi.String(""),
				},
				Resources: pulumi.StringArray{
					pulumi.String("nodes/proxy"),
				},
				Verbs: pulumi.StringArray{
					pulumi.String("get"),
				},
			},
			&rbacv1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{
					pulumi.String(""),
				},
				Resources: pulumi.StringArray{
					pulumi.String("nodes/stats"),
					pulumi.String("configmaps"),
					pulumi.String("events"),
				},
				Verbs: pulumi.StringArray{
					pulumi.String("create"),
				},
			},
			&rbacv1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{
					pulumi.String(""),
				},
				Resources: pulumi.StringArray{
					pulumi.String("configmaps"),
				},
				ResourceNames: pulumi.StringArray{
					pulumi.String("cwagent-clusterleader"),
				},
				Verbs: pulumi.StringArray{
					pulumi.String("get"),
					pulumi.String("update"),
				},
			},
		},
	})
	if err != nil { return nil }

	_, err = rbacv1.NewClusterRoleBinding(ctx, "cloudwatch-agent-cluster-role-binding", &rbacv1.ClusterRoleBindingArgs{
		Kind:       pulumi.String("ClusterRoleBinding"),
		ApiVersion: pulumi.String("rbac.authorization.k8s.io/v1"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("cloudwatch-agent-role-binding"),
		},
		Subjects: rbacv1.SubjectArray{
			&rbacv1.SubjectArgs{
				Kind:      pulumi.String("ServiceAccount"),
				Name:      pulumi.String("cloudwatch-agent"),
				Namespace: pulumi.String(namespace),
			},
		},
		RoleRef: &rbacv1.RoleRefArgs{
			Kind:     pulumi.String("ClusterRole"),
			Name:     pulumi.String("cloudwatch-agent-role"),
			ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
		},
	})
	if err != nil { return nil }

	type CloudwatchConfig struct {
		CLUSTER_NAME string
		REGION string
	}

	config := CloudwatchConfig {
		CLUSTER_NAME: clusterName,
		REGION: region,
	}
	cwConfig, err := utils.ParseTemplateAsBytes(config, filepath.Join(filepath.Dir(moduleDir), "/cloudwatchagent/config.json"))
	if err != nil { return nil }
	_, err = corev1.NewConfigMap(ctx, "cloudwatchagentconfigConfigMap", &corev1.ConfigMapArgs{
			ApiVersion: pulumi.String("v1"),
			Data: pulumi.StringMap{
				"cwagentconfig.json": pulumi.String(cwConfig.String()),
			},
			Kind: pulumi.String("ConfigMap"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("cwagentconfig"),
				Namespace: pulumi.String(namespace),
			},
		})
	if err != nil { return nil }

	_, err = appsv1.NewDaemonSet(ctx, "cloudwatch-agent-daemon-set", &appsv1.DaemonSetArgs{
			ApiVersion: pulumi.String("apps/v1"),
			Kind:       pulumi.String("DaemonSet"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("cloudwatch-agent"),
				Namespace: pulumi.String(namespace),
			},
			Spec: &appsv1.DaemonSetSpecArgs{
				Selector: &metav1.LabelSelectorArgs{
					MatchLabels: pulumi.StringMap{
						"name": pulumi.String("cloudwatch-agent"),
					},
				},
				Template: &corev1.PodTemplateSpecArgs{
					Metadata: &metav1.ObjectMetaArgs{
						Labels: pulumi.StringMap{
							"name": pulumi.String("cloudwatch-agent"),
						},
					},
					Spec: &corev1.PodSpecArgs{
						Containers: corev1.ContainerArray{
							&corev1.ContainerArgs{
								Name:  pulumi.String("cloudwatch-agent"),
								Image: pulumi.String("amazon/cloudwatch-agent:1.230621.0"),
								Resources: &corev1.ResourceRequirementsArgs{
									Limits: pulumi.StringMap{
										"cpu":    pulumi.String("100m"),
										"memory": pulumi.String("100Mi"),
									},
									Requests: pulumi.StringMap{
										"cpu":    pulumi.String("100m"),
										"memory": pulumi.String("100Mi"),
									},
								},
								Env: corev1.EnvVarArray{
									&corev1.EnvVarArgs{
										Name: pulumi.String("HOST_IP"),
										ValueFrom: &corev1.EnvVarSourceArgs{
											FieldRef: &corev1.ObjectFieldSelectorArgs{
												FieldPath: pulumi.String("status.hostIP"),
											},
										},
									},
									&corev1.EnvVarArgs{
										Name: pulumi.String("HOST_NAME"),
										ValueFrom: &corev1.EnvVarSourceArgs{
											FieldRef: &corev1.ObjectFieldSelectorArgs{
												FieldPath: pulumi.String("spec.nodeName"),
											},
										},
									},
									&corev1.EnvVarArgs{
										Name: pulumi.String("K8S_NAMESPACE"),
										ValueFrom: &corev1.EnvVarSourceArgs{
											FieldRef: &corev1.ObjectFieldSelectorArgs{
												FieldPath: pulumi.String("metadata.namespace"),
											},
										},
									},
									&corev1.EnvVarArgs{
										Name:  pulumi.String("CI_VERSION"),
										Value: pulumi.String("k8s/1.0.1"),
									},
								},
								VolumeMounts: corev1.VolumeMountArray{
									&corev1.VolumeMountArgs{
										Name:      pulumi.String("cwagentconfig"),
										MountPath: pulumi.String("/etc/cwagentconfig"),
									},
									&corev1.VolumeMountArgs{
										Name:      pulumi.String("rootfs"),
										MountPath: pulumi.String("/rootfs"),
										ReadOnly:  pulumi.Bool(true),
									},
									&corev1.VolumeMountArgs{
										Name:      pulumi.String("dockersock"),
										MountPath: pulumi.String("/var/run/docker.sock"),
										ReadOnly:  pulumi.Bool(true),
									},
									&corev1.VolumeMountArgs{
										Name:      pulumi.String("varlibdocker"),
										MountPath: pulumi.String("/var/lib/docker"),
										ReadOnly:  pulumi.Bool(true),
									},
									&corev1.VolumeMountArgs{
										Name:      pulumi.String("sys"),
										MountPath: pulumi.String("/sys"),
										ReadOnly:  pulumi.Bool(true),
									},
									&corev1.VolumeMountArgs{
										Name:      pulumi.String("devdisk"),
										MountPath: pulumi.String("/dev/disk"),
										ReadOnly:  pulumi.Bool(true),
									},
								},
							},
						},
						Volumes: corev1.VolumeArray{
							&corev1.VolumeArgs{
								Name: pulumi.String("cwagentconfig"),
								ConfigMap: &corev1.ConfigMapVolumeSourceArgs{
									Name: pulumi.String("cwagentconfig"),
								},
							},
							&corev1.VolumeArgs{
								Name: pulumi.String("rootfs"),
								HostPath: &corev1.HostPathVolumeSourceArgs{
									Path: pulumi.String("/"),
								},
							},
							&corev1.VolumeArgs{
								Name: pulumi.String("dockersock"),
								HostPath: &corev1.HostPathVolumeSourceArgs{
									Path: pulumi.String("/var/run/docker.sock"),
								},
							},
							&corev1.VolumeArgs{
								Name: pulumi.String("varlibdocker"),
								HostPath: &corev1.HostPathVolumeSourceArgs{
									Path: pulumi.String("/var/lib/docker"),
								},
							},
							&corev1.VolumeArgs{
								Name: pulumi.String("sys"),
								HostPath: &corev1.HostPathVolumeSourceArgs{
									Path: pulumi.String("/sys"),
								},
							},
							&corev1.VolumeArgs{
								Name: pulumi.String("devdisk"),
								HostPath: &corev1.HostPathVolumeSourceArgs{
									Path: pulumi.String("/dev/disk/"),
								},
							},
						},
						TerminationGracePeriodSeconds: pulumi.Int(60),
						ServiceAccountName:            pulumi.String("cloudwatch-agent"),
					},
				},
			},
		})
	if err != nil { return nil }

	return nil
}
