package main

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apiextensions"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"gitlab.com/fynbos/infra/aws/modules/utils"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		k8sStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-shared-eu-west-1-shared-k8s/main", nil)
		if err != nil {
			return err
		}
		kubeConfig := k8sStack.GetStringOutput(pulumi.String("kubeconfig"))
		kubeProvider, err := kubernetes.NewProvider(ctx, "kubernetes-provider", &kubernetes.ProviderArgs{
			Kubeconfig: kubeConfig,
			Namespace:  pulumi.String("argocd"),
		})
		if err != nil {
			return err
		}
		oidcProvider := k8sStack.GetStringOutput(pulumi.String("oidcProvider"))
		cfg := config.New(ctx, "fynbos")
		accountID := cfg.Require("accountId")

		namespace, err := corev1.NewNamespace(ctx, "namespace", &corev1.NamespaceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name: pulumi.String("argocd"),
				Labels: pulumi.StringMap{
					"pod-security.kubernetes.io/enforce": pulumi.String("baseline"),
					"pod-security.kubernetes.io/audit":   pulumi.String("baseline"),
					"pod-security.kubernetes.io/warn":    pulumi.String("baseline"),
				},
			},
		}, pulumi.Provider(kubeProvider))
		if err != nil {
			return err
		}

		role, err := newArgoCdRole(ctx, pulumi.String(accountID), oidcProvider, namespace.Metadata.Name().Elem())
		if err != nil {
			return err
		}

		argo, err := yaml.NewConfigFile(ctx, "argo", &yaml.ConfigFileArgs{
			File: "./argocd.yaml",
			Transformations: []yaml.Transformation{
				// Set the aws role arn on the correct service accounts
				func(state map[string]interface{}, opts ...pulumi.ResourceOption) {
					name := state["metadata"].(map[string]interface{})["name"]
					if state["kind"] == "ServiceAccount" && (name == "argocd-server" || name == "argocd-application-controller") {
						metadata := state["metadata"].(map[string]interface{})
						annotations := make(map[string]interface{})
						annotations["eks.amazonaws.com/role-arn"] = role.Arn
						metadata["annotations"] = annotations
					}
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{role, namespace}))
		if err != nil {
			return err
		}

		_, err = apiextensions.NewCustomResource(ctx, fmt.Sprintf("argocd-mapping-http"), &apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("getambassador.io/v3alpha1"),
			Kind:       pulumi.String("Mapping"),
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String("argocd-server-ui"),
				Namespace: namespace.Metadata.Name(),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": pulumi.Map{
					"hostname": pulumi.String("argocd.mgnt.fynbos.dev"),
					"prefix":   pulumi.String("/"),
					"rewrite":  pulumi.String("/"),
					"service":  pulumi.String("argocd-server:443"),
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{namespace, argo}))
		if err != nil {
			return err
		}

		_, err = apiextensions.NewCustomResource(ctx, "root-application", &apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("argoproj.io/v1alpha1"),
			Kind:       pulumi.String("Application"),
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String("root-app"),
				Namespace: namespace.Metadata.Name(),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": pulumi.Map{
					"source": pulumi.Map{
						"repoURL":        pulumi.String("https://gitlab.com/fynbos/rooibos.git"),
						"targetRevision": pulumi.String("main"),
						"path":           pulumi.String("root-app"),
					},
					"syncPolicy": pulumi.Map{
						"automated": pulumi.Map{},
					},
					"destination": pulumi.Map{
						"server":    pulumi.String("https://kubernetes.default.svc"),
						"namespace": namespace.Metadata.Name(),
					},
					"project": pulumi.String("default"),
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{argo, namespace}))
		if err != nil {
			return err
		}

		ctx.Export("argoRoleArn", role.Arn)
		return nil
	})
}

func newArgoCdRole(ctx *pulumi.Context, accountId pulumi.StringPtrInput, oidcProvider pulumi.StringPtrInput, namespace pulumi.StringPtrInput) (*iam.Role, error) {

	policy, err := newRolePolicy(ctx)
	if err != nil {
		return nil, err
	}
	trustPolicy := newTrustPolicy(ctx, accountId, oidcProvider, namespace)

	role, err := iam.NewRole(ctx, "argocd-role", &iam.RoleArgs{
		Name:             pulumi.String("eksArgoRole"),
		Description:      pulumi.String(""),
		AssumeRolePolicy: trustPolicy,
		InlinePolicies: iam.RoleInlinePolicyArray{
			iam.RoleInlinePolicyArgs{
				Name:   pulumi.String("assumePolicy"),
				Policy: pulumi.String(policy),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return role, err
}

func newRolePolicy(ctx *pulumi.Context) (string, error) {
	policy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Effect: utils.StringPtr("Allow"),
				Actions: []string{
					"sts:AssumeRole",
				},
				Resources: []string{
					"*",
				},
			},
		},
	})
	if err != nil {
		return "", err
	}

	return policy.Json, nil
}

