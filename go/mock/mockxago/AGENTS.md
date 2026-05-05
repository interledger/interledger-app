# MockXago — AI Agent Development Guide

Comprehensive guidance for AI coding agents working on the MockXago project.

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

**Important**: MockXago does NOT have its own `go.mod`. It's part of the parent `go/` workspace module at `/home/stephan/interledger/interledger-app/go/go.mod`. Never create `go.mod` or `go.sum` files in `go/mock/mockxago/`.

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
└── AGENTS.md                  # This file
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

**CRITICAL**: These endpoints are ONLY available when specific environment variables are set. They enable E2E tests to manipulate state without real external integrations.

**Endpoints** (require `XAGO_MOCK_TEST_MODE=true` or equivalent):
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

MockXago has a **comprehensive three-tier testing strategy**:

1. **Unit Tests** (handler, storage, jobs, auth, utils)
2. **Integration Tests** (not yet implemented - future work)
3. **E2E BDD Tests** (Godog/Cucumber in `testenv/`)

### Unit Tests

**Location**: `internal/*/`  
**Pattern**: Table-driven tests with testify assertions  
**Coverage Target**: 65%+ for `internal/` packages

**Example Test Structure**:
```go
func TestAddBeneficiary(t *testing.T) {
    tests := []struct {
        name       string
        accountID  string
        body       string
        wantStatus int
        wantErr    bool
    }{
        {
            name:       "valid beneficiary",
            accountID:  "valid-account-id",
            body:       `{"name":"Test Bank","accountNumber":"123456"}`,
            wantStatus: 201,
            wantErr:    false,
        },
        {
            name:       "missing name",
            accountID:  "valid-account-id",
            body:       `{"accountNumber":"123456"}`,
            wantStatus: 400,
            wantErr:    true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

**Test Helpers**:
```go
// Setup test handler with in-memory storage
func setupTestHandler(t *testing.T) *Handler {
    store := storage.NewMemoryStorage()
    queue := jobs.NewMemoryQueue()
    return NewHandler(store, queue)
}

// Setup handler with authentication
func setupAuthHandler(t *testing.T) *Handler {
    h := setupTestHandler(t)
    // Create test user and issue token
    return h
}

// Issue bearer token for tests
func issueToken(t *testing.T, h *Handler) string {
    // Login and return tokenValue
}

// Create authorized request
func authorizedRequest(method, url, token string, body io.Reader) *http.Request {
    req := httptest.NewRequest(method, url, body)
    req.Header.Set("Authorization", "Bearer "+token)
    return req
}
```

**Running Unit Tests**:
```bash
# All unit tests
cd go/mock/mockxago
go test ./internal/... -v

# Specific package
go test ./internal/handler -v

# With coverage
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Coverage summary
go tool cover -func=coverage.out | grep total
```

**Current Coverage** (as of Phase 5 completion):
- **Overall**: ~44%
- **auth**: 95.0% ✓ (excellent)
- **handler**: 33.8% (expanded from 22.4%, target 65%+)
- **jobs**: 41.4% (needs expansion)
- **storage**: 91.0% ✓ (memory only)
- **utils**: 66.7% ✓ (acceptable)
- **logger**: 0.0% (not critical)
- **models**: 0.0% (not critical)

### E2E BDD Tests (Godog)

**Location**: `testenv/`  
**Framework**: Godog (Go implementation of Cucumber)  
**Test Definition**: Gherkin `.feature` files  
**Step Implementations**: `*_steps.go` files

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
- **Network**: All services on shared Docker network

**Running E2E Tests**:
```bash
cd go/mock/mockxago

# All E2E tests (builds Docker, runs scenarios)
go test ./testenv/... -v

# Specific tag
go test ./testenv/... -v -godog.tags="@authentication"

# Multiple tags (AND)
go test ./testenv/... -v -godog.tags="@deposits&&@webhooks"

# Pretty output
go test ./testenv/... -v -godog.format=pretty

