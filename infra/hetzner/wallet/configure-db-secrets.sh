#!/bin/bash

cd configuration
export VAULT_TOKEN=$(cat vault-tokens.txt | grep 'Root' | cut -d':' -f2 | xargs)
export VAULT_ADDR=http://127.0.0.1:8200

export PROD_DATABASE_PASSWORD=$(cat prod-database-root-password.txt)
export DEV_DATABASE_PASSWORD=$(cat dev-database-root-password.txt)

function vault_manage_prod_database_credentials () {
  vault secrets enable -path=database-prod database

  tee accessdb.sql <<EOF
CREATE USER "{{name}}" WITH ENCRYPTED PASSWORD '{{password}}' VALID UNTIL
'{{expiration}}';
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO "{{name}}";
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO "{{name}}";
GRANT ALL ON SCHEMA public TO "{{name}}";
EOF

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
    vault write database-prod/roles/$role db_name=$role \
    creation_statements=@accessdb.sql default_ttl=1h max_ttl=168h
  done

}

function vault_manage_dev_database_credentials () {
  vault secrets enable -path=database-dev database

  tee accessdb.sql <<EOF
CREATE USER "{{name}}" WITH ENCRYPTED PASSWORD '{{password}}' VALID UNTIL
'{{expiration}}';
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO "{{name}}";
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO "{{name}}";
GRANT ALL ON SCHEMA public TO "{{name}}";
EOF

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
    vault write database-dev/roles/$role db_name=$role \
    creation_statements=@accessdb.sql default_ttl=1h max_ttl=168h
  done

}

vault_manage_prod_database_credentials
# vault_manage_dev_database_credentials
