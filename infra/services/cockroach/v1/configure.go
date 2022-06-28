package crdb

import (
	"errors"
	batchv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/batch/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"io/ioutil"
	"path/filepath"
	"runtime"
)

type InitCrdbJobArgs struct {
	Namespace      pulumi.StringPtrInput
	RootCertSecret pulumi.StringPtrInput
}

func InitCrdbJob(ctx *pulumi.Context, args InitCrdbJobArgs, opts ...pulumi.ResourceOption) error {

	// cockroach init file read from `dbinit.sql` file.
	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok {
		return errors.New("could not get directory path for cockroach module")
	}

	dat, err := ioutil.ReadFile(filepath.Join(filepath.Dir(moduleDir), "dbinit.sql"))
	if err != nil {
		return err
	}

	sqlInitConfig, err := corev1.NewConfigMap(ctx, "crdb-init-sql", &corev1.ConfigMapArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("cockroachdb-init"),
			Namespace: args.Namespace,
		},
		Data: pulumi.StringMap{"dbinit.sql": pulumi.String(dat)},
	}, opts...)

	_, err = batchv1.NewJob(ctx, "cockroachdb-init", &batchv1.JobArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("Job"),
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("cockroachdb-init"),
			Namespace: args.Namespace,
		},
		Spec: &batchv1.JobSpecArgs{
			Template: &corev1.PodTemplateSpecArgs{
				Spec: &corev1.PodSpecArgs{
					Containers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:            pulumi.String("cockroachdb-client"),
							Image:           pulumi.String("cockroachdb/cockroach:v21.1.11"),
							ImagePullPolicy: pulumi.String("IfNotPresent"),
							VolumeMounts: corev1.VolumeMountArray{
								&corev1.VolumeMountArgs{
									Name:      pulumi.String("sql-init-config"),
									MountPath: pulumi.String("/cockroach/init/"),
								},
								&corev1.VolumeMountArgs{
									Name:      pulumi.String("certs"),
									MountPath: pulumi.String("/cockroach/cockroach-certs/"),
								},
							},
							Command: pulumi.StringArray{
								pulumi.String("./cockroach"),
								pulumi.String("sql"),
								pulumi.String("--certs-dir=/cockroach/cockroach-certs"),
								pulumi.String("--file=/cockroach/init/dbinit.sql"),
								pulumi.String("--user=root"),
								pulumi.String("--host=cockroachdb-public"),
							},
						},
					},
					Volumes: corev1.VolumeArray{
						&corev1.VolumeArgs{
							Name: pulumi.String("sql-init-config"),
							ConfigMap: &corev1.ConfigMapVolumeSourceArgs{
								Name: sqlInitConfig.Metadata.Name(),
							},
						},
						&corev1.VolumeArgs{
							Name: pulumi.String("certs"),
							Projected: &corev1.ProjectedVolumeSourceArgs{
								Sources: &corev1.VolumeProjectionArray{
									&corev1.VolumeProjectionArgs{
										Secret: corev1.SecretProjectionArgs{
											Name: args.RootCertSecret,
											Items: corev1.KeyToPathArray{
												corev1.KeyToPathArgs{
													Key:  pulumi.String("tls.crt"),
													Path: pulumi.String("client.root.crt"),
													Mode: pulumi.Int(256),
												},
												corev1.KeyToPathArgs{
													Key:  pulumi.String("tls.key"),
													Path: pulumi.String("client.root.key"),
													Mode: pulumi.Int(256),
												},
												corev1.KeyToPathArgs{
													Key:  pulumi.String("ca.crt"),
													Path: pulumi.String("ca.crt"),
													Mode: pulumi.Int(256),
												},
											},
										},
									},
								},
							},
						},
					},
					RestartPolicy: pulumi.String("OnFailure"),
				},
			},
		},
	}, opts...)
	if err != nil {
		return err
	}
	return nil
}
