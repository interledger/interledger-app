package main

import (
	"github.com/pulumi/pulumi-cloudflare/sdk/v3/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateDevClusterAccess(ctx *pulumi.Context, zoneId pulumi.IDOutput) error {
	app, err := cloudflare.NewAccessApplication(ctx, "dev-cluster-app", &cloudflare.AccessApplicationArgs{
		//CorsHeaders: cloudflare.AccessApplicationCorsHeaderArray{
		//	&cloudflare.AccessApplicationCorsHeaderArgs{
		//		AllowCredentials: pulumi.Bool(true),
		//		AllowedMethods: pulumi.StringArray{
		//			pulumi.String("GET"),
		//			pulumi.String("POST"),
		//			pulumi.String("OPTIONS"),
		//		},
		//		AllowedOrigins: pulumi.StringArray{
		//			pulumi.String("https://example.com"),
		//		},
		//		MaxAge: pulumi.Int(10),
		//	},
		//},
		AutoRedirectToIdentity: pulumi.Bool(true),
		Domain:                 pulumi.String("dev.fynbos.dev"),
		Name:                   pulumi.String("Dev Cluster"),
		SessionDuration:        pulumi.String("24h"),
		Type:                   pulumi.String("self_hosted"),
		ZoneId:                 zoneId,
		// Allowed IDP was manually created so we don't need to store secrets when deploying it
		AllowedIdps: pulumi.StringArray{
			pulumi.String("4271ed52-0611-4386-b1f8-f8e8c1c8391d"),
		},
	})
	if err != nil {
		return err
	}

	_, err = cloudflare.NewAccessPolicy(ctx, "dev-cluster-access-policy", &cloudflare.AccessPolicyArgs{
		ApplicationId: app.ID(),
		ZoneId:        zoneId,
		Name:          pulumi.String("Dev cluster policy"),
		Precedence:    pulumi.Int(1),
		Decision:      pulumi.String("allow"),
		Includes: cloudflare.AccessPolicyIncludeArray{
			&cloudflare.AccessPolicyIncludeArgs{
				EmailDomains: pulumi.StringArray{
					pulumi.String("fynbos.dev"),
				},
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func CreateDevMailAccess(ctx *pulumi.Context, zoneId pulumi.IDOutput) error {
	app, err := cloudflare.NewAccessApplication(ctx, "dev-mail-app", &cloudflare.AccessApplicationArgs{
		AutoRedirectToIdentity: pulumi.Bool(true),
		Domain:                 pulumi.String("mail.fynbos.dev"),
		Name:                   pulumi.String("Dev Cluster"),
		SessionDuration:        pulumi.String("24h"),
		Type:                   pulumi.String("self_hosted"),
		ZoneId:                 zoneId,
		// Allowed IDP was manually created so we don't need to store secrets when deploying it
		AllowedIdps: pulumi.StringArray{
			pulumi.String("4271ed52-0611-4386-b1f8-f8e8c1c8391d"),
		},
	})
	if err != nil {
		return err
	}

	_, err = cloudflare.NewAccessPolicy(ctx, "dev-mail-access-policy", &cloudflare.AccessPolicyArgs{
		ApplicationId: app.ID(),
		ZoneId:        zoneId,
		Name:          pulumi.String("Dev mail policy"),
		Precedence:    pulumi.Int(1),
		Decision:      pulumi.String("allow"),
		Includes: cloudflare.AccessPolicyIncludeArray{
			&cloudflare.AccessPolicyIncludeArgs{
				EmailDomains: pulumi.StringArray{
					pulumi.String("fynbos.dev"),
				},
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func CreateDevRetoolAccess(ctx *pulumi.Context, zoneID pulumi.IDOutput) error {
	app, err := cloudflare.NewAccessApplication(ctx, "dev-retool-app", &cloudflare.AccessApplicationArgs{
		AutoRedirectToIdentity: pulumi.Bool(true),
		Domain:                 pulumi.String("retool.fynbos.dev"),
		Name:                   pulumi.String("Dev Cluster"),
		SessionDuration:        pulumi.String("24h"),
		Type:                   pulumi.String("self_hosted"),
		ZoneId:                 zoneID,
		// Allowed IDP was manually created so we don't need to store secrets when deploying it
		AllowedIdps: pulumi.StringArray{
			pulumi.String("4271ed52-0611-4386-b1f8-f8e8c1c8391d"),
		},
	})
	if err != nil {
		return err
	}

	_, err = cloudflare.NewAccessPolicy(ctx, "dev-retool-access-policy", &cloudflare.AccessPolicyArgs{
		ApplicationId: app.ID(),
		ZoneId:        zoneID,
		Name:          pulumi.String("Dev retool policy"),
		Precedence:    pulumi.Int(1),
		Decision:      pulumi.String("allow"),
		Includes: cloudflare.AccessPolicyIncludeArray{
			&cloudflare.AccessPolicyIncludeArgs{
				EmailDomains: pulumi.StringArray{
					pulumi.String("fynbos.dev"),
				},
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func CreateDevPayAccess(ctx *pulumi.Context, zoneID pulumi.IDOutput) error {
	app, err := cloudflare.NewAccessApplication(ctx, "dev-pay-app", &cloudflare.AccessApplicationArgs{
		AutoRedirectToIdentity: pulumi.Bool(true),
		Domain:                 pulumi.String("pay.fynbos.dev"),
		Name:                   pulumi.String("Dev Cluster"),
		SessionDuration:        pulumi.String("24h"),
		Type:                   pulumi.String("self_hosted"),
		ZoneId:                 zoneID,
		// Allowed IDP was manually created so we don't need to store secrets when deploying it
		AllowedIdps: pulumi.StringArray{
			pulumi.String("4271ed52-0611-4386-b1f8-f8e8c1c8391d"),
		},
	})
	if err != nil {
		return err
	}

	_, err = cloudflare.NewAccessPolicy(ctx, "dev-pay-access-policy", &cloudflare.AccessPolicyArgs{
		ApplicationId: app.ID(),
		ZoneId:        zoneID,
		Name:          pulumi.String("Dev pay policy"),
		Precedence:    pulumi.Int(1),
		Decision:      pulumi.String("allow"),
		Includes: cloudflare.AccessPolicyIncludeArray{
			&cloudflare.AccessPolicyIncludeArgs{
				EmailDomains: pulumi.StringArray{
					pulumi.String("fynbos.dev"),
				},
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func CreateEu1DevClusterAccess(ctx *pulumi.Context, zoneId pulumi.IDOutput) error {
	app, err := cloudflare.NewAccessApplication(ctx, "eu1-dev-cluster-app", &cloudflare.AccessApplicationArgs{
		AutoRedirectToIdentity: pulumi.Bool(true),
		Domain:                 pulumi.String("eu1.fynbos.dev"),
		Name:                   pulumi.String("EU1 Dev Cluster"),
		SessionDuration:        pulumi.String("24h"),
		Type:                   pulumi.String("self_hosted"),
		ZoneId:                 zoneId,
		// Allowed IDP was manually created, so we don't need to store secrets when deploying it
		AllowedIdps: pulumi.StringArray{
			pulumi.String("4271ed52-0611-4386-b1f8-f8e8c1c8391d"),
		},
	})
	if err != nil {
		return err
	}

	_, err = cloudflare.NewAccessPolicy(ctx, "eu1-dev-cluster-access-policy", &cloudflare.AccessPolicyArgs{
		ApplicationId: app.ID(),
		ZoneId:        zoneId,
		Name:          pulumi.String("Eu1 Dev cluster policy"),
		Precedence:    pulumi.Int(1),
		Decision:      pulumi.String("allow"),
		Includes: cloudflare.AccessPolicyIncludeArray{
			&cloudflare.AccessPolicyIncludeArgs{
				EmailDomains: pulumi.StringArray{
					pulumi.String("fynbos.dev"),
				},
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}
