package kubernetes

type KubectlCluster struct {
	Server                   string  `json:"server,omitempty"`
	CertificateAuthorityData *string `json:"certificate-authority-data,omitempty"`
}

type KubectlClusterWithName struct {
	Name    string         `json:"name"`
	Cluster KubectlCluster `json:"cluster"`
}

type KubectlConfig struct {
	Kind           string                    `json:"kind"`
	ApiVersion     string                    `json:"apiVersion"`
	CurrentContext string                    `json:"current-context"`
	Clusters       []*KubectlClusterWithName `json:"clusters"`
	Contexts       []*KubectlContextWithName `json:"contexts"`
	Users          []*KubectlUserWithName    `json:"users"`
}

type KubectlContext struct {
	Cluster string `json:"cluster"`
	User    string `json:"user"`
}

type KubectlContextWithName struct {
	Name    string         `json:"name"`
	Context KubectlContext `json:"context"`
}

type KubectlUser struct {
	ClientCertificateData []byte          `json:"client-certificate-data,omitempty"`
	ClientKeyData         []byte          `json:"client-key-data,omitempty"`
	Password              string          `json:"password,omitempty"`
	Username              string          `json:"username,omitempty"`
	Token                 string          `json:"token,omitempty"`
	Exec                  KubectlUserExec `json:"exec,omitempty"`
}

type KubectlUserExec struct {
	ApiVersion string   `json:"apiVersion,omitempty"`
	Args       []string `json:"args,omitempty"`
	Command    string   `json:"command,omitempty"`
}

type KubectlUserWithName struct {
	Name string      `json:"name"`
	User KubectlUser `json:"user"`
}

func GenerateKubeconfig(name string, certData *string, server string) KubectlConfig {
	return KubectlConfig{
		Kind:           "Config",
		ApiVersion:     "v1",
		CurrentContext: "aws",
		Clusters: []*KubectlClusterWithName{
			{
				Name: "kubernetes",
				Cluster: KubectlCluster{
					CertificateAuthorityData: certData,
					Server:                   server,
				},
			},
		},
		Contexts: []*KubectlContextWithName{
			{
				Context: KubectlContext{
					Cluster: "kubernetes",
					User:    "aws",
				},
				Name: "aws",
			},
		},
		Users: []*KubectlUserWithName{
			{
				Name: "aws",
				User: KubectlUser{
					Exec: KubectlUserExec{
						ApiVersion: "client.authentication.k8s.io/v1alpha1",
						Args: []string{
							"eks",
							"get-token",
							"--cluster-name",
							name,
						},
						Command: "aws",
					},
				},
			},
		},
	}
}
