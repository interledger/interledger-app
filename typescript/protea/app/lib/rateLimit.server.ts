import { redisClient } from './redis.server'

interface RateLimitOptions {
  limit: number
  ttlSeconds: number
}
function getRateLimitDefaults(): RateLimitOptions {
  const limit = Number(process.env.DEFAULT_RATE_LIMIT_REQUESTS) || 4
  const ttlSeconds = Number(process.env.DEFAULT_RATE_LIMIT_TIME) || 3600

  return { limit, ttlSeconds }
}

const DEFAULT_RATE_LIMIT = getRateLimitDefaults()

/**
 * Generic Redis-based rate limiter.
 * Only increments the counter if the callback succeeds.
 *
 * @param key Unique key per action/user
 * @param callback Async function to execute
 * @param options { limit, ttlSeconds }
 * @returns { result, error }
 */
export async function rateLimit<T>(
  key: string,
  callback: () => Promise<T>,
  options: RateLimitOptions = DEFAULT_RATE_LIMIT
): Promise<{ result?: T; error?: string }> {
  const { limit, ttlSeconds } = options

  // Read current count
  const current = await redisClient.get(key)
  const count = current ? parseInt(current) : 0

  if (count >= limit) {
    return { error: `Too many attempts. Please try again later.` }
  }

  try {
    const result = await callback()
    await redisClient.set(key, count + 1, { EX: ttlSeconds })

    return { result }
  } catch (err) {
    throw err
  }
}
