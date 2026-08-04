# Contributing to interledger-app <!-- omit in toc -->

Thank you for contributing to the Interledger Wallet :tada: Your contributions are essential to making this project better.

## Before you begin

- Check the [existing issues](https://github.com/interledger/interledger-app/issues) to see if what you want to work on is already tracked.
- For anything non-trivial, open an issue (or comment on an existing one) before starting work, so we can discuss the approach.

## Table of Contents <!-- omit in toc -->

- [Types of contributions](#types-of-contributions)
  - [:beetle: Issues](#beetle-issues)
  - [:hammer_and_wrench: Pull requests](#hammer_and_wrench-pull-requests)
  - [:books: Documentation](#books-documentation)
- [Working in this repository](#working-in-this-repository)
  - [Repository layout](#repository-layout)
  - [Local development environment](#local-development-environment)
  - [Code quality](#code-quality)
    - [Go](#go)
    - [TypeScript](#typescript)
    - [Commit messages](#commit-messages)
    - [CI](#ci)
  - [Labels](#labels)
- [Submitting pull requests](#submitting-pull-requests)
- [Review process](#review-process)
- [Releases](#releases)

## Types of contributions

### :beetle: Issues

We use GitHub issues to track public reported bugs and features these will be triaged and picked up by our team. If you've found something that needs fixing, search open issues first to see if someone else has already reported it. If it's something new, open an issue with:

- A clear and descriptive title.
- A detailed description of the problem, including steps to reproduce if applicable.
- Information about your environment (OS, browser, version) where relevant.
- Any relevant screenshots, logs, or error messages (with secrets redacted).

### :hammer_and_wrench: Pull requests

Feel free to fork and open a pull request for changes you'd like to contribute. A maintainer will review it as soon as possible.

### :books: Documentation

Project documentation lives under [`documentation/`](../documentation) and various `README.md` / `AGENTS.md` files scattered through the repo (e.g. [`local/README.md`](../local/README.md), [`e2e/README.md`](../e2e/README.md)). Improvements and clarifications there are always welcome.

## Working in this repository

### Repository layout

This is a monorepo combining a Go workspace and a pnpm/TypeScript workspace:

```
go/                 Go module used via the root go.work workspace
  backend/          Main wallet backend (gRPC, GraphQL, Temporal workers, provider integrations)
  pacioli/          Double-entry ledger service
  geo/              Geographic/country data service
  configa/          Configuration loading/validation library
  log/, tracing/    Shared logging and OpenTelemetry helpers
  mock/             Mock provider services used for local dev and CI
typescript/         React Router 7 frontends (each its own pnpm project)
  protea/           User-facing wallet frontend
  botanist/         Admin dashboard
e2e/                BDD end-to-end tests (Godog + Playwright)
local/              Docker Compose local development environment
proto/              Protocol buffer definitions
helm/               Helm chart for deployment
documentation/      Project documentation
```

### Local development environment

See [local/README.md](../local/README.md) for the full setup (TLS certs, `/etc/hosts`, Docker Compose services). In short:

```sh
cd local
cp example.env .env
make hosts    # add /etc/hosts entries (sudo)
make certs    # generate self-signed TLS certs
make trust    # trust the certs (macOS)
make all      # start infrastructure, services, and the app
```

Once the environment is up, seed Rafiki assets:

```sh
cd local/scripts
make build
./local-dev-tool rafiki --skip-ui --wait-for-ready 120
```

### Code quality

#### Go

[golangci-lint](https://golangci-lint.run/) is used for linting, configured at `go/.golangci.yml`.

```sh
cd go
make lint          # golangci-lint run ./...
```

Unit tests require Postgres. Use the dedicated ephemeral unit-test database rather than your normal local environment's:

```sh
cd local && make unit-test-db-up && cd ..

# Generate test migrations (from go/backend)
cd go/backend
atlas migrate diff create_all \
  --dir "file://db/testmigrations" \
  --to "file://db/schema.hcl" \
  --dev-url "postgres://postgres:password@127.0.0.1:55432/postgres?sslmode=disable"

go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total
```

CI enforces per-package coverage thresholds defined in [`go/coverage.thresholds`](../go/coverage.thresholds) — changes must meet or exceed them.

If you touch `go.mod`/`go.sum` in any workspace module, keep the workspace tidy:

```sh
go work sync
for m in go e2e local/scripts; do (cd "$m" && go mod tidy); done
go work sync
```

#### TypeScript

[ESLint](https://eslint.org/) and [Prettier](https://prettier.io/) are used, configured per-package (`typescript/protea`, `typescript/botanist`). From within a package:

```sh
pnpm lint          # eslint --fix
pnpm format        # prettier --write
pnpm typecheck
pnpm build          # also runs typecheck
```

`pnpm precommit` runs format, lint, and build together — worth running before opening a PR.

Dependencies are managed with **pnpm** (`packageManager` pinned in the root `package.json`); don't use `npm install`. `protea` and `botanist` each have their own lockfile — install from inside the package you're working on.

#### Commit messages

PR titles must follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) — this is enforced by CI (`pr-title-check.yml`). Allowed types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`, `local`.

The type of your PR title determines whether — and how — a release is cut when it's merged (see [Releases](#releases)), so pick it deliberately:

| Type                                                                 | Effect             |
| -------------------------------------------------------------------- | ------------------ |
| `feat`                                                               | minor version bump |
| `fix`, `perf`                                                        | patch version bump |
| `feat!`, `fix!`, or a `BREAKING CHANGE:` footer                      | major version bump |
| `refactor`, `chore`, `docs`, `test`, `ci`, `build`, `style`, `local` | no release         |

#### CI

GitHub Actions workflows live in [`.github/workflows`](workflows). The main ones you'll hit on a PR:

- `pr-title-check.yml` — validates the PR title.
- `go-tests.yml` / `go-test-template.yml` / `mock-tester.yml` — Go unit tests + coverage thresholds for `configa`, `backend`, `pacioli`, and the mock services.
- `linting.yml` — `golangci-lint` across all Go code, plus a check that `go.mod`/`go.sum`/`go.work.sum` are tidy.
- `typescript-checks.yml` — lint/typecheck/build for `protea` and `botanist`.
- `e2e-tests.yml` — full Playwright/Godog suite against a full local-style environment.
- `helm-tests.yml` — `helm unittest` + `kubeconform` for the `interledger-app` and `mock-services` charts (only when `helm/**` changes).

Documentation-only changes (`documentation/**`) skip tests, builds, and linting. Local-only changes (`local/**`) skip unit tests, builds, and linting, but E2E still runs since it exercises that environment.

### Labels

PRs are auto-labeled by [`.github/workflows/labeler.yml`](workflows/labeler.yml) based on the paths changed (config: [`.github/labeler.yml`](labeler.yml)) — e.g. `backend`, `pacioli`, `geo`, `protea`, `botanist`, provider-specific labels like `gatehub`/`xago`/`fiant/pti`, `e2e`, `local`, `helm`, `documentation`. Labels are only added, never removed, so feel free to add more manually if useful.

## Submitting pull requests

1. Fork the repository (or create a branch if you have write access).
2. Create a new branch from `main`.
3. Make your changes and commit them.
4. Run the relevant lint/test/build commands locally (see [Code quality](#code-quality)).
5. Open a pull request against `main` with a title following [Conventional Commits](#commit-messages).
6. Fill out the PR template checklist (tests updated, no sensitive data in logs, DevOps notified of release-impacting changes, docs updated).
7. If your PR closes an issue, reference it in the description using `Closes #123`.
8. Be patient and be prepared to address feedback.

## Review process

- A maintainer will review your PR for correctness, code quality, and adherence to these guidelines.
- Please respond to feedback promptly and push follow-up commits rather than force-pushing over review comments, unless asked to.
- Once approved and CI is green, a maintainer will merge it.

## Releases

Releases are fully automated via [semantic-release](https://semantic-release.gitbook.io) based on commit types on `main` — **do not push version tags manually or open `release/v*` branches**. See the [Releases section of the README](../README.md#releases) for the full process, including maintenance/backport branches.

Thank you for contributing!
