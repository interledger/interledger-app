# GitHub Copilot Instructions for Interledger Wallet

## Project Context

**Purpose**: Interledger Wallet is a comprehensive fintech application enabling users to manage digital wallets, make payments, handle KYC verification, and interact with the Interledger Protocol network.

**Why it exists**: Provides a production-grade reference implementation of a wallet application integrating with multiple payment providers (GateHub, Xago, Chimoney, PTI), Interledger Protocol via Rafiki, and identity management via Ory Kratos.

**Tech Stack**:
- **Backend**: Go 1.24+ (Chi HTTP router, gRPC, Temporal workflows)
- **Frontend**: TypeScript with Remix framework, Tailwind CSS
- **Database**: PostgreSQL 17 with Atlas migrations (HCL format)
- **Identity/Auth**: Ory Kratos (self-hosted)
- **Workflow Engine**: Temporal
- **Interledger**: Rafiki connector
- **Reverse Proxy**: Traefik with TLS
- **Cache/Queue**: Redis
- **Testing**: Go standard library, testify, Godog (BDD)
- **Protobuf**: Buf for code generation

## Repository Structure

```
interledger-app/
├── go/                          # Go workspace root (module: gitlab.com/fynbos)
│   ├── backend/                 # Main wallet backend service
│   │   ├── main.go              # Entry point (subcommands: start, migrate, worker, dev)
│   │   ├── db/                  # Database schema (Atlas HCL)
│   │   │   └── schema.hcl       # Source of truth for DB schema
│   │   ├── payments/            # Payment domain
│   │   │   ├── *.go             # HTTP handlers, routes
│   │   │   ├── client/          # Internal client interface
│   │   │   ├── ops/             # Business logic operations
│   │   │   └── types.go         # Domain models
│   │   ├── wallets/             # Wallet domain (similar structure)
│   │   ├── kyc/                 # KYC verification domain
│   │   ├── identities/          # Identity management
│   │   ├── transactions/        # Transaction history
│   │   ├── rafiki/              # Interledger Protocol integration
│   │   ├── signup/              # User registration flows
│   │   ├── limits/              # Transaction/balance limits
│   │   ├── contacts/            # User contacts
│   │   ├── agreements/          # Terms & conditions
│   │   ├── linkedaccounts/      # Provider account linking
│   │   ├── providers/           # Payment provider integrations
│   │   │   ├── gatehub/         # GateHub provider
│   │   │   ├── xago/            # Xago provider
│   │   │   ├── chimoney/        # Chimoney provider
│   │   │   └── pti/             # PTI/BOS provider
│   │   ├── temporal/            # Temporal workflow definitions
│   │   ├── jobs/                # Async job implementations
│   │   ├── middleware/          # HTTP middleware (auth, logging, etc.)
│   │   └── Makefile             # Backend build/test commands
│   ├── mock/                    # Mock services
│   │   ├── mockxago/             # Mock Xago payment provider
│   │   └── mockbos/              # Mock BOS/PTI service
│   ├── pacioli/                 # Double-entry accounting ledger (gRPC)
│   ├── proto/                   # Generated Go protobuf code
│   ├── env/                     # Environment utilities
│   ├── geo/                     # Geolocation services
│   ├── log/                     # Logging utilities
│   ├── tracing/                 # Distributed tracing
│   ├── go.mod                   # Go module definition
│   ├── Makefile                 # Go workspace commands
│   └── coverage.thresholds      # Test coverage requirements
├── proto/                       # Protobuf source definitions
│   ├── buf.gen.yaml             # Buf generate configuration
│   ├── backend/                 # Backend service proto files
│   ├── pacioli/                 # Pacioli service proto files
│   ├── geo/                     # Geo service proto files
│   └── Makefile                 # Proto generation commands
├── typescript/                  # Frontend applications
│   ├── protea/                  # Main user-facing wallet UI
│   ├── botanist/                # Admin panel
│   └── hortus/                  # Additional UI app
├── local/                       # Docker Compose development environment
│   ├── docker-compose.yaml      # Main compose file
│   ├── Makefile                 # Local environment commands
│   ├── config/                  # Service configurations
├── e2e/                         # BDD E2E tests (Godog)
│   ├── godog_test.go            # Test runner
│   ├── features/                # Gherkin feature files
│   ├── *_steps.go               # Step definitions
│   ├── AGENTS.md                # E2E testing guide for AI agents
│   └── STEP_REFERENCE.md        # Available step definitions
│   ├── config/                  # Service configurations
│   ├── initdb/                  # Database initialization scripts
│   ├── scripts/                 # Helper scripts
│   ├── *.yaml                   # Service-specific compose files
│   └── docs/                    # Local environment documentation
├── go.work                      # Go workspace configuration
├── renovate.json                # Dependency update automation
├── CLAUDE.md                    # AI assistant instructions (legacy)
└── README.md                    # Project documentation
```

