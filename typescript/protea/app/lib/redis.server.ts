import type { RedisClientType } from '@redis/client'
import { createClient } from '@redis/client'
import logger from './logger.server'

let redisClient: RedisClientType

const url = process.env.REDIS_URL || 'redis://0.0.0.0:6379'

declare global {
  var __redisClient: RedisClientType | undefined
}

// this is needed because in development we don't want to restart
// the server with every change, but we want to make sure we don't
// create a new connection to the Client with every change either.
if (process.env.NODE_ENV === 'production') {
  redisClient = createClient({ url })
  redisClient.connect()
  redisClient.on('error', (err) => {
    logger.error({ error: err.message }, 'Redis error')
  })
} else {
  if (!global.__redisClient) {
    global.__redisClient = createClient({
      url,
      socket: {
        reconnectStrategy: 5000
      }
    })
    global.__redisClient.connect()
    global.__redisClient.on('error', (err) => {
      logger.error({ error: err.message }, 'Redis error')
    })
  }
  redisClient = global.__redisClient
}

// Function to wait for Redis connection
const waitForRedisConnection = (timeout: number = 5000): Promise<void> => {
  return new Promise((resolve, reject) => {
    if (redisClient.isReady) {
      resolve()
      return
    }

    const timer = setTimeout(() => {
      reject(new Error('Timeout waiting for Redis connection'))
    }, timeout)

    const onConnect = () => {
      clearTimeout(timer)
      resolve()
    }

    const onError = (error: Error) => {
      clearTimeout(timer)
      reject(error)
    }

    redisClient.once('connect', onConnect)
    redisClient.once('error', onError)
  })
}

export { redisClient, waitForRedisConnection }
