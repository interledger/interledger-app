package main

import (
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	k8s "gitlab.com/fynbos/infra/aws/modules/kubernetes"
	"gitlab.com/fynbos/infra/aws/modules/utils"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		clusterName := "live-eu-west-1-eks-cluster"
		// Setting this to `true` will publically expose the clusters admin endpoint.
		// We do this so that we can configure the cluster afterwhich we will disable
		// the public access.
		allowPublicConfiguration := os.Getenv("ALLOW_PUBLIC_CONFIGURATION") == "true"
		fynbosConf := config.New(ctx, "fynbos")
		awsAccountId := fynbosConf.Get("accountId")
		vpcStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-shared-euwest1-networking/main", nil)
		if err != nil {
			return err
		}
		glRunnerStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-shared-euwest1-glrunner/main", nil)
		if err != nil {
			return err
		}

		vpcId := vpcStack.GetIDOutput(pulumi.String("vpcId"))
		publicSubnetIds := utils.StringArrayOutputFromStack(vpcStack, "publicSubnets")
		privateSubnetIds := utils.StringArrayOutputFromStack(vpcStack, "privateSubnets")
		glRunnerRoleArn := glRunnerStack.GetStringOutput(pulumi.String("glRunnerRoleArn"))
		glRunnerSg := glRunnerStack.GetIDOutput(pulumi.String("glRunnerSecurityGroupID"))

		roles, err := k8s.NewEksRoles(ctx, awsAccountId, glRunnerRoleArn)
		if err != nil {
			return err
		}
		ctx.Export("adminRoleArn", roles.Admin.Arn)
		ctx.Export("automationRoleArn", roles.Automation.Arn)

		cluster, err := k8s.NewEksControlPlane(ctx, k8s.EksControlPlaneArgs{
			Name:                         clusterName,
			Version:                      "1.21",
			VpcId:                        vpcId,
			PublicSubnets:                publicSubnetIds,
			PrivateSubnets:               privateSubnetIds,
			AccountId:                    awsAccountId,
			IamRoles:                     roles,
			ExposeAdminEndpoint:          allowPublicConfiguration,
			ClusterAllowedSecurityGroups: pulumi.StringArray{glRunnerSg},
		})
		if err != nil {
			return err
		}
		ctx.Export("kubeconfig", cluster.Kubeconfig)
		ctx.Export("oidcProvider", cluster.Core.OidcProvider())
		ctx.Export("clusterEndpoint", cluster.Core.Endpoint())

		_, err = k8s.NewManagedNodeGroup(ctx, k8s.ManagedNodeGroupArgs{
			Name:          "managed-ng-0",
			Cluster:       cluster,
			NodeRole:      roles.NodeGroup,
			InstanceTypes: pulumi.StringArray{pulumi.String("t2.medium")},
			MinSize:       1,
			MaxSize:       5,
			DesiredSize:   2,
			SubnetIds:     privateSubnetIds,
		})

		return nil
	})
}