## Critical Constraints

1. **Test-Driven Development**: Always write or update tests before making changes. Run unit tests first, then E2E tests.
2. **Small Incremental Changes**: Make one focused change at a time, verify it works, then move on.
3. **Refactor Aggressively**: Clean up surrounding code when adding features. Extract shared logic, consolidate patterns, remove duplication.
4. **No Sensitive Data in Logs**: Never log passwords, tokens, API keys, or PII. See `docs/logging-policy.md`.
5. **Database Changes via Atlas**: Schema changes must be made in `go/backend/db/schema.hcl`, not raw SQL.
6. **Phone Number Format**: Phone numbers must use E.164 format (e.g., `+49987654321`). Kratos validation is strict.

## Architecture Overview

### Backend Domain Pattern

Every domain follows a consistent structure:

```
domain/
├── *.go                   # HTTP handlers, routes, middleware
├── client/                # Internal client interface (for cross-domain calls)
│   └── client.go
├── ops/                   # Business operations (core logic)
│   ├── operation1.go
│   ├── operation2.go
│   └── operation_test.go
└── types.go or types/     # Domain models, DTOs, enums
```

**Key Domains**:
- `payments/` - Payment processing, provider orchestration
- `wallets/` - Wallet CRUD, balance management
- `kyc/` - KYC verification flows (Persona, provider-specific)
- `identities/` - User identity management (interfaces with Kratos)
- `transactions/` - Transaction history, ledger queries
- `rafiki/` - Interledger Protocol integration
- `signup/` - User registration, onboarding
- `limits/` - Transaction and balance limits
- `contacts/` - User contact management
- `agreements/` - Terms of service, user agreements
- `linkedaccounts/` - Payment provider account linking

### Payment Providers

Located in `go/backend/providers/{provider}/`:
- **GateHub** (`providers/gatehub/`) - Multi-currency custodial wallet
- **Xago** (`providers/xago/`) - South African payment provider
- **Chimoney** (`providers/chimoney/`) - International payments
- **PTI/BOS** (`providers/pti/`) - Payment processing

Each provider implements a common interface with `client/` and `ops/` packages.

### Supporting Services

**Pacioli** (`go/pacioli/`)
- Double-entry accounting ledger
- Accessed via gRPC
- Records all financial transactions
- Ensures balance integrity

**Rafiki** (`go/backend/rafiki/`)
- Interledger Protocol connector integration
- Payment pointers, ILP packets
- Cross-border payments

**Temporal** (`go/backend/temporal/`, `go/backend/jobs/`)
- Async workflow orchestration
- Long-running operations (KYC, deposits, withdrawals)
- Retry logic, state management
- Worker processes run separately from HTTP server

### Key Infrastructure Dependencies

| Service | Purpose | Port |
|---------|---------|------|
| PostgreSQL 17 | Primary database | 5433 |
| Temporal | Workflow engine | 7233 |
| Ory Kratos | Identity/auth | 4433/4434 |
| Rafiki | ILP connector | Various |
| Traefik | Reverse proxy/router | 80/443 |
| Redis | Caching, session storage | 6379 |

## Development Workflow

### Test-Driven Development (TDD)

**CRITICAL**: Always follow this workflow:

1. **Make a small change** — Focus on one thing at a time. Avoid large, sweeping modifications.
2. **Run unit tests** — Verify the change compiles and passes unit tests first.
   ```bash
   cd go/backend
   go test -count=1 -v ./domain/ops/operation_test.go
   # OR run all domain tests
   go test -count=1 -v ./domain/...
   ```
3. **Run E2E tests** — Once unit tests pass, run relevant E2E scenarios.
   ```bash
   cd e2e
   go test -v -timeout 10m -args -tags "@domain&&@provider"
   ```
4. **Commit** — Only commit once both unit and E2E tests pass.
5. **Move on** — Proceed to the next change and repeat.

**Do not batch multiple unrelated changes into a single step.** If something breaks, it should be immediately obvious which change caused it.

### Refactoring Guidelines

