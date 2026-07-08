# Local Kubernetes Environment (kind)

This folder stands up the **full interledger-app stack** on a local
[kind](https://kind.sigs.k8s.io/) (Kubernetes in Docker) cluster. It deploys
Traefik (ingress + TLS), PostgreSQL, Valkey (Redis), Ory Kratos, Temporal, the
mock payment providers (mockpti, mockgatehub, mockxago), and the wallet
application itself (backend, protea frontend, botanist admin) — all reachable
from the browser over HTTPS at `*.interledger.kind`.

Everything lives in a single `interledger` namespace on a cluster named
`interledger`.

## Prerequisites

- **Docker** — builds and runs the container images and the kind node.
- **[kind](https://kind.sigs.k8s.io/docs/user/quick-start/)** — local Kubernetes.
  - macOS: `brew install kind`
  - Linux: `curl -Lo ./kind https://kind.sigs.k8s.io/kind-linux-amd64 && chmod +x ./kind && sudo mv ./kind /usr/local/bin/`
  - Verify: `kind version`
- **kubectl** — `kubectl version`
- **Helm 3+** — used to install Traefik, PostgreSQL, Valkey, Kratos, Temporal,
  and the local `helm/interledger-app` and `helm/mock-services` charts.
- **openssl** — generates the self-signed wildcard TLS certificate.
- **sudo** access — needed to add `*.interledger.kind` entries to `/etc/hosts`
  and (optionally) to trust the generated certificate.

## Getting Started

### Bring the whole stack up

```bash
cd k8s
make up
```

`make up` runs the full lifecycle in order:

1. `cluster` — creates the kind cluster `interledger` (control-plane only, with
   host ports 80/443 mapped into the node so Traefik is reachable from the browser).
2. `certs` — generates a self-signed wildcard cert for `*.interledger.kind`.
3. `certs-apply` — uploads it as the `kind-tls` Kubernetes TLS secret.
4. `hosts` — adds `*.interledger.kind` hostnames to `/etc/hosts` (requires sudo).
5. `build` — builds all Docker images: app (`backend`, `protea`, `botanist`)
   and mock services (`mockpti`, `mockgatehub`, `mockxago`).
6. `load` — loads those images into the kind cluster.
7. `deploy` — deploys everything (Traefik → manifests → infra → services → mock → app).
8. `validate` — port-forwards each service and checks its health endpoint.

Run `make help` to see every available target.

### Access the app

Once `make up` finishes, the stack is reachable at:

| URL | Service |
|---|---|
| `https://interledger.kind` | Wallet frontend (protea) |
| `https://interledger.kind/webhooks/*`, `/.well-known/*` | Wallet backend (HTTP) |
| `https://admin.mgnt.interledger.kind` | Admin UI (botanist) |
| `https://identity.interledger.kind` | Ory Kratos (public) |
| `https://mockpti.interledger.kind` | Mock PTI |
| `https://mockgatehub.interledger.kind` | Mock GateHub |
| `https://mockxago.interledger.kind` | Mock Xago |

The certificate is self-signed. Your browser will warn on first visit unless you
trust it — see [Trusting the TLS certificate](#trusting-the-tls-certificate).

### Tear down

```bash
make down
```

This deletes the entire kind cluster. To rebuild only part of the stack without
recreating the cluster, use the granular targets (e.g. `make build-app load-app
deploy-app`).

## Makefile Targets

| Target | Description |
|---|---|
| `up` | Full lifecycle: cluster, certs, hosts, build, load, deploy, validate |
| `down` | Delete the kind cluster |
| `cluster` | Create the kind cluster only |
| `certs` / `certs-apply` / `print-certs` | Generate / upload / inspect the TLS cert |
| `trust*` / `untrust*` | Trust/untrust the cert (macOS, Debian/Ubuntu, Arch — see below) |
| `hosts` | Add `*.interledger.kind` entries to `/etc/hosts` |
| `build` / `build-app` / `build-mock` | Build all / app-only / mock-only images |
| `load` / `load-app` / `load-mock` | Load all / app-only / mock-only images into kind |
| `manifests` | Apply the kustomize manifests (secrets, IngressRoutes, TLSStore, soketi) |
| `deploy` | Deploy everything in order |
| `deploy-traefik` | Traefik ingress controller (in `kube-system`) |
| `deploy-infra` | PostgreSQL + Valkey |
| `deploy-services` | Kratos + Temporal |
| `deploy-mock` | Mock services chart |
| `deploy-app` | Wallet app chart (backend, protea, botanist) |
| `validate` | Health-check each service via port-forward |
| `status` | `kubectl get pods -n interledger` |

## TLS

### Generating certs

`make certs` writes a self-signed wildcard certificate for `interledger.kind`
and `*.interledger.kind` to `config/certs/kind.{crt,key}`, using the SAN config
in [config/san.cnf](config/san.cnf). `make certs-apply` then uploads it as the
`kind-tls` secret, which the Traefik `TLSStore` (see
[manifests/tls-store.yaml](manifests/tls-store.yaml)) uses as the default
certificate for all HTTPS routes.

In headless/CI environments set `HEADLESS=1` so `openssl` runs non-interactively:

```bash
make certs HEADLESS=1
```

### Trusting the TLS certificate

To avoid browser warnings, trust the generated cert for your OS:

- **macOS**: `make trust` (remove with `make untrust`)
- **Debian/Ubuntu**: `make trust-debian` (remove with `make untrust-debian`)
- **Arch Linux**: `make trust-arch` (remove with `make untrust-arch`)

Use the matching `print-trust*` target to verify it is trusted.

## Configuration Strategy: configa

The app and mock services resolve their configuration at startup from Kubernetes
Secrets using the [configa](https://github.com/interledger/configa) library —
no secrets are baked into images or Helm values.

### How it works

1. **Config templates**: Service YAML configs contain templates like
   `{{ secret "wallet-backend" "dbUrl" }}` or `{{ secret "mock-secrets" "redisUrl" }}`.
2. **Secret injection**: The Kubernetes Secrets `wallet-backend`, `wallet-frontend`
   ([manifests/app-secrets.yaml](manifests/app-secrets.yaml)) and `mock-secrets`
   ([manifests/mock-secrets.yaml](manifests/mock-secrets.yaml)) hold the plaintext
   values.
3. **Runtime resolution**: On startup, configa's in-cluster secret client
   authenticates to the Kubernetes API (via a ServiceAccount granted get access
   to the named secrets) and fetches them.
4. **Template substitution**: Templates are replaced with the real values before
   the config is loaded.

> The secret values in this folder are **non-sensitive placeholders for local
> development only**. Never reuse them in a real environment.

### Example

In [kind/values.mock-services.yaml](kind/values.mock-services.yaml):

```yaml
mockpti:
  config:
    redis_url: '{{ secret "mock-secrets" "redisUrl" }}'
```

In [manifests/mock-secrets.yaml](manifests/mock-secrets.yaml):

```yaml
stringData:
  redisUrl: redis://valkey-primary.interledger:6379
```

At startup `redis_url` resolves to `redis://valkey-primary.interledger:6379`.

## Validation

After deployment, verify health:

```bash
make validate
# or directly:
bash ./scripts/validate.sh interledger kind-interledger
```

The script waits for each deployment to become available, then port-forwards each
service and checks its health endpoint (`/health` for mock services, `/healthz`
for the wallet backend/frontend/admin). It exits non-zero if any check fails.

Check pod status at any time with `make status`.

## Folder Structure

```
k8s/
├── Makefile                         # Cluster lifecycle + build/deploy/validate targets
├── config/
│   ├── san.cnf                      # openssl SAN config for the wildcard cert
│   └── certs/                       # Generated kind.crt / kind.key (git-ignored)
├── kind/
│   ├── cluster.yaml                 # kind cluster config (control-plane, host ports 80/443)
│   ├── values.traefik.yaml          # Traefik ingress controller overrides
│   ├── values.postgres.yaml         # PostgreSQL (Bitnami chart) overrides
│   ├── values.valkey.yaml           # Valkey/Redis (Bitnami chart) overrides
│   ├── values.kratos.yaml           # Ory Kratos overrides
│   ├── values.temporal.yaml         # Temporal overrides
│   ├── values.mock-services.yaml    # Mock services chart overrides (configa templates)
│   └── values.interledger-app.yaml  # Wallet app chart overrides (configa templates)
├── manifests/
│   ├── kustomization.yaml           # Aggregates the manifests below
│   ├── namespaces.yaml              # interledger namespace
│   ├── mock-secrets.yaml            # Secret for the mock services
│   ├── app-secrets.yaml             # ServiceAccount + Secrets for the wallet app
│   ├── soketi.yaml                  # Soketi (Pusher-compatible websockets) deployment
│   ├── tls-store.yaml               # Traefik default TLS certificate
│   └── ingressroutes/               # Traefik IngressRoutes
│       ├── frontend.yaml            # interledger.kind → protea (+ backend webhooks)
│       ├── admin.yaml               # admin.mgnt.interledger.kind → botanist
│       ├── kratos.yaml              # identity.interledger.kind → kratos-public
│       └── mock-services.yaml       # mock{pti,gatehub,xago}.interledger.kind
├── scripts/
│   └── validate.sh                  # Health-check validation
└── README.md                        # This file
```

The Helm charts themselves live outside this folder, at `helm/interledger-app`
and `helm/mock-services` in the repo root.

## Troubleshooting

**Pods stuck in `CrashLoopBackOff` / `Pending`** — inspect logs and events:

```bash
kubectl --context kind-interledger -n interledger get pods
kubectl --context kind-interledger -n interledger logs <pod-name>
kubectl --context kind-interledger -n interledger describe pod <pod-name>
```

**configa secret resolution failed** — confirm the secret exists and the pod's
ServiceAccount can read it:

```bash
kubectl --context kind-interledger -n interledger get secret wallet-backend mock-secrets
```

**Browser can't reach `*.interledger.kind`** — ensure `make hosts` ran (the
entries are tagged `# generated by k8s make hosts` in `/etc/hosts`) and that
Traefik is running in `kube-system`:

```bash
kubectl --context kind-interledger -n kube-system get pods -l app.kubernetes.io/name=traefik
```

**TLS warnings in the browser** — trust the cert with the `make trust*` target
for your OS (see [Trusting the TLS certificate](#trusting-the-tls-certificate)).

**Helm deployment timeout** — images may still be pulling, or a dependency isn't
ready. Watch the pods and re-run the relevant `deploy-*` target once they settle:

```bash
kubectl --context kind-interledger -n interledger get pods -w
```

**Ports 80/443 already in use** — another process (or a previous kind cluster)
is bound to them. Free the ports or `make down` any stale cluster before `make up`.
```
