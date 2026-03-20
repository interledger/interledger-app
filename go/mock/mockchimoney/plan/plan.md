# MockChimoney — Implementation Plan

## Overview

MockChimoney is a local-development HTTP mock of the Chimoney payment provider API. It follows the same patterns as `mockgatehub` and `mockxago`: a Go service using `go-chi`, in-memory or Redis-backed storage, a background webhook worker, and a Traefik-fronted service in the local Docker Compose environment.

The mock must satisfy two integration points:
1. **Backend (`go/backend`)** — HTTP REST API calls from `providers/chimoney/external/client.go`
2. **Protea (frontend)** — KYC iframe served by `mockchimoney`, rendered inside an `<iframe>` in the browser

---

## What the Backend Calls (API Surface to Implement)

All authenticated via `X-API-KEY` header (validated against a configured `MOCKCHIMONEY_API_KEY`).

Every response has the envelope:
```json
{ "status": "success" | "error", "error": "...", "data": { ... } }
```

### 1. `POST /v0.2.4/multicurrency-wallets/create`
Create a chimoney sub-account (called during `CreateChimoneyUserWorkflow`).

**Request body:**
```json
{ "name": "...", "email": "...", "firstName": "...", "lastName": "...", "phoneNumber": "..." }
```
**Response `data`:**
```json
{ "id": "<uuid>", "parent": "...", "uid": "...", "name": "...", "subAccount": true, "verification": { "status": "pending" } }
```
The returned `id` is stored in `chi_money_wallets.external_id` and used as the chimoney sub-account ID in all subsequent calls.

---

### 2. `GET /v0.2.4/multicurrency-wallets/get?id=<subAccountID>`
Retrieve a sub-account by ID (used in `GetChimoneyWallet` activity).

**Response `data`:** same shape as create response.

---

### 3. `POST /v0.2.4/multicurrency-wallets/transfer`
Transfer between two sub-accounts (P2P payments between chimoney users).

**Request body:**
```json
{
  "subAccount": "<sender-subAccountID>",
  "receiver": "<receiver-subAccountID>",
  "amountToSend": "50.00",
  "originCurrency": "CAD",
  "destinationCurrency": "CAD",
  "turnOffNotification": true
}
```
**Response:** `{ "status": "success", "data": {} }` — no meaningful data content needed.

---

### 4. `POST /v0.2.4/payment/initiate`
Create a deposit payment link. Returns a URL the user opens to pay.

**Request body:**
```json
{
  "amount": "100.00",
  "currency": "CAD",
  "subAccount": "<chiWalletID>",
  "payerEmail": "user@example.com",
  "turnOffNotification": true,
  "redirect_url": "https://protea.interledger.test/callbacks/chimoney"
}
```

**Validation (mock enforces):**
- `payerEmail` is **required** — the official Chimoney docs mark it as required. The mock must return `400` if absent.
- `currency` must be one of `USD`, `NGN`, `CAD` (the complete enum from the official docs). Any other value returns `400`. The backend currently always sends `CAD`; the mock validates against the full documented enum so that invalid values are caught early.

**Response `data`:**
```json
{
  "paymentLink": "https://mockchimoney.interledger.test/pay/<issueID>",
  "issueID": "<chiWalletID>_<uuid>",
  "chiRef": "<uuid>",
  "status": "pending",
  "payerEmail": "user@example.com"
}
```

**Critical:** The `issueID` format must be `{chiWalletID}_{random}`. The backend parses it with `ExtractChiWalletIDFromIssueID` (splits on `_` and takes `parts[0]`) to recover the chimoney sub-account ID from the issue ID alone — including in webhooks that arrive without explicit wallet context.

The `paymentLink` must point to `GET /pay/<issueID>` on mockchimoney, which serves an HTML page where the user "completes" the deposit (see Section 5).

---

### 5. `GET /pay/<issueID>` — Deposit confirmation page (browser-facing)
This is **not** a backend API call; it is loaded by the browser after the user is redirected to `paymentLink`. The page simulates the Chimoney payment form on which the user provides card/Interac details and confirms payment.

On confirmation, the mock:
1. Marks the payment as `redeemed` in storage.
2. Redirects the browser to `redirect_url?issueID=<issueID>&status=success` (from the stored deposit record), which is `/callbacks/chimoney?issueID=...&status=success` in protea.
3. Asynchronously sends two webhooks to the backend (with a short configurable delay):
   - `charge.interac.completed` or `charge.chimoney-wallet.completed` — triggers `CreateChimoneyDepositWorkflow`
   - `chimoney.redeem.completed` — triggers `FinishChimoneyDepositWorkflow`

