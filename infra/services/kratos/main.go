package kratos

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/services/cockroach"
	"gitlab.com/fynbos/infra/services/ingress"
)

func DeployKratos(ctx *pulumi.Context) (*helm.Chart, error) {

	crCert, err := cockroach.CreateClientCert(ctx, &cockroach.ClientCertArgs{
		Issuer:    "ca-issuer",
		Namespace: "default",
		Name:      "kratos",
	})

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
				"tag": pulumi.String("v0.7.6-alpha.1"),
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
				"config": pulumi.Map{
					"dsn": pulumi.String("cockroach://kratos@cockroachdb-public:26257/kratos?sslmode=verify-full&max_conns=20&max_idle_conns=4&sslcert=/cockroach-certs/client.kratos.crt&sslkey=/cockroach-certs/client.kratos.key&sslrootcert=/cockroach-certs/ca.crt"),
					"serve": pulumi.Map{
						"public": pulumi.Map{
							"base_url": pulumi.String("http://fynbos.test"),
							"cors": pulumi.Map{
								"enabled": pulumi.Bool(true),
							},
						},
					},
					"selfservice": pulumi.Map{
						"default_browser_return_url": pulumi.String("http://fynbos.test"),
						"whitelisted_return_urls": pulumi.StringArray{
							pulumi.String("http://fynbos.test"),
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
								"ui_url": pulumi.String("http://fynbos.test/error"),
							},
							"settings": pulumi.Map{
								"ui_url":                     pulumi.String("http://fynbos.test/profile"),
								"privileged_session_max_age": pulumi.String("15m"),
							},
							"recovery": pulumi.Map{
								"enabled":  pulumi.Bool(true),
								"ui_url":   pulumi.String("http://fynbos.test/error"),
								"lifespan": pulumi.String("15m"),
							},
							"verification": pulumi.Map{
								"enabled": pulumi.Bool(true),
								"ui_url":  pulumi.String("http://fynbos.test/verify"),
								"after": pulumi.Map{
									"default_browser_return_url": pulumi.String("http://fynbos.test/profile"),
								},
							},
							"logout": pulumi.Map{
								"after": pulumi.Map{
									"default_browser_return_url": pulumi.String("http://fynbos.test/"),
								},
							},
							"login": pulumi.Map{
								"ui_url":   pulumi.String("http://fynbos.test/login"),
								"lifespan": pulumi.String("10m"),
								"after": pulumi.Map{
									"default_browser_return_url": pulumi.String("http://fynbos.test/profile"),
								},
							},
							"registration": pulumi.Map{
								"ui_url":   pulumi.String("http://fynbos.test/signup"),
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
						"session": pulumi.StringArray{
							pulumi.String("PLEASE-CHANGE-ME-I-AM-VERY-INSECURE"),
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
							"connection_uri": pulumi.String("smtp://mailhog:1025/"),
							"from_address":   pulumi.String("no-reply@fynbos.test"),
						},
					},
				},
			},
			"ingress": pulumi.Map{
				"admin": pulumi.Map{
					"enabled": pulumi.Bool(true),
				},
				"public": pulumi.Map{
					"enabled": pulumi.Bool(true),
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
	}, pulumi.DependsOn([]pulumi.Resource{crCert}))

	err = ingress.DeployMapping(ctx, &ingress.MappingArgs{
		Name:     "kratos-self-service",
		Hostname: "*",
		Prefix:   "/self-service/",
		Service:  "kratos-public",
	})
	err = ingress.DeployMapping(ctx, &ingress.MappingArgs{
		Name:     "kratos-sessions",
		Hostname: "*",
		Prefix:   "/sessions/",
		Service:  "kratos-public",
	})

	return chart, err
}
