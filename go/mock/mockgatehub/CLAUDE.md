# MockGatehub — Development Guide

Comprehensive guidance for working on the MockGatehub project.

## Official API Documentation

**Always consult the official GateHub API docs when uncertain about endpoint behavior, request/response formats, or field semantics:**
https://docs.gatehub.net/api-documentation/c3OPAp5dM191CDAdwyYS

When implementing or modifying endpoints, check the official docs first to ensure MockGatehub matches real GateHub behavior.

## Project Context

MockGatehub is a Go mock implementation of the GateHub API, designed to support local development and testing of wallet applications that integrate with GateHub. It removes the dependency on real credentials and services, enabling fully local development, automated testing without API rate limits, and predictable behavior for CI/CD pipelines.

### Critical Constraints

1. **API Compliance**: MockGatehub must be a drop-in GateHub replacement. Applications expect exact API format compliance.
2. **Sandbox Parity Only**: Focus on happy paths and sandbox behavior. Production GateHub features are out of scope.
3. **Multi-Currency Required**: Support all 11 currencies (XRP, USD, EUR, GBP, ZAR, MXN, SGD, CAD, EGG, PEB, PKR).
4. **Immutable Vault UUIDs**: Vault identifiers are hardcoded in `internal/consts/consts.go` and must never change (application databases store these).
5. **Redis Required for Webhooks**: The webhook queue always needs Redis, even in in-memory storage mode.

## Architecture Overview

### Tech Stack

- **Language**: Go 1.24+
- **HTTP Router**: chi v5 (lightweight, idiomatic)
- **Storage**: Dual backend (memory for tests, Redis for runtime)
- **Webhook Queue**: Redis-backed sorted-set job queue with background worker
- **Logging**: Zap structured logging
- **Testing**: testify (unit), godog/Cucumber (BDD E2E)
- **Containerization**: Docker multi-stage build
- **CI/CD**: GitHub Actions with semantic-release

### Directory Structure

```
mockgatehub/
├── cmd/mockgatehub/               # Application entry point
│   ├── main.go                    # Server setup, routing, startup
│   └── main_test.go               # Route registration tests
├── internal/                      # Private application code
│   ├── auth/                      # HMAC signature generation/validation
│   │   ├── middleware.go          # HTTP auth middleware (public endpoint whitelist)
│   │   ├── middleware_test.go
│   │   ├── signature.go           # Signature generation and verification
│   │   └── signature_test.go
│   ├── config/                    # Centralized configuration
│   │   └── config.go             # Environment variable loading with defaults
│   ├── consts/                    # Constants
│   │   └── consts.go             # Currencies, vault IDs, rates, statuses, events
│   ├── handler/                   # HTTP request handlers
│   │   ├── handler.go            # Handler struct, RequestLogger, health check, root/iframe handlers
│   │   ├── handler_test.go       # Handler tests
│   │   ├── helpers.go            # sendJSON, sendError, setCORSHeaders, decodeJSON
│   │   ├── auth.go               # /auth/v1 endpoints (tokens, managed users)
│   │   ├── identity.go           # /id/v1 endpoints (KYC flow, iframe)
│   │   ├── core.go               # /core/v1 endpoints (wallets, transactions)
│   │   ├── cards.go              # /cards/v1 endpoints (full card lifecycle)
│   │   ├── fees.go               # Fee configuration and admin endpoints
│   │   ├── fees_test.go          # Fee calculation and API tests
│   │   └── rates.go              # /rates/v1 endpoints (rates, vaults)
│   ├── logger/                    # Logging setup
│   │   └── logger.go            # Zap dual-core logger (stdout + stderr)
│   ├── models/                    # Domain & API models
│   │   ├── models.go            # User, Wallet, Transaction
│   │   ├── api.go               # Request/response DTOs
│   │   ├── messages.go          # Complex response types (user profile, wallet balance)
│   │   └── cards.go             # Card-related models (Customer, Card, Account, etc.)
│   ├── storage/                   # Storage layer
│   │   ├── interface.go         # Storage contract (~30 methods)
│   │   ├── memory.go            # In-memory implementation (sync.RWMutex)
│   │   ├── memory_test.go
│   │   ├── redis.go             # Redis implementation (JSON + atomic ops)
│   │   ├── redis_test.go
│   │   └── seeder.go            # Pre-seeds test users with balances
│   ├── utils/                     # Utilities
│   │   └── utils.go             # UUID, mock XRPL address, transaction hash
│   └── webhook/                   # Webhook delivery system
│       ├── manager.go            # Webhook manager (enqueue, send, sign)
│       ├── manager_test.go
│       ├── queue.go              # Redis sorted-set backed job queue
│       ├── worker.go             # Background worker (polls every 5s)
│       └── job.go                # Job struct and serialization
├── features/                      # Gherkin BDD feature files (8 files)
│   ├── auth_user_kyc.feature
│   ├── cards.feature
│   ├── fee_configuration.feature
│   ├── organization_configuration.feature
│   ├── rates_and_vaults.feature
│   ├── service_health.feature
│   ├── signature_authentication.feature
│   ├── transactions.feature
│   └── wallets_and_balances.feature
├── test/integration/              # Go integration tests
│   └── integration_test.go
├── testenv/                       # E2E test environment
│   ├── docker-compose.yml        # Redis (26380) + MockGatehub (25151)
│   ├── godog_test.go             # Godog BDD runner (//go:build e2e)
│   ├── *_steps.go                # Step definitions for each feature area
│   ├── helpers.go                # Test helpers
│   ├── http_client.go            # Signed HTTP client for tests
│   ├── services.go               # Docker service management
│   ├── test_context.go           # Shared test state
│   ├── types.go                  # Test constants
│   └── README.md                 # Test environment documentation
├── web/                           # Static web assets
│   ├── index.html                # Deposit/withdraw iframe
│   └── kyc-iframe.html           # KYC onboarding iframe
├── docs/                          # GateHub API reference docs
├── Dockerfile                     # Multi-stage Docker build
├── Makefile                       # Build, test, lint, clean targets
├── .releaserc.json               # Semantic-release configuration
├── go.mod                         # Go module (chi, redis, godog, zap, testify, uuid)
├── README.md                      # User-facing project documentation
└── CLAUDE.md                      # This file
```

