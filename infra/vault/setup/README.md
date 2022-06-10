This project is designed to get the initial setup of vault complete. Once vault was created an initial root token
is generated for us. Using this root token we want to create an admin policy, setup okta connection and then 
bind admin policy to an okta admin group. Finally, the goal is to revoke the root token as a best practise.

Note: If the admin policy needs to be changed, a new root token will have to be generated using the keys that are 
currently shared between the founders. [See here](https://learn.hashicorp.com/tutorials/vault/generate-root)

## Make Vault accessible locally

Using `strongdm` connect to the vault cluster at `vault1.fynbos.cloud`, if you need permission request from someone
that can elevate your access.

```shell
sdm connect vault1 8200
```

This will bind vault to `http://127.0.0.1:8200`

## Login to Vault

If using the root token
```shell
export VAULT_ADDR=http://127.0.0.1:8200
vault login
```

Using okta
```shell
export VAULT_ADDR=http://127.0.0.1:8200
vault login -method=okta username=
```

## Okta Setup
First setup okta to be able to be used as an auth method by following this [guide](https://www.vaultproject.io/docs/auth/okta#via-the-cli-1) 

## Run Pulumi
Next run pulumi project that will configure the project as required.

```shell
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=xxx
pulumi up -s fynbos/main
```

## Revoke the root token
Finally, revoke the root token

```shell
vault token revoke ${ROOT_TOKEN}
```

