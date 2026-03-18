# Environment Variables Reference

This document covers all environment variables for the three deployed application services:

- **Protea** — user-facing frontend (`typescript/protea`)
- **Botanist** — admin portal (`typescript/botanist`)
- **Backend** — Go wallet backend (`go/backend`)

Variables are grouped by secrecy:

- **Secret** — must be stored in 1Password and injected via a Kubernetes secret; never committed to config files.
- **Non-secret** — safe to store in a ConfigMap / compose file.

Environment column values are the effective values after all Helm value layers resolve (base + environment override). `SECRET` means the value comes from 1Password and is not documented here. `(not set)` means the variable is absent in local compose and the code falls back to a built-in default.

---

## Protea (Frontend)

Protea is a Remix application serving the user-facing wallet UI.

### Non-secret variables

| Variable | Description | Secret? | Production | Sandbox | Development | Local |
|---|---|---|---|---|---|---|
| `FYNBOS_ENV` | Runtime environment tag used for feature flags and logging context | No | `prod` | `dev` | `dev` | `local` |
| `LOG_LEVEL` | Log verbosity (`debug`, `info`, `warn`, `error`) | No | `info` | `info` | `info` | `debug` |
| `LOG_PRETTY` | Human-readable log formatting. Set `false` for JSON in deployed environments | No | `false` | `false` | `false` | `true` |
| `KRATOS_URL` | Internal URL for the Ory Kratos public API (session/login flows) | No | `http://kratos-public` | `http://kratos-public` | `http://kratos-public` | `http://kratos:4433` |
| `RAFIKI_AUTH_ENDPOINT` | Internal URL for the Rafiki auth gRPC/GraphQL endpoint | No | `http://rafiki-auth-service:3009` | `http://rafiki-auth-service:3009` | `http://rafiki-auth-service:3009` | `http://rafiki_auth:3009` |
| `PAYMENT_POINTER_BASE` | Domain used to build Open Payments payment pointer addresses | No | `ilp.link` | `sandbox.ilp.link` | `development.ilp.link` | `local.ilp.link` |
| `BACKEND_GRPC_URL` | Internal URL for the wallet backend gRPC server | No | `http://wallet-backend-service-grpc:8443` | `http://wallet-backend-service-grpc:8443` | `http://wallet-backend-service-grpc:8443` | `http://backend:8443` |
| `GATE_SIGNUP` | When `true`, new user registrations are gated/disabled | No | TODO | TODO | TODO | `false` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OpenTelemetry collector endpoint for traces | No | `grpc://api.honeycomb.io:443` | `grpc://api.honeycomb.io:443` | `grpc://api.honeycomb.io:443` | `grpc://api.honeycomb.io:443` |
| `OTEL_SERVICE_NAME` | Service name reported in OpenTelemetry traces | No | `protea` | `protea` | `protea` | `protea` |
| `DEFAULT_RATE_LIMIT_REQUESTS` | Max requests allowed per time window before rate limiting kicks in | No | TODO | TODO | TODO | `4` (code default) |
| `DEFAULT_RATE_LIMIT_TIME` | Rate limit time window in seconds | No | TODO | TODO | TODO | `3600` (code default) |
| `PTI_CLIENT_ID` | PTI/Fiant payment provider client UUID, passed to the browser for payment widget initialisation | No | TODO | TODO | TODO | `''` (code default) |
| `SENTRY_RELEASE` | Identifies the deployed version in Sentry error reports | No | TODO | TODO | TODO | `''` (code default) |
| `CHOKIDAR_USEPOLLING` | Enable filesystem polling for hot-reload in containers (dev only) | No | N/A | N/A | N/A | `true` |

### Secret variables

