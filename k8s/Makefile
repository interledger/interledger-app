CI_DIR   := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
REPO_ROOT := $(shell git -C "$(CI_DIR)" rev-parse --show-toplevel)

CLUSTER_NAME := interledger-ci
KUBE_CONTEXT := kind-$(CLUSTER_NAME)
NAMESPACE    := mock-services
RELEASE      := mock-services

REGISTRY := europe-west4-docker.pkg.dev/wallet-451208/interledger-app
TAG      := ci
SERVICES := mockpti mockgatehub mockxago

CHART_DIR   := $(REPO_ROOT)/helm/mock-services
VALUES_CI   := $(CI_DIR)/kind/values.mock-services.yaml
VALUES_VALKEY := $(CI_DIR)/kind/values.valkey.yaml
KIND_CONFIG := $(CI_DIR)/kind/cluster.yaml

export HELM_KUBECONTEXT := $(KUBE_CONTEXT)

.PHONY: up down cluster build load deploy validate

## Create cluster, build images, deploy charts, and validate
up: cluster build load deploy validate

## Delete namespace and shut down the kind cluster
down:
	@echo "==> Deleting namespace $(NAMESPACE)..."
	-kubectl --context $(KUBE_CONTEXT) delete namespace $(NAMESPACE) --ignore-not-found
	@echo "==> Deleting cluster $(CLUSTER_NAME)..."
	kind delete cluster --name $(CLUSTER_NAME)

## Create the kind cluster
cluster:
	@echo "==> Creating kind cluster $(CLUSTER_NAME)..."
	kind create cluster --name $(CLUSTER_NAME) --config $(KIND_CONFIG)

## Build mock service images from source
build:
	@echo "==> Building images (tag: $(TAG))..."
	@for svc in $(SERVICES); do \
		echo "  $$svc"; \
		docker build \
			--tag $(REGISTRY)/$$svc:$(TAG) \
			--file $(REPO_ROOT)/go/mock/$$svc/Dockerfile \
			$(REPO_ROOT); \
	done

## Load built images into the kind cluster
load:
	@echo "==> Loading images into kind cluster $(CLUSTER_NAME)..."
	@for svc in $(SERVICES); do \
		echo "  $(REGISTRY)/$$svc:$(TAG)"; \
		kind load docker-image $(REGISTRY)/$$svc:$(TAG) --name $(CLUSTER_NAME); \
	done

## Deploy valkey and mock-services charts
deploy:
	@echo "==> Deploying valkey..."
	helm --kube-context $(KUBE_CONTEXT) upgrade --install valkey \
		oci://registry-1.docker.io/bitnamicharts/valkey \
		--namespace $(NAMESPACE) --create-namespace \
		--values $(VALUES_VALKEY) \
		--wait --timeout 2m
	@echo "==> Applying kustomization (mock-secrets)..."
	kubectl --context $(KUBE_CONTEXT) apply -k $(CI_DIR)/manifests
	@echo "==> Updating chart dependencies..."
	helm dependency update $(CHART_DIR)
	@echo "==> Deploying $(RELEASE)..."
	helm --kube-context $(KUBE_CONTEXT) upgrade --install $(RELEASE) $(CHART_DIR) \
		--namespace $(NAMESPACE) \
		--values $(VALUES_CI) \
		--wait --timeout 3m

## Port-forward each service and check /health
validate:
	@$(CI_DIR)/scripts/validate.sh $(NAMESPACE) $(KUBE_CONTEXT)
