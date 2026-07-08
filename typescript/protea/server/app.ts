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

// Compared with 'entry.server', this middleware logs the loader/action requests
if (logger.level === 'debug' || logger.level === 'trace') {
  app.use((req, res, next) => {
    if (!PROBE_PATHS.has(req.path)) {
      res.on('finish', () => {
        const level =
          res.statusCode >= 500
            ? 'error'
            : res.statusCode >= 400
              ? 'warn'
              : 'debug'
        logger[level](`${req.method} ${req.path} ${res.statusCode}`)
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
