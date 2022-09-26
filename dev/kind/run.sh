#!/usr/bin/env sh
set -ex

ROOIBOS_DIR=$(dirname "$0")/../../../rooibos

# Create cluster
ctlptl apply -f "$(dirname "$0")/config.yaml"

# Build backend base docker and upload to cluster
docker build "$(dirname "$0")/../../go/" -f "$(dirname "$0")/../../go/backend/Dockerfile" -t localhost:5005/backend:latest
docker push localhost:5005/backend:latest

# Build protea base docker and upload to cluster
docker build "$(dirname "$0")/../../typescript/protea" -f "$(dirname "$0")/../../typescript/protea/Dockerfile" --target dev -t localhost:5005/protea:latest
docker push localhost:5005/protea:latest

# Pull and push temporal
docker pull 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalite
docker tag 823058932981.dkr.ecr.eu-west-1.amazonaws.com/temporalite localhost:5005/temporalite
docker push localhost:5005/temporalite

# deploy
[ ! -d $ROOIBOS_DIR ] && git clone git@gitlab.com:fynbos/rooibos.git $ROOIBOS_DIR
[ -d $ROOIBOS_DIR/certmanager/envs/local ] && while ! kubectl apply -k $ROOIBOS_DIR/certmanager/envs/local; do echo "Retrying to apply certmanager resources in 5s."; sleep 5; done
[ -d $ROOIBOS_DIR/emissary/envs/local ] && while ! kubectl apply -k $ROOIBOS_DIR/emissary/envs/local; do echo "Retrying to apply emissary resources in 5s."; sleep 5; done
[ -d $ROOIBOS_DIR/cockroach/envs/local ] && while ! kubectl apply -k $ROOIBOS_DIR/cockroach/envs/local; do echo "Retrying to apply cockroach resources in 5s."; sleep 5; done
[ -d $ROOIBOS_DIR/backend/envs/local ] && while ! kubectl apply -k $ROOIBOS_DIR/backend/envs/local; do echo "Retrying to apply backend resources in 5s."; sleep 5; done
[ -d $ROOIBOS_DIR/kratos/envs/local ] && while ! kubectl apply -k $ROOIBOS_DIR/kratos/envs/local; do echo "Retrying to apply kratos resources in 5s."; sleep 5; done
[ -d $ROOIBOS_DIR/protea/envs/local ] && while ! kubectl apply -k $ROOIBOS_DIR/protea/envs/local; do echo "Retrying to apply protea resources in 5s."; sleep 5; done
[ -d $ROOIBOS_DIR/temporalite/envs/local ] && while ! kubectl apply -k $ROOIBOS_DIR/temporalite/envs/local; do echo "Retrying to apply temporal resources in 5s."; sleep 5; done
