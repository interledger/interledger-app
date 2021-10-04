package main

import (
	"os"
	"github.com/pulumi/pulumi-vault/sdk/v4/go/vault"
	"github.com/pulumi/pulumi-vault/sdk/v4/go/vault/jwt"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main () {
	VAULT_TOKEN := os.Getenv("VAULT_TOKEN")
	VAULT_CACERT := os.Getenv("VAULT_CACERT")
	pulumi.Run(func(ctx *pulumi.Context) error {
		provider, err := vault.NewProvider(ctx, "vault", &vault.ProviderArgs{
			Address: pulumi.String("https://127.0.0.1:8200"), // use Boundary/jump box to tunnel to Vault
			CaCertFile: pulumi.String(VAULT_CACERT),
			Token: pulumi.String(VAULT_TOKEN),
		})
		if err != nil { return err }

		_, err = newAdminUserPolicy(ctx, provider)
		if err != nil { return err }

		auth, err := jwt.NewAuthBackend(ctx, "google-auth", &jwt.AuthBackendArgs{
			Description:      pulumi.String("Google"),
			Path:             pulumi.String("oidc"),
			OidcDiscoveryUrl: pulumi.String("https://accounts.google.com"),
			OidcClientId:     pulumi.String(os.Getenv("OIDC_CLIENT_ID")),
			OidcClientSecret: pulumi.String(os.Getenv("OIDC_CLIENT_SECRET")),
			DefaultRole:      pulumi.String("admin"), // TODO: change this when we set up more
			Tune: jwt.AuthBackendTuneArgs{
				DefaultLeaseTtl: pulumi.String("30m"),
				MaxLeaseTtl: pulumi.String("30m"),
			},
		})
		if err != nil {
			return err
		}

		_, err = jwt.NewAuthBackendRole(ctx, "admin-role", &jwt.AuthBackendRoleArgs{
			Backend:  auth.Path,
			RoleName: pulumi.String("admin"),
			TokenPolicies: pulumi.StringArray{
				pulumi.String("admin"),
			},
			UserClaim: pulumi.String("sub"),
			RoleType:  pulumi.String("oidc"),
			AllowedRedirectUris: pulumi.StringArray{
				pulumi.String("http://localhost:8250/oidc/callback"), // cli login
				pulumi.String("https://127.0.0.1:8200/ui/vault/auth/oidc/oidc/callback"), // web login
			},
		})
		if err != nil {
			return err
		}

		clientCertAuth, err := setupSshClientKeySigning(ctx, provider)
		if err != nil { return err }
		ctx.Export("clientCertAuthPublicKey", clientCertAuth.PublicKey) // Servers we want ssh access to must trust this certificate.

		err = setupSshHostKeySigning(ctx, provider)
		if err != nil { return err }

		_, err = newBoundaryControllerPolicy(ctx, provider)
		if err != nil { return err }

		return nil
	})
}