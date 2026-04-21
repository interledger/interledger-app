// Custom production server for protea.
//
// Replaces `react-router-serve`, which unconditionally mounts
// `morgan("tiny")` and floods the logs with one line per HTTP request
// (including Kubernetes liveness/readiness probes against /healthz and
// /live).
//
// Responsibilities:
//   - Serve static assets with the same cache headers as
//     `react-router-serve`.
//   - Delegate all other routes to React Router's express adapter.
//   - Emit HTTP access logs only when `LOG_LEVEL` is `debug` or `trace`,
//     and never for liveness/readiness probes.
//   - Respect reverse-proxy headers (`X-Forwarded-*`) so `req.ip`,
//     `req.protocol`, and `req.hostname` reflect the real client.
//   - Shut down gracefully on SIGTERM / SIGINT so in-flight requests
//     complete during Kubernetes rolling deploys.

import { createRequestHandler } from '@react-router/express'
import compression from 'compression'
import express from 'express'
import morgan from 'morgan'

const build = await import('./build/server/index.js')

const app = express()

app.disable('x-powered-by')

// Trust the reverse proxy in front of us (Traefik locally, ingress
// controller in k8s). Without this, Express sees the proxy's IP and
// always reports `req.protocol === 'http'` even when TLS terminated
// upstream.
app.set('trust proxy', true)

app.use(compression())

// Static assets, identical to react-router-serve defaults.
app.use(
  '/assets',
  express.static('build/client/assets', { immutable: true, maxAge: '1y' }),
)
app.use(express.static('build/client', { maxAge: '1h' }))

// Access logs: only when LOG_LEVEL is debug/trace, and never for probes.
const logLevel = (process.env.LOG_LEVEL || 'warn').toLowerCase()
const accessLogsEnabled = logLevel === 'debug' || logLevel === 'trace'

if (accessLogsEnabled) {
  const PROBE_PATHS = new Set(['/healthz', '/live', '/ready'])
  app.use(
    morgan('combined', {
      skip: (req) => PROBE_PATHS.has(req.path),
    }),
  )
}

app.all(
  '*',
  createRequestHandler({
    build,
    mode: process.env.NODE_ENV,
  }),
)

const port = Number(process.env.PORT) || 3000
const server = app.listen(port, () => {
  // eslint-disable-next-line no-console
  console.log(`protea listening on http://localhost:${port}`)
})

// Graceful shutdown: stop accepting new connections, let in-flight
// requests finish, then exit. Kubernetes sends SIGTERM on pod
// termination; without this, active requests are cut mid-flight during
// rolling deploys. Hard cap at 10s so we don't block the pod forever
// if something hangs.
const SHUTDOWN_TIMEOUT_MS = 10_000

function shutdown(signal) {
  // eslint-disable-next-line no-console
  console.log(`Received ${signal}, shutting down gracefully`)
  server.close((err) => {
    if (err) {
      // eslint-disable-next-line no-console
      console.error('Error during server close', err)
      process.exit(1)
    }
    process.exit(0)
  })
  setTimeout(() => {
    // eslint-disable-next-line no-console
    console.error(
      `Forcing shutdown after ${SHUTDOWN_TIMEOUT_MS}ms; in-flight requests were cut`,
    )
    process.exit(1)
  }, SHUTDOWN_TIMEOUT_MS).unref()
}

for (const signal of ['SIGTERM', 'SIGINT']) {
  process.on(signal, () => shutdown(signal))
}
