// Resolves botanist's config.yaml (via configa, awaiting any {{ secret }}
// lookups) and re-exposes it as the flat env vars botanist's app code already
// reads (process.env.BACKEND_GRPC_URL / PAYMENT_POINTER_BASE), then execs the
// real dev/prod command as a child process.
//
// Unlike protea, botanist has no custom server.js entry point (it runs
// `react-router dev` / `react-router-serve` directly) and only reads two
// plain, non-secret config values — so there's no need for a
// globalThis/module-graph handoff. Usage: node bootstrap.mjs <command> [...args]
import { spawnSync } from 'node:child_process'
import { newInClusterSecretClient, parseConfig } from '@interledger/configa'
import { BotanistConfigSchema } from './app/config.schema.mjs'

function configFiles() {
  const raw = process.env.CONFIG
  if (!raw) {
    throw new Error(
      'CONFIG environment variable is required (comma-separated list of YAML config file paths)'
    )
  }
  const files = raw
    .split(',')
    .map((f) => f.trim())
    .filter(Boolean)
  if (files.length === 0) {
    throw new Error(
      'CONFIG environment variable contains no valid file paths'
    )
  }
  return files
}

const config = await parseConfig(configFiles(), {
  secretClient: newInClusterSecretClient()
}).resolve(BotanistConfigSchema)

const [command, ...args] = process.argv.slice(2)
const result = spawnSync(command, args, {
  stdio: 'inherit',
  env: {
    ...process.env,
    BACKEND_GRPC_URL: config.backend_grpc_url,
    PAYMENT_POINTER_BASE: config.payment_pointer_base
  }
})

if (result.error) {
  throw result.error
}
process.exit(result.status ?? 1)
