// Custom server for protea. Runs the identical request pipeline
// (server/app.ts) in both dev and prod so the two environments stay as
// close as possible, per the React Router node-custom-server pattern:
// https://github.com/remix-run/react-router-templates/tree/main/node-custom-server
//
// Replaces `react-router-serve` (prod) and `vite dev` (dev).
import { newInClusterSecretClient, parseConfig } from '@interledger/configa'
import compression from 'compression'
import express from 'express'
import { ProteaConfigSchema } from './app/config.schema.mjs'

const BUILD_PATH = './build/server/index.js'
const DEVELOPMENT = process.env.NODE_ENV === 'development'
const PORT = Number(process.env.PORT) || 3000

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
    throw new Error('CONFIG environment variable contains no valid file paths')
  }
  return files
}

// Resolved once, here, before the app module graph is imported below (both
// the dev ssrLoadModule path and the prod BUILD_PATH import). app/config.server.ts
// reads the result back off globalThis — see that file for why the load
// can't just be a top-level await there.
globalThis.__proteaConfig = await parseConfig(configFiles(), {
  secretClient: newInClusterSecretClient()
}).resolve(ProteaConfigSchema)

const app = express()

if (DEVELOPMENT) {
  const viteDevServer = await import('vite').then((vite) =>
    vite.createServer({
      server: { middlewareMode: true },
      appType: 'custom'
    })
  )
  app.use(viteDevServer.middlewares)
  app.use(async (req, res, next) => {
    try {
      const source = await viteDevServer.ssrLoadModule('./server/app.ts')
      return await source.app(req, res, next)
    } catch (error) {
      if (error instanceof Error) {
        viteDevServer.ssrFixStacktrace(error)
      }
      next(error)
    }
  })
} else {
  app.use(compression())
  app.use(
    '/assets',
    express.static('build/client/assets', { immutable: true, maxAge: '1y' })
  )
  app.use(express.static('build/client', { maxAge: '1h' }))
  app.use(await import(BUILD_PATH).then((mod) => mod.app))
}

const server = app.listen(PORT, () => {
  console.log(`protea listening on http://localhost:${PORT}`)
})

// Graceful shutdown: stop accepting new connections, let in-flight requests
// finish, then exit. Kubernetes sends SIGTERM on pod termination; without
// this, active requests are cut mid-flight during rolling deploys. Hard cap
// at 10s so a stuck connection can't block the pod forever.
const SHUTDOWN_TIMEOUT_MS = 10_000

function shutdown(signal) {
  console.log(`Received ${signal}, shutting down gracefully`)
  server.close((err) => {
    if (err) {
      console.error('Error during server close', err)
      process.exit(1)
    }
    process.exit(0)
  })
  setTimeout(() => {
    console.error(
      `Forcing shutdown after ${SHUTDOWN_TIMEOUT_MS}ms; in-flight requests were cut`
    )
    process.exit(1)
  }, SHUTDOWN_TIMEOUT_MS).unref()
}

for (const signal of ['SIGTERM', 'SIGINT']) {
  process.on(signal, () => shutdown(signal))
}
