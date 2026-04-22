import type { Route } from './+types/healthz'
import { getRedisHealth } from '~/lib/redis.server'
import logger from '~/lib/logger.server'

export const loader = async ({ request }: Route.LoaderArgs) => {
  const redisHealth = await getRedisHealth()

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

  return new Response('OK', { status: 200 })
}
