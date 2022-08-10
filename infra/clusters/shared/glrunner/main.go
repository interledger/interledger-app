package main

import (
	"errors"
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	v1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	fynbosK8s "gitlab.com/fynbos/infra/aws/modules/kubernetes"
	"os"
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
		})
		if err != nil {
			return err
		}
		oidcProvider := k8sStack.GetStringOutput(pulumi.String("oidcProvider"))
		cfg := config.New(ctx, "fynbos")
		accountId := cfg.Require("accountId")

		namespace, err := corev1.NewNamespace(ctx, "gl-runner", &corev1.NamespaceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name: pulumi.String("gitlab-runner"),
			},
		}, pulumi.Provider(kubeProvider))
		if err != nil {
			return err
		}

		effect := "Allow"
		s3Access, err := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
			Statements: []iam.GetPolicyDocumentStatement{
				{
					Effect: &effect,
					Actions: []string{
						"s3:PutObject",
						"s3:GetObject",
						"s3:ListBucket",
						"s3:DeleteObject",
						"s3:GetBucketLocation",
					},
					Resources: []string{
						fmt.Sprintf("arn:aws:s3:::%s/*", "fynbos-gl-cache"),
						fmt.Sprintf("arn:aws:s3:::%s", "fynbos-gl-cache"),
					},
				},
			},
		})
		if err != nil {
			return err
		}

		trustPolicy := fynbosK8s.NewIamTrustPolicyDocumentV2(ctx, pulumi.String(accountId), oidcProvider, namespace.Metadata.Name().Elem(), pulumi.String("gl-runner"))
		role, err := iam.NewRole(ctx, "eks-glrunner-sa-role", &iam.RoleArgs{
			Name:             pulumi.String("eksGlrunnerSaRole"),
			Description:      pulumi.String("Service account role for gitlab runner"),
			AssumeRolePolicy: trustPolicy,
			ManagedPolicyArns: pulumi.StringArray{
				iam.ManagedPolicyAmazonEC2ContainerRegistryPowerUser,
			},
			InlinePolicies: iam.RoleInlinePolicyArray{
				iam.RoleInlinePolicyArgs{
					Name:   pulumi.String("s3AccessPolicy"),
					Policy: pulumi.String(s3Access.Json),
				},
			},
		})
		if err != nil {
			return err
		}

		runnerToken := os.Getenv("GITLAB_RUNNER_TOKEN")
		if runnerToken == "" {
			return errors.New("GITLAB_RUNNER_TOKEN not set")
		}

		serviceAccount, err := corev1.NewServiceAccount(ctx, "glrunner-sa", &corev1.ServiceAccountArgs{
			ApiVersion: pulumi.String("v1"),
			Kind:       pulumi.String("ServiceAccount"),
			Metadata: metav1.ObjectMetaArgs{
				Annotations: pulumi.StringMap{
					"eks.amazonaws.com/role-arn": role.Arn,
				},
				Name:      pulumi.String("gl-runner"),
				Namespace: namespace.Metadata.Name(),
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{
			role,
			namespace,
			kubeProvider,
		}))
		if err != nil {
			return err
		}

		gitlabRole, err := v1.NewRole(ctx, "gl-runner-role", &v1.RoleArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String("gitlab-runner"),
				Namespace: namespace.Metadata.Name(),
			},
			Rules: v1.PolicyRuleArray{
				v1.PolicyRuleArgs{
					ApiGroups: pulumi.StringArray{
						pulumi.String(""),
					},
					Resources: pulumi.StringArray{
						pulumi.String("*"),
					},
					Verbs: pulumi.StringArray{
						pulumi.String("*"),
					},
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{
			namespace,
			kubeProvider,
		}))
		if err != nil {
			return err
		}

		_, err = v1.NewRoleBinding(ctx, "gl-runner-role-binding", &v1.RoleBindingArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String("gitlab-runner"),
				Namespace: namespace.Metadata.Name(),
			},
			RoleRef: v1.RoleRefArgs{
				Kind: pulumi.String("Role"),
				Name: gitlabRole.Metadata.Name().Elem(),
			},
			Subjects: v1.SubjectArray{
				v1.SubjectArgs{
					Kind:      pulumi.String("ServiceAccount"),
					Name:      serviceAccount.Metadata.Name().Elem(),
					Namespace: namespace.Metadata.Name().Elem(),
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{
			namespace,
			kubeProvider,
			gitlabRole,
			serviceAccount,
		}))
		if err != nil {
			return err
		}

		_, err = helm.NewRelease(ctx, "gitlab", &helm.ReleaseArgs{
			Version:         pulumi.String("0.43.1"),
			Chart:           pulumi.String("gitlab-runner"),
			Namespace:       pulumi.String("gitlab-runner"),
			CreateNamespace: pulumi.BoolPtr(false),
			RepositoryOpts: &helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://charts.gitlab.io"),
			},
			Values: pulumi.Map{
				"runnerRegistrationToken": pulumi.String(runnerToken),
				"gitlabUrl":               pulumi.String("https://gitlab.com/"),
				"rbac": pulumi.Map{
					"create":             pulumi.Bool(false),
					"serviceAccountName": serviceAccount.Metadata.Name(),
				},
				"runners": pulumi.Map{
					"tags": pulumi.String("k8s"),
					"config": pulumi.String(`
[[runners]]
  name = "k8s-runner"
  executor = "kubernetes"
  [runners.cache]
    Type = "s3"
	Shared = true
	[runners.cache.s3]
      ServerAddress = "s3.amazonaws.com"
	  AuthenticationType = "iam"
	  BucketName = "fynbos-gl-cache"
	  BucketLocation = "eu-west-1"
	  Insecure = false
  [runners.kubernetes]
	namespace = "gitlab-runner"
    poll_interval = 5
    poll_timeout = 3600
	service_account = "gl-runner"
	privileged = true
	[runners.kubernetes.node_selector]
      glrunner_in_k8s = "true"
	[runners.kubernetes.node_tolerations]
	  "taint_for_gl_runner=true" = "NoExecute"
`),
				},
			},
		}, pulumi.Provider(kubeProvider), pulumi.DependsOn([]pulumi.Resource{serviceAccount}))
		if err != nil {
			return err
		}

		return nil
	})
}
