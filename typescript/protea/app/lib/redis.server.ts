import { createClient } from '@redis/client'
import logger from './logger.server'

type RedisClient = ReturnType<typeof createClient>

const REDIS_STARTUP_ATTEMPTS = 3
const REDIS_RETRY_DELAY_MS = 3000
const DEFAULT_WAIT_TIMEOUT_MS = 5000

const configuredRedisUrl = process.env.REDIS_URL?.trim()

if (!configuredRedisUrl) {
  logger.error('REDIS_URL config is empty. Exiting process.')
  process.exit(1)
}

interface RedisRuntimeState {
  client?: RedisClient
  startupConnectionPromise: Promise<void>
}

declare global {
  var __redisRuntimeState: RedisRuntimeState | undefined
}

let redisClient!: RedisClient

const getRedisTargetForLogs = (redisUrl: string) => {
  try {
    const parsed = new URL(redisUrl)
    return `${parsed.hostname}:${parsed.port || '6379'}`
  } catch {
    // Avoid leaking credentials from malformed URLs into logs.
    return '<invalid redis url>'
  }
}

logger.info(
  {
    nodeEnv: process.env.NODE_ENV,
    redisTarget: getRedisTargetForLogs(configuredRedisUrl),
    redisUrlFromEnv: Boolean(configuredRedisUrl)
  },
  'Initializing Redis client'
)

const attachRedisErrorLogger = (client: RedisClient) => {
  client.on('error', (err) => {
    logger.error({ error: err.message }, 'Redis error')
  })
}

const createRedisClientWithLogging = (): RedisClient => {
  const client = createClient({
    url: configuredRedisUrl,
    socket: {
      reconnectStrategy: 5000
    }
  })
  attachRedisErrorLogger(client)
  return client
}

const delay = (ms: number): Promise<void> => new Promise((resolve) => setTimeout(resolve, ms))

const connectWithRetry = async (): Promise<void> => {
  let lastError: unknown

  for (let attempt = 1; attempt <= REDIS_STARTUP_ATTEMPTS; attempt++) {
    redisClient = createRedisClientWithLogging()
    if (global.__redisRuntimeState) {
      global.__redisRuntimeState.client = redisClient
    }

    try {
      await redisClient.connect()
      logger.info({ attempt, redisTarget: getRedisTargetForLogs(configuredRedisUrl) }, 'Redis connected')
      return
    } catch (error) {
      lastError = error
      logger.error(
        {
          attempt,
          maxAttempts: REDIS_STARTUP_ATTEMPTS,
          redisTarget: getRedisTargetForLogs(configuredRedisUrl),
          error: error instanceof Error ? error.message : String(error)
        },
        'Failed to connect to Redis during startup'
      )

      try {
        await redisClient.disconnect()
      } catch {
        // Ignore cleanup errors and continue retry loop.
      }

      if (attempt < REDIS_STARTUP_ATTEMPTS) {
        await delay(REDIS_RETRY_DELAY_MS)
      }
    }
  }

  logger.error(
    {
      redisTarget: getRedisTargetForLogs(configuredRedisUrl),
      maxAttempts: REDIS_STARTUP_ATTEMPTS,
      error: lastError instanceof Error ? lastError.message : String(lastError)
    },
    'Unable to establish Redis connection at startup. Exiting process.'
  )
  process.exit(1)
}

let startupConnectionPromise: Promise<void>

if (global.__redisRuntimeState) {
  if (global.__redisRuntimeState.client) {
    redisClient = global.__redisRuntimeState.client
  }
  startupConnectionPromise = global.__redisRuntimeState.startupConnectionPromise
} else {
  startupConnectionPromise = connectWithRetry()
  global.__redisRuntimeState = {
    client: redisClient,
    startupConnectionPromise
  }
}

const waitForRedisConnection = async (timeout: number = DEFAULT_WAIT_TIMEOUT_MS): Promise<void> => {
  await Promise.race([
    startupConnectionPromise,
    new Promise<never>((_, reject) => {
      setTimeout(() => reject(new Error('Timeout waiting for Redis startup connection')), timeout)
    })
  ])

  if (redisClient.isReady) {
    return
  }

  await new Promise<void>((resolve, reject) => {
    const timer = setTimeout(() => {
      cleanup()
      reject(new Error('Timeout waiting for Redis connection'))
    }, timeout)

    const onReady = () => {
      cleanup()
      resolve()
    }

    const onError = (error: Error) => {
      cleanup()
      reject(error)
    }

    const cleanup = () => {
      clearTimeout(timer)
      redisClient.off('ready', onReady)
      redisClient.off('error', onError)
    }

    redisClient.once('ready', onReady)
    redisClient.once('error', onError)
  })
}

const getRedisHealth = async (): Promise<{ ok: true } | { ok: false; error: string }> => {
  try {
    await waitForRedisConnection(2000)
    const response = await redisClient.ping()

    if (response !== 'PONG') {
      return { ok: false, error: `Unexpected Redis ping response: ${response}` }
    }

    return { ok: true }
  } catch (error) {
    return {
      ok: false,
      error: error instanceof Error ? error.message : String(error)
    }
  }
}

export { redisClient, waitForRedisConnection, getRedisHealth }