// Need a custom one as it has different requirements for the conditions
func newTrustPolicy(ctx *pulumi.Context, accountId pulumi.StringPtrInput, oidcProvider pulumi.StringPtrInput, namespace pulumi.StringPtrInput) pulumi.StringOutput {
	return pulumi.All(accountId, oidcProvider, namespace).ApplyT(func(args []interface{}) (string, error) {
		aid := args[0].(string)
		oidc := args[1].(string)
		ns := args[2].(string)

		conditions := []iam.GetPolicyDocumentStatementCondition{
			{
				Test:     "StringEquals",
				Values:   []string{"sts.amazonaws.com"},
				Variable: fmt.Sprintf("%s:aud", oidc),
			},
		}

		conditions = append(conditions, iam.GetPolicyDocumentStatementCondition{
			Test: "StringLike",
			Values: []string{
				fmt.Sprintf("system:serviceaccount:%s:*", ns),
			},
			Variable: fmt.Sprintf("%s:sub", oidc),
		})

		policy, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
			Statements: []iam.GetPolicyDocumentStatement{
				{
					Effect: utils.StringPtr("Allow"),
					Actions: []string{
						"sts:AssumeRoleWithWebIdentity",
					},
					Principals: []iam.GetPolicyDocumentStatementPrincipal{
						{
							Type: "Federated",
							Identifiers: []string{
								fmt.Sprintf("arn:aws:iam::%s:oidc-provider/%s", aid, oidc),
							},
						},
					},
					Conditions: conditions,
				},
			},
		})
		if err != nil {
			return "", err
		}

		return policy.Json, nil
	}).(pulumi.StringOutput)
}

type clusterArgs struct {
	Name             string
	ClusterName      string
	Namespace        pulumi.StringInput
	ClusterStackName string
}

func newCluster(ctx *pulumi.Context, args clusterArgs, opts ...pulumi.ResourceOption) error {
	stack, err := pulumi.NewStackReference(ctx, args.ClusterStackName, nil)
	if err != nil {
		return err
	}
	ca := stack.GetStringOutput(pulumi.String("ca"))
	deployRoleArn := stack.GetStringOutput(pulumi.String("deployRoleArn"))
	endpoint := stack.GetStringOutput(pulumi.String("clusterEndpoint"))

	clusterCf := newClusterConfig(clusterConfigArgs{
		ClusterName: pulumi.String(args.ClusterName),
		RoleARN:     deployRoleArn,
		CaData:      ca,
	})

	_, err = corev1.NewSecret(ctx, args.Name, &corev1.SecretArgs{
		ApiVersion: pulumi.String("v1"),
		Type:       pulumi.String("Opaque"),
		Metadata: metav1.ObjectMetaArgs{
			Name:      pulumi.Sprintf("%s-cluster", args.Name),
			Namespace: args.Namespace,
			Labels: pulumi.StringMap{
				"argocd.argoproj.io/secret-type": pulumi.String("cluster"),
			},
		},
		StringData: pulumi.StringMap{
			"name":   pulumi.String(args.Name),
			"server": endpoint,
			"config": clusterCf,
		},
	}, opts...)

	return nil
}

type awsAuthConfig struct {
	ClusterName string `json:"clusterName"`
	RoleARN     string `json:"roleARN"`
}
type tlsClientConfig struct {
	CaData string `json:"caData"`
}

type clusterConfig struct {
	AwsAuthConfig   awsAuthConfig   `json:"awsAuthConfig"`
	TlsClientConfig tlsClientConfig `json:"tlsClientConfig"`
}

type clusterConfigArgs struct {
	ClusterName pulumi.StringInput
	RoleARN     pulumi.StringInput
	CaData      pulumi.StringInput
}

func newClusterConfig(args clusterConfigArgs) pulumi.StringOutput {
	return pulumi.All(args.ClusterName, args.RoleARN, args.CaData).ApplyT(func(args []interface{}) (string, error) {

		c := clusterConfig{
			AwsAuthConfig: awsAuthConfig{
				ClusterName: args[0].(string),
				RoleARN:     args[1].(string),
			},
			TlsClientConfig: tlsClientConfig{
				CaData: args[2].(string),
			},
		}

		configJson, err := json.Marshal(c)

		return string(configJson), err
	}).(pulumi.StringOutput)
}
