#!/usr/bin/env sh
set -e

go install filippo.io/mkcert@v1.4.4

echo "Using certmanager root-secret as a local CA and adding it to your trust store." 
echo "This will require sudo priviledge."

CAROOT=$(mkcert -CAROOT)
kubectl get secret/root-secret --namespace cert-manager -o json | jq -r '.data."ca.crt"' | base64 -d > $CAROOT/rootCA.pem
mkcert -install
