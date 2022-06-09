package main

import (
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/kms"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/route53"
	"gitlab.com/fynbos/infra/aws/modules/utils"

	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	fynbosK8s "gitlab.com/fynbos/infra/aws/modules/kubernetes"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		vpcStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-shared-euwest1-networking/main", nil)
		if err != nil {
			return err
		}
		dnsZoneId := vpcStack.GetStringOutput(pulumi.String("dnsZoneId"))
		k8sStack, err := pulumi.NewStackReference(ctx, "fynbos/aws-shared-eu-west-1-shared-k8s/main", nil)
		if err != nil {
			return err
		}
		oidcProvider := k8sStack.GetStringOutput(pulumi.String("oidcProvider"))
		kubeConfig := k8sStack.GetStringOutput(pulumi.String("kubeconfig"))
		kubeProvider, err := kubernetes.NewProvider(ctx, "kubernetes-provider", &kubernetes.ProviderArgs{
			Kubeconfig: kubeConfig,
		})
		fynbosConfig := config.New(ctx, "fynbos")
		accountId := fynbosConfig.Get("accountId")

		namespace, err := corev1.NewNamespace(ctx, "vault", &corev1.NamespaceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name: pulumi.String("vault"),
			},
		}, pulumi.Provider(kubeProvider))
		if err != nil {
			return err
		}

		trustPolicy := fynbosK8s.NewIamTrustPolicyDocumentV2(ctx, pulumi.String(accountId), oidcProvider, namespace.Metadata.Name().Elem(), pulumi.String("vault"))
		role, err := iam.NewRole(ctx, "eks-vault-sa-role", &iam.RoleArgs{
			Name:             pulumi.String("eksVaultSaRole"),
			Description:      pulumi.String("Service account role for vault statefulset"),
			AssumeRolePolicy: trustPolicy,
		})
		if err != nil {
			return err
		}

		serviceAccount, err := corev1.NewServiceAccount(ctx, "vault-sa", &corev1.ServiceAccountArgs{
			ApiVersion: pulumi.String("v1"),
			Kind:       pulumi.String("ServiceAccount"),
			Metadata: metav1.ObjectMetaArgs{
				Annotations: pulumi.StringMap{
					"eks.amazonaws.com/role-arn": role.Arn,
				},
				Name:      pulumi.String("vault"),
				Namespace: namespace.Metadata.Name(),
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{
			role,
			namespace,
			kubeProvider,
		}))

		keyPolicy := createKmsPolicy(ctx, role.Arn, pulumi.String(accountId))
		key, err := kms.NewKey(ctx, "vault-encryption-key", &kms.KeyArgs{
			Description: pulumi.String("Key for vault to auto-unseal in shared k8s"),
			Policy:      keyPolicy,
		})
		if err != nil {
			return err
		}

		vaultServerConfig := createVaultConfig(key.KeyId)

		chart, err := helm.NewChart(ctx, "vault", helm.ChartArgs{
			Namespace: namespace.Metadata.Name().Elem(),
			Chart:     pulumi.String("vault"),
			Version:   pulumi.String("0.20.0"),
			FetchArgs: helm.FetchArgs{
				Repo: pulumi.String("https://helm.releases.hashicorp.com"),
			},
			Values: pulumi.Map{
				"server": pulumi.Map{
					"nodeSelector": pulumi.StringMap{
						"vault_in_k8s": pulumi.String("true"),
					},
					"tolerations": pulumi.MapArray{
						pulumi.Map{
							"key":      pulumi.String("taint_for_consul_xor_vault"),
							"operator": pulumi.String("Equal"),
							"value":    pulumi.String("true"),
							"effect":   pulumi.String("NoExecute"),
						},
					},
					"ha": pulumi.Map{
						"enabled": pulumi.Bool(true),
						"raft": pulumi.Map{
							"enabled": pulumi.Bool(true),
							"config":  vaultServerConfig,
						},
					},
					"serviceAccount": pulumi.Map{
						"create": pulumi.Bool(false),
						"name":   serviceAccount.Metadata.Name(),
					},
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{namespace, serviceAccount, key}))

		lbVault, err := corev1.NewService(ctx, "vault-lb", &corev1.ServiceArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Namespace: namespace.Metadata.Name(),
				Labels: pulumi.StringMap{
					"app.kubernetes.io/name": pulumi.String("vault"),
				},
				Annotations: pulumi.StringMap{
					"service.beta.kubernetes.io/aws-load-balancer-type":            pulumi.String("external"),
					"service.beta.kubernetes.io/aws-load-balancer-nlb-target-type": pulumi.String("instance"),
					"service.beta.kubernetes.io/aws-load-balancer-scheme":          pulumi.String("internal"),
				},
				Name: pulumi.String("vault-lb"),
			},
			Spec: &corev1.ServiceSpecArgs{
				Ports: corev1.ServicePortArray{
					corev1.ServicePortArgs{
						Port:       pulumi.Int(8200),
						TargetPort: pulumi.Int(8200),
					},
				},
				Selector: pulumi.StringMap{
					"app.kubernetes.io/name":     pulumi.String("vault"),
					"app.kubernetes.io/instance": pulumi.String("vault"),
					"component":                  pulumi.String("server"),
				},
				Type: pulumi.String("LoadBalancer"),
				LoadBalancerSourceRanges: pulumi.StringArray{
					pulumi.String("10.100.0.0/16"),
					pulumi.String("10.10.0.0/16"),
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{chart, namespace}))

		endpoint := lbVault.Status.ApplyT(
			func(status *corev1.ServiceStatus) *string {
				if status.LoadBalancer.Ingress != nil {
					ingress := status.LoadBalancer.Ingress[0]
					if ingress.Hostname != nil {
						return ingress.Hostname
					}
					return ingress.Ip
				}

				return nil
			}).(pulumi.StringPtrOutput)

		ctx.Export("endpoint", endpoint)
		// create a simple record for vault.fynbos.dev for now as we only have one instance.
		_, err = route53.NewRecord(ctx, "vault1.fynbos.cloud", &route53.RecordArgs{
			ZoneId: dnsZoneId,
			Name:   pulumi.String("vault1"),
			Type:   pulumi.String("CNAME"),
			Ttl:    pulumi.Int(300),
			Records: pulumi.StringArray{
				endpoint.Elem(),
			},
		})

		return err
	})
}

