# MockXago — Development Guide

Comprehensive guidance for working on the MockXago project.

## Official API Documentation

**Always consult the official Xago API docs when uncertain about endpoint behavior, request/response formats, or field semantics:**
https://documenter.getpostman.com/view/49463771/2sB3QRo7pf

When implementing or modifying endpoints, check the official docs first to ensure MockXago matches real Xago behavior.

## Project Context

MockXago is a Go mock implementation of the Xago API, designed to support local development and testing of wallet applications that integrate with Xago for multi-currency payments, KYC verification, and balance account management.

### Why MockXago Exists

The Interledger Wallet integrates with Xago for:
- User identity and KYC verification (Persona integration)
- Multi-currency balance accounts (ZAR, USD)
- Bank account beneficiary management
- Fund deposits and transfers
- Transaction history and reporting

MockXago removes the dependency on real Xago credentials and services, enabling:
- **Fully local development** without external API dependencies
- **Automated testing** without API rate limits or sandbox restrictions
- **Predictable behavior** for CI/CD pipelines
- **Rapid iteration** without affecting real Xago sandbox data
- **Test-mode endpoints** for simulating deposits and balance changes

### Critical Constraints

1. **API Compliance**: MockXago must be a drop-in replacement for Xago. Applications expect exact API format compatibility.
2. **Sandbox Parity Only**: Focus on happy paths and sandbox behavior. Production Xago features are out of scope.
3. **Multi-Currency Required**: Support ZAR (South African Rand) and USD (US Dollar) balance accounts.
4. **Test Mode**: Test-only endpoints (under `/v1/test/*`) must be gated by environment variables.
5. **Webhook Delivery**: Must support async webhook delivery with proper HMAC signatures.

## Architecture Overview

### Tech Stack

- **Language**: Go 1.24+
- **HTTP Router**: chi v5 (lightweight, idiomatic)
- **Storage**: Dual backend (in-memory for tests, Redis for runtime/E2E)
- **Job Queue**: Redis-backed job queue with background worker
- **Logging**: uber/zap structured logging
- **Testing**: Godog (BDD), testify (assertions), httptest (HTTP mocking)
- **Containerization**: Docker multi-stage build

### Module Structure

**Important**: MockXago does NOT have its own `go.mod`. It's part of the parent `go/` workspace module at `go/go.mod`. Never create `go.mod` or `go.sum` files in `go/mock/mockxago/`.

### Directory Structure

