import { register } from '~/lib/metrics.server'
import type { Route } from './+types/metrics'

// Prometheus scrape endpoint. This is a resource route (no default export), so
// it renders no UI — it returns the current metrics registry in Prometheus
// text exposition format. Scraped by the local Prometheus container (see
// local/monitoring.yaml) at protea:3000/metrics.
//
// `Cache-Control: no-store` so a proxy can never serve a stale scrape.
export const loader = async (_args: Route.LoaderArgs) => {
  const body = await register.metrics()
  return new Response(body, {
    status: 200,
    headers: {
      'Content-Type': register.contentType,
      'Cache-Control': 'no-store'
    }
  })
}
