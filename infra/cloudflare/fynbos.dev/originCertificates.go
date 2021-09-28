package main

import (
	"github.com/pulumi/pulumi-cloudflare/sdk/v3/go/cloudflare"
	"github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func newBoundaryCertificate(ctx *pulumi.Context) (*tls.PrivateKey ,*cloudflare.OriginCaCertificate, error) {
	privateKey, err := tls.NewPrivateKey(ctx, "boundary-cert-private-key", &tls.PrivateKeyArgs{
		Algorithm: pulumi.String("RSA"), // Defaults to 2048
	}, pulumi.Protect(true), pulumi.AdditionalSecretOutputs([]string{"PrivateKeyPem"})) // encrypt private key
	if err != nil {
		return nil, nil, err
	}

	certRequest, err := tls.NewCertRequest(ctx, "boundary-cert-request", &tls.CertRequestArgs{
		KeyAlgorithm:  privateKey.Algorithm,
		PrivateKeyPem: privateKey.PrivateKeyPem,
		DnsNames: pulumi.StringArray{
			pulumi.String("boundary.fynbos.dev"),
			pulumi.String("*.boundary.fynbos.dev"),
		},
		Subjects: tls.CertRequestSubjectArray{
			&tls.CertRequestSubjectArgs{
				Country:   pulumi.String("IRE"),
				Organization: pulumi.String("Fynbos"),
			},
		},
	})
	if err != nil {
		return nil, nil, err
	}

	cert, err := cloudflare.NewOriginCaCertificate(ctx, "boundary-cert", &cloudflare.OriginCaCertificateArgs{
		Csr: certRequest.CertRequestPem,
		Hostnames: pulumi.StringArray{
			pulumi.String("boundary.fynbos.dev"),
			pulumi.String("*.boundary.fynbos.dev"),
		},
		RequestType:       pulumi.String("origin-rsa"),
		RequestedValidity: pulumi.Int(365),
	}, pulumi.Protect(true))
	if err != nil {
		return nil, nil, err
	}

	return privateKey, cert, nil
}