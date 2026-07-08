import type { z } from 'zod'
import type { ProteaConfigSchema } from './config.schema.mjs'

export type ProteaConfig = z.infer<typeof ProteaConfigSchema>

declare global {
  // eslint-disable-next-line no-var
  var __proteaConfig: ProteaConfig | undefined
}

// server.js resolves the config (via configa, awaiting any {{ secret }}
// lookups) and assigns it to globalThis before the app module graph is
// ever imported — see server.js. A top-level `await parseConfig(...)` here
// would be simpler, but this file is reachable (via root.tsx's loader) from
// the client bundle's build graph too, and Vite's client build target
// predates top-level-await support; a real await here breaks that build.
if (!globalThis.__proteaConfig) {
  throw new Error(
    'protea config accessed before server.js initialized it — check server bootstrap ordering'
  )
}

export const config: ProteaConfig = globalThis.__proteaConfig
