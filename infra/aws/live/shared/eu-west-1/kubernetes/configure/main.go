package main

import (

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	k8s "gitlab.com/fynbos/infra/aws/modules/kubernetes"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		clusterName := "live-eu-west-1-eks-cluster"
		conf := config.New(ctx, "cluster")
		// TODO: figure out how to extract oidcID from OidcProviderOutput exported in kubernetes stack.
		oidcId := conf.Get("oidcId")

		// TODO: Figure out why we can't import kubeconfig from provision stack and create an eks provider without it erroring out.
		err := k8s.UpdateAwsNodeDaemonSetToUseIrsa(ctx, oidcId)
		if err != nil {
			return err
		}

		err = k8s.ConfigureClusterRolesAndPsp(ctx)
		if err != nil { return err }

		err = k8s.DeployLoggingAndMonitoring(ctx, clusterName, "eu-west-1")
		if err != nil { return err }

		calicoOperator, calicoCrds, err := k8s.DeployCalico(ctx)
		if err != nil { return err }

		err = k8s.ConfigureDefaultNetworkPolicy(ctx, []pulumi.Resource{calicoOperator, calicoCrds})
		if err != nil { return err }		

		return nil
	})
}
