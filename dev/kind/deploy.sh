#!/usr/bin/env sh
set -ex

ROOIBOS_DIR=$(dirname "$0")/../../../rooibos

[ ! -d $ROOIBOS_DIR ] && git clone git@gitlab.com:fynbos/rooibos.git $ROOIBOS_DIR
[ -d $ROOIBOS_DIR/certmanager/envs/local ] && while ! kubectl apply -k $ROOIBOS_DIR/certmanager/envs/local; do echo "Retrying to apply certmanager resources in 5s."; sleep 5; done
[ -d $ROOIBOS_DIR/emissary/envs/local ] && while ! kubectl apply -k $ROOIBOS_DIR/emissary/envs/local; do echo "Retrying to apply emissary resources in 5s."; sleep 5; done
[ -d $ROOIBOS_DIR/cockroach/envs/local ] && while ! kubectl apply -k $ROOIBOS_DIR/cockroach/envs/local; do echo "Retrying to apply cockroach resources in 5s."; sleep 5; done
[ -d $ROOIBOS_DIR/backend/envs/local ] && while ! kubectl apply -k $ROOIBOS_DIR/backend/envs/local; do echo "Retrying to apply backend resources in 5s."; sleep 5; done
[ -d $ROOIBOS_DIR/kratos/envs/local ] && while ! kubectl apply -k $ROOIBOS_DIR/kratos/envs/local; do echo "Retrying to apply kratos resources in 5s."; sleep 5; done
[ -d $ROOIBOS_DIR/protea/envs/local ] && while ! kubectl apply -k $ROOIBOS_DIR/protea/envs/local; do echo "Retrying to apply protea resources in 5s."; sleep 5; done
[ -d $ROOIBOS_DIR/temporalite/envs/local ] && while ! kubectl apply -k $ROOIBOS_DIR/temporalite/envs/local; do echo "Retrying to apply temporal resources in 5s."; sleep 5; done
