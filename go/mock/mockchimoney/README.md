# MockChimoney

> **Status**: This service is implemented and used for local development and testing. See [`plan/plan.md`](plan/plan.md) for the detailed design and planned enhancements.

MockChimoney is a lightweight Go mock of the Chimoney API for local development and testing of the Interledger Wallet. It follows the same patterns as `mockgatehub` and `mockxago`: go-chi router, Redis-backed storage, background webhook delivery, and Traefik-fronted Docker Compose service.

Official Chimoney API documentation: <https://chimoney.readme.io/reference/introduction>

---

## Overview

MockChimoney replaces the Chimoney sandbox for local development, enabling:

- Creating and retrieving multicurrency wallets (sub-accounts)
- Initiating deposit payment links and simulating Interac funding via a browser pay page
- Executing Interac withdrawal payouts
- KYC widget simulation with approve/decline controls
- Signed webhook delivery (svix-style HMAC-SHA256) matching the exact format the backend verifies
- Fee estimation and CAD→USD currency conversion at a configurable fixed rate

## API Endpoints

### Multicurrency Wallets

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v0.2.4/multicurrency-wallets/create` | Create a new sub-account |
| `GET` | `/v0.2.4/multicurrency-wallets/get?id=<id>` | Retrieve sub-account by ID |
| `POST` | `/v0.2.4/multicurrency-wallets/transfer` | Transfer between sub-accounts |

### Payments (Deposits)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v0.2.4/payment/initiate` | Create a deposit payment link |
| `POST` | `/v0.2.4/payment/verify` | Check payment status by issueID |
| `GET` | `/pay/<issueID>` | Browser-facing deposit confirmation page |

### Payouts (Withdrawals)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v0.2.4/payouts/interac` | Initiate Interac withdrawal |
| `POST` | `/v0.2.4/payouts/status` | Check payout status by chiRef |

### Info

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v0.2.4/info/fee-estimate` | Estimate transaction fee |
| `GET` | `/v0.2.4/info/convert/local-amount-to-usd` | Convert currency amount to USD |

### KYC

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/verify/kyc/<externalID>?redirect=<url>` | Browser-facing KYC widget |

### Utility

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |

## Webhook Events

The mock delivers the following events (signed with svix-style HMAC-SHA256) to the configured `WEBHOOK_URL`:

| Event | Trigger |
|-------|---------|
| `charge.interac.completed` | User completes payment on the pay page |
| `chimoney.redeem.completed` | User completes payment on the pay page (second webhook) |
| `chimoney.redeem.failed` | Payment explicitly marked failed |
| `payout.interac.completed` | Interac withdrawal succeeds (default) |
| `payout.interac.expired` | Interac withdrawal expires (configurable) |
| `payout.interac.cancelled` | Interac withdrawal cancelled (configurable) |
| `user.kyc.completed` | User approves KYC on the widget page |
| `user.kyc.declined` | User declines KYC on the widget page |

> **Note**: The official Chimoney webhook documentation at <https://chimoney.readme.io/reference/webhooks-and-events> only lists a small subset of these events (`chimoney.redeem.completed`, `payout.bank.*`, etc.). The events above were discovered from the backend's `ops/webhooks.go` switch statement and are **not publicly documented**. They are only visible in the Chimoney sandbox dashboard.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `MOCKCHIMONEY_PORT` | `8080` | HTTP listen port |
| `MOCKCHIMONEY_API_KEY` | `local-test-api-key` | Expected `X-API-KEY` header value |
| `MOCKCHIMONEY_ENFORCE_AUTHENTICATION` | `true` | Whether to validate `X-API-KEY` |
| `MOCKCHIMONEY_REDIS_URL` | `""` | Redis URL; empty = in-memory storage |
| `MOCKCHIMONEY_REDIS_DB` | `5` | Redis database number |
| `CHIMONEY_WEBHOOK_SECRET` | `local_bG9jYWwtdGVzdC13ZWJob29rLXNlY3JldA==` | Svix webhook signing secret (same format as backend expects) |
| `WEBHOOK_URL` | `http://backend:8080/webhooks/chimoney` | Backend webhook endpoint |
| `WEBHOOK_MIN_DELAY_SEC` | `0.5` | Seconds before deposit/withdrawal webhooks fire |
| `INTERAC_FEE_FLAT` | `1.50` | Flat CAD fee for Interac payouts |
| `CAD_TO_USD_RATE` | `0.735` | Exchange rate for currency conversion |
| `LOG_LEVEL` | `info` | Zap log level |