| Variable | Description | Secret? | Production | Sandbox | Development | Local |
|---|---|---|---|---|---|---|
| `COOKIE_SECRETS` | JSON array of strings used to sign session cookies. Rotate periodically. | **Yes** | SECRET | SECRET | SECRET | `["localsecret"]` |
| `RAFIKI_AUTH_SECRET` | Shared secret between Protea and the Rafiki auth service | **Yes** | SECRET | SECRET | SECRET | `my-super-secret-identity-key` |
| `PUSHER_APP_KEY` | Pusher application key for real-time push notifications (public-facing) | **Yes** | SECRET | SECRET | SECRET | `91988d6075551d29760a` |
| `PUSHER_APP_CLUSTER` | Pusher cluster region for the application | **Yes** | SECRET | SECRET | SECRET | `eu` |
| `REDIS_URL` | Redis connection URL used for session storage and caching | **Yes** | SECRET | SECRET | SECRET | `redis://redis:6379/2` |
| `SENTRY_DSN` | Sentry DSN for client-side and server-side error reporting | **Yes** | SECRET | SECRET | SECRET | `(not set)` |
| `SEGMENT_API_KEY` | Segment analytics write key for event tracking | **Yes** | SECRET | SECRET | SECRET | `(not set)` |
| `BT_TOKEN` | Basis Theory token for card tokenisation flows in the frontend | **Yes** | SECRET | SECRET | SECRET | `(not set)` |
| `OTEL_EXPORTER_OTLP_HEADERS` | Auth headers for the OTLP endpoint (e.g. `x-honeycomb-team=<key>`) | **Yes** | SECRET | SECRET | SECRET | `(not set)` |

### Legacy Variables

| Variable | Description | Secret? | Notes |
|---|---|---|---|
| `DATO_API_TOKEN` | DatoCMS API token used by legacy CMS content loading paths | **Yes** | Keep documented, no environment value guidance |
| `CF_TURNSTILE_SITE_KEY` | Cloudflare Turnstile site key used in legacy bot-protection integration | **Yes** | Keep documented, no environment value guidance |
| `CF_TURNSTILE_SECRET_KEY` | Cloudflare Turnstile secret key used in legacy server verification | **Yes** | Keep documented, no environment value guidance |
| `GOOGLE_MAPS_API_KEY` | Google Maps API key used in legacy geocode/autocomplete endpoints | **Yes** | Keep documented, no environment value guidance |

---

## Botanist (Admin Portal)

Botanist is a Remix application providing the internal admin interface. It connects to the backend admin gRPC port.

### Non-secret variables

| Variable | Description | Secret? | Production | Sandbox | Development | Local |
|---|---|---|---|---|---|---|
| `FYNBOS_ENV` | Runtime environment tag | No | `prod` | `dev` | `dev` | `local` |
| `BACKEND_GRPC_URL` | Internal URL for the backend admin gRPC port (8448) | No | `wallet-backend-service-grpc:8448` | `wallet-backend-service-grpc:8448` | `wallet-backend-service-grpc:8448` | `backend:8448` |
| `KRATOS_ADMIN_URL` | Internal URL for the Ory Kratos Admin API | No | `http://kratos-admin:4434` | `http://kratos-admin:4434` | `http://kratos-admin:4434` | `http://kratos:4434` |
| `PAYMENT_POINTER_BASE` | Domain used to display payment pointer addresses for users | No | `ilp.link` | `sandbox.ilp.link` | `development.ilp.link` | `(not set)` |

> **Note:** Botanist has no secret variables — all sensitive operations are delegated to the backend. The admin gRPC port must only be accessible from within the cluster.

### Legacy Variables

| Variable | Description | Secret? | Notes |
|---|---|---|---|
| `(none)` | No Botanist-specific legacy variables from the current legacy list | N/A | Includes Discord/DatoCMS/Turnstile/Google Maps/Vault/Smarty/Astra/Twitter scope |

---

## Backend (Wallet)

The Go backend is the core of the wallet, handling payments, provider integrations, webhooks, Temporal workflows, and gRPC APIs.

### General & Observability

| Variable | Description | Secret? | Production | Sandbox | Development | Local |
|---|---|---|---|---|---|---|
| `FYNBOS_ENV` | Runtime environment tag | No | `prod` | `dev` | `dev` | `local` |
| `LOG_LEVEL` | Log verbosity | No | `info` | `info` | `debug` | `info` |
| `PORT` | HTTP server port (webhooks, health) | No | `8080` | `8080` | `8080` | `8080` |
| `OPEN_PAYMENTS_PORT` | Open Payments protocol server port | No | `8081` | `8081` | `8081` | `8081` |
| `AUTHORISATION_PORT` | Authorisation server port | No | `8082` | `8082` | `8082` | `8082` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OpenTelemetry collector endpoint | No | `grpc://api.honeycomb.io:443` | `grpc://api.honeycomb.io:443` | `grpc://api.honeycomb.io:443` | `grpc://api.honeycomb.io:443` |
| `OTEL_SERVICE_NAME` | Service name in traces | No | `backend` | `backend` | `backend` | `backend` |
| `OTEL_EXPORTER_OTLP_HEADERS` | Auth headers for the OTLP endpoint | **Yes** | SECRET | SECRET | SECRET | `x-honeycomb-team=7Qskhns7Dc7wgazrDe6yZD` |
| `SENTRY_DSN` | Sentry DSN for error reporting | **Yes** | SECRET | SECRET | SECRET | `(not set)` |
| `SENTRY_RELEASE` | Sentry release identifier | No | TODO | TODO | TODO | `(not set)` |

