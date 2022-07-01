package kratos

import (
	"errors"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apiextensions"
	v1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"io/ioutil"
	"path/filepath"
	"runtime"

	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DeployKratosArgs struct {
	Domain             string
	DefaultSecret      string
	CertSecretName     pulumi.StringPtrInput
	Namespace          pulumi.StringInput
	ServiceAccountName pulumi.StringPtrInput
}

func DeployKratos(ctx *pulumi.Context, args DeployKratosArgs, opts ...pulumi.ResourceOption) (*helm.Chart, error) {

	emailTemplates, err := GetEmailTemplates()
	if err != nil {
		return nil, err
	}

	chart, err := helm.NewChart(ctx, "kratos", helm.ChartArgs{
		Version: pulumi.String("0.21.5"),
		Chart:   pulumi.String("kratos"),
		FetchArgs: &helm.FetchArgs{
			Repo: pulumi.String("https://k8s.ory.sh/helm/charts"),
		},
		Namespace: args.Namespace,
		Transformations: []yaml.Transformation{
			// Omit a resource from the Chart by transforming the specified resource definition
			// to an empty List.
			func(state map[string]interface{}, opts ...pulumi.ResourceOption) {
				name := state["metadata"].(map[string]interface{})["name"]
				if state["kind"] == "Service" && name == "kratos-courier" {
					state["apiVersion"] = "v1"
					state["kind"] = "List"
				}
			},
		},
		Values: pulumi.Map{
			"image": pulumi.Map{
				"tag": pulumi.String("v0.8.2-alpha.1"),
			},
			"autoscaling": pulumi.Map{
				"enabled": pulumi.Bool(false),
			},
			"kratos": pulumi.Map{
				"autoMigrate": pulumi.Bool(true),
				"development": pulumi.Bool(true),
				"identitySchemas": pulumi.Map{
					"identity.schema.json": pulumi.String("{\n          \"$id\": \"https://fynbos.dev/users/email-password/identity.schema.json\",\n          \"$schema\": \"http://json-schema.org/draft-07/schema#\",\n          \"title\": \"User\",\n          \"type\": \"object\",\n          \"properties\": {\n            \"traits\": {\n              \"type\": \"object\",\n              \"properties\": {\n                \"email\": {\n                  \"type\": \"string\",\n                  \"format\": \"email\",\n                  \"title\": \"E-Mail\",\n                  \"ory.sh/kratos\": {\n                    \"credentials\": {\n                      \"password\": {\n                        \"identifier\": true\n                      }\n                    },\n                    \"verification\": {\n                      \"via\": \"email\"\n                    },\n                    \"recovery\": {\n                      \"via\": \"email\"\n                    }\n                  }\n                }\n              },\n              \"required\": [\n                \"email\"\n              ],\n              \"additionalProperties\": false\n            }\n          }\n        }"),
				},
				"emailTemplates": emailTemplates,
				"config": pulumi.Map{
					"dsn": pulumi.String("cockroach://kratos@cockroachdb-public.cockroachdb:26257/kratos?sslmode=verify-full&max_conns=20&max_idle_conns=4&sslcert=/cockroach-certs/client.kratos.crt&sslkey=/cockroach-certs/client.kratos.key&sslrootcert=/cockroach-certs/ca.crt"),
					"serve": pulumi.Map{
						"public": pulumi.Map{
							"base_url": pulumi.String(args.Domain),
							"cors": pulumi.Map{
								"enabled": pulumi.Bool(true),
							},
						},
					},
					"selfservice": pulumi.Map{
						"default_browser_return_url": pulumi.String(args.Domain),
						"whitelisted_return_urls": pulumi.StringArray{
							pulumi.String(args.Domain),
						},
						"methods": pulumi.Map{
							"password": pulumi.Map{
								"enabled": pulumi.Bool(true),
							},
							"link": pulumi.Map{
								"enabled": pulumi.Bool(true),
								"config": pulumi.Map{
									"lifespan": pulumi.String("1h"),
								},
							},
						},
						"flows": pulumi.Map{
							"error": pulumi.Map{
								"ui_url": pulumi.Sprintf("%s/error", args.Domain),
							},
							"settings": pulumi.Map{
								"ui_url":                     pulumi.Sprintf("%s/settings", args.Domain),
								"privileged_session_max_age": pulumi.String("15m"),
							},
							"recovery": pulumi.Map{
								"enabled":  pulumi.Bool(true),
								"ui_url":   pulumi.Sprintf("%s/error", args.Domain),
								"lifespan": pulumi.String("15m"),
							},
							"verification": pulumi.Map{
								"enabled": pulumi.Bool(true),
								"ui_url":  pulumi.Sprintf("%s/verify", args.Domain),
								"after": pulumi.Map{
									"default_browser_return_url": pulumi.Sprintf("%s/home", args.Domain),
								},
							},
							"logout": pulumi.Map{
								"after": pulumi.Map{
									"default_browser_return_url": pulumi.String(args.Domain),
								},
							},
							"login": pulumi.Map{
								"ui_url":   pulumi.Sprintf("%s/login", args.Domain),
								"lifespan": pulumi.String("10m"),
								"after": pulumi.Map{
									"default_browser_return_url": pulumi.Sprintf("%s/home", args.Domain),
								},
							},
							"registration": pulumi.Map{
								"ui_url":   pulumi.Sprintf("%s/signup", args.Domain),
								"lifespan": pulumi.String("10m"),
								"after": pulumi.Map{
									"password": pulumi.Map{
										"hooks": pulumi.Array{
											pulumi.Map{
												"hook": pulumi.String("session"),
											},
										},
									},
								},
							},
						},
					},
					"log": pulumi.Map{
						"level":                 pulumi.String("debug"),
						"format":                pulumi.String("text"),
						"leak_sensitive_values": pulumi.Bool(true),
					},
					"secrets": pulumi.Map{
						"cookie": pulumi.StringArray{
							pulumi.String(args.DefaultSecret),
						},
						"cipher": pulumi.StringArray{
							pulumi.String(args.DefaultSecret),
						},
						"default": pulumi.StringArray{
							pulumi.String(args.DefaultSecret),
						},
					},
					"hashers": pulumi.Map{
						"argon2": pulumi.Map{
							"parallelism": pulumi.Int(1),
							"memory":      pulumi.String("128MB"),
							"iterations":  pulumi.Int(2),
							"salt_length": pulumi.Int(16),
							"key_length":  pulumi.Int(16),
						},
					},
					"identity": pulumi.Map{
						"default_schema_url": pulumi.String("file:///etc/config/identity.schema.json"),
					},
					"courier": pulumi.Map{
						"smtp": pulumi.Map{
							"connection_uri": pulumi.String("smtp://mailhog:1025/?disable_starttls=true"),
							"from_address":   pulumi.String("no-reply@fynbos.test"),
						},
						"template_override_path": pulumi.String("/conf/courier-templates"),
					},
				},
			},
			"ingress": pulumi.Map{
				"admin": pulumi.Map{
					"enabled": pulumi.Bool(false),
				},
				"public": pulumi.Map{
					"enabled": pulumi.Bool(false),
				},
			},
			"deployment": pulumi.Map{
				"serviceAccount": pulumi.Map{
					"create": pulumi.Bool(false),
					"name":   args.ServiceAccountName,
				},
				"extraVolumes": pulumi.Array{
					pulumi.Map{
						"name": pulumi.String("cockroach-certs"),
						"projected": pulumi.Map{
							"sources": pulumi.Array{
								pulumi.Map{
									"secret": pulumi.Map{
										"name": args.CertSecretName,
										"items": pulumi.Array{
											pulumi.Map{
												"key":  pulumi.String("ca.crt"),
												"path": pulumi.String("ca.crt"),
												//"mode": pulumi.Int(256),
											},
											pulumi.Map{
												"key":  pulumi.String("tls.crt"),
												"path": pulumi.String("client.kratos.crt"),
												//"mode": pulumi.Int(256),
											},
											pulumi.Map{
												"key":  pulumi.String("tls.key"),
												"path": pulumi.String("client.kratos.key"),
												//"mode": pulumi.Int(256),
											},
										},
									},
								},
							},
						},
					},
				},
				"extraVolumeMounts": pulumi.Array{
					pulumi.Map{
						"name":      pulumi.String("cockroach-certs"),
						"mountPath": pulumi.String("/cockroach-certs"),
					},
				},
			},
		},
	}, opts...)

	return chart, err
}

type DeployIngressArgs struct {
	Hostname  pulumi.StringPtrInput
	Namespace pulumi.StringInput
}

func DeployIngress(ctx *pulumi.Context, args DeployIngressArgs, opts ...pulumi.ResourceOption) error {
	_, err := apiextensions.NewCustomResource(ctx, "kratos-selfservice-mapping", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("getambassador.io/v3alpha1"),
		Kind:       pulumi.String("Mapping"),
		Metadata: v1.ObjectMetaArgs{
			Name:      pulumi.String("kratos-self-service"),
			Namespace: args.Namespace,
		},
		OtherFields: kubernetes.UntypedArgs{
			"spec": pulumi.Map{
				"hostname": args.Hostname,
				"prefix":   pulumi.String("/self-service/"),
				"rewrite":  pulumi.String("/self-service/"),
				"service":  pulumi.String("kratos-public.kratos"),
			},
		},
	}, opts...)
	if err != nil {
		return err
	}

	_, err = apiextensions.NewCustomResource(ctx, "kratos-sessions-mapping", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("getambassador.io/v3alpha1"),
		Kind:       pulumi.String("Mapping"),
		Metadata: v1.ObjectMetaArgs{
			Name:      pulumi.String("kratos-sessions"),
			Namespace: args.Namespace,
		},
		OtherFields: kubernetes.UntypedArgs{
			"spec": pulumi.Map{
				"hostname": args.Hostname,
				"prefix":   pulumi.String("/sessions/"),
				"rewrite":  pulumi.String("/sessions/"),
				"service":  pulumi.String("kratos-public.kratos"),
			},
		},
	}, opts...)
	if err != nil {
		return err
	}

	return nil
}

