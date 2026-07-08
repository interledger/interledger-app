# Environment Variables Reference

> **Configuration reference.** Central source for runtime environment variables, secret classification, and per-environment values.

> **Note:** Values shown as `<...>` are placeholders, not required for local development.

**Related documents:**

- [Home](index.md) - Full documentation map and topic index
- [Backend Configuration Guide](backend-configuration-guide.md) - The Go backend's YAML config scheme, secret handling, and full settings reference
- [Frontend Configuration Guide](frontend-configuration-guide.md) - Protea's and Botanist's YAML config scheme, secret handling, and full settings reference
- [Logging Reference](logging-reference.md) - Log configuration and structured logging policy
- [Signup Explainer](signup-guide.md) - Signup flow and auth-related runtime dependencies
- [KYC Explainer](kyc-guide.md) - Provider verification flows and integration dependencies
- [Provider Payments Guide](provider-payments-reference.md) - Provider behavior and configuration differences
- [GateHub Cards](gatehub-cards-guide.md) - GateHub integration configuration and card-specific secrets

**Quick Navigation:**

- **Looking for Protea or Botanist configuration?** -> Both are configured via YAML, not environment variables — see the [Frontend Configuration Guide](frontend-configuration-guide.md)
- **Looking for backend configuration?** -> The backend is configured via YAML, not environment variables — see the [Backend Configuration Guide](backend-configuration-guide.md)

This document previously covered per-key environment variables for the three application services:

- **Backend** — wallet backend (`go/backend`)
- **Protea** — user-facing frontend (`typescript/protea`)
- **Botanist** — admin portal (`typescript/botanist`)

> **None of the three are configured through per-key environment variables anymore.** Each loads YAML config files via the `CONFIG` variable and resolves secrets through Kubernetes secret templating (`{{ secret "name" "key" }}`), using the Go `configa` package (backend) or its TypeScript port `@interledger/configa` (Protea, Botanist). See the [Backend Configuration Guide](backend-configuration-guide.md) and [Frontend Configuration Guide](frontend-configuration-guide.md) for the full configuration schemes and settings references.

Variables with **Secret: Yes** must be stored in 1Password and injected via a Kubernetes secret (referenced from YAML via a `{{ secret ... }}` template); real production secret values must never be committed to repository config files. Variables with **Secret: No** are safe to store in a ConfigMap or compose file.

---

## Protea (Frontend)

Protea is a React Router application serving the user-facing wallet UI. It is configured via YAML — see the [Frontend Configuration Guide: Protea reference](frontend-configuration-guide.md#protea-reference) for the full settings table. Its only genuine environment variables:

| Variable | Required | Purpose |
|---|---|---|
| `CONFIG` | **Yes** | Comma-separated list of YAML config file paths to merge (base first, overrides last). |
| `NODE_ENV` | No | `production` or `development`. |
| `SENTRY_RELEASE` | No | Tied to the built image (same value as `IMAGE_HASH`), not the environment — stays a build-time env var. |
| `INTERLEDGER_APP_VERSION` | No | Surfaced on `/healthz`. Also image-build-time metadata. |
| `KUBERNETES_SERVICE_HOST` / `KUBERNETES_SERVICE_PORT` | Auto | Injected by Kubernetes; used to resolve `{{ secret }}` templates. Not set manually. |

---

## Botanist (Admin Portal)

Botanist is a React Router application providing the internal admin interface. It connects to the backend admin gRPC port, which must only be accessible from within the cluster. It is configured via YAML — see the [Frontend Configuration Guide: Botanist reference](frontend-configuration-guide.md#botanist-reference) for the full settings table (currently just two keys — Botanist has no secret configuration). Its only genuine environment variables:

| Variable | Required | Purpose |
|---|---|---|
| `CONFIG` | **Yes** | Comma-separated list of YAML config file paths to merge (base first, overrides last). |
| `NODE_ENV` | No | `production` or `development`. |

---

## Backend (Wallet)

The Go backend is the core of the wallet, handling payments, provider integrations, webhooks, Temporal workflows, and gRPC APIs.

**The backend is not configured through environment variables.** It loads YAML configuration files listed in the `CONFIG` environment variable, deep-merges them (base + override), and resolves secrets via Kubernetes secret templating (`{{ secret "name" "key" }}`). Configuration is validated against a typed struct at startup.

Its only genuine environment variables are:

| Variable | Required | Purpose |
|---|---|---|
| `CONFIG` | **Yes** | Comma-separated list of YAML config file paths to merge (base first, overrides last). |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | No | OTLP/gRPC trace endpoint. Overrides the YAML `otel.endpoint` value when set (warning logged). |
| `OTEL_EXPORTER_OTLP_HEADERS` | No | Auth headers for the OTLP endpoint. Overrides the YAML `otel.headers` value when set (warning logged). |

For the full configuration scheme, secret handling, per-setting reference (including all provider, database, and feature-flag settings), and the migration config, see the **[Backend Configuration Guide](backend-configuration-guide.md)**.

> **Legacy note:** Earlier revisions of this document listed the backend's settings as individual `BACKEND_*` environment variables. Those are obsolete — the backend now reads them from YAML config. The per-environment values previously tracked here now live in the deployed environments' config files.
