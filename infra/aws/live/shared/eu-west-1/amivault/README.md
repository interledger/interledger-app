This will create an Amazon Linux 2 based machine image that has Vault installed.

An 'm5.large' machine is spun up to build the image. Logs from the build process can be enabled by setting the `enableLogging` variable to the `newImageBuilderConfiguration` function.
It creates the following directory structure for you to use
 - /opt/vault/tls
 - /opt/vault/bin
 - /opt/vault/data
 - /opt/vault/config

When using this image, the tls cert and private key must be provisioned at `/opt/vault/tls` as `vault.crt.pem` and `vault.key.pem` respectively. You are required to then supply the systemd config and start Vault.