When adding features or fixing bugs:
- **Extract shared logic** into helpers or utility functions
- **Consolidate repeated patterns** (e.g., HTTP error handling, validation)
- **Remove dead code** immediately
- **Factor out test helpers** rather than copy-pasting test setup
- **Simplify complex conditionals** with early returns or guard clauses

## Build & Development Commands

### Local Environment (from `local/`)

```bash
# Start infrastructure (Traefik, Postgres, Redis)
make infra

# Start services (Kratos, Temporal, Rafiki) — requires infra
make svc

# Start application (backends + frontends) with watch mode
make app

# Start everything with watch mode
make all

# Start everything WITHOUT watch mode (for E2E tests)
make all-nowatch

# Stop all services
make down

# Stop all and remove volumes (full reset)
make reset

# Add required /etc/hosts entries
make hosts

# Generate self-signed TLS certificates
make certs

# Trust certificates on Debian/Ubuntu
make trust-debian
```

### Backend Testing (from `go/backend/`)

**Prerequisites**: Postgres instance + Atlas CLI for migrations.

```bash
# Install Atlas CLI (if not already installed)
curl -sSf https://atlasgo.sh | sh

# Start test Postgres instance
make testenvup

# Generate test migrations from schema.hcl
make generatetestmigrations

# Run all unit tests
make test

# Run tests for a specific package
go test -count=1 -v ./payments/...

# Run a single test file
go test -count=1 -v ./kyc/ops/persona_test.go

# Run with coverage
go test -cover ./...

# Tear down test database
make testenvdown
```

**Test Database Behavior**:
- Each test package creates its own database for parallel execution
- Database name: `test_<package_name>_<timestamp>`
- Automatically cleaned up after test completion

### Go Linting (from `go/`)

```bash
# Run golangci-lint
make lint

# CI format with code-climate output
make lintci
```

**Configuration**: `go/.golangci.yml` (timeout 10m, build tags: `codeanalysis`)

### Protobuf Generation (from `proto/`)

```bash
# Generate Go code from proto files
make gen

# This runs: buf generate + cleanup
```

### Backend Binary Subcommands

The main binary (`go/backend/main.go`) accepts:

```bash
# Start HTTP + gRPC server only
./backend start

# Run database migrations
./backend migrate

# Start Temporal worker only
./backend worker

# Start both server and worker (for development)
./backend dev
```

## Testing Strategy

### Unit Tests

**Location**: Alongside production code in each package (e.g., `ops/*_test.go`)

**Guidelines**:
- Use table-driven tests for multiple scenarios
- Use testify assertions: `require.NoError()`, `assert.Equal()`, etc.
- Each test should create its own database (handled automatically)
- Mock external dependencies (HTTP clients, provider APIs)
- Test both success and error cases

**Example**:
```go
func TestCreatePayment(t *testing.T) {
    tests := []struct {
        name       string
        input      CreatePaymentRequest
        wantErr    bool
        wantStatus PaymentStatus
    }{
        {
            name: "valid payment",
            input: CreatePaymentRequest{
                Amount:   "100.00",
                Currency: "USD",
                UserID:   "user-123",
            },
            wantErr:    false,
            wantStatus: PaymentStatusPending,
        },
        {
            name: "invalid amount",
            input: CreatePaymentRequest{
                Amount:   "-50",
                Currency: "USD",
                UserID:   "user-123",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            db := setupTestDB(t)
            defer db.Close()

            ops := NewPaymentOps(db)
            payment, err := ops.CreatePayment(context.Background(), tt.input)

            if tt.wantErr {
                require.Error(t, err)
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tt.wantStatus, payment.Status)
        })
    }
}
```

### E2E Tests (BDD with Godog)

**Location**: `e2e/`

**Key Files**:
- `features/*.feature` - Gherkin feature files (Given/When/Then scenarios)
- `*_steps.go` - Go step definitions implementing feature steps
- `godog_test.go` - Test runner configuration
- `AGENTS.md` - Comprehensive E2E guide for AI agents
- `STEP_REFERENCE.md` - Available step definitions reference

**Running E2E Tests**:
```bash
cd e2e

# Run all tests
go test -v -timeout 30m

# Run specific scenarios by tag
go test -v -timeout 5m -args -tags @signuponly
go test -v -timeout 10m -args -tags "@kyc&&@xago"
go test -v -timeout 10m -args -tags "@withdrawal&&@fees"

# Common tag combinations
go test -v -timeout 10m -args -tags "@payments"
go test -v -timeout 10m -args -tags "@gatehub"
go test -v -timeout 10m -args -tags "@rafiki"
```

