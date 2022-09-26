package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/services/cockroach"
	"gitlab.com/fynbos/infra/services/kratos"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		crCert, err := cockroach.CreateClientCert(ctx, &cockroach.ClientCertArgs{
			Issuer:    "ca-issuer",
			Namespace: "kratos",
			Name:      "kratos",
		})
		if err != nil {
			return err
		}
		_, err = kratos.DeployKratos(ctx, crCert, "http://fynbos.test", "CHANGE-ME-I-AM-VERY-INSECURE1234")
		if err != nil {
			return err
		}
		err = kratos.DeployKratosIngress(ctx)
		if err != nil {
			return err
		}

		return nil
	})
}
