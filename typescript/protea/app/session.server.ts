import { createCookie, createSessionStorage } from '@remix-run/node'
import { redisClient } from '~/lib/redis.server'
import { v4 } from 'uuid'

const EXPIRATION_DURATION_IN_SECONDS = 60 * 60 * 24 // a day

const cookie = createCookie('user_settings', {
  httpOnly: true,
  path: '/',
  secrets: ['TODO:secrets'],
  sameSite: 'lax', //true,
  expires: new Date(Date.now() + EXPIRATION_DURATION_IN_SECONDS * 1000)
})

export const { getSession, commitSession, destroySession } =
  createSessionStorage({
    cookie,
    async createData(data, expires) {
      const id = v4()
      await redisClient.set(id, JSON.stringify(data), {
        PXAT: expires?.valueOf(),
        NX: true
      })
      return id
    },
    async readData(id) {
      const data = await redisClient.get(id)
      if (data == null) return null
      return JSON.parse(data)
    },
    async updateData(id, data, expires) {
      await redisClient.set(id, JSON.stringify(data), {
        PXAT: expires?.valueOf(),
        XX: true
      })
    },
    async deleteData(id) {
      await redisClient.del(id)
    }
  })
