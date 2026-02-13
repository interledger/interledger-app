# E2E Workflow Plan (PR)

## Goal
Create a GitHub Actions workflow that runs the Playwright/Godog E2E suite against the current PR using the local docker-compose stack defined in `local/`.

## Inputs I Reviewed
- `local/Makefile` for environment setup targets
- `local/e2e-playwright/README.md` for test and Playwright install steps
- `local/docker-compose.yaml` and component YAMLs for service profiles and images
- `local/scripts/README.md` + `local/scripts/Makefile` for Rafiki seeding tool
- Existing workflows for Go version and general CI patterns

## Proposed Workflow Outline
1. **Checkout PR code**
   - Use `actions/checkout@v4` with `fetch-depth: 0`.

2. **Authenticate to GHCR (MockGatehub is private)**
   - Add workflow permissions `packages: read`.
   - Run `docker login ghcr.io` with `GITHUB_TOKEN` before any `docker compose` pulls.

3. **Host setup (Linux runner)**
   - Run `sudo make hosts` from `local/` to add host entries.
   - Run `make certs` to generate TLS certs.
   - Trust the certs on Linux using `make trust-debian` (requires `libnss3-tools`) so browsers accept `https://interledger.test`.

4. **Install Go 1.25.x**
   - Use `actions/setup-go@v5` with `go-version: '1.25.1'` (aligns with repo workflows; satisfies “Go 25” requirement).

5. **Install Playwright system dependencies**
   - `sudo apt-get update` and install libraries listed in `local/e2e-playwright/README.md`.
   - `cd local/e2e-playwright` and run `go mod download`.
   - Run `go run github.com/playwright-community/playwright-go/cmd/playwright install`.

6. **Build the local facilitator (Rafiki seeding tool)**
   - `cd local/scripts` and `make build` (produces `./local-dev-tool`).

7. **Start infrastructure profile and wait for readiness**
   - `cd local` and run `docker compose --profile infrastructure up -d`.
   - Add explicit waits: check Postgres and Redis containers are healthy/running, and Traefik reachable (e.g., `curl -k https://traefik.test`).

8. **Start services profile and wait**
   - `docker compose --profile services up -d`.
   - Wait for Kratos, Temporal, and Rafiki containers to be healthy/running.

9. **Seed Rafiki assets**
   - `cd local/scripts` and run `./local-dev-tool rafiki --skip-ui`.
   - Ensure it points at local Rafiki GraphQL endpoint (environment should already be wired by `local/.env` defaults).

10. **Start application profile and wait**
   - `cd local` and run `docker compose --profile application up -d`.
   - Wait for backend + frontends + mockgatehub to be reachable.

11. **Verify app endpoint**
    - `curl -k https://interledger.test` until HTTP 200 (or 30x) to confirm the stack is serving.

12. **Run E2E tests**
   - `cd local/e2e-playwright`.
   - `go test -v -timeout 10m -args -tags @signuponly` (no debug suppression, per AGENTS guidance).

13. **Collect logs on failure**
    - On test failure, dump `docker compose logs` for core services to aid diagnosis.

## Notes / Risks
- Trusting the self-signed cert is important for Playwright; without it, browser TLS errors can fail tests. Use `make trust-debian` and install `libnss3-tools`.
- Some containers are pulled from GHCR; ensure the workflow can pull public images without auth.
- MockGatehub is private in GHCR; the workflow must log in to `ghcr.io` using `GITHUB_TOKEN` and have `packages: read` permissions.
- The Playwright runner should not suppress output (`-debug=false` must not be used).
- Expect multi-minute runtime; use timeouts accordingly.

## Decisions
- Run only `@signuponly` scenarios for PRs.
- Do not upload Playwright artifacts yet.
- Keep explicit profile-by-profile startup (no `make all-nowatch`).
