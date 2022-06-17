ui = true

listener "tcp" {
  tls_disable      = 0
  address          = "[::]:8202"
  tls_cert_file    = "/etc/vault/remote-tls/tls.crt"
  tls_key_file     = "/etc/vault/remote-tls/tls.key"
  tls_ca_cert_file = "/etc/vault/remote-tls/ca.crt"
}

listener "tcp" {
  tls_disable      = 0
  address          = "[::]:8200"
  cluster_address  = "[::]:8201"
  tls_cert_file    = "/etc/vault/tls/tls.crt"
  tls_key_file     = "/etc/vault/tls/tls.key"
  tls_ca_cert_file = "/etc/vault/tls/ca.crt"
}

storage "raft" {
  path = "/vault/data"
}

service_registration "kubernetes" {}

seal "awskms" {
  kms_key_id = "{{.KeyId}}"
}