```
go/mock/mockxago/
├── cmd/mockxago/              # Application entry point
│   └── main.go                # Server setup, routing, startup
├── internal/                  # Private application code
│   ├── auth/                  # Authentication & token validation
│   │   ├── errors.go          # Auth error types
│   │   ├── validator.go       # Token validation logic
│   │   └── validator_test.go
│   ├── handler/               # HTTP request handlers
│   │   ├── handler.go         # Handler struct, dependencies, health check
│   │   ├── handler_test.go    # Handler setup tests
│   │   ├── balance.go         # GET /v1/accounts/{id}/balance
│   │   ├── balance_test.go
│   │   ├── beneficiary.go     # /v1/accounts/{id}/beneficiaries endpoints
│   │   ├── beneficiary_test.go
│   │   ├── currency.go        # GET /v1/currencies (bank deposit details)
│   │   ├── currency_test.go
│   │   ├── deposit.go         # Deposit processing, webhook handling
│   │   ├── deposit_test.go
│   │   ├── kyc.go             # Legacy KYC iframe endpoints
│   │   ├── persona.go         # Persona KYC integration
│   │   ├── persona_test.go
│   │   ├── subaccount.go      # Sub-account CRUD operations
│   │   ├── subaccount_test.go
│   │   ├── test_balance.go    # Test-mode balance manipulation (/v1/test/*)
│   │   ├── test_balance_ext_test.go
│   │   └── transactions.go    # Transaction listing, transfers
│   │   └── transactions_test.go
│   ├── jobs/                  # Background jobs & queue
│   │   ├── job.go             # Job struct definition
│   │   ├── job_test.go
│   │   ├── queue.go           # Redis-backed job queue
│   │   └── queue_test.go
│   │   └── worker.go          # Background job worker
│   ├── logger/                # Logging setup
│   │   └── logger.go          # Zap structured logger initialization
│   ├── models/                # Domain & API models
│   │   ├── models.go          # SubAccount, Beneficiary, Transaction, Deposit, Job
│   │   └── api.go             # Request/response DTOs
│   ├── storage/               # Storage layer
│   │   ├── interface.go       # Storage contract (~40 methods)
│   │   ├── memory.go          # In-memory implementation (sync.RWMutex)
│   │   ├── memory_test.go
│   │   ├── redis.go           # Redis implementation (JSON + atomic ops)
│   │   └── redis_test.go
│   └── utils/                 # Utilities
│       ├── token.go           # Token generation helpers
│       ├── token_test.go
│       └── uuid.go            # UUID generation
├── features/                  # Gherkin BDD feature files (8 files)
│   ├── authentication.feature
│   ├── balance_management.feature
│   ├── beneficiary_management.feature
│   ├── currencies_and_deposits.feature
│   ├── deposits_and_webhooks.feature
│   ├── integration_workflows.feature
│   ├── subaccount_management.feature
│   └── transactions.feature
├── testenv/                   # E2E test environment
│   ├── docker-compose.yml     # Redis (26379) + MockXago (28080)
│   ├── godog_test.go          # BDD test runner (//go:build e2e)
│   ├── *_steps.go             # Step definitions per feature area
│   ├── helpers.go             # Test helpers
│   ├── http_client.go         # Signed HTTP client
│   ├── services.go            # Docker service management
│   ├── test_context.go        # Shared test state
│   ├── types.go               # Test constants
│   └── webhook_server.go      # Mock webhook receiver
├── web/                       # Static web assets
│   └── kyc-iframe.html        # KYC verification iframe
├── Dockerfile                 # Multi-stage Docker build
├── Makefile                   # Build, test, lint targets
├── README.md                  # User-facing documentation
└── CLAUDE.md                  # This file
```

### Design Principles

1. **Dependency Injection**: Handler receives storage and job queue via constructor
2. **Interface-Based Storage**: Enables swapping memory/Redis without code changes
3. **Table-Driven Tests**: Use testify assertions with comprehensive coverage
4. **Idiomatic Go**: Standard project layout, effective Go patterns
5. **Minimal Dependencies**: chi, redis, uuid, testify, zap, godog

## Core Systems

### 1. Storage Layer (`internal/storage/`)

**Interface** (`interface.go`) defines ~40 methods covering:
- **Tokens**: Store/retrieve bearer tokens with TTL
- **Sub-Accounts**: CRUD operations, lookup by wallet ID
- **Beneficiaries**: CRUD, list by account, auto-approval tracking
- **Transactions**: CRUD, list by account with pagination
- **Balances**: Get/update per account per currency (atomic operations)
- **Deposits**: CRUD, lookup by reference
- **Jobs**: Job queue operations (enqueue, dequeue, update status)
- **Personas**: Inquiry tracking for Persona KYC

**Memory** (`memory.go`):
- `sync.RWMutex` for thread safety
- Maps for all entities
- Used for unit tests
- No external dependencies

**Redis** (`redis.go`):
- JSON serialization for complex objects
- Atomic operations for balance updates (`INCRBYFLOAT`)
- Connection pooling (10 connections, 2 min idle)
- Used for integration tests and runtime
- Key patterns:
  ```
  token:{tokenValue}                        → JSON (with TTL)
  token:account:{tokenValue}                → accountId
  subaccount:{accountId}                    → JSON
  subaccount:wallet:{walletId}              → accountId
  beneficiary:{beneficiaryId}               → JSON
  beneficiaries:account:{accountId}         → List of beneficiary IDs
  transaction:{transactionId}               → JSON
  transactions:account:{accountId}          → List of transaction IDs
  balance:{accountId}:{currency}            → Float64 (atomic)
  deposit:{depositId}                       → JSON
  deposit:ref:{reference}                   → depositId
  job:{jobId}                               → JSON
  jobs:ready                                → Sorted set (score = notBefore)
  ```