# Stop on first failure
go test ./testenv/... -v -godog.stop-on-failure
```

**BDD Test Structure**:

**Feature File Example** (`features/authentication.feature`):
```gherkin
Feature: Authentication
  As a wallet application
  I want to authenticate with MockXago
  So that I can access protected endpoints

  @authentication
  Scenario: Successful login with valid credentials
    When I POST to "/v1/login" with body:
      """
      {
        "policyId": "5e2585a474b0e90012ce8ff1",
        "fields": [
          {"fieldName": "publicKey", "fieldValue": "test-public-key"},
          {"fieldName": "secret", "fieldValue": "test-secret"}
        ]
      }
      """
    Then the response status should be 200
    And the response should have field "tokenValue"
    And I save the field "tokenValue" as "authToken"

  @authentication
  Scenario: Access protected endpoint with valid token
    Given I am authenticated with valid credentials
    When I GET "/v1/company/accounts?walletId=test-wallet-123" with auth
    Then the response status should be 200
```

**Step Definition Example** (`testenv/auth_steps.go`):
```go
func (tc *TestContext) iAmAuthenticatedWithValidCredentials() error {
    // Login and store token
    resp, err := tc.httpClient.Login(
        "test-public-key",
        "test-secret",
    )
    if err != nil {
        return err
    }
    tc.authToken = resp.TokenValue
    return nil
}

func (tc *TestContext) iGETWithAuth(path string) error {
    req, _ := http.NewRequest("GET", tc.baseURL+path, nil)
    req.Header.Set("Authorization", "Bearer "+tc.authToken)
    
    resp, err := tc.httpClient.Do(req)
    if err != nil {
        return err
    }
    tc.lastResponse = resp
    return nil
}
```

**Test Context** (`testenv/test_context.go`):
```go
type TestContext struct {
    baseURL       string
    httpClient    *http.Client
    authToken     string
    lastResponse  *http.Response
    lastBody      []byte
    savedValues   map[string]interface{}
}
```

**Docker Compose** (`testenv/docker-compose.yml`):
```yaml
services:
  redis:
    image: valkey/valkey:8-alpine
    ports:
      - "26379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 2s
      timeout: 1s
      retries: 5

  mockxago:
    build:
      context: ../
      dockerfile: Dockerfile
    ports:
      - "28080:8080"
    environment:
      - XAGO_MOCK_PORT=8080
      - XAGO_API_PUBLIC_KEY=test-public-key
      - XAGO_API_SECRET=test-secret
      - XAGO_MOCK_TEST_MODE=true
      - MOCKXAGO_REDIS_URL=redis://redis:6379
      - WEBHOOK_URL=http://webhook-server:8081/webhook
      - WEBHOOK_SECRET=test-webhook-secret
    depends_on:
      redis:
        condition: service_healthy