### Design Principles

1. **Dependency Injection**: Handler receives storage and webhook manager via constructor
2. **Interface-Based Storage**: Enables swapping memory/Redis without code changes
3. **Table-Driven Tests**: Use testify assertions with comprehensive coverage
4. **Idiomatic Go**: Standard project layout, effective Go patterns
5. **Minimal Dependencies**: chi, redis, uuid, testify, zap, godog

## Core Systems

### 1. Configuration (`internal/config/`)

All environment variables are centralized in `config.Load()`:

| Env Var | Default | Notes |
|---------|---------|-------|
| `MOCKGATEHUB_PORT` | `8080` | HTTP server port |
| `LOG_LEVEL` | `info` | Zap log level |
| `MOCKGATEHUB_REDIS_URL` | `""` | Enables Redis storage if set |
| `MOCKGATEHUB_REDIS_DB` | `0` | Redis database number |
| `WEBHOOK_URL` | `""` | Webhook delivery target |
| `WEBHOOK_SECRET` | `mock-secret` | Webhook HMAC signing secret |
| `WEBHOOK_MIN_DELAY_SEC` | `0.05` → clamped ≥ `2` | Minimum delay before webhook delivery |
| `MOCKGATEHUB_ENFORCE_AUTHENTICATION` | `true` | Enable/disable HMAC middleware |
| `MOCKGATEHUB_VALID_CREDENTIALS` | `local-test-app-id:local-test-app-secret` | Format: `appId:secret,appId2:secret2` |
| `DEFAULT_ORGANIZATION_ID` | `default-org` | Organization ID for callback routing |

### 2. Storage Layer (`internal/storage/`)

**Interface** (`interface.go`) defines ~30 methods covering:
- Users: CRUD operations
- Card Customers: CRUD + lookup by source ID
- Card Accounts: CRUD
- Customer Delivery Addresses: create + list
- Cards: CRUD + list by customer/account + limits
- Card Transactions: CRUD + index by card
- Wallets: CRUD + list by user
- Transactions: CRUD + status update
- Balances: get/add/deduct per user per currency
- 3DS Challenges: CRUD + list pending
- Organizations: CRUD for organization configuration

**Memory** (`memory.go`): `sync.RWMutex` for thread safety, maps for all entities.

**Redis** (`redis.go`): JSON serialization, `INCRBYFLOAT` for atomic balance updates. Keys follow patterns like `user:{id}`, `wallet:{address}`, `balance:{userID}:{currency}`.

