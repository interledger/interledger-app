disable_mlock = true

telemetry {
  prometheus_retention_time = "24h"
  disable_hostname = true
}

controller {
  name = "controller-${name_suffix}"
  description = "Boundary controller"

  database {
    url = "postgresql://boundary:{{ .DbPassword }}@{{ .DbEndpoint }}/boundary"
  }
}

listener "tcp" {
  address = "<private-ip>:9200"
  purpose = "api"
  {{if .TlsDisabled }}
  tls_disable = true
  {{ else }}
  tls_disable   = false
  tls_cert_file = "/etc/boundary.crt.pem"
  tls_key_file  = "/etc/boundary.key.pem"
  {{ end }}

  cors_enabled = true
  cors_allowed_origins = ["*"]
  # proxy_protocol_behavior         = "allow_authorized"
  # proxy_protocol_authorized_addrs = "127.0.0.1"
}

listener "tcp" {
  address = "<private-ip>:9201"
  purpose = "cluster"
  tls_disable = true
  # proxy_protocol_behavior         = "allow_authorized"
  # proxy_protocol_authorized_addrs = "127.0.0.1"
}

kms "awskms" {
  purpose = "root"
  key_id = "global_root"
  kms_key_id = "{{ .KmsRootKeyId }}"
}

kms "awskms" {
  purpose = "worker-auth"
  key_id = "global_worker_auth"
  kms_key_id = "{{ .KmsWorkerKeyId }}"
}

kms "awskms" {
  purpose = "recovery"
  key_id = "global_recovery"
  kms_key_id = "{{ .KmsRecoveryKeyId }}"
}
