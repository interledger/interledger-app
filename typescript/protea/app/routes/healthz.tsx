import type { Route } from './+types/healthz'
import { getBackendHealth } from '~/lib/backend-health.server'
import { getRedisHealth } from '~/lib/redis.server'
import logger from '~/lib/logger.server'

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

    return new Response('Redis unavailable', { status: 503 })
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

    return new Response('Backend unavailable', { status: 503 })
  }

  return new Response('OK', { status: 200 })
}
