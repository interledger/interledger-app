#!/bin/bash

export CONSUL_CACERT=/etc/consul.d/certs/consul-agent-ca.pem
export CONSUL_CLIENT_CERT=/etc/consul.d/certs/dc1-client-consul-0.pem
export CONSUL_CLIENT_KEY=/etc/consul.d/certs/dc1-client-consul-0-key.pem

consul acl bootstrap | sudo tee /etc/consul.d/acl_key.txt
mkdir ~/policies
cat <<EOF | tee ~/policies/node-policy.hcl
agent_prefix "" {
  policy = "write"
}
node_prefix "" {
  policy = "write"
}
service_prefix "" {
  policy = "read"
}
session_prefix "" {
  policy = "read"
}
EOF

export CONSUL_MGMT_TOKEN=<get secretID from /etc/consul.d/acl_key.txt>
consul acl policy create \
  -token=${CONSUL_MGMT_TOKEN} \
  -name node-policy \
  -rules @policies/node-policy.hcl

consul acl token create \
  -token=${CONSUL_MGMT_TOKEN} \
  -description "node token" \
  -policy-name node-policy | sudo tee /etc/consul.d/node_acl_token.txt


export NODE_TOKEN=<get secretID from /etc/consul.d/node_acl_token.txt>
consul acl set-agent-token \
  -token=${CONSUL_MGMT_TOKEN} \
  agent $NODE_TOKEN
