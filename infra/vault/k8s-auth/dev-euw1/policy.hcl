path "secret/data/k8s-dev-euw1/{{"{{"}}identity.entity.aliases.{{.AuthAccessor}}.metadata.service_account_namespace{{"}}"}}/{{"{{"}}identity.entity.aliases.{{.AuthAccessor}}.metadata.service_account_name{{"}}"}}/*"
{
  capabilities = ["read"]
}

path "secret/metadata/k8s-dev-euw1/{{"{{"}}identity.entity.aliases.{{.AuthAccessor}}.metadata.service_account_namespace{{"}}"}}/{{"{{"}}identity.entity.aliases.{{.AuthAccessor}}.metadata.service_account_name{{"}}"}}/*"
{
  capabilities = ["read", "list"]
}

path "pki/dev-int/sign/crdb-client"
{
  capabilities = ["read", "create", "update"]
}