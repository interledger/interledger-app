package postgres

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"runtime"

	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DeployPostgresArgs struct {
	ReadReplicasCount    uint8
	PostgresUserPassword *random.RandomPassword // Password for "postgres" super user
}

func DeployPostgres(ctx *pulumi.Context, args *DeployPostgresArgs, opts ...pulumi.ResourceOption) error {
	nc, err := createNodeCert(ctx, &NodeCertArgs{
		Issuer:      "ca-issuer",
		Namespace:   "default",
		ServiceName: "postgres-postgresql",
	}, opts...)
	if err != nil {
		return err
	}

	// read from `dbinit.sql` file.
	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok {
		return errors.New("Could not get directory path for postgres module.")
	}

	dat, err := ioutil.ReadFile(filepath.Join(filepath.Dir(moduleDir), "dbinit.sql"))
	if err != nil {
		return err
	}

	sqlInitConfig, err := corev1.NewConfigMap(ctx, "postgres-init-sql", &corev1.ConfigMapArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("postgres-init"),
		},
		Data: pulumi.StringMap{"dbinit.sql": pulumi.String(dat)},
	})
	if err != nil {
		return err
	}

	_, err = helm.NewChart(ctx, "postgres", helm.ChartArgs{
		Version: pulumi.String("11.1.19"),
		Chart:   pulumi.String("postgresql"),
		FetchArgs: &helm.FetchArgs{
			Repo: pulumi.String("https://charts.bitnami.com/bitnami"),
		},
		Values: pulumi.Map{
			"auth": pulumi.Map{
				// enabling super user to run the init scripts.
				"enablePostgresUser": pulumi.Bool(true),
				"postgresPassword":   args.PostgresUserPassword.Result,
			},
			"readReplicas": pulumi.Map{
				"replicaCount": pulumi.Int(args.ReadReplicasCount),
			},
			"serviceAccount": pulumi.Map{
				"create":                       pulumi.Bool(true),
				"name":                         pulumi.String("postgres"),
				"automountServiceAccountToken": pulumi.Bool(false),
			},
			"rbac": pulumi.Map{
				"create": pulumi.Bool(true),
				"rules": pulumi.MapArray{
					pulumi.Map{
						"apiGroups": pulumi.StringArray{pulumi.String("")},
						"resources": pulumi.StringArray{pulumi.String("secrets")},
						"verbs":     pulumi.StringArray{pulumi.String("get")},
					},
				},
			},
			"tls": pulumi.Map{
				"enabled":            pulumi.Bool(true),
				"certificatesSecret": pulumi.String("postgresdb-node"),
				"certFilename":       pulumi.String("tls.crt"),
				"certKeyFilename":    pulumi.String("tls.key"),
				"certCAFilename":     pulumi.String("ca.crt"),
			},
			"volumePermissions": pulumi.Map{
				"enabled": pulumi.Bool(true),
			},
			"primary": pulumi.Map{
				"initdb": pulumi.Map{
					"scriptsConfigMap": sqlInitConfig.Metadata.Name(),
				},
				"persistence": pulumi.Map{
					"size": pulumi.String("1Gi"), // just for dev.
				},
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{nc}))
	if err != nil {
		return err
	}

	return nil
}
