This deploys a single Vault instance backed by EBS in the private subnet. We use Keybase to help with the unsealing process.
When Vault is deployed for the first time, it needs to be initiated and told that we will use Keybase to unseal it. If you don't have Keybase set up, see [how to sign up](#signing-up-for-keybase).
Remember to first build the Vault AMI using the `amivault` project and generate the self-signed certificate and keys using the `vaulttlscertificates` project.

### Prerequisites
The baseline project needs to be run first so that an AWS KMS key is provisioned with which we encrypt secrets in Pulumi.

## Initialising the stack
We use the an AWS Kms Key that was set up in the baseline project to encrypt the Vault certificate's private key (and any other secret). You need to initialise the stack with the following command to
`aws-vault exec <account that assumes shared role> -- pulumi stack init <name> --secrets-provider="aws://<kms-key-id>?region=eu-west-1"`

If you forgot to initialise it like that then change the secrets provider by running
`aws-vault exec <account that assumes shared role> -- pulumi stack change-secrets-provider "aws://<kms-key-id>?region=eu-west-1"`
This will update the `Pulumi.main.yaml` file with details about the external encryption provider (AWS KMS).

### Create an SSH tunnel to Vault instance
Currenlty we are using a jump box to get access to Vault. First create a key pair using the aws console. You should have the private key downloaded locally. Then create a `t2.nano` instance in the public subnet using a security group that allows inbound SSH access and all egress. Be sure to allocate a public IP to the jump box. You can then ssh into the vault instance by running
```sh
ssh-add <path to private key>
ssh -NT -L 8200:127.0.0.1:8200 -J ec2-user@<jump box ip> ec2-user@<vault instance ip>
```

### 1. Initiating Vault
You will need to get the Vault TLS certificate from the Pulumi state by running 
```sh
aws-vault exec shared-from-don -- pulumi stack output vaultTlsCert > <path-to-cert>
```
You can test that you are able to speak to Vault by running the following from your local machine
```sh
export VAULT_CACERT=<path-to-cert>
vault status
```

Then you can initialise the instance by running
```sh
vault operator init -key-shares=4 -key-threshold=2 -pgp-keys="keybase:donchangfoot,keybase:matdehaast,keybase:cairin,keybase:adrianhopebailie"
```
This will generate 4 unseal keys and we will need a minumum of 2 keys to unseal Vault.
If you get the following error
```sh
failed to initialize barrier: failed to persist keyring: mkdir /vault/core: permission denied
```
then run 
```sh
sudo chown vault:vault /vault
vault operator init -key-shares=4 -key-threshold=2 -pgp-keys="keybase:donchangfoot,keybase:matdehaast,keybase:cairin,keybase:adrianhopebailie"
```

The unseal keys and root token will be displayed now ONLY ONCE. The unseal keys are encrypted with the users public keys from Keybase and the base64 encoded.
The order matches that used in the `init` command. i.e. `donchangfoot` will be the first unseal key, `matdehaast` the second etc.
Distribute the encrypted unseal keys to the appopriate people. This needs to be safely stored in a password manager like 1Password. Also store the root token securely.

### 2. Unsealing Vault
Vault needs to be unsealed every time it is started. At least 2 of the unseal keys are necessary for this.
To get the plain-text unseal key run
```sh
echo "<base64 encoded and encrypted key>" | base64 --decode | keybase pgp decrypt
```
Then SSH into the Vault instance again and run
```sh
export VAULT_CACERT=/opt/vault/tls/vault.crt.pem
vault operator unseal
```
It will ask you to enter the plain-text unseal key.

### 3. Enabling secrets engine and AWS authentication
https://blog.gruntwork.io/a-guide-to-automating-hashicorp-vault-3-authenticating-with-an-iam-user-or-role-a3203a3ee088
We can use the AWS IAM roles to allow EC2 instances (or Lambda functions) to authenticate with Vault.
SSH into the Vault instance and enable the auth by running
```sh
export VAULT_CACERT=/opt/vault/tls/vault.crt.pem
vault auth enable aws
TODO: script that will set up policies
```


### Signing up for Keybase
You can use your package manager to install the Keybase CLI (the GUI is optional). 
When you register a device on Keybase, the public key is stored on their servers and the private key is stored locally on your device (don't choose the option to store the private key on Keybase). 
Be sure to have more than one registered device so you can recover CLI access.

You can then sign up by running `keybase signup` and following the prompts.

### Set up encryption keys
You need to set up encryption keys in Keybase. i.e. add a gpg key pair. *NB* the Vault unseal process will use the *first* key in the list from  `keybase pgp list` Be sure that this first key is not expired.
If you don't have a pgp key pair set up on Keybase then run and follow the prompts.
```sh
keybase pgp gen
```
