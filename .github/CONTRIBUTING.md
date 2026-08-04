# Working with interledger-app <!-- omit in toc -->

> [!IMPORTANT]
> **This repository is not actively maintained.**
>
> The Interledger Foundation is not focusing on this product at the moment, so we are not
> triaging issues, reviewing pull requests, or cutting releases. The code is public under the
> [Apache 2.0 licence](../LICENSE) so that others can learn from it and build on it.
>
> **You are welcome to fork it and take it wherever you like.** If you want to talk about what
> you are building, or find out where things stand, come and chat with the community on the
> [Interledger Slack](https://communityinviter.com/apps/interledger/interledger-working-groups-slack).

This document is no longer a contribution workflow — it is an orientation guide for anyone
working in their own fork. It explains how the codebase is laid out, how to get it running
locally, and how the tooling that ships with it behaves.

## Table of Contents <!-- omit in toc -->

- [What to expect](#what-to-expect)
- [Working in a fork](#working-in-a-fork)
  - [Repository layout](#repository-layout)
  - [Local development environment](#local-development-environment)
  - [Code quality](#code-quality)
    - [Go](#go)
    - [TypeScript](#typescript)
    - [Commit messages](#commit-messages)
    - [CI](#ci)
  - [Labels](#labels)
  - [Releases](#releases)
- [Getting in touch](#getting-in-touch)

## What to expect

To set expectations clearly:

- **Issues** — the tracker may be closed or left unattended. Please do not rely on a response.
  If you have found a bug, the fix belongs in your fork.
- **Pull requests** — we are not reviewing or merging them. If you open one it will most likely
  sit unanswered, which is a waste of your time; a fork is the better route.
- **Releases** — no further versions are planned. The automation described below is still in the
  repository and still works, but nobody is running it on this side.
- **Security** — since the project is unmaintained, treat the code as-is. Review it yourself
  before running anything derived from it in production, and note that pinned dependency
  versions will drift out of date and are not being patched.
- **Documentation** — `documentation/` and the various `README.md` / `AGENTS.md` files were
  accurate when written, but will age along with the code.

None of this stops you from using the project. Apache 2.0 lets you fork, modify, redistribute,
and run it commercially, provided you keep the licence and attribution intact (see [NOTICE](../NOTICE)).

## Working in a fork

Everything below describes the repository as it stands, so you can find your way around it.

### Repository layout

This is a monorepo combining a Go workspace and a set of TypeScript frontends:

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
helm/               Helm charts for deployment
documentation/      Project documentation
```

The wallet integrates with several third-party payment providers (GateHub, PTI/Fiant, Xago,
Chimoney, Plaid) and with a [Rafiki](https://github.com/interledger/rafiki) ILP node. Mock
implementations of each provider live under `go/mock/`, so the whole stack runs locally without
any real provider credentials — which is the easiest way to explore it.

### Local development environment

See [local/README.md](../local/README.md) for the full setup (TLS certs, `/etc/hosts`, Docker
Compose services). In short:

```sh
cd local
cp example.env .env
make hosts    # add /etc/hosts entries (sudo)
make certs    # generate self-signed TLS certs
make trust    # trust the certs (macOS; trust-debian / trust-arch on Linux)
make all      # start infrastructure, services, and the app
```

Once the environment is up, seed Rafiki assets:

```sh
cd local/scripts
make build
./local-dev-tool rafiki --skip-ui --wait-for-ready 120
```

### Code quality

The linting, testing, and formatting setup is intact and worth keeping if you fork.

#### Go

[golangci-lint](https://golangci-lint.run/) is used for linting, configured at `go/.golangci.yml`.

```sh
cd go
make lint          # golangci-lint run ./...
```

Unit tests require Postgres. Use the dedicated ephemeral unit-test database rather than your
normal local environment's:

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

CI enforces per-package coverage thresholds defined in [`go/coverage.thresholds`](../go/coverage.thresholds).

If you touch `go.mod`/`go.sum` in any workspace module, keep the workspace tidy:

```sh
go work sync
for m in go e2e local/scripts; do (cd "$m" && go mod tidy); done
go work sync
```

#### TypeScript

[ESLint](https://eslint.org/) and [Prettier](https://prettier.io/) are used, configured
per-package (`typescript/protea`, `typescript/botanist`). From within a package:

```sh
pnpm lint          # eslint --fix
pnpm format        # prettier --write
pnpm typecheck
pnpm build          # also runs typecheck
```

`pnpm precommit` runs format, lint, and build together.

Dependencies are managed with **pnpm** (`packageManager` pinned in the root `package.json`);
don't use `npm install`. `protea` and `botanist` each have their own lockfile — install from
inside the package you're working on.

#### Commit messages

The repository uses [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/),
enforced on PR titles by `pr-title-check.yml`. Allowed types: `feat`, `fix`, `docs`, `style`,
`refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`, `local`.

The commit type drives the version bump that semantic-release would apply (see
[Releases](#releases)):

| Type                                                                 | Effect             |
| -------------------------------------------------------------------- | ------------------ |
| `feat`                                                               | minor version bump |
| `fix`, `perf`                                                        | patch version bump |
| `feat!`, `fix!`, or a `BREAKING CHANGE:` footer                      | major version bump |
| `refactor`, `chore`, `docs`, `test`, `ci`, `build`, `style`, `local` | no release         |

#### CI

GitHub Actions workflows live in [`.github/workflows`](workflows):

- `pr-title-check.yml` — validates the PR title.
- `go-tests.yml` / `go-test-template.yml` / `mock-tester.yml` — Go unit tests + coverage
  thresholds for `configa`, `backend`, `pacioli`, and the mock services.
- `linting.yml` — `golangci-lint` across all Go code, plus a check that
  `go.mod`/`go.sum`/`go.work.sum` are tidy.
- `typescript-checks.yml` — lint/typecheck/build for `protea` and `botanist`.
- `e2e-tests.yml` — full Playwright/Godog suite against a full local-style environment.
- `helm-tests.yml` — `helm unittest` + `kubeconform` for the `interledger-app` and
  `mock-services` charts (only when `helm/**` changes).

Documentation-only changes (`documentation/**`) skip tests, builds, and linting. Local-only
changes (`local/**`) skip unit tests, builds, and linting, but E2E still runs since it exercises
that environment.

If you fork, note that several workflows depend on infrastructure that is not public: the E2E
suite needs a self-hosted `e2e-tester-dynamic` runner, and the build, publish, and deploy
workflows expect Interledger Foundation GCP Artifact Registry credentials, a release GitHub App,
and the companion `interledger-app-deploy` repository. Expect to disable or rewire those. The
lint, unit-test, and Helm-test workflows run on standard GitHub-hosted runners and should work
in a fork as-is.

### Labels

PRs are auto-labeled by [`.github/workflows/labeler.yml`](workflows/labeler.yml) based on the
paths changed (config: [`.github/labeler.yml`](labeler.yml)) — e.g. `backend`, `pacioli`, `geo`,
`protea`, `botanist`, provider-specific labels like `gatehub`/`xago`/`fiant/pti`, `e2e`, `local`,
`helm`, `documentation`. Labels are only added, never removed.

### Releases

Release automation is fully driven by [semantic-release](https://semantic-release.gitbook.io)
from commit types on `main`, rather than by manual version tags or `release/v*` branches. See the
[Releases section of the README](../README.md#releases) for how the pipeline is wired, including
maintenance/backport branches. As noted above, no further releases are planned here — the
description is there so you understand the mechanism before adapting or removing it.

## Getting in touch

The best place to reach the wider Interledger community — to discuss ideas, ask what happened to
this project, or share what you have built on top of it — is the
[Interledger Slack](https://communityinviter.com/apps/interledger/interledger-working-groups-slack).

For related, actively developed projects, see [Rafiki](https://github.com/interledger/rafiki)
(the ILP node this wallet builds on), the
[Open Payments](https://github.com/interledger/open-payments) specification, and
[interledger.org/developers](https://interledger.org/developers).

Interactions in community spaces are covered by the [Code of Conduct](code_of_conduct.md).

Thanks for your interest in the project.
