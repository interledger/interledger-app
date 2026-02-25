# MockXago - AI Agent Development Guide

This document provides comprehensive guidance for AI coding agents working on the MockXago project.

## Project Context

MockXago is a lightweight Go mock implementation of the Xago API, designed to support local development and testing of wallet applications that integrate with Xago for multi-currency payments, KYC verification, and balance account management.

### Why MockXago Exists

Wallet applications that integrate with Xago typically use it for:
- User identity and KYC verification (including Persona integration)
- Multi-currency balance accounts (ZAR, USD)
- Bank account beneficiary management
- Fund deposits and transfers
- Transaction history and reporting

MockXago removes the dependency on real Xago credentials and services, enabling:
- Fully local development without external dependencies
- Automated testing without API rate limits
- Predictable behavior for CI/CD pipelines
- Rapid iteration without affecting real Xago sandbox data
- Test-mode endpoints for simulating deposits and transfers

### Critical Constraints

1. **API Compliance**: MockXago must be a drop-in replacement for Xago. Applications expect exact API compatibility.
2. **Sandbox Parity Only**: Focus on happy paths and sandbox environment behavior. Production Xago features are out of scope.
3. **Multi-Currency Required**: Support ZAR (South African Rand) and USD (US Dollar) balance accounts.
4. **Test Mode**: Test-only endpoints (under `/v1/test/*`) must be gated by `XAGO_MOCK_TEST_MODE=true` environment variable.
5. **Webhook Delivery**: Must support async webhook delivery with HMAC signature authentication.

## Architecture Overview

### Tech Stack

- **Language**: Go 1.24+
- **HTTP Router**: chi v5 (lightweight, idiomatic)
- **Storage**: In-memory (with mutex-based concurrency control)
- **Containerization**: Docker multi-stage build
- **Testing**: Godog for BDD-style integration tests, testify for assertions

### Directory Structure

```
go/mock/mockxago/
├── cmd/mockxago/              # Application entry point
│   └── main.go                # HTTP server setup, routing
├── internal/                  # Private application code
│   ├── auth/                  # JWT token generation/validation
│   │   ├── auth.go
│   │   └── auth_test.go
│   ├── handler/               # HTTP handlers
│   │   ├── handler.go         # Handler struct & dependencies
│   │   ├── auth.go            # /v1/login endpoints
│   │   ├── auth_test.go
│   │   ├── accounts.go        # /v1/company/accounts endpoints
│   │   ├── accounts_test.go
│   │   ├── balances.go        # /v1/accounts/{id}/balance endpoints
│   │   ├── balances_test.go
│   │   ├── beneficiaries.go   # /v1/accounts/{id}/beneficiaries endpoints
│   │   ├── beneficiaries_test.go
│   │   ├── currencies.go      # /v1/currencies endpoints
│   │   ├── currencies_test.go
│   │   ├── transactions.go    # /v1/company/transactions endpoints
│   │   ├── transactions_test.go
│   │   ├── transfers.go       # /v1/transfers endpoints (stubs)
│   │   ├── transfers_test.go
│   │   ├── persona.go         # Persona KYC endpoints
│   │   ├── persona_test.go
│   │   ├── kyc.go             # Legacy KYC iframe endpoints
│   │   ├── kyc_test.go
│   │   ├── test_helpers.go    # Test-mode endpoints (/v1/test/*)
│   │   ├── test_helpers_test.go
│   │   └── health.go          # Health check
│   ├── jobs/                  # Background jobs
│   │   ├── beneficiary_approval.go  # Auto-approve beneficiaries after 3s
│   │   └── webhook_sender.go        # Async webhook delivery
│   ├── logger/                # Logging
│   │   └── logger.go          # Structured logger (zap)
│   ├── models/                # Domain & API models
│   │   ├── models.go          # SubAccount, Beneficiary, Transaction
│   │   └── api.go             # Request/response DTOs
│   ├── storage/               # Storage layer
│   │   ├── storage.go         # In-memory storage implementation
│   │   └── storage_test.go
│   └── utils/                 # Utilities
│       ├── signature.go       # HMAC signature generation
│       └── utils_test.go
├── features/                  # BDD feature files (Gherkin)
│   ├── auth_user_kyc.feature
│   ├── cards.feature
│   ├── fee_configuration.feature
│   ├── rates_and_vaults.feature
│   ├── service_health.feature
│   ├── signature_authentication.feature
│   ├── transactions.feature
│   └── wallets_and_balances.feature
├── testenv/                   # Integration test environment
│   ├── docker-compose.yml     # Test-only compose
│   ├── godog_test.go          # BDD test runner
│   ├── *_steps.go             # Step definitions
│   └── helpers.go             # Test helpers
├── web/                       # Static web assets
│   └── kyc-iframe.html        # KYC iframe HTML
├── Dockerfile                 # Multi-stage Docker build
├── go.mod                     # Go module definition
├── go.sum                     # Dependency checksums
├── Makefile                   # Build commands
├── README.md                  # User documentation
└── AGENTS.md                  # This file
```