### Database

| Variable | Description | Secret? | Production | Sandbox | Development | Local |
|---|---|---|---|---|---|---|
| `DB_URL` | Wallet backend PostgreSQL connection URL | **Yes** | SECRET | SECRET | SECRET | `postgres://postgres:postgres@postgres:5432/backend?sslmode=disable` |
| `DB_URL_WITH_CERTS` | Backend DB URL with TLS certificates (IAM auth in GCP) | **Yes** | SECRET | SECRET | SECRET | Same as `DB_URL` |
| `PACIOLI_DB_URL` | Pacioli ledger PostgreSQL connection URL | **Yes** | SECRET | SECRET | SECRET | `postgres://postgres:postgres@postgres:5432/pacioli?sslmode=disable` |
| `PACIOLI_DB_URL_WITH_CERTS` | Pacioli DB URL with TLS certificates | **Yes** | SECRET | SECRET | SECRET | Same as `PACIOLI_DB_URL` |
| `RAFIKI_DB_URL` | Rafiki backend PostgreSQL connection URL | **Yes** | SECRET | SECRET | SECRET | `postgres://postgres:postgres@postgres:5432/rafiki_backend?sslmode=disable` |
| `RAFIKI_AUTH_DB_URL` | Rafiki auth PostgreSQL connection URL | **Yes** | SECRET | SECRET | SECRET | `postgres://postgres:postgres@postgres:5432/rafiki_auth?sslmode=disable` |

### Ledger & Wallet Core

| Variable | Description | Secret? | Production | Sandbox | Development | Local |
|---|---|---|---|---|---|---|
| `USD_LEDGER_ID` | Internal Pacioli ledger ID for USD accounts | No | `1` | `1` | `1` | `1` |
| `NOOP_EQUITY_ACCOUNT_ID` | Equity account UUID used as no-op placeholder in ledger entries | No | `00000000-0000-0000-0000-000000000000` | `00000000-0000-0000-0000-000000000000` | `00000000-0000-0000-0000-000000000000` | `43d4b2bd-e29b-4a63-9aa8-7990776c714e` |
| `REDIS_URL` | Redis connection URL for queues and caching | **Yes** | SECRET | SECRET | SECRET | `(not set)` |
| `ALLOWED_WALLET_IDS` | Comma-separated wallet IDs allowed through regional blocks | No | `(empty)` | `(empty)` | `(empty)` | `(empty)` |
| `BLOCKED_REGIONS` | Comma-separated ISO country codes to block wallet access from | No | `US` | `(empty)` | `(empty)` | `(empty)` |
| `APPLE_APP_ID` | Apple App Site Association app ID for deep-links | No | `6B7AFCRT3V.app.wallet.interledger` | TODO | `6B7AFCRT3V.app.wallet.interledger.dev` | `6B7AFCRT3V.app.wallet.interledger.test` |
| `ANDROID_PACKAGE_NAME` | Android App Links package name for deep-links | No | `org.interledger.walletandroid` | TODO | `org.interledger.walletandroid.debugDevelopment` | `org.interledger.walletandroid.test` |
| `ANDROID_SHA256` | Android signing certificate SHA-256 fingerprint for deep-links | No | `25:D0:6A:A1:4F:BD:F6:05:93:2C:80:A2:AC:DA:DE:E4:2F:72:BC:23:9C:8F:2C:78:85:49:9A:AC:D2:4B:A4:FF` | TODO | `25:D0:6A:A1:...` | `25:D0:6A:A1:...` |

### Infrastructure Services

| Variable | Description | Secret? | Production | Sandbox | Development | Local |
|---|---|---|---|---|---|---|
| `TEMPORAL_URL` | Temporal frontend address for workflow orchestration | No | `temporal-frontend-headless:7233` | `temporal-frontend-headless:7233` | `temporal-frontend-headless:7233` | `temporal:7233` |
| `KRATOS_URL` | Kratos public API URL for session validation | No | `http://kratos-public` | `http://kratos-public` | `http://kratos-public` | `http://kratos:4433` |
| `KRATOS_ADMIN_URL` | Kratos admin API URL for user management | No | `http://kratos-admin` | `http://kratos-admin` | `http://kratos-admin` | `http://kratos:4434` |
| `RAFIKI_BACKEND_GRAPHQL_URL` | Rafiki backend GraphQL endpoint | No | `http://rafiki-backend-service:3001/graphql` | `http://rafiki-backend-service:3001/graphql` | `http://rafiki-backend-service:3001/graphql` | `http://rafiki_backend:3001/graphql` |
| `RAFIKI_AUTH_GRAPHQL_URL` | Rafiki auth GraphQL endpoint | No | `http://rafiki-auth-service:3003/graphql` | `http://rafiki-auth-service:3003/graphql` | `http://rafiki-auth-service:3003/graphql` | `http://rafiki_auth:3003/graphql` |
| `PUSHER_ADDR` | Internal URL for the Pusher/Soketi service used for push notifications | **Yes** | SECRET | SECRET | SECRET | `http://pusher:8080` |

### Admin Portal

| Variable | Description | Secret? | Production | Sandbox | Development | Local |
|---|---|---|---|---|---|---|
| `ADMIN_POLICY_AUD` | Cloudflare Access policy audience for the admin portal | **Yes** | SECRET | SECRET | SECRET | `admin-policy` |
| `ADMIN_TEAM_DOMAIN` | Cloudflare Access team domain for the admin portal | **Yes** | SECRET | SECRET | SECRET | `admin-team` |

### Twilio (Phone Verification)

| Variable | Description | Secret? | Production | Sandbox | Development | Local |
|---|---|---|---|---|---|---|
| `TWILIO_ACCOUNT_SID` | Twilio account SID | **Yes** | SECRET | SECRET | SECRET | `SK021f793191208ba69c3bea87dd426085` |
| `TWILIO_SERVICE_SID` | Twilio Verify service SID | **Yes** | SECRET | SECRET | SECRET | `VA8af4e130da63b9fac4c042acbc33a267` |
| `TWILIO_ACCOUNT_TOKEN` | Twilio auth token | **Yes** | SECRET | SECRET | SECRET | `test` |

### SendGrid (Transactional Email)

| Variable | Description | Secret? | Production | Sandbox | Development | Local |
|---|---|---|---|---|---|---|
| `SENDGRID_API_KEY` | SendGrid API key for sending emails | **Yes** | SECRET | SECRET | SECRET | `test-sendgrid-api-key` |
| `ZENDESK_USER` | Zendesk account email (currently not actively used) | No | `support@interledger-app.dev` | `support@interledger-app.dev` | `support@interledger-app.dev` | `matt@fynbos.dev` |
| `ZENDESK_TOKEN` | Zendesk API token | **Yes** | SECRET | SECRET | SECRET | `test` |

### Segment (Analytics)

| Variable | Description | Secret? | Production | Sandbox | Development | Local |
|---|---|---|---|---|---|---|
| `SEGMENT_KEY` | Segment write key for server-side event tracking | **Yes** | SECRET | SECRET | SECRET | `test-segment-key` |

### Sentry (Error Tracking)

| Variable | Description | Secret? | Production | Sandbox | Development | Local |
|---|---|---|---|---|---|---|
| `SENTRY_DSN` | Sentry DSN for server-side error reporting | **Yes** | SECRET | SECRET | SECRET | `(not set)` |

### Persona (Identity Verification)

| Variable | Description | Secret? | Production | Sandbox | Development | Local |
|---|---|---|---|---|---|---|
| `PERSONA_TOKEN` | Persona API token for identity verification | **Yes** | SECRET | SECRET | SECRET | `(not set)` |
| `PERSONA_WEBHOOK_TOKEN` | Persona webhook verification token | **Yes** | SECRET | SECRET | SECRET | `(not set)` |

### Basis Theory (Card Tokenisation)

