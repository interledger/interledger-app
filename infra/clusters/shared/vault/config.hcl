ui = true
listener "tcp" {
  tls_disable = 1
  address = "[::]:8200"
  cluster_address = "[::]:8201"
}
storage "raft" {
  path = "/vault/data"
}
service_registration "kubernetes" {}

seal "awskms" {
  kms_key_id = "{{.KeyId}}"
}