package kratos

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"runtime"

	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apiextensions"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/services/ingress"
)

func DeployKratos(ctx *pulumi.Context, cert *apiextensions.CustomResource, domain string, defaultSecret string) (*helm.Chart, error) {

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
					"dsn": pulumi.String("cockroach://kratos@cockroachdb-public:26257/kratos?sslmode=verify-full&max_conns=20&max_idle_conns=4&sslcert=/cockroach-certs/client.kratos.crt&sslkey=/cockroach-certs/client.kratos.key&sslrootcert=/cockroach-certs/ca.crt"),
					"serve": pulumi.Map{
						"public": pulumi.Map{
							"base_url": pulumi.String(domain),
							"cors": pulumi.Map{
								"enabled": pulumi.Bool(true),
							},
						},
					},
					"selfservice": pulumi.Map{
						"default_browser_return_url": pulumi.String(domain),
						"whitelisted_return_urls": pulumi.StringArray{
							pulumi.String(domain),
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
								"ui_url": pulumi.Sprintf("%s/error", domain),
							},
							"settings": pulumi.Map{
								"ui_url":                     pulumi.Sprintf("%s/settings", domain),
								"privileged_session_max_age": pulumi.String("15m"),
							},
							"recovery": pulumi.Map{
								"enabled":  pulumi.Bool(true),
								"ui_url":   pulumi.Sprintf("%s/error", domain),
								"lifespan": pulumi.String("15m"),
							},
							"verification": pulumi.Map{
								"enabled": pulumi.Bool(true),
								"ui_url":  pulumi.Sprintf("%s/verify", domain),
								"after": pulumi.Map{
									"default_browser_return_url": pulumi.Sprintf("%s/home", domain),
								},
							},
							"logout": pulumi.Map{
								"after": pulumi.Map{
									"default_browser_return_url": pulumi.String(domain),
								},
							},
							"login": pulumi.Map{
								"ui_url":   pulumi.Sprintf("%s/login", domain),
								"lifespan": pulumi.String("10m"),
								"after": pulumi.Map{
									"default_browser_return_url": pulumi.Sprintf("%s/home", domain),
								},
							},
							"registration": pulumi.Map{
								"ui_url":   pulumi.Sprintf("%s/signup", domain),
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
							pulumi.String(defaultSecret),
						},
						"cipher": pulumi.StringArray{
							pulumi.String(defaultSecret),
						},
						"default": pulumi.StringArray{
							pulumi.String(defaultSecret),
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
				"extraVolumes": pulumi.Array{
					pulumi.Map{
						"name": pulumi.String("cockroach-certs"),
						"projected": pulumi.Map{
							"sources": pulumi.Array{
								pulumi.Map{
									"secret": pulumi.Map{
										"name": pulumi.String("cockroachdb-kratos"),
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
	}, pulumi.DependsOn([]pulumi.Resource{cert}))

	return chart, err
}

func DeployKratosIngress(ctx *pulumi.Context, opts ...pulumi.ResourceOption) error {
	err := ingress.DeployMapping(ctx, &ingress.MappingArgs{
		Name:     "kratos-self-service",
		Hostname: "*",
		Prefix:   "/self-service/",
		Rewrite:  "/self-service/",
		Service:  "kratos-public",
	}, opts...)
	if err != nil {
		return err
	}

	err = ingress.DeployMapping(ctx, &ingress.MappingArgs{
		Name:     "kratos-sessions",
		Hostname: "*",
		Prefix:   "/sessions/",
		Rewrite:  "/sessions/",
		Service:  "kratos-public",
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