### Design Principles

1. **Dependency Injection**: Handler receives storage & webhook sender via constructor
2. **In-Memory Storage**: Uses `sync.RWMutex` for thread-safe access to data structures
3. **Table-Driven Tests**: Use testify's suite pattern for comprehensive coverage
4. **Idiomatic Go**: Follow standard project layout, effective Go patterns
5. **Minimal Dependencies**: Only essential libraries (chi, jwt-go, testify, godog, zap)

## Core Functionality

### 1. Storage Layer

**Interface** (`internal/storage/interface.go`):
Defines all storage operations with two implementations:

**Memory Storage** (`internal/storage/memory.go`):
- Used for unit tests
- Fast, no external dependencies
- Thread-safe with `sync.RWMutex`

**Redis Storage** (`internal/storage/redis.go`):
- Used for E2E tests and production deployments
- Persistent, production-ready
- Connection pooling (10 connections, 2 min idle)
- Atomic operations for balance updates

**Key Patterns** (Redis):
```
token:{tokenValue}                        → JSON (with TTL)
token:account:{tokenValue}                → accountId
subaccount:{accountId}                    → JSON
subaccount:wallet:{walletId}              → accountId
beneficiary:{beneficiaryId}               → JSON
beneficiaries:wallet:{walletId}           → List of beneficiary IDs
transaction:{transactionId}               → JSON
transactions:account:{accountId}          → List of transaction IDs
balance:{walletId}:{currency}:available   → Float64 (atomic)
balance:{walletId}:{currency}:reserved    → Float64 (atomic)
deposit:{depositId}                       → JSON
deposit:ref:{reference}                   → depositId
deposits:all                              → List of deposit IDs
job:{jobId}                               → JSON
jobs:ready                                → Sorted set (score = notBefore timestamp)
```

**Key Methods**:
- `CreateSubAccount(account *models.SubAccount) error`
- `GetSubAccount(accountId string) (*models.SubAccount, error)`
- `GetSubAccountByWallet(walletId string) (*models.SubAccount, error)`
- `UpdateSubAccount(account *models.SubAccount) error`
- `AddBeneficiary(accountId string, ben *models.Beneficiary) error`
- `GetBeneficiaries(accountId string, limit, page int) ([]*models.Beneficiary, int, error)`
- `GetBalance(accountId, currency string) (float64, error)`
- `UpdateBalance(accountId, currency string, amount float64) error`
- `CreateTransaction(tx *models.Transaction) error`
- `GetTransaction(txId string) (*models.Transaction, error)`

**Thread Safety**:
- Uses `sync.RWMutex` for concurrent read/write access
- Read operations use `RLock()` / `RUnlock()`
- Write operations use `Lock()` / `Unlock()`

