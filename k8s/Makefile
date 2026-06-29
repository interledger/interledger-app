KIND_DIR  := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
REPO_ROOT := $(shell git -C "$(KIND_DIR)" rev-parse --show-toplevel)

# ─── Cluster ──────────────────────────────────────────────────────────────────
CLUSTER_NAME := interledger
KUBE_CONTEXT := kind-$(CLUSTER_NAME)
NAMESPACE    := interledger
KIND_CONFIG  := $(KIND_DIR)/kind/cluster.yaml

# ─── Images ───────────────────────────────────────────────────────────────────
# "registry.kind" is a fake registry prefix.  Images are loaded directly via
# `kind load docker-image` and referenced with imagePullPolicy: IfNotPresent.
REGISTRY     := registry.kind
TAG          := kind
APP_SERVICES := backend protea botanist
MOCK_SERVICES := mockpti mockgatehub mockxago

# ─── Chart paths and values ───────────────────────────────────────────────────
CHART_INTERLEDGER := $(REPO_ROOT)/helm/interledger-app
CHART_MOCK        := $(REPO_ROOT)/helm/mock-services
VALUES_TRAEFIK    := $(KIND_DIR)/kind/values.traefik.yaml
VALUES_POSTGRES   := $(KIND_DIR)/kind/values.postgres.yaml
VALUES_VALKEY     := $(KIND_DIR)/kind/values.valkey.yaml
VALUES_TEMPORAL   := $(KIND_DIR)/kind/values.temporal.yaml
VALUES_KRATOS     := $(KIND_DIR)/kind/values.kratos.yaml
VALUES_MOCK       := $(KIND_DIR)/kind/values.mock-services.yaml
VALUES_APP        := $(KIND_DIR)/kind/values.interledger-app.yaml
MANIFESTS_DIR     := $(KIND_DIR)/manifests
CERTS_DIR         := $(KIND_DIR)/config/certs

export HELM_KUBECONTEXT := $(KUBE_CONTEXT)

# ─── Hosts entries needed for browser access ─────────────────────────────────
HOSTS := \
	interledger.kind \
	admin.mgnt.interledger.kind \
	identity.interledger.kind \
	mockgatehub.interledger.kind \
	mockxago.interledger.kind \
	mockpti.interledger.kind

.PHONY: help up down \
	cluster \
	certs certs-apply print-certs \
	trust print-trust untrust \
	trust-debian print-trust-debian untrust-debian \
	trust-arch print-trust-arch untrust-arch \
	hosts \
	build-app load-app \
	build-mock load-mock \
	build load \
	deploy-traefik deploy-infra deploy-services deploy-mock deploy-app deploy \
	manifests \
	validate status

## Show this help message
help:
	@echo "make <target>"
	@echo ""
	@echo "Full lifecycle:"
	@echo "  up                  Create cluster, build + load all images, deploy everything"
	@echo "  down                Delete all resources and the kind cluster"
	@echo ""
	@echo "Cluster:"
	@echo "  cluster             Create the kind cluster ($(CLUSTER_NAME))"
	@echo ""
	@echo "TLS:"
	@echo "  certs               Generate self-signed TLS certs for *.interledger.kind"
	@echo "  certs-apply         Create/update the kind-tls K8s secret from the generated cert"
	@echo "  print-certs         Show cert details"
	@echo "  trust               Trust the cert (macOS)"
	@echo "  trust-debian        Trust the cert (Debian/Ubuntu)"
	@echo "  trust-arch          Trust the cert (Arch Linux)"
	@echo "  untrust             Remove the cert from trust (macOS)"
	@echo "  untrust-debian      Remove the cert from trust (Debian/Ubuntu)"
	@echo "  untrust-arch        Remove the cert from trust (Arch Linux)"
	@echo ""
	@echo "Hosts:"
	@echo "  hosts               Add *.interledger.kind entries to /etc/hosts (requires sudo)"
	@echo ""
	@echo "Images:"
	@echo "  build               Build all images (app + mock services)"
	@echo "  load                Load all images into the kind cluster"
	@echo "  build-app           Build backend, protea, botanist images"
	@echo "  load-app            Load app images into the kind cluster"
	@echo "  build-mock          Build mock service images"
	@echo "  load-mock           Load mock service images into the kind cluster"
	@echo ""
	@echo "Deploy:"
	@echo "  deploy              Deploy everything (traefik, infra, services, mock, app)"
	@echo "  deploy-traefik      Deploy Traefik ingress controller"
	@echo "  deploy-infra        Deploy PostgreSQL and Valkey"
	@echo "  deploy-services     Deploy Kratos and Temporal"
	@echo "  deploy-mock         Deploy mock services"
	@echo "  deploy-app          Deploy interledger-app"
	@echo "  manifests           Apply kustomize manifests (secrets, IngressRoutes, etc.)"
	@echo ""
	@echo "Validate:"
	@echo "  validate            Port-forward each service and check /health"
	@echo "  status              Show pod status in the $(NAMESPACE) namespace"

