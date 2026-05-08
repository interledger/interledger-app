# GitHub Copilot Instructions for Interledger App

## Project Overview

**Purpose**: Interledger Wallet application enabling cross-currency peer-to-peer payments through integration with multiple payment providers (GateHub, PTI, Xago, Chimoney) and the Rafiki ILP node.

**Stack**: Go 1.25+, TypeScript (Remix), PostgreSQL 17, Redis, Ory Kratos, Temporal, Docker, Playwright (E2E tests)

**Repository Type**: Monorepo with Go workspace (`go.work`) containing:
- `go/` - Backend services (gRPC APIs, Temporal workflows, provider integrations)
- `typescript/` - Frontend applications (protea=user-facing, botanist=admin, hortus=public)
- `e2e/` - BDD end-to-end tests (Godog + Playwright)
- `local/` - Docker Compose development environment
- `proto/` - Protocol buffers definitions

**Size**: ~50k+ LoC across Go and TypeScript, 38+ documentation files, 11 CI/CD workflows

## Build and Validation

### Prerequisites
Always run these commands in sequence before any development work:

```bash
# 1. Ensure Go 1.25.1+ and Docker are installed
go version  # Must be >= 1.25.1
docker --version

# 2. Install Atlas CLI (database migrations)
curl -sSf https://atlasgo.sh | sh

# 3. Start Postgres for Go tests
docker run -d --name ilf-postgres -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=password -p 5432:5432 postgres:17
```

### Go Backend Testing

**CRITICAL**: Always generate test migrations before running tests:

```bash
cd go/backend  # or go/pacioli
export DB_URL=postgres://postgres:password@127.0.0.1:5432?sslmode=disable
atlas migrate diff create_all \
  --dir "file://db/testmigrations" \
  --to "file://db/schema.hcl" \
  --dev-url "${DB_URL}"
```

Then run tests:

```bash
# From go/backend or go/pacioli
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

**Coverage Thresholds** (enforced by CI, defined in `go/coverage.thresholds`):
- `backend=10.0%`
- `pacioli=34.8%`
- `geo=90.0%`

Changes must meet or exceed these thresholds.

### Linting

**Known Issue**: `golangci-lint` version mismatch may occur if the installed version was built with Go 1.24 but system Go is 1.25+. If `make lint` fails with "the Go language version used to build golangci-lint is lower than the targeted Go version", upgrade golangci-lint:

```bash
# Upgrade golangci-lint to version 2.5.0+
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.5.0
```

Then lint:

```bash
cd go
make lint  # Or: golangci-lint run ./...
```

### Local Development Environment

**Complete setup** (first time only):

```bash
cd local

# 1. Add DNS entries to /etc/hosts (requires sudo)
make hosts

# 2. Generate self-signed TLS certificates
make certs HEADLESS=1

# 3. Trust certificates (Linux Debian/Ubuntu)
sudo apt-get install libnss3-tools
make trust-debian

# For macOS, use: make trust
```

**Starting services**:

```bash
cd local

# All services (infrastructure + services + application)
make all

# Or step-by-step:
make infrastructure  # Traefik, Postgres, Redis
make services        # Kratos, Temporal, Rafiki
make application     # Wallet backends + frontends

# Stop all services
make down

# Reset (stop + delete volumes)
make reset
```

**After starting**, seed Rafiki assets:

```bash
cd local/scripts
make build
./local-dev-tool rafiki --skip-ui --wait-for-ready 120
```

### E2E Tests

**Prerequisites** (in addition to local environment setup):

```bash
# Install Playwright browsers
sudo apt-get install -y libnss3 libnspr4 libatk1.0-0 libatk-bridge2.0-0 \
  libcups2 libxkbcommon0 libatspi2.0-0 libxcomposite1 libxdamage1 libxext6 \
  libxfixes3 libxrandr2 libgbm1 libpango-1.0-0 libcairo2 libasound2

cd e2e
go mod download
go run github.com/playwright-community/playwright-go/cmd/playwright install chromium
```

**Running tests**:

```bash
cd e2e