## Known Nuances and Discrepancies

### `payerEmail` is required on `/payment/initiate`

The [official docs](https://chimoney.readme.io/reference/post_v0-2-4-payment-initiate) mark `payerEmail` as **required**. The mock enforces this and returns `400` if it is absent. The backend always populates this field from the user record, so this validation should never be triggered in normal operation — but the mock enforces correctness.

### Supported currencies on `/payment/initiate`

The official docs specify a `currency` enum of **`USD`**, **`NGN`**, **`CAD`** only. The mock validates against this list and returns `400` for any other value. The backend currently only sends `CAD`, but the mock enforces the full documented constraint.

### The `/info/fee-estimate` endpoint is newly added and not on production

Confirmed by the Chimoney backend developer (Adi): `/info/fee-estimate` was added recently and is **not yet deployed to the Chimoney production API**. The backend has a fallback formula (`$1.00 + 0.5% × amount`) when this endpoint returns an error. The mock implements the endpoint regardless — the backend tries it first before falling back.

### Webhook event list in official docs is outdated

The public docs at `webhooks-and-events` only list a handful of event types (`chimoney.redeem.completed`, `payout.bank.*`, `payout.gift-card.*`). The events the backend actually handles (`charge.interac.completed`, `payout.interac.*`, `user.kyc.*`, etc.) are **not listed in public documentation**. They were found by reading `go/backend/providers/chimoney/ops/webhooks.go` directly.

### KYC API vs. hardcoded URL

The Chimoney API has a `GET /sub-account/kyc/link` endpoint that returns a hosted KYC page URL. The backend does **not** call this — it constructs the KYC URL directly by concatenating `CHIMONEY_KYC_BASE_URL + /verify/kyc/{externalID}?redirect={callbackURL}`. The mock implements the `/verify/kyc/<externalID>` path directly; the `/sub-account/kyc/link` endpoint is not implemented.

### `issueID` format encodes the sub-account ID

The backend extracts the chimoney sub-account ID from an `issueID` by splitting on `_` and taking the first part: `ExtractChiWalletIDFromIssueID`. This means the mock **must** generate `issueID` values in the format `{chiWalletID}_{uuid}` — otherwise the backend cannot resolve which sub-account a webhook belongs to.

### Webhook structure: `data` wrapper

The official webhook docs show events wrapped in a `data` key: `{ "eventType": "...", "status": "...", "data": { ... } }`. The backend's webhook handler parses fields at the top level of the JSON body (it binds `issueID`, `status`, `meta` directly). The mock should **not** wrap webhook payloads in an extra `data` key.

### Transfer endpoint has `sendViaInterledger` field

The `POST /multicurrency-wallets/transfer` endpoint has a `sendViaInterledger` boolean that is documented but not used by the backend. The mock can accept and ignore it.

## Open Questions

- **What does `GetWallet` actually return?** The response schema for `GET /multicurrency-wallets/get` is login-gated in the docs. The mock returns the same shape as `CreateWallet` (id, name, parent, uid, verification.status). If the backend expects additional fields (e.g., balances), this will need to be updated once observed in the sandbox.
- **What rail values are valid for `/info/fee-estimate`?** The docs say "9 enum values" but don't list them publicly. The backend sends `"interac"`. The mock accepts any string for `rail` but returns a fee estimate regardless.
- **`payout.interac.expired` vs `payout.interac.cancelled`**: The backend handles both identically (calls `ExecuteFinishWithdraw`). The mock sends `completed` by default; `expired`/`cancelled` are only sent if configured. The exact triggering conditions in production are unknown.
- **`chimoney.redeem.failed`**: The backend handles this event by calling `ExecuteFinishDeposit` with `"failed"` status. The mock only sends this if the deposit is explicitly failed (not via the normal pay page flow). A way to trigger this from the UI may be desirable. 

## Backend Changes Required

Two environment variables must be added to the backend before mockchimoney can be used locally:

1. `CHIMONEY_API_BASE_URL` — overrides the hardcoded prod/sandbox base URL in `external/client.go`
2. `CHIMONEY_KYC_BASE_URL` — overrides the hardcoded `dash.chimoney.io` URL in `ops/ops.go`

See [`plan/plan.md`](plan/plan.md) for the exact code changes.
