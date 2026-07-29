# Perf provisioner redesign — implementation plan & handoff

Status: **research complete, design agreed, implementation NOT started.** This
doc is self-contained; you should not need to re-run the investigation.

## Goal (what the user asked for)

Redesign `go/performance` provisioning to be as simple as possible:

- CLI takes a **comma-separated list of countries**, e.g. `-countries za,de,us`.
- Default **100 wallets per country**, always funded.
- **No senders/receivers concept** in the provisioner — it just creates wallets.
- **Deterministic wallet names** (so re-runs are stable).
- Fund each wallet with **5000 units** of that country's currency (major units).
  (Originally 10 000; lowered to 5000 to stay under USD/PTI KYC limits.)
- **Idempotent**: on re-run, create the wallet only if it doesn't exist, then
  check its balance and top up only the shortfall to reach 5000.

Countries in scope now: **za→ZAR (Xago), de→EUR (GateHub), us→USD (PTI)**. Build
all three in one pass.

## Why this is non-trivial

Funding is provider-specific; there is no uniform "fund N units" RPC. Each
currency uses a different provider + mechanism. All three were confirmed
**doable headlessly** (no browser). Complexity: ZAR trivial, EUR medium, USD heavy.

## Current codebase state

- `go/performance/` is an untracked new dir (`?? go/performance/` in git status).
- Existing provisioner (`provision/provision.go`) creates **ZA/ZAR only**, funds
  via Xago, splits into senders/receivers, writes an overlay via `scenario.go`.
- A recent fix already landed in `provision/provision.go` `createWallet`: it
  tolerates `WALLETS_WALLET_CONFLICT` because the wallets gRPC middleware
  auto-creates the user's default wallet (with the profile country) on the first
  authenticated call. **Keep that behavior** in the rewrite.
- Runner (`runner/`, `config/config.go`) consumes `senders:` + `receivers:` and
  requires at least one of each (`config.go` `validate()`).

## Agreed design

### CLI (`cmd/perf/main.go` `cmdProvision`)
- Replace `-count`/`-senders`/`-country`/`-asset`/`-asset-scale` with:
  - `-countries` (string, comma-separated, e.g. `za,de,us`) — required.
  - `-per-country` (int, default `100`).
  - `-target` (int, default `5000`) — target balance in **major** units.
  - keep `-prefix` (optional, default empty), `-password`, `-target`(grpc addr →
    rename to `-grpc`/keep `-target` as address? currently `-target` is the grpc
    address; pick a non-colliding name, e.g. keep `-grpc-address`), `-kratos`,
    `-dsn`, `-out`.
- Always fund (drop `-fund`/`-deposits`).

### Naming (deterministic)
- Label: `<country>-<index3>`, e.g. `za-001` … `za-100`, `de-001`, `us-001`.
- Email: `<label>@perf.interledger.test`.
- Phone: keep the existing `+2782%07d`-style generator but make it unique across
  countries (e.g. seed the numeric suffix from a global running index, or prefix
  per country) — phone numbers must be globally unique or Kratos/ signup rejects.
- Wallet address: `https://local.ilp.link/<label>`. **Change the default address
  host from `https://ilp.link` to `https://local.ilp.link`** to match the local
  `PAYMENT_POINTER_BASE` (`local.ilp.link`). The old `ilp.link` default looks wrong
  for local.

### Output (`provision/scenario.go`) + runner (`config/config.go`)
- Provisioner writes a **flat `wallets:` list** (label, email, password,
  wallet_address) — drop `Split` and senders/receivers from the provisioner.
- Extend `config/config.go`:
  - Add `Wallets []WalletEntry` field (`yaml:"wallets"`), each `{label, email,
    password, wallet_address}`.
  - In `applyDefaults()` (or a new expand step in `Load`), if `Wallets` is
    non-empty, expand every wallet into **both** `Senders` (label/email/password)
    and `Receivers` (label/wallet_address), unless senders/receivers were given
    explicitly. This keeps the runner unchanged and hand-written scenarios still
    work.