### 2. Authentication (JWT Tokens)

**Format**:
JWT tokens with HS256 signing, expires in 55 minutes.

**Claims**:
```go
{
  "account_id": "sub-account-uuid",  // Optional, for context
  "exp": 1234567890                  // Expiration timestamp
}
```

**Login Flow**:
1. `POST /v1/login` with `apiPublicKey` and `apiSecretKey`
2. Validate against environment variables `XAGO_API_PUBLIC_KEY` and `XAGO_API_SECRET`
3. Generate JWT token with 55-minute expiration
4. Return `{tokenValue: "..."}`

**Request Headers**:
- `Authorization: Bearer {tokenValue}`

**Implementation Notes**:
- Token validation is optional in test mode for flexibility
- Tokens can be mapped to accountId for context-aware operations

### 3. Multi-Currency System

**Supported Currencies**:
```go
ZAR (South African Rand)
USD (US Dollar)
```

**Balance Behavior**:
- `GET /v1/accounts/{accountId}/balance` returns all currencies
- Format: `[{"currencyCode": "ZAR", "amount": 10000.00}, {"currencyCode": "USD", "amount": 5000.00}]`
- If balance is 0.00, still include the currency in response

**Currency Details**:
`GET /v1/currencies` returns bank deposit details for each currency:
```json
[
  {
    "currencyCode": "ZAR",
    "accountNumber": "1234567890",
    "accountName": "Xago ZAR Account",
    "bankName": "Standard Bank",
    "branchCode": "051001",
    "swiftCode": "SBZAZAJJ"
  },
  {
    "currencyCode": "USD",
    "accountNumber": "9876543210",
    "accountName": "Xago USD Account",
    "bankName": "First National Bank",
    "branchCode": "250655",
    "swiftCode": "FIRNZAJJ"
  }
]
```

### 4. KYC (Know Your Customer) Flow

MockXago supports two KYC flows:

#### Legacy KYC Iframe Flow
1. `GET /kyc/iframe?token={token}&user_id={walletId}` – Serves KYC iframe HTML
2. `POST /kyc/submit` – Iframe form submission
3. Webhook `id.verification.accepted` emitted asynchronously

