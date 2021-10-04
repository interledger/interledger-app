package main

import (
	"github.com/pulumi/pulumi-vault/sdk/v4/go/vault"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	utils "gitlab.com/fynbos/infra/aws/modules/utils"
)

func newBoundaryControllerPolicy (ctx *pulumi.Context, provider *vault.Provider) (*vault.Policy, error) {
	type Data struct {}
	data := Data{}
	policy, err := vault.NewPolicy(ctx, "boundary-controller-policy", &vault.PolicyArgs{
		Name: pulumi.String("boundary-controller"),
		Policy: pulumi.String(utils.ParseTemplateAsBytes(data, "./policy/boundary-controller-policy.hcl")),
	}, pulumi.Provider(provider))
	if err != nil { return nil, err }


	return policy, nil
}

func newAdminUserPolicy (ctx *pulumi.Context, provider *vault.Provider) (*vault.Policy, error) {
	type Data struct {}
	data := Data{}
	policy, err := vault.NewPolicy(ctx, "admin-user-policy", &vault.PolicyArgs{
		Name: pulumi.String("admin"),
		Policy: pulumi.String(utils.ParseTemplateAsBytes(data, "./policy/admin.hcl")),
	}, pulumi.Provider(provider))
	if err != nil { return nil, err }


	return policy, nil
}