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
[ -d $ROOIBOS_DIR/botanist/envs/local ] && while ! kubectl apply -k $ROOIBOS_DIR/botanist/envs/local; do echo "Retrying to apply botanist resources in 5s."; sleep 5; done
[ -d $ROOIBOS_DIR/temporal/envs/local ] && while ! kubectl apply -k $ROOIBOS_DIR/temporal/envs/local; do echo "Retrying to apply temporal resources in 5s."; sleep 5; done
[ -d $ROOIBOS_DIR/mockbos/envs/local ] && while ! kubectl apply -k $ROOIBOS_DIR/mockbos/envs/local; do echo "Retrying to apply mockbos resources in 5s."; sleep 5; done


# update coredns config to resolve local.fynbos.me to emissary
if [ -d $ROOIBOS_DIR/coredns/envs/local ]
then
	echo "Updating coredns config"
	EMISSARY_SERVICE_IP=$(kubectl get --namespace=emissary service/emissary-emissary-ingress -o=json | jq -r '.spec.clusterIP')
	sed "s/{EMISSARY_SERVICE_IP}/$EMISSARY_SERVICE_IP/g" $ROOIBOS_DIR/coredns/envs/local/config-example.yaml > $ROOIBOS_DIR/coredns/envs/local/config.yaml
	kubectl apply -k $ROOIBOS_DIR/coredns/envs/local
	kubectl --namespace=kube-system rollout restart deployment/coredns
fi

RED="\033[1;31m"
echo -e "${RED}Add fynbos.test mail.fynbos.test auth.fynbos.test local.fynbos.me crdb.fynbos.test kratos.fynbos.test kratos-admin.fynbos.test temporal.fynbos.test to your hosts file."