# All tests (default timeout: 10m, may need 20m for full suite)
go test -v -timeout 20m

# Run specific tests by tag
go test -v -timeout 10m -args -tags @signuponly
go test -v -timeout 10m -args -tags "@withdrawal&&@fees"

# Without debug output
go test -v -timeout 10m -args -debug=false

# Parallel execution (CI mode)
go test -v -timeout 20m -args -concurrency=10 -debug=false
```

**Important**: E2E tests require the full local environment running (`make all` in `local/`). Screenshots of failures are saved to `e2e/debug/`.

## Project Structure and Architecture

### Go Workspace Layout

```
go/
├── backend/        # Main wallet backend (gRPC, GraphQL, Temporal workers)
│   ├── db/schema.hcl              # Atlas database schema
│   ├── db/testmigrations/         # Auto-generated test migrations
│   ├── grpc/                      # gRPC server definitions
│   ├── providers/{gatehub,pti,xago,chimoney}/  # Provider integrations
│   ├── temporal/workflows/        # Temporal workflow definitions
│   └── main.go
├── pacioli/        # Double-entry ledger service
├── geo/            # Geographic/country data service
└── Makefile        # Top-level Go commands (lint)
```

### TypeScript Layout

```
typescript/
├── protea/         # User-facing wallet frontend (Remix)
├── botanist/       # Admin dashboard (Remix)
└── hortus/         # Public-facing website (Remix)
```

### Key Configuration Files

- `go.work` - Go workspace (includes go/, e2e/, local/scripts/)
- `go/coverage.thresholds` - Coverage enforcement thresholds
- `go/.golangci.yml` - Linter configuration
- `local/docker-compose.yml` - Full stack orchestration (50+ services)
- `proto/buf.gen.yaml` - Protocol buffer generation config

### Critical Concepts (from `docs/concepts.md`)

**Wallet != Provider Wallet**: One Interledger wallet contains multiple "linked accounts" from different providers. Never confuse:
- Interledger `Wallet` (one per user, one ILP address, one country)
- GateHub "wallet" (XRPL address, multiple per user)
- PTI "wallet" (per-currency balance)
- Xago "SubAccount"

**Always use `linked_accounts.provider_id`** when calling provider APIs, not `wallets.id`.

**Transaction vs Payment**: A Payment is an intent (P2P, deposit, withdrawal), which creates one or more Transactions (ledger entries).

## CI/CD Validation Pipeline

### PR Checks (must pass before merge)

1. **PR Title** - Must follow Conventional Commits (feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert, local)
2. **Go Tests** - Coverage thresholds enforced for backend, pacioli, geo
3. **E2E Tests** - Full Playwright suite on Google Cloud VM (90m timeout)
4. **Linting** - golangci-lint v2.5.0 in `go/` directory

### Workflow Path Filtering

Workflows skip unnecessary CI runs based on which files changed:
- **Documentation-only changes** (`documentation/**`): All tests, builds, and linting are skipped.
- **Local-only changes** (`local/**`): Unit tests, builds, and linting are skipped. E2E tests still run since they exercise the local environment.
- These filters apply only to `pull_request` triggers — `push` to main and manual dispatches always run.

### PR Auto-Labeling

PRs are automatically labeled by `.github/workflows/labeler.yml` using the config in `.github/labeler.yml`. Labels are only added, never removed (`sync-labels: false`). If a label is already present, it is left as-is.

**Current label mappings** (keep `.github/labeler.yml` updated when the project evolves):

| Label | Paths |
|-------|-------|
| `backend` | `go/backend/**` |
| `pacioli` | `go/pacioli/**` |
| `geo` | `go/geo/**` |
| `mock services` | `go/mock/**` |
| `gatehub` | `go/backend/providers/gatehub/**`, `go/mock/mockgatehub/**` |
| `fiant/pti` | `go/backend/providers/pti/**`, `go/backend/providers/fiant/**`, `go/mock/mockpti/**` |
| `xago` | `go/backend/providers/xago/**`, `go/mock/mockxago/**` |
| `chimoney` | `go/backend/providers/chimoney/**`, `go/mock/mockchimoney/**` |
| `CI` | `.github/workflows/**`, `tooling/**` |
| `botanist` | `typescript/botanist/**` |
| `protea` | `typescript/protea/**` |
| `hortus` | `typescript/hortus/**` |
| `documentation` | `documentation/**` |
| `e2e` | `e2e/**` |
| `local` | `local/**` |

**When adding new providers, mock services, or frontend apps**: Update `.github/labeler.yml` with appropriate path globs and add the label mapping to this table.

### Workflows

- `.github/workflows/pr-title-check.yml` - Validates PR title format
- `.github/workflows/labeler.yml` - Auto-labels PRs based on changed files (config: `.github/labeler.yml`)
- `.github/workflows/go-tests.yml` - Runs `go-test-template.yml` for backend, pacioli (skipped for docs/local-only changes)
- `.github/workflows/e2e-tests.yml` - Starts VM, runs E2E suite with concurrency=10 (skipped for docs-only changes)
- `.github/workflows/linting.yml` - Runs golangci-lint on all Go code (skipped for docs/local-only changes)
- `.github/workflows/build-and-publish.yml` - Builds Docker images on PRs (build only) and pushes to GCP Artifact Registry when triggered by a version tag or `workflow_dispatch`
- `.github/workflows/release.yml` - Runs semantic-release on every push to `main`; creates a git tag, GitHub Release, and release notes from commit history (see Release Process below)

### Release Process

Releases are fully automated via **semantic-release** — do not create `release/v*` branches or manually push version tags.

**How it works:**

1. A PR is merged to `main`.
2. `release.yml` runs semantic-release, which analyses all commits since the last tag.
3. The version bump is determined from commit types:
   - `feat:` → minor (`1.1.0`)
   - `fix:`, `perf:` → patch (`1.0.1`)
   - `BREAKING CHANGE:` footer or `feat!:` / `fix!:` → major (`2.0.0`)
   - `refactor:`, `chore:`, `docs:`, `test:`, `ci:`, `build:`, `style:`, `local:` → **no release**
4. If there is a releasable commit, semantic-release creates a `vX.Y.Z` git tag and a GitHub Release with auto-generated notes.
5. The new tag triggers `build-and-publish.yml`, which builds all Docker images and pushes them to GCP Artifact Registry tagged with that version.

**Config files**: `.releaserc.json` (release config), `package.json` + `pnpm-lock.yaml` at repo root (semantic-release dependencies).

**Authentication**: `release.yml` authenticates as a GitHub App rather than using the default `GITHUB_TOKEN`. This is required because tag pushes made with `GITHUB_TOKEN` do **not** trigger downstream workflows, so `build-and-publish.yml` would never see the new tag. Required repo secrets:

- `RELEASE_APP_ID` — the GitHub App's numeric ID
- `RELEASE_APP_PRIVATE_KEY` — the App's PEM private key (not the OAuth client secret)

The App must be installed on this repository with **Contents: write** permission. The installation ID is auto-discovered at runtime by `actions/create-github-app-token`. `release.yml` validates the token, app installation, and permissions before invoking semantic-release, so misconfiguration fails fast with an explicit error.

**If `main` is protected**: ensure the App (its bot user, e.g. `your-app[bot]`) is listed as an allowed actor that can bypass branch protection for tag creation, or that the protection rules permit tag pushes from Apps.

### Testing Locally Before Push

```bash
# 1. Lint
cd go && make lint

# 2. Go tests with coverage
cd go/backend
atlas migrate diff create_all --dir "file://db/testmigrations" \
  --to "file://db/schema.hcl" \
  --dev-url "postgres://postgres:password@127.0.0.1:5432?sslmode=disable"
go test -coverprofile=coverage.out ./...

# 3. Check coverage thresholds
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
echo "Coverage: $COVERAGE (threshold: see go/coverage.thresholds)"

# 4. E2E tests (requires local environment)
cd ../../local && make all
# Wait ~30s for services to stabilize
cd ../local/scripts && ./local-dev-tool rafiki --skip-ui --wait-for-ready 120
cd ../../e2e && go test -v -timeout 20m
```

## Common Pitfalls and Workarounds

### Issue: Atlas Test Migrations Fail

**Symptom**: `go test` fails with "migration not found" or schema errors

**Fix**: Always regenerate test migrations before running tests:

```bash
export DB_URL=postgres://postgres:password@127.0.0.1:5432?sslmode=disable
atlas migrate diff create_all \
  --dir "file://db/testmigrations" \
  --to "file://db/schema.hcl" \
  --dev-url "${DB_URL}"
```

### Issue: golangci-lint Version Mismatch

**Symptom**: `make lint` fails with "Go language version used to build golangci-lint is lower than targeted Go version"

**Fix**: Upgrade golangci-lint to v2.5.0+ (built with Go 1.25+)

### Issue: E2E Tests Hang on WaitForLoadState

**Symptom**: Playwright tests timeout indefinitely

**Fix**: Always pass timeout when using `WaitForLoadState(networkidle)`:

```go
page.WaitForLoadState(playwright.LoadStateNetworkidle, 
  playwright.PageWaitForLoadStateOptions{Timeout: playwright.Float(10000)})
```

### Issue: Kratos Phone Number Validation Fails

**Symptom**: Valid phone numbers rejected with `format: "tel"` error

**Fix**: Clean environment with `cd local && make reset`, then restart. Root cause unknown but observed after environment changes.

### Issue: Local Environment Not Ready

**Symptom**: E2E tests fail with connection errors

**Fix**: Always wait for Rafiki seeding to complete:

```bash
cd local/scripts
./local-dev-tool rafiki --skip-ui --wait-for-ready 120
```

Then wait additional 10-30 seconds for all services to stabilize.

## Logging Policy

Follow `docs/logging-policy.md`:

- **fatal** → stderr, exit process (invalid config, security breach)
- **error** → stderr, immediate attention needed (provider API failures after retries)
- **warning** → stdout, periodic review needed (timeouts with retry, non-critical issues)
- **info** → stdout, noteworthy events (payment initiated, report generated)
- **debug** → stdout, troubleshooting only (strip before merge to main unless needed in deployed environments)

All logs must be JSON, one object per line. Include `ts` (Unix timestamp), optional `caller`, `requestId`, `correlationId`.

## Additional Resources

- `e2e/AGENTS.md` - Agent-specific E2E testing guidance (tag usage, troubleshooting)
- `docs/concepts.md` - Provider terminology mapping
- `documentation/docs/env-variables.md` - Full environment variable reference for Protea, Botanist, and Backend
- `local/README.md` - Detailed local environment documentation
- `e2e/README.md` - E2E test setup and execution guide
- `e2e/STEP_REFERENCE.md` - Gherkin step definitions reference

## Environment Variables

All environment variables for Protea (frontend), Botanist (admin), and the Go Backend are documented in `documentation/docs/env-variables.md`. This file includes:
- Whether each variable is a secret or not
- Suggested values per environment (production, sandbox, development, local)
- Notes on GateHub, Xago, PTI, Chimoney provider-specific config

**When reviewing PRs that touch environment configuration, always check:**

1. **New env vars added to `go/backend` or TypeScript apps** — add the variable to `documentation/docs/env-variables.md` with correct secret classification and per-environment values.
2. **Changes to `local/*.yaml` compose files or `local/example.env`** — keep `documentation/docs/env-variables.md` in sync.
3. **Changes to `interledger-app-deploy` values files** — update the Production/Sandbox/Development columns in `documentation/docs/env-variables.md`.
4. **Secrets must never be committed** — any new sensitive variable must reference a 1Password vault item; flag any plaintext secrets immediately.
5. **Production vs sandbox Gatehub client IDs differ** — production uses different OAuth client IDs. See the GateHub section of `documentation/docs/env-variables.md`.

## Trust These Instructions

These instructions were validated by running commands and reading the codebase comprehensively. Only perform additional searches if you encounter information that contradicts these instructions or if you need details not covered here.