```

**Test Workflow**:
1. **Setup**: Start Docker Compose services
2. **Before Scenario**: Clean Redis, reset test context
3. **Execute Steps**: Run Gherkin steps sequentially
4. **Assertions**: Validate responses, state, webhooks
5. **After Scenario**: Record results
6. **Teardown**: Stop containers after all scenarios

### Testing Best Practices

**Unit Testing**:
- ✅ Test both success and error paths
- ✅ Use table-driven tests for multiple scenarios
- ✅ Mock external dependencies (storage, job queue)
- ✅ Test concurrent access where applicable
- ✅ Validate JSON marshaling/unmarshaling
- ✅ Test edge cases (empty strings, nil values, boundary conditions)

**BDD Testing**:
- ✅ Write scenarios from user perspective
- ✅ Use `@tags` for organizing related scenarios
- ✅ Keep scenarios focused (test one behavior)
- ✅ Use Background for common setup steps
- ✅ Validate webhooks are delivered correctly
- ✅ Test full end-to-end workflows

**Redis Storage Testing**:
- ✅ Skip Redis tests if Redis not available (use `t.Skip()`)
- ✅ Clean database before each test (`FLUSHDB`)
- ✅ Test atomic operations (concurrent balance updates)
- ✅ Validate TTL for tokens
- ✅ Test job queue ordering (sorted set by timestamp)

**Example Redis Test Guard**:
```go
func TestRedisStorage(t *testing.T) {
    redisURL := os.Getenv("MOCKXAGO_REDIS_URL")
    if redisURL == "" {
        t.Skip("Skipping Redis tests: MOCKXAGO_REDIS_URL not set")
    }
    
    store, err := NewRedisStorage(redisURL, 1) // Use DB 1 for tests
    require.NoError(t, err)
    defer store.FlushDB() // Clean up
    
    // Test implementation
}
```

## Development Workflow

### CRITICAL: Test-Driven Development (TDD) Approach

**All changes to MockXago MUST follow a strict TDD workflow**:

1. **Write/Update BDD Tests First** (`features/*.feature`)
   - Define expected behavior in Gherkin scenarios
   - Add new scenarios for new features
   - Update existing scenarios if changing behavior
   - Ensures behavior is well-defined before implementation

2. **Write/Update Unit Tests** (`internal/*/test.go`)
   - Write failing tests that define the specification
   - Cover success paths, error paths, edge cases
   - Aim for 65%+ coverage on new code
   - Use table-driven test patterns

3. **Implement Changes**
   - Write minimal code to make tests pass
   - Refactor for clarity while keeping tests green
   - Add logging, documentation, error handling

4. **Iterate Until Unit Tests Pass**
   ```bash
   go test ./internal/... -v -run TestYourNewFeature
   ```
   - Fix failures one by one
   - Don't move on until all unit tests pass

5. **Iterate Until BDD Tests Pass**
   ```bash
   go test ./testenv/... -v -godog.tags="@your-feature"
   ```
   - Implement missing step definitions
   - Fix integration issues
   - Verify end-to-end behavior

6. **Verify Full Test Suite**
   ```bash
   go test ./... -v
   ```
   - All tests must pass before committing
   - Check coverage: `go test ./internal/... -coverprofile=coverage.out`

**Why TDD?**
- ✅ Prevents regressions
- ✅ Documents expected behavior
- ✅ Ensures testability from the start
- ✅ Faster debugging (failing tests show exactly what broke)
- ✅ Confidence in refactoring

### CRITICAL: Documentation Maintenance

**EVERY time you make changes to MockXago, you MUST**:

1. **Review and Update** `/home/stephan/interledger/interledger-app/documentation/docs/`
   - Check if any Markdown files document the changed feature
   - Update API examples if endpoint behavior changed
   - Add new documentation for new features
   - Files to check: `mockxago-*.md`, integration guides, architecture docs

2. **Review and Update** `go/mock/mockxago/README.md`
   - Update API reference section if endpoints changed
   - Update configuration section if env vars added
   - Update examples if request/response formats changed
   - Update feature status section

3. **Review and Update** `go/mock/mockxago/AGENTS.md` (this file)
   - Update "Core Systems" if architecture changed
   - Update "Adding New Endpoints" workflow if patterns changed
   - Update "Testing Strategy" if test approaches changed
   - Update "Configuration Reference" if env vars added
   - Update "Key Files Reference" if new critical files added

**Documentation Review Checklist**:
```bash
# After making changes, run these checks:

# 1. Check if your changes affect documentation
grep -r "YourChangedFeature" documentation/docs/
grep "YourChangedFeature" go/mock/mockxago/README.md

# 2. Update documentation files
vim documentation/docs/mockxago-*.md  # If feature documented
vim go/mock/mockxago/README.md        # API reference, config
vim go/mock/mockxago/AGENTS.md        # Agent instructions

