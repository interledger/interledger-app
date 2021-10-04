# Vault configuration management
This project uses Pulumi's declaritive api to configure:
- auth methods
- policies
- SSH client signing

## Prerequisites
Vault needs to have been deployed, initialised and unsealed. See the `vault` project for more details.
You then need to use Boundary to get a secure connection to Vault's api.
```sh
boundary connect -target-id $VAULT_TARGET_ID -listen-port 8200
```
The `VAULT_TARGET_ID` can be looked up using the boundary ui. It is important to make sure the `listen-port` is set to 8200 as
the OIDC redirect urls are configured to use that port.

## Auth Methods
Google is set up as the OIDC provider. A web application client needs to be configured on Google and the client secret and client id used when running the pulumi script.
**NB** The `admin` role, described below, is assigned to any Google login.

## Policies
The policies can be found in `./policy`. In summary, they are:

- admin: are allowed full CRUD access to auth methods, secrets and SSH client signing. **NB** This is the default policy that is assigned
- boundary-controller: allow boundary to issue and renew tokens so that it can use Vault as a credential broker when setting up secure connections to things like a database.
- ssh-admin: allowed to get Vault to sign ssh public key's for the `admin` role

## SSH Client signing
Vault is set up as a private certificate authority that is used to sign your local ssh public key. Vault's certificate is then trusted on
servers into which we want ssh access.

```sh
# generate a key pair if you don't have one
ssh-keygen -b 2048 -t rsa -f ~/.ssh/<key name>

# make sure you have a connection to Vault through Boundary
export VAULT_CACERT=<path to vault's certificate>
export VAULT_ADDR=https://127.0.0.1:8200 # not localhost as that will raise X509 invalid host errors
vault login -method=oidc role=admin
vault write -field=signed_key ssh-client-signer/sign/admin-role public_key=@$HOME/.ssh/<key-name>.pub > ~/.ssh/signed-key.pub
```
The signed key is valid for 30min. You can this to ssh into a machine through Boundary using
```sh
boundary connect ssh -target-id $VAULT_TARGET_ID -- -i ~/.ssh/<key name> -i .ssh/signed-key.pub -l admin
```
**NB** It is assumed that the machine that you are ssh-ing into is configured with the `admin` user.

### Configuring an SSH target
The target machine will need to be configured to trust Vault's certificate (TrustedCaPublicKey). In addition, an admin user will need to be provisioned.
We use ssh authorized principals (admin is one) so that in future we can assign different priviledges to different roles.
An example cloudinit file is shown below.
```sh
#cloud-config

package_upgrade: true
package_update: true

users:
  - default
  - name: admin
    gecos: admin
    sudo: ALL=(ALL) NOPASSWD:ALL

write_files:
  - path: /etc/ssh/auth_principals/admin
    content: admin
  - path: /etc/ssh/trusted-ca.pem
    content: {{ .TrustedCaPublicKey }}

runcmd:
  - echo 'AuthorizedPrincipalsFile /etc/ssh/auth_principals/%u' >> /etc/ssh/sshd_config
  - echo 'TrustedUserCAKeys /etc/ssh/trusted-ca.pem' >> /etc/ssh/sshd_config
  - sed -i "s/PubkeyAuthentication.*/c\PubkeyAuthentication yes/" /etc/ssh/sshd_config
  - service sshd restart

```