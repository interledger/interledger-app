package cert_manager

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func DeployCertManager(ctx *pulumi.Context) error {
	// TODO this didn't work :/
	_, err := yaml.NewConfigFile(ctx, "cert-manager", &yaml.ConfigFileArgs{
		File: "https://github.com/jetstack/cert-manager/releases/download/v1.6.1/cert-manager.yaml",
	})
	if err != nil {
		return err
	}

	return nil
}
