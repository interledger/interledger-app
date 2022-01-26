#!/usr/bin/env bash
COCKROACH_PATH=../../services/cockroach

rm -rf ./certs

mkdir -p certs
cp $COCKROACH_PATH/certs/ca.crt ./certs/
cockroach cert create-client pacioli --certs-dir=certs --ca-key=$COCKROACH_PATH/safe/ca.key