Both webhooks contain `{ "eventType": "...", "issueID": "<issueID>", "status": "completed" }`.

---

### 6. `POST /v0.2.4/payment/verify`
Called during `VerifyChimoneyPayment` activity to get payment status. Called after both deposit webhooks.

**Request body:**
```json
{ "id": "<issueID>", "subAccount": "<chiWalletID>" }
```
**Response `data`:** a `Payment` object:
```json
{
  "id": "<uuid>",
  "issueID": "<issueID>",
  "amount": "100.00",
  "currency": "CAD",
  "subAccount": "<chiWalletID>",
  "status": "redeemed",
  "chiRef": "<uuid>",
  "payerEmail": "user@example.com",
  "issueDate": "<RFC3339>",
  "meta": {
    "amount": 100.0,
    "processingFee": { "amount": 1.5, "currency": "CAD", "grossAmount": 101.5, "netAmount": 100.0, "provider": "interac" }
  }
}
```
The `status` field drives `FinishChimoneyDepositWorkflow`: `"redeemed"` → complete the deposit; `"failed"` → fail it.

---

### 7. `POST /v0.2.4/payouts/interac`
Execute an Interac withdrawal. Called during `ChimoneyWithdraw` activity.

**Request body:**
```json
{
  "interacs": [{ "name": "...", "email": "...", "amount": 95.00, "narration": "..." }],
  "debitCurrency": "CAD",
  "subAccount": "<chiWalletID>",
  "turnOffNotification": true
}
```
**Response `data`:**
```json
{
  "data": [{
    "id": "<uuid>",
    "issueID": "<chiWalletID>_<uuid>",
    "amount": 95.00,
    "fee": 1.50,
    "debitCurrency": "CAD",
    "type": "interac",
    "chiref": "<uuid>"
  }]
}
```
The `issueID` follows the same `{chiWalletID}_{random}` format.

After returning 200, the mock enqueues a background job to send a `payout.interac.completed` (or `payout.interac.expired`) webhook to the backend after a configurable delay. The withdrawal webhook payload:
```json
{
  "eventType": "payout.interac.completed",
  "issueID": "<issueID>",
  "status": "completed",
  "amount": "95.00",
  "meta": {
    "issuer": "<chiWalletID>",
    "amount": 95.00,
    "currency": "CAD",
    "paymentType": "interac"
  }
}
```

---

### 8. `POST /v0.2.4/payouts/status`
Check payout status by `chiRef`. Called during `ChimoneyPayoutStatus` activity.

**Request body:**
```json
{ "chiRef": "<uuid>", "subAccount": "<chiWalletID>", "turnOffNotification": true }
```
**Response `data`:**
```json
{ "id": "<uuid>", "amount": 95.0, "fee": 1.5, "type": "interac", "issueID": "<issueID>", "status": "completed" }
```

---

### 9. `POST /v0.2.4/info/fee-estimate`
Estimate Interac withdrawal fee. Called in `GetEstimatedFee`.

**Request body:**
```json
{ "amount": "100.00", "currency": "CAD", "rail": "interac", "direction": "payout" }
```
**Response `data`:**
```json
{ "amount": 100.0, "currency": "CAD", "rail": "interac", "direction": "payout", "totalFee": 1.50, "netAmount": 98.50 }
```
The fee amounts can be configured via env vars (default: flat `$1.50`).

---

### 10. `GET /v0.2.4/info/convert/local-amount-to-usd?amountInOriginCurrency=<N>&originCurrency=<currency>`
Convert a currency amount to USD. Used internally.

**Response `data`:**
```json
{ "originCurrency": "CAD", "amountInOriginCurrency": "100", "amountInUSD": 73.50, "validUntil": "...", "expiresAt": "...", "expiresAtTimestamp": 1234567890 }
```
A fixed exchange rate (e.g., CAD→USD = 0.735) is sufficient for local testing.

---

## KYC Widget (Browser-Facing)

The backend constructs the KYC widget URL as:
```
{CHIMONEY_KYC_BASE_URL}/verify/kyc/{chiSubAccountID}?redirect={callbackURL}
```
Currently this URL is hardcoded to `https://dash.chimoney.io` (prod) or `https://sandbox.chimoney.io` (non-prod). A new env var `CHIMONEY_KYC_BASE_URL` must be added to the backend to allow routing this to mockchimoney in local environments.

### `GET /verify/kyc/<externalID>?redirect=<callbackURL>`
Returns an HTML page. It shows a simple "Complete KYC" form and a simulate-approval button.

