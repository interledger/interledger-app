package redis

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DeployRedisArgs struct {
	EnableSentinal bool
	ReplicaCount   uint8
}

func DeployRedis(ctx *pulumi.Context, args *DeployRedisArgs, opts ...pulumi.ResourceOption) error {
	architecture := "standalone"
	if args.ReplicaCount > 0 {
		architecture = "replication"
	}

	nc, err := createNodeCert(ctx, &NodeCertArgs{
		Issuer:      "ca-issuer",
		Namespace:   "default",
		ServiceName: "redis-master",
	}, opts...)
	if err != nil {
		return err
	}

	_, err = helm.NewChart(ctx, "redis", helm.ChartArgs{
		Version: pulumi.String("16.8.5"),
		Chart:   pulumi.String("redis"),
		FetchArgs: &helm.FetchArgs{
			Repo: pulumi.String("https://charts.bitnami.com/bitnami"),
		},
		Values: pulumi.Map{
			"architecture": pulumi.String(architecture),
			"auth": pulumi.Map{
				"enabled": pulumi.Bool(false), // TODO: enable tls and mount secrets
			},
			"replica": pulumi.Map{
				"replicaCount": pulumi.Int(args.ReplicaCount),
			},
			"serviceAccount": pulumi.Map{
				"create":                       pulumi.Bool(true),
				"name":                         pulumi.String("redis"),
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
			"sentinel": pulumi.Map{
				"enabled": pulumi.Bool(args.EnableSentinal),
			},
			"tls": pulumi.Map{
				"enabled":            pulumi.Bool(true),
				"certificatesSecret": pulumi.String("redis-node"),
				"certFilename":       pulumi.String("tls.crt"),
				"certKeyFilename":    pulumi.String("tls.key"),
				"certCAFilename":     pulumi.String("ca.crt"),
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{nc}))
	if err != nil {
		return err
	}

	return nil
}