- Update the committed scenario files only if needed (they still use
  senders/receivers directly — fine).

### Idempotency (per wallet)
1. Register in Kratos. If the identity already exists, **log in instead** using
   the existing `auth.Client.Login(email, password)` (`auth/auth.go:48`,
   `CreateNativeLoginFlow`). Detect "already exists" from the register error.
2. Ensure the platform wallet exists (middleware auto-creates it; the
   `createWallet` conflict-tolerant path already handles this).
3. Ensure the provider linked account + KYC exist — **skip creation if already
   present** (check via `GetBalances`/provider list before creating).
4. Read current balance for the wallet's currency; **top up only the shortfall**
   to reach the target. If already ≥ target, do nothing.

Report per-wallet outcome as notes (like the existing `fund()` returns notes) so
one wallet/provider failing doesn't abort the whole run.

## Exact API surface (verified — file:line)

### Perf client wiring
- `client/client.go`: `Pool`, `Pool.Wallet(label, token) *Wallet`, `Wallet.ctx()`
  attaches bearer token + timeout. Add new RPC wrappers in `client/provision.go`.
- `client.Classify(stage, err) *Failure` exposes `.AppCode` (string alias of
  `errcodes.AppErrorCode`) — use `errcodes.ErrCodeWalletsWalletConflict` etc.
- `auth.Client.Login(email, password) (token, error)` already exists.
- `currency.ParseCurrency("EUR").Scale()` gives asset scale
  (`go/backend/currency/amount.go`: USD/EUR scale 2, ZAR scale 2).

### Backend gRPC RPCs (all on `pb.BackendServiceClient`, `go/proto/backend/v1`)
- `GetBalances(Empty) -> GetBalancesResponse{Balances []Balance}`; `Balance{Balance
  *Amount, Currency string, LinkedAccount string}` — already wrapped
  (`client/client.go:150`).
- `Amount{Amount int64, Asset string, AssetScale int32, Country string}`.
- **Xago (ZAR):** `AddXagoBalanceAccount(AddXagoBalanceAccountRequest{CurrencyCode,
  Nickname, Title}) -> LinkedAccount{Id,...}`; `DepositTestXago(Empty)` — no
  amount, fixed credit into the wallet's Xago sub-account; loop until balance ≥
  target. Both already wrapped in `client/provision.go`. Handler is non-prod
  gated (`go/backend/grpc/xago.go:246`).
- **GateHub (EUR):** `GetGatehubOnboardingWidget(Empty) -> GatehubWidget` — the
  gRPC call itself creates the managed user + EUR "EUR Balance" linked account +
  Pacioli account as a side-effect (`go/backend/grpc/gatehub.go:10`,
  `providers/gatehub/ops/ops.go:81`, `ops/workflows.go:17`). EU-country gated, NOT
  KYC gated. (Admin `CreateGatehubUser(walletID)` is an alternative but is on the
  admin service.)
