package main

import (
	"github.com/pulumi/pulumi-vault/sdk/v4/go/vault"
	"github.com/pulumi/pulumi-vault/sdk/v4/go/vault/ssh"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"gitlab.com/fynbos/infra/aws/modules/utils"
)

// Background on setting up SSH at scale https://www.hashicorp.com/blog/managing-ssh-access-at-scale-with-hashicorp-vault

func setupSshClientKeySigning (ctx *pulumi.Context, provider *vault.Provider) (*ssh.SecretBackendCa, error) {
	sshClientSigning, err := vault.NewMount(ctx, "ssh-client-signer", &vault.MountArgs{
			Type: pulumi.String("ssh"),
			Path: pulumi.String("ssh-client-signer"),
			SealWrap: pulumi.Bool(true),
		}, pulumi.Provider(provider))
	if err != nil { return nil, err }

	certAuth, err := ssh.NewSecretBackendCa(ctx, "ssh-client-signing-ca", &ssh.SecretBackendCaArgs{
		Backend: sshClientSigning.Path,
		GenerateSigningKey: pulumi.Bool(true), // let Vault generate a key pair for us. We will be able to get public key via Vault api.
	}, pulumi.Provider(provider))
	if err != nil { return nil, err }

	_, err = ssh.NewSecretBackendRole(ctx, "ssh-admin-role", &ssh.SecretBackendRoleArgs{
		Name:                  pulumi.String("admin-role"),
		Backend:               sshClientSigning.Path,
		KeyType:               pulumi.String("ca"),
		AllowUserCertificates: pulumi.Bool(true),
		AllowedUsers:          pulumi.String("admin"),
		DefaultExtensions:     pulumi.Map{
			"permit-pty": pulumi.String(""),
		},
		DefaultUser:           pulumi.String("admin"),
		Ttl:                   pulumi.String("30m0s"),
	}, pulumi.Provider(provider))
	if err != nil { return nil, err }

	type Data struct {}
	data := Data{}
	_, err = vault.NewPolicy(ctx, "ssh-admin-policy", &vault.PolicyArgs{
		Name: pulumi.String("ssh-admin"),
		Policy: pulumi.String(utils.ParseTemplateAsBytes(data, "./policy/ssh-admin.hcl")),
	}, pulumi.Provider(provider))
	if err != nil { return nil, err }	

	return certAuth, nil
}

func setupSshHostKeySigning (ctx *pulumi.Context, provider *vault.Provider) error {
	sshHostSigning, err := vault.NewMount(ctx, "ssh-host-signer", &vault.MountArgs{
			Type: pulumi.String("ssh"),
			Path: pulumi.String("ssh-host-signer"),
			SealWrap: pulumi.Bool(true),
			MaxLeaseTtlSeconds: pulumi.Int(365 * 24 * 60), // 1 year
		}, pulumi.Provider(provider))
	if err != nil { return err }

	_, err = ssh.NewSecretBackendCa(ctx, "ssh-host-signing-ca", &ssh.SecretBackendCaArgs{
		Backend: sshHostSigning.Path,
		GenerateSigningKey: pulumi.Bool(true),
	}, pulumi.Provider(provider))
	if err != nil { return err }

	_, err = ssh.NewSecretBackendRole(ctx, "ssh-host-signing-role", &ssh.SecretBackendRoleArgs{
		Name:                  pulumi.String("host-role"),
		Backend:               sshHostSigning.Path,
		KeyType:               pulumi.String("ca"),
		AllowHostCertificates: pulumi.Bool(true),
		AllowedDomains:        pulumi.String("fynbos.cloud"),
		AllowSubdomains:       pulumi.Bool(true),
		Ttl:                   pulumi.String("8760h"),
	}, pulumi.Provider(provider))
	if err != nil { return err }

	return nil
}