On approval, the mock:
1. Stores KYC status as `completed` for the sub-account.
2. Redirects the browser to `{redirect_url}&kyc` (the protea callback page).
3. Sends a `user.kyc.completed` webhook to the backend:
   ```json
   { "eventType": "user.kyc.completed", "userID": "<externalID>" }
   ```

The protea callback page at `/callbacks/chimoney?kyc` posts `{ kyc: true }` to the parent window, signalling the `personal-details` page to submit the KYC completion POST.

A "Decline KYC" button sends `user.kyc.declined` instead and redirects with `status=failed`.

---

## Webhook Delivery (Svix Signature Format)

The backend verifies chimoney webhooks using svix-style HMAC-SHA256:
- **Headers:** `svix-id: <uuid>`, `svix-timestamp: <unix-seconds>`, `svix-signature: v1,<base64-hmac>`
- **Signed content:** `{svix-id}.{svix-timestamp}.{raw-body}`
- **Secret parsing:** The backend reads `CHIMONEY_WEBHOOK_SECRET`, splits on `_`, base64-decodes the part after `_`.

Therefore, mockchimoney's webhook secret env var must be in the same format: `prefix_<base64(secret)>`. A convenient fixed local value:
```
CHIMONEY_WEBHOOK_SECRET=local_bG9jYWwtdGVzdC13ZWJob29rLXNlY3JldA==
```
(`local-test-webhook-secret` base64-encoded.)

The mock signs every outgoing webhook with this secret using the same algorithm. The backend's `Verify()` function will accept it.

---

## Project Structure

Mirrors mockgatehub and mockxago:

```
go/mock/mockchimoney/
├── .gitignore
├── .golangci.yml
├── AGENTS.md
├── Dockerfile
├── Makefile
├── README.md
├── cmd/
│   └── mockchimoney/
│       └── main.go              # Bootstrap: config, storage, webhook worker, chi router
├── features/
│   ├── wallet_management.feature
│   ├── deposits.feature
│   ├── withdrawals.feature
│   ├── fee_estimation.feature
│   ├── kyc.feature
│   └── webhooks.feature
├── internal/
│   ├── config/
│   │   └── config.go            # Env-var-driven config struct
│   ├── handler/
│   │   ├── handler.go           # Handler struct, request logger middleware, health check
│   │   ├── wallet.go            # CreateWallet, GetWallet
│   │   ├── payment.go           # Deposit initiate, verify, pay page (browser)
│   │   ├── payout.go            # Interac withdrawal, payout status
│   │   ├── fee.go               # Fee estimate
│   │   ├── convert.go           # Currency conversion
│   │   ├── kyc.go               # KYC widget page, KYC completion form handler
│   │   └── transfer.go          # Wallet-to-wallet transfer
│   ├── jobs/                    # Background webhook job queue (same pattern as mockxago)
│   │   ├── job.go
│   │   ├── queue.go
│   │   └── worker.go
│   ├── logger/
│   │   └── logger.go            # Zap logger wrapper
│   ├── models/
│   │   ├── api.go               # APIResponse envelope type
│   │   └── models.go            # SubAccount, Payment, Payout storage models
│   ├── storage/
│   │   ├── interface.go         # Storage interface
│   │   ├── memory.go            # In-memory implementation
│   │   └── redis.go             # Redis-backed implementation (optional)
│   └── webhook/
│       └── sender.go            # Signs and POSTs webhooks with svix headers
├── testenv/                     # Godog BDD integration tests
│   ├── godog_test.go
│   ├── docker-compose.yml
│   ├── http_client.go
│   ├── test_context.go
│   ├── helpers.go
│   ├── wallet_steps.go
│   ├── payment_steps.go
│   ├── withdrawal_steps.go
│   ├── kyc_steps.go
│   └── webhook_server.go        # Test webhook receiver
└── web/
    ├── pay.html                 # Deposit confirmation page served to browser
    └── kyc.html                 # KYC completion page served in iframe
```

---

## Storage Model

### SubAccount
```go
type SubAccount struct {
    ID          string    // chimoney external sub-account ID
    ParentID    string    // parent account ID (the org/root account)
    Name        string
    Email       string
    KYCStatus   string    // "pending", "completed", "declined"
    CreatedAt   time.Time
}
```

### Payment (deposit)
```go
type Payment struct {
    IssueID     string    // "{subAccountID}_{uuid}"
    SubAccount  string    // chimoney sub-account ID
    Amount      float64
    Currency    string
    Status      string    // "pending", "redeemed", "failed"
    PayerEmail  string
    RedirectURL string
    ChiRef      string
    CreatedAt   time.Time
}
```

