#!/bin/bash

# abort on unbound variable
set -o nounset

# don't hide errors within pipes
set -o pipefail

# exit on error
set -o errexit

go install filippo.io/mkcert@v1.4.4
mkcert -install

mkcert -cert-file ../nomad/tls.cert.pem -key-file ../nomad/tls.key.pem interledger.test "*.interledger.test" local.fynbos.me local.ilp.link "*.mgnt.interledger.test"

sudo grep 10.9.99.10 /etc/hosts || echo "10.9.99.10 interledger.test admin.mgnt.interledger.test temporal.mgnt.interledger.test local.fynbos.me  local.ilp.link rafiki.mgnt.interledger.test auth.interledger.test" | sudo tee -a /etc/hosts