func createKmsPolicy(ctx *pulumi.Context, roleArn pulumi.StringPtrInput, accountId pulumi.StringPtrInput) pulumi.StringOutput {
	return pulumi.All(roleArn, accountId).ApplyT(func(args []interface{}) (string, error) {
		roleArn := args[0].(string)
		aid := args[1].(string)
		doc, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
			Version: utils.StringPtr("2012-10-17"),
			Statements: []iam.GetPolicyDocumentStatement{
				{
					Principals: []iam.GetPolicyDocumentStatementPrincipal{
						{
							Type: "AWS",
							Identifiers: []string{
								roleArn,
							},
						},
					},
					Effect: utils.StringPtr("Allow"),
					Actions: []string{
						"kms:Encrypt",
						"kms:Decrypt",
						"kms:DescribeKey",
					},
					Resources: []string{
						"*",
					},
				},
				{
					Principals: []iam.GetPolicyDocumentStatementPrincipal{
						{
							Type: "AWS",
							Identifiers: []string{
								fmt.Sprintf("arn:aws:iam::%s:root", aid),
							},
						},
					},
					Effect: utils.StringPtr("Allow"),
					Actions: []string{
						"kms:*",
					},
					Resources: []string{
						"*",
					},
				},
			},
		})
		return doc.Json, err
	}).(pulumi.StringOutput)
}

func createVaultConfig(keyId pulumi.StringOutput) pulumi.StringOutput {
	return pulumi.All(keyId).ApplyT(func(args []interface{}) (string, error) {
		kid := args[0].(string)
		data := struct {
			KeyId string
		}{
			KeyId: kid,
		}

		vaultServerConfig, err := utils.ParseTemplateAsBytes(data, "./config.hcl")
		if err != nil {
			return "", err
		}

		return vaultServerConfig.String(), nil
	}).(pulumi.StringOutput)
}
