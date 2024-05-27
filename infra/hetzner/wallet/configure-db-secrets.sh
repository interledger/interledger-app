#!/bin/bash

cd configuration
export VAULT_TOKEN=$(cat vault-tokens.txt | grep 'Root' | cut -d':' -f2 | xargs)
export VAULT_ADDR=http://127.0.0.1:8200

export PROD_DATABASE_PASSWORD=$(cat prod-database-root-password.txt)
export DEV_DATABASE_PASSWORD=$(cat dev-database-root-password.txt)

function vault_manage_prod_database_credentials () {
  vault secrets enable -path=database-prod database

  for role in "backend" "pacioli" "temporal" "temporal_visibility" "rafiki_backend" "rafiki_auth" "kratos"
  do
    echo "role: $role"
    tee "prod-$role-connection.json" <<EOF
{
  "plugin_name": "postgresql-database-plugin",
  "allowed_roles": "$role",
  "connection_url": "postgresql://{{username}}:{{password}}@127.0.0.1:5432/$role?sslmode=disable",
  "username": "postgres",
  "password": "$PROD_DATABASE_PASSWORD"
}
EOF

    vault write database-prod/config/$role @prod-$role-connection.json
    vault write database-prod/static-roles/$role db_name=$role rotation_period="168h" \
    username=$role rotation_statements=`ALTER USER "{{name}}" WITH PASSWORD '{{password}}';`
    rm prod-$role-connection.json
  done

}

function vault_manage_dev_database_credentials () {
  vault secrets enable -path=database-dev database

  for role in "backend" "pacioli" "temporal" "temporal_visibility" "rafiki_backend" "rafiki_auth" "kratos"
  do
    tee "dev-$role-connection.json" <<EOF
{
  "plugin_name": "postgresql-database-plugin",
  "allowed_roles": "$role",
  "connection_url": "postgresql://{{username}}:{{password}}@127.0.0.1:6432/$role?sslmode=disable",
  "username": "postgres",
  "password": "$DEV_DATABASE_PASSWORD"
}
EOF

    vault write database-dev/config/$role @dev-$role-connection.json
    vault write database-dev/static-roles/$role db_name=$role rotation_period="168h" \
    username=$role rotation_statements=`ALTER USER "{{name}}" WITH PASSWORD '{{password}}';`
    rm dev-$role-connection.json
  done

}

vault_manage_prod_database_credentials
vault_manage_dev_database_credentials
