# MockGatehub Test Environment

This directory contains an isolated test environment for running MockGatehub integration tests.

## Structure

```
testenv/
├── docker-compose.yml              # Isolated compose environment
├── godog_test.go                   # BDD-style E2E tests with Godog
└── README.md                       # This file
```

## Quick Start

```bash
go test -tags e2e -v -count=1 ./testenv/ -run TestFeatures
```

The tests will:
1. Start MockGatehub and Redis in isolated containers (ports 25151, 26380)
2. Wait for services to be ready
3. Run all integration tests
4. Print detailed results with color-coded output
5. Clean up containers and volumes automatically

## What Gets Tested

The integration test suite validates:
- ✅ Service health and availability
- ✅ User creation and management
- ✅ **Wallet auto-creation** (GET /core/v1/users/{userId} creates wallet if none exists)
- ✅ **Wallet persistence** (subsequent calls return same wallet)
- ✅ Authentication token generation
- ✅ **Iframe token generation** (with user mapping for deposit flow)
- ✅ KYC workflow (auto-approval in sandbox)
- ✅ Additional wallet creation via POST
- ✅ Multi-currency balance queries (11 currencies)
- ✅ Exchange rate data
- ✅ Vault information
- ✅ **Dynamic deposits** (custom amount/currency from iframe)
- ✅ Transaction creation

## Configuration

The test environment uses:
- **Port 25151** for MockGatehub (avoiding conflicts with port 8080)
- **Port 26380** for Redis (avoiding conflicts with port 6379)
- **No Redis persistence** - data is cleared after each test run
- **Isolated network** - `mockgatehub-test` network

## Tests Included

The integration test suite covers all critical MockGatehub endpoints:

1. **Health check** - Service availability
2. **Create managed user** - User registration
3. **Get authorization token** - Authentication flow
4. **Start KYC** - Identity verification (auto-approved in sandbox)
5. **Get user KYC state** - Verification status check
6. **Create wallet** - XRPL wallet generation
7. **Get wallet balance** - Multi-currency balance retrieval (11 currencies)
8. **Get exchange rates** - Real-time rate data
9. **Get vault information** - Liquidity provider vault data
10. **Create transaction** - Transaction processing

All tests pass against the isolated test environment.

## Requirements

- **Go 1.24+** (for running tests)
- **Docker and Docker Compose** (for containers)
- **MockGatehub Docker image** built as `local-mockgatehub`

No additional tools needed - the Godog tests handle all HTTP requests and assertions.

## Building the Docker Image

Before running tests, ensure the MockGatehub image is built:

```bash
cd /home/stephan/interledger/testnet
docker build -f packages/mockgatehub/Dockerfile -t local-mockgatehub .
```

## Troubleshooting

**Services fail to start:**
- Check if ports 25151 and 26380 are available: `lsof -i :25151 -i :26380`
- Ensure Docker daemon is running: `docker ps`
- Verify the `local-mockgatehub` image exists: `docker images | grep mockgatehub`
- Rebuild if needed: `cd ../../.. && docker build -f packages/mockgatehub/Dockerfile -t local-mockgatehub .`

**Tests fail:**
- Check service logs: `docker compose logs mockgatehub`
- Manually test endpoints: `curl http://localhost:25151/health`
- Ensure previous cleanup ran: `docker compose down -v`
- Check for port conflicts with main `docker/local` environment

**Go issues:**
- Ensure Go 1.24+: `go version`
- If tests fail to start services, check Docker daemon and port availability

## Manual Usage

### Start Environment Only

```bash
# Start containers without running tests
docker compose up -d

# Check service health
curl http://localhost:25151/health

# View logs
docker compose logs -f mockgatehub

# Stop and clean up
docker compose down -v
```

### Modify Tests

Edit the BDD feature files (`features/*.feature`) and test scenario implementations in the testenv directory. The e2e tests use Godog for BDD-style testing with clear step definitions.

### Run Specific Tests

You can run specific tests by using Go test flags:

1. Run a specific test file: `go test ./testenv/godog_test.go`
2. Use verbose mode for detailed output: `go test -v ./...