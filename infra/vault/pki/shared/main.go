package main

import (
	"github.com/pulumi/pulumi-vault/sdk/v5/go/vault"
	"github.com/pulumi/pulumi-vault/sdk/v5/go/vault/pkisecret"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		rootPki, err := vault.NewMount(ctx, "root-mount", &vault.MountArgs{
			Path:                   pulumi.String("pki/shared"),
			Type:                   pulumi.String("pki"),
			Description:            pulumi.String("Shared root PKI"),
			DefaultLeaseTtlSeconds: pulumi.Int(60 * 60 * 24 * 365 * 20),
			MaxLeaseTtlSeconds:     pulumi.Int(60 * 60 * 24 * 365 * 20),
		})
		if err != nil {
			return err
		}

		rootCert, err := pkisecret.NewSecretBackendRootCert(ctx, "root-cert", &pkisecret.SecretBackendRootCertArgs{
			Backend:      rootPki.Path,
			Type:         pulumi.String("internal"),
			CommonName:   pulumi.String("Fynbos Shared Root CA"),
			Ttl:          pulumi.String("175320h"),
			KeyType:      pulumi.String("ed25519"),
			Organization: pulumi.String("Fynbos"),
		})
		if err != nil {
			return err
		}

		_, err = pkisecret.NewSecretBackendConfigUrls(ctx, "root-backend-urls", &pkisecret.SecretBackendConfigUrlsArgs{
			Backend: rootPki.Path,
			IssuingCertificates: pulumi.StringArray{
				pulumi.String("http://vault1.fynbos.cloud/v1/pki/ca"),
			},
			CrlDistributionPoints: pulumi.StringArray{
				pulumi.String("http://vault1.fynbos.cloud/v1/pki/crl"),
			},
		})
		if err != nil {
			return err
		}

		intPki, err := vault.NewMount(ctx, "int-mount", &vault.MountArgs{
			Path:                   pulumi.String("pki/shared-int"),
			Type:                   pulumi.String("pki"),
			Description:            pulumi.String("Shared Intermediate PKI"),
			DefaultLeaseTtlSeconds: pulumi.Int(60 * 60 * 24 * 365 * 5),
			MaxLeaseTtlSeconds:     pulumi.Int(60 * 60 * 24 * 365 * 5),
		})
		if err != nil {
			return err
		}

		// Create a CSR (Certificate Signing Request)
		// Generates a new private key and a CSR for signing the PKI Secret Backend.
		intCSR, err := pkisecret.NewSecretBackendIntermediateCertRequest(ctx, "int-csr", &pkisecret.SecretBackendIntermediateCertRequestArgs{
			Backend:      intPki.Path,
			Type:         pulumi.String("internal"),
			CommonName:   pulumi.String("Fynbos Shared Intermediate Authority"),
			KeyType:      pulumi.String("ed25519"),
			Organization: pulumi.String("Fynbos"),
		})
		if err != nil {
			return err
		}

		intSigned, err := pkisecret.NewSecretBackendRootSignIntermediate(ctx, "int-signed", &pkisecret.SecretBackendRootSignIntermediateArgs{
			Backend:      rootPki.Path,
			Csr:          intCSR.Csr,
			CommonName:   pulumi.String("Fynbos Shared Intermediate Authority"),
			Organization: pulumi.String("Fynbos"),
			Ttl:          pulumi.String("43800h"),
		})
		if err != nil {
			return err
		}

		_, err = pkisecret.NewSecretBackendIntermediateSetSigned(ctx, "int-cert", &pkisecret.SecretBackendIntermediateSetSignedArgs{
			Backend:     intPki.Path,
			Certificate: intSigned.Certificate,
		})
		if err != nil {
			return err
		}

		// Allow emissary to sign for domains for dev cluster.
		_, err = pkisecret.NewSecretBackendRole(ctx, "emissary", &pkisecret.SecretBackendRoleArgs{
			Name:             pulumi.String("emissary"),
			Backend:          intPki.Path,
			AllowBareDomains: pulumi.Bool(true),
			AllowSubdomains:  pulumi.Bool(true),
			AllowedDomains: pulumi.StringArray{
				pulumi.String("mgnt.fynbos.dev"),
			},
			MaxTtl: pulumi.String("276500"), // 32days in seconds
		})

		ctx.Export("rootCert", rootCert.Certificate)
		return nil
	})
}
