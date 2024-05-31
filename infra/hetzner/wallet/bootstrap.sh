#!/bin/bash

# abort on nonzero exitstatus
set -o errexit

# abort on unbound variable
set -o nounset

# don't hide errors within pipes
set -o pipefail

function init_vault () {
  # Store the vault unseal keys securely
  vault operator init | tee vault-tokens.txt

  export VAULT_TOKEN=$(cat vault-tokens.txt | grep 'Root' | cut -d':' -f2 | xargs)
}

function bootstrap_nomad_acl () {
  nomad acl bootstrap | tee nomad-acl-token.txt

  export NOMAD_TOKEN=$(cat nomad-acl-token.txt | grep 'Secret ID' | cut -d'=' -f2 | xargs)

  nomad namespace apply -description "Production environment" prod
  nomad namespace apply -description "Development environment" dev
}

function enable_vault_transit_keys () {
  vault secrets enable -path 'transit/dev/backend' transit
  vault secrets enable -path 'transit/prod/backend' transit
}

function configure_vault_workload_identities () {
  vault auth enable -path 'jwt-nomad' 'jwt'

  cat << EOF | tee vault-auth-method-jwt-nomad.json
{
  "jwks_url": "http://0.0.0.0:4646/.well-known/jwks.json",
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

  local auth_method_accessor=$(vault auth list | grep 'jwt-nomad' | xargs | cut -d' ' -f3)

  cat << EOF| tee vault-policy-nomad-workloads.hcl
path "kv/data/{{identity.entity.aliases.${auth_method_accessor}.metadata.nomad_namespace}}/{{identity.entity.aliases.$AUTH_METHOD_ACCESSOR.metadata.nomad_job_id}}/*" {
  capabilities = ["read"]
} 

path "kv/data/{{identity.entity.aliases.${auth_method_accessor}.metadata.nomad_namespace}}/{{identity.entity.aliases.$AUTH_METHOD_ACCESSOR.metadata.nomad_job_id}}" {
  capabilities = ["read"]
}

path "kv/metadata/{{identity.entity.aliases.${auth_method_accessor}.metadata.nomad_namespace}}/*" {
  capabilities = ["list"]
}

path "kv/metadata/*" {
  capabilities = ["list"]
}

path "database-{{identity.entity.aliases.${auth_method_accessor}.metadata.nomad_namespace}}/static-creds/*" {
  capabilities = ["read"]
}

path "transit/{{identity.entity.aliases.${auth_method_accessor}.metadata.nomad_namespace}}/{{identity.entity.aliases.$AUTH_METHOD_ACCESSOR.metadata.nomad_job_id}}/keys/*" {
    capabilities = ["create", "update", "read", "list"]
}

path "transit/{{identity.entity.aliases.${auth_method_accessor}.metadata.nomad_namespace}}/{{identity.entity.aliases.$AUTH_METHOD_ACCESSOR.metadata.nomad_job_id}}/sign/*" {
    capabilities = ["create", "update"]
}

path "transit/{{identity.entity.aliases.${auth_method_accessor}.metadata.nomad_namespace}}/{{identity.entity.aliases.$AUTH_METHOD_ACCESSOR.metadata.nomad_job_id}}/verify/*" {
    capabilities = ["create", "update"]
}
EOF

  vault policy write 'nomad-workloads' 'vault-policy-nomad-workloads.hcl'
}

function enable_vault_kv_store () {
  vault secrets enable -version '2' 'kv'
}

function vault_manage_prod_database_credentials () {
  sudo apt-get install --assume-yes uuid-runtime
  vault secrets enable -path=database-prod database
  local prod_database_password=$(uuidgen)

  for role in "backend" "pacioli" "temporal" "temporal_visibility" "rafiki_backend" "rafiki_auth" "kratos"
  do
    echo "role: $role"
    tee "prod-$role-connection.json" <<EOF
{
  "plugin_name": "postgresql-database-plugin",
  "allowed_roles": "$role",
  "connection_url": "postgresql://{{username}}:{{password}}@127.0.0.1:5432/$role?sslmode=disable",
  "username": "postgres",
  "password": "${prod_database_password}"
}
EOF

    vault write database-prod/config/$role @prod-$role-connection.json
    vault write database-prod/static-roles/$role db_name=$role rotation_period="168h" \
    username=$role rotation_statements=`ALTER USER "{{name}}" WITH PASSWORD '{{password}}';`
    rm prod-$role-connection.json
  done
  
  echo ${prod_database_password} > prod-database-root-password.txt
}

function vault_manage_dev_database_credentials () {
  vault secrets enable -path=database-dev database
  local dev_database_password=$(uuidgen)

  for role in "backend" "pacioli" "temporal" "temporal_visibility" "rafiki_backend" "rafiki_auth" "kratos"
  do
    tee "dev-$role-connection.json" <<EOF
{
  "plugin_name": "postgresql-database-plugin",
  "allowed_roles": "$role",
  "connection_url": "postgresql://{{username}}:{{password}}@127.0.0.1:6432/$role?sslmode=disable",
  "username": "postgres",
  "password": "${dev_database_password}"
}
EOF

    vault write database-dev/config/$role @dev-$role-connection.json
    vault write database-dev/static-roles/$role db_name=$role rotation_period="168h" \
    username=$role rotation_statements=`ALTER USER "{{name}}" WITH PASSWORD '{{password}}';`
    rm dev-$role-connection.json
  done

  echo ${dev_database_password} > dev-database-root-password.txt
}

function main () {
	mkdir configuration && cd configuration
	export VAULT_ADDR=http://127.0.0.1:8200
	export CONSUL_HTTP_ADDR=http://127.0.0.1:8500

	bootstrap_nomad_acl
	init_vault
	enable_vault_transit_keys
	enable_vault_kv_store
	configure_vault_workload_identities
	vault_manage_prod_database_credentials
	vault_manage_dev_database_credentials

	unset VAULT_TOKEN
	unset NOMAD_TOKEN
}

main
