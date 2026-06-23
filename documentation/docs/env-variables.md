# Environment Variables Reference

> **Configuration reference.** Central source for runtime environment variables, secret classification, and per-environment values.

**Related documents:**

- [Home](index.md) - Full documentation map and topic index
- [Logging Reference](logging-reference.md) - Log configuration and structured logging policy
- [Signup Explainer](signup-guide.md) - Signup flow and auth-related runtime dependencies
- [KYC Explainer](kyc-guide.md) - Provider verification flows and integration dependencies
- [Provider Payments Guide](provider-payments-reference.md) - Provider behavior and configuration differences
- [GateHub Cards](gatehub-cards-guide.md) - GateHub integration configuration and card-specific secrets

**Quick Navigation:**

- **Need to compare service-level variables?** -> See [Protea (Frontend)](#protea-frontend), [Botanist (Admin Portal)](#botanist-admin-portal), and [Backend (Wallet)](#backend-wallet)
- **Need legacy variable context?** -> See [Legacy Variables](#legacy-variables)

This document covers all environment variables for the three deployed application services:

- **Protea** — user-facing frontend (`typescript/protea`)
- **Botanist** — admin portal (`typescript/botanist`)
- **Backend** — Go wallet backend (`go/backend`)

Variables with **Secret: Yes** must be stored in 1Password and injected via a Kubernetes secret; real production secret values must never be committed to repository config files. Variables with **Secret: No** are safe to store in a ConfigMap or compose file.

---

## Protea (Frontend)

Protea is a Remix application serving the user-facing wallet UI.

| Variable | Description | Secret | Notes |
|---|---|---|---|
| `NODE_ENV` | Node.js runtime mode | No | Deployed: `production`; Local: `development` |
| `FYNBOS_ENV` | Runtime environment tag used for feature flags (e.g. Xago test-deposit gate) | No | Prod: `prod`; Sandbox/Dev: `dev`; Local: `local` |
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
| `SENTRY_RELEASE` | Identifies the deployed version in Sentry error reports | No | Not set by default; deployed values TBD |
| `SENTRY_ENV_LABEL` | Environment label sent to Sentry with every event (e.g. `prod`, `dev`, `local`). Only relevant when `SENTRY_DSN` is set. **Sentry is disabled by default in local.** | No | Prod: `prod`; Sandbox: `sandbox`; Dev: `dev`; Local: not set (Sentry off) |
| `CHOKIDAR_USEPOLLING` | Enable filesystem polling for hot-reload in containers (dev only) | No | Local only: `true`; not applicable in deployed environments |
| `COOKIE_SECRETS` | JSON array of strings used to sign session cookies. Rotate periodically. | Yes | Local default: `["localsecret"]` |
| `RAFIKI_AUTH_SECRET` | Shared secret between Protea and the Rafiki auth service | Yes | Local default: `my-super-secret-identity-key` |
| `PUSHER_APP_KEY` | Pusher application key for real-time push notifications (public-facing) | Yes | Local default: `91988d6075551d29760a` |
| `PUSHER_APP_CLUSTER` | Pusher cluster region for the application | Yes | Local default: `eu` |
| `REDIS_URL` | Redis connection URL used for session storage and caching | Yes | Local default: `redis://redis:6379/2` |
| `SENTRY_DSN` | Sentry DSN for client-side and server-side error reporting. **Not set in local** — Sentry is disabled when this is empty. | Yes | Not set locally; required in deployed environments |
| `SEGMENT_API_KEY` | Segment analytics write key for event tracking | Yes | Not set locally |
| `GOOGLE_MAPS_API_KEY` | Google Maps API key for geocoding and places autocomplete endpoints used during onboarding | Yes | Not set locally |

### Legacy Variables

| Variable | Description | Secret | Notes |
|---|---|---|---|
| `DATO_API_TOKEN` | DatoCMS API token used by legacy CMS content loading paths | Yes | Keep documented, no environment value guidance |
| `BT_TOKEN` | Basis Theory token for legacy frontend card tokenisation integration paths | Yes | Legacy variable in local compose; not used in current Protea app code |
| `CF_TURNSTILE_SITE_KEY` | Cloudflare Turnstile site key used in legacy bot-protection integration | Yes | Keep documented, no environment value guidance |
| `CF_TURNSTILE_SECRET_KEY` | Cloudflare Turnstile secret key used in legacy server verification | Yes | Keep documented, no environment value guidance |
| `GATE_SIGNUP` | Signup gating feature flag from older frontend config wiring | No | Legacy variable in local compose; not used in current Protea app code |

---

## Botanist (Admin Portal)

Botanist is a Remix application providing the internal admin interface. It connects to the backend admin gRPC port. Botanist has no secret variables — all sensitive operations are delegated to the backend. The admin gRPC port must only be accessible from within the cluster.

| Variable | Description | Secret | Notes |
|---|---|---|---|
| `FYNBOS_ENV` | Runtime environment tag | No | Prod: `prod`; Sandbox/Dev: `dev`; Local: `local` |
| `BACKEND_GRPC_URL` | Internal URL for the backend admin gRPC port (8448) | No | Deployed: `wallet-backend-service-grpc:8448`; Local: `backend:8448` |
| `PAYMENT_POINTER_BASE` | Domain used to display payment pointer addresses for users | No | Prod: `ilp.link`; Sandbox: `sandbox.ilp.link`; Dev: `development.ilp.link`; Local: `local.ilp.link` |

---

## Backend (Wallet)

The Go backend is the core of the wallet, handling payments, provider integrations, webhooks, Temporal workflows, and gRPC APIs.

### General & Observability

| Variable | Description | Secret | Notes |
|---|---|---|---|
| `FYNBOS_ENV` | Runtime environment tag | No | Prod: `prod`; Sandbox/Dev: `dev`; Local: `local` |
| `LOG_LEVEL` | Log verbosity | No | Dev: `debug`; all others: `info` |
| `PORT` | HTTP server port (webhooks, health) | No | All environments: `8080` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OpenTelemetry collector endpoint | No | All environments: `grpc://api.honeycomb.io:443` |
| `OTEL_SERVICE_NAME` | Service name in traces | No | All environments: `backend` |
| `OTEL_EXPORTER_OTLP_HEADERS` | Auth headers for the OTLP endpoint | Yes | Local default: `x-honeycomb-team=7Qskhns7Dc7wgazrDe6yZD` |
| `SENTRY_DSN` | Sentry DSN for server-side error reporting | Yes | Not set locally |
| `SENTRY_RELEASE` | Sentry release identifier | No | Not set locally; deployed values TBD |

### Database

| Variable | Description | Secret | Notes |
|---|---|---|---|
| `DB_URL` | Wallet backend PostgreSQL connection URL | Yes | Local default: `postgres://postgres:postgres@postgres:5432/backend?sslmode=disable` |
| `DB_URL_WITH_CERTS` | Backend DB URL with TLS certificates (IAM auth in GCP) | Yes | Local default: same as `DB_URL` |
| `PACIOLI_DB_URL` | Pacioli ledger PostgreSQL connection URL | Yes | Local default: `postgres://postgres:postgres@postgres:5432/pacioli?sslmode=disable` |
| `PACIOLI_DB_URL_WITH_CERTS` | Pacioli DB URL with TLS certificates | Yes | Local default: same as `PACIOLI_DB_URL` |
| `RAFIKI_DB_URL` | Rafiki backend PostgreSQL connection URL | Yes | Local default: `postgres://postgres:postgres@postgres:5432/rafiki_backend?sslmode=disable` |
| `RAFIKI_AUTH_DB_URL` | Rafiki auth PostgreSQL connection URL | Yes | Local default: `postgres://postgres:postgres@postgres:5432/rafiki_auth?sslmode=disable` |

### Ledger & Wallet Core

| Variable | Description | Secret | Notes |
|---|---|---|---|
| `REDIS_URL` | Redis connection URL for queues and caching | Yes | Local default: `redis://redis:6379/0` (provided by local compose); currently no direct `REDIS_URL` reference under `go/backend` |
| `ALLOWED_WALLET_IDS` | Comma-separated wallet IDs allowed through regional blocks | No | All environments: empty |
| `BLOCKED_REGIONS` | Comma-separated ISO country codes to block wallet access from | No | Prod: `US`; all others: empty |
| `APPLE_APP_ID` | Apple App Site Association app ID for deep-links | No | Prod: `6B7AFCRT3V.app.wallet.interledger`; Sandbox: TBD; Dev: `6B7AFCRT3V.app.wallet.interledger.dev`; Local: `6B7AFCRT3V.app.wallet.interledger.test` |
| `ANDROID_PACKAGE_NAME` | Android App Links package name for deep-links | No | Prod: `org.interledger.walletandroid`; Sandbox: TBD; Dev: `org.interledger.walletandroid.debugDevelopment`; Local: `org.interledger.walletandroid.test` |
| `ANDROID_SHA256` | Android signing certificate SHA-256 fingerprint for deep-links | No | Prod: `25:D0:6A:A1:4F:BD:F6:05:93:2C:80:A2:AC:DA:DE:E4:2F:72:BC:23:9C:8F:2C:78:85:49:9A:AC:D2:4B:A4:FF`; Sandbox: TBD |
| `ADMIN_BASE_URL` | Base URL of the admin portal, used to build links to it for support purposes. **Required** | No | Full URL including scheme, e.g. `https://admin.interledger.app` |

### Infrastructure Services

| Variable | Description | Secret | Notes |
|---|---|---|---|
| `TEMPORAL_URL` | Temporal frontend address for workflow orchestration | No | Deployed: `temporal-frontend-headless:7233`; Local: `temporal:7233` |
| `KRATOS_URL` | Kratos public API URL for session validation | No | Deployed: `http://kratos-public`; Local: `http://kratos:4433` |
| `KRATOS_ADMIN_URL` | Kratos admin API URL for user management | No | Deployed: `http://kratos-admin`; Local: `http://kratos:4434` |
| `RAFIKI_BACKEND_GRAPHQL_URL` | Rafiki backend GraphQL endpoint | No | Deployed: `http://rafiki-backend-service:3001/graphql`; Local: `http://rafiki_backend:3001/graphql` |
| `RAFIKI_AUTH_GRAPHQL_URL` | Rafiki auth GraphQL endpoint | No | Deployed: `http://rafiki-auth-service:3003/graphql`; Local: `http://rafiki_auth:3003/graphql` |
| `PUSHER_ADDR` | Internal URL for the Pusher/Soketi service used for push notifications | Yes | Local default: `http://pusher:8080` |

### Admin Portal

| Variable | Description | Secret | Notes |
|---|---|---|---|
| `ADMIN_POLICY_AUD` | Cloudflare Access policy audience for the admin portal | Yes | Local default: `admin-policy` |
| `ADMIN_TEAM_DOMAIN` | Cloudflare Access team domain for the admin portal | Yes | Local default: `admin-team` |

### Twilio (Phone Verification)

| Variable | Description | Secret | Notes |
|---|---|---|---|
| `TWILIO_ACCOUNT_SID` | Twilio account SID | Yes | Local default: `SK021f793191208ba69c3bea87dd426085` |
| `TWILIO_SERVICE_SID` | Twilio Verify service SID | Yes | Local default: `VA8af4e130da63b9fac4c042acbc33a267` |
| `TWILIO_ACCOUNT_TOKEN` | Twilio auth token | Yes | Local default: `test` |
| `TWILIO_ENABLED` | Enables live Twilio Verify calls. When `false`, all phone verification methods return stub responses and credentials are not required. | No | Prod/Sandbox: `true`; Dev/Local: `false` |

### SendGrid (Transactional Email)

| Variable | Description | Secret | Notes |
|---|---|---|---|
| `EMAIL_ENABLED` | Set to `false` to disable all outgoing emails. Defaults to enabled if unset. When disabled, SendGrid env vars are not required. | No | Local default: `false` |
| `SENDGRID_API_KEY` | SendGrid API key for sending emails. Required when `EMAIL_ENABLED` is not `false`. | Yes | Local default: `test-sendgrid-api-key` |
| `SENDGRID_FROM_NAME` | Display name for outgoing emails (e.g. "Interledger Wallet"). Required when `EMAIL_ENABLED` is not `false`. | No | Local default: `Interledger Wallet` |
| `SENDGRID_FROM_EMAIL` | Sender email address for outgoing emails. Required when `EMAIL_ENABLED` is not `false`. | No | Local default: `support@interledger.app` |
| `SENDGRID_ONE_TEMPLATE_ID` | SendGrid Dynamic Template ID used by backend transactional emails. Required when `EMAIL_ENABLED` is not `false`. | No | Local default: `d-12030774d225454ea91720034b9adb97` |

### Segment (Analytics)

| Variable | Description | Secret | Notes |
|---|---|---|---|
| `SEGMENT_KEY` | Segment write key for server-side event tracking | Yes | Local default: `test-segment-key` |

### Persona (Identity Verification)

| Variable | Description | Secret | Notes |
|---|---|---|---|
| `PERSONA_TOKEN` | Persona API token for identity verification. Required at backend startup. | Yes | Prod/Sandbox/Dev: secret from 1Password; Local default: `test-persona-token` |
| `PERSONA_WEBHOOK_TOKEN` | Persona webhook verification token. Required at backend startup. | Yes | Prod/Sandbox/Dev: secret from 1Password; Local default: `test-persona-webhook-token` |
| `PERSONA_BASE_URL` | Persona API base URL override used by backend Persona client | No | Default when unset: `https://api.withpersona.com/api/v1/`; Local default in compose: `http://mockxago:8080/v1/` |
| `PERSONA_SANDBOX_ZA_FAKE_ZA_ID` | **Persona sandbox workaround only.** When `true`, the backend generates a synthetic South African ID number instead of fetching one from Persona. This is required in the Persona sandbox environment because Persona's sandbox always returns an American user profile — meaning the South African ID field is always null. Without this flag, Xago subaccount creation fails for all sandbox users since a valid ZA ID is a required field. Has no effect in production, where Persona returns real ZA ID documents. | No | Default: `false`; Set to `true` in sandbox |

### Slack (Notifications)

| Variable | Description | Secret | Notes |
|---|---|---|---|
| `SLACK_TOKEN` | Slack bot token for sending notifications | Yes | Not set locally |
| `SIGNUP_KYC_SLACK_CHANNEL` | Slack channel ID for signup, identity, linked-account, and KYC notifications. Empty disables this category. | No | Not set locally |
| `TRANSACTION_SLACK_CHANNEL` | Slack channel ID for transaction-lifecycle notifications (new transactions, fiant transaction-status webhooks). Empty disables this category. | No | Not set locally |
| `ERROR_SLACK_CHANNEL` | Slack channel ID for error/ops alerts: provider RollbackReserve errors, withdrawal failures, asset-not-configured warnings, gatehub card-transaction anomalies. Empty disables this category. | No | Not set locally |
| `SLACK_CLIENT_ID` | Slack OAuth app client ID | Yes | Not set locally |
| `SLACK_CLIENT_SECRET` | Slack OAuth app client secret | Yes | Not set locally |
| `SLACK_SIGNING_SECRET` | Slack signing secret for webhook verification | Yes | Not set locally |

### Chimoney (Payment Provider)

| Variable | Description | Secret | Notes |
|---|---|---|---|
| `CHIMONEY_TOKEN` | Chimoney API token | Yes | Not set locally |
| `CHIMONEY_WEBHOOK_SECRET` | Chimoney webhook signature secret | Yes | Not set locally |

### GateHub (Payment Provider)

GateHub is used for EUR XRPL transactions, onboarding, and debit card issuing.

> **Sandbox/Development note:** The Sandbox environment is missing explicit Gatehub API URL configuration in its values file. It is expected to inherit from the development environment base. Verify this is intentional before promoting changes.

> **Production client IDs differ:** GateHub OAuth client IDs for on/off-ramp, onboarding, and exchange are different in production vs sandbox/development.

| Variable | Description | Secret | Notes |
|---|---|---|---|
| `GATEHUB_API_BASE_URL` | Base URL for GateHub API calls | No | Prod: `https://api.gatehub.net`; Sandbox/Dev: `https://api.sandbox.gatehub.net`; Local: `http://mockgatehub:8080` |
| `GATEHUB_ONBOARDING_BASE_URL` | Frontend widget URL for GateHub KYC onboarding | No | Prod: `https://onboarding.gatehub.net`; Sandbox/Dev: `https://onboarding.sandbox.gatehub.net`; Local: `https://mockgatehub.interledger.test/iframe/onboarding` |
| `GATEHUB_ON_OFF_RAMP_BASE_URL` | Frontend widget URL for GateHub deposit/withdrawal | No | Prod: `https://managed-ramp.gatehub.net`; Sandbox/Dev: `https://managed-ramp.sandbox.gatehub.net`; Local: `https://mockgatehub.interledger.test` |
| `GATEHUB_ON_OFF_RAMP_CLIENT_ID` | OAuth client ID for GateHub on/off-ramp token issuance | No | Prod: `f4c8f30f-7fc3-4aa1-8573-520cb67565e3`; Sandbox/Dev/Local: `f8119dfd-e563-44ee-9ae2-1e60a4fce74f` |
| `GATEHUB_ONBOARDING_CLIENT_ID` | OAuth client ID for GateHub onboarding token issuance | No | Prod: `40a22fc5-9091-4c6f-aff6-a3fddf475b33`; Sandbox/Dev/Local: `4df24d1b-5796-4eec-951b-21699d61b970` |
| `GATEHUB_EXCHANGE_CLIENT_ID` | OAuth client ID for GateHub exchange token issuance | No | Prod: `50e7c590-f6f9-4fa9-9498-260bd978c5d6`; Sandbox/Dev/Local: `4e28d4df-22d7-414c-97a3-d71956df29ba` |
| `GATEHUB_FALLBACK_WEBHOOK_URL` | Secondary webhook URL for unrecognised GateHub users (card fallback) | No | Prod: `https://api.interledger.cards/gatehub-webhooks`; Sandbox/Dev: empty; Local: not set |
| `GATEHUB_CARD_ACCOUNT_PRODUCT_CODE` | Card product code applied when creating GateHub card accounts | Yes | Local default: `PWSR_DEBP_2404` |
| `GATEHUB_ORGANIZATION_ID` | GateHub organisation UUID for org-level operations | No | Prod: TBD; Sandbox/Dev: `a5b3de06-3c61-47aa-82fa-9d9e1875f42c`; Local: `default-org` |
| `GATEHUB_EUR_OPS_ACCOUNT` | Internal EUR operations account ID for ledger movements | No | All environments: `1854f171-eafa-4e30-bf66-7dbfe167ccfa` |
| `GATEHUB_EUR_OPS_LEDGER_ID` | Internal EUR ledger ID for GateHub reserve/assign posting | No | All environments: `4482387` |
| `GATEHUB_APP_ID` | GateHub API app ID header value | Yes | Local default: `local-test-app-id` |
| `GATEHUB_SECRET` | GateHub API signing secret | Yes | Local default: `local-test-app-secret` |
| `GATEHUB_WEBHOOK_SECRET` | Hex-encoded webhook signing secret for verifying GateHub webhooks | Yes | Local default: `6d6f636b5f776562686f6f6b5f736563726574` |
| `GATEHUB_GATEWAY_ID` | GateHub gateway UUID for linking users to a hub | Yes | Local default: `b18965b5-8bb0-486e-88b2-7be47b974ac1` |
| `GATEHUB_PAYWISER_EURO_VAULT_ID` | GateHub EUR vault UUID used in hosted transactions | Yes | Local default: `a09a0a2c-1a3a-44c5-a1b9-603a6eea9341` |
| `GATEHUB_CARD_APP_ID` | GateHub card app ID header value (separate from main app) | Yes | Local default: `local-test-card-app-id` |
| `GATEHUB_SENDING_USER_ID` | Managed GateHub user UUID used for backfill payment operations | Yes | Dev: `5dd8e7f7-ead5-4f8d-b956-1acfecc231b4`; Local: `test-sending-user-id`; Prod/Sandbox: secret |
| `GATEHUB_SENDING_USER_ADDRESS` | XRPL wallet address of the managed sending user | Yes | Dev: `191968355`; Local: `rN7n7otQDd6FczFgLdVZGMbpKRtFVfT4hb`; Prod/Sandbox: secret |

### Xago (Payment Provider)

Xago is used for ZAR and USD transactions in South Africa.

| Variable | Description | Secret | Notes |
|---|---|---|---|
| `XAGO_API_BASE_URL` | Xago business API base URL | No | Sandbox/Dev: `https://test.xago.io/exchange/v1`; Local: `http://mockxago:8080/v1`; Prod: TBD |
| `XAGO_IDENTITY_BASE_URL` | Xago identity/login API base URL | No | Sandbox/Dev: `https://test.xago.io/identity/v1`; Local: `http://mockxago:8080/v1`; Prod: TBD |
| `XAGO_POLICY_ID` | Xago login policy ID for requesting access tokens | No | Sandbox/Dev/Local: `5e2585a474b0e90012ce8ff1`; Prod: TBD |
| `XAGO_API_SECRET` | Xago API secret key credential | Yes | Local default: `test-secret` |
| `XAGO_API_PUBLIC_KEY` | Xago API public key credential | Yes | Local default: `test-public-key` |

### PTI / Fiant (Payment Provider)

PTI is a US-focused payment provider currently disabled except in production.

All PTI endpoint URLs follow a consistent environment pattern documented in the [Fiant Front-End SDK usage guide](https://developers.platform.fiant.io/docs/front-end-sdk-usage):

- **Staging**: `*.staging.fiant.io` (used by sandbox.and development environments)
- **Production**: `*.platform.fiant.io`
- **Local**: mockpti emulator (`mockpti.interledger.test`)

For deployed environments the `PTI_JWK`, `PTI_PUBLIC_KEY_JWK`, and `PTI_BASE_URL` values are secrets stored in the appropriate environment's **1Password vault** and injected via Kubernetes `OnePasswordItem` resources. Non-secret variables (`PTI_CLIENT_ID`, `PTI_SDK_URL`, `PTI_FORMS_URL`) are safe to store in ConfigMaps.

| Variable | Description | Secret | Notes |
|---|---|---|---|
| `PTI_ENABLED` | Enables PTI provider integration and runtime config validation | No | Prod: `true`; all others: `false` |
| `PTI_BASE_URL` | PTI REST API base URL (see [Fiant API Reference](https://developers.platform.fiant.io/reference)) | Yes | Prod: secret (1Password); Sandbox: `https://api.staging.fiant.io/v1/`; Dev: empty; Local: `http://mockpti:8080` |
| `PTI_SDK_URL` | URL to the Fiant Web SDK JavaScript bundle. The [Fiant docs](https://developers.platform.fiant.io/docs/front-end-sdk-usage) specify the pattern `https://sdk.{env}.fiant.io/latest/index.js` where `{env}` is `staging` or `platform` | No | Prod: `https://sdk.platform.fiant.io/0.0.23/index.js`; Sandbox/Dev: `https://sdk.staging.fiant.io/latest/index.js`; Local: `https://mockpti.interledger.test/sdk/index.js` |
| `PTI_FORMS_URL` | URL to the Fiant hosted forms (Elements) endpoint. The [Fiant docs](https://developers.platform.fiant.io/docs/front-end-sdk-usage) document the `ptiDomain` init parameter with values `staging.fiant.io` (staging) and `platform.fiant.io` (production); the forms URL is `https://forms.{ptiDomain}` | No | Prod: `https://forms.platform.fiant.io`; Sandbox/Dev: `https://forms.staging.fiant.io`; Local: `https://mockpti.interledger.test/forms` |
| `PTI_CLIENT_ID` | PTI client UUID passed to the browser for `PTI.init()` widget initialisation and used server-side for webhook validation | No | Prod: `f4c8f30f-...` (confirm with PTI); Sandbox: `81a556d8-106d-4c93-8c6f-f2f8e555b4f0`; Dev: empty; Local default: `04d3e1b5-96d4-47e4-9eaa-13e9b4b0f219` |
| `PTI_JWK` | PTI private RSA JWK used for request signing and webhook crypto | Yes | Prod/Sandbox: secret (1Password); Local default: test RSA key (see local compose) |
| `PTI_PUBLIC_KEY_JWK` | PTI public RSA JWK used for webhook signature verification | Yes | Prod/Sandbox: secret (1Password); Local default: test RSA public key (see local compose) |
| `FYNBOS_BACKEND_HOST` | Host used by the PTI mock webhook proxy (`/webhooks/pti`) when forwarding requests to the wallet backend | No | Not set in any environment |

### Plaid (POC)

Plaid integration is a proof-of-concept (see `documentation/poc/plaid/`). The feature is gated behind `PLAID_ENABLED`. When disabled (the default), no runtime validation is performed and the `/plaid/*` HTTP routes are not registered. The current POC only targets Plaid Sandbox.

`PLAID_ENABLED` is read by **two** services and they MUST agree per environment: the **backend** uses it to wire the Plaid provider/routes, and **Protea** reads its own `PLAID_ENABLED` (server-side) to pick the US bank-link flow in the UI — `true` = Plaid (`/connect/bank`), `false` = manual bank-details form (`/connect/bank/us`). If they disagree the UI and backend mismatch. In local both default off the same `BACKEND_PLAID_ENABLED` root var (`local/wallet.yaml`, `local/protea.yaml`); in deploy configs set the var on both workloads.

| Variable | Description | Secret | Notes |
|---|---|---|---|
| `PLAID_ENABLED` | Backend: enables Plaid POC integration, runtime config validation, and `/plaid/*` HTTP routes. Protea: selects the US bank-link flow (Plaid vs manual). Must match across both services. | No | Default: `false`. Local POC: `true`; all deployed environments: `false` until POC promoted |
| `PLAID_CLIENT_ID` | Plaid API client ID from the Plaid dashboard | Yes | Local POC: developer's sandbox client ID (1Password / personal); not set in deployed environments |
| `PLAID_SECRET` | Plaid API secret matching the chosen `PLAID_ENV` | Yes | Local POC: developer's sandbox secret; not set in deployed environments |
| `PLAID_ENV` | Plaid environment selector (`sandbox` or `production`) | No | Local POC: `sandbox`; deployed: unset |
| `PLAID_PRODUCTS` | Comma-separated list of Plaid products requested at Link creation | No | POC: `auth,transactions,balance,identity` |
| `PLAID_COUNTRY_CODES` | Comma-separated ISO-3166-1 alpha-2 country codes for Link institution filtering | No | POC: `US` (sandbox only supports US institutions out of the box) |
| `PLAID_PROCESSOR` | Plaid processor partner used when minting a processor token via `processor/token/create` (`fiant` or `zero_hash`). Phase 2 only. | No | Default: `fiant`. Use `zero_hash` to validate plumbing if `fiant` is not enabled in your Plaid team. |
| `PLAID_API_URL` | Overrides the Plaid SDK base URL. When set, all Plaid REST calls target this host instead of the real `sandbox.plaid.com`/`production.plaid.com`. Used to point the backend at the local **mockplaid** service. | No | Local POC: `http://mockplaid:8080` (mock). Blank/unset for real Plaid. Deployed: unset. See `documentation/poc/mockplaid/`. |

#### MockPlaid service (local only)

The `mockplaid` mock service (`go/mock/mockplaid/`, started by `local/docker-compose.yaml`) reads its own env. These are **local-only**; not set in deployed environments.

| Variable | Description | Secret | Notes |
|---|---|---|---|
| `MOCKPLAID_PORT` | HTTP listen port | No | Default `8080` |
| `MOCKPLAID_REDIS_URL` | Redis connection URL | No | Set → Redis store; empty → in-memory. Local: `redis://redis:6379` |
| `MOCKPLAID_REDIS_DB` | Redis DB index | No | Default `6` |

> `cdn.plaid.com` is **not** an env var — it's an `/etc/hosts` redirect to mockplaid (added by `make hosts`), so the browser loads the mock Plaid Link SDK. Toggle with `make plaid-mock` / `make plaid-real`. See `documentation/poc/plaid/mockplaid.md`.

### Legacy Variables

| Variable | Description | Secret | Notes |
|---|---|---|---|
| `VAULT_ADDR` | HashiCorp Vault server address used by legacy Vault integration paths | No | Keep documented, no environment value guidance |
| `VAULT_TRANSIT_ENGINE_PATH` | Vault transit engine path used by legacy encryption/decryption flows | No | Keep documented, no environment value guidance |
| `SMARTY_AUTH_ID` | Smarty auth ID used by legacy address validation integration | No | Keep documented, no environment value guidance |
| `SMARTY_AUTH_TOKEN` | Smarty auth token used by legacy address validation integration | No | Keep documented, no environment value guidance |
| `RAFIKI_GRAPHQL_URL` | Legacy Rafiki GraphQL URL alias from older local compose manifests | No | Historic YAML-only variable; should not be used for active configuration |
| `XAGO_USD_OPS_ACCOUNT` | Legacy Xago USD operations account UUID from older local compose manifests | No | Historic YAML-only variable; should not be used for active configuration |
| `XAGO_ZAR_OPS_ACCOUNT` | Legacy Xago ZAR operations account UUID from older local compose manifests | No | Historic YAML-only variable; should not be used for active configuration |
| `XAGO_LEDGER_ID_ZAR` | Legacy Xago ZAR ledger ID from older local compose manifests | No | Historic YAML-only variable; should not be used for active configuration |
| `XAGO_LEDGER_ID_USD` | Legacy Xago USD ledger ID from older local compose manifests | No | Historic YAML-only variable; should not be used for active configuration |
| `XAGO_WEBHOOK_SECRET` | Legacy Xago webhook secret placeholder from older local compose manifests | Yes | Historic YAML-only variable; should not be used for active configuration |

> **Discord:** No Discord-related environment variables are currently present in the active service configuration.
