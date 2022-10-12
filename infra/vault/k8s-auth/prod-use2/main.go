package main

import (
	"encoding/base64"

	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	v1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-vault/sdk/v5/go/vault"
	vaultk8s "github.com/pulumi/pulumi-vault/sdk/v5/go/vault/kubernetes"
	"github.com/pulumi/pulumi-vault/sdk/v5/go/vault/pkisecret"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/aws/modules/utils"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		k8sStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-prod-us-east-2-prod-k8s/main", nil)
		if err != nil {
			return err
		}
		endpoint := k8sStack.GetStringOutput(pulumi.String("clusterEndpoint"))
		ca := k8sStack.GetStringOutput(pulumi.String("ca"))
		oidcProvider := k8sStack.GetStringOutput(pulumi.String("oidcProvider"))

		kubeConfig := k8sStack.GetStringOutput(pulumi.String("kubeconfig"))
		kubeProvider, err := kubernetes.NewProvider(ctx, "kubernetes-provider", &kubernetes.ProviderArgs{
			Kubeconfig: kubeConfig,
		})
		if err != nil {
			return err
		}

		// the pulumiID value is {NAMESPACE/SECRET_NAME}. Manually lookup as its easier.
		secret, err := v1.GetSecret(ctx, "secret", pulumi.ID("vault/vault-token-ntx7s"), nil, pulumi.Provider(kubeProvider))
		if err != nil {
			return err
		}
		token := secret.Data.MapIndex(pulumi.String("token"))

		k8sAuth, err := vault.NewAuthBackend(ctx, "k8s-auth", &vault.AuthBackendArgs{
			Type: pulumi.String("kubernetes"),
			Path: pulumi.String("k8s-prod-use2"),
		})
		if err != nil {
			return err
		}

		_, err = vaultk8s.NewAuthBackendConfig(ctx, "k8s-auth-config", &vaultk8s.AuthBackendConfigArgs{
			Backend:              k8sAuth.Path,
			KubernetesHost:       endpoint,
			KubernetesCaCert:     base64Decode(ca),
			TokenReviewerJwt:     base64Decode(token),
			Issuer:               oidcProvider,
			DisableIssValidation: pulumi.Bool(true),
		})
		if err != nil {
			return err
		}

		policy, err := vault.NewPolicy(ctx, "kubernetes-reader", &vault.PolicyArgs{
			Name:   pulumi.String("prod-use2-kubernetes-reader"),
			Policy: newPolicy(k8sAuth.Accessor),
		})
		if err != nil {
			return err
		}

		_, err = vaultk8s.NewAuthBackendRole(ctx, "kubernetes-role", &vaultk8s.AuthBackendRoleArgs{
			RoleName: pulumi.String("k8s-app"),
			Backend:  k8sAuth.Path,
			BoundServiceAccountNames: pulumi.StringArray{
				pulumi.String("*"),
			},
			BoundServiceAccountNamespaces: pulumi.StringArray{
				pulumi.String("*"),
			},
			TokenPolicies: pulumi.StringArray{
				policy.Name,
			},
		}, pulumi.DependsOn([]pulumi.Resource{policy}))
		if err != nil {
			return err
		}

		err = createCrdbNodeRole(ctx, k8sAuth.Path)
		if err != nil {
			return err
		}

		err = createEmissaryRole(ctx, k8sAuth.Path)
		if err != nil {
			return err
		}

		err = createCrdbClientRole(ctx, k8sAuth.Accessor, pulumi.String("pki/prod-int"))
		if err != nil {
			return err
		}

		return nil
	})
}

func base64Decode(input pulumi.StringPtrInput) pulumi.StringOutput {
	return input.ToStringPtrOutput().ApplyT(func(arg interface{}) (string, error) {
		str := arg.(*string)
		data, err := base64.StdEncoding.DecodeString(*str)
		if err != nil {
			return "", err
		}

		return string(data), nil
	}).(pulumi.StringOutput)
}

func newPolicy(authAccessor pulumi.StringOutput) pulumi.StringOutput {
	return pulumi.All(authAccessor).ApplyT(func(args []interface{}) (string, error) {
		data := struct {
			AuthAccessor string
		}{
			AuthAccessor: args[0].(string),
		}
		parsedPolicy, err := utils.ParseTemplateAsBytes(data, "./policy.hcl")
		if err != nil {
			return "", err
		}
		return parsedPolicy.String(), nil
	}).(pulumi.StringOutput)
}

func createCrdbNodeRole(ctx *pulumi.Context, path pulumi.StringPtrInput) error {
	policy, err := vault.NewPolicy(ctx, "crdb-node", &vault.PolicyArgs{
		Name: pulumi.String("prod-use2-crdb-node"),
		Policy: pulumi.String(`
path "pki/prod-int/sign/crdb-node"
{
  capabilities = ["read", "create", "update"]
}
`),
	})
	if err != nil {
		return err
	}

	_, err = vaultk8s.NewAuthBackendRole(ctx, "crdb-node-role", &vaultk8s.AuthBackendRoleArgs{
		RoleName: pulumi.String("crdb-node"),
		Backend:  path,
		BoundServiceAccountNames: pulumi.StringArray{
			pulumi.String("*"),
		},
		BoundServiceAccountNamespaces: pulumi.StringArray{
			pulumi.String("cockroachdb"),
		},
		TokenPolicies: pulumi.StringArray{
			policy.Name,
		},
	}, pulumi.DependsOn([]pulumi.Resource{policy}))
	if err != nil {
		return err
	}

	return nil
}

func createEmissaryRole(ctx *pulumi.Context, path pulumi.StringPtrInput) error {
	policy, err := vault.NewPolicy(ctx, "emissary", &vault.PolicyArgs{
		Name: pulumi.String("prod-use2-emissary"),
		Policy: pulumi.String(`
path "pki/prod-int/sign/emissary"
{
  capabilities = ["read", "create", "update"]
}
`),
	})
	if err != nil {
		return err
	}

	_, err = vaultk8s.NewAuthBackendRole(ctx, "emissary-role", &vaultk8s.AuthBackendRoleArgs{
		RoleName: pulumi.String("emissary"),
		Backend:  path,
		BoundServiceAccountNames: pulumi.StringArray{
			pulumi.String("emissary"),
		},
		BoundServiceAccountNamespaces: pulumi.StringArray{
			pulumi.String("emissary"),
		},
		TokenPolicies: pulumi.StringArray{
			policy.Name,
		},
	}, pulumi.DependsOn([]pulumi.Resource{policy}))
	if err != nil {
		return err
	}

	return nil
}

// This is not where I would want to put this function, but its difficult to deal with as if it is in PKI, then PKI
// needs to know the k8s-auth mount accessor. So not sure how best to handle that. As it becomes circular dependency.
func createCrdbClientRole(ctx *pulumi.Context, authAccessor pulumi.StringOutput, pkiPath pulumi.StringInput) error {
	_, err := pkisecret.NewSecretBackendRole(ctx, "crdb-client", &pkisecret.SecretBackendRoleArgs{
		Name:                   pulumi.String("crdb-client"),
		Backend:                pkiPath,
		AllowBareDomains:       pulumi.Bool(true),
		AllowLocalhost:         pulumi.Bool(false),
		AllowSubdomains:        pulumi.Bool(false),
		AllowedDomainsTemplate: pulumi.Bool(true),
		AllowedDomains: pulumi.StringArray{
			pulumi.Sprintf("{{identity.entity.aliases.%s.metadata.service_account_namespace}}", authAccessor),
		},
		MaxTtl: pulumi.String("2765000"), // 32days in seconds
	})
	return err
}
