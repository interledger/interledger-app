ui = true

listener "tcp" {
  address         = "0.0.0.0:8200"
  cluster_address = "0.0.0.0:8201"
  tls_cert_file   = "/opt/vault/tls/vault.crt.pem"
  tls_key_file    = "/opt/vault/tls/vault.key.pem"
}


storage "file" {
  path    = "/opt/vault/data"
}