- **PTI (USD):**
  - `GetKYCProviderWidget(GetKYCProviderWidgetRequest{IdempotencyKey}) ->
    KYCProviderWidget{Provider, PtiWidget *PtiWidget{UserId, RequestId, ...}}` —
    for US this creates the PTI user (`go/backend/grpc/kyc.go:93`). Capture
    `PtiWidget.UserId` (=external user id) and `RequestId`.
  - `SetKYCStatusPending(Empty)` — spawns `CreateWalletWorkflow` for US, which
    retries `CheckUserAssessmentAccepted` until the assessment is accepted
    (`go/backend/grpc/kyc.go:378`, `kyc/ops/workflows.go:89`). Call BEFORE
    accepting the assessment (once accepted, status flips to Level2 and
    SetKYCStatusPending is a no-op).
  - `CreatePtiBankAccount(CreatePtiBankAccountRequest{BankName, AccountNumber,
    RoutingNumber, AccountType}) -> LinkedAccount` — the ACH source. Gated by
    `feats.BanksEnabled` (a per-wallet DB feature flag `banks_enabled`,
    `go/backend/features/types.go:8`; verify it's on in local — if off, US bank
    creation fails and you must enable it or `SetFeatures`).
  - `DepositBalance(TransferBalanceRequest{FromLinkedAccount, ToLinkedAccount,
    Amount *Amount, Note}) -> Payment` — creates the deposit payment; enforces KYC
    limits (`go/backend/grpc/deposit.go:42`). Amount is **minor units** (5000 USD
    → `500000`).
  - `PtiCreateDeposit(PtiCreateDepositRequest{Id, IpAddress *string}) -> Empty` —
    `Id` is the **payment id** from `DepositBalance` (not a linked-account id);
    `IpAddress` unused. Executes the deposit (`go/backend/grpc/deposit.go:159`).
  - `GetPtiBalances(Empty) -> GetPtiBalancesResponse{Balances []PtiBalance{Balance
    *Amount, Currency, LinkedAccount, ...}}` — poll for the balance to appear/grow.

### Mock HTTP calls (needed for EUR create-credit and USD assessment)
Reachability: from **inside compose** use `http://mockgatehub:8080` /
`http://mockpti:8080`; from the **host** use `https://mockgatehub.interledger.test`
/ `https://mockpti.interledger.test` (self-signed TLS → use an
`InsecureSkipVerify` client + `make hosts` entries). The perf tool runs on the
host today (`localhost:8443` gRPC, `localhost:4433` Kratos), so use the Traefik
hostnames with a TLS-skip client, OR document running the tool inside compose.

- **GateHub deposit (EUR)** — `POST /core/v1/transactions`:
  - Headers: `Content-Type: application/json`,
    `x-gatehub-managed-user-uuid: <external_id>`, plus **HMAC auth** (enforced in
    local): `x-gatehub-app-id: local-test-app-id`,
    `x-gatehub-timestamp: <unix millis>`,
    `x-gatehub-signature: GenerateSignature(ts, "POST", <fullURL>, <body>, "local-test-app-secret")`.
  - Signature format (copy verbatim from
    `go/mock/mockgatehub/internal/auth/signature.go:16` `GenerateSignature`):
    `hex(HMAC_SHA256("<ts>|POST|<fullURL>|<body>", secret))`, then
    `strings.Trim(msg, "|")`. The middleware checks the **full URL** first, then
    falls back to a **path-only** signature (`middleware.go:131-138`), so signing
    over the exact URL string you request (scheme+host+path) works; if unsure,
    the path-only fallback (`/core/v1/transactions`) is the safety net.
  - Body (amount is a **float in major EUR units**, NOT minor):
    `{"type":1,"deposit_type":"external","receiving_address":"<addr>","amount":5000.00,"currency":"EUR"}`
    (`type:1` = deposit; `deposit_type:"external"`). `receiving_address` = the
    gatehub linked account `provider_id` (mock XRPL address); the wallet is
    resolved from the header so the exact value is echoed/stored only.
  - Credit is **async**: mock delays webhook `webhook_min_delay_sec: 2`
    (`local/config/mockgatehub.yaml`) → POSTs `core.deposit.completed` to
    `http://backend:8080/webhooks/gatehub` → backend runs `CreateGatehubDeposit`
    Temporal workflow → credits Pacioli. **Poll `GetBalances` (EUR "EUR Balance"
    linked account) until it reflects.** Requires the Temporal `backend` worker
    running. Default deposit fee 0 → full 5000 credited.
  - Confirmed config: `local/config/mockgatehub.yaml`
    (`enforce_authentication: true`, `valid_credentials: local-test-app-id:
    local-test-app-secret`, `webhook_secret: 6d6f636b...`,
    `webhook_url: http://backend:8080/webhooks/gatehub`); backend
    `local/config/backend.yaml` gatehub `app_id: local-test-app-id`,
    `secret: local-test-app-secret`, matching `webhook_secret`,
    `api_base_url: http://mockgatehub:8080`.