**Seeder** (`seeder.go`): Pre-creates two test users:
- User 1: `00000000-0000-0000-0000-000000000001`, `testuser1@mockgatehub.local`, 10,000 USD, KYC `action_required`
- User 2: `00000000-0000-0000-0000-000000000002`, `testuser2@mockgatehub.local`, 10,000 EUR, KYC `action_required`

### 3. Authentication (`internal/auth/`)

**Signature Format**: `HMAC-SHA256(timestamp|method|full_url|body, secret)`, hex-encoded. Empty segments are stripped.

**Webhook Signatures**: Different algorithm — `HMAC-SHA256(json_body, hex_decoded_secret)` via `GenerateGateHubWebhookSignature`.

**Middleware** (`middleware.go`): Applied globally. Skips public endpoints:
- `/health`, `/`, `/iframe/onboarding`, `/iframe/submit`, `/transaction/complete`, `/api/user-currencies`, `/admin/fees`

Validates `x-gatehub-app-id`, `x-gatehub-timestamp`, `x-gatehub-signature` headers. Reconstructs full URL using `X-Forwarded-Proto`/`X-Forwarded-Host` for proxy environments.

### 4. Webhook System (`internal/webhook/`)

**Architecture**: Redis-backed job queue with a background worker and organization-aware callback routing.

- **Queue** (`queue.go`): Redis sorted set (`webhooks:queue`) + per-job hash (`webhooks:job:{id}`). Jobs are scored by `NotBefore` timestamp.
- **Worker** (`worker.go`): Polls every 5 seconds, processes batches of 10 ready jobs.
- **Manager** (`manager.go`): `SendAsync(eventType, userID, data, offsetDelaySeconds)` enqueues jobs. `send()` performs HTTP delivery with HMAC signing. Supports organization-aware callback routing and 2FA callback test workflows.
- **Job** (`job.go`): Tracks ID, event type, user ID, data, attempts, status, timestamps.

**Callback Routing**: The webhook manager resolves the callback URL per event:
1. Looks up the organization config using `DEFAULT_ORGANIZATION_ID`
2. If the organization has an `apiBaseUrl`, routes callbacks there
3. Falls back to the global `WEBHOOK_URL` environment variable

**2FA Callback Test**: When an organization config is updated via `PATCH /auth/v1/users/organization/{orgID}`, a test job is enqueued with a 3-second delay. The worker executes INITIATE and VERIFY callbacks to the organization's `apiBaseUrl` to verify connectivity.

**Retry policy**: 10 max attempts, 30-second fixed backoff between retries.

**Payload format**:
```json
{
  "uuid": "...",
  "timestamp": "1768920404045",
  "event_type": "core.deposit.completed",
  "user_uuid": "...",
  "environment": "sandbox",
  "data": { ... }
}
```

### 5. Handler Structure (`internal/handler/`)

```go
type Handler struct {
    store          storage.Storage
    webhookManager *webhook.Manager
    tokenToUser    sync.Map     // Maps bearer tokens → user UUIDs
    feeConfig      *FeeConfig   // Thread-safe fee percentages
}
```

Constructor: `NewHandler(store, webhookManager)` — initializes with `NewFeeConfig()` (0% defaults).

**Handler files**:
- `handler.go` — Handler struct, RequestLogger, HealthCheck, RootHandler, TransactionCompleteHandler, processDeposit, processWithdrawal, GetUserCurrencies
- `helpers.go` — sendJSON, sendError, setCORSHeaders, sendJSONWithCORS, sendErrorWithCORS, decodeJSON
- `auth.go` — CreateToken, CreateManagedUser, GetManagedUser, UpdateManagedUserEmail
- `identity.go` — GetUser, StartKYC, UpdateKYCState, OverrideRiskLevel, KYCIframe, KYCIframeSubmit
- `core.go` — CreateWallet, GetUserWallets, GetWallet, GetWalletBalance, CreateTransaction, GetTransaction
- `cards.go` — Full card lifecycle (~20 handlers)
- `fees.go` — FeeConfig, GetFees, SetFees, CalculateFee
- `rates.go` — GetCurrentRates, GetVaults

### 6. Fee System (`internal/handler/fees.go`)

Thread-safe `FeeConfig` with `depositFeePercent` and `withdrawalFeePercent` (default 0%). Admin API at `GET/PUT /admin/fees` for runtime configuration. Validates 0–100 range.

Fee logic: `CalculateFee(amount, percent) = round(amount * percent / 100, 2)`
- Deposits: net credited = amount - fee
- Withdrawals: total deducted = amount + fee
- Hosted transfers: always 0% fee

### 7. Transaction Lifecycle

1. `POST /core/v1/transactions` creates transaction in `pending` (status `1`)
2. If webhook URL configured: sends pending webhook immediately, then after ~2s goroutine delay marks `completed` (status `100`), credits balance, sends completed webhook
3. If no webhook URL: completes synchronously

**Transaction types**: `0`=Withdrawal, `1`=Deposit, `2`=Hosted

**Transaction statuses**: `1`=Pending, `100`=Completed, `3`=Failed

### 8. KYC Flow

1. `POST /id/v1/users/{userID}/hubs/{gatewayID}` — sets user to `action_required`
2. `GET /iframe/onboarding?bearer={token}[&user_id={uuid}]` — serves `web/kyc-iframe.html`
3. `POST /iframe/submit` — parses multipart or URL-encoded form, updates user to `accepted`, sends `id.verification.accepted` webhook with 2s offset delay
4. Iframe posts `{ type: 'OnboardingCompleted', value: JSON.stringify({ applicantStatus: 'submitted' }) }` to parent window

Token → user mapping: `/auth/v1/tokens` stores bearer → user UUID. The iframe submit handler resolves `user_id` from this mapping if not provided in the form.

### 9. Constants (`internal/consts/consts.go`)

**Webhook events**: `id.verification.accepted`, `id.verification.rejected`, `id.verification.action_required`, `core.deposit.completed`, `cards.card.created`, `cards.transaction.event`, `cards.3ds.auth_3ds_confirmation`

**Card statuses**: `Active`, `Blocked`, `TemporaryBlocked`, `Replaced`, `SoftDelete`, `AccountBlocked`, `InCreation`

## Testing

### Test Structure

| Location | Type | Description |
|----------|------|-------------|
| `internal/*/` | Unit | Package-level tests (handler, auth, storage, webhook, fees) |
| `test/integration/` | Integration | Full user journey, API compliance, multi-currency |
| `testenv/` | E2E (BDD) | Godog/Cucumber tests driven by `features/*.feature` |

### Running Tests

```bash
# All tests (unit + integration)
make unit-tests

# BDD E2E tests (builds Docker, runs against containers)
make e2e-tests

# Both
make test

# Coverage
make coverage
```

### E2E Test Environment

`testenv/docker-compose.yml` runs:
- Redis on port `26380`
- MockGatehub on port `25151` (built from Dockerfile, auth enforced)

Tests use `//go:build e2e` tag. Each feature area has a `*_steps.go` file implementing step definitions.

### Feature Files (~55 scenarios across 8 files)

- `service_health.feature` — health endpoint
- `auth_user_kyc.feature` — user creation, KYC flow
- `wallets_and_balances.feature` — wallet provisioning
- `transactions.feature` — deposits, hosted transfers, formatting
- `rates_and_vaults.feature` — exchange rates, vault listing
- `cards.feature` — full card lifecycle (~17 scenarios)
- `signature_authentication.feature` — HMAC validation (~12 scenarios)
- `fee_configuration.feature` — fee setup and application (~16 scenarios)
- `organization_configuration.feature` — organization config update (~6 scenarios)

## CI/CD

### GitHub Actions Workflows

**PR Validation** (`pr-validation.yml`):
- Validates PR title (Conventional Commits)
- Runs unit and E2E tests
- Docker build (no push) with GHA cache

**Release** (`release.yml`), runs on push to `main`:
- All tests (unit + E2E)
- semantic-release to determine version and create tag
- Docker build and push to `ghcr.io/interledger/mockgatehub` (linux/amd64 + linux/arm64)

### Semantic Versioning

- `feat` → minor release
- `fix`, `perf`, `docs`, `refactor`, `build`, `ci` → patch release
- `chore`, `test` → no release
- Tag format: `v${version}`
- Config: `.releaserc.json`

## Development Workflow

### Making Changes

```bash
go mod tidy
go test ./...
go build ./cmd/mockgatehub
```

### Docker Build & Test

```bash
docker build -t local-mockgatehub .
make e2e-tests
```

## Logging

It's safe to log sensitive values in MockGatehub — this is a development/testing mock, not a production system.

Use zap structured logging:
```go
logger.Info("deposit created",
    zap.String("user_id", userID),
    zap.String("amount", amountStr),
    zap.String("currency", currency),
)
```

Log levels:
- `Info` — normal operations (user created, deposit received)
- `Warn` — non-fatal issues (invalid input, fallback behavior)
- `Error` — failed operations (database error, webhook failure)
- `Debug` — detailed diagnostics (token resolution, balance calculations)

## When Adding New Endpoints

1. Define models in `internal/models/api.go` or `internal/models/cards.go`
2. Implement handler in `internal/handler/{domain}.go`
3. Register route in `cmd/mockgatehub/main.go` `setupRoutes()`
4. Write unit tests in `internal/handler/{domain}_test.go`
5. Add BDD scenarios in `features/{domain}.feature` and step definitions in `testenv/{domain}_steps.go`
6. Update `README.md` API section

## When Modifying Storage

1. Update `internal/storage/interface.go`
2. Implement in **both** `memory.go` AND `redis.go`
3. Add tests covering new functionality
4. Update seeder if affecting test user creation

## When Changing Constants

1. Update `internal/consts/consts.go`
2. **NEVER** change existing vault UUIDs
3. Update all test files with hardcoded values
4. Update README.md tables

## Testing Checklist

- [ ] All tests pass: `make test`
- [ ] Coverage acceptable: `make coverage` (aim for ≥80%)
- [ ] Docker build succeeds: `docker build -t local-mockgatehub .`
- [ ] Application works with MockGatehub

## Key Files Reference

**Must review before coding**:
1. `internal/consts/consts.go` — all constants
2. `internal/storage/interface.go` — storage contract
3. `internal/models/models.go` — domain models
4. `cmd/mockgatehub/main.go` — routing configuration

**Frequently modified**:
1. `internal/handler/*.go` — API endpoint implementations
2. `internal/storage/memory.go` — in-memory storage logic
3. `internal/webhook/manager.go` — webhook delivery

## Troubleshooting

**Tests failing with Redis**: Ensure Redis is running and DB is clean (`redis-cli -n 1 FLUSHDB`)

**Docker build fails**: Verify `go.mod` and `go.sum` are present, check for syntax errors (`go build ./...`)

**Webhooks not arriving**: Check `WEBHOOK_URL`, verify app backend is running, check Docker logs

**E2E tests fail**: Ensure Docker is running, port 25151 and 26380 are free

### Check a user balance (Interledger App stack)

When MockGatehub runs behind the Interledger App stack:

```bash
# 1) Find Kratos identity ID by email
docker compose -f local/docker-compose.yaml exec -T postgres \
    psql -U postgres -d kratos -c \
    "SELECT i.id FROM identities i \
     JOIN identity_credentials ic ON ic.identity_id = i.id \
     JOIN identity_credential_identifiers ici ON ici.identity_credential_id = ic.id \
     WHERE ici.identifier='716461-sender-p2p@example.com';"

# 2) Find wallet_id using identity ID
docker compose -f local/docker-compose.yaml exec -T postgres \
    psql -U postgres -d backend -c \
    "SELECT wallet_id FROM user_wallets WHERE user_id='IDENTITY_ID';"

# 3) Find GateHub wallet address (provider_id)
docker compose -f local/docker-compose.yaml exec -T postgres \
    psql -U postgres -d backend -c \
    "SELECT id, provider_id, send_currency, receive_currency \
     FROM linked_accounts WHERE wallet_id='WALLET_ID' AND provider='gatehub';"

# 4) Query MockGatehub balances
curl -sk https://mockgatehub.interledger.test/core/v1/wallets/PROVIDER_ID/balances | jq
```

## Critical Notes

1. **ALWAYS run all tests after changes**: `make test`
2. **Maintain API compatibility**: Applications rely on exact GateHub response format
3. **Never modify vault UUIDs or currency codes**: These are immutable
4. **Test both storage backends**: Changes must work with memory AND Redis
5. **Update CLAUDE.md if you discover outdated guidance**: Keep instructions fresh
6. **Use zap structured logging**: Always include context fields for debugging
7. **Webhook queue requires Redis**: Even in in-memory storage mode
