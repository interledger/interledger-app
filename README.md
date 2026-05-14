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

Unit tests currently require the database to run. Each test will create a new database to run against on the given instance
allowing tests to run in parralel.

To avoid touching your main local development database, use the dedicated unit-test Postgres service from the `local/` environment.
This service uses a RAM-backed `tmpfs` data directory (ephemeral) to reduce disk I/O during tests.

```
# Start dedicated unit-test Postgres on localhost:55432
cd local
make unit-test-db-up
cd ..

# Install Atlas CLI on your local host
curl -sSf https://atlasgo.sh | sh

# Use Atlas to generate test migrations using a clean scratch database
export ATLAS_DEV_URL=postgres://postgres:password@127.0.0.1:55432/atlas_dev_tmp?sslmode=disable
psql "postgres://postgres:password@127.0.0.1:55432/postgres?sslmode=disable" -c "DROP DATABASE IF EXISTS atlas_dev_tmp;"
psql "postgres://postgres:password@127.0.0.1:55432/postgres?sslmode=disable" -c "CREATE DATABASE atlas_dev_tmp;"
atlas migrate diff create_all \
    --dir "file://go/backend/db/testmigrations" \
    --to  "file://go/backend/db/schema.hcl" \
    --dev-url "${ATLAS_DEV_URL}"

# Run tests against the dedicated unit-test DB instance
export DB_URL=postgres://postgres:password@127.0.0.1:55432/%s?sslmode=disable
```

You will now be able to run specific test files like this
```
go test -count=4 -v go/backend/kyc/ops/persona_test.go
```

When done:
```
cd local
make unit-test-db-down
```

Some notes
- In previous iterations the project used Postgres 15, so be on the lookout for issues relating to the move to Postgres 17
- After performing the steps above you should be able to run the tests directly from vscode

The end.