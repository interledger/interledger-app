# Interledger Wallet

## Local Development Environment

Go to [local/README.md](local/README.md) for the local development setup.

## Releases

### Standard releases

Releases are automated via [semantic-release](https://semantic-release.gitbook.io). Merging non-doc-only changes to `main` triggers the release workflow, which analyses commit messages and — if any releasable change is present — creates a git tag, a GitHub Release with generated notes, and kicks off a Docker image build and Helm chart publish. The new version is then automatically promoted to the `development` environment. Documentation-only merges (`documentation/**`) are excluded from the release trigger.

Version bumps follow [Conventional Commits](https://www.conventionalcommits.org):

| Commit type | Version bump |
|---|---|
| `feat:` | minor |
| `fix:`, `perf:` | patch |
| `BREAKING CHANGE:` / `feat!:` / `fix!:` | major |
| `refactor:`, `chore:`, `docs:`, `test:`, `ci:`, `build:`, `style:`, `local:` | none |

Do not push version tags manually.

### Promoting to sandbox or production

Use `.github/workflows/deploy.yml` (Actions → Deploy → Run workflow). Supply the target environment (`sandbox` or `production`) and optionally a version; leave version blank to deploy the latest published chart. The workflow opens an auto-merge PR in `interledger-app-deploy`, waits for the environment's healthz endpoint to confirm the rollout, and registers the release with Linear.

### Maintenance releases

Use a maintenance branch when production is on an older version and you need to backport a critical fix without shipping everything that has landed on `main` since.

**Branch naming:** `N.x` covers any patch in the `N.*` series; `N.M.x` is scoped to `N.M.*`. Both are recognised by semantic-release and by the release workflow.

**Create the branch from the tag you want to patch:**

```bash
# Example: patch the v1.13.0 line
git checkout -b 1.13.x v1.13.0
git push -u origin 1.13.x
```

**Apply the fix:**

```bash
# Cherry-pick from main, or commit directly to the branch
git cherry-pick <commit-sha>   # must use fix: / feat: prefix for a release to be cut
git push origin 1.13.x
```

Pushing a releasable commit triggers semantic-release on the branch. It creates `v1.13.1` (or the next patch in the series), publishes a GitHub Release, and builds the Docker image and Helm chart.

**Maintenance releases do not auto-deploy.** The automatic `development` promotion is skipped for maintenance branches — you decide exactly when and where the version lands. Deploy it manually once it is ready:

```
Actions → Deploy → Run workflow → environment: sandbox → version: 1.13.1
```

**Port the fix forward to `main`:**

After the maintenance release is cut, the same fix must be applied to `main` so it is not silently absent from future releases. Open a PR targeting `main` with the cherry-picked commit(s):

```bash
git checkout -b fix/backport-<description> origin/main
git cherry-pick <commit-sha>
# open a PR as normal
```

If the fix cannot be cleanly cherry-picked (e.g. the affected code has changed significantly), re-implement it on `main` directly. The goal is that `main` always contains every fix that has ever been shipped.

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