#### Persona KYC Flow (Primary)
1. `POST /v1/inquiries` – Creates inquiry, returns iframe URL
2. `GET /v1/inquiries/{inquiryId}` – Get inquiry status
3. `GET /v1/inquiries/{inquiryId}/iframe` – Serves Persona KYC iframe
4. `POST /v1/inquiries/{inquiryId}/submit` – Iframe form submission
5. Webhook to `PERSONA_WEBHOOK_URL` (mimics Persona's format)
6. Webhook `id.verification.accepted` to `WEBHOOK_URL` (wallet-facing)

**Approval Logic**:
- KYC is auto-approved on submission
- Final status: `"approved"` with `riskLevel: "low"`
- Webhooks are emitted asynchronously via goroutine

**KYC Iframe** (`web/kyc-iframe.html`):
- Uses `FormData` to submit as `multipart/form-data`
- Posts message to parent window on success: `{type: 'OnboardingCompleted', value: {...}}`
- Sends `{type: 'OnboardingError', value: {message}}` on failure

### 5. Sub-Account Management

**Create Sub-Account**:
- `POST /v1/company/accounts`
- Input: `{firstName, lastName, email, mobileNumber, identityType, idNumber, physicalAddress, thirdPartyVerificationUrl, walletId (optional)}`
- Generate UUID for `accountId` if not provided
- Store sub-account with deposit references
- Return sub-account object with bank deposit details

**Update Sub-Account**:
- `PUT /v1/company/accounts/{accountId}`
- Updates KYC-related fields
- Used for KYC status updates

**Get Sub-Account by Wallet**:
- `GET /v1/company/accounts?walletId={walletId}`
- Returns sub-account associated with wallet
- Used by wallet backend to resolve accountId from walletId

### 6. Beneficiary Management

**Add Beneficiary**:
- `POST /v1/accounts/{accountId}/beneficiaries`
- Input: `{name, accountNumber, currencyCode, branchCode (optional), bankName (optional), accountName (optional)}`
- Create beneficiary with `status: "pending"`
- Auto-approve after 3 seconds (background job)
- Return beneficiary object

**List Beneficiaries**:
- `GET /v1/accounts/{accountId}/beneficiaries?limit={limit}&page={page}`
- Supports pagination (default: limit=10, page=1)
- Returns array of beneficiaries with pagination metadata

**Auto-Approval**:
Beneficiaries are automatically transitioned from `"pending"` to `"approved"` after 3 seconds (simulating sandbox approval process).

Implementation: `internal/jobs/beneficiary_approval.go`

### 7. Transaction Handling

**Types**:
1. **DEPOSIT**: External deposit (e.g., bank transfer)
   - Creates transaction record
   - Updates balance
   - Emits webhook to `WEBHOOK_URL`

2. **WITHDRAWAL**: Transfer to beneficiary
   - Creates transfer record
   - Deducts balance
   - Auto-completes in mock (no async processing)

**Test-Mode Endpoints** (`XAGO_MOCK_TEST_MODE=true`):
- `POST /v1/test/balances/deposit` - Simulate deposit + webhook
- `POST /v1/test/balances/set` - Set balance directly
- `POST /v1/test/balances/transfer` - Simulate transfer
- `POST /v1/test/transactions` - Create test transaction

**Transaction Listing**:
- `GET /v1/company/transactions?limit={limit}&page={page}` - List transactions
- `GET /v1/company/transactions/{id}` - Get single transaction

### 8. Webhook System

**Manager** (`internal/jobs/webhook_sender.go`):
```go
func SendWebhookAsync(url string, payload interface{}, secret string)
```

**Event Types**:
1. **KYC Verification**: `id.verification.accepted`
2. **Deposit Completed**: Deposit transaction payload

**Event Format (KYC)**:
```json
{
  "event_type": "id.verification.accepted",
  "wallet_id": "wallet-uuid",
  "timestamp": "2026-02-23T10:00:00Z",
  "data": {
    "message": "Persona KYC verification accepted"
  }
}
```

**Event Format (Deposit)**:
```json
{
  "accountId": "account-uuid",
  "amount": 5000.00,
  "currencyCode": "ZAR",
  "transactionId": "tx-uuid",
  "code": 104,
  "status": "settled",
  "createdAt": "2026-02-23T10:00:00Z",
  "settledAt": "2026-02-23T10:00:00Z"
}
```

**Delivery**:
- Async (goroutine)
- 3 retry attempts with exponential backoff (1s, 2s, 4s)
- Signs with HMAC-SHA256: `X-Signature` header = hex(HMAC-SHA256(body, secret))
- Logs errors but does not block main flow

**Persona Webhooks**:
MockXago also sends webhooks to `PERSONA_WEBHOOK_URL` mimicking Persona's format:
```json
{
  "type": "inquiry.completed",
  "data": {
    "id": "inquiry-id",
    "attributes": {
      "status": "approved",
      "reference-id": "wallet-id"
    }
  }
}
```

### 9. Pagination Pattern

All list endpoints support pagination:

**Query Parameters**:
- `limit`: Items per page (default: 10, max: 100)
- `page`: Page number (1-indexed, default: 1)

**Response Format**:
```json
{
  "data": [...],
  "pagination": {
    "limit": 10,
    "page": 1,
    "numberOfPages": 5,
    "total": 42
  }
}
```

**Implementation Helper**:
```go
func paginateSlice(items []interface{}, limit, page int) ([]interface{}, *PaginationMeta)
```

## Testing Strategy

### Coverage Goal: 80%+

### BDD Integration Tests (Godog)

**Location**: `testenv/`

**Feature Files**: `features/*.feature` (Gherkin scenarios)

**Step Definitions**: `testenv/*_steps.go`

**Running Tests**:
```bash
# All features
go test ./testenv/...

# Specific feature
go test ./testenv/ -godog.tags="@auth"

# With verbose output
go test ./testenv/ -v -godog.format=pretty
```

**Test Environment**:
- Uses docker-compose.yml with isolated MockXago instance
- Ports: 28080 (MockXago), 28081 (webhook receiver)
- Clean state for each test scenario

**Coverage**:
- 91/91 scenarios passing (100% pass rate)
- Coverage across all domains: auth, accounts, balances, beneficiaries, transactions, KYC, etc.

### Unit Tests

**Handler Tests** (`internal/handler/*_test.go`):
- Use `httptest.NewRecorder()` for response capture
- Table-driven tests for multiple scenarios
- Test both success and error cases

Example:
```go
func TestCreateSubAccount(t *testing.T) {
    tests := []struct {
        name       string
        body       string
        wantStatus int
        wantErr    bool
    }{
        {"valid account", `{"firstName":"John","lastName":"Doe",...}`, 201, false},
        {"missing firstName", `{"lastName":"Doe"}`, 400, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

**Storage Tests** (`internal/storage/storage_test.go`):
- Concurrent access tests (goroutines)
- Edge case handling
- Balance update atomicity

## Common Patterns

### Error Handling

```go
func (h *Handler) CreateSubAccount(w http.ResponseWriter, r *http.Request) {
    var req CreateSubAccountRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "Invalid request body")
        return
    }
    
    // Validation
    if req.FirstName == "" {
        respondError(w, http.StatusBadRequest, "firstName is required")
        return
    }
    
    // Business logic
    account, err := h.storage.CreateSubAccount(&req)
    if err != nil {
        logger.Log.Error("Failed to create sub-account", zap.Error(err))
        respondError(w, http.StatusInternalServerError, "Internal server error")
        return
    }
    
    respondJSON(w, http.StatusCreated, account)
}
```

### JSON Response Helpers

```go
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
    respondJSON(w, status, map[string]string{"error": message})
}
```

### Async Operations

```go
// Launch webhook delivery in background
go jobs.SendWebhookAsync(webhookURL, event, webhookSecret)
```

### Concurrency-Safe Storage Access

```go
// Read operation
func (s *Storage) GetSubAccount(accountId string) (*models.SubAccount, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    account, exists := s.subAccounts[accountId]
    if !exists {
        return nil, fmt.Errorf("sub-account not found")
    }
    return account, nil
}

