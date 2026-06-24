# k8s CI Testing

This folder contains scripts and configuration for local Kubernetes testing using [kind](https://kind.sigs.k8s.io/) (Kubernetes in Docker). It orchestrates the deployment of Valkey (Redis), and mock payment provider services (mockpti, mockgatehub, mockxago) in an isolated cluster.

## Prerequisites

### Installing kind

kind is a tool for running local Kubernetes clusters using Docker containers.

- **Official installation guide**: https://kind.sigs.k8s.io/docs/user/quick-start/
- **Quick install** (macOS with Homebrew): `brew install kind`
- **Quick install** (Linux): `curl -Lo ./kind https://kind.sigs.k8s.io/kind-linux-amd64 && chmod +x ./kind && sudo mv ./kind /usr/local/bin/`

Verify installation: `kind version`

### Required tools

- Docker (for building and running container images)
- kubectl (installed automatically with kind; verify with `kubectl version`)
- Helm 3+ (for deploying charts)

## Getting Started

### Start the cluster

```bash
cd k8s
make up
```

This command:
1. Creates a local kind cluster named `interledger-ci`
2. Builds Docker images for all three mock services (mockpti, mockgatehub, mockxago)
3. Loads images into the cluster
4. Deploys Valkey (Redis) using Helm
5. Applies Kubernetes manifests (including configuration secrets via Kustomize)
6. Deploys mock services using Helm with configa integration

Expected result: All pods in the `mock-services` namespace should be `1/1 Running`.

### Stop and clean up

```bash
make down
```

This command:
1. Deletes the `mock-services` namespace
2. Destroys the entire kind cluster

To preserve the cluster but only remove the namespace (for faster redeploys):

```bash
kubectl --context kind-interledger-ci delete namespace mock-services --ignore-not-found
helm --kube-context kind-interledger-ci uninstall valkey mock-services --namespace mock-services --ignore-not-found
```

## Configuration Strategy: configa

All mock services use the [configa](https://github.com/interledger/configa) library to resolve Kubernetes Secrets into configuration values.

### How it works

1. **Config Templates**: Service YAML configs contain templates like `{{ secret "mock-secrets" "redisUrl" }}`
2. **Secret Injection**: The Kubernetes Secret `mock-secrets` (defined in `manifests/mock-secrets.yaml`) stores plaintext config values
3. **Runtime Resolution**: When pods start, configa's `InClusterSecretClient` authenticates to the Kubernetes API and fetches the secret
4. **Template Substitution**: Templates are replaced with actual values before config is loaded

### Benefits

- **No hardcoded secrets** in values files or environment variables
- **Kubernetes-native**: Uses standard Kubernetes Secrets API
- **RBAC-protected**: ServiceAccounts have minimal permissions (get-only on specific secrets)
- **Dynamic**: Update secrets without redeploying pods

### Example

In [kind/values.mock-services.yaml](kind/values.mock-services.yaml):

```yaml
mockpti:
  config:
    redis_url: "{{ secret \"mock-secrets\" \"redisUrl\" }}"
```

In [manifests/mock-secrets.yaml](manifests/mock-secrets.yaml):

```yaml
stringData:
  redisUrl: redis://valkey-primary:6379
```

Result: At startup, `redis_url` is resolved to `redis://valkey-primary:6379`.

## Validation

After deployment, verify health:

```bash
bash ./scripts/validate.sh
```

This script checks that all deployments are ready and all services respond to `/health` endpoint.

## Folder Structure

```
k8s/
├── Makefile                    # Cluster lifecycle commands (up, down, build)
├── kind/
│   ├── cluster.yaml            # kind cluster configuration (1 control plane, no workers)
│   ├── values.valkey.yaml      # Helm overrides for Valkey deployment
│   └── values.mock-services.yaml  # Helm overrides for mock services (includes configa templates)
├── manifests/
│   ├── kustomization.yaml      # Kustomize configuration (namespace, resources)
│   └── mock-secrets.yaml       # Kubernetes Secret with configuration values
├── scripts/
│   ├── validate.sh             # Health check validation
│   └── (future test scripts)
└── README.md                   # This file
```

## Troubleshooting

**Pods stuck in CrashLoopBackOff**: Check pod logs:
```bash
kubectl --context kind-interledger-ci logs -n mock-services <pod-name>
```

**configa secret resolution failed**: Verify the Secret exists:
```bash
kubectl --context kind-interledger-ci get secret -n mock-services mock-secrets
```

**Helm deployment timeout**: Wait longer for images to pull:
```bash
kubectl --context kind-interledger-ci get pods -n mock-services
```

**Services not responding to health checks**: Ensure all pods are 1/1 Running before running validation script.