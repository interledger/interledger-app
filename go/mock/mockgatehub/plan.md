# MockGatehub Web UI — Implementation Plan

## Goal

Add a minimalistic server-rendered web UI at `/ui/` that lets developers:

1. **Browse** all data stored in the mock (users, balances, wallets, transactions, cards, card transactions)
2. **Trigger actions** — KYC state transitions (with webhook dispatch) and simulated card transactions

---

## Design Decisions

| Question | Decision |
|---|---|
| Data enumeration approach | Add `List*` methods to `Storage` interface — works with both memory and Redis |
| KYC actions exposed | `action_required`, `rejected`, `accepted` — all fire the corresponding webhook |
| Card transaction control | Full form: pick user → card, set amount / currency / merchant / type |
| Frontend stack | Plain HTML + vanilla JS, Go `html/template` — consistent with existing iframe pages |
| Auth | No auth — open like `/admin/fees` (dev-only service) |
| Balance adjustment | Out of scope for this iteration |

---

## Webhook Compatibility Note

The UI KYC trigger will send the webhook payload matching **real GateHub behavior** as documented in `docs/gatehub-kyc.md`, so the wallet backend processes it correctly:

```json
// id.verification.accepted
{ "gateway": "paywiser-eu-sandbox", "verified": { "short": "accepted", "status": 1 } }

// id.verification.rejected
{ "gateway": "paywiser-eu-sandbox", "verified": { "short": "rejected", "status": 2 } }

// id.verification.action_required
{ "gateway": "paywiser-eu-sandbox", "verified": { "short": "action_required", "status": 0 } }
```

> **Note**: The existing `KYCIframeSubmit` and `UpdateKYCState` handlers send a simplified payload that does not include `gateway` or `verified`. The UI will send the documented real-GateHub format to maximise realism. If this causes issues (e.g., the wallet backend filters it out by gateway name), the format can be adjusted to match the simpler existing handlers. Watch backend logs after the first test trigger to confirm.

---

## Test Strategy Overview

**`make test`** runs three things in sequence: `lint` → `unit-tests` → `e2e-tests`.

- **lint**: `gofmt -l .` + `go vet ./...` + `golangci-lint run ./...` — code must be lint-clean.
- **unit-tests**: `go test -v ./...` — all packages under the repo root.
- **e2e-tests**: `go test -tags e2e ./testenv/` — godog BDD tests against Docker containers (require Docker to be running).

**Gate rule**: `make test` must pass at the end of every phase before work on the next phase begins. New phases add new tests; they never weaken existing ones. The e2e tests cover existing behaviour and must not regress.

**Handler test pattern** (consistent with the existing `internal/handler/*_test.go` files):
- Use `storage.NewMemoryStorage()` + `storage.SeedTestUsers(store)` in each test
- Call handlers directly via `httptest.NewRecorder()` / `httptest.NewRequest()`
- Assert HTTP status codes and that key response content is present
- Use `webhook.NewManager("", "test-secret", nil, nil, "")` (no URL, no queue) so webhooks don't actually fire but the manager can be called without panicking

**Storage test pattern** (consistent with the existing `internal/storage/memory_test.go` / `redis_test.go`):
- Memory tests: pure unit tests, no external dependencies
- Redis tests: require a running Redis instance; already skipped by the existing test suite when Redis is unavailable

---

## Phases Overview

```
Phase 1 ──────────────────────────────────────────────── Storage Foundation
    1a: interface.go additions (sequential prerequisite)
    1b: memory.go implementation + tests  ─┐  (parallel after 1a)
    1c: redis.go implementation + tests   ─┘
    ✅ make test gate

Phase 2 ──────────────────────────────────────────────── Routing Skeleton
    2a: routes + auth middleware bypass + stub handlers   (can run in parallel with Phase 1c)
    2b: HTML template files with placeholder content      (can run in parallel with 2a)
    ✅ make test gate

Phase 3 ──────────────────────────────────────────────── Data Browsing UI
    3a: Dashboard handler + template + tests  ─┐  (parallel after Phase 2)
    3b: User detail handler + template + tests ─┘
    ✅ make test gate

Phase 4 ──────────────────────────────────────────────── KYC Action
    ✅ make test gate

Phase 5 ──────────────────────────────────────────────── Card Transaction Action
    ✅ make test gate
```

