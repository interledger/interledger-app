# E2E Tests with Cucumber & Godog

## What is Cucumber?

Cucumber is a behavior-driven development (BDD) tool that lets you write tests in plain language using Gherkin syntax. Instead of writing test code directly, you describe what the application should do in `.feature` files. These files bridge the gap between stakeholders and developers by making tests readable to everyone.

## Important detail
In order to be able to run tests in parralel, it is critical to design all tests to be completely independent of each other. Otherwise running tests become unfeasible.

**How it works here:**

1. **Feature files** (`.feature`) contain human-readable test scenarios written in Gherkin
2. **Godog** is the Go implementation of Cucumber that parses these feature files
3. **Step definitions** (Go functions in files like `gatehub_signup.go`, `gatehub_payments.go`) implement the actual test logic
4. **Playwright** automates browser interactions to execute the test steps

The test runner reads feature files, matches each line to a corresponding step definition, and executes the test logic.

## Running Tests Locally

### Prerequisites

1. **Local environment setup** (first time only):
   ```bash
   cd local
   make hosts                  # Add entries to /etc/hosts (requires sudo)
   make certs                  # Generate self-signed certificates
   make trust                  # Trust certs on macOS (or see getting-started.md for Linux)
   ```
   See [getting-started.md](docs/gatehub/getting-started.md) for detailed environment setup.

2. **Playwright dependencies** – Install the browser binaries:
   ```bash
   sudo apt-get install -y libnss3 libnspr4 libatk1.0-0 libatk-bridge2.0-0 libcups2 libxkbcommon0 libatspi2.0-0 libxcomposite1 libxdamage1 libxext6 libxfixes3 libxrandr2 libgbm1 libpango-1.0-0 libcairo2 libasound2
   cd local/e2e-playwright
   go mod download
   go run github.com/playwright-community/playwright-go/cmd/playwright install
   ```

3. **All services running** – Start the full stack:
   ```bash
   cd local
   make all
   ```
   This starts Traefik, Postgres, Redis, Kratos, Temporal, Rafiki, and application services.

4. **Rafiki assets seeded** – Automatically happens with `make all`, but can be run manually:
   ```bash
   cd local/scripts
   go run rafiki-setup.go
   ```
   This creates USD, EUR, GBP, and other currency assets with initial liquidity. See [rafiki-seeding.md](docs/rafiki-seeding.md) for details.

### Running All Tests

```bash
cd local/e2e-playwright
go test -v -timeout 10m
```

Output will show:
- Each scenario passing/failing
- Step-by-step execution in `pretty` format
- Screenshots saved to `debug/` if tests fail

**Note**: Some tests (especially those with multiple KYC flows like `@p2p-payment`) require more than 5 minutes. Use `-timeout 10m` to allow sufficient time for all scenarios to complete.

### Running Specific Tests with Tags

Use the `-tags` flag to run only scenarios tagged with specific labels:

```bash
# Run only signup tests
go test -v -timeout 10m -args -tags @signuponly

# Run KYC verification tests
go test -v -timeout 10m -args -tags @kyc

# Run multiple tags (OR logic)
go test -v -timeout 10m -args -tags "@signuponly or @kyc"

# Run all except a tag (NOT logic)
go test -v -timeout 10m -args -tags "not @wip"
```

Available tags in feature files:
- `@signup` – User registration flow
- `@kyc` – KYC verification (identity verification)
- `@deposit` – Deposit transactions
- `@p2p` – Peer-to-peer payments
- `@wip` or `@skip` – Work in progress (skipped by default)
- `@phone-debug` – Phone number validation debugging

### Running with Debug Output

#### Standard Output (minimal)
```bash
cd local/e2e-playwright
go test -v -timeout 10m
```

#### Suppressing Debug Output

Debug output is enabled by default. To run tests without debug logs:
```bash
cd local/e2e-playwright
go test -v -timeout 10m -args -debug=false
```

This disables `debugPrintln()` output but still runs tests normally.


### Debugging Failures

When a test fails:

1. **Check screenshots** in `debug/` – Shows page state at failure
2. **Review logs** from failed steps
3. **Enable debug mode** for that specific test:
   ```bash
   go test -v -timeout 10m -args -tags @kyc
   ```

## Feature Files

Located in `features/`:

The idea is that we expand the features and scenarios as the project grows.

Each feature follows Gherkin syntax:
- `Feature:` – Overall test suite name
- `Background:` – Steps that run before each scenario
- `Scenario:` – Individual test case
- `Given` – Initial state
- `When` – Action taken
- `Then` – Expected result
- `@tag` – Metadata for filtering

## MockGatehub Integration

KYC and deposit tests interact with MockGatehub, a lightweight mock implementation of the Gatehub API. It is:
- Running in the local environment at `https://mockgatehub.interledger.test` (via Traefik HTTPS)
- Accessed by the backend at `http://mockgatehub:8080` (internal Docker network)
- Pre-configured with test users and balances for E2E testing

See [gatehub-mock-setup.md](docs/gatehub/gatehub-mock-setup.md) for architecture details.

## Troubleshooting

- **Playwright install fails** – Ensure `go mod download` completed, then try again
- **Tests timeout** – Increase `-timeout` flag (e.g., `-timeout 10m`)
- **Database conflicts** – Each test run generates random test identifiers to avoid clashes
- **Flaky wallet tests** – See [AGENTS.md](AGENTS.md) for known issues with wallet address submission
- **Phone number validation errors** – If tel format errors occur, try `make reset` in `local/` to clean the environment
- **Mixed content errors** – Verify MockGatehub is accessible at `https://mockgatehub.interledger.test` and that Traefik is running
- **Certificate trust issues** – Run `make trust` in `local/` (macOS) or see [getting-started.md](docs/gatehub/getting-started.md#certs) for Linux