**Critical Methods**:
```go
// Sub-accounts
CreateSubAccount(account *SubAccount) error
GetSubAccount(accountId string) (*SubAccount, error)
GetSubAccountByWallet(walletId string) (*SubAccount, error)
UpdateSubAccount(account *SubAccount) error

// Beneficiaries
AddBeneficiary(accountId string, ben *Beneficiary) error
GetBeneficiary(benId string) (*Beneficiary, error)
ListBeneficiariesByAccountID(accountId string) ([]*Beneficiary, error)
UpdateBeneficiary(ben *Beneficiary) error

// Balances (atomic)
GetBalance(accountId, currency string) (float64, error)
UpdateBalance(accountId, currency string, delta float64) error

// Transactions
CreateTransaction(tx *Transaction) error
GetTransaction(txId string) (*Transaction, error)
ListTransactionsByAccountID(accountId string, limit, page int) ([]*Transaction, int, error)

// Jobs
EnqueueJob(job *Job) error
DequeueReadyJobs(maxJobs int) ([]*Job, error)
UpdateJobStatus(jobId, status string, failureReason *string) error
```

### 2. Authentication (`internal/auth/`)

**Token Format**: Bearer tokens stored in storage with TTL

**Login Flow**:
1. `POST /v1/login` with `policyId`, `publicKey`, and `secret` fields
2. Validates against `XAGO_API_PUBLIC_KEY` and `XAGO_API_SECRET` environment variables
3. Generates random token (32-character hex string)
4. Stores token → accountId mapping (if accountId provided)
5. Returns `{tokenValue: "..."}`

**Request Headers**:
- `Authorization: Bearer {tokenValue}`

**Token Validation**:
- All protected endpoints require valid bearer token
- Token must exist in storage and not be expired
- Public endpoints: `/health`, `/v1/login`

### 3. Multi-Currency System

**Supported Currencies**:
- **ZAR** (South African Rand) - Ledger ID: `9246927`
- **USD** (US Dollar) - Ledger ID: `9246873`

**Balance Behavior**:
- `GET /v1/accounts/{accountId}/balance` returns array of all currencies
- Format: `[{"currencyCode": "ZAR", "amount": "10000.00"}, {"currencyCode": "USD", "amount": "5000.00"}]`
- Amounts are strings formatted to 2 decimal places
- If balance is 0.00, still include the currency in response

**Currency Details**:
`GET /v1/currencies` returns bank deposit details for each currency:
```json
{
  "data": [
    {
      "currencyCode": "ZAR",
      "accountNumber": "62XXXXXXX",
      "accountName": "Xago (Pty) Ltd",
      "bankName": "First National Bank",
      "branchCode": "250655",
      "swiftCode": "FIRNZAJJ",
      "reference": "{uniqueDepositReference}"
    },
    {
      "currencyCode": "USD",
      "accountNumber": "100XXXXXX",
      "accountName": "Xago (Pty) Ltd",
      "bankName": "Bidvest Bank",
      "swiftCode": "BIDZAJJJ",
      "reference": "{uniqueDepositReference}"
    }
  ]
}
```

### 4. KYC (Know Your Customer) Flow

MockXago supports **Persona KYC integration**:

**Persona Inquiry Flow**:
1. `POST /v1/inquiries` - Creates inquiry, returns iframe URL
   - Input: `{referenceId, nameFirst, nameLast, emailAddress, phoneNumber, addressStreet1, addressCity, addressCountryCode, ...}`
   - Returns: `{id, iframeUrl, status: "pending"}`

2. `GET /v1/inquiries/{inquiryId}` - Get inquiry status
   - Returns: `{id, status, referenceId, ...}`

3. `GET /v1/inquiries/{inquiryId}/iframe` - Serves KYC iframe HTML
   - Serves `web/kyc-iframe.html` with inquiry context