**E2E Environment Notes**:
- Start with `make all-nowatch` from `local/` (NOT `make all`)
- Mock services accessible via public URLs (e.g., `mockxago.interledger.test`)
- Users are unique per test run — database cleanup has limited value
- **Never suppress test output** — visibility is critical for debugging
- Phone numbers must use E.164 format: `+49987654321`
- Kratos `tel` validation issues may require `make reset` to resolve
- MockXago serves KYC iframe at `/kyc/iframe` for South African flows

**Adding New E2E Tests**:
1. Create or update `.feature` file in `features/`
2. Implement step definitions in appropriate `*_steps.go` file
3. Use existing step definitions when possible (see `STEP_REFERENCE.md`)
4. Tag scenarios appropriately for selective execution
5. Run test to verify it works end-to-end
6. Update `AGENTS.md` if introducing new patterns

## Database & Migrations

### Schema Definition

**Source of Truth**: `go/backend/db/schema.hcl` (Atlas HCL format)

**Never write raw SQL migrations.** All schema changes must be defined in `schema.hcl`.

**Example Schema**:
```hcl
table "users" {
  schema = schema.public
  column "id" {
    type = uuid
    default = sql("gen_random_uuid()")
  }
  column "email" {
    type = text
    null = false
  }
  column "created_at" {
    type = timestamptz
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
  index "idx_users_email" {
    columns = [column.email]
    unique = true
  }
}
```

### Running Migrations

```bash
cd go/backend

# Generate migrations from schema.hcl
make generatetestmigrations

# Apply migrations (via backend binary)
./backend migrate
```

### Atlas CLI Commands

```bash
# Inspect current database schema
atlas schema inspect -u "postgres://user:pass@localhost:5433/dbname?sslmode=disable"

# Apply schema changes
atlas schema apply \
  -u "postgres://user:pass@localhost:5433/dbname?sslmode=disable" \
  --to "file://db/schema.hcl" \
  --dev-url "docker://postgres/17/dev?search_path=public"

# Generate migration diff
atlas migrate diff migration_name \
  --dir "file://db/migrations" \
  --to "file://db/schema.hcl" \
  --dev-url "docker://postgres/17/dev"
```

## Common Patterns

### HTTP Handler Structure

```go
package payments

import (
    "encoding/json"
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "gitlab.com/fynbos/backend/middleware"
)

func (h *Handler) RegisterRoutes(r chi.Router) {
    r.Group(func(r chi.Router) {
        r.Use(middleware.RequireAuth)
        r.Post("/payments", h.CreatePayment)
        r.Get("/payments/{id}", h.GetPayment)
        r.Get("/payments", h.ListPayments)
    })
}

func (h *Handler) CreatePayment(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    userID := middleware.GetUserID(ctx)
    
    var req CreatePaymentRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }
    
    // Validate request
    if err := req.Validate(); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Call business logic
    payment, err := h.ops.CreatePayment(ctx, userID, req)
    if err != nil {
        h.logger.Error("failed to create payment", 
            zap.String("user_id", userID),
            zap.Error(err))
        http.Error(w, "internal server error", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(payment)
}
```

### Business Logic (Ops) Pattern

```go
package ops

import (
    "context"
    "database/sql"
    "fmt"
)

type PaymentOps struct {
    db     *sql.DB
    logger *zap.Logger
}

func NewPaymentOps(db *sql.DB, logger *zap.Logger) *PaymentOps {
    return &PaymentOps{db: db, logger: logger}
}

func (o *PaymentOps) CreatePayment(ctx context.Context, userID string, req CreatePaymentRequest) (*Payment, error) {
    // Start transaction
    tx, err := o.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback()
    
    // Business logic
    payment := &Payment{
        ID:       generateID(),
        UserID:   userID,
        Amount:   req.Amount,
        Currency: req.Currency,
        Status:   PaymentStatusPending,
    }
    
    // Insert payment
    _, err = tx.ExecContext(ctx, `
        INSERT INTO payments (id, user_id, amount, currency, status, created_at)
        VALUES ($1, $2, $3, $4, $5, NOW())
    `, payment.ID, payment.UserID, payment.Amount, payment.Currency, payment.Status)
    if err != nil {
        return nil, fmt.Errorf("insert payment: %w", err)
    }
    
    // Commit transaction
    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("commit transaction: %w", err)
    }
    
    return payment, nil
}
```

### Error Handling

Always wrap errors with context:
```go
if err != nil {
    return nil, fmt.Errorf("operation description: %w", err)
}
```

