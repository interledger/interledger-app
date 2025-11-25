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
export enum RateLimitKeys {
  RecoveryEmail = 'recovery.email',
  VerifyEmail = 'verify.email'
}

type RateLimitKeyType = `${RateLimitKeys}_${string}`

/**
 * Generic Redis-based rate limiter.
 * Only increments the counter if the callback succeeds.
 *
 * @param key Unique key per action/user
 * @param callback Async function to execute
 * @param options { limit, ttlSeconds }
 * @returns { result, error }
 */
export async function rateLimit(
  key: RateLimitKeyType,
  options: RateLimitOptions = DEFAULT_RATE_LIMIT
): Promise<string | undefined> {
  const { limit, ttlSeconds } = options

  try {
    const current = await redisClient.get(key)
    const count = current ? Number(current) : 0
    if (count >= limit) {
      return 'Too many attempts. Please try again later.'
    }
    await redisClient.set(key, count + 1, { EX: ttlSeconds })
  } catch (err) {
    console.error('Rate limit read failed:', err)
  }
}

export function getKey(
  rateLimitKeys: RateLimitKeys,
  id: string
): RateLimitKeyType {
  return `${rateLimitKeys}_${id}`
}