# 3. Verify documentation builds (if using mkdocs)
cd documentation && mkdocs build
```

**Why Documentation Maintenance Matters**:
- ✅ Prevents documentation drift from implementation
- ✅ Helps future agents understand the system
- ✅ Assists developers integrating with MockXago
- ✅ Maintains consistency across the codebase

### 1. Local Development

```bash
# Start Redis (required for webhooks and job queue)
docker run -d --name mockxago-redis -p 6379:6379 valkey/valkey:8-alpine

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

### 2. Making Changes

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

# Run linter (if configured)
golangci-lint run ./...
```

### 3. Adding New Endpoints (TDD Workflow)

**CRITICAL**: Follow TDD approach - write tests BEFORE implementation!

**Step-by-Step Checklist**:

**PHASE 1: Define Behavior (Tests First)**

1. **Add BDD Scenario** (`features/foo.feature`)
   ```gherkin
   @foo
   Scenario: Create foo successfully
     Given I am authenticated with valid credentials
     When I POST to "/v1/foos" with body:
       """
       {"name": "Test Foo", "amount": 100.00}
       """
     Then the response status should be 201
     And the response should have field "id"
     And the response field "name" should equal "Test Foo"
   ```

2. **Add Step Definitions** (`testenv/foo_steps.go`) - if needed
   ```go
   func (tc *TestContext) iPOSTToWithBody(path, body string) error {
       // Implementation (likely already exists)
   }
   ```

3. **Write Unit Tests** (`internal/handler/foo_test.go`)
   ```go
   func TestCreateFoo(t *testing.T) {
       tests := []struct {
           name       string
           body       string
           wantStatus int
           wantErr    bool
       }{
           {"valid", `{"name":"Test","amount":100.00}`, 201, false},
           {"missing name", `{"amount":100.00}`, 400, true},
           {"negative amount", `{"name":"Test","amount":-50}`, 400, true},
       }
       
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               h := setupTestHandler(t)
               // Test implementation
           })
       }
   }
   ```

**PHASE 2: Implementation**

4. **Define Models** (`internal/models/api.go` or `models.go`)
   ```go
   type CreateFooRequest struct {
       Name   string `json:"name"`
       Amount float64 `json:"amount"`
   }
   
   type FooResponse struct {
       ID        string    `json:"id"`
       Name      string    `json:"name"`
       Amount    float64   `json:"amount"`
       CreatedAt time.Time `json:"createdAt"`
   }
   ```

2. **Add Storage Methods** (`internal/storage/interface.go`)
   ```go
   CreateFoo(foo *Foo) error
   GetFoo(id string) (*Foo, error)
   ```

3. **Implement Storage** (both `memory.go` AND `redis.go`)
   ```go
   // memory.go
   func (s *MemoryStorage) CreateFoo(foo *Foo) error {
       s.mu.Lock()
       defer s.mu.Unlock()
       s.foos[foo.ID] = foo
       return nil
   }
   
   // redis.go
   func (s *RedisStorage) CreateFoo(foo *Foo) error {
       data, _ := json.Marshal(foo)
       return s.client.Set(ctx, "foo:"+foo.ID, data, 0).Err()
   }
   ```

7. **Implement Handler** (`internal/handler/foo.go`)
   ```go
   func (h *Handler) CreateFoo(w http.ResponseWriter, r *http.Request) {
       var req CreateFooRequest
       if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
           respondError(w, http.StatusBadRequest, "Invalid request")
           return
       }
       
       // Validate (makes tests pass)
       if req.Name == "" {
           respondError(w, http.StatusBadRequest, "name is required")
           return
       }
       if req.Amount < 0 {
           respondError(w, http.StatusBadRequest, "amount must be non-negative")
           return
       }
       
       foo := &Foo{
           ID:        uuid.New().String(),
           Name:      req.Name,
           Amount:    req.Amount,
           CreatedAt: time.Now(),
       }
       
       if err := h.storage.CreateFoo(foo); err != nil {
           logger.Log.Error("Failed to create foo", zap.Error(err))
           respondError(w, http.StatusInternalServerError, "Internal error")
           return
       }
       
       respondJSON(w, http.StatusCreated, FooResponse{...})
   }
   ```

8. **Register Route** (`cmd/mockxago/main.go`)
   ```go
   r.Post("/v1/foos", handler.CreateFoo)
   r.Get("/v1/foos/{id}", handler.GetFoo)
   ```

**PHASE 3: Verification & Documentation**

9. **Run Unit Tests** (iterate until passing)
   ```bash
   go test ./internal/handler -v -run TestCreateFoo
   go test ./internal/storage -v -run TestFoo
   ```

10. **Run BDD Tests** (iterate until passing)
    ```bash
    go test ./testenv/... -v -godog.tags="@foo"
    ```

11. **Verify Full Coverage**
    ```bash
    go test ./internal/... -coverprofile=coverage.out
    go tool cover -html=coverage.out
    # Ensure new code has 65%+ coverage
    ```

12. **Update Documentation**
    - `README.md`: Add API endpoint documentation
    - `AGENTS.md`: Update if patterns changed
    - `documentation/docs/`: Add user-facing documentation

13. **Final Validation**
    ```bash
    # All tests must pass
    go test ./... -v
    
    # Build must succeed
    go build ./cmd/mockxago
    
    # Docker build must succeed
    docker build -t mockxago:test .
    ```

**TDD Mantra**: 🔴 Red (write failing test) → 🟢 Green (make it pass) → 🔵 Refactor (improve code)

### 4. Debugging

**Enable Debug Logging**:
```bash
export LOG_LEVEL=debug
go run ./cmd/mockxago
```

**Check Redis State**:
```bash
# Connect to Redis
redis-cli -p 6379