4. `POST /v1/inquiries/{inquiryId}/submit` - Iframe form submission
   - Auto-approves inquiry (status → "completed")
   - Sends webhook to `PERSONA_WEBHOOK_URL` (Persona format)
   - Sends webhook to `WEBHOOK_URL` (wallet format)
   - Iframe posts message to parent: `{type: 'OnboardingCompleted', value: {...}}`

**Auto-Approval Logic**:
- All KYC inquiries are auto-approved on submission
- Final status: `"completed"`
- Webhooks are emitted asynchronously

### 5. Sub-Account Management

**Create Sub-Account**:
- `POST /v1/company/accounts`
- Input: `{firstName, lastName, email, mobileNumber, identityType, idNumber, physicalAddress, thirdPartyVerificationUrl, walletId (optional)}`
- Generates UUID for `accountId` if not provided
- Creates unique deposit references for each currency
- Returns sub-account object with bank deposit details

**Update Sub-Account**:
- `PUT /v1/company/accounts/{accountId}`
- Updates KYC-related fields
- Used for KYC status updates

**Query Sub-Account by Wallet**:
- `GET /v1/company/accounts?walletId={walletId}`
- Returns sub-account associated with wallet
- Used by wallet backend to resolve accountId from walletId

### 6. Beneficiary Management

**Add Beneficiary**:
- `POST /v1/accounts/{accountId}/beneficiaries`
- Input: `{name, accountNumber, currencyCode, branchCode (optional), bankName (optional)}`
- Creates beneficiary with `status: "pending"`
- Enqueues auto-approval job (3-second delay)
- Returns beneficiary object with `id`, `status`, timestamps

**List Beneficiaries**:
- `GET /v1/accounts/{accountId}/beneficiaries`
- Returns array of beneficiaries
- Includes `id`, `name`, `accountNumber`, `currencyCode`, `status`, timestamps

**Auto-Approval**:
- Background job transitions beneficiaries from `"pending"` → `"approved"` after 3 seconds
- Implemented via job queue system

### 7. Transaction & Transfer Handling

**Transaction Types**:
1. **DEPOSIT**: External deposit (e.g., bank transfer)
   - Creates transaction record
   - Updates balance atomically
   - Emits webhook to `WEBHOOK_URL`

2. **TRANSFER**: Transfer to beneficiary (withdrawal)
   - Creates transfer record
   - Deducts balance atomically
   - Returns transfer object

**Transaction Listing**:
- `GET /v1/company/transactions?limit={limit}&page={page}` - List all transactions
- `GET /v1/company/transactions/{id}` - Get single transaction
- Supports pagination (default: limit=10, page=1)
- Returns: `{data: [...], pagination: {limit, page, numberOfPages, total}}`

**Transfer Creation**:
- `POST /v1/transfers`
- Input: `{accountId, beneficiaryId, currencyCode, amount, narration}`
- Validates beneficiary exists and is approved
- Validates sufficient balance
- Deducts balance atomically
- Returns transfer object

### 8. Deposit Processing (`internal/handler/deposit.go`)

**Deposit Workflow**:
1. Client calls `POST /v1/test/balances/deposit` (test mode only)
2. Create deposit record with unique reference
3. Update balance atomically
4. Create transaction record (type: "DEPOSIT", status: "settled")
5. Enqueue webhook job for async delivery

**Webhook Payload** (sent to `WEBHOOK_URL`):
```json
{
  "accountId": "account-uuid",
  "amount": 5000.00,
  "currencyCode": "ZAR",
  "transactionId": "tx-uuid",
  "code": 104,
  "status": "settled",
  "createdAt": "2026-03-06T10:00:00Z",
  "settledAt": "2026-03-06T10:00:00Z"
}
```

**Webhook Signing**:
- `X-Signature` header = hex(HMAC-SHA256(JSON body, webhook secret))
- Secret from `WEBHOOK_SECRET` environment variable

### 9. Job Queue System (`internal/jobs/`)

**Architecture**:
- Redis-backed sorted set for ready jobs (score = `notBefore` timestamp)
- Background worker polls every 5 seconds
- Processes up to 10 jobs per batch

