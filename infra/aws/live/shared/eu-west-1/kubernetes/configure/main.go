package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	k8s "gitlab.com/fynbos/infra/aws/modules/kubernetes"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		conf := config.New(ctx, "cluster")
		// TODO: figure out how to extract oidcID from OidcProviderOutput exported in kubernetes stack.
		oidcProvider := conf.Get("oidcProvider")
		clusterName := conf.Get("name")

		fynbosConf := config.New(ctx, "fynbos")
		accountId := fynbosConf.Get("accountId")

		// TODO: Figure out why we can't import kubeconfig from provision stack and create an eks provider without it erroring out.
		err := k8s.UpdateAwsNodeDaemonSetToUseIrsa(ctx, accountId, oidcProvider)
		if err != nil {
			return err
		}

		err = k8s.ConfigureClusterRolesAndPsp(ctx)
		if err != nil {
			return err
		}

		err = k8s.DeployLoggingAndMonitoring(ctx, clusterName, "eu-west-1")
		if err != nil {
			return err
		}

		baselineStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-shared-baseline/main", nil)
		if err != nil {
			return err
		}
		ebsKmsKeyArn := baselineStack.GetStringOutput(pulumi.String("ebsEncryptionKeyArn"))

		err = k8s.DeployEbsCsi(ctx, k8s.EbsCsiArgs{
			ClusterName:  clusterName,
			EbsKmsKeyArn: ebsKmsKeyArn,
			OidcProvider: oidcProvider,
			AccountId:    accountId,
		})
		if err != nil {
			return err
		}

		err = k8s.DeployDefaultCSIStorageClass(ctx)
		if err != nil {
			return err
		}

		//calicoOperator, calicoCrds, err := k8s.DeployCalico(ctx)
		//if err != nil { return err }
		//
		//err = k8s.ConfigureDefaultNetworkPolicy(ctx, []pulumi.Resource{calicoOperator, calicoCrds})
		//if err != nil { return err }

		// Configure Automation roles
		err = k8s.ConfigureAutomationRole(ctx, "default")
		if err != nil {
			return err
		}
		err = k8s.ApplyAutomationRoleBindingToNamespace(ctx, "default")
		if err != nil {
			return err
		}

		return nil
	})
}