// Write operation
func (s *Storage) CreateSubAccount(account *models.SubAccount) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    s.subAccounts[account.AccountID] = account
    s.subAccountsByWallet[account.WalletID] = account.AccountID
    return nil
}
```

## Logging Guidelines

**Important**: It's safe to log sensitive values in MockXago because:
- This is a development/testing mock service, not a production system
- Applications running against it are also in local test environments
- Verbose logging helps with debugging integrations
- No real credentials or production data flows through this service

**Logging Standards**:
- Use zap structured logging via `logger.Log.Info()`, `logger.Log.Warn()`, `logger.Log.Error()`, `logger.Log.Debug()`
- Include contextual fields: `zap.String("key", value)`, `zap.Error(err)`, `zap.Float64("amount", amt)`
- Log all significant operations: account creation, balance updates, webhook delivery
- Include identifiers (accountId, walletId, transactionId) for traceability

**Log Levels**:
- `Info`: Normal operations (account created, deposit received, webhook sent)
- `Warn`: Non-fatal issues (invalid input, missing optional fields)
- `Error`: Operations that failed (storage error, webhook delivery failed)
- `Debug`: Detailed diagnostic info (token validation, balance calculations)

**Example**:
```go
logger.Log.Info("deposit processed successfully", 
    zap.String("account_id", accountID),
    zap.Float64("amount", amount),
    zap.String("currency", currency),
    zap.String("transaction_id", txID),
)
```

## Configuration

**Environment Variables**:
```bash
XAGO_MOCK_PORT=8080                          # HTTP port
XAGO_API_PUBLIC_KEY=test-public-key          # Expected login public key
XAGO_API_SECRET=test-secret                  # Expected login secret key
XAGO_MOCK_TEST_MODE=true                     # Enable /v1/test/* endpoints

