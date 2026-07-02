// Shared Express app for protea's custom server (see ../server.js).
//
// Loaded two ways so dev and prod run the exact same request pipeline:
//   - dev:  server.js runs this via vite's `ssrLoadModule`, so it's
//     transformed on every request and gets HMR like the rest of the app.
//   - prod: `react-router build` compiles this file to `build/server/index.js`
//     (see vite.config.ts `build.rollupOptions.input`), and server.js imports
//     that build output directly.
//
// Replaces `react-router-serve`, which unconditionally mounts
// `morgan("tiny")` and floods logs with one line per HTTP request
// (including Kubernetes liveness/readiness probes), ignoring `LOG_LEVEL`.
import { createRequestHandler } from '@react-router/express'
import express from 'express'
import logger from '~/lib/logger.server'

const PROBE_PATHS = new Set(['/healthz', '/live', '/ready'])

export const app = express()

app.disable('x-powered-by')

// Trust exactly one reverse-proxy hop (Traefik locally, ingress controller
// in k8s) so `req.ip`/`req.protocol`/`req.hostname` reflect the real client
// without letting direct clients spoof `X-Forwarded-*`.
app.set('trust proxy', 1)

// Access log: one compact line per request, message only (no bound
// req/res/headers), only emitted at debug/trace so it never fires in
// normal operation. Probe paths are always skipped since they're pure
// noise even when debugging. Deliberately not using pino-http here — it
// binds the full `req` object onto a per-request child logger before its
// completion log fires, so even minimal serializers still print a `req`/
// `res` sub-object on every line. A plain `res.on('finish', ...)` gives
// full control over the line's shape at the cost of not seeing raw
// socket-level errors (aborted connections before a response was sent),
// which pino-http's `err` callback would have caught.
if (logger.level === 'debug' || logger.level === 'trace') {
  app.use((req, res, next) => {
    if (!PROBE_PATHS.has(req.path)) {
      res.on('finish', () => {
        const level =
          res.statusCode >= 500 ? 'error' : res.statusCode >= 400 ? 'warn' : 'debug'
        logger[level](`${req.method} ${req.originalUrl} ${res.statusCode}`)
      })
    }
    next()
  })
}

app.use(
  createRequestHandler({
    build: () => import('virtual:react-router/server-build'),
    mode: process.env.NODE_ENV
  })
)