# ─── Full lifecycle ───────────────────────────────────────────────────────────

## Create cluster, generate certs, build + load images, deploy everything
up: cluster certs certs-apply hosts build load deploy validate

## Delete all resources and the kind cluster
down:
	@echo "==> Deleting kind cluster $(CLUSTER_NAME)..."
	kind delete cluster --name $(CLUSTER_NAME)

# ─── Cluster ──────────────────────────────────────────────────────────────────

## Create the kind cluster
cluster:
	@echo "==> Creating kind cluster $(CLUSTER_NAME)..."
	kind create cluster --name $(CLUSTER_NAME) --config $(KIND_CONFIG)

# ─── TLS certificates ─────────────────────────────────────────────────────────

## Generate self-signed wildcard TLS cert for *.interledger.kind
certs:
	@mkdir -p $(CERTS_DIR)
	@if [ "$(HEADLESS)" = "1" ]; then \
		openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
			-keyout $(CERTS_DIR)/kind.key -out $(CERTS_DIR)/kind.crt \
			-subj "/CN=interledger.kind/O=Interledger/C=US" \
			-config $(KIND_DIR)/config/san.cnf; \
	else \
		openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
			-keyout $(CERTS_DIR)/kind.key -out $(CERTS_DIR)/kind.crt \
			-config $(KIND_DIR)/config/san.cnf; \
	fi
	@echo "Certificate written to $(CERTS_DIR)/kind.crt"

## Upload the cert as a K8s TLS secret (requires the cluster to be running)
certs-apply:
	@echo "==> Creating/updating kind-tls secret in namespace $(NAMESPACE)..."
	kubectl --context $(KUBE_CONTEXT) create namespace $(NAMESPACE) --dry-run=client -o yaml \
		| kubectl --context $(KUBE_CONTEXT) apply -f -
	kubectl --context $(KUBE_CONTEXT) create secret tls kind-tls \
		--cert=$(CERTS_DIR)/kind.crt \
		--key=$(CERTS_DIR)/kind.key \
		--namespace=$(NAMESPACE) \
		--dry-run=client -o yaml \
		| kubectl --context $(KUBE_CONTEXT) apply -f -

print-certs:
	openssl x509 -in $(CERTS_DIR)/kind.crt -text -noout

# ─── Trust ───────────────────────────────────────────────────────────────────

trust:
	sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $(CERTS_DIR)/kind.crt

print-trust:
	sudo security find-certificate -a -c "interledger.kind" -Z /Library/Keychains/System.keychain

untrust:
	sudo security find-certificate -a -c "interledger.kind" -Z /Library/Keychains/System.keychain | \
	grep 'SHA-1 hash:' | \
	awk '{print $$NF}' | \
	xargs -I {} sudo security delete-certificate -Z {} /Library/Keychains/System.keychain