# Storage backend (Redis recommended for E2E/production)
REDIS_URL=redis:6379                         # Redis connection (or MOCKXAGO_REDIS_URL)
REDIS_DB=0                                   # Redis database number (or MOCKXAGO_REDIS_DB)

# Webhooks
WEBHOOK_URL=http://backend:8080/xago/webhooks/event  # Wallet webhook URL
WEBHOOK_SECRET=local-webhook-secret          # HMAC secret for webhooks
PERSONA_WEBHOOK_URL=http://backend:8080/webhooks/persona  # Persona webhook URL
PERSONA_WEBHOOK_TOKEN=persona-hmac-secret    # HMAC secret for Persona webhooks
```

**Docker Compose Integration**:
Configure in your `docker-compose.yml`:
```yaml
mockxago:
  image: ghcr.io/interledger/mockxago:latest
  ports:
    - "8080:8080"
  environment:
    - XAGO_MOCK_PORT=8080
    - XAGO_API_PUBLIC_KEY=test-public-key
    - XAGO_API_SECRET=test-secret
    - XAGO_MOCK_TEST_MODE=true
    - WEBHOOK_URL=http://backend:8080/xago/webhooks/event
    - WEBHOOK_SECRET=local-webhook-secret
    - PERSONA_WEBHOOK_URL=http://backend:8080/webhooks/persona
    - PERSONA_WEBHOOK_TOKEN=persona-hmac-secret
```

## Development Workflow

### 1. Making Changes

```bash
go mod tidy                    # Update dependencies
go test ./...                  # Run all tests (unit + integration)
go build ./cmd/mockxago        # Build binary
```

### 2. Running Locally

```bash
# Basic mode
./mockxago

# With test mode enabled
export XAGO_MOCK_TEST_MODE=true
export WEBHOOK_URL=http://localhost:3000/xago/webhooks
export WEBHOOK_SECRET=test-secret
./mockxago
```

### 3. Docker Build & Test

```bash
# Build fresh image
docker build -t local-mockxago .

# Test in isolated environment
go test ./testenv/...

# Deploy with Docker Compose
docker compose up -d mockxago
docker compose logs -f mockxago
```

### 4. Full Integration Testing

```bash
# Run all BDD tests
go test ./testenv/ -v -godog.format=pretty

# Run specific feature
go test ./testenv/ -godog.tags="@beneficiaries"

# Check test coverage
go test ./... -cover
```

## Troubleshooting

### Tests failing with "sub-account not found"
**Cause**: Storage state not properly initialized or cleaned between tests  
**Solution**: Ensure `NewStorage()` is called for each test scenario, or call `Reset()` if storage has a reset method

### Webhooks not arriving
**Cause**: `WEBHOOK_URL` incorrect or backend not listening  
**Solution**:
- Check `WEBHOOK_URL` environment variable
- Verify backend is running: `curl http://backend:8080/health`
- Check MockXago logs: `docker compose logs mockxago`
- Verify signature validation on receiving end

### Docker build fails
**Cause**: Missing dependencies or incorrect paths  
**Solution**:
- Ensure `go.mod` and `go.sum` are present
- Run `go mod tidy` before building
- Check Dockerfile paths match actual structure

