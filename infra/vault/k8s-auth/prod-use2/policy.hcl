path "secret/data/k8s-prod-use2/{{"{{"}}identity.entity.aliases.{{.AuthAccessor}}.metadata.service_account_namespace{{"}}"}}/{{"{{"}}identity.entity.aliases.{{.AuthAccessor}}.metadata.service_account_name{{"}}"}}/*"
{
  capabilities = ["read"]
}

path "secret/metadata/k8s-prod-use2/{{"{{"}}identity.entity.aliases.{{.AuthAccessor}}.metadata.service_account_namespace{{"}}"}}/{{"{{"}}identity.entity.aliases.{{.AuthAccessor}}.metadata.service_account_name{{"}}"}}/*"
{
  capabilities = ["read", "list"]
}

path "pki/prod-int/sign/crdb-client"
{
  capabilities = ["read", "create", "update"]
}

# Transit secret engine
path "transit/k8s-prod-use2/{{"{{"}}identity.entity.aliases.{{.AuthAccessor}}.metadata.service_account_namespace{{"}}"}}/{{"{{"}}identity.entity.aliases.{{.AuthAccessor}}.metadata.service_account_name{{"}}"}}/keys/*" {
  capabilities = ["create", "update", "read", "list"]
}

path "transit/k8s-prod-use2/{{"{{"}}identity.entity.aliases.{{.AuthAccessor}}.metadata.service_account_namespace{{"}}"}}/{{"{{"}}identity.entity.aliases.{{.AuthAccessor}}.metadata.service_account_name{{"}}"}}/sign/*" {
  capabilities = ["create", "update"]
}

path "transit/k8s-prod-use2/{{"{{"}}identity.entity.aliases.{{.AuthAccessor}}.metadata.service_account_namespace{{"}}"}}/{{"{{"}}identity.entity.aliases.{{.AuthAccessor}}.metadata.service_account_name{{"}}"}}/verify/*" {
  capabilities = ["create", "update"]
}