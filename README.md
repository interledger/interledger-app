# Interledger Wallet

## Local Development Environment

Go to [local/README.md](local/README.md) for the local development setup.

## Releases

Releases are automated via [semantic-release](https://semantic-release.gitbook.io). Merging non-doc-only changes to `main` triggers the release workflow, which analyses commit messages and — if any releasable change is present — creates a git tag, a GitHub Release with generated notes, and kicks off a Docker image build. Documentation-only merges (`documentation/**`) are excluded from the release trigger.

Version bumps follow [Conventional Commits](https://www.conventionalcommits.org):

| Commit type | Version bump |
|---|---|
| `feat:` | minor |
| `fix:`, `perf:` | patch |
| `BREAKING CHANGE:` / `feat!:` / `fix!:` | major |
| `refactor:`, `chore:`, `docs:`, `test:`, `ci:`, `build:`, `style:`, `local:` | none |

Do not create `release/v*` branches or push version tags manually.

## Testing

### GoLang unittests

Unit tests require Postgres. Each test creates its own temporary database, so tests run in
parallel without interfering with each other or the main development database.

Use the dedicated unit-test Postgres service from the `local/` environment — it uses a
RAM-backed `tmpfs` data directory so it is fast and fully ephemeral.

```bash
# Start dedicated unit-test Postgres on localhost:55432
cd local && make unit-test-db-up && cd ..

# Install Atlas CLI (generates test migrations from the schema)
curl -sSf https://atlasgo.sh | sh

# Generate test migrations (run from go/backend).
# The `postgres` database is used as Atlas' scratch space — no separate DB setup needed.
cd go/backend
atlas migrate diff create_all \
    --dir "file://db/testmigrations" \
    --to  "file://db/schema.hcl" \
    --dev-url "postgres://postgres:password@127.0.0.1:55432/postgres?sslmode=disable"

# Run all backend tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total
```

The test harness defaults to `postgres://postgres:password@127.0.0.1:55432/%s?sslmode=disable`
(hardcoded in `db/migrate.go`), so **`DB_URL` does not need to be set** when using the local
unit-test container. Only set it if pointing at a different Postgres instance.

When done:
```bash
cd local && make unit-test-db-down
```

Notes:
- The project moved from Postgres 15 to 17 — flag any unexpected behaviour related to that
- After the steps above, tests can also be run directly from VS Code
