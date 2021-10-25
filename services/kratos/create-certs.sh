#!/usr/bin/env bash
rm -rf ./certs

mkdir certs
cp ./../cockroach/certs/ca.crt ./certs/
cockroach cert create-client kratos --certs-dir=certs --ca-key=./../cockroach/safe/ca.key
kubectl create secret generic cockroachdb.client.kratos --from-file=certs