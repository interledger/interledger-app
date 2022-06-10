package main

import (
	"github.com/pulumi/pulumi-vault/sdk/v4/go/vault"
	"github.com/pulumi/pulumi-vault/sdk/v4/go/vault/okta"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/aws/modules/utils"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		type Data struct{}
		data := Data{}
		policyTemplate, err := utils.ParseTemplateAsBytes(data, "./admin-policy.hcl")
		if err != nil {
			return err
		}
		adminPolicy, err := vault.NewPolicy(ctx, "admin-policy", &vault.PolicyArgs{
			Name:   pulumi.String("admin"),
			Policy: pulumi.String(policyTemplate.String()),
		})
		if err != nil {
			return err
		}

		oktaBackend, err := vault.LookupAuthBackend(ctx, &vault.LookupAuthBackendArgs{
			Path: "okta",
		}, nil)
		if err != nil {
			return err
		}

		_, err = okta.NewAuthBackendGroup(ctx, "okta-admin", &okta.AuthBackendGroupArgs{
			Path:      pulumi.String(oktaBackend.Path),
			GroupName: pulumi.String("vault-admin"),
			Policies: pulumi.StringArray{
				adminPolicy.Name,
			},
		})
		if err != nil {
			return err
		}

		return nil
	})
}