trust-debian:
	sudo cp $(CERTS_DIR)/kind.crt /usr/local/share/ca-certificates/interledger-kind.crt
	sudo update-ca-certificates
	mkdir -p $(HOME)/.pki/nssdb
	[ -f "$(HOME)/.pki/nssdb/cert9.db" ] || certutil -d sql:$(HOME)/.pki/nssdb -N --empty-password
	certutil -d sql:$(HOME)/.pki/nssdb -A -t "C,," -n "interledger.kind" -i $(CERTS_DIR)/kind.crt

print-trust-debian:
	certutil -d sql:$(HOME)/.pki/nssdb -L | grep "interledger.kind"

untrust-debian:
	sudo rm -f /usr/local/share/ca-certificates/interledger-kind.crt
	sudo update-ca-certificates
	certutil -d sql:$(HOME)/.pki/nssdb -D -n "interledger.kind"

trust-arch:
	sudo cp $(CERTS_DIR)/kind.crt /etc/ca-certificates/trust-source/anchors/interledger-kind.crt
	sudo update-ca-trust
	mkdir -p $(HOME)/.pki/nssdb
	[ -f "$(HOME)/.pki/nssdb/cert9.db" ] || certutil -d sql:$(HOME)/.pki/nssdb -N --empty-password
	certutil -d sql:$(HOME)/.pki/nssdb -A -t "C,," -n "interledger.kind" -i $(CERTS_DIR)/kind.crt

print-trust-arch:
	trust list | grep -A5 "interledger.kind"

untrust-arch:
	sudo rm -f /etc/ca-certificates/trust-source/anchors/interledger-kind.crt
	sudo update-ca-trust
	certutil -d sql:$(HOME)/.pki/nssdb -D -n "interledger.kind"

# ─── /etc/hosts ───────────────────────────────────────────────────────────────

## Add interledger.kind entries to /etc/hosts
hosts:
	@sudo sed -i.bak '/# generated by k8s make hosts/d' /etc/hosts && sudo rm -f /etc/hosts.bak
	@$(foreach host,$(HOSTS),echo "127.0.0.1 $(host) # generated by k8s make hosts" | sudo tee -a /etc/hosts > /dev/null;)
	@echo "Added $(words $(HOSTS)) entries to /etc/hosts"

# ─── Image building ───────────────────────────────────────────────────────────

## Build all application images (backend, protea, botanist)
build-app:
	@echo "==> Building backend..."
	docker build \
		--tag $(REGISTRY)/backend:$(TAG) \
		--file $(REPO_ROOT)/go/backend/Dockerfile \
		$(REPO_ROOT)/go
	@echo "==> Building protea..."
	docker build \
		--tag $(REGISTRY)/protea:$(TAG) \
		--file $(REPO_ROOT)/typescript/protea/Dockerfile \
		$(REPO_ROOT)/typescript/protea
	@echo "==> Building botanist..."
	docker build \
		--tag $(REGISTRY)/botanist:$(TAG) \
		--file $(REPO_ROOT)/typescript/botanist/Dockerfile \
		$(REPO_ROOT)/typescript/botanist

## Build all mock service images
build-mock:
	@echo "==> Building mock services..."
	@for svc in $(MOCK_SERVICES); do \
		echo "  $$svc"; \
		docker build \
			--tag $(REGISTRY)/$$svc:$(TAG) \
			--file $(REPO_ROOT)/go/mock/$$svc/Dockerfile \
			$(REPO_ROOT); \
	done

## Build all images (app + mock services)
build: build-app build-mock

## Load application images into the kind cluster
load-app:
	@echo "==> Loading app images into kind cluster $(CLUSTER_NAME)..."
	@for svc in $(APP_SERVICES); do \
		echo "  $(REGISTRY)/$$svc:$(TAG)"; \
		kind load docker-image $(REGISTRY)/$$svc:$(TAG) --name $(CLUSTER_NAME); \
	done

## Load mock service images into the kind cluster
load-mock:
	@echo "==> Loading mock service images into kind cluster $(CLUSTER_NAME)..."
	@for svc in $(MOCK_SERVICES); do \
		echo "  $(REGISTRY)/$$svc:$(TAG)"; \
		kind load docker-image $(REGISTRY)/$$svc:$(TAG) --name $(CLUSTER_NAME); \
	done