- **PTI assessment accept (USD)** — `POST /users/assessments`:
  - Header: `x-pti-client-id: 04d3e1b5-96d4-47e4-9eaa-13e9b4b0f219`
    (`local/config/backend.yaml` pti.client_id / `local/config/mockpti.yaml`).
  - Body: `{"id":"<externalUserID>","type":"PERSON"}` → mock stores
    `Assessment:"ACCEPTED"` and fires the `USER_ASSESSMENT` webhook
    (`go/mock/mockpti/internal/handler/assessment.go:19`). This unblocks the
    pending `CreateWalletWorkflow` → USD balance linked account gets created.
  - mockpti deposits auto-**SETTLE** and webhook back to
    `http://backend:8080/webhooks/pti`; you do NOT POST to
    `/transactions/deposits` yourself — the `PtiCreateDeposit` RPC drives it.

### DB reads (provisioner already holds `*sql.DB` on `BackendDSN`, `provision.go:113`)
- GateHub managed-user UUID:
  `SELECT gu.external_id FROM gatehub_users gu JOIN user_wallets uw ON
  uw.wallet_id = gu.wallet_id WHERE uw.user_id = $1 LIMIT 1`
  (mirrors `e2e/db_helpers.go:146`).
- GateHub receiving address (optional):
  `SELECT provider_id FROM linked_accounts WHERE wallet_id = <ghWalletID> AND
  provider='gatehub' AND type='balance' AND deleted_at IS NULL`.
- KYC approval (ZA/EU still need it to transact — US gets Level2 from the
  assessment, but likely also wants the approved row): existing `approveKYC`
  writes `wallet_kyc_status (wallet_id, status=3)` (`provision.go:342`). Reuse for
  ZA/EU. For US, the PTI workflow drives status; confirm whether the approved row
  is still needed for sending payments and add it if so.

## Per-provider funding recipes (ordered)

### ZAR / za (Xago) — reuse existing `fund()`
1. `AddXagoBalanceAccount{CurrencyCode:"ZAR", Nickname/Title:"ZAR Balance"}`.
2. Loop `DepositTestXago()` until `GetBalances` ZAR ≥ target (each call credits a
   fixed amount; existing `waitForBalance` pattern, `provision.go:393`).

### EUR / de (GateHub)
1. Create US wallet-equivalent but country `DE`, asset `EUR`.
2. gRPC `GetGatehubOnboardingWidget(Empty)` (as the wallet owner) → creates
   managed user + EUR linked account. (Idempotent: safe to call if exists.)
3. DB: read `gatehub_users.external_id` + linked-account `provider_id`.
4. Signed HTTP `POST /core/v1/transactions` (see above) with `amount: <target>.00`.
5. Poll `GetBalances` EUR until ≥ target (≈2s+ webhook + workflow).
6. KYC approve via DB write so it can transact.

### USD / us (PTI) — heaviest
1. Create wallet country `US`, asset `USD`.
2. gRPC `GetKYCProviderWidget{IdempotencyKey:<uuid>}` → capture
   `PtiWidget.UserId` (externalUserID), `RequestId`.
3. gRPC `SetKYCStatusPending(Empty)` → spawns `CreateWalletWorkflow` (retries).
4. HTTP `POST /users/assessments` `{id:externalUserID, type:"PERSON"}` with
   `x-pti-client-id` → assessment ACCEPTED, unblocks balance-account creation.
5. Poll `GetPtiBalances` until the USD balance linked account exists → capture
   its `LinkedAccount` id (`balanceLA`).
6. gRPC `CreatePtiBankAccount{BankName, AccountNumber, RoutingNumber,
   AccountType:"CHECKING"}` → capture `bankLA` (needs `BanksEnabled`).
7. gRPC `DepositBalance{FromLinkedAccount:bankLA, ToLinkedAccount:balanceLA,
   Amount:{Amount: target*100, Asset:"USD", AssetScale:2}, Note}` → capture
   `paymentID`.
