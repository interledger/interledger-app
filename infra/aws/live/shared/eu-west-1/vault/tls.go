package main

import (
	tls "github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func newVaultCertificate(ctx *pulumi.Context, serialNumber string) (*tls.PrivateKey, *tls.SelfSignedCert, error) {
	privateKey, err := tls.NewPrivateKey(ctx, "vault-tls-private-key", &tls.PrivateKeyArgs{
		Algorithm: pulumi.String("RSA"),
	}, pulumi.Protect(true), pulumi.AdditionalSecretOutputs([]string{"PrivateKeyPem"})) // ensure private key is encrypted
	if err != nil {
		return nil, nil, err
	}

	cert, err := tls.NewSelfSignedCert(ctx, "vault-cert", &tls.SelfSignedCertArgs{
		IsCaCertificate: pulumi.Bool(true),
		DnsNames: pulumi.StringArray{
			pulumi.String("vault.fynbos.cloud"),
			pulumi.String("*.vault.fynbos.cloud"),
		},
		KeyAlgorithm: privateKey.Algorithm,
		PrivateKeyPem: privateKey.PrivateKeyPem,
		Subjects: tls.SelfSignedCertSubjectArray{
			tls.SelfSignedCertSubjectArgs{
				SerialNumber: pulumi.String(serialNumber),
				Country: pulumi.String("IRE"),
				Organization: pulumi.String("Fynbos"),
			},
		},
		ValidityPeriodHours: pulumi.Int(24 * 365),
		AllowedUses: pulumi.StringArray{
			pulumi.String("cert_signing"),
			pulumi.String("digital_signature"),
			pulumi.String("server_auth"),
		},
		IpAddresses: pulumi.StringArray{
			pulumi.String("127.0.0.1"), // so that the vault client can be used locally
		},
	})		
	if err != nil {
		return nil, nil, err
	}

	return privateKey, cert, nil
}
