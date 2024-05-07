# Initial configuration
# This will
# - init and unseal vault
# - configure nomad to use vault secrets
# - init vault kv v2


function init_vault () {
  # Store the vault unseal keys securely
  vault operator init | tee vault-tokens.txt

  vault operator unseal $(cat vault-tokens.txt | grep 'Key 1' | cut -d':' -f2 | xargs)
  vault operator unseal $(cat vault-tokens.txt | grep 'Key 2' | cut -d':' -f2 | xargs)
  vault operator unseal $(cat vault-tokens.txt | grep 'Key 3' | cut -d':' -f2 | xargs)

  export VAULT_TOKEN=$(cat vault-tokens.txt | grep 'Root' | cut -d':' -f2 | xargs)
}

function bootstrap_nomad_acl () {
  nomad acl bootstrap | tee nomad-acl-token.txt

  export NOMAD_TOKEN=$(cat nomad-acl-token.txt | grep 'Secret ID' | cut -d'=' -f2 | xargs)

  nomad namespace apply -description "Production environment" prod
  nomad namespace apply -description "Development environment" dev
}

function configure_vault_workload_identities () {
  vault auth enable -path 'jwt-nomad' 'jwt'

  cat << EOF | tee vault-auth-method-jwt-nomad.json
{
  "jwks_url": "http://127.0.0.1:4646/.well-known/jwks.json",
  "jwt_supported_algs": ["RS256", "EdDSA"],
  "default_role": "nomad-workloads"
}
EOF

  vault write auth/jwt-nomad/config '@vault-auth-method-jwt-nomad.json'

cat << EOF | tee vault-role-nomad-workloads.json
{
  "role_type": "jwt",
  "bound_audiences": ["vault.io"],
  "user_claim": "/nomad_job_id",
  "user_claim_json_pointer": true,
  "claim_mappings": {
    "nomad_namespace": "nomad_namespace",
    "nomad_job_id": "nomad_job_id",
    "nomad_task": "nomad_task"
  },
  "token_type": "service",
  "token_policies": ["nomad-workloads"],
  "token_period": "30m",
  "token_explicit_max_ttl": 0
}
EOF

  vault write auth/jwt-nomad/role/nomad-workloads '@vault-role-nomad-workloads.json'

  export AUTH_METHOD_ACCESSOR=$(vault auth list | grep 'jwt-nomad' | xargs | cut -d' ' -f3)

  cat << EOF| tee vault-policy-nomad-workloads.hcl
path "kv/data/{{identity.entity.aliases.$AUTH_METHOD_ACCESSOR.metadata.nomad_namespace}}/{{identity.entity.aliases.$AUTH_METHOD_ACCESSOR.metadata.nomad_job_id}}/*" {
  capabilities = ["read"]
} 

path "kv/data/{{identity.entity.aliases.$AUTH_METHOD_ACCESSOR.metadata.nomad_namespace}}/{{identity.entity.aliases.$AUTH_METHOD_ACCESSOR.metadata.nomad_job_id}}" {
  capabilities = ["read"]
}

path "kv/metadata/{{identity.entity.aliases.$AUTH_METHOD_ACCESSOR.metadata.nomad_namespace}}/*" {
  capabilities = ["list"]
}

path "kv/metadata/*" {
  capabilities = ["list"]
}

path "database-{{identity.entity.aliases.$AUTH_METHOD_ACCESSOR.metadata.nomad_namespace}}/creds/*" {
  capabilities = ["read"]
}
EOF

  vault policy write 'nomad-workloads' 'vault-policy-nomad-workloads.hcl'
}

function generate_database_root_passwords () {
  sudo apt-get install --assume-yes uuid-runtime
  vault secrets enable -version '2' 'kv'

  PROD_DB_PASSWORD=$(uuidgen)
  DEV_DB_PASSWORD=$(uuidgen)

  vault kv put -mount=kv prod/postgres/config password=$PROD_DB_PASSWORD
  vault kv put -mount=kv dev/postgres/config password=$DEV_DB_PASSWORD

  echo $PROD_DB_PASSWORD > prod-database-root-password.txt
  echo $DEV_DB_PASSWORD > dev-database-root-password.txt
}


mkdir configuration && cd configuration
export VAULT_ADDR=http://127.0.0.1:8200

init_vault
bootstrap_nomad_acl
configure_vault_workload_identities
generate_database_root_passwords

unset VAULT_TOKEN
unset NOMAD_TOKEN
unset AUTH_METHOD_ACCESSOR
