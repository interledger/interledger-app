package kubernetes

import (
	"errors"
	"path/filepath"
	"runtime"

	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func DeployCalico(ctx *pulumi.Context) (*yaml.ConfigFile, *yaml.ConfigFile, error) {
	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok { return nil, nil, errors.New("Could not get directory path for kubernetes module.") }

	calicoOperator, err := yaml.NewConfigFile(ctx, "calico-operator", &yaml.ConfigFileArgs{
		File: filepath.Join(filepath.Dir(moduleDir), "calico/operator.yaml"),
	})
	if err != nil { return nil, nil, err }

	calicoCrds, err := yaml.NewConfigFile(ctx, "calico-crds", &yaml.ConfigFileArgs{
		File: filepath.Join(filepath.Dir(moduleDir), "calico/crds.yaml"),
	})
	if err != nil { return nil, nil, err }

	return calicoOperator, calicoCrds, nil
}

// This will apply policies that:
// - global default deny all
// - allow CoreDns queries
func ConfigureDefaultNetworkPolicy(ctx *pulumi.Context, deployDependencies []pulumi.Resource) error {
	_, err := yaml.NewConfigGroup(ctx, "default-network-policy",
		&yaml.ConfigGroupArgs{
			YAML: []string{
// default deny all
				`
apiVersion: crd.projectcalico.org/v1
kind: GlobalNetworkPolicy
metadata:
  name: default-deny
spec:
  selector: all()
  types:
    - Ingress
    - Egress`,
// allow CoreDns queries
				`
apiVersion: crd.projectcalico.org/v1
kind: GlobalNetworkPolicy
metadata:
  name: allow-dns-egress
spec:
  selector: all()
  types:
    - Egress
  egress:
    - action: Allow
      protocol: UDP
      destination:
        namespaceSelector: name == "kube-system"
        ports:
          - 53`,
			},
	}, pulumi.DependsOn(deployDependencies))
	if err != nil { return err }

	return nil
}
