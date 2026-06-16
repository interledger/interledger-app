# AGENTS.md — mockplaid

Guidance for AI agents working on the Plaid mock. Read alongside `README.md` and
`documentation/poc/plaid/mockplaid.md`.

## What this service is

A mock of Plaid covering **two surfaces**: the Plaid REST API (backend SDK target) and the
Plaid Link CDN/SDK (browser). Goal: run the wallet's Plaid → Fiant flow locally, offline,
deterministically, with **zero app-code branching** (only the network target changes).

## Architecture

```
cmd/mockplaid/main.go        server, store selection (memory|redis), route table
internal/config/             env (MOCKPLAID_PORT, LOG_LEVEL, MOCKPLAID_REDIS_URL/DB)
internal/logger/             zap (copied from mockpti)
internal/models/             models.go (LinkSession, Account, Item) + institutions.go (the 2 banks)
internal/storage/            interface + memory + redis + shared contract test
internal/handler/            handler.go (helpers) + link.go + item.go + processor.go + products.go + link_ui.go
web/                         link-initialize.js (window.Plaid shim) + link.html (dropdown UI), go:embed
testenv/ + features/         godog e2e (//go:build e2e)
```

## Core invariants — do NOT break

1. **Determinism.** `ins_mock_a` (Tartan) → stable `account_id` per account; `ins_mock_b`
   (Platypus) → fresh `acc_mock_b_<seq>` each select. The duplicate/multi-account tests depend
   on this. `account_id` is set at select time (`link.go`) and stored on the `Item`; every later
   endpoint reads it back — never regenerate it downstream.
2. **Plaid wire shapes.** REST responses must deserialize into `plaid-go/v42` SDK structs
   (snake_case JSON). Keep `request_id` on every response; errors use Plaid's
   `{error_type,error_code,error_message,display_message,request_id}` envelope (`sendPlaidError`).
3. **Never log secrets.** `logCreds` logs header *presence* only — never the `PLAID-SECRET` value
   or any token value beyond the mock link/public tokens.
4. **CDN shim path is exact.** `GET /link/v2/stable/link-initialize.js` — react-plaid-link
   hardcodes it. The shim's postMessage bridge validates `event.origin === https://cdn.plaid.com`.

## Adding a Plaid endpoint

1. Handler in the matching `internal/handler/*.go` (resolve item via `requireItem` where access-token-scoped).
2. Swap the `h.NotImplemented` stub → real handler in `cmd/mockplaid/main.go` `setupRoutes`.
3. Match the `plaid-go/v42` response struct exactly (check the backend caller in
   `go/backend/providers/plaid/client/client.go`).
4. Unit test in `internal/handler/*_test.go`; extend godog if it's part of the link flow.

## Testing

```bash
make unit-test   # includes storage contract on memory + miniredis
make e2e-test    # godog; in-memory mockplaid, /test/reset between scenarios
make lint        # golangci-lint v2.5.0 (gofmt + vet + linters)
```
After ANY code change, rebuild the compose container before browser/curl checks:
```bash
cd local && docker compose up -d --build mockplaid
```
(Compose images do not auto-rebuild on source change — stale container = 404 on new routes.)

## Gotchas

- **HSTS**: `cdn.plaid.com` is preloaded → cert must be trusted (`make hosts/certs/trust`), no bypass.
- **Processor token**: mockplaid mints it; it is NOT validated against real Plaid anywhere. The real
  Fiant *does* validate → local PTI must point at **mockpti**, not real Fiant, for link-to-fiant to pass.
- **`interface{}` vs `any`**: gopls hints `any`; golangci (the CI linter) does not enforce it — match
  the existing `interface{}` style; `make lint` is the source of truth.
- **`unused` linter**: don't add an unexported helper before its first caller exists — it trips
  `unused`. Add it in the same change that uses it.

## Key reference files
- Backend caller (must match shapes): `go/backend/providers/plaid/client/client.go`
- Plan/tasks: `documentation/poc/mockplaid/tasks.md`
- How-it-works: `documentation/poc/plaid/mockplaid.md`
