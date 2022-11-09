import { createSessionStorage } from '@remix-run/node'
import { redisClient } from '~/lib/redis.server'
import { v4 } from 'uuid'

export const { getSession, commitSession, destroySession } =
  createSessionStorage({
    cookie: {
      name: 'user_settings',
      httpOnly: true,
      maxAge: 86400,
      path: '/',
      sameSite: 'lax'
    },
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
