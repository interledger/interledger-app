#!/usr/bin/env bash
rm -rf ./certs ./safe

mkdir certs
mkdir safe
cockroach cert create-ca --certs-dir=certs --ca-key=safe/ca.key
cockroach cert create-client root --certs-dir=certs --ca-key=safe/ca.key
kubectl create secret generic cockroachdb.client.root --from-file=certs
cockroach cert create-node --certs-dir=certs --ca-key=safe/ca.key localhost 127.0.0.1 cockroachdb-public cockroachdb-public.default cockroachdb-public.default.svc.cluster.local *.cockroachdb *.cockroachdb.default *.cockroachdb.default.svc.cluster.local
kubectl create secret generic cockroachdb.node --from-file=certs

