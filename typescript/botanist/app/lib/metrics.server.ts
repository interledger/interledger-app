import { collectDefaultMetrics, Registry } from 'prom-client'

// A single process-wide Prometheus registry, exposed at GET /metrics by
// app/routes/metrics.ts. It is guarded on globalThis so that Vite/HMR module
// re-evaluation in dev does not create duplicate registries or register the
// default collectors twice (which throws).
declare global {
  // eslint-disable-next-line no-var
  var __metricsRegistry: Registry | undefined
}

function createRegistry(): Registry {
  const registry = new Registry()
  // Node.js runtime + process metrics only (event-loop lag, heap, GC, CPU,
  // open handles, ...). No application metrics are registered yet — this is
  // wiring only. Add custom metrics to this registry later.
  collectDefaultMetrics({ register: registry })
  return registry
}

export const register: Registry =
  globalThis.__metricsRegistry ??
  (globalThis.__metricsRegistry = createRegistry())
