listener "tcp" {
  address = "<private-ip>:9202"
  purpose = "proxy"
  {{if .TlsDisabled }}
  tls_disable = true
  {{ else }}
  tls_disable   = false
  tls_cert_file = "/etc/boundary.crt.pem"
  tls_key_file  = "/etc/boundary.key.pem"
  {{ end }}

  #proxy_protocol_behavior = "allow_authorized"
  #proxy_protocol_authorized_addrs = "127.0.0.1"
}

worker {
  # Name attr must be unique
  name = "worker-${name_suffix}"
  public_addr = "{{ .PublicHost }}"
  description = "A default worker created for demonstration"
  controllers = [
    {{ range $ip := .ControllerIps }}
    "{{ $ip }}",
    {{ end }}
  ]
}

kms "awskms" {
  purpose = "worker-auth"
  key_id = "global_worker_auth"
  kms_key_id = "{{ .KmsWorkerKeyId }}"
}