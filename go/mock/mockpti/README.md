# Mock PTI Service

> **Work in progress** — this service is under active development.

Mock PTI (`mockpti`) is a local stand-in for the PTI/Fiant provider API, used
during development and automated testing. It follows the same patterns as
`mockgatehub` and `mockxago`.

## Current status

Phase 1 (Foundations and Signup Core) is complete. Subsequent phases are tracked
in the [plan/](plan/) folder:

| Phase | Description | Status |
|-------|-------------|--------|
| [Phase 1](plan/phase-1-foundations-signup.md) | Foundations and Signup Core | ✅ Done |
| [Phase 2](plan/phase-2-wallets-payment-info.md) | Wallets and Payment Information | Not started |
| [Phase 3](plan/phase-3-transactions.md) | Transactions | Not started |
| [Phase 4](plan/phase-4-webhook-jobs.md) | Webhook Jobs and Delivery | Not started |
| [Phase 5](plan/phase-5-local-integration-sdk.md) | Local Integration and SDK | Not started |

See [plan/plan.md](plan/plan.md) for the full design and [plan/roadmap.md](plan/roadmap.md)
for sequencing details.

## Quick start

```bash
make build      # compile binary
make test       # lint + unit tests + e2e tests
make lint       # linters only
make unit-test  # unit tests only
```

## Webhook crypto settings CLI

`mockpti` includes helper CLI commands to generate local webhook JWK settings
for wallet/backend + mockpti, and to derive a public JWK from a private JWK.

Generate a full local key set:

```bash
go run ./cmd/mockpti gen-webhook-settings
```

This command prints ready-to-paste values for:

- `.env` variables (`BACKEND_PTI_JWK`, `BACKEND_PTI_PUBLIC_KEY_JWK`, `MOCKPTI_WEBHOOK_SIGNING_JWK`, `MOCKPTI_WEBHOOK_ENCRYPTION_JWK`)
- `local/wallet.yaml` backend env lines (`PTI_JWK`, `PTI_PUBLIC_KEY_JWK`)
- `local/mockpti.yaml` env lines (`MOCKPTI_WEBHOOK_SIGNING_JWK`, `MOCKPTI_WEBHOOK_ENCRYPTION_JWK`)

Derive public JWK from private JWK JSON on stdin:

```bash
cat private.jwk.json | go run ./cmd/mockpti derive-public-jwk
```

Notes:

- Webhook security is asymmetric, not a shared symmetric secret.
- `mockpti` signs webhook payloads with `MOCKPTI_WEBHOOK_SIGNING_JWK` and encrypts
	to the backend decrypt key (`MOCKPTI_WEBHOOK_ENCRYPTION_JWK`).
- Backend verifies signatures with `PTI_PUBLIC_KEY_JWK` and decrypts with
	`PTI_JWK`.
