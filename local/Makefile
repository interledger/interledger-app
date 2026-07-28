COMPOSE := docker compose

# Monitoring (Prometheus + Grafana) is opt-in: monitoring.yaml is intentionally
# left out of docker-compose.yaml's `include:` graph so `make all` /
# --profile "*" can never start it. It is loaded explicitly (via -f) only by the
# `monitoring` targets and by `down`/`delete-volumes`, so teardown stays complete.
MONITORING := $(COMPOSE) -f docker-compose.yaml -f monitoring.yaml

help:
	@echo "make <command>"
	@echo "e.g. make infrastructure"
	@echo ""

	@echo "help                    Show this help message"
	@echo "hosts                   Add required entries to /etc/hosts (requires sudo, includes mockgatehub.interledger.test, docs.interledger.test)"
	@echo ""

	@echo "infrastructure|infra    Traefik, Postgres, Redis"
	@echo "services|svc            Kratos, Temporal, Rafiki, Ngrok (requires infra)"
	@echo "application|app         Wallet application, backends + frontends (requires infra, svc)"
	@echo "backend|back            Just backends services"
	@echo "frontend|front          Just frontends services"
	@echo "all                     All services (infra, svc, app)"
	@echo "monitoring|mon          Prometheus + Grafana (opt-in; scrapes running services)"
	@echo "monitoring-down         Stop Prometheus + Grafana"
	@echo "unit-test-db-up         Start isolated Postgres for Go unit tests (localhost:55432)"
	@echo "unit-test-db-down       Stop isolated Postgres for Go unit tests"
	@echo ""
	@echo "build                   Build backend images"
	@echo "pull                    Pull all images"
	@echo ""
	@echo "down                    Stop all running services"
	@echo "delete-volumes|reset    Stop all services and remove associated volumes"
	@echo ""

	@echo "certs                   Generate self-signed TLS certificates for local development"
	@echo "print-certs             Print details of the self-signed TLS certificates"
	@echo ""

	@echo "trust                   Trust the self-signed TLS certificates on macOS"
	@echo "print-trust             List all self-signed TLS certificates on macOS"
	@echo "untrust                 Remove the self-signed TLS certificates from macOS"

	@echo "trust-debian           Trust the self-signed TLS certificates on Debian-based systems"
	@echo "print-trust-debian     List all self-signed TLS certificates on Debian-based systems"
	@echo "untrust-debian         Remove the self-signed TLS certificates from Debian-based systems"

	@echo "trust-arch             Trust the self-signed TLS certificates on Arch-based systems"
	@echo "print-trust-arch       List all self-signed TLS certificates on Arch-based systems"
	@echo "untrust-arch           Remove the self-signed TLS certificates from Arch-based systems"

	@echo ""

	@echo ""

.PHONY: \
	help \

	infrastructure infra \
	services svc \
	application app \
	backend back \
	frontend front \
	all \
	monitoring mon monitoring-down \
	unit-test-db-up \
	unit-test-db-down \
	build \
	pull \
	down \
	delete-volumes reset \

	certs print-certs \

	trust print-trust untrust \
	trust-debian print-trust-debian untrust-debian \
	trust-arch print-trust-arch untrust-arch \

	hosts \
	plaid-mock plaid-real \

prep:
	@mkdir -p config-override
	@touch config-override/mockgatehub.yaml
	@touch config-override/mockpti.yaml
	@touch config-override/mockxago.yaml
	@touch config-override/backend.yaml
	@touch config-override/backend.migration.yaml

infrastructure: prep
	$(COMPOSE) --profile infrastructure up

infra: infrastructure

services:
	$(COMPOSE) --profile services up

svc: services

application:
	$(COMPOSE) --profile application up --watch

app: application

backend:
	$(COMPOSE) --profile backend up --watch

back: backend

frontend:
	$(COMPOSE) --profile frontend up --watch

front: frontend

all: prep
	$(COMPOSE) --profile "*" up --watch

all-nowatch: prep
	$(COMPOSE) --profile "*" up -d

all-ci: prep
	$(COMPOSE) -f docker-compose.yaml -f protea.runner.mounts.yaml --profile infrastructure --profile services --profile application --profile backend --profile frontend up -d

# Prometheus + Grafana. Start the app services separately (e.g. `make all`) so
# Prometheus has targets to scrape. Runs in the foreground like `make infra`.
monitoring:
	$(MONITORING) --profile monitoring up

mon: monitoring

monitoring-down:
	$(MONITORING) --profile monitoring down

unit-test-db-up:
	$(COMPOSE) --profile unit-tests up -d postgres-unit-tests

unit-test-db-down:
	$(COMPOSE) stop postgres-unit-tests
	$(COMPOSE) rm -f postgres-unit-tests

build:
	$(COMPOSE) --profile "*" build

pull:
	$(COMPOSE) --profile "*" pull

down:
	$(MONITORING) --profile "*" down

delete-volumes:
	$(MONITORING) --profile "*" down -v

reset: delete-volumes

certs:
	mkdir -p config/certs
	@if [ "$(HEADLESS)" = "1" ]; then \
		openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
			-keyout config/certs/local.key -out config/certs/local.crt \
			-subj "/CN=interledger.test/O=Interledger/C=US" \
			-config config/san.cnf; \
	else \
		openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
			-keyout config/certs/local.key -out config/certs/local.crt \
			-config config/san.cnf; \
	fi