### Payout (withdrawal)
```go
type Payout struct {
    ID          string    // uuid
    IssueID     string    // "{subAccountID}_{uuid}"
    SubAccount  string
    Amount      float64
    Fee         float64
    Currency    string
    Status      string    // "pending", "completed", "expired", "cancelled"
    ChiRef      string
    InteracEmail string
    CreatedAt   time.Time
}
```

---

## Configuration (Environment Variables)

| Variable | Default | Description |
|---|---|---|
| `MOCKCHIMONEY_PORT` | `8080` | HTTP listen port |
| `MOCKCHIMONEY_API_KEY` | `local-test-api-key` | Value expected in `X-API-KEY` header |
| `MOCKCHIMONEY_ENFORCE_AUTHENTICATION` | `true` | Whether to check `X-API-KEY` |
| `MOCKCHIMONEY_REDIS_URL` | `""` | Redis URL (empty → in-memory storage) |
| `MOCKCHIMONEY_REDIS_DB` | `5` | Redis database number |
| `CHIMONEY_WEBHOOK_SECRET` | `local_bG9jYWwtdGVzdC13ZWJob29rLXNlY3JldA==` | Used to sign outgoing webhooks (same env var name as backend reads) |
| `WEBHOOK_URL` | `http://backend:8080/webhooks/chimoney` | Backend webhook endpoint |
| `WEBHOOK_MIN_DELAY_SEC` | `0.5` | Minimum delay before sending deposit/withdrawal webhooks |
| `INTERAC_FEE_FLAT` | `1.50` | Flat CAD fee applied to Interac payouts |
| `CAD_TO_USD_RATE` | `0.735` | Exchange rate used for ConvertToUSD |
| `LOG_LEVEL` | `info` | Zap log level |

---

## Required Backend Changes

### 1. `external/client.go` — Configurable base URL

The `New()` constructor must read `CHIMONEY_API_BASE_URL` to override the hardcoded prod/sandbox URLs:

```go
func New(transport *http.Client) Client {
    baseURL := os.Getenv("CHIMONEY_API_BASE_URL")
    if baseURL == "" {
        if env.IsProd() {
            baseURL = "https://api.chimoney.io/v0.2.4"
        } else {
            baseURL = "https://api-v2-sandbox.chimoney.io/v0.2.4"
        }
    }
    // ...
}
```

### 2. `ops/ops.go` — Configurable KYC widget base URL

`GetKYCWidget` currently hardcodes the chimoney dashboard URL. Add `CHIMONEY_KYC_BASE_URL`:

```go
func GetKYCWidget(ctx context.Context, b Backends, walletID string) (string, error) {
    // ... wallet lookup / creation ...

    baseURL := os.Getenv("CHIMONEY_KYC_BASE_URL")
    if baseURL == "" {
        if env.IsProd() {
            baseURL = "https://dash.chimoney.io"
        } else {
            baseURL = "https://sandbox.chimoney.io"
        }
    }
    // ...
}
```

These two changes mean no application logic changes are needed; only configuration wiring differs between environments.

---

## Required Local Environment Changes

### New file: `local/mockchimoney.yaml`

```yaml
services:
  mockchimoney:
    build:
      context: ..
      dockerfile: go/mock/mockchimoney/Dockerfile
    profiles:
      - application
    restart: always
    expose:
      - "8080"
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.mockchimoney.rule=Host(`mockchimoney.interledger.test`)"
      - "traefik.http.routers.mockchimoney.entrypoints=websecure"
      - "traefik.http.services.mockchimoney.loadbalancer.server.port=8080"
    environment:
      LOG_LEVEL: debug
      MOCKCHIMONEY_REDIS_URL: redis://redis:6379
      MOCKCHIMONEY_REDIS_DB: '5'
      MOCKCHIMONEY_API_KEY: ${BACKEND_CHIMONEY_API_KEY:-local-test-api-key}
      WEBHOOK_URL: http://backend:8080/webhooks/chimoney
      CHIMONEY_WEBHOOK_SECRET: ${BACKEND_CHIMONEY_WEBHOOK_SECRET:-local_bG9jYWwtdGVzdC13ZWJob29rLXNlY3JldA==}
      WEBHOOK_MIN_DELAY_SEC: '0.5'
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 3
      start_period: 5s
```

### Update `local/docker-compose.yaml`

Add `- mockchimoney.yaml` to the `include:` list alongside `mockgatehub.yaml` and `mockxago.yaml`.

### Update `local/wallet.yaml` — backend service environment

Add the following env vars to the `backend` service:

```yaml
# Chimoney
- CHIMONEY_API_BASE_URL=${BACKEND_CHIMONEY_API_BASE_URL:-http://mockchimoney:8080/v0.2.4}
- CHIMONEY_KYC_BASE_URL=${BACKEND_CHIMONEY_KYC_BASE_URL:-https://mockchimoney.interledger.test}
- CHIMONEY_TOKEN=${BACKEND_CHIMONEY_TOKEN:-local-test-api-key}
- CHIMONEY_WEBHOOK_SECRET=${BACKEND_CHIMONEY_WEBHOOK_SECRET:-local_bG9jYWwtdGVzdC13ZWJob29rLXNlY3JldA==}
```

---

## Webhook Signing Implementation

The sender in `internal/webhook/sender.go` must implement the svix signature scheme that the backend `Verify()` function checks:

```
svix-id: msg_<uuid>
svix-timestamp: <unix-epoch-seconds>
svix-signature: v1,<base64(HMAC-SHA256(key, "{svix-id}.{svix-timestamp}.{body}"))>
```

Where `key` is the base64-decoded suffix of `CHIMONEY_WEBHOOK_SECRET` (i.e., after splitting on `_` and decoding `parts[1]`). This exactly mirrors the backend's `ParseWebhookSecret` + `Verify` logic.

---

## Deposit Flow (End-to-End in Local Environment)

1. User clicks "Deposit" → protea calls `GetChimoneyDepositLink` gRPC → backend calls `POST /v0.2.4/payment/initiate` → mockchimoney returns `paymentLink`.
2. Protea renders `paymentLink` in an `<iframe>` → browser loads `GET /pay/<issueID>` from mockchimoney.
3. User clicks "Pay" → mockchimoney marks payment as `redeemed`, redirects browser to `/callbacks/chimoney?issueID=...&status=success`.
4. Protea's `/callbacks/chimoney` route reads query params and posts `{ issueID, status }` to the parent window.
5. Protea's deposit page listens for the message, sees `status=success`, auto-submits `chimoney-successfull-deposit` form → redirect to `/`.
6. Meanwhile, mockchimoney sends (after `WEBHOOK_MIN_DELAY_SEC`):
   - `charge.interac.completed` webhook → backend starts `CreateChimoneyDepositWorkflow` → calls `VerifyPayment` on mockchimoney (returns `status: "redeemed"`) → creates deposit transaction record.
   - `chimoney.redeem.completed` webhook → backend starts `FinishChimoneyDepositWorkflow` → calls `VerifyPayment` again → `AssignChimoneyBalance` → `CompleteChimoneyTransaction` → sends deposit confirmation email.

---

## Withdrawal Flow (End-to-End in Local Environment)

1. User initiates withdrawal → backend starts `ExecuteChimoneyWithdrawalWorkflow` → `ReserveBalance` → calls `POST /v0.2.4/payouts/interac` on mockchimoney → stores `issueID` as `foreignID`.
2. Mockchimoney returns success with `issueID`.
3. After `WEBHOOK_MIN_DELAY_SEC`, mockchimoney sends `payout.interac.completed` (with `meta.issuer = chiWalletID`, `meta.amount`, `meta.currency`).
4. Backend `handleWithdrawal` → `ExecuteFinishWithdraw` → `FinalizeReserve` → `UpdateWithdrawalFeeAndCompleteTransaction` → sends withdrawal completion email.

---

## KYC Flow (End-to-End in Local Environment)

1. User opens personal-details page → protea calls `GetKYCProviderWidget` gRPC → backend calls `GetKYCWidget` → constructs `https://mockchimoney.interledger.test/verify/kyc/<externalID>?redirect=<callbackURL>` → returns as `chimoneyWidget`.
2. Protea renders the URL in an `<iframe>`.
3. User clicks "Complete KYC" on the mockchimoney KYC page.
4. Mockchimoney redirects browser to `<callbackURL>&kyc` (`/callbacks/chimoney?kyc`).
5. Protea's callback page posts `{ kyc: true }` to parent → `personal-details` submits POST → marks user as KYC-complete.
6. Mockchimoney sends `user.kyc.completed` webhook with `{ eventType: "user.kyc.completed", userID: "<externalID>" }`.
7. Backend `handleKYC` → resolves `walletID` from `externalID` via `GetWalletID` → `b.KYC().SetKYCStatus(walletID, kyc.StatusLevel1)`.

---

## Testing Plan

### Unit tests (`internal/...`)
- Handler tests for each endpoint (table-driven, no external dependencies)
- Storage tests (memory store): wallet CRUD, payment state transitions, payout state transitions
- Webhook sender: signature generation matches backend `Verify()` expectations

### Godog BDD tests (`testenv/`)

Feature scenarios covering:
- **`wallet_management.feature`**: Create wallet, get wallet, duplicate creation is idempotent
- **`deposits.feature`**: Initiate deposit → browser pays → webhooks fire → payment verifiable as `redeemed`
- **`withdrawals.feature`**: Interac payout → webhook fires `completed` or `expired`
- **`fee_estimation.feature`**: Fee estimate returns correct totalFee
- **`kyc.feature`**: KYC page redirects correctly, `user.kyc.completed` webhook fires
- **`webhooks.feature`**: Outgoing webhooks are correctly svix-signed; rejects webhooks sent with wrong secret

### Integration with local environment
MockChimoney participates in E2E tests via the existing `go test` suite in `e2e/`. No separate E2E step is needed beyond the local environment already running.

---

## Official API Documentation Cross-Reference

All endpoints below were cross-referenced against the [Chimoney API Reference](https://chimoney.readme.io/reference) (fetched directly; response body schemas are login-gated but request fields are public).

### Summary Table

| Endpoint | Doc URL | Backend match | Notes |
|---|---|---|---|
| `POST /multicurrency-wallets/create` | [doc](https://chimoney.readme.io/reference/post_v0-2-4-multicurrency-wallets-create) | ✅ | `name` required; `email`, `firstName`, `lastName`, `phoneNumber` optional |
| `GET /multicurrency-wallets/get` | [doc](https://chimoney.readme.io/reference/get_v0-2-4-multicurrency-wallets-get) | ✅ | Single query param `id` (required) |
| `POST /multicurrency-wallets/transfer` | [doc](https://chimoney.readme.io/reference/post_v0-2-4-multicurrency-wallets-transfer) | ✅ | `amountToSend`, `originCurrency`, `destinationCurrency` required; `subAccount`, `receiver`, `turnOffNotification`, `sendViaInterledger` optional |
| `POST /payment/initiate` | [doc](https://chimoney.readme.io/reference/post_v0-2-4-payment-initiate) | ✅ | `payerEmail` **required** — mock validates and returns 400 if absent; `currency` must be one of `USD`, `NGN`, `CAD` |
| `POST /payment/verify` | [doc](https://chimoney.readme.io/reference/post_v0-2-4-payment-verify) | ✅ | `id` = issueID (required), `subAccount` optional |
| `POST /payouts/interac` | [doc](https://chimoney.readme.io/reference/post_v0-2-4-payouts-interac) | ✅ | `debitCurrency` required, `interacs` array required, `subAccount` optional |
| `POST /payouts/status` | [doc](https://chimoney.readme.io/reference/post_v0-2-4-payouts-status) | ✅ | `chiRef` required, `subAccount` optional |
| `POST /info/fee-estimate` | [doc](https://chimoney.readme.io/reference/post_v0-2-4-info-fee-estimate) | ✅ | `amount` (number, required); `currency` (defaults to USD); `rail` (enum, optional, 9 values); `direction` enum: `payout` \| `funding` |
| `GET /info/convert/local-amount-to-usd` | [doc](https://chimoney.readme.io/reference/get_v0-2-4-info-convert-local-amount-to-usd) | ✅ | Query params: `originCurrency` (required), `amountInOriginCurrency` (integer, required) |
| `GET /sub-account/kyc/link` | [doc](https://chimoney.readme.io/reference/get_v0-2-4-sub-account-kyc-link-1) | N/A | Backend does NOT call this; instead hardcodes KYC widget URLs. Mock implements its own KYC page at `GET /verify/kyc/<externalID>`. |

### Detailed Findings Per Endpoint

#### `POST /multicurrency-wallets/create`
- ✅ Field names confirmed: `name` (required), `email`, `firstName`, `lastName`, `phoneNumber`, `meta` (optional object).
- Response shape not rendered (login-gated), but the backend `CreateWalletResp` struct (with `ID`, `Parent`, `SubAccount`, `UID`, `Name`, `Verification.Status`) is derived from actual API observations.

#### `GET /multicurrency-wallets/get`
- ✅ Single query param: `id` (required) — matches `?id=<id>` in Go client.
- Doc description: "Retrieves the details of a specified multicurrency wallet (using the walletID) such as wallet balance, transaction history, and account metadata."

#### `POST /multicurrency-wallets/transfer`
- ✅ Full field list confirmed: `sender` (optional, not used by backend), `subAccount` (optional = sender sub-account), `amountToSend` (required), `originCurrency` (required), `receiver` (optional = receiver sub-account), `email` (optional), `phoneNumber` (optional), `destinationCurrency` (required), `narration` (optional), `turnOffNotification` (boolean), `sendViaInterledger` (boolean, not used by backend).
- Backend sends `subAccount`, `receiver`, `amountToSend`, `originCurrency`, `destinationCurrency`, `turnOffNotification` — all confirmed valid fields.

#### `POST /payment/initiate`
- ⚠️ **`payerEmail` is documented as required** — the backend always populates it (from user lookup), so this is functionally correct but worth noting.
- ⚠️ **`currency` enum is restricted**: docs show only `USD`, `NGN`, `CAD`. Backend only uses `CAD`. Mock should accept at minimum `CAD`.
- `redirect_url` (optional) is the callback URL. `subAccount` (optional) is the settlement wallet.
- `paymentMethod` object exists in the doc for directing to specific rails (card, Interac, etc.) — backend doesn't use it; mock can ignore it.

#### `POST /payment/verify`
- ✅ Field `id` maps to `issueID` in Go struct tag `json:"id,omitempty"`. Field `subAccount` is optional.
- No discrepancies.

#### `POST /payouts/interac`
- ✅ Exact match: `debitCurrency` (required), `interacs` (required array), `subAccount` (optional), `turnOffNotification` (boolean).

#### `POST /payouts/status`
- ✅ Exact match: `chiRef` (required), `subAccount` (optional), `turnOffNotification` (boolean).

#### `POST /info/fee-estimate`
- ✅ Field names confirmed: `amount` (number, required), `currency` (string, defaults to USD, must be USD when rail is omitted), `rail` (enum, optional — 9 values not listed publicly), `direction` (enum: `payout` | `funding`, defaults to `payout`).
- ⚠️ **Newly added endpoint** (doc updated ~22 days before this plan was written). Confirmed by Adi (backend developer) that this endpoint is **not yet deployed to production**. The backend has a fallback estimation formula (`$1.00 + 0.5% × amount`) for when this call fails.
- Mock must implement this endpoint regardless — the backend tries it first before falling back.
- The backend sends `currency: "CAD"` and `rail: "interac"`, `direction: "payout"`.

#### `GET /info/convert/local-amount-to-usd`
- ✅ Query params confirmed: `originCurrency` (currency code, required), `amountInOriginCurrency` (integer, required). Doc description: "Converts a specified amount in a local currency (e.g., KES) to its equivalent value in USD."
- Backend passes `amountInOriginCurrency=<N>&originCurrency=<currency>` — matches exactly.

### Webhook Events — Public Docs vs. Actual Backend Usage

The [webhooks documentation](https://chimoney.readme.io/reference/webhooks-and-events) lists only a **partial and outdated** set of event types:
- `chimoney.payment.completed`
- `chimoney.redeem.completed`
- `payout.bank.initiated`, `payout.bank.failed`, `payout.bank.completed`
- `payout.gift-card.initiated`, `payout.gift-card.failed`, `payout.gift-card.completed`

**The actual events handled by the backend** (from `ops/webhooks.go`) are **not listed** in the public docs. These must be discovered from the backend code itself:

| Event type | Trigger source | Backend handler |
|---|---|---|
| `charge.interac.completed` | Interac deposit funded | `handleConfirmedOrCompletedCharge` → `CreateDeposit` |
| `charge.card.completed` | Card deposit funded | `handleConfirmedOrCompletedCharge` → `CreateDeposit` |
| `charge.chimoney-wallet.completed` | Chimoney-wallet deposit | `handleConfirmedOrCompletedCharge` → `CreateDeposit` |
| `charge.crypto.xrpl.confirmed` | XRPL crypto deposit | `handleConfirmedOrCompletedCharge` → `CreateDeposit` |
| `charge.crypto.celo.confirmed` | Celo crypto deposit | `handleConfirmedOrCompletedCharge` → `CreateDeposit` |
| `chimoney.redeem.completed` | Payment redeemed by user | `handleRedeemWebhook("completed")` → `ExecuteFinishDeposit` |
| `chimoney.redeem.failed` | Payment redemption failed | `handleRedeemWebhook("failed")` → `ExecuteFinishDeposit` |
| `payout.interac.completed` | Interac withdrawal completed | `handleWithdrawal` → `ExecuteFinishWithdraw` |
| `payout.interac.expired` | Interac withdrawal expired | `handleWithdrawal` → `ExecuteFinishWithdraw` |
| `payout.interac.cancelled` | Interac withdrawal cancelled | `handleWithdrawal` → `ExecuteFinishWithdraw` |
| `user.kyc.completed` | KYC approved | `handleKYC` → `SetKYCStatus(Level1)` |
| `user.kyc.declined` | KYC rejected | `handleKYC` → `SetKYCStatus(Declined)` |

Mock should send the events relevant to local testing: `charge.interac.completed`, `chimoney.redeem.completed`, `chimoney.redeem.failed`, `payout.interac.completed`, `payout.interac.expired`, `user.kyc.completed`, `user.kyc.declined`.

### Sandbox vs. Production

| | Production | Sandbox |
|---|---|---|
| Base URL | `https://api.chimoney.io/v0.2.4` | `https://api-v2-sandbox.chimoney.io/v0.2.4` |
| KYC widget | `https://dash.chimoney.io/verify/kyc/…` | `https://sandbox.chimoney.io/verify/kyc/…` |
| Funded with | Real funds | $1000 (10,000 Chimoney) test amount |
| Simulate endpoints | Not available | `POST /payment/simulate-interac-funding`, `POST /payment/simulate-funding`, `POST /payment/simulate` |

Mock replaces the sandbox entirely for local development — all simulate endpoints are irrelevant since the mock controls its own webhook delivery directly.

### KYC Widget API (for reference)
The official API has `GET /sub-account/kyc/link?subAccountID=<id>&redirectUrl=<url>&skip=<bool>` which returns a hosted KYC page URL. The backend bypasses this and constructs the KYC URL directly:
```
{CHIMONEY_KYC_BASE_URL}/verify/kyc/{chiSubAccountID}?redirect={callbackURL}
```
The mock must implement this URL pattern at `GET /verify/kyc/<externalID>`.

---

## Implementation Phases

Each phase follows strict TDD discipline:

1. **Red** — Run the relevant feature scenarios. Confirm they fail (compilation errors or assertion failures).
2. **Green** — Write the minimum code to make the failing scenarios pass. No gold-plating.
3. **Refactor** — Eliminate duplication, improve naming, tighten error handling, re-run all scenarios to stay green. Lint (`golangci-lint run ./...`) must pass before closing the phase.

Phases are ordered by dependency. Do not start a phase until all previous phases are green and linted.

Each phase has its own document in the `plan/` directory:

| Phase | Title | Dependencies | Status |
|-------|-------|------|--------|
| [Phase 0](phase-0-bootstrap.md) | Bootstrap: Repository Skeleton and Health Check | None | Planned |
| [Phase 1](phase-1-storage-and-wallets.md) | Storage Layer and Wallet APIs | Phase 0 | Completed (2026-03-20) |
| [Phase 2](phase-2-authentication.md) | Authentication Middleware | Phases 0-1 | Planned |
| [Phase 3](phase-3-wallet-transfer.md) | Wallet Transfer | Phases 0-2 | Planned |
| [Phase 4](phase-4-payment-initiation.md) | Payment Initiation and Verification | Phases 0-3 | Planned |
| [Phase 5](phase-5-deposit-webhooks.md) | Webhook Infrastructure and Deposit Webhooks | Phases 0-4 | Planned |
| [Phase 6](phase-6-withdrawals.md) | Withdrawal Flow and Webhooks | Phases 0-5 | Planned |
| [Phase 7](phase-7-fee-and-conversion.md) | Fee Estimation and Currency Conversion | Phases 0-6 | Planned |
| [Phase 8](phase-8-kyc.md) | KYC Widget | Phases 0-7 | Planned |
| [Phase 9](phase-9-redis.md) | Redis Storage Backend | Phases 0-8 | Planned |
| [Phase 10](phase-10-local-integration.md) | Local Environment Integration | Phases 0-9 | Planned |
| [Phase 11 (Optional)](phase-11-bdd-tests.md) | BDD Test Suite | Phases 0-9 (can run in parallel with Phase 10) | Planned |

---

## Non-Goals / Out of Scope

- Real Chimoney API compatibility beyond what the backend actually calls
- Support for currencies beyond `USD`, `NGN`, `CAD` — these three are the complete documented enum; the backend currently only sends `CAD` but the mock validates against the full set
- Card/crypto payment methods (the pay page only needs to simulate Interac for local dev; the mock may just show a single "Pay" button regardless of payment method)
- Webhook retry logic (a single delivery attempt is sufficient for local dev; unlike mockgatehub's Redis-backed retry queue, a simple in-process goroutine delay is acceptable)
- The `PaymentType` field on responses (`charge.interac.completed` vs `charge.card.completed`) can be hardcoded to `interac` since the backend uses all charge events identically