Phases 4 and 5 are independent of each other and can be developed in parallel — both depend only on Phase 3 being complete.

---

## Phase 1 — Storage Foundation

**Goal**: Add three new methods to the `Storage` interface and implement them in both backends.

### 1a — Interface (sequential, must be first)

**File**: `internal/storage/interface.go`

Add:

```go
// List all users in storage (for admin UI)
ListUsers() ([]*models.User, error)

// List all transactions belonging to a user
ListTransactionsByUser(userID string) ([]*models.Transaction, error)

// Return all non-zero balances for a user as a map of currency → amount
GetAllBalances(userID string) (map[string]float64, error)
```

No tests at this step — the interface is not yet implemented. This is a compile-time contract.

### 1b — Memory implementation (parallel with 1c)

**File**: `internal/storage/memory.go`

**`ListUsers`**: acquire read lock, copy all values from `s.users` map into a slice, return.

**`ListTransactionsByUser`**: acquire read lock, iterate `s.transactions`, collect entries where `tx.UserID == userID`.

**`GetAllBalances`**: acquire read lock, return a shallow copy of `s.balances[userID]` (returns empty map — not nil — if user has no balances).

**File**: `internal/storage/memory_test.go` — add tests:

```
TestListUsers_Empty               → returns empty slice, no error
TestListUsers_WithSeededUsers     → returns both seeded users after SeedTestUsers
TestListUsers_AfterCreate         → creates a user, verifies it appears in list
TestListTransactionsByUser_Empty  → returns empty slice for user with no txns
TestListTransactionsByUser        → creates two txns for user1, one for user2; verifies filter
TestGetAllBalances_Empty          → returns empty map for user with no balances
TestGetAllBalances                → adds balances for USD and EUR, verifies both returned
```

### 1c — Redis implementation (parallel with 1b)

**File**: `internal/storage/redis.go`

**`ListUsers`**: use `SCAN 0 MATCH user:* COUNT 100` with cursor iteration; for each key, skip any that contain a second colon (e.g., `user:{id}:wallets`, `user:{id}:3ds:challenges`) by checking `strings.Count(key, ":") > 1`; `GET` and unmarshal each remaining key.

**`ListTransactionsByUser`**: maintain a new per-user Redis list `user:{id}:transactions`. Modify `CreateTransaction` in `redis.go` to also `LPUSH user:{userID}:transactions {txID}` immediately after storing the transaction. `ListTransactionsByUser` calls `LRANGE user:{id}:transactions 0 -1` and fetches each transaction by ID.

**`GetAllBalances`**: use `SCAN 0 MATCH balance:{userID}:* COUNT 100`; for each key, extract the currency suffix, call `GET` for the value, parse as float64 and add to result map.

**File**: `internal/storage/redis_test.go` — add tests mirroring the memory tests above:

```
TestRedisListUsers_Empty
TestRedisListUsers_WithSeededUsers
TestRedisListTransactionsByUser_Empty
TestRedisListTransactionsByUser
TestRedisGetAllBalances_Empty
TestRedisGetAllBalances
TestRedisCreateTransaction_AddsUserIndex   → creates txn, verifies LRANGE returns its ID
```

### Phase 1 — Test gate

```bash
make test
```

All existing tests must continue to pass. The 7 new memory tests and 7 new Redis tests must pass. No new BDD scenarios yet.

Verify manually that compilation succeeds: `make build`.

---

## Phase 2 — Routing Skeleton