### Balance not updating
**Cause**: Currency mismatch or accountId not found  
**Solution**:
- Verify currency code matches exactly: "ZAR" or "USD"
- Check accountId exists: `GET /v1/company/accounts?walletId={walletId}`
- Check balance endpoint: `GET /v1/accounts/{accountId}/balance`

### KYC webhook not triggering
**Cause**: Webhook URL not configured or backend not accepting webhooks  
**Solution**:
- Set `WEBHOOK_URL` in environment
- Check webhook signature validation on backend
- Verify backend endpoint exists: `POST /xago/webhooks/event`

### Test mode endpoints returning 404
**Cause**: `XAGO_MOCK_TEST_MODE` not set to `true`  
**Solution**: Export environment variable before starting: `export XAGO_MOCK_TEST_MODE=true`

## API Compliance Analysis

### Known Divergences from Official Xago API

1. **Beneficiary Endpoints Path**:
   - **Official API**: `POST /v1/beneficiaries`, `GET /v1/beneficiaries`
   - **MockXago**: `POST /v1/accounts/{accountId}/beneficiaries`, `GET /v1/accounts/{accountId}/beneficiaries`
   - **Impact**: Medium - beneficiaries are scoped to accountId in URL path
   - **Status**: Accepted divergence for clarity, may add aliases in future

2. **Transaction Query Endpoint**:
   - **Official API**: `GET /v1/transactions?transactionId=X`
   - **MockXago**: `GET /v1/company/transactions/{id}`
   - **Impact**: Low - wallet backend uses specific transaction lookup by ID
   - **Status**: Works for current wallet needs

3. **Test-Mode Endpoints**:
   - **MockXago Only**: `/v1/test/*` endpoints for balance manipulation
   - **Purpose**: Essential for local testing and E2E automation
   - **Status**: Intentional addition, not in official API

### Recommendations for Future Work

**Phase 12: API Compliance Improvements** (Optional):
- Add `/v1/beneficiaries` endpoints as aliases
- Add `/v1/transactions?transactionId=X` query support
- Keep existing endpoints for backward compatibility

**Effort**: 2-3 hours  
**Benefit**: Full API compliance, easier migration to real Xago

## AI Agent Best Practices

### When Adding New Endpoints

1. **Define Models**: Add request/response DTOs to `internal/models/api.go`
2. **Implement Handler**: Add method to appropriate `internal/handler/{domain}.go`
3. **Add Route**: Register in `cmd/mockxago/main.go` setupRoutes
4. **Write Tests**: Create table-driven test in `{domain}_test.go`
5. **Add BDD Scenario**: Create Gherkin scenario in `features/{domain}.feature`
6. **Add Step Definitions**: Implement steps in `testenv/{domain}_steps.go`
7. **Update Docs**: Add endpoint to README.md API section

### When Modifying Storage

1. **Update Interface**: Check if `Storage` interface needs new methods
2. **Thread Safety**: Use appropriate mutex locks (RLock for reads, Lock for writes)
3. **Add Tests**: Cover new functionality in `storage_test.go` with concurrent access tests
4. **Update Seeder**: If affecting initialization, update data seeding logic

### When Changing Webhook Behavior

1. **Update Payload**: Modify struct in `internal/models/api.go`
2. **Update Sender**: Modify `internal/jobs/webhook_sender.go` if changing retry logic
3. **Test Delivery**: Add test scenario for new webhook type
4. **Update Docs**: Document webhook payload format in README.md

### Testing Checklist

Before committing:
- [ ] All tests pass: `go test ./...`
- [ ] BDD tests pass: `go test ./testenv/ -v`
- [ ] Coverage acceptable: `go test ./... -cover` (aim for 80%+)
- [ ] Docker build succeeds: `docker build -t local-mockxago .`
- [ ] No race conditions: `go test ./... -race`
- [ ] Code formatted: `gofmt -w .`
- [ ] Documentation updated if API changed

### Critical: Maintain testenv/

**The `testenv/` directory is NOT optional**. Future agents MUST maintain it when making changes:

1. **When adding new endpoints**: Add corresponding BDD scenarios to feature files
2. **When changing API responses**: Update step definitions and assertions
3. **When adding new features**: Add comprehensive test coverage
4. **When modifying authentication**: Ensure test headers/tokens are valid

**testenv/ provides**:
- Isolated integration testing (no conflicts with main environment)
- Fast feedback loop for full-stack changes
- Regression prevention for critical user journeys
- CI/CD validation readiness

**Running tests**:
```bash
go test ./testenv/...  # Runs all BDD tests
go test ./... -cover   # Runs all tests with coverage
```

**Expected outcome**: 91/91 scenarios passing

**If tests fail after your changes**:
1. Check what changed in API responses
2. Update step definitions in `testenv/*_steps.go`
3. Update assertions to match new behavior
4. Ensure backward compatibility where possible
5. Document breaking changes clearly

## Key Files Reference

**Must Review Before Coding**:
1. `internal/models/models.go` - Domain models (SubAccount, Beneficiary, Transaction, Inquiry)
2. `internal/models/api.go` - Request/response DTOs
3. `internal/storage/storage.go` - Storage interface and implementation
4. `cmd/mockxago/main.go` - Routing configuration

**Frequently Modified**:
1. `internal/handler/*.go` - API endpoint implementations
2. `internal/storage/storage.go` - Storage logic
3. `internal/jobs/*.go` - Background job implementations (webhooks, beneficiary approval)
4. `features/*.feature` - BDD test scenarios
5. `testenv/*_steps.go` - BDD step definitions

**Rarely Touch**:
1. `internal/logger/logger.go` - Basic logging setup
2. `internal/auth/auth.go` - JWT token generation
3. `Dockerfile` - Container build configuration
4. `web/kyc-iframe.html` - KYC iframe HTML

## Success Metrics

Your changes should maintain or improve:
- **Test Coverage**: ≥80%
- **BDD Pass Rate**: 100% (91/91 scenarios)
- **API Compliance**: Applications run without modification
- **Docker Build Time**: Keep under 2 minutes
- **Response Time**: All endpoints < 100ms (local)
- **Memory Usage**: < 50MB for in-memory storage

## Implementation Status Reference

| Phase | Status | Notes |
|-------|--------|-------|
| 1: Authentication & Foundation | ✅ Complete | JWT tokens, login, health check |
| 2: Sub-Account Management | ✅ Complete | Create, update, query by wallet |
| 3: Currencies & Deposit Details | ✅ Complete | Bank details for ZAR and USD |
| 4: Balance Management | ✅ Complete | Multi-currency balances, updates |
| 5: Beneficiary Management | ✅ Complete | Add, list, auto-approval |
| 6: Transactions & Transfers | ✅ Complete | Create, list, query |
| 7: Deposits & Webhooks | ✅ Complete | Test deposits, webhook delivery |
| 8: Integration Testing | ✅ Complete | 91/91 BDD scenarios passing |
| 9: Redis Storage | ⬜ Optional | In-memory sufficient for tests |
| 10: Docker & Deployment | ✅ Complete | Multi-stage build, compose |
| 11: Documentation & Polish | ✅ Complete | README.md, AGENTS.md |
| 12: API Compliance Improvements | ⬜ Optional | Beneficiary endpoint aliases |

## Questions? Issues?

When encountering ambiguity:
1. Check existing implementation in similar endpoints
2. Refer to Xago API documentation (if accessible)
3. Test against wallet backend's expected behavior
4. Check BDD scenarios in `features/` for expected behavior
5. Default to simplest solution that maintains API compatibility

Remember: MockXago is a development tool. Prioritize simplicity, testability, and API compatibility over feature completeness.

---

**Last Updated**: February 23, 2026  
**Maintainers**: Interledger Foundation  
**Repository**: https://github.com/interledger/interledger-app
