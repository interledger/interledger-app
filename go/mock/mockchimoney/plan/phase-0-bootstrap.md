# Phase 0 — Bootstrap: Repository Skeleton and Health Check

**Goal**: A runnable service that can be built, started, and health-checked. Nothing else.

## Deliverables

- `go.mod` (module `gitlab.com/fynbos/mock/mockchimoney`)
- `cmd/mockchimoney/main.go` — wires config, logger, router, HTTP server
- `internal/config/config.go` — reads all env vars from the configuration table with typed defaults
- `internal/logger/logger.go` — thin zap wrapper (JSON output, level from config)
- `internal/handler/handler.go` — `Handler` struct; request-logger middleware; `GET /health` returning `{"status":"ok"}`
- `Makefile` — targets: `build`, `run`, `test` (runs linting + unit tests + e2e tests), `lint`, `docker-build`
- `Dockerfile` — multi-stage, Go binary only
- `.gitignore`, `.golangci.yml`

## Feature File

`features/service_health.feature`

## Test-Driven Development Notes

1. **Red**: Run `go test -v ./features` or the equivalent godog command. Expect compilation errors and/or 404s.
2. **Green**: Wire the health route into the chi router. Ensure it returns 200 with the expected JSON.
3. **Refactor**: Extract middleware into its own file. Ensure config validation (e.g., panic on invalid `LOG_LEVEL`) is testable.

## Acceptance Criteria

- [ ] `make build` produces a working binary
- [ ] `make run` starts the service and listens on the configured port
- [ ] `GET /health` returns `{"status":"ok"}` with status 200
- [ ] Health check does not require authentication (test in Phase 2)
- [ ] All code passes `golangci-lint run ./...`
- [ ] Feature scenarios in `service_health.feature` all pass- [ ] `make test` runs successfully (linting + unit tests + e2e tests)