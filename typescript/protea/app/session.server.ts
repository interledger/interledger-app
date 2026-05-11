import { createCookie, createSessionStorage } from 'react-router';
import { v4 } from 'uuid'
import { redisClient, waitForRedisConnection } from '~/lib/redis.server'
import { envValue } from './env.server';

const EXPIRATION_DURATION_IN_SECONDS = 60 * 60 * 24 // a day
const COOKIE_SECRETS = JSON.parse(
  envValue("COOKIE_SECRETS")
)

const cookie = createCookie('user_settings', {
  httpOnly: true,
  path: '/',
  secrets: COOKIE_SECRETS,
  sameSite: 'lax', //true,
  maxAge: EXPIRATION_DURATION_IN_SECONDS
})

export const { getSession, commitSession, destroySession } =
  createSessionStorage({
    cookie,
    async createData(data, expires) {
      await waitForRedisConnection()
      const id = v4()
      await redisClient.set(id, JSON.stringify(data), {
        PXAT: expires?.valueOf()
      })
      return id
    },
    async readData(id) {
      await waitForRedisConnection()
      const data = await redisClient.get(id)
      if (data == null) return null
      return JSON.parse(data)
    },
    async updateData(id, data, expires) {
      await waitForRedisConnection()
      /**
       * NOTE: Don't set the update only flag here, because we want to always set the data.
       * Remix doesn't know if the data has changed or not, so we need to always set it.
       */
      await redisClient.set(id, JSON.stringify(data), {
        PXAT: expires?.valueOf()
      })
    },
    async deleteData(id) {
      await waitForRedisConnection()
      await redisClient.del(id)
    }
  })