Log errors with structured fields:
```go
logger.Error("operation failed",
    zap.String("user_id", userID),
    zap.String("payment_id", paymentID),
    zap.Error(err))
```

### Temporal Workflow Pattern

```go
package jobs

import (
    "time"
    
    "go.temporal.io/sdk/workflow"
)

func ProcessPaymentWorkflow(ctx workflow.Context, paymentID string) error {
    logger := workflow.GetLogger(ctx)
    
    // Set activity options
    opts := workflow.ActivityOptions{
        StartToCloseTimeout: 5 * time.Minute,
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts: 3,
        },
    }
    ctx = workflow.WithActivityOptions(ctx, opts)
    
    // Execute activities
    var result ProcessPaymentResult
    err := workflow.ExecuteActivity(ctx, ProcessPaymentActivity, paymentID).Get(ctx, &result)
    if err != nil {
        logger.Error("payment processing failed", "error", err)
        return err
    }
    
    logger.Info("payment processed successfully", "payment_id", paymentID)
    return nil
}
```

## Key Files Reference

### Must Review Before Coding

1. **`go/backend/db/schema.hcl`** - Database schema (source of truth)
2. **`go.work`** - Go workspace configuration
3. **`local/docker-compose.yaml`** - Service orchestration
4. **`docs/logging-policy.md`** - What NOT to log
5. **`e2e/AGENTS.md`** - E2E testing guide

### Frequently Modified

1. **`go/backend/{domain}/*.go`** - Domain handlers
2. **`go/backend/{domain}/ops/*.go`** - Business logic
3. **`go/backend/{domain}/types.go`** - Domain models
4. **`e2e/features/*.feature`** - E2E scenarios
5. **`e2e/*_steps.go`** - E2E step definitions

### Rarely Touch

1. **`go/backend/middleware/*.go`** - HTTP middleware
2. **`proto/**/*.proto`** - Protobuf definitions
3. **`local/*.yaml`** - Docker Compose service configs
4. **`go/backend/main.go`** - Application entry point

## Environment Variables

### Backend (`go/backend/`)

```bash
DATABASE_URL=postgres://user:pass@localhost:5433/wallet?sslmode=disable
REDIS_URL=redis://localhost:6379
TEMPORAL_HOST=localhost:7233
KRATOS_PUBLIC_URL=https://kratos.interledger.test
KRATOS_ADMIN_URL=http://kratos:4434
RAFIKI_ADMIN_URL=http://rafiki-admin:3001
LOG_LEVEL=debug
PORT=3000
GRPC_PORT=50051
```

### Frontend (TypeScript apps)

```bash
API_BASE_URL=https://api.interledger.test
KRATOS_PUBLIC_URL=https://kratos.interledger.test
```

## Logging Guidelines

**CRITICAL**: Follow `docs/logging-policy.md` strictly.

**Never log**:
- Passwords or password hashes
- API keys, secrets, tokens (except in mock services for debugging)
- Full credit card numbers or CVV
- Social security numbers or government IDs
- Personally identifiable information (PII) without user consent

**Safe to log**:
- User IDs (UUIDs, not emails)
- Transaction IDs
- Payment amounts and currencies
- Request/response metadata (method, path, status code)
- Timing information, performance metrics

**Use structured logging**:
```go
logger.Info("payment created",
    zap.String("payment_id", paymentID),
    zap.String("user_id", userID),
    zap.String("amount", amount),
    zap.String("currency", currency))
```

## PR Checklist

Before submitting a pull request:

- [ ] Tests updated and in place (unit + E2E where applicable)
- [ ] All tests pass: `make test` (backend), E2E scenarios
- [ ] No sensitive data in logs or error messages
- [ ] Database schema changes reflected in `db/schema.hcl`
- [ ] Linter passes: `make lint`
- [ ] Code refactored to reduce duplication
- [ ] Documentation updated if API or behavior changed
- [ ] DevOps notified of release-impacting changes:
  - New environment variables
  - Database migrations requiring downtime
  - New services or dependencies
  - Breaking changes to external integrations

## Troubleshooting

### Tests Failing with Database Errors

**Symptom**: `relation "table_name" does not exist`

**Solution**:
```bash
cd go/backend
make testenvdown
make testenvup
make generatetestmigrations
make test
```

### Docker Compose Services Not Starting

**Symptom**: Services fail to start or hang