# List all keys
KEYS *

# Get specific value
GET token:eyJhbGci...

# Check job queue
ZRANGE jobs:ready 0 -1 WITHSCORES

# Flush test database
SELECT 1
FLUSHDB
```

**Webhook Debugging**:
```bash
# Start simple webhook receiver
python3 -m http.server 8081

# Or use netcat
nc -l 8081

# Check MockXago logs for webhook delivery
docker logs mockxago-test -f | grep webhook
```

**Test Specific Scenario**:
```bash
# Run single scenario by line number
go test ./testenv/... -v -godog.paths="features/authentication.feature:12"

# Run scenarios matching name
go test ./testenv/... -v -godog.filter="Successful login"
```

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

**Important**: It's safe to log sensitive values in MockXago because:
- This is a development/testing mock, not production
- Applications using it are in local test environments
- Verbose logging helps with integration debugging
- No real credentials or production data flows through this service

**Logging Standards**:
```go
import (
    "go.uber.org/zap"
    "github.com/interledger/interledger-app/go/mock/mockxago/internal/logger"
)

// Info - Normal operations
logger.Log.Info("sub-account created",
    zap.String("account_id", accountID),
    zap.String("wallet_id", walletID),
)

// Warn - Non-fatal issues
logger.Log.Warn("beneficiary approval delayed",
    zap.String("beneficiary_id", benID),
    zap.Int("attempts", attempts),
)

// Error - Operations failed
logger.Log.Error("failed to update balance",
    zap.Error(err),
    zap.String("account_id", accountID),
    zap.String("currency", currency),
    zap.Float64("delta", delta),
)