8. gRPC `PtiCreateDeposit{Id: paymentID}` → executes; mockpti auto-settles →
   webhook → Pacioli credit.
9. Poll `GetPtiBalances` until ≥ target. If `DepositBalance` rejects on KYC
   limits, split into several smaller `DepositBalance`+`PtiCreateDeposit` cycles
   (5000 was chosen to avoid this, but keep the split logic as a fallback and
   surface the limit message if it still fails).
10. KYC approve DB write if needed for sending.

## Proposed file layout

- `provision/countries.go` (new): `type spec struct{ currency string; scale int32;
  provider string }`; `map[string]spec` keyed by lowercase country code (za/de/us,
  extensible); label/email/phone/address helpers.
- `provision/funder.go` (new): `type Funder interface { Ensure(ctx, w, target)
  (balanceMinor int64, notes []string, err error) }`; a selector by provider.
- `provision/fund_xago.go`, `fund_gatehub.go`, `fund_pti.go` (new): one per
  provider (or a single `funders.go`).
- `provision/mocks.go` (new): host/in-network base URL resolution + a signed
  mockgatehub client (copy `GenerateSignature`) + a mockpti client
  (`x-pti-client-id` header). Add a `-mock-gatehub-url` / `-mock-pti-url` flag or
  derive from a `-in-network` bool (default host: Traefik hostnames + TLS skip).
- `provision/provision.go` (rewrite): `Options{Countries []string, PerCountry int,
  TargetMajor int64, ...}`; `Run` loops countries × index, calls `provisionOne`
  (create-or-login → ensure linked account/KYC → fund via `Funder`), collects a
  flat `[]Wallet`.
- `provision/scenario.go` (rewrite): `WriteWallets(w, []Wallet)` → flat `wallets:`.
- `config/config.go`: add `Wallets` + expansion into senders/receivers.
- `cmd/perf/main.go` `cmdProvision`: new flags; drop `Split`.
- `Makefile`: `provision: $(BIN) provision -countries $(COUNTRIES) ...` with
  `COUNTRIES ?= za,de,us`.

## Risks / caveats to handle

1. **Async funding** (EUR + USD): always poll with a deadline; require Temporal
   `backend` worker running. Return a clear note on timeout.
2. **HMAC signing** (EUR): must match `GenerateSignature`; path-only fallback is
   the safety net. Test against a running mockgatehub early.
3. **Host reachability of mocks**: mockgatehub has no host port — from the host
   you must go through Traefik (`https://mockgatehub.interledger.test`, TLS skip +
   hosts entry). Simplest alternative: run the provisioner inside compose so
   `http://mockgatehub:8080` / `http://mockpti:8080` work directly. Decide and
   document.
4. **USD KYC limits**: 5000 chosen to stay under; keep split-and-retry + surface
   the backend limit message if it still rejects.
5. **`BanksEnabled`** must be on for the US wallet (`CreatePtiBankAccount`). If
   off by default in local, call `SetFeatures` or flip the DB flag first.
6. **Phone uniqueness** across 300 wallets — generate globally-unique E.164.
7. **Idempotency**: register→login fallback; skip provider setup if the linked
   account already exists; top-up only the shortfall.
8. Keep the existing `WALLETS_WALLET_CONFLICT`-tolerant `createWallet` logic.

## Verification

- `cd go && GOWORK=off go build ./performance/... && go test ./performance/...`
  and `gofmt -l`.
- Live smoke (needs full local stack `make all` + monitoring optional): run
  `perf provision -countries za,de,us -per-country 1 -target 5000` and confirm one
  funded wallet per currency (check balances, and the generated `wallets.local.yaml`).
- Then a `perf run` with a committed scenario layered over the overlay.

## Reference: prior research lives in the conversation

Two Explore agents produced detailed, cited recipes for GateHub (EUR) and PTI
(USD); their key facts are captured above. The originals were thorough — if a
detail here is ambiguous, re-derive from the cited file:line locations.
