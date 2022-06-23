package main

import (
	"encoding/base64"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	v1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-vault/sdk/v5/go/vault"
	vaultk8s "github.com/pulumi/pulumi-vault/sdk/v5/go/vault/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/aws/modules/utils"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		k8sStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-shared-eu-west-1-shared-k8s/main", nil)
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
		secret, err := v1.GetSecret(ctx, "secret", pulumi.ID("vault/vault-token-96gx2"), nil, pulumi.Provider(kubeProvider))
		if err != nil {
			return err
		}
		token := secret.Data.MapIndex(pulumi.String("token"))

		k8sAuth, err := vault.NewAuthBackend(ctx, "k8s-auth", &vault.AuthBackendArgs{
			Type: pulumi.String("kubernetes"),
			Path: pulumi.String("k8s-shared-euw1"),
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
			Name:   pulumi.String("shared-euw1-kubernetes-reader"),
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
