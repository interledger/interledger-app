package main

import (
	"errors"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
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

		trustPolicy := fynbosK8s.NewIamTrustPolicyDocumentV2(ctx, pulumi.String(accountId), oidcProvider, namespace.Metadata.Name().Elem(), pulumi.String("gl-runner"))
		role, err := iam.NewRole(ctx, "eks-glrunner-sa-role", &iam.RoleArgs{
			Name:             pulumi.String("eksGlrunnerSaRole"),
			Description:      pulumi.String("Service account role for gitlab runner"),
			AssumeRolePolicy: trustPolicy,
			ManagedPolicyArns: pulumi.StringArray{
				iam.ManagedPolicyAmazonEC2ContainerRegistryPowerUser,
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
					"create": pulumi.Bool(true),
				},
				"runners": pulumi.Map{
					"tags": pulumi.String("k8s"),
					"config": pulumi.String(`
[[runners]]
  name = "k8s-runner"
  executor = "kubernetes"
  [runners.kubernetes]
	namespace = "gitlab-runner"
    poll_interval = 5
    poll_timeout = 3600
	service_account = "gl-runner"
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