**Solution**:
```bash
cd local
make down
make reset  # Nuclear option: removes all volumes
make hosts  # Ensure /etc/hosts is configured
make certs  # Regenerate TLS certificates
make all-nowatch
```

### Kratos Phone Number Validation Errors

**Symptom**: `invalid phone number format`

**Cause**: Kratos requires strict E.164 format

**Solution**:
- Use format: `+[country_code][number]` (e.g., `+49987654321`)
- No spaces, dashes, or parentheses
- If persistent, try `make reset` to clear Kratos state

### E2E Tests Timeout

**Symptom**: Tests exceed timeout limit

**Solution**:
- Run specific scenarios with tags instead of full suite
- Increase timeout: `go test -timeout 30m ...`
- Check if services are healthy: `docker compose ps`
- Review logs: `docker compose logs -f [service_name]`

### Temporal Workflow Stuck

**Symptom**: Workflow shows as "Running" but not progressing

**Solution**:
- Check worker logs: `docker compose logs temporal-worker`
- Verify worker is processing tasks
- Terminate stuck workflow: Temporal Web UI → Workflows → Terminate
- Restart worker: `docker compose restart temporal-worker`

### Go Module Cache Issues

**Symptom**: `go: cannot find module providing package`

**Solution**:
```bash
go clean -modcache
go mod tidy
go mod download
```

### Frontend Not Reflecting Backend Changes

**Symptom**: UI shows stale data or errors

**Solution**:
- Clear browser cache and cookies
- Restart frontend: `docker compose restart protea`
- Check API responses: Browser DevTools → Network tab
- Verify backend logs for errors

## Best Practices for AI Agents

### When Adding New Features

1. **Identify the domain** - Which domain does this feature belong to? (payments, wallets, kyc, etc.)
2. **Check existing patterns** - Look at similar features in the same domain
3. **Define models** - Add types to `types.go` or `types/` directory
4. **Implement ops** - Add business logic in `ops/` package with tests
5. **Add handlers** - Implement HTTP handlers with proper error handling
6. **Register routes** - Add to router in domain's `RegisterRoutes()`
7. **Write E2E tests** - Add feature file and step definitions
8. **Update docs** - If behavior is non-obvious, document it

### When Fixing Bugs

1. **Reproduce the bug** - Write a failing test that demonstrates the issue
2. **Identify root cause** - Use logs, debugger, or additional logging
3. **Fix minimally** - Change only what's necessary to fix the bug
4. **Verify fix** - Ensure test now passes
5. **Check for similar issues** - Search codebase for similar patterns
6. **Refactor if needed** - Clean up surrounding code if it contributed to the bug

### When Refactoring

1. **Ensure tests exist** - Don't refactor untested code without adding tests first
2. **Make small changes** - One refactor at a time, verify tests still pass
3. **Extract incrementally** - Move code to helpers gradually, not all at once
4. **Preserve behavior** - Tests should pass without modification (unless fixing bugs)
5. **Remove dead code** - Clean up unused functions, variables, imports
6. **Document non-obvious changes** - Add comments for complex logic

### When Reviewing Code

1. **Check test coverage** - Ensure new code has appropriate tests
2. **Verify error handling** - All errors should be wrapped with context
3. **Check for sensitive data** - No PII, passwords, tokens in logs
4. **Review database queries** - Use parameterized queries, avoid N+1 patterns
5. **Test E2E impact** - Consider if changes affect user-facing behavior
6. **Validate performance** - Large loops, database queries, external API calls

## Success Metrics

Your changes should maintain or improve:

- **Test Coverage**: ≥80% for backend packages
- **E2E Test Pass Rate**: 100% (all scenarios must pass)
- **Build Time**: Backend builds < 30 seconds
- **Test Runtime**: Unit tests < 5 minutes, E2E tests < 30 minutes
- **API Response Time**: P95 < 500ms for typical requests
- **Code Maintainability**: Reduced duplication, clear separation of concerns

## Questions? Need Help?

When encountering ambiguity:
1. Check existing implementation in similar domains
2. Refer to `AGENTS.md` in relevant directory (e.g., `e2e/AGENTS.md`)
3. Review logs for error context: `docker compose logs -f [service]`
4. Search codebase for similar patterns: `grep -r "pattern" go/backend/`
5. Default to simplest solution that maintains test coverage

Remember: This is a production-grade application. Prioritize correctness, testability, and maintainability over speed of implementation.

---

**Last Updated**: February 20, 2026  
**Maintainers**: Interledger Foundation  
**Repository**: https://gitlab.com/fynbos/interledger-wallet