| Variable | Description | Secret? | Production | Sandbox | Development | Local |
|---|---|---|---|---|---|---|
| `BASIS_THEORY_API_KEY` | Basis Theory private API key (server-side) | **Yes** | SECRET | SECRET | SECRET | `(not set)` |
| `BT_TOKEN` | Basis Theory session token for frontend card operations | **Yes** | SECRET | SECRET | SECRET | `(not set)` |

### CDN

| Variable | Description | Secret? | Production | Sandbox | Development | Local |
|---|---|---|---|---|---|---|
| `CDN_KEY` | CDN authentication key for signed asset URLs | **Yes** | SECRET | SECRET | SECRET | `(not set)` |

### Slack (Notifications)

| Variable | Description | Secret? | Production | Sandbox | Development | Local |
|---|---|---|---|---|---|---|
| `SLACK_TOKEN` | Slack bot token for sending notifications | **Yes** | SECRET | SECRET | SECRET | `(not set)` |
| `SLACK_CLIENT_ID` | Slack OAuth app client ID | **Yes** | SECRET | SECRET | SECRET | `(not set)` |
| `SLACK_CLIENT_SECRET` | Slack OAuth app client secret | **Yes** | SECRET | SECRET | SECRET | `(not set)` |
| `SLACK_SIGNING_SECRET` | Slack signing secret for webhook verification | **Yes** | SECRET | SECRET | SECRET | `(not set)` |

### Chimoney (Payment Provider)

| Variable | Description | Secret? | Production | Sandbox | Development | Local |
|---|---|---|---|---|---|---|
| `CHIMONEY_TOKEN` | Chimoney API token | **Yes** | SECRET | SECRET | SECRET | `(not set)` |
| `CHIMONEY_WEBHOOK_SECRET` | Chimoney webhook signature secret | **Yes** | SECRET | SECRET | SECRET | `(not set)` |

### GateHub (Payment Provider)

GateHub is used for EUR XRPL transactions, onboarding, and debit card issuing.

> **Sandbox/Development note:** The Sandbox environment is missing explicit Gatehub API URL configuration in its values file. It is expected to inherit from the development environment base. Verify this is intentional before promoting changes.

#### Non-secret GateHub variables

| Variable | Description | Secret? | Production | Sandbox | Development | Local |
|---|---|---|---|---|---|---|
| `GATEHUB_API_BASE_URL` | Base URL for GateHub API calls | No | `https://api.gatehub.net` | `https://api.sandbox.gatehub.net` | `https://api.sandbox.gatehub.net` | `http://mockgatehub:8080` |
| `GATEHUB_ONBOARDING_BASE_URL` | Frontend widget URL for GateHub KYC onboarding | No | `https://onboarding.gatehub.net` | `https://onboarding.sandbox.gatehub.net` | `https://onboarding.sandbox.gatehub.net` | `https://mockgatehub.interledger.test/iframe/onboarding` |
| `GATEHUB_ON_OFF_RAMP_BASE_URL` | Frontend widget URL for GateHub deposit/withdrawal | No | `https://managed-ramp.gatehub.net` | `https://managed-ramp.sandbox.gatehub.net` | `https://managed-ramp.sandbox.gatehub.net` | `https://mockgatehub.interledger.test` |
| `GATEHUB_ON_OFF_RAMP_CLIENT_ID` | OAuth client ID for GateHub on/off-ramp token issuance | No | `f4c8f30f-7fc3-4aa1-8573-520cb67565e3` | `f8119dfd-e563-44ee-9ae2-1e60a4fce74f` | `f8119dfd-e563-44ee-9ae2-1e60a4fce74f` | `f8119dfd-e563-44ee-9ae2-1e60a4fce74f` |
| `GATEHUB_ONBOARDING_CLIENT_ID` | OAuth client ID for GateHub onboarding token issuance | No | `40a22fc5-9091-4c6f-aff6-a3fddf475b33` | `4df24d1b-5796-4eec-951b-21699d61b970` | `4df24d1b-5796-4eec-951b-21699d61b970` | `4df24d1b-5796-4eec-951b-21699d61b970` |
| `GATEHUB_EXCHANGE_CLIENT_ID` | OAuth client ID for GateHub exchange token issuance | No | `50e7c590-f6f9-4fa9-9498-260bd978c5d6` | `4e28d4df-22d7-414c-97a3-d71956df29ba` | `4e28d4df-22d7-414c-97a3-d71956df29ba` | `4e28d4df-22d7-414c-97a3-d71956df29ba` |
| `GATEHUB_FALLBACK_WEBHOOK_URL` | Secondary webhook URL for unrecognised GateHub users (card fallback) | No | `https://api.interledger.cards/gatehub-webhooks` | `(empty)` | `(empty)` | `(not set)` |
| `GATEHUB_CARD_ACCOUNT_PRODUCT_CODE` | Card product code applied when creating GateHub card accounts | **Yes** | SECRET | SECRET | SECRET | `PWSR_DEBP_2404` |
| `GATEHUB_ORGANIZATION_ID` | GateHub organisation UUID for org-level operations | No | TODO | `a5b3de06-3c61-47aa-82fa-9d9e1875f42c` | `a5b3de06-3c61-47aa-82fa-9d9e1875f42c` | `default-org` |
| `GATEHUB_EUR_OPS_ACCOUNT` | Internal EUR operations account ID for ledger movements | No | `1854f171-eafa-4e30-bf66-7dbfe167ccfa` | `1854f171-eafa-4e30-bf66-7dbfe167ccfa` | `1854f171-eafa-4e30-bf66-7dbfe167ccfa` | `1854f171-eafa-4e30-bf66-7dbfe167ccfa` |
| `GATEHUB_EUR_OPS_LEDGER_ID` | Internal EUR ledger ID for GateHub reserve/assign posting | No | `4482387` | `4482387` | `4482387` | `4482387` |

