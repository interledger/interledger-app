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
  VerifyPhone = 'verify.phone'
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
  options: RateLimitOptions = DEFAULT_RATE_LIMIT
): Promise<{ result?: T; error?: string }> {
  const { limit, ttlSeconds } = options

  let count = 0
  try {
    const current = await redisClient.get(key)
    count = current ? Number(current) : 0
  } catch (err) {
    console.error('Rate limit read failed:', err)
  }

  if (count >= limit) {
    return { error: 'Too many attempts. Please try again later.' }
  }

  let result: T
  try {
    result = await callback()
  } catch (err) {
    throw err
  }

  try {
    await redisClient.set(key, count + 1, { EX: ttlSeconds })
  } catch (err) {
    console.error('Rate limit increment failed:', err)
  }

  return { result }
}

export function getKey(rateLimitKeys: RateLimitKeys, id: string) {
  return `${rateLimitKeys}_${id}`
}
