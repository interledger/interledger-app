# Frontend Configuration Guide

> **Configuration reference.** How Protea and Botanist are configured: the YAML config scheme, how secrets are resolved, and a complete table of every available setting.

**Related documents:**

- [Home](index.md) - Full documentation map and topic index
- [Environment Variables](env-variables.md) - The small set of genuine runtime environment variables that remain
- [Backend Configuration Guide](backend-configuration-guide.md) - The Go backend's equivalent YAML config scheme
- [Logging Reference](logging-reference.md) - Log levels and structured logging policy

**Quick Navigation:**

- **How is Protea configured?** -> See [Protea reference](#protea-reference)
- **How is Botanist configured?** -> See [Botanist reference](#botanist-reference)
- **How do I set a value locally?** -> See [Local development](#local-development)
- **How are secrets handled?** -> See [Secrets](#secrets)

---

## Overview

Protea and Botanist are configured from **YAML files**, the same scheme the Go backend uses. At startup each app reads the `CONFIG` environment variable — a comma-separated list of YAML file paths — merges those files, resolves any secret templates, validates the result against a typed [Zod](https://zod.dev) schema, and then runs. The loader is [`@interledger/configa`](../../typescript/packages/configa), a TypeScript port of the backend's `configa` package — same merge rules, same `{{ secret "name" "key" }}` template syntax, same YAML files are readable by both the Go and TypeScript services.

Only a small number of genuine environment variables remain per app (see [Environment Variables](env-variables.md)); everything else lives in YAML.

---

## Load flow

1. **Read `CONFIG`** — a comma-separated list of YAML file paths. Required; startup fails if unset.
2. **Parse & merge** — files are parsed in order and deep-merged, with **later files overriding earlier ones** (maps deep-merge; scalars/arrays overlay-wins).
3. **Resolve secrets** — any `{{ secret "name" "key" }}` template expressions are substituted with values fetched from the Kubernetes Secret API via the pod's service account. If no templates are present, no Kubernetes call is made.
4. **Validate** — the merged config is parsed against the app's Zod schema (`app/config.schema.mjs`). Missing required fields or wrong types cause an immediate, descriptive startup failure.

**Protea** resolves this once, in `server.js`, before the app's route module graph is even imported (this ordering matters — see the comments in `typescript/protea/app/config.server.ts` for why a plain top-level `await` doesn't work here). **Botanist** has no custom server entry point, so `typescript/botanist/bootstrap.mjs` resolves the config and re-exposes the two resulting values as env vars before exec'ing the real `react-router dev` / `react-router-serve` command.

### Base + override overlay

Both apps mount `CONFIG` as two files, a base config and an override, exactly like the backend:

```
CONFIG=/etc/protea/config.yaml,/etc/protea/config-override.yaml
```

The override file may be empty.

---

## Local development

Local config lives in `local/config/` and `local/config-override/`, mirroring the backend's pattern:

| File | Role |
|---|---|
| `local/config/protea.yaml` | Base Protea config for local development (checked in, plaintext test values). |
| `local/config-override/protea.yaml` | Local overrides. **Git-ignored** — put your personal or real local secret values here. |
| `local/config/botanist.yaml` | Base Botanist config for local development. |
| `local/config-override/botanist.yaml` | Local overrides (git-ignored). |

To change a setting locally, edit the `config-override/*.yaml` file (which wins over the base file) rather than the checked-in base. The override files are created empty by `make prep`. No `{{ secret }}` templates are needed locally, so `configa` never calls the Kubernetes API in this environment.

Both apps keep `target: dev` and bind-mount `app/` into the container for live reload — mounting config from a file doesn't affect that.

---

## Secrets

Secret resolution works identically to the backend — see the [Backend Configuration Guide's Secrets section](backend-configuration-guide.md#secrets) for the full explanation. In short:

```yaml
rafiki:
  auth_secret: '{{ secret "wallet-frontend" "rafikiAuthSecret" }}'
```

- If **no** templates are present after merging, no Kubernetes call is made.
- A missing secret, a missing key within a secret, or insufficient RBAC permissions all cause a descriptive startup failure.
- The Helm chart grants each app's service account read-only access to the secrets listed in `frontend.secretAccess` / `admin.secretAccess` via a Role + RoleBinding (`templates/rbac.frontend.yaml`, `templates/rbac.admin.yaml`), mirroring `templates/rbac.backend.yaml`.

**One TypeScript-specific quirk:** `cookie_secrets` is a `string[]`, but a single Kubernetes Secret key can only ever hold a string. In-cluster, this field is set to a `{{ secret ... }}` template that resolves to a JSON-encoded array (e.g. `'["real-secret-value"]'`), and Protea's Zod schema (`config.schema.mjs`) JSON-decodes it automatically. Locally, it's just a native YAML array.

---

## Protea reference

Loaded from `typescript/protea/app/config.schema.mjs` (`ProteaConfigSchema`).

### Environment variables

These are the only environment variables Protea itself reads — everything else is YAML config.

| Variable | Required | Purpose |
|---|---|---|
| `CONFIG` | **Yes** | Comma-separated list of YAML config file paths to merge (base first, overrides last). |
| `NODE_ENV` | No | `production` or `development`. Controls the request pipeline and log formatting; not moved to YAML since it's a Node/React Router intrinsic. |
| `SENTRY_RELEASE` | No | Same value as the deploy `IMAGE_HASH` — tied to the built image, not the environment, so it stays a build-time env var rather than a YAML setting. |
| `INTERLEDGER_APP_VERSION` | No | Surfaced on `/healthz` as the `interledger-app-version` header. Also image-build-time metadata. |
| `KUBERNETES_SERVICE_HOST` / `KUBERNETES_SERVICE_PORT` | Auto | Injected by Kubernetes. Used by `configa` to reach the Kubernetes Secret API when resolving `{{ secret }}` templates. Not set manually. |

### Configuration reference

| Key | Type | Required | Secret | Notes |
|---|---|---|---|---|
| `environment` | string | No | No | Runtime environment tag surfaced to the client. Prod: `prod`; Sandbox/Dev: `dev`; Local: `local`. |
| `cookie_secrets` | []string | Yes | Yes | Used to sign session cookies. See the JSON-encoding note above. |
| `target_host` | string | Yes | No | Base URL of the app itself. Used to build self-referential links in legal and contact pages. |
| `support_email` | string | Yes | No | Support email shown in error messages, legal pages, and notifications. |
| `payment_pointer_base` | string | Yes | No | Domain used to build Open Payments payment pointer addresses. |
| `public_op_auth_host` | string | Yes | No | Public-facing hostname for the Open Payments auth server. |
| `persona_sdk_url` | string | Yes | No | URL to the Persona identity verification SDK. |
| `mockxago_endpoint` | string | No | No | Base URL for the Xago/Persona KYC iframe. Empty falls back to the Persona SDK flow. Only set locally. |
| `log_level` | string | No | No | `debug`/`info`/`warn`/`error`. |
| `log_pretty` | bool | No | No | Human-readable log formatting. Default `true`. |
| `google_maps_api_key` | string | No | Yes | Used by geocoding/places-autocomplete endpoints during onboarding. |
| `backend.grpc_url` | string | No | No | Internal wallet backend gRPC URL. Defaults to `http://<fullname>-backend-service-grpc:8443` in Helm when unset. |
| `backend.http_url` | string | No | No | Internal wallet backend HTTP URL. Defaults to `http://<fullname>-backend-service-http:8080` in Helm when unset. |
| `redis.url` | string | Yes | Yes | Redis connection URL for session storage and caching. |
| `kratos.url` | string | Yes | No | Ory Kratos public API URL. |
| `rafiki.auth_endpoint` | string | Yes | No | Rafiki auth gRPC/GraphQL endpoint. |
| `rafiki.auth_secret` | string | Yes | Yes | Shared secret between Protea and the Rafiki auth service. |
| `pti.sdk_url` | string | Yes | No | Fiant Web SDK JavaScript bundle URL. |
| `pti.forms_url` | string | Yes | No | Fiant hosted forms (Elements) URL. |
| `pti.client_id` | string | Yes | No | PTI/Fiant payment provider client UUID. |
| `sentry.dsn` | string | No | Yes | Sentry DSN. Sentry is disabled when empty (default in local). |
| `sentry.env_label` | string | No | No | Environment label sent to Sentry. Only relevant when `sentry.dsn` is set. |
| `pusher.app_key` | string | No | Yes | Pusher application key for real-time push notifications. Realtime degrades gracefully when unset. |
| `pusher.app_cluster` | string | No | No | Pusher cluster region. Default `eu`. |
| `rate_limit.requests` | number | No | No | Max requests per time window before rate limiting. Default `4`. |
| `rate_limit.window_seconds` | number | No | No | Rate limit time window in seconds. Default `3600`. |
| `op_intpay.enabled` | bool | No | No | Feature flag for Open Payments Interledger Pay (`/quick-pay` routes). Default `false`. |
| `op_intpay.host` | string | Conditional | No | Public base URL used to build quick-pay links. Required when `op_intpay.enabled` is true. |
| `op_intpay.redirect_url` | string | Conditional | No | URL the client is redirected to once a quick-pay grant completes. Required when enabled. |
| `op_intpay.wallet_address` | string | Conditional | No | Wallet address used as the Open Payments client identity. Required when enabled. |
| `op_intpay.key_id` | string | Conditional | Yes | Key id of the Open Payments client. Required when enabled. |
| `op_intpay.private_key` | string | Conditional | Yes | Base64-encoded Ed25519 private key, paired with `op_intpay.key_id`. Required when enabled. |

> **Removed, not carried over:** `SEGMENT_API_KEY` and the `OTEL_EXPORTER_OTLP_*` frontend variables were set by the old env-var ConfigMap but never actually read by any Protea code — they were dropped rather than migrated. `BT_TOKEN`, `CF_TURNSTILE_SITE_KEY`, `CF_TURNSTILE_SECRET_KEY`, and `GATE_SIGNUP` were already-dead legacy variables and were also dropped.

---

## Botanist reference

Loaded from `typescript/botanist/app/config.schema.mjs` (`BotanistConfigSchema`). Botanist has a much smaller config surface than Protea — it only proxies gRPC calls to the backend admin API and displays payment pointer addresses.

### Environment variables

| Variable | Required | Purpose |
|---|---|---|
| `CONFIG` | **Yes** | Comma-separated list of YAML config file paths to merge (base first, overrides last). |
| `NODE_ENV` | No | `production` or `development`. |

### Configuration reference

| Key | Type | Required | Secret | Notes |
|---|---|---|---|---|
| `backend_grpc_url` | string | No | No | Internal backend admin gRPC port (8448) target (host:port, no scheme). Defaults to `<fullname>-backend-service-grpc:8448` in Helm when unset. |
| `payment_pointer_base` | string | Yes | No | Domain used to display payment pointer addresses for users. |

> **Removed, not carried over:** `KRATOS_ADMIN_URL` was set by the old env-var ConfigMap but never actually read by any Botanist code — it was dropped rather than migrated. Botanist has no secret configuration today; all sensitive operations are delegated to the backend.

---

> **Legacy note:** Earlier revisions of [Environment Variables](env-variables.md) listed Protea's and Botanist's settings as individual environment variables (`ILW_ENV`, `BACKEND_GRPC_URL`, `COOKIE_SECRETS`, etc.), injected via a per-key ConfigMap and `secretKeyRef` env vars. Those are obsolete — both apps now read them from YAML config via `@interledger/configa`, the same scheme the backend already used. The per-environment values previously tracked as env vars now live in the deployed environments' config files.
