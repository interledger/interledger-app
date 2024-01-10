import type { RedisClientType } from '@redis/client'
import { createClient } from '@redis/client'

let redisClient: RedisClientType

// const url = process.env.REDIS_URL || 'redis://redis/'
const url = 'redis://localhost:6379'


declare global {
  var __redisClient: RedisClientType | undefined
}

// this is needed because in development we don't want to restart
// the server with every change, but we want to make sure we don't
// create a new connection to the Client with every change either.
if (process.env.NODE_ENV === 'production') {
  redisClient = createClient({ url })
  redisClient.connect()
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
      console.error('Redis error:', err)
    })
  }
  redisClient = global.__redisClient
}

export { redisClient }
