This project is designed for our PKI setup.

## Make Vault accessible locally

Using `strongdm` connect to the vault cluster at `vault1.fynbos.cloud`, if you need permission request from someone
that can elevate your access.

```shell
sdm connect vault1 8200
```

This will bind vault to `http://127.0.0.1:8200`

## Login to Vault

Using okta
```shell
export VAULT_ADDR=http://127.0.0.1:8200
vault login -method=okta username=
```

## Run Pulumi
Next run pulumi project that will configure the project as required.

```shell
export VAULT_ADDR=http://127.0.0.1:8200
pulumi up -s fynbos/main
```