// Debug - Detailed diagnostics
logger.Log.Debug("token validated",
    zap.String("token", token[:12]+"..."),
    zap.String("account_id", accountID),
)
```

**Log Levels**:
- `Debug`: Token validation, balance calculations, job processing details
- `Info`: Account created, deposit processed, webhook sent
- `Warn`: Invalid input, missing optional fields, retry attempts
- `Error`: Storage failures, webhook delivery failed, unexpected errors

## Troubleshooting

### Tests Failing

**Unit tests fail with "sub-account not found"**:
- **Cause**: Storage state not properly initialized
- **Solution**: Ensure `setupTestHandler(t)` is called for each test

**E2E tests fail with "connection refused"**:
- **Cause**: Docker services not started or not ready
- **Solution**: Check `docker ps`, verify health checks pass, wait for services

**Coverage below target**:
- **Cause**: Need more test cases
- **Solution**: Run `go tool cover -html=coverage.out`, identify untested paths, add tests

### Webhooks Not Arriving

**Webhook URL not reachable**:
- **Cause**: `WEBHOOK_URL` incorrect or backend not running
- **Solution**: 
  - Verify: `curl -X POST $WEBHOOK_URL -d '{"test": true}'`
  - Check backend logs
  - Verify Docker network connectivity

**Webhook signature validation fails**:
- **Cause**: Secret mismatch between MockXago and backend
- **Solution**: Ensure `WEBHOOK_SECRET` matches backend configuration

**Webhooks not triggering**:
- **Cause**: Job queue not processing
- **Solution**:
  - Check Redis connection: `redis-cli -p 6379 ping`
  - Verify job enqueued: `redis-cli ZRANGE jobs:ready 0 -1`
  - Check worker logs for errors

### Redis Issues

**"connection refused" errors**:
- **Cause**: Redis not running or wrong port
- **Solution**: `docker run -d -p 6379:6379 valkey/valkey:8-alpine`

**Balance not updating**:
- **Cause**: Using in-memory storage instead of Redis
- **Solution**: Set `MOCKXAGO_REDIS_URL=redis://localhost:6379`

**Job queue not processing**:
- **Cause**: Worker not started or Redis connection lost
- **Solution**: Check logs, verify Redis health, restart MockXago

### Build/Runtime Issues

**Build fails with "package not found"**:
- **Cause**: Dependencies not downloaded
- **Solution**: `cd /home/stephan/interledger/interledger-app/go && go mod tidy`

**Docker build fails**:
- **Cause**: Incorrect paths or missing files
- **Solution**: Verify Dockerfile paths, ensure all files committed

**Import path errors**:
- **Cause**: Wrong module path
- **Solution**: All imports must use `github.com/interledger/interledger-app/go/mock/mockxago/internal/...`

## API Compliance Notes

### Known Divergences from Official Xago API

1. **Beneficiary Endpoint Paths**:
   - **Official**: `POST /v1/beneficiaries`, `GET /v1/beneficiaries`
   - **MockXago**: `POST /v1/accounts/{accountId}/beneficiaries`, `GET /v1/accounts/{accountId}/beneficiaries`
   - **Impact**: Medium - beneficiaries scoped to accountId in URL
   - **Status**: Accepted divergence for clarity

2. **Test-Mode Endpoints**:
   - **MockXago Only**: `/v1/test/*` endpoints
   - **Purpose**: Essential for E2E automation
   - **Status**: Intentional addition, not in official API

3. **Auto-Approval Timing**:
   - **Official**: Variable approval time (hours to days)
   - **MockXago**: 3 seconds (configurable via job queue)
   - **Status**: Intentional for fast testing

### Future Compliance Work

**Phase 12: API Compliance Improvements** (Optional):
- Add `/v1/beneficiaries` endpoints as aliases
- Make approval timing configurable
- Add rate limiting simulation
- Add error simulation endpoints

## AI Agent Best Practices

### Critical: Always Follow TDD

**NO EXCEPTIONS**: Every code change follows the TDD workflow:
1. Write/update tests FIRST (BDD + unit)
2. Run tests (they should FAIL)
3. Write minimal code to pass tests
4. Refactor while keeping tests green
5. Update documentation

**Never**:
- ❌ Write implementation before tests
- ❌ Skip BDD tests because "unit tests are enough"
- ❌ Commit code with failing tests
- ❌ Skip documentation updates

### Critical: Maintain testenv/

**The `testenv/` directory is NOT optional**. When making changes:

✅ **DO**:
- Add BDD scenarios for new endpoints
- Update step definitions when API responses change
- Ensure backward compatibility
- Test webhook delivery in E2E scenarios
- Update assertions to match new behavior
- Document breaking changes clearly

❌ **DON'T**:
- Skip E2E tests ("unit tests are enough")
- Break existing scenarios without updating them
- Change response formats without updating all tests
- Remove testenv infrastructure

### Documentation Update Checklist

After EVERY change, check these files:

```bash
# 1. Project documentation
ls documentation/docs/mockxago-*  # Update if feature documented

# 2. User-facing documentation
vim go/mock/mockxago/README.md     # API reference, examples

# 3. Agent instructions
vim go/mock/mockxago/AGENTS.md     # This file - update patterns/workflow

# 4. Verify consistency
grep -r "YourFeatureName" documentation/docs/ go/mock/mockxago/
```

**Ask yourself**:
- Did I change API endpoint paths or parameters? → Update README.md
- Did I add environment variables? → Update README.md + AGENTS.md
- Did I change test patterns or workflow? → Update AGENTS.md
- Did I add new features? → Update documentation/docs/

### When Stuck

1. **Check existing implementation** in similar endpoints
2. **Review feature files** for expected behavior
3. **Check README.md** for API documentation
4. **Run tests** to see what's failing
5. **Check logs** with `LOG_LEVEL=debug`
6. **Inspect Redis** with `redis-cli` to see actual state
7. **Test against real wallet backend** if available
8. **Review this AGENTS.md** for patterns and workflows

### Common Patterns to Follow

**Handler Error Handling**:
```go
// Decode request
var req CreateFooRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    respondError(w, http.StatusBadRequest, "Invalid request body")
    return
}

// Validate input
if req.Name == "" {
    respondError(w, http.StatusBadRequest, "name is required")
    return
}

// Storage operation
foo, err := h.storage.CreateFoo(&models.Foo{...})
if err != nil {
    logger.Log.Error("Failed to create foo", zap.Error(err))
    respondError(w, http.StatusInternalServerError, "Internal server error")
    return
}

// Success response
respondJSON(w, http.StatusCreated, foo)
```

**Async Webhook with Job Queue**:
```go
// Create webhook job
job := &models.Job{
    ID:          uuid.New().String(),
    Type:        "webhook",
    Payload: map[string]interface{}{
        "url":    webhookURL,
        "data":   webhookData,
        "secret": webhookSecret,
    },
    Status:      "pending",
    NotBefore:   time.Now().Add(2 * time.Second).UnixMilli(),
    Attempts:    0,
    MaxAttempts: 3,
    CreatedAt:   time.Now(),
}

// Enqueue
if err := h.jobQueue.EnqueueJob(job); err != nil {
    logger.Log.Error("Failed to enqueue webhook job", zap.Error(err))
}
```

**Atomic Balance Update**:
```go
// Get current balance
currentBalance, err := h.storage.GetBalance(accountID, currency)
if err != nil {
    return err
}

// Validate sufficient funds (for withdrawals)
if currentBalance + delta < 0 {
    return errors.New("insufficient balance")
}

// Update atomically
if err := h.storage.UpdateBalance(accountID, currency, delta); err != nil {
    return err
}
```

## Success Metrics

Your changes should maintain or improve:
- ✅ **Test Coverage**: ≥65% for `internal/` packages
- ✅ **E2E Pass Rate**: 100% (all scenarios passing)
- ✅ **API Compliance**: Wallet backend runs without modification
- ✅ **Build Time**: < 2 minutes for Docker build
- ✅ **Response Time**: All endpoints < 100ms (local)
- ✅ **Memory Usage**: < 100MB (in-memory), < 150MB (Redis)

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

---

**Last Updated**: March 6, 2026  
**Phase**: 5 (Foundation - Complete)  
**Maintainers**: Interledger Foundation  
**Repository**: https://github.com/interledger/interledger-app