**Job Lifecycle**:
1. **Enqueue**: `EnqueueJob(job)` → status "pending", notBefore timestamp
2. **Dequeue**: Worker fetches ready jobs where `notBefore <= now()`
3. **Process**: Execute job handler (webhook delivery, beneficiary approval)
4. **Complete**: Update status to "completed" or "failed"
5. **Retry**: Max 3 attempts, 30-second backoff between retries

**Job Types**:
- **webhook**: HTTP POST with HMAC signature
- **beneficiary_approval**: Auto-approve pending beneficiary

**Job Model**:
```go
type Job struct {
    ID            string
    Type          string
    Payload       map[string]interface{}
    Status        string  // "pending", "processing", "completed", "failed"
    NotBefore     int64   // Unix timestamp (milliseconds)
    Attempts      int
    MaxAttempts   int
    FailureReason *string
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

### 10. Test-Mode Endpoints (`internal/handler/test_balance.go`)

**CRITICAL**: These endpoints are ONLY available when `XAGO_MOCK_TEST_MODE=true`. They enable E2E tests to manipulate state without real external integrations.

**Endpoints**:
- `POST /v1/test/balances/set` - Set balance directly
  - Input: `{accountId, currencyCode, amount}`
  - Bypasses all business logic

- `POST /v1/test/balances/deposit` - Simulate deposit with webhook
  - Input: `{accountId, currencyCode, amount, reference}`
  - Full deposit flow including webhook delivery

- `POST /v1/test/balances/transfer` - Simulate transfer
  - Input: `{accountId, beneficiaryId, currencyCode, amount}`
  - Bypasses beneficiary approval checks

- `POST /v1/test/balances/reset` - Clear all balances
  - Input: `{accountId}`
  - Sets all currencies to 0.00

- `DELETE /v1/test/deposits` - Clear all deposits
  - Used for test cleanup

## Testing Strategy

### Test Architecture

MockXago has a **three-tier testing strategy**:

1. **Unit Tests** (handler, storage, jobs, auth, utils)
2. **Integration Tests** (not yet implemented - future work)
3. **E2E BDD Tests** (Godog/Cucumber in `testenv/`)

### Unit Tests

**Location**: `internal/*/`
**Pattern**: Table-driven tests with testify assertions
**Coverage Target**: 65%+ for `internal/` packages

**Running Unit Tests**:
```bash
# All unit tests
cd go/mock/mockxago
go test ./internal/... -v

# Specific package
go test ./internal/handler -v

# With coverage
go test ./internal/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total
```

**Current Coverage**:
- **Overall**: ~44%
- **auth**: 95.0% (excellent)
- **handler**: 33.8% (target 65%+)
- **jobs**: 41.4% (needs expansion)
- **storage**: 91.0% (memory only)
- **utils**: 66.7% (acceptable)

### E2E BDD Tests (Godog)

**Location**: `testenv/`
**Framework**: Godog (Go implementation of Cucumber)
**Test Definition**: Gherkin `.feature` files

**Feature Files** (8 total):
1. `authentication.feature` - Login, token validation
2. `balance_management.feature` - Balance queries, updates
3. `beneficiary_management.feature` - Add, list, approval flow
4. `currencies_and_deposits.feature` - Currency details, deposit references
5. `deposits_and_webhooks.feature` - Deposit processing, webhook delivery
6. `integration_workflows.feature` - End-to-end user journeys
7. `subaccount_management.feature` - Create, update, query sub-accounts
8. `transactions.feature` - Transaction listing, transfers

**Test Environment**:
- **Docker Compose**: Isolated Redis + MockXago containers
- **Ports**:
  - MockXago: `28080`
  - Redis: `26379`
  - Webhook receiver: `28081`
- **State Management**: Clean Redis before each scenario

**Running E2E Tests**:
```bash
cd go/mock/mockxago

# All E2E tests (builds Docker, runs scenarios)
go test ./testenv/... -v

# Specific tag
go test ./testenv/... -v -godog.tags="@authentication"

# Multiple tags (AND)
go test ./testenv/... -v -godog.tags="@deposits&&@webhooks"

# Stop on first failure
go test ./testenv/... -v -godog.stop-on-failure
```

**Redis Storage Testing**:
```go
func TestRedisStorage(t *testing.T) {
    redisURL := os.Getenv("MOCKXAGO_REDIS_URL")
    if redisURL == "" {
        t.Skip("Skipping Redis tests: MOCKXAGO_REDIS_URL not set")
    }

    store, err := NewRedisStorage(redisURL, 1) // Use DB 1 for tests
    require.NoError(t, err)
    defer store.FlushDB() // Clean up
}
```

## Development Workflow

### CRITICAL: Test-Driven Development (TDD)

**All changes to MockXago MUST follow TDD**:

1. **Write/Update BDD Tests First** (`features/*.feature`)
2. **Write/Update Unit Tests** (`internal/*/test.go`) - aim for 65%+ coverage
3. **Implement Changes** - minimal code to make tests pass
4. **Iterate Until Unit Tests Pass**: `go test ./internal/... -v -run TestYourNewFeature`
5. **Iterate Until BDD Tests Pass**: `go test ./testenv/... -v -godog.tags="@your-feature"`
6. **Verify Full Test Suite**: `go test ./... -v`

### Local Development

```bash
# Start Redis (required for webhooks and job queue)
docker run -d --name mockxago-redis -p 6379:6379 redis:7-alpine

