# Backend Configuration Guide

> **Configuration reference.** How the Go backend is configured: the YAML config scheme, how secrets are resolved, and a complete table of every available setting.

**Related documents:**

- [Home](index.md) - Full documentation map and topic index
- [Environment Variables](env-variables.md) - Runtime environment variables for Protea and Botanist (the frontends)
- [Logging Reference](logging-reference.md) - Log levels and structured logging policy

**Quick Navigation:**

- **How is the backend configured?** -> See [Configuration scheme](#configuration-scheme)
- **How do I set a value locally?** -> See [Local development](#local-development)
- **How are secrets handled?** -> See [Secrets](#secrets)
- **What settings exist?** -> See [Configuration reference](#configuration-reference-startconfig) and [Migration configuration](#migration-configuration-migrationconfig)

---

## Overview

The Go backend (`go/backend`) is configured entirely from **YAML files**, not environment variables. At startup it reads the `CONFIG` environment variable — a comma-separated list of YAML file paths — merges those files, resolves any secret templates, validates the result against a typed struct, and then runs. The loader is the internal [`configa`](https://github.com/interledger/interledger-app/tree/main/go/configa) package.

Only a small number of genuine environment variables remain (see [Environment variables](#environment-variables) below); everything else lives in YAML.

> This applies to the backend `start`, `worker`, and `migrate` commands. The Pacioli ledger service and the mock provider services have their own, separate configuration and are not covered here.

---

## Configuration scheme

### Load flow

At startup the backend performs the following (`go/backend/config`):

1. **Read `CONFIG`** — a comma-separated list of YAML file paths. Required; startup fails if unset.
2. **Parse & merge** — files are parsed in order and deep-merged, with **later files overriding earlier ones**. This is the base + override pattern (see below).
3. **Resolve secrets** — any `{{ secret "name" "key" }}` template expressions are substituted with values fetched from the Kubernetes Secret API. If no templates are present, no Kubernetes call is made.
4. **Apply defaults** — a few fields default when omitted (see the [reference table](#configuration-reference-startconfig)).
5. **Validate** — the merged config is unmarshalled into the `StartConfig` struct and validated. Missing required fields, invalid enum values, and unmet conditional requirements cause an immediate, descriptive startup failure.

Because validation runs at startup, a misconfigured backend fails fast rather than erroring later at runtime.

### Base + override overlay

`CONFIG` typically lists two files, a base config and an override:

```
CONFIG=/etc/backend/config.yaml,/etc/backend/config-override.yaml
```

Merge behavior:

| Value kind | Behavior |
|---|---|
| Maps / nested objects | Deep-merged recursively |
| Scalars (strings, numbers, bools) | Override replaces base |
| Arrays | Override replaces base (not concatenated) |

This lets a base file carry shared defaults while an override file supplies environment-specific or secret values. The override file may be empty.

---

## Environment variables

These are the **only** environment variables the backend itself reads. Everything else is YAML config.

| Variable | Required | Purpose |
|---|---|---|
| `CONFIG` | **Yes** | Comma-separated list of YAML config file paths to merge (base first, overrides last). |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | No | OTLP/gRPC trace endpoint. Read by the OpenTelemetry SDK. **Takes priority over** the YAML `otel.endpoint` value (a warning is logged when it overrides config). |
| `OTEL_EXPORTER_OTLP_HEADERS` | No | Auth headers for the OTLP endpoint (e.g. the Honeycomb ingest key). **Takes priority over** the YAML `otel.headers` value (a warning is logged when it overrides config). |
| `KUBERNETES_SERVICE_HOST` / `KUBERNETES_SERVICE_PORT` | Auto | Injected by Kubernetes. Used by `configa` to reach the Kubernetes Secret API when resolving `{{ secret }}` templates. Not set manually. |

> The `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` and `OTEL_EXPORTER_OTLP_TRACES_HEADERS` traces-specific variants are also honored, matching the OpenTelemetry SDK convention.

See [`otel`](#observability) in the reference for the YAML equivalents.

---

## Local development

Local config lives in `local/config/` and `local/config-override/`:

| File | Role |
|---|---|
| `local/config/backend.yaml` | Base backend config for local development (checked in, plaintext test values). |
| `local/config-override/backend.yaml` | Local overrides. **Git-ignored** — put your personal or real local secret values here. |
| `local/config/backend.migration.yaml` | Base config for the `migrate` command. |
| `local/config-override/backend.migration.yaml` | Overrides for the `migrate` command (git-ignored). |

The local compose file mounts these into the container and sets `CONFIG` accordingly:

```yaml
environment:
  - CONFIG=/etc/backend/config.yaml,/etc/backend/config-override.yaml
volumes:
  - ./config/backend.yaml:/etc/backend/config.yaml:ro
  - ./config-override/backend.yaml:/etc/backend/config-override.yaml:ro
```

To change a setting locally, edit `config-override/backend.yaml` (which wins over the base file) rather than the checked-in base. The override file is created empty by `make prep`.

---

## Secrets

### How secret resolution works

Any string value in the YAML may contain a template expression:

```yaml
gatehub:
  webhook_secret: '{{ secret "wallet-backend" "gatehubWebhookSecret" }}'

db:
  url: 'postgres://user:{{ secret "wallet-backend" "dbPassword" }}@db:5432/backend'
```

At startup `configa` scans the merged YAML for `{{ secret "<name>" "<key>" }}` expressions and replaces each with the value of key `<key>` in the Kubernetes Secret named `<name>` (in the pod's namespace). Templates may appear anywhere in a string, including embedded in a larger value like a connection URL.

Key points:

- If **no** templates are present after merging, no Kubernetes call is made — plaintext-only configs work with no cluster access. This is how local development runs.
- If templates **are** present but the loader has no secret client configured, startup fails with `configa.ErrNoSecretClient`.
- A missing secret, a missing key within a secret, or insufficient RBAC permissions all cause a descriptive startup failure (`ErrSecretNotFound`, `ErrSecretForbidden`, `ErrSecretFetchFailed`).

### Rules for handling secrets

- **Never commit real secret values.** Real credentials must never appear in checked-in config files. Deployed configs reference them via `{{ secret ... }}` templates; the actual values live in the team secret manager and are synced into Kubernetes Secrets.
- **Local test values are fine to commit.** The plaintext values in `local/config/backend.yaml` are throwaway test credentials for the mock services — they are not real secrets.
- **Put real local values in the override file.** If you need a real credential locally (e.g. to hit a real provider sandbox), place it in the git-ignored `local/config-override/backend.yaml`, never in the base file.
- **Which fields are secret** is marked in the [reference table](#configuration-reference-startconfig) below. Any field marked *Secret: Yes* should be supplied via a `{{ secret }}` template in deployed environments.

---

## Configuration reference (`StartConfig`)

Loaded by the backend `start` and `worker` commands. **Required** means the field must be present (validation fails otherwise). **Secret** marks fields that should be injected via `{{ secret }}` templates in deployed environments.

### General

| Key | Type | Required | Secret | Notes |
|---|---|---|---|---|
| `environment.mode` | string | Yes | No | One of `prod`, `sandbox`, `dev`, `local`, `test`. Drives behavioural switches. |
| `environment.label` | string | No | No | Human-readable tag attached to telemetry (Sentry, OTEL). |
| `port` | string | No | No | HTTP port. Default `8080`. |
| `application_url` | string | Yes | No | Base URL of the wallet application. |
| `open_payments_base_url` | string | Yes | No | Open Payments base URL. |
| `auth_base_url` | string | Yes | No | Open Payments auth server base URL. |
| `log_level` | string | No | No | `debug`/`info`/`warn`/`error`. Default `info`. |
| `log_output_path` | string | No | No | Log destination. Default `stdout`. |
| `allowed_wallet_ids` | []string | No | No | Wallet IDs allowed through regional blocks. |
| `blocked_regions` | []string | No | No | ISO country codes to block. Prod: `US`. |

### Database

| Key | Type | Required | Secret | Notes |
|---|---|---|---|---|
| `db.url` | string | Yes | Yes | Wallet backend PostgreSQL URL (contains credentials). |
| `db.pacioli_url` | string | Yes | Yes | Pacioli ledger PostgreSQL URL. |

### Kratos & Temporal

| Key | Type | Required | Secret | Notes |
|---|---|---|---|---|
| `kratos.url` | string | Yes | No | Kratos public API URL. Defaults to `http://localhost:4433` if omitted. |
| `kratos.admin_url` | string | Yes | No | Kratos admin API URL. |
| `temporal.url` | string | No | No | Temporal frontend address. Default `temporal:7233`. |

### Rafiki

| Key | Type | Required | Secret | Notes |
|---|---|---|---|---|
| `rafiki.node_enabled` | bool | No | No | Feature flag: enables Rafiki full-node event orchestration flows. |
| `rafiki.backend_graphql_url` | string | Yes | No | Rafiki backend GraphQL endpoint. |
| `rafiki.auth_graphql_url` | string | Yes | No | Rafiki auth GraphQL endpoint. |
| `rafiki.operator_tenant_id` | string | Yes | No | Operator tenant UUID. |
| `rafiki.admin_api_secret` | string | Yes | Yes | Shared secret for the Rafiki admin API. |
| `rafiki.signature_version` | string | Yes | No | Request signature version. |
| `rafiki.db_url` | string | No | Yes | Rafiki backend DB URL (used by worker jobs). |
| `rafiki.auth_db_url` | string | No | Yes | Rafiki auth DB URL (used by worker jobs). |

### GateHub

Required in **production** (validated when `environment.mode` is `prod`). See conditional requirements below.

| Key | Type | Required | Secret | Notes |
|---|---|---|---|---|
| `gatehub.app_id` | string | Prod | Yes | API app ID header. |
| `gatehub.secret` | string | Prod | Yes | API signing secret. |
| `gatehub.card_app_id` | string | Prod | Yes | Card app ID header. |
| `gatehub.gateway_id` | string | Prod | Yes | Gateway UUID. |
| `gatehub.card_account_product_code` | string | Prod | Yes | Card product code. |
| `gatehub.paywiser_euro_vault_id` | string | Prod | Yes | EUR vault UUID. |
| `gatehub.sending_user_id` | string | Prod | Yes | Managed sending user UUID. |
| `gatehub.sending_user_address` | string | Prod | Yes | Sending user XRPL address. |
| `gatehub.intermediary_user_id` | string | Prod + node | Yes | Intermediary user UUID. Required in prod when `rafiki.node_enabled`. |
| `gatehub.intermediary_user_address` | string | Prod + node | Yes | Intermediary user XRPL address. Required in prod when `rafiki.node_enabled`. |
| `gatehub.webhook_secret` | string | No | Yes | Hex-encoded webhook signing secret. |
| `gatehub.fallback_webhook_url` | string | No | No | Secondary webhook URL for unrecognised users. |
| `gatehub.on_off_ramp_client_id` | string | Prod | No | OAuth client ID for on/off-ramp. |
| `gatehub.onboarding_client_id` | string | Prod | No | OAuth client ID for onboarding. |
| `gatehub.exchange_client_id` | string | Prod | No | OAuth client ID for exchange. |
| `gatehub.api_base_url` | string | Prod | No | GateHub API base URL. |
| `gatehub.onboarding_base_url` | string | Prod | No | Onboarding widget URL. |
| `gatehub.on_off_ramp_base_url` | string | Prod | No | Deposit/withdrawal widget URL. |
| `gatehub.eur_ops_account` | string | Prod | No | Internal EUR operations account ID. |
| `gatehub.eur_ops_ledger_id` | uint32 | Prod | No | Internal EUR ledger ID (must be non-zero in prod). |
| `gatehub.organization_id` | string | Prod | No | GateHub organisation UUID. |

### Xago

| Key | Type | Required | Secret | Notes |
|---|---|---|---|---|
| `xago.api_base_url` | string | Yes | No | Business API base URL. |
| `xago.identity_base_url` | string | Yes | No | Identity/login API base URL. |
| `xago.api_public_key` | string | Yes | Yes | API public key credential. |
| `xago.api_secret` | string | Yes | Yes | API secret key credential. |
| `xago.policy_id` | string | Yes | No | Login policy ID. |

### PTI / Fiant

Provider is gated by `pti.enabled`; the remaining fields are required only when enabled.

| Key | Type | Required | Secret | Notes |
|---|---|---|---|---|
| `pti.enabled` | bool | No | No | Enables PTI integration and config validation. |
| `pti.base_url` | string | If enabled | Yes | PTI REST API base URL. |
| `pti.jwk` | string | If enabled | Yes | Private RSA JWK for request signing. |
| `pti.client_id` | string | If enabled | No | Client UUID. |
| `pti.sdk_url` | string | If enabled | No | Fiant Web SDK bundle URL. |
| `pti.forms_url` | string | If enabled | No | Fiant hosted forms URL. |
| `pti.public_key_jwk` | string | If enabled | Yes | Public RSA JWK for webhook verification. |

### Persona

| Key | Type | Required | Secret | Notes |
|---|---|---|---|---|
| `persona.base_url` | string | Yes | No | Persona API base URL. |
| `persona.token` | string | Yes | Yes | Persona API token. |
| `persona.webhook_token` | string | Yes | Yes | Persona webhook verification token. |
| `persona.sandbox_fake_za_id` | bool | No | No | Sandbox-only: generate a synthetic South African ID. Has no effect in prod. |

### Twilio

Gated by `twilio.enabled`; credentials required only when enabled.

| Key | Type | Required | Secret | Notes |
|---|---|---|---|---|
| `twilio.enabled` | bool | Yes when `environment.mode: prod` | No | Enables live Twilio Verify calls. See behaviour below. |
| `twilio.account_sid` | string | If enabled | Yes | Twilio account SID. |
| `twilio.account_token` | string | If enabled | Yes | Twilio auth token. |
| `twilio.service_sid` | string | If enabled | Yes | Twilio Verify service SID. |

**Enabled vs. disabled behaviour**

- **`twilio.enabled: true`** — the backend initialises the real Twilio Verify
  service. `account_sid`, `account_token`, and `service_sid` are required, and
  the service SID is validated against Twilio at startup (startup fails fast on
  bad credentials).
- **`twilio.enabled: false`** — the backend initialises a **no-op verification
  service** instead of Twilio. It never contacts Twilio: verification codes are
  always reported as sent, and **any submitted OTP is accepted**. No credentials
  are required. This is intended for local development and testing.

**Production guardrail**

Running the no-op service in production would let anyone bypass phone
verification, so it is not allowed:

- When `environment.mode: prod`, `twilio.enabled` **must** be `true`. Setting it
  to `false` (or leaving it unset) is rejected at two layers:
  - **Chart render** — `helm template`/`helm install` fails with
    `backend.config.twilio.enabled must be true when environment.mode is prod`.
  - **Backend startup** — config validation returns
    `twilio.enabled must be true when environment.mode is prod`.
- In every other mode (`sandbox`, `dev`, `local`, `test`) `twilio.enabled`
  defaults to `false`. Override it to `true` to exercise the real Twilio
  integration in a non-production environment.

### Email (SendGrid)

Gated by `email.enabled`; SendGrid fields required only when enabled.

| Key | Type | Required | Secret | Notes |
|---|---|---|---|---|
| `email.enabled` | bool | No | No | Enables outgoing email. |
| `email.sendgrid.api_key` | string | If enabled | Yes | SendGrid API key. |
| `email.sendgrid.from_name` | string | If enabled | No | Sender display name. |
| `email.sendgrid.from_email` | string | If enabled | No | Sender email address. |
| `email.sendgrid.one_template_id` | string | If enabled | No | Dynamic template ID. |
| `email.sendgrid.support_email` | string | If enabled | No | Support inbox address. |

### Slack

| Key | Type | Required | Secret | Notes |
|---|---|---|---|---|
| `slack.token` | string | No | Yes | Bot token. |
| `slack.channel_signup_kyc` | string | No | No | Channel ID for signup/KYC notifications. Empty disables. |
| `slack.channel_transaction` | string | No | No | Channel ID for transaction notifications. Empty disables. |
| `slack.channel_error` | string | No | No | Channel ID for error/ops alerts. Empty disables. |
| `slack.client_id` | string | No | Yes | OAuth app client ID. |
| `slack.client_secret` | string | No | Yes | OAuth app client secret. |

### Chimoney

| Key | Type | Required | Secret | Notes |
|---|---|---|---|---|
| `chimoney.token` | string | No | Yes | Chimoney API token. |
| `chimoney.webhook_secret` | string | No | Yes | Webhook signature secret. |

### Admin portal

| Key | Type | Required | Secret | Notes |
|---|---|---|---|---|
| `admin.policy_aud` | string | Yes | No | Cloudflare Access policy audience. |
| `admin.team_domain` | string | Yes | No | Cloudflare Access team domain. |
| `admin.base_url` | string | Yes | No | Base URL of the admin portal. |

### Mobile deep-links

| Key | Type | Required | Secret | Notes |
|---|---|---|---|---|
| `mobile.apple_app_id` | string | Yes | No | Apple App Site Association app ID. |
| `mobile.android_package_name` | string | Yes | No | Android App Links package name. |
| `mobile.android_sha256` | string | Yes | No | Android signing certificate SHA-256. |

### Other integrations

| Key | Type | Required | Secret | Notes |
|---|---|---|---|---|
| `vault.addr` | string | No | No | HashiCorp Vault address (legacy encryption paths). |
| `vault.transit_engine_path` | string | No | No | Vault transit engine path. |
| `vault.token` | string | No | Yes | Vault token. |
| `sentry.dsn` | string | No | Yes | Sentry DSN for server-side error reporting. |
| `smarty.auth_id` | string | Yes | No | Smarty address-validation auth ID. |
| `smarty.auth_token` | string | Yes | Yes | Smarty address-validation auth token. |
| `pusher.addr` | string | Yes | No | Internal Pusher/Soketi service URL. |
| `segment.key` | string | Yes | Yes | Segment server-side write key. |

### Agreements

| Key | Type | Required | Secret | Notes |
|---|---|---|---|---|
| `agreements.folder_name` | string | Yes | No | Agreements folder name. |
| `agreements.signup_agreement_ids` | []string | No | No | Agreement IDs required at signup. |

### Observability

| Key | Type | Required | Secret | Notes |
|---|---|---|---|---|
| `otel.enabled` | bool | No | No | Explicit on/off switch for trace export. When `false`, a no-op exporter is used. |
| `otel.endpoint` | string | No | No | OTLP/gRPC endpoint. Overridden by `OTEL_EXPORTER_OTLP_ENDPOINT` if set (warning logged). |
| `otel.headers` | map[string]string | No | Yes | OTLP auth headers (e.g. Honeycomb ingest key). Overridden by `OTEL_EXPORTER_OTLP_HEADERS` if set (warning logged). |

> The service name reported in traces (`backend` / `backend-worker`) is set in code, not configuration.

### Conditional requirements

Some fields become required based on feature flags or environment mode:

| Condition | Additionally required |
|---|---|
| `twilio.enabled: true` | `twilio.account_sid`, `twilio.account_token`, `twilio.service_sid` |
| `environment.mode: prod` | `twilio.enabled: true` (disabling Twilio in prod is rejected) |
| `email.enabled: true` | `email.sendgrid.api_key`, `email.sendgrid.from_name`, `email.sendgrid.from_email`, `email.sendgrid.one_template_id`, `email.sendgrid.support_email` |
| `pti.enabled: true` | `pti.base_url`, `pti.jwk`, `pti.client_id`, `pti.sdk_url`, `pti.forms_url`, `pti.public_key_jwk` |
| `environment.mode: prod` | All GateHub fields marked *Prod* above (plus a non-zero `gatehub.eur_ops_ledger_id`) |
| `environment.mode: prod` **and** `rafiki.node_enabled: true` | `gatehub.intermediary_user_id`, `gatehub.intermediary_user_address` |

---

## Migration configuration (`MigrationConfig`)

The `migrate` command loads a smaller config (from its own `CONFIG` files, e.g. `backend.migration.yaml`). It uses the same `configa` loader, overlay, and secret-templating scheme.

| Key | Type | Required | Secret | Notes |
|---|---|---|---|---|
| `environment.mode` | string | Yes | No | One of `prod`, `sandbox`, `dev`, `local`, `test`. |
| `environment.label` | string | No | No | Telemetry tag. |
| `db_url` | string | Yes | Yes | Wallet backend PostgreSQL URL. |
| `pacioli_db_url` | string | Yes | Yes | Pacioli ledger PostgreSQL URL. |
| `open_payments_base_url` | string | Yes | No | Open Payments base URL. |
| `kratos_url` | string | No | No | Kratos public API URL. |
| `log_level` | string | No | No | Log verbosity. |
| `log_output_path` | string | No | No | Log destination. |
| `label` | string | No | No | Telemetry tag attached to monitoring signals. |
| `agreements.folder_name` | string | Yes | No | Agreements folder name. |
| `agreements.signup_agreement_ids` | []string | No | No | Agreement IDs required at signup. |
| `sentry.dsn` | string | No | Yes | Sentry DSN. |
