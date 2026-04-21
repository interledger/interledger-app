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
//   - Emit structured JSON access logs via pino, only when `LOG_LEVEL`
//     is `debug` or `trace`, and never for liveness/readiness probes.
//   - Respect reverse-proxy headers (`X-Forwarded-*`) for exactly one
//     upstream hop so `req.ip`, `req.protocol`, and `req.hostname`
//     reflect the real client without allowing header spoofing from
//     direct connections.
//   - Shut down gracefully on SIGTERM / SIGINT so in-flight requests
//     complete during Kubernetes rolling deploys.
//
// Logging follows the repo logging policy
// (documentation/docs/logging-reference.md): JSON, one object per line.

import { createRequestHandler } from '@react-router/express'
import compression from 'compression'
import express from 'express'
import pino from 'pino'
import pinoHttp from 'pino-http'

const VALID_LOG_LEVELS = ['fatal', 'error', 'warn', 'info', 'debug', 'trace']
const logLevel = (process.env.LOG_LEVEL || 'warn').toLowerCase()
if (!VALID_LOG_LEVELS.includes(logLevel)) {
  process.stderr.write(
    JSON.stringify({
      level: 'fatal',
      ts: Date.now() / 1000,
      caller: 'server.js',
      msg: 'Invalid LOG_LEVEL configuration',
      error: `LOG_LEVEL must be one of: ${VALID_LOG_LEVELS.join(', ')}`,
      providedValue: logLevel,
    }) + '\n',
  )
  process.exit(1)
}

const logger = pino({
  level: logLevel,
  timestamp: pino.stdTimeFunctions.isoTime,
  formatters: {
    level: (label) => ({ level: label }),
  },
  base: { pid: undefined, hostname: undefined },
})

const build = await import('./build/server/index.js')

const app = express()

app.disable('x-powered-by')

// Trust exactly one reverse-proxy hop in front of us (Traefik locally,
// ingress controller in k8s). Scoping to a single hop prevents direct
// clients from spoofing `X-Forwarded-*` headers if the service is ever
// reachable without the intended proxy.
app.set('trust proxy', 1)

app.use(compression())

// Static assets, identical to react-router-serve defaults.
app.use(
  '/assets',
  express.static('build/client/assets', { immutable: true, maxAge: '1y' }),
)
app.use(express.static('build/client', { maxAge: '1h' }))

// Access logs: structured JSON via pino-http, only active when the
// logger is at debug/trace. Probe paths are skipped so liveness /
// readiness traffic never appears in logs.
const PROBE_PATHS = new Set(['/healthz', '/live', '/ready'])
const accessLogsEnabled = logLevel === 'debug' || logLevel === 'trace'

if (accessLogsEnabled) {
  app.use(
    pinoHttp({
      logger,
      autoLogging: {
        ignore: (req) => PROBE_PATHS.has(req.url?.split('?')[0] ?? ''),
      },
      customLogLevel: (_req, res, err) => {
        if (err || res.statusCode >= 500) return 'error'
        if (res.statusCode >= 400) return 'warn'
        return 'debug'
      },
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
  logger.info({ port }, 'protea listening')
})

// Graceful shutdown: stop accepting new connections, let in-flight
// requests finish, then exit. Kubernetes sends SIGTERM on pod
// termination; without this, active requests are cut mid-flight during
// rolling deploys. Hard cap at 10s so a stuck connection can't block
// the pod forever.
const SHUTDOWN_TIMEOUT_MS = 10_000

function shutdown(signal) {
  logger.info({ signal }, 'received shutdown signal, draining connections')
  server.close((err) => {
    if (err) {
      logger.error({ err: err.message }, 'error during server close')
      process.exit(1)
    }
    process.exit(0)
  })
  setTimeout(() => {
    logger.error(
      { timeoutMs: SHUTDOWN_TIMEOUT_MS },
      'forcing shutdown; in-flight requests were cut',
    )
    process.exit(1)
  }, SHUTDOWN_TIMEOUT_MS).unref()
}

for (const signal of ['SIGTERM', 'SIGINT']) {
  process.on(signal, () => shutdown(signal))
}
