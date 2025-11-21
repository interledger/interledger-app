import { redisClient } from './redis.server'

interface RateLimitOptions {
  limit: number
  ttlSeconds: number
}

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
  options: RateLimitOptions = { limit: 4, ttlSeconds: 60 * 60 }
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