## Load all images into the kind cluster
load: load-app load-mock

# ─── Deployment ───────────────────────────────────────────────────────────────

## Apply kustomize manifests (secrets, IngressRoutes, TLSStore, soketi)
manifests:
	@echo "==> Applying kustomize manifests..."
	kubectl --context $(KUBE_CONTEXT) apply -k $(MANIFESTS_DIR)

## Deploy Traefik as a DaemonSet with hostPort bindings for 80/443
deploy-traefik:
	@echo "==> Deploying Traefik..."
	helm repo add traefik https://traefik.github.io/charts 2>/dev/null || true
	helm repo update traefik
	helm --kube-context $(KUBE_CONTEXT) upgrade --install traefik traefik/traefik \
		--namespace kube-system \
		--values $(VALUES_TRAEFIK) \
		--wait --timeout 3m

## Deploy PostgreSQL and Valkey
deploy-infra:
	@echo "==> Deploying PostgreSQL..."
	helm repo add bitnami https://charts.bitnami.com/bitnami 2>/dev/null || true
	helm repo update bitnami
	helm --kube-context $(KUBE_CONTEXT) upgrade --install postgres \
		oci://registry-1.docker.io/bitnamicharts/postgresql \
		--namespace $(NAMESPACE) --create-namespace \
		--values $(VALUES_POSTGRES) \
		--wait --timeout 3m
	@echo "==> Deploying Valkey..."
	helm --kube-context $(KUBE_CONTEXT) upgrade --install valkey \
		oci://registry-1.docker.io/bitnamicharts/valkey \
		--namespace $(NAMESPACE) \
		--values $(VALUES_VALKEY) \
		--wait --timeout 2m

## Deploy Kratos and Temporal
deploy-services:
	@echo "==> Deploying Kratos..."
	helm repo add ory https://k8s.ory.sh/helm/charts 2>/dev/null || true
	helm repo update ory
	helm --kube-context $(KUBE_CONTEXT) upgrade --install kratos ory/kratos \
		--namespace $(NAMESPACE) \
		--values $(VALUES_KRATOS) \
		--wait --timeout 5m
	@echo "==> Deploying Temporal..."
	helm repo add temporal https://go.temporal.io/helm-charts 2>/dev/null || true
	helm repo update temporal
	helm --kube-context $(KUBE_CONTEXT) upgrade --install temporal temporal/temporal \
		--namespace $(NAMESPACE) \
		--values $(VALUES_TEMPORAL) \
		--wait --timeout 5m

## Deploy mock services
deploy-mock:
	@echo "==> Updating mock-services chart dependencies..."
	helm dependency update $(CHART_MOCK)
	@echo "==> Deploying mock-services..."
	helm --kube-context $(KUBE_CONTEXT) upgrade --install mock-services $(CHART_MOCK) \
		--namespace $(NAMESPACE) \
		--values $(VALUES_MOCK) \
		--wait --timeout 3m

## Deploy interledger-app (backend, frontend, admin)
deploy-app:
	@echo "==> Updating interledger-app chart dependencies..."
	helm dependency update $(CHART_INTERLEDGER)
	@echo "==> Deploying interledger-app..."
	helm --kube-context $(KUBE_CONTEXT) upgrade --install wallet $(CHART_INTERLEDGER) \
		--namespace $(NAMESPACE) \
		--values $(VALUES_APP) \
		--wait --timeout 5m

## Deploy everything in order
deploy: deploy-traefik manifests deploy-infra deploy-services deploy-mock deploy-app

# ─── Validation / status ──────────────────────────────────────────────────────

## Show pod status in the interledger namespace
status:
	kubectl --context $(KUBE_CONTEXT) get pods --namespace $(NAMESPACE)

## Port-forward each service and check /healthz
validate:
	@$(KIND_DIR)/scripts/validate.sh $(NAMESPACE) $(KUBE_CONTEXT)
