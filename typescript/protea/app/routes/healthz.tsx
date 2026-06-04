import { getBackendHealth } from '~/lib/backend-health.server'
import logger from '~/lib/logger.server'
import { getRedisHealth } from '~/lib/redis.server'
import type { Route } from './+types/healthz'

// Returns the response headers for /healthz.
//
// Always sets `Cache-Control: no-store` so a CDN/proxy can never serve a
// stale health response, and conditionally exposes the build version via
// `interledger-app-version` so the Deploy workflow can poll for a new
// release going live. The version env is intentionally only set in
// production images (see typescript/protea/Dockerfile), so dev/local
// servers won't advertise a version — matching the "omitted in development
// modes" requirement.
function healthzHeaders(): HeadersInit {
  const headers: Record<string, string> = {
    'Cache-Control': 'no-store'
  }
  const version = process.env.INTERLEDGER_APP_VERSION
  if (version) {
    headers['interledger-app-version'] = version
  }
  return headers
}

export const loader = async ({ request }: Route.LoaderArgs) => {
  const [redisHealth, backendHealth] = await Promise.all([
    getRedisHealth(),
    getBackendHealth()
  ])

  if (!redisHealth.ok) {
    logger.error(
      {
        method: request.method,
        path: new URL(request.url).pathname,
        error: redisHealth.error
      },
      'Health check failed: Redis is unavailable'
    )

    return new Response('Redis unavailable', {
      status: 503,
      headers: healthzHeaders()
    })
  }

  if (!backendHealth.ok) {
    logger.error(
      {
        method: request.method,
        path: new URL(request.url).pathname,
        error: backendHealth.error
      },
      'Health check failed: Backend is unavailable'
    )

    return new Response('Backend unavailable', {
      status: 503,
      headers: healthzHeaders()
    })
  }

  return new Response('OK', { status: 200, headers: healthzHeaders() })
}
