# @interledger/configa

`@interledger/configa` is a small TypeScript library for loading YAML config into typed, validated objects, supporting:

- layered config files (base + overlays)
- optional Kubernetes secret template resolution
- validation via [Zod](https://zod.dev) schemas

It is a TypeScript port of [`go/configa`](../../../go/configa) — same merge rules, same `{{ secret "name" "key" }}` template syntax, same YAML files are readable by both the Go and TypeScript services. It's designed for apps that want simple file-based config in all environments, with Kubernetes Secret lookup only when needed.

## Features

- Parse one or more YAML files into a typed config object
- Deep-merge overlays (later files win)
- Resolve templates like `{{ secret "my-secret" "password" }}`
- Skip Kubernetes calls when no templates are present
- Validate the merged config against a Zod schema

## Quick Start (No Secrets)

### 1. Define a schema

```ts
import { z } from 'zod'

export const ConfigSchema = z.object({
  port: z.string().default('8080'),
  db_url: z.string().min(1),
  log_level: z.string().default('info')
})

export type Config = z.infer<typeof ConfigSchema>
```

### 2. Create YAML config

```yaml
# config.yaml
port: '8080'
db_url: postgres://localhost:5432/app
log_level: info
```

### 3. Parse and resolve

```ts
import { parseConfig } from '@interledger/configa'
import { ConfigSchema } from './config.schema'

async function loadConfig() {
  return parseConfig(['config.yaml']).resolve(ConfigSchema)
}
```

`resolve()` parses the merged YAML, substitutes any `{{ secret }}` templates (see below), and validates the result against the schema — throwing a descriptive error if validation fails.

## Overlay Files

Pass multiple files to `parseConfig`. Files are merged in order:

```ts
parseConfig(['config/base.yaml', 'config/dev.yaml'])
```

Merge behavior:

- Maps: deep-merged recursively
- Scalars and arrays: overlay value replaces base value entirely
- Last file wins for conflicts

A common pattern is to load the list from an env var, exactly like the Go backend does:

```ts
const files = (process.env.CONFIG ?? '')
  .split(',')
  .map((f) => f.trim())
  .filter(Boolean)

const config = await parseConfig(files, {
  secretClient: newInClusterSecretClient()
}).resolve(ConfigSchema)
```

## Kubernetes Secret Templates

If the YAML contains templates, pass a `secretClient` to `parseConfig`.

Template format:

```yaml
api_key: '{{ secret "wallet-secrets" "apiKey" }}'
database_url: 'postgres://user:{{ secret "wallet-secrets" "dbPassword" }}@db:5432/app'
```

Only the `secret` function is supported — this is a lightweight regex-based resolver, not a full template engine. Any other `{{ ... }}` expression fails validation with a clear error.

### In-cluster usage

```ts
import { newInClusterSecretClient, parseConfig } from '@interledger/configa'

const secretClient = newInClusterSecretClient()

const config = await parseConfig(['config.yaml'], {
  secretClient,
  // Optional: override the namespace used for secret lookups.
  // Defaults to the service account's own namespace.
  namespace: 'wallet'
}).resolve(ConfigSchema)
```

`newInClusterSecretClient()` reads the pod's service account token/CA/namespace from `/var/run/secrets/kubernetes.io/serviceaccount/*` and calls the Kubernetes Secrets API directly, over CA-pinned TLS. Construction never fails — credential errors only surface on the first actual secret fetch, so it's safe to construct outside a cluster.

Notes:

- If no `{{ ... }}` templates survive the merge, `resolve()` never touches Kubernetes — plaintext-only configs (e.g. local development) work with zero cluster access.
- If templates **are** present but no `secretClient` was provided, `resolve()` throws `ErrNoSecretClient`.
- Multiple references to the same secret name are fetched only once per `resolve()` call.
- A Kubernetes Secret key can only ever hold a string. If you need an array or object value from a secret, model the field as a string in your schema and use `z.preprocess` to `JSON.parse` it — see `cookie_secrets` in [`typescript/protea/app/config.schema.mjs`](../../protea/app/config.schema.mjs) for a real example.

## Local Testing with a Fake SecretClient

```ts
import { parseConfig, FakeSecretClient } from '@interledger/configa'

const secretClient = new FakeSecretClient({
  'my-secret': { apiKey: 'test-key', dbPassword: 'test-pass' }
})

const config = await parseConfig(['testdata/config.yaml'], {
  secretClient
}).resolve(ConfigSchema)
```

`FakeSecretClient` also exposes `callCount: Map<string, number>` if a test wants to assert that a given secret was fetched (and deduplicated) as expected.

You can also implement the `SecretClient` interface directly for more control:

```ts
import type { SecretClient } from '@interledger/configa'

const secretClient: SecretClient = {
  async getSecret(namespace, name) {
    return { apiKey: 'test-key' }
  }
}
```

## Error Handling

| Error | Thrown when |
|---|---|
| `ErrNoSecretClient` | The config contains `{{ secret }}` templates but no `secretClient` was provided. |
| `ErrSecretNotFound` | The Kubernetes API returned 404 for a referenced secret. |
| `ErrSecretForbidden` | The Kubernetes API returned 401/403 — the pod's service account lacks RBAC access (see `rbac.frontend.yaml` / `rbac.admin.yaml` / `rbac.backend.yaml` in the Helm chart). |
| `ErrSecretFetchFailed` | The Kubernetes API request failed for any other reason (network error, 5xx, etc). |

`parseConfig()` itself throws plain `Error`s for file-read failures, invalid YAML, and an empty file list. `resolve()` throws a plain `Error` (with the Zod error as `cause`) on schema validation failure, and for unsupported (non-`secret`) template expressions or missing secret keys.

## Real Usage in This Repo

Both `typescript/protea` and `typescript/botanist` load their config through this library, but wire the async load into their startup sequence differently because of how each app boots:

- **Protea** has a custom `server.js` entry point. It resolves config there — before the app's route module graph is imported — and hands the result to the rest of the app via `globalThis`. See [`typescript/protea/app/config.server.ts`](../../protea/app/config.server.ts) for why this uses a `globalThis` handoff instead of a top-level `await` (a top-level `await` in a file reachable from `root.tsx`'s loader breaks Vite's client build, since browser targets predate top-level-await support).
- **Botanist** has no custom server entry point — it runs `react-router dev` / `react-router-serve` directly. [`typescript/botanist/bootstrap.mjs`](../../botanist/bootstrap.mjs) resolves config and re-exposes the couple of values botanist needs as plain env vars, then execs the real command as a child process.

Both apps' config.yaml files live in `local/config/*.yaml` (checked in, plaintext test values) with git-ignored overrides in `local/config-override/*.yaml` for local development — see the [Frontend Configuration Guide](../../../documentation/docs/frontend-configuration-guide.md) for the full settings reference and the [Backend Configuration Guide](../../../documentation/docs/backend-configuration-guide.md) for the Go-side equivalent.

## Development

```bash
pnpm install    # from the typescript/ workspace root
pnpm --filter @interledger/configa build      # compile to dist/
pnpm --filter @interledger/configa test       # run the Vitest suite
pnpm --filter @interledger/configa typecheck
```