# Set environment variables
export XAGO_API_PUBLIC_KEY=test-public-key
export XAGO_API_SECRET=test-secret
export XAGO_MOCK_PORT=8080
export XAGO_MOCK_TEST_MODE=true
export MOCKXAGO_REDIS_URL=redis://localhost:6379
export WEBHOOK_URL=http://localhost:3000/xago/webhooks/event
export WEBHOOK_SECRET=test-webhook-secret
export LOG_LEVEL=debug

# Run MockXago
cd go/mock/mockxago
go run ./cmd/mockxago
```

### Making Changes

```bash
# Always run from go/mock/mockxago directory
cd go/mock/mockxago

# Verify code compiles
go build ./cmd/mockxago

# Run unit tests
go test ./internal/... -v

# Check coverage
go test ./internal/... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total

# Run E2E tests (requires Docker)
go test ./testenv/... -v

# Format code
go fmt ./...
```

### Adding New Endpoints (TDD Workflow)

1. Add BDD scenario in `features/foo.feature`
2. Add step definitions in `testenv/foo_steps.go` (if needed)
3. Write unit tests in `internal/handler/foo_test.go`
4. Define models in `internal/models/api.go`
5. Add storage methods to `internal/storage/interface.go`
6. Implement in **both** `memory.go` AND `redis.go`
7. Implement handler in `internal/handler/foo.go`
8. Register route in `cmd/mockxago/main.go`
9. Run unit tests, then BDD tests until passing
10. Update `README.md` and `CLAUDE.md` if patterns changed

## Configuration Reference

### Environment Variables

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `XAGO_MOCK_PORT` | `8080` | No | HTTP server port |
| `XAGO_API_PUBLIC_KEY` | `test-public-key` | Yes | Expected login public key |
| `XAGO_API_SECRET` | `test-secret` | Yes | Expected login secret |
| `XAGO_MOCK_TEST_MODE` | `false` | No | Enable `/v1/test/*` endpoints |
| `MOCKXAGO_REDIS_URL` | `""` | Recommended | Redis connection (enables Redis storage) |
| `MOCKXAGO_REDIS_DB` | `0` | No | Redis database number |
| `WEBHOOK_URL` | `""` | No | Wallet webhook delivery URL |
| `WEBHOOK_SECRET` | `""` | No | HMAC secret for webhook signatures |
| `PERSONA_WEBHOOK_URL` | `""` | No | Persona webhook delivery URL |
| `PERSONA_WEBHOOK_TOKEN` | `""` | No | HMAC secret for Persona webhooks |
| `LOG_LEVEL` | `info` | No | Logging level (debug, info, warn, error) |

### Xago Constants

**Policy ID**: `5e2585a474b0e90012ce8ff1` (hardcoded, matches real Xago sandbox)

**Currency Ledger IDs**:
- ZAR: `9246927`
- USD: `9246873`

**Operations Accounts** (for E2E tests):
- USD: `868196c3-f6b4-4920-bbfb-d1c7f6a98183`
- ZAR: `b0944908-16e6-4ef4-8677-192165e33c59`

**Transaction Codes**:
- `104`: Deposit completed

**Beneficiary Statuses**:
- `pending`: Newly created, not yet approved
- `approved`: Approved and ready for transfers

## Logging Guidelines

It's safe to log sensitive values in MockXago — this is a development/testing mock, not production.

```go
import (
    "go.uber.org/zap"
    "github.com/interledger/interledger-app/go/mock/mockxago/internal/logger"
)

logger.Log.Info("sub-account created",
    zap.String("account_id", accountID),
    zap.String("wallet_id", walletID),
)
```

**Log Levels**:
- `Debug`: Token validation, balance calculations, job processing details
- `Info`: Account created, deposit processed, webhook sent
- `Warn`: Invalid input, missing optional fields, retry attempts
- `Error`: Storage failures, webhook delivery failed, unexpected errors

## Troubleshooting

**Unit tests fail with "sub-account not found"**: Ensure `setupTestHandler(t)` is called for each test

**E2E tests fail with "connection refused"**: Check `docker ps`, verify health checks pass, wait for services

**Webhooks not arriving**:
- Verify: `curl -X POST $WEBHOOK_URL -d '{"test": true}'`
- Check `WEBHOOK_SECRET` matches backend configuration
- Verify job enqueued: `redis-cli ZRANGE jobs:ready 0 -1`

**Redis "connection refused"**: `docker run -d -p 6379:6379 redis:7-alpine`

**Build fails with "package not found"**: `cd go && go mod tidy`

**Import path errors**: All imports must use `github.com/interledger/interledger-app/go/mock/mockxago/internal/...`

### Debugging

```bash
# Enable debug logging
export LOG_LEVEL=debug
go run ./cmd/mockxago

# Check Redis state
redis-cli -p 6379
KEYS *
ZRANGE jobs:ready 0 -1 WITHSCORES

# Run single scenario by line number
go test ./testenv/... -v -godog.paths="features/authentication.feature:12"
```

## API Compliance Notes

### Known Divergences from Official Xago API

1. **Beneficiary Endpoint Paths**:
   - **Official**: `POST /v1/beneficiaries`, `GET /v1/beneficiaries`
   - **MockXago**: `POST /v1/accounts/{accountId}/beneficiaries`, `GET /v1/accounts/{accountId}/beneficiaries`
   - **Status**: Accepted divergence for clarity

2. **Test-Mode Endpoints**: `/v1/test/*` endpoints are MockXago-only, essential for E2E automation

3. **Auto-Approval Timing**: MockXago approves in 3 seconds vs. hours/days in production

## Key Files Reference

**Must Review Before Coding**:
1. `internal/models/models.go` - Domain models
2. `internal/models/api.go` - Request/response DTOs
3. `internal/storage/interface.go` - Storage contract
4. `cmd/mockxago/main.go` - Routing configuration

**Frequently Modified**:
1. `internal/handler/*.go` - API implementations
2. `internal/storage/memory.go` - In-memory storage
3. `internal/storage/redis.go` - Redis storage
4. `features/*.feature` - BDD scenarios
5. `testenv/*_steps.go` - BDD step definitions

**Rarely Touch**:
1. `internal/logger/logger.go` - Logging setup
2. `internal/jobs/worker.go` - Job queue worker
3. `Dockerfile` - Container build
4. `web/kyc-iframe.html` - KYC iframe

## Success Metrics

- **Test Coverage**: ≥65% for `internal/` packages
- **E2E Pass Rate**: 100% (all scenarios passing)
- **API Compliance**: Wallet backend runs without modification
- **Build Time**: < 2 minutes for Docker build
- **Response Time**: All endpoints < 100ms (local)
