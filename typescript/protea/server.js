// Custom production server for protea.
//
// Replaces `react-router-serve`, which unconditionally mounts
// `morgan("tiny")` and floods the logs with one line per HTTP request
// (including Kubernetes liveness/readiness probes against /healthz and
// /live).
//
// Behaviour:
//   - Access logs are only emitted when LOG_LEVEL is `debug` or `trace`.
//   - Health/probe endpoints (/healthz, /live, /ready) are always skipped,
//     since they are pure noise even when debugging.
//
// Everything else mirrors @react-router/serve so deployment is unchanged.

import { createRequestHandler } from '@react-router/express'
import compression from 'compression'
import express from 'express'
import morgan from 'morgan'

const build = await import('./build/server/index.js')

const app = express()

app.disable('x-powered-by')
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
    morgan('tiny', {
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
app.listen(port, () => {
  // eslint-disable-next-line no-console
  console.log(`protea listening on http://localhost:${port}`)
})
