## GitHub scripts

JavaScript modules used by GitHub Actions workflows via `actions/github-script`.

### Scripts

| Script | Used by | Purpose |
| ------ | ------- | ------- |
| `select-packages.js` | `build-and-publish.yml` | Outputs the build matrix (Go + TypeScript packages × platforms) |
| `determine-version.js` | `build-and-publish.yml` | Computes the Docker image tag and whether to push, based on the trigger (tag push / dispatch / PR) |
| `e2e-runner-scaler.js` | `e2e-runner-scaler.yml` | Scales GCP VM runners for E2E test jobs |

### Local development

#### Prerequisites

- [NodeJS](https://nodejs.org) version 20 or later
- [PNPM](https://pnpm.io)

#### Environment setup

1. Install dependencies
```sh
pnpm install
```

---

#### Useful commands

| Description                    | Command                        |
| ------------------------------ | ------------------------------ |
| Test scripts                   | `node --test`                  |
