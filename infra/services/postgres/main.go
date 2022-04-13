package postgres

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DeployPostgresArgs struct {
	ReadReplicasCount uint8

	// For now we are using this for Rafiki. Might be neccessary later to use an init.sql file.
	Username string                 // Custom user to create
	Password *random.RandomPassword // Password for custom user
	Database string
}

func DeployPostgres(ctx *pulumi.Context, args *DeployPostgresArgs) error {
	_, err := helm.NewChart(ctx, "postgres", helm.ChartArgs{
		Version: pulumi.String("11.1.19"),
		Chart:   pulumi.String("postgresql"),
		FetchArgs: &helm.FetchArgs{
			Repo: pulumi.String("https://charts.bitnami.com/bitnami"),
		},
		Values: pulumi.Map{
			"auth": pulumi.Map{
				"enablePostgresUser": pulumi.Bool(false),
				"username":           pulumi.String(args.Username),
				"password":           args.Password.Result,
				"database":           pulumi.String(args.Database),
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
		},
	})
	if err != nil {
		return err
	}

	return nil
}