print-certs:
	openssl x509 -in config/certs/local.crt -text -noout

trust:
	sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain config/certs/local.crt

print-trust:
	sudo security find-certificate -a -c "interledger.test" -Z /Library/Keychains/System.keychain

untrust:
	sudo security find-certificate -a -c "interledger.test" -Z /Library/Keychains/System.keychain | \
	grep 'SHA-1 hash:' | \
	awk '{print $$NF}' | \
	xargs -I {} sudo security delete-certificate -Z {} /Library/Keychains/System.keychain

trust-debian:
	sudo cp config/certs/local.crt /usr/local/share/ca-certificates/
	sudo update-ca-certificates
	rm -rf $(HOME)/.pki/nssdb
	mkdir -p $(HOME)/.pki/nssdb
	certutil -d sql:$(HOME)/.pki/nssdb -N --empty-password
	certutil -d sql:$(HOME)/.pki/nssdb -A -t "C,," -n "interledger.test" -i config/certs/local.crt

print-trust-debian:
	certutil -d sql:$(HOME)/.pki/nssdb -L | grep "interledger.test"

untrust-debian:
	sudo rm /usr/local/share/ca-certificates/local.crt
	sudo update-ca-certificates
	certutil -d sql:$(HOME)/.pki/nssdb -D -n "interledger.test"

trust-arch:
	sudo cp config/certs/local.crt /etc/ca-certificates/trust-source/anchors/interledger-test.crt
	sudo update-ca-trust
	mkdir -p $(HOME)/.pki/nssdb
	[ -f "$(HOME)/.pki/nssdb/cert9.db" ] || certutil -d sql:$(HOME)/.pki/nssdb -N --empty-password
	certutil -d sql:$(HOME)/.pki/nssdb -A -t "C,," -n "interledger.test" -i config/certs/local.crt

print-trust-arch:
	trust list | grep -A5 "interledger"

untrust-arch:
	sudo rm -f /etc/ca-certificates/trust-source/anchors/interledger-test.crt
	sudo update-ca-trust
	certutil -d sql:$(HOME)/.pki/nssdb -D -n "interledger.test"

HOSTS := interledger.test \
	admin.mgnt.interledger.test \
	temporal.mgnt.interledger.test \
	local.ilp.link \
	auth.local.ilp.link \
	connector.local.ilp.link \
	rafiki.mgnt.interledger.test \
	traefik.test \
	ngrok.test \
	identity.interledger.test \
	mockgatehub.interledger.test \
	mockxago.interledger.test \
	mockpti.interledger.test \
	mockplaid.interledger.test \
	cdn.plaid.com \
	docs.interledger.test

hosts:
	@sudo sed -i.bak '/# generated by make hosts/d' /etc/hosts && sudo rm -f /etc/hosts.bak
	@$(foreach host,$(HOSTS),echo "127.0.0.1 $(host) # generated by make hosts" | sudo tee -a /etc/hosts > /dev/null;)

# Toggle local Plaid between the mock and the real Plaid. Flips both layers:
# backend SDK base URL (config-override/backend.yaml plaid.api_url) + the
# cdn.plaid.com /etc/hosts redirect (browser SDK). Restarts the backend so the
# new config takes effect. There must only ever be one top-level `plaid:` key
# in this file - YAML doesn't merge duplicate top-level keys (the last one
# wins), which wipes client_id/secret/env/etc and breaks against real Plaid.
PLAID_OVERRIDE := config-override/backend.yaml

# $(1) = api_url value to set
define set-plaid-api-url
	@mkdir -p config-override
	@touch $(PLAID_OVERRIDE)
	@if grep -q '  api_url:' $(PLAID_OVERRIDE); then \
		awk -v url='$(1)' '/^[[:space:]]*api_url:/ { print "  api_url: \"" url "\""; next } { print }' \
			$(PLAID_OVERRIDE) > $(PLAID_OVERRIDE).tmp && mv $(PLAID_OVERRIDE).tmp $(PLAID_OVERRIDE); \
	elif grep -q '^plaid:[[:space:]]*$$' $(PLAID_OVERRIDE); then \
		awk -v url='$(1)' '/^plaid:[[:space:]]*$$/ { print; print "  api_url: \"" url "\""; next } { print }' \
			$(PLAID_OVERRIDE) > $(PLAID_OVERRIDE).tmp && mv $(PLAID_OVERRIDE).tmp $(PLAID_OVERRIDE); \
	else \
		printf 'plaid:\n  api_url: "$(1)"\n' >> $(PLAID_OVERRIDE); \
	fi
endef

plaid-mock:
	$(call set-plaid-api-url,http://mockplaid:8080)
	@grep -q 'cdn.plaid.com' /etc/hosts || echo "127.0.0.1 cdn.plaid.com # generated by make hosts" | sudo tee -a /etc/hosts > /dev/null
	@echo "Plaid -> MOCK (mockplaid). Restarting backend..."
	$(COMPOSE) up -d backend

plaid-real:
	$(call set-plaid-api-url,)
	@sudo sed -i.bak '/cdn.plaid.com/d' /etc/hosts && sudo rm -f /etc/hosts.bak
	@echo "Plaid -> REAL (sandbox.plaid.com, selected by plaid.env). Ensure plaid.client_id/plaid.secret are real sandbox creds in $(PLAID_OVERRIDE). Restarting backend..."
	$(COMPOSE) up -d backend