**Goal**: Register all UI routes, bypass auth, and install stub handlers that return HTTP 200 with placeholder HTML. No real data yet — just structural scaffolding.

### 2a — Routes + auth middleware (can start in parallel with Phase 1c)

**File**: `internal/auth/middleware.go`

Add a prefix check for `/ui/` before the existing exact-match lookup:

```go
if strings.HasPrefix(path, "/ui/") || path == "/ui" {
    return true // public — no HMAC required
}
```

> This is the only change needed to auth. All `/ui/*` paths become public.

**File**: `cmd/mockgatehub/main.go`

Register routes inside a new `r.Route("/ui", ...)` block:

```go
r.Route("/ui", func(r chi.Router) {
    r.Get("/", h.UIDashboard)
    r.Get("/users/{userID}", h.UIUserDetail)
    r.Get("/actions/kyc", h.UIKYCForm)
    r.Post("/actions/kyc", h.UIKYCAction)
    r.Get("/actions/card-transaction", h.UICardTxForm)
    r.Post("/actions/card-transaction", h.UICardTxAction)
    r.Get("/actions/card-transaction/cards", h.UICardTxCards)
})
```

**File**: `internal/handler/ui.go` (new)

Stub implementations — each handler returns `200 OK` with a minimal HTML response:

```go
func (h *Handler) UIDashboard(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    fmt.Fprint(w, "<html><body><h1>MockGatehub UI — coming soon</h1></body></html>")
}
// ... same pattern for all other UI handlers
```

### 2b — Template files (parallel with 2a)

Create `web/ui/` directory with placeholder HTML files:

```
web/ui/dashboard.html          — user list page
web/ui/user.html               — user detail page
web/ui/kyc_action.html         — KYC trigger form
web/ui/card_tx_action.html     — card transaction form
```

Each file is a fully valid HTML document with a `<title>`, consistent header/nav, and a `{{/* placeholder */}}` comment where dynamic content will go. Include the shared CSS in a `<style>` block in each file (same palette as `web/index.html` — no external CDN).

> These files are not served yet (the stub handlers don't use them), but creating them now allows template and handler work to proceed in parallel in Phase 3.

### Phase 2 — Tests

**File**: `cmd/mockgatehub/main_test.go` — add route registration tests:

```
TestUIRoutesRegistered   → builds router via setupRoutes, sends GET /ui/ — expects 200
TestUIUserDetailRoute    → sends GET /ui/users/some-id — expects 200 (stub, no real lookup)
TestUIKYCFormRoute       → sends GET /ui/actions/kyc — expects 200
TestUICardTxFormRoute    → sends GET /ui/actions/card-transaction — expects 200
TestUINotFoundRoute      → sends GET /ui/nonexistent — expects 404
```

**File**: `internal/auth/middleware_test.go` — add:

```
TestPublicEndpoints_UIPrefix   → verifies /ui/, /ui/users/abc, /ui/actions/kyc all bypass auth
```

### Phase 2 — Test gate

```bash
make test
```

All existing tests pass. 5 new route tests and 1 new auth test pass. The stub handlers return 200 without touching storage.

---

## Phase 3 — Data Browsing UI

**Goal**: Replace stub handlers with real handlers that query storage and render the HTML templates. Both the dashboard and user detail page can be built in parallel.

### 3a — Dashboard (parallel with 3b)

**Handler**: `(h *Handler) UIDashboard(w http.ResponseWriter, r *http.Request)` in `internal/handler/ui.go`

Logic:
1. Call `h.store.ListUsers()` → slice of `*models.User`
2. For each user, call `h.store.GetAllBalances(user.ID)` to get their balances
3. Render `web/ui/dashboard.html` with the collected data

Template data struct:
```go
type DashboardData struct {
    Users []UserSummary
    Config ConfigSummary  // port, redis mode, webhook URL, auth enforcement
}
type UserSummary struct {
    User     *models.User
    Balances map[string]float64
}
type ConfigSummary struct {
    Port                  string
    RedisEnabled          bool
    WebhookURL            string
    EnforceAuthentication bool
}
```

**Template** `web/ui/dashboard.html`:
- Page header with service name and nav links ("Dashboard" | "Trigger KYC" | "Trigger Card Transaction")
- Config summary box (port, storage mode, webhook URL, auth status)
- Users table: Email | ID (truncated to 8 chars, full ID on hover) | KYC state (colour-coded badge: green=accepted, orange=action_required, red=rejected) | Balances (non-zero only) | Link to detail page
- "No users" message when list is empty

**Tests**: `internal/handler/ui_test.go` (new file):

```
TestUIDashboard_Renders            → seeds two users, calls handler, expects 200, body contains both emails
TestUIDashboard_Empty              → empty storage, expects 200, body contains "No users"
TestUIDashboard_KYCBadge          → user with accepted state → body contains "accepted"
```

### 3b — User Detail (parallel with 3a)

**Handler**: `(h *Handler) UIUserDetail(w http.ResponseWriter, r *http.Request)` in `internal/handler/ui.go`

Logic:
1. Extract `userID` from URL via `chi.URLParam(r, "userID")`
2. `h.store.GetUser(userID)` — 404 if not found
3. `h.store.GetAllBalances(userID)`
4. `h.store.GetWalletsByUser(userID)`
5. `h.store.ListTransactionsByUser(userID)`
6. `h.store.GetCustomerBySourceID(userID)` — nil is fine (no cards yet)
7. If customer found: `h.store.GetCardsByCustomer(*customer.ID)` and for each card `h.store.GetCardTransactionIDs(card.ID)` → `h.store.GetCardTransaction(id)`
8. Render `web/ui/user.html`

Template data struct:
```go
type UserDetailData struct {
    User         *models.User
    Balances     map[string]float64
    Wallets      []*models.Wallet
    Transactions []*models.Transaction
    Customer     *models.Customer       // nil if no card customer
    Cards        []CardWithTransactions
}
type CardWithTransactions struct {
    Card         *models.Card
    Transactions []*models.CardTransaction
}
```

**Template** `web/ui/user.html`:
- Back link to Dashboard
- Profile section: email, KYC state badge, risk level, name/address/DOB (if set), creation date
- Action buttons: "Trigger KYC Event" (links to `/ui/actions/kyc?userID={id}`) and "Trigger Card Transaction" (links to `/ui/actions/card-transaction?userID={id}`)
- Balances table: Currency | Amount | shown greyed if 0.00
- Wallets section: address (monospace, copyable), network, created at
- Transactions table: UUID (truncated) | type label | amount + currency | status badge | created at
- Cards section: masked PAN | status badge | account ID | created at; each card has a collapsible transaction sub-table

**Tests**: in `internal/handler/ui_test.go`:

```
TestUIUserDetail_KnownUser         → seeds user, expects 200, body contains email
TestUIUserDetail_UnknownUser       → random UUID, expects 404
TestUIUserDetail_WithBalances      → seeds user + balance, body contains "USD"
TestUIUserDetail_WithWallet        → seeds user + wallet, body contains wallet address
TestUIUserDetail_WithTransactions  → seeds user + transaction, body contains transaction ID (truncated)
TestUIUserDetail_WithCards         → seeds customer + card, body contains masked PAN
```

### Phase 3 — Test gate

```bash
make test
```

All existing tests pass. ~9 new handler tests pass. Both handlers fully replace their stubs. Templates are loaded from the filesystem using the same multi-path fallback pattern as `KYCIframe`:

```go
possiblePaths := []string{
    "web/ui/dashboard.html",
    "./web/ui/dashboard.html",
    "../web/ui/dashboard.html",
    "../../web/ui/dashboard.html",
}
```

**Manual smoke test** (if the service is running locally):
1. Open `http://localhost:8080/ui/` — should list seeded users
2. Click a user — should show profile, empty balances/wallets/transactions sections
3. Confirm existing endpoints (`/health`, `/core/v1/...`) still work

---

## Phase 4 — KYC Action

**Goal**: Implement the KYC trigger form and its POST handler. Selecting a user and a target state updates the user's KYC state in storage and fires the appropriate webhook.

### Handler: `UIKYCForm` (GET)

Serves `web/ui/kyc_action.html` pre-populated with all users.

- Calls `h.store.ListUsers()` to populate the user dropdown
- If `userID` query param is present (linked from user detail page), pre-select that user
- Template shows: current KYC state next to each user's name in the dropdown

### Handler: `UIKYCAction` (POST)

Form fields: `userID`, `kycState` (`action_required` | `rejected` | `accepted`), `riskLevel` (`low` | `medium` | `high`).

Logic:
1. Parse and validate form fields; reject unknown `kycState` values
2. `h.store.GetUser(userID)` → 400 if not found
3. Update `user.KYCState` and `user.RiskLevel`, call `h.store.UpdateUser(user)`
4. Build webhook data matching the real GateHub format:
   ```go
   statusCode := map[string]int{
       "accepted":        1,
       "rejected":        2,
       "action_required": 0,
   }[kycState]
   webhookData := map[string]interface{}{
       "gateway": "paywiser-eu-sandbox",
       "verified": map[string]interface{}{
           "short":  kycState,
           "status": statusCode,
       },
   }
   ```
5. Map `kycState` to `eventType` using `consts.WebhookEventKYC*` constants
6. `go h.webhookManager.SendAsync(eventType, userID, webhookData, 0)`
7. HTTP redirect to `/ui/users/{userID}?flash=kyc_triggered`

The user detail page should display the `flash` query param as a success banner (simple check: `if r.URL.Query().Get("flash") != ""`).

### Tests in `internal/handler/ui_test.go`

```
TestUIKYCForm_Renders                → GET /ui/actions/kyc — 200, body contains both seeded user emails
TestUIKYCForm_PreSelectsUser         → GET /ui/actions/kyc?userID=... — body contains "selected" for that user
TestUIKYCAction_AcceptsUser          → POST accepted state; verify user.KYCState == "accepted" in storage; response is redirect
TestUIKYCAction_RejectsUser          → POST rejected state; verify user.KYCState == "rejected"; response is redirect
TestUIKYCAction_ActionRequired       → POST action_required state; verify storage updated
TestUIKYCAction_UnknownUser          → POST with random UUID; expects 400
TestUIKYCAction_InvalidState         → POST with kycState="invalid"; expects 400
TestUIKYCAction_DefaultRiskLevel     → POST accepted with no riskLevel; defaults to "low"
```

> **Webhook dispatch verification**: The tests use `webhook.NewManager("", "test-secret", nil, nil, "")` (no queue). `SendAsync` with a nil queue is a no-op (check the existing implementation; if it panics on nil queue, wrap in a nil-check or pass a no-op manager). The tests verify the _storage state_ changed — that is the primary assertion. Webhook payload format is covered by manual testing against the real wallet backend.

### Phase 4 — Test gate

```bash
make test
```

All existing tests pass. 8 new KYC action tests pass. The form handler replaces its stub.

**Manual smoke test**:
1. Navigate to `http://localhost:8080/ui/` → click a user → "Trigger KYC Event"
2. Select `rejected`, submit
3. Verify the user detail page shows KYC state = `rejected`
4. Check wallet backend logs for the received webhook (confirms the `paywiser` gateway format is accepted)

---

## Phase 5 — Card Transaction Action

**Goal**: Implement the card transaction trigger form and its POST handler. The form lets the developer pick a user, pick one of their active cards, enter transaction details, and fire a simulated card transaction.

### Handler: `UICardTxCards` (GET — AJAX helper)

**Path**: `GET /ui/actions/card-transaction/cards?userID={id}`

Returns a JSON array of active cards for the given user (for populating the card dropdown via `fetch()` when the user selection changes):

```json
[
  { "id": "...", "maskedPan": "512345******2346", "status": "Active", "accountId": "..." }
]
```

Logic:
1. `h.store.GetCustomerBySourceID(userID)` → empty array (200) if no customer
2. `h.store.GetCardsByCustomer(*customer.ID)` → filter to `status == "Active"` only
3. `h.sendJSON(w, 200, cards)`

### Handler: `UICardTxForm` (GET)

Serves `web/ui/card_tx_action.html`.

- Calls `h.store.ListUsers()` to populate the user dropdown
- If `userID` query param is present, pre-select that user and eagerly fetch their cards (same logic as the AJAX helper) to populate the card dropdown without a round-trip
- Template includes a `<script>` block that calls the AJAX helper on user dropdown `change` events to reload the card options

### Handler: `UICardTxAction` (POST)

Form fields: `userID`, `cardID`, `amount` (string, e.g., `"50.00"`), `currency`, `txType` (int), `merchantName`, `merchantCity`, `merchantCountry`.

Logic (mirrors `CreateCardTransaction` handler in `internal/handler/cards.go`):
1. Validate all required fields; parse amount as float64
2. `h.store.GetCard(cardID)` → 400 if not found or not Active
3. `h.store.GetUser(userID)` → 400 if not found
4. `h.store.GetBalance(userID, currency)` → 400 if insufficient
5. Build a `models.CardTransaction` with a new UUID, set fields, status `"COMPLETED"`, ghResponseCode `"00"`
6. `h.store.CreateCardTransaction(tx)`
7. `h.store.AddCardTransactionIndex(cardID, tx.TransactionID)`
8. `h.store.DeductBalance(userID, currency, amount)`
9. Build and dispatch `cards.transaction.event` webhook:
   ```go
   webhookData := map[string]interface{}{
       "title":         "Card Purchase",
       "body":          fmt.Sprintf("%s %.2f at %s", currency, amountF, merchantName),
       "transactionId": tx.TransactionID,
       "cardId":        cardID,
   }
   go h.webhookManager.SendAsync(consts.WebhookEventCardTransaction, userID, webhookData, 0)
   ```
10. HTTP redirect to `/ui/users/{userID}?flash=card_tx_triggered`

### Template `web/ui/card_tx_action.html`

- User dropdown (onchange triggers `fetch("/ui/actions/card-transaction/cards?userID=" + val)` and re-renders card options)
- Card dropdown (populated dynamically; shows masked PAN + account ID)
- Amount field (numeric, required)
- Currency dropdown (all 11 supported currencies, default EUR)
- Transaction type dropdown: Purchase (0) | ATM Withdrawal (1) | Refund (20)
- Merchant name (text, optional)
- Merchant city (text, optional)
- Merchant country (text, 2-letter, optional)

### Tests in `internal/handler/ui_test.go`

```
TestUICardTxCards_NoCustomer          → user with no card customer → 200, empty JSON array
TestUICardTxCards_WithCards           → seeds customer + active card → card appears in response
TestUICardTxCards_FiltersInactive     → blocked card excluded; active card included

TestUICardTxForm_Renders              → GET form — 200, body contains user dropdown

TestUICardTxAction_HappyPath          → valid card, sufficient balance; transaction created in storage,
                                        balance deducted, response is redirect to user detail
TestUICardTxAction_InsufficientFunds  → balance < amount → 400
TestUICardTxAction_InactiveCard       → card.Status == "Blocked" → 400
TestUICardTxAction_UnknownCard        → random cardID → 400
TestUICardTxAction_UnknownUser        → random userID → 400
TestUICardTxAction_InvalidAmount      → amount = "not-a-number" → 400
TestUICardTxAction_DefaultsApplied    → merchantName empty → transaction created with empty merchant fields (not panic)
```

The happy-path test verifies the storage state directly:

```go
// After POST:
tx, err := store.GetCardTransaction(txID)
require.NoError(t, err)
assert.Equal(t, "COMPLETED", *tx.TxStatus)

bal, err := store.GetBalance(userID, "EUR")
require.NoError(t, err)
assert.InDelta(t, startBalance-amount, bal, 0.01)
```

### Phase 5 — Test gate

```bash
make test
```

All existing tests pass. ~10 new card transaction tests pass. Both card transaction handlers replace their stubs.

**Manual smoke test**:
1. Navigate to a user's detail page → "Trigger Card Transaction"
2. Select the user (pre-filled), select a card from the dropdown (loaded via AJAX)
3. Enter amount `50.00`, currency `EUR`, merchant `Test Store`
4. Submit → redirected to user detail page showing the new card transaction in the card's transaction list
5. Verify the balance decreased by `50.00`
6. Check wallet backend logs for the `cards.transaction.event` webhook

---

## Complete File Inventory

### New files

| File | Created in phase |
|------|-----------------|
| `internal/handler/ui.go` | Phase 2 (stubs), completed Phase 3–5 |
| `internal/handler/ui_test.go` | Phase 3 (created), extended Phase 4–5 |
| `web/ui/dashboard.html` | Phase 2 (placeholder), completed Phase 3 |
| `web/ui/user.html` | Phase 2 (placeholder), completed Phase 3 |
| `web/ui/kyc_action.html` | Phase 2 (placeholder), completed Phase 4 |
| `web/ui/card_tx_action.html` | Phase 2 (placeholder), completed Phase 5 |

### Modified files

| File | Modified in phase |
|------|------------------|
| `internal/storage/interface.go` | Phase 1a |
| `internal/storage/memory.go` | Phase 1b |
| `internal/storage/memory_test.go` | Phase 1b |
| `internal/storage/redis.go` | Phase 1c (new methods + `CreateTransaction` index) |
| `internal/storage/redis_test.go` | Phase 1c |
| `cmd/mockgatehub/main.go` | Phase 2a |
| `cmd/mockgatehub/main_test.go` | Phase 2a |
| `internal/auth/middleware.go` | Phase 2a |
| `internal/auth/middleware_test.go` | Phase 2a |

---

## Non-Goals / Out of Scope

- Pagination (mock scale doesn't need it)
- Real-time updates / WebSockets
- Balance adjustment from UI
- Search or filtering
- BDD E2E scenarios for the UI (can be added later as `features/ui.feature`)
- Dark mode

---

## Open Questions

1. **KYC webhook gateway field**: The wallet backend filters KYC webhooks by `data.gateway` containing `"paywiser"`. The UI sends `"paywiser-eu-sandbox"`. Verify this is accepted by watching backend logs after the first real trigger. If not, adjust to match the simpler format used by `UpdateKYCState` (i.e., `{"state": "accepted", "risk_level": "low"}`).

2. **`SendAsync` with nil queue**: When the webhook manager is constructed with `nil` as the queue (as in unit tests), `SendAsync` must not panic. Verify this before Phase 4 — if it does panic, wrap the call in a nil-queue guard or add a `NoopManager` helper for tests.

3. **In-memory `ListTransactionsByUser` performance**: Iterates all transactions and filters. Fine at mock scale (tens of records). Not a concern for this use case.

4. **Template file resolution in tests**: Handler tests call handlers directly without starting an HTTP server, so the working directory may not contain `web/ui/`. Tests should verify handler status codes and parse the response body for key content. If the template cannot be found, the handler returns 500 — tests will fail clearly. Consider an alternative: embed templates in the binary using `//go:embed web/ui/*.html` if filesystem resolution proves unreliable in the test environment. This is an optional refinement.