#### Secret GateHub variables

| Variable | Description | Secret? | Production | Sandbox | Development | Local |
|---|---|---|---|---|---|---|
| `GATEHUB_APP_ID` | GateHub API app ID header value | **Yes** | SECRET | SECRET | SECRET | `local-test-app-id` |
| `GATEHUB_SECRET` | GateHub API signing secret | **Yes** | SECRET | SECRET | SECRET | `local-test-app-secret` |
| `GATEHUB_WEBHOOK_SECRET` | Hex-encoded webhook signing secret for verifying GateHub webhooks | **Yes** | SECRET | SECRET | SECRET | `6d6f636b5f776562686f6f6b5f736563726574` |
| `GATEHUB_GATEWAY_ID` | GateHub gateway UUID for linking users to a hub | **Yes** | SECRET | SECRET | SECRET | `b18965b5-8bb0-486e-88b2-7be47b974ac1` |
| `GATEHUB_PAYWISER_EURO_VAULT_ID` | GateHub EUR vault UUID used in hosted transactions | **Yes** | SECRET | SECRET | SECRET | `a09a0a2c-1a3a-44c5-a1b9-603a6eea9341` |
| `GATEHUB_CARD_APP_ID` | GateHub card app ID header value (separate from main app) | **Yes** | SECRET | SECRET | SECRET | `local-test-card-app-id` |
| `GATEHUB_SENDING_USER_ID` | Managed GateHub user UUID used for backfill payment operations | **Yes** | SECRET | SECRET (uses dev default) | `5dd8e7f7-ead5-4f8d-b956-1acfecc231b4` | `test-sending-user-id` |
| `GATEHUB_SENDING_USER_ADDRESS` | XRPL wallet address of the managed sending user | **Yes** | SECRET | SECRET (uses dev default) | `191968355` | `rN7n7otQDd6FczFgLdVZGMbpKRtFVfT4hb` |

### Xago (Payment Provider)

Xago is used for ZAR and USD transactions in South Africa.

| Variable | Description | Secret? | Production | Sandbox | Development | Local |
|---|---|---|---|---|---|---|
| `XAGO_API_BASE_URL` | Xago business API base URL | No | TODO | `https://test.xago.io/exchange/v1` | `https://test.xago.io/exchange/v1` | `http://mockxago:8080/v1` |
| `XAGO_IDENTITY_BASE_URL` | Xago identity/login API base URL | No | TODO | `https://test.xago.io/identity/v1` | `https://test.xago.io/identity/v1` | `http://mockxago:8080/v1` |
| `XAGO_POLICY_ID` | Xago login policy ID for requesting access tokens | No | TODO | `5e2585a474b0e90012ce8ff1` | `5e2585a474b0e90012ce8ff1` | `5e2585a474b0e90012ce8ff1` |
| `XAGO_API_SECRET` | Xago API secret key credential | **Yes** | SECRET | SECRET | SECRET | `test-secret` |
| `XAGO_API_PUBLIC_KEY` | Xago API public key credential | **Yes** | SECRET | SECRET | SECRET | `test-public-key` |

