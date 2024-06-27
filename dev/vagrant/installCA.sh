#!/bin/bash

# abort on unbound variable
set -o nounset

# don't hide errors within pipes
set -o pipefail

# exit on error
set -o errexit

go install filippo.io/mkcert@v1.4.4
mkcert -install

mkcert -cert-file ../nomad/tls.cert.pem -key-file ../nomad/tls.key.pem fynbos.test "*.fynbos.test" local.fynbos.me "*.mgnt.fynbos.test"

sudo grep 10.9.99.10 /etc/hosts || echo "10.9.99.10 wallet.fynbos.test admin.mgnt.fynbos.test temporal.mgnt.fynbos.test local.fynbos.me" | sudo tee -a /etc/hosts