func GetEmailTemplates() (pulumi.Map, error) {
	_, moduleDir, _, ok := runtime.Caller(0)
	if !ok {
		return nil, errors.New("could not get directory path for kratos module")
	}

	recoveryValidSubject, err := ioutil.ReadFile(filepath.Join(filepath.Dir(moduleDir), "templates/recovery/valid/email.subject.gotmpl"))
	if err != nil {
		return nil, err
	}

	recoveryValidBody, err := ioutil.ReadFile(filepath.Join(filepath.Dir(moduleDir), "templates/recovery/valid/email.body.gotmpl"))
	if err != nil {
		return nil, err
	}

	recoveryValidPlainBody, err := ioutil.ReadFile(filepath.Join(filepath.Dir(moduleDir), "templates/recovery/valid/email.body.plaintext.gotmpl"))
	if err != nil {
		return nil, err
	}

	recoveryInvalidSubject, err := ioutil.ReadFile(filepath.Join(filepath.Dir(moduleDir), "templates/recovery/invalid/email.subject.gotmpl"))
	if err != nil {
		return nil, err
	}

	recoveryInvalidBody, err := ioutil.ReadFile(filepath.Join(filepath.Dir(moduleDir), "templates/recovery/invalid/email.body.gotmpl"))
	if err != nil {
		return nil, err
	}

	recoveryInvalidPlainBody, err := ioutil.ReadFile(filepath.Join(filepath.Dir(moduleDir), "templates/recovery/invalid/email.body.plaintext.gotmpl"))
	if err != nil {
		return nil, err
	}

	verificationValidSubject, err := ioutil.ReadFile(filepath.Join(filepath.Dir(moduleDir), "templates/verification/valid/email.subject.gotmpl"))
	if err != nil {
		return nil, err
	}

	verificationValidBody, err := ioutil.ReadFile(filepath.Join(filepath.Dir(moduleDir), "templates/verification/valid/email.body.gotmpl"))
	if err != nil {
		return nil, err
	}

	verificationValidPlainBody, err := ioutil.ReadFile(filepath.Join(filepath.Dir(moduleDir), "templates/verification/valid/email.body.plaintext.gotmpl"))
	if err != nil {
		return nil, err
	}

	verificationInvalidSubject, err := ioutil.ReadFile(filepath.Join(filepath.Dir(moduleDir), "templates/verification/invalid/email.subject.gotmpl"))
	if err != nil {
		return nil, err
	}

	verificationInvalidBody, err := ioutil.ReadFile(filepath.Join(filepath.Dir(moduleDir), "templates/verification/invalid/email.body.gotmpl"))
	if err != nil {
		return nil, err
	}

	verificationInvalidPlainBody, err := ioutil.ReadFile(filepath.Join(filepath.Dir(moduleDir), "templates/verification/invalid/email.body.plaintext.gotmpl"))
	if err != nil {
		return nil, err
	}

	return pulumi.Map{
		"recovery": pulumi.Map{
			"valid": pulumi.Map{
				"subject":   pulumi.String(recoveryValidSubject),
				"body":      pulumi.String(recoveryValidBody),
				"plainBody": pulumi.String(recoveryValidPlainBody),
			},
			"invalid": pulumi.Map{
				"subject":   pulumi.String(recoveryInvalidSubject),
				"body":      pulumi.String(recoveryInvalidBody),
				"plainBody": pulumi.String(recoveryInvalidPlainBody),
			},
		},
		"verification": pulumi.Map{
			"valid": pulumi.Map{
				"subject":   pulumi.String(verificationValidSubject),
				"body":      pulumi.String(verificationValidBody),
				"plainBody": pulumi.String(verificationValidPlainBody),
			},
			"invalid": pulumi.Map{
				"subject":   pulumi.String(verificationInvalidSubject),
				"body":      pulumi.String(verificationInvalidBody),
				"plainBody": pulumi.String(verificationInvalidPlainBody),
			},
		},
	}, nil
}
