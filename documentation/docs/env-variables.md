# Environment Variables Reference

> **Configuration reference.** Central source for runtime environment variables, secret classification, and per-environment values.

> **Note:** Values shown as `<...>` are placeholders, not required for local development.

**Related documents:**

- [Home](index.md) - Full documentation map and topic index
- [Backend Configuration Guide](backend-configuration-guide.md) - The Go backend's YAML config scheme, secret handling, and full settings reference
- [Logging Reference](logging-reference.md) - Log configuration and structured logging policy
- [Signup Explainer](signup-guide.md) - Signup flow and auth-related runtime dependencies
- [KYC Explainer](kyc-guide.md) - Provider verification flows and integration dependencies
- [Provider Payments Guide](provider-payments-reference.md) - Provider behavior and configuration differences
- [GateHub Cards](gatehub-cards-guide.md) - GateHub integration configuration and card-specific secrets

**Quick Navigation:**

- **Need to compare service-level variables?** -> See [Protea (Frontend)](#protea-frontend) and [Botanist (Admin Portal)](#botanist-admin-portal)
- **Looking for backend configuration?** -> The backend is configured via YAML, not environment variables — see the [Backend Configuration Guide](backend-configuration-guide.md)
- **Need legacy variable context?** -> See [Legacy Variables](#legacy-variables)

This document covers the environment variables for the two **frontend** application services:

- **Protea** — user-facing frontend (`typescript/protea`)
- **Botanist** — admin portal (`typescript/botanist`)

> **The Go backend (`go/backend`) is not configured through environment variables.** It loads YAML config files via the `CONFIG` variable and resolves secrets through Kubernetes secret templating. Its only genuine environment variables are `CONFIG` and the `OTEL_EXPORTER_OTLP_*` overrides. See the [Backend Configuration Guide](backend-configuration-guide.md) for the full configuration scheme and settings reference.

Variables with **Secret: Yes** must be stored in 1Password and injected via a Kubernetes secret; real production secret values must never be committed to repository config files. Variables with **Secret: No** are safe to store in a ConfigMap or compose file.

---

## Protea (Frontend)

Protea is a Remix application serving the user-facing wallet UI.

| Variable | Description | Secret | Notes |
|---|---|---|---|
| `NODE_ENV` | Node.js runtime mode | No | Deployed: `production`; Local: `development` |
| `ILW_ENV` | Runtime environment tag used for feature flags (e.g. Xago test-deposit gate) | No | Prod: `prod`; Sandbox/Dev: `dev`; Local: `local` |
| `LOG_LEVEL` | Log verbosity (`debug`, `info`, `warn`, `error`) | No | Deployed: `info`; Local: `debug` |
| `LOG_PRETTY` | Human-readable log formatting. Set `false` for JSON in deployed environments | No | Deployed: `false`; Local: `true` |
| `TARGET_HOST` | Base URL of the app itself (scheme + host). Used to build self-referential links in legal and contact pages | No | Prod: `https://interledger.app`; Sandbox/Dev: environment URL; Local: `https://interledger.test` |
| `SUPPORT_EMAIL` | Support email address shown to users in error messages, legal pages, and notifications | No | Prod/Sandbox/Dev: `support@interledger.app`; Local: `support@interledger.app` |
| `KRATOS_URL` | Internal URL for the Ory Kratos public API (session/login flows) | No | Deployed: `http://kratos-public`; Local: `http://kratos:4433` |
| `RAFIKI_AUTH_ENDPOINT` | Internal URL for the Rafiki auth gRPC/GraphQL endpoint | No | Deployed: `http://rafiki-auth-service:3009`; Local: `http://rafiki_auth:3009` |
| `PUBLIC_OP_AUTH_HOST` | Public-facing hostname for the Open Payments auth server (used by the Rafiki auth flow) | No | Prod: `auth.ilp.link`; Sandbox: `auth.sandbox.ilp.link`; Dev: `auth.development.ilp.link`; Local: `auth.local.ilp.link` |
| `PAYMENT_POINTER_BASE` | Domain used to build Open Payments payment pointer addresses | No | Prod: `ilp.link`; Sandbox: `sandbox.ilp.link`; Dev: `development.ilp.link`; Local: `local.ilp.link` |
| `BACKEND_GRPC_URL` | Internal URL for the wallet backend gRPC server | No | Deployed: `http://wallet-backend-service-grpc:8443`; Local: `http://backend:8443` |
| `BACKEND_HTTP_URL` | Internal URL for the wallet backend HTTP server | No | Deployed: `http://wallet-backend-service:8080`; Local: `http://backend:8080` |
| `DEFAULT_RATE_LIMIT_REQUESTS` | Max requests allowed per time window before rate limiting kicks in | No | Local default: `4` (code default); deployed values TBD |
| `DEFAULT_RATE_LIMIT_TIME` | Rate limit time window in seconds | No | Local default: `3600` (code default); deployed values TBD |
| `PTI_CLIENT_ID` | PTI/Fiant payment provider client UUID, passed to the browser for payment widget initialisation | No | Local default: `''` (code default); deployed values TBD |
| `PTI_SDK_URL` | URL to the Fiant Web SDK JavaScript bundle loaded by the PTI payment widget. Value pattern documented in [Fiant Front-End SDK usage](https://developers.platform.fiant.io/docs/front-end-sdk-usage) as `https://sdk.{env}.fiant.io/latest/index.js` | No | Prod: `https://sdk.platform.fiant.io/0.0.23/index.js`; Sandbox/Dev: `https://sdk.staging.fiant.io/latest/index.js`; Local: `https://mockpti.interledger.test/sdk/index.js` |
| `PTI_FORMS_URL` | URL to the Fiant hosted forms (Elements) used for KYC, onboarding, and payment collection widgets. Derived from the `ptiDomain` init parameter documented in [Fiant Front-End SDK usage](https://developers.platform.fiant.io/docs/front-end-sdk-usage): `https://forms.{ptiDomain}` | No | Prod: `https://forms.platform.fiant.io`; Sandbox/Dev: `https://forms.staging.fiant.io`; Local: `https://mockpti.interledger.test/forms` |
| `PERSONA_SDK_URL` | URL to the Persona identity verification JavaScript SDK loaded by the KYC flow | No | Prod/Sandbox: `https://cdn.withpersona.com/dist/persona-v4.8.0-alpha.js`; Local: `https://mockxago.interledger.test/v1/persona-sdk.js` |
| `MOCKXAGO_ENDPOINT` | Base URL for the Xago/Persona KYC iframe (`/v1/inquiries/<id>/iframe` appended at runtime). When set, the Xago iframe flow is used; when empty, the Persona SDK flow is used instead. Optional. | No | Local: `https://mockxago.interledger.test`; not set in deployed environments (Persona SDK used instead) |
| `OP_INTPAY_ENABLED` | Feature flag for Open Payments Interledger Pay (`/quick-pay` routes). When `false`, the remaining `OP_INTPAY_*` variables are unused and the quick-pay routes reject requests | No | Local default: `false` |
| `OP_INTPAY_HOST` | Public base URL of the wallet's Open Payments auth server, used to build quick-pay links. Required when `OP_INTPAY_ENABLED=true` | No | Local: `https://interledger.test/`; deployed values TBD |
| `OP_INTPAY_REDIRECT_URL` | URL the client is redirected back to once a quick-pay grant completes. Required when `OP_INTPAY_ENABLED=true` | No | Local: `https://interledger.test/quick-pay/finish`; deployed values TBD |
| `OP_INTPAY_WALLET_ADDRESS` | Wallet address (payment pointer) used as the Open Payments client identity for quick-pay grant requests. Required when `OP_INTPAY_ENABLED=true` | No | Local default: empty; deployed values TBD |
| `SENTRY_RELEASE` | Identifies the deployed version in Sentry error reports | No | Not set by default; deployed values TBD |
| `SENTRY_ENV_LABEL` | Environment label sent to Sentry with every event (e.g. `prod`, `dev`, `local`). Only relevant when `SENTRY_DSN` is set. **Sentry is disabled by default in local.** | No | Prod: `prod`; Sandbox: `sandbox`; Dev: `dev`; Local: not set (Sentry off) |
| `CHOKIDAR_USEPOLLING` | Enable filesystem polling for hot-reload in containers (dev only) | No | Local only: `true`; not applicable in deployed environments |
| `COOKIE_SECRETS` | JSON array of strings used to sign session cookies. Rotate periodically. | Yes | Local default: `["localsecret"]` |
| `RAFIKI_AUTH_SECRET` | Shared secret between Protea and the Rafiki auth service | Yes | Local default: `my-super-secret-identity-key` |
| `PUSHER_APP_KEY` | Pusher application key for real-time push notifications (public-facing). Optional - realtime degrades gracefully when unset. | Yes | Local default: empty |
| `PUSHER_APP_CLUSTER` | Pusher cluster region for the application | Yes | Local default: `eu` |
| `REDIS_URL` | Redis connection URL used for session storage and caching | Yes | Local default: `redis://redis:6379/2` |
| `SENTRY_DSN` | Sentry DSN for client-side and server-side error reporting. **Not set in local** — Sentry is disabled when this is empty. | Yes | Not set locally; required in deployed environments |
| `SEGMENT_API_KEY` | Segment analytics write key for event tracking | Yes | Not set locally |
| `GOOGLE_MAPS_API_KEY` | Google Maps API key for geocoding and places autocomplete endpoints used during onboarding | Yes | Not set locally |
| `OP_INTPAY_KEY_ID` | Key id of the Open Payments client used to authenticate quick-pay grant requests. Required when `OP_INTPAY_ENABLED=true` | Yes | Local default: empty |
| `OP_INTPAY_PRIVATE_KEY` | Base64-encoded Ed25519 private key of the Open Payments client, paired with `OP_INTPAY_KEY_ID`. Required when `OP_INTPAY_ENABLED=true` | Yes | Local default: empty |

### Legacy Variables

| Variable | Description | Secret | Notes |
|---|---|---|---|
| `BT_TOKEN` | Basis Theory token for legacy frontend card tokenisation integration paths | Yes | Legacy variable in local compose; not used in current Protea app code |
| `CF_TURNSTILE_SITE_KEY` | Cloudflare Turnstile site key used in legacy bot-protection integration | Yes | Keep documented, no environment value guidance |
| `CF_TURNSTILE_SECRET_KEY` | Cloudflare Turnstile secret key used in legacy server verification | Yes | Keep documented, no environment value guidance |
| `GATE_SIGNUP` | Signup gating feature flag from older frontend config wiring | No | Legacy variable in local compose; not used in current Protea app code |

---

## Botanist (Admin Portal)

Botanist is a Remix application providing the internal admin interface. It connects to the backend admin gRPC port. Botanist has no secret variables — all sensitive operations are delegated to the backend. The admin gRPC port must only be accessible from within the cluster.

| Variable | Description | Secret | Notes |
|---|---|---|---|
| `ILW_ENV` | Runtime environment tag | No | Prod: `prod`; Sandbox/Dev: `dev`; Local: `local` |
| `BACKEND_GRPC_URL` | Internal URL for the backend admin gRPC port (8448) | No | Deployed: `wallet-backend-service-grpc:8448`; Local: `backend:8448` |
| `PAYMENT_POINTER_BASE` | Domain used to display payment pointer addresses for users | No | Prod: `ilp.link`; Sandbox: `sandbox.ilp.link`; Dev: `development.ilp.link`; Local: `local.ilp.link` |

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
