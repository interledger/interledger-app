package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/kms"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/route53"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/s3"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apiextensions"
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
				Labels: pulumi.StringMap{
					"pod-security.kubernetes.io/enforce": pulumi.String("privileged"),
					"pod-security.kubernetes.io/audit":   pulumi.String("privileged"),
					"pod-security.kubernetes.io/warn":    pulumi.String("privileged"),
				},
			},
		}, pulumi.Provider(kubeProvider))
		if err != nil {
			return err
		}

		// Create KMSIssuer for internal Vault TLS
		kmsIssuer, err := apiextensions.NewCustomResource(ctx, "vault-kms-issuer", &apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("cert-manager.skyscanner.net/v1alpha1"),
			Kind:       pulumi.String("KMSIssuer"),
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String("vault-kms-issuer"),
				Namespace: namespace.Metadata.Name(),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": pulumi.Map{
					"keyId":      pulumi.String("alias/kms-issuer"),
					"commonName": pulumi.String("Fynbos Vault Root CA"),
					"duration":   pulumi.String("87600h"),
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{namespace}))

		cert, err := apiextensions.NewCustomResource(ctx, "vault-tls-cert", &apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("cert-manager.io/v1"),
			Kind:       pulumi.String("Certificate"),
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String("vault-tls-cert"),
				Namespace: namespace.Metadata.Name(),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": pulumi.Map{
					"secretName":  pulumi.String("vault-tls"),
					"duration":    pulumi.String("720h"),
					"renewBefore": pulumi.String("168h"),
					"usages": pulumi.StringArray{
						pulumi.String("server auth"),
						pulumi.String("client auth"),
					},
					"privateKey": pulumi.Map{
						"algorithm": pulumi.String("RSA"),
						"size":      pulumi.Int(2048),
					},
					"commonName": pulumi.String("vault"),
					"subject": pulumi.Map{
						"organizations": pulumi.StringArray{
							pulumi.String("Cockroach"),
						},
					},
					"dnsNames": pulumi.StringArray{
						pulumi.String("localhost"),
						pulumi.String("127.0.0.1"),
						pulumi.String("vault-internal"),
						pulumi.String("*.vault-internal"),
						pulumi.String("*.vault-internal.vault"),
						pulumi.String("*.vault-internal.vault.svc.cluster.local"),
					},
					"issuerRef": pulumi.Map{
						"name":  kmsIssuer.Metadata.Name(),
						"kind":  kmsIssuer.Kind,
						"group": pulumi.String("cert-manager.skyscanner.net"),
					},
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{namespace, kmsIssuer}))

		remoteCert, err := apiextensions.NewCustomResource(ctx, "vault-remote-cert", &apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("cert-manager.io/v1"),
			Kind:       pulumi.String("Certificate"),
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String("vault-remote-cert"),
				Namespace: namespace.Metadata.Name(),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": pulumi.Map{
					"secretName":  pulumi.String("vault-remote-tls"),
					"duration":    pulumi.String("2160h"),
					"renewBefore": pulumi.String("360h"),
					"usages": pulumi.StringArray{
						pulumi.String("server auth"),
						pulumi.String("client auth"),
					},
					"privateKey": pulumi.Map{
						"algorithm": pulumi.String("RSA"),
						"size":      pulumi.Int(2048),
					},
					"dnsNames": pulumi.StringArray{
						pulumi.String("vault1.fynbos.cloud"),
					},
					"issuerRef": pulumi.Map{
						"name":  pulumi.String("fynbos-cloud-issuer"),
						"kind":  pulumi.String("ClusterIssuer"),
						"group": pulumi.String("cert-manager.io"),
					},
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{namespace}))

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
				"global": pulumi.Map{
					"tlsDisable": pulumi.Bool(false),
				},
				"server": pulumi.Map{
					"updateStrategyType": pulumi.String("RollingUpdate"),
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
					"volumes": pulumi.MapArray{
						pulumi.Map{
							"name": pulumi.String("vault-tls"),
							"secret": pulumi.Map{
								"secretName": pulumi.String("vault-tls"),
							},
						},
						pulumi.Map{
							"name": pulumi.String("vault-remote-tls"),
							"secret": pulumi.Map{
								"secretName": pulumi.String("vault-remote-tls"),
							},
						},
						pulumi.Map{
							"name": pulumi.String("certwatcher"),
						},
					},
					"volumeMounts": pulumi.MapArray{
						pulumi.Map{
							"name":      pulumi.String("vault-tls"),
							"mountPath": pulumi.String("/etc/vault/tls/"),
						},
						pulumi.Map{
							"name":      pulumi.String("vault-remote-tls"),
							"mountPath": pulumi.String("/etc/vault/remote-tls/"),
						},
					},
					"extraContainers": pulumi.MapArray{
						pulumi.Map{
							"name":            pulumi.String("watch-certs"),
							"image":           pulumi.Sprintf("%s.dkr.ecr.eu-west-1.amazonaws.com/certwatcher:3.17.0", accountId),
							"imagePullPolicy": pulumi.String("Always"),
							"args": pulumi.StringArray{
								pulumi.String("/etc/vault/remote-tls/* /etc/vault/tls/*"), // watched folders
								pulumi.String("60"),                     // interval to validate checksums of watched folders
								pulumi.String("vault"),                  // process name to SIGHUP
								pulumi.String("/etc/vault/certwatcher"), // folder for checksums files
							},
							"volumeMounts": pulumi.MapArray{
								pulumi.Map{
									"name":      pulumi.String("vault-tls"),
									"mountPath": pulumi.String("/etc/vault/tls/"),
								},
								pulumi.Map{
									"name":      pulumi.String("vault-remote-tls"),
									"mountPath": pulumi.String("/etc/vault/remote-tls/"),
								},
								pulumi.Map{
									"name":      pulumi.String("certwatcher"),
									"mountPath": pulumi.String("/etc/vault/certwatcher/"),
								},
							},
						},
					},
					"shareProcessNamespace": pulumi.Bool(true),
				},
				"injector": pulumi.Map{
					"port":        pulumi.Int(10285),
					"hostNetwork": pulumi.Bool(true),
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{namespace, serviceAccount, key, cert, remoteCert}))

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
						TargetPort: pulumi.Int(8202),
					},
				},
				Selector: pulumi.StringMap{
					"app.kubernetes.io/name":     pulumi.String("vault"),
					"app.kubernetes.io/instance": pulumi.String("vault"),
					"component":                  pulumi.String("server"),
				},
				Type: pulumi.String("LoadBalancer"),
				LoadBalancerSourceRanges: pulumi.StringArray{
					pulumi.String("10.100.0.0/16"), // shared-euw1
					pulumi.String("10.10.0.0/16"),  // dev-euw1
					pulumi.String("10.30.0.0/16"),  // prod-use2
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
		if err != nil {
			return err
		}

		snapshotBucket, err := s3.NewBucket(ctx, "snapshot-bucket", &s3.BucketArgs{
			Versioning: &s3.BucketVersioningArgs{
				Enabled: pulumi.Bool(true),
			},
			LifecycleRules: s3.BucketLifecycleRuleArray{
				s3.BucketLifecycleRuleArgs{
					Expiration: s3.BucketLifecycleRuleExpirationArgs{
						Days: pulumi.Int(365),
					},
					Enabled: pulumi.Bool(true),
				},
			},
			ServerSideEncryptionConfiguration: s3.BucketServerSideEncryptionConfigurationArgs{
				Rule: s3.BucketServerSideEncryptionConfigurationRuleArgs{
					BucketKeyEnabled: pulumi.Bool(true), // use bucket level kms key
					ApplyServerSideEncryptionByDefault: s3.BucketServerSideEncryptionConfigurationRuleApplyServerSideEncryptionByDefaultArgs{
						SseAlgorithm: pulumi.String("aws:kms"),
					},
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.Protect(true))
		if err != nil {
			return err
		}
		ctx.Export("snapshotBucket", snapshotBucket.Arn)

		snapshotTrustPolicy := fynbosK8s.NewIamTrustPolicyDocumentV2(ctx, pulumi.String(accountId), oidcProvider, pulumi.String("vault"), pulumi.String("vault-snapshot"))
		snapshotAccessPolicy := snapshotBucket.Arn.ApplyT(func(arn string) (string, error) {
			policy, err := fynbosK8s.NewBucketReadWriteDeleteAccessPolicy(ctx, arn)
			if err != nil {
				return "", err
			}

			return policy.Json, nil
		}).(pulumi.StringOutput)

		snapshotRole, err := iam.NewRole(ctx, "snapshot-role", &iam.RoleArgs{
			AssumeRolePolicy: snapshotTrustPolicy,
			InlinePolicies: iam.RoleInlinePolicyArray{
				iam.RoleInlinePolicyArgs{
					Name:   pulumi.String("read-write-access"),
					Policy: snapshotAccessPolicy,
				},
			},
		}, pulumi.Provider(kubeProvider))
		if err != nil {
			return err
		}
		ctx.Export("snapshotRole", snapshotRole.Arn)

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