### PTI / Fiant (Payment Provider)

PTI is a US-focused payment provider currently disabled except in production.

| Variable | Description | Secret? | Production | Sandbox | Development | Local |
|---|---|---|---|---|---|---|
| `PTI_ENABLED` | Enables PTI provider integration and runtime config validation | No | `true` | `false` | `false` | `false` |
| `PTI_BASE_URL` | PTI API base URL | **Yes** | SECRET | SECRET | `(empty)` | `https://api.staging.fiant.io/v1/` |
| `PTI_CLIENT_ID` | PTI client UUID for API and webhook validation | **Yes** | SECRET | SECRET | SECRET | `04d3e1b5-96d4-47e4-9eaa-13e9b4b0f219` |
| `PTI_JWK` | PTI private RSA JWK used for request signing and webhook crypto | **Yes** | SECRET | SECRET | SECRET | *(test RSA key in local compose)* |
| `PTI_PUBLIC_KEY_JWK` | PTI public RSA JWK used for webhook signature verification | **Yes** | SECRET | SECRET | SECRET | *(test RSA public key in local compose)* |

### Legacy Variables

| Variable | Description | Secret? | Notes |
|---|---|---|---|
| `VAULT_ADDR` | HashiCorp Vault server address used by legacy Vault integration paths | No | Keep documented, no environment value guidance |
| `VAULT_TRANSIT_ENGINE_PATH` | Vault transit engine path used by legacy encryption/decryption flows | No | Keep documented, no environment value guidance |
| `SMARTY_AUTH_ID` | Smarty auth ID used by legacy address validation integration | No | Keep documented, no environment value guidance |
| `SMARTY_AUTH_TOKEN` | Smarty auth token used by legacy address validation integration | No | Keep documented, no environment value guidance |
| `ASTRA_CLIENT_ID` | Astra client ID used by legacy Astra integration paths | **Yes** | Keep documented, no environment value guidance |
| `ASTRA_CLIENT_SECRET` | Astra client secret used by legacy Astra integration paths | **Yes** | Keep documented, no environment value guidance |
| `ASTRA_WEBHOOK_BEARER_TOKEN` | Astra webhook bearer token used by legacy webhook verification paths | **Yes** | Keep documented, no environment value guidance |
| `TWITTER_CLIENT_ID` | Twitter/X OAuth client ID used by legacy social login integration | No | Keep documented, no environment value guidance |
| `TWITTER_CLIENT_SECRET` | Twitter/X OAuth client secret used by legacy social login integration | No | Keep documented, no environment value guidance |
| `TWITTER_BEARER_TOKEN` | Twitter/X API bearer token used by legacy integration paths | No | Keep documented, no environment value guidance |
| `TWITTER_REDIRECT_URL` | Twitter/X OAuth redirect URL used by legacy callback handling | No | Keep documented, no environment value guidance |

> **Discord:** No Discord-related environment variables are currently present in the active service configuration.

### Google OAuth

| Variable | Description | Secret? | Production | Sandbox | Development | Local |
|---|---|---|---|---|---|---|
| `GOOGLE_OAUTH2_CLIENT_ID` | Google OAuth 2.0 client ID (currently not actively used) | No | `google_oauth` | `google_oauth` | `google_oauth` | `google_oauth` |

---

## Reviewing Environment Variable Changes

When reviewing pull requests that touch service configurations, verify the following:

1. **Secrets are never committed** — any new secret must reference a 1Password vault item via a Kubernetes secret.
2. **Local compose defaults are safe** — local defaults must be fake/mock values, never real credentials.
3. **All environments are accounted for** — if a new variable is added, update `base/values` and all environment `values/` overrides in `interledger-app-deploy`.
4. **Missing local defaults** — if a new variable has no meaningful local default, document it in [env-cleanup.txt](../../env-cleanup.txt) and consider whether a mock value should be added.
5. **Production-specific values differ from sandbox** — especially for GateHub client IDs and API base URLs.
