package retool

import (
	"errors"
	"path/filepath"
	"runtime"

	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	utils "gitlab.com/fynbos/infra/aws/modules/utils"
	"gitlab.com/fynbos/infra/services/ingress"
)

func DeployRetool(ctx *pulumi.Context, hn string) (*helm.Chart, error) {
	err := deployConfigmap(ctx)
	if err != nil {
		return nil, err
	}

	values := pulumi.Map{
		"config": pulumi.Map{
			"licenseKey":         pulumi.String("SSOP_0645bcae-fd52-4ce5-ab33-8001d81b7d38"),
			"useInsecureCookies": pulumi.Bool(true),
			"encryptionKey":      pulumi.String("9i/IwIhNuuXczH9q6mU8YyqzVPY3eM3r7H3qBSdKJk9XHJkSLj1C3Zeey24zfMbdzlEzWsieZ5G+4vtJkK/F4w=="),
			"jwtSecret":          pulumi.String("9i/IwIhNuuXczH9q6mU8YyqzVPY3eM3r7H3qBSdKJk9XHJkSLj1C3Zeey24zfMbdzlEzWsieZ5G+4vtJkK/F4w=="),
		},
		"image": pulumi.Map{
			"repository": pulumi.String("tryretool/backend"),
			"tag":        pulumi.String("latest"),
		},
		"env": pulumi.Map{
			"PROTO_DIRECTORY_PATH": pulumi.String("/retool_backend/protos"),
		},
		"extraVolumeMounts": pulumi.Array{
			pulumi.Map{
				"name":      pulumi.String("retool-protos"),
				"mountPath": pulumi.String("/retool_backend/protos"),
			},
		},
		"extraVolumes": pulumi.Array{
			pulumi.Map{
				"name": pulumi.String("retool-protos"),
				"configMap": pulumi.Map{
					"name": pulumi.String("retool-config"),
					"items": pulumi.Array{
						pulumi.Map{
							"key":  pulumi.String("backend"),
							"path": pulumi.String("backend.proto"),
						},
					},
				},
			},
		},
	}

	chart, err := helm.NewChart(ctx, "retool", helm.ChartArgs{
		Version: pulumi.String("4.8.0"),
		Chart:   pulumi.String("retool"),
		FetchArgs: &helm.FetchArgs{
			Repo: pulumi.String("https://charts.retool.com"),
		},
		Values: values,
	})
	if err != nil {
		return nil, err
	}

	err = deployIngress(ctx, hn)
	if err != nil {
		return nil, err
	}

	return chart, nil
}

func deployIngress(ctx *pulumi.Context, hn string) error {
	err := ingress.DeployMapping(ctx, &ingress.MappingArgs{
		Name:            "retool",
		Hostname:        hn,
		Prefix:          "/",
		Rewrite:         "/",
		Service:         "retool:3000",
		EnableWebsocket: true,
	})
	if err != nil {
		return err
	}

	return nil
}

func deployConfigmap(ctx *pulumi.Context) error {
	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok {
		return errors.New("Could not get directory path for utils/testing.")
	}
	backendProtoPath := filepath.Join(filepath.Dir(moduleDir), "../../../proto/backend/v1/backend.proto")
	backendProto := utils.ParseTemplate(struct{}{}, backendProtoPath)
	_, err := corev1.NewConfigMap(ctx, "retool-proto-configmaps", &corev1.ConfigMapArgs{
		ApiVersion: pulumi.String("v1"),
		Kind:       pulumi.String("ConfigMap"),
		Metadata: &metav1.ObjectMetaArgs{
			Labels: pulumi.StringMap{
				"app": pulumi.String("retool"),
			},
			Name: pulumi.String("retool-config"),
		},
		BinaryData: pulumi.StringMap{
			"backend": pulumi.String(backendProto),
		},
	})
	if err != nil {
		return err
	}

	return nil
}
