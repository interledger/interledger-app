import { createCookie, createSessionStorage } from '@remix-run/node'
import { redisClient } from '~/lib/redis.server'
import { v4 } from 'uuid'

const EXPIRATION_DURATION_IN_SECONDS = 10 //60 * 60 * 24 // a day

const cookie = createCookie('user_settings', {
  httpOnly: true,
  path: '/',
  secrets: ['TODO:secrets'],
  sameSite: 'lax', //true,
  maxAge: EXPIRATION_DURATION_IN_SECONDS
  // expires: new Date(Date.now() + EXPIRATION_DURATION_IN_SECONDS * 1000)
})

export const { getSession, commitSession, destroySession } =
  createSessionStorage({
    cookie,
    async createData(data, expires) {
      console.log('CREATE', data, expires)
      const id = v4()
      await redisClient.set(id, JSON.stringify(data), {
        PXAT: expires?.valueOf(),
        NX: true
      })
      return id
    },
    async readData(id) {
      const data = await redisClient.get(id)
      console.log('READ', data)
      if (data == null) return null
      return JSON.parse(data)
    },
    async updateData(id, data, expires) {
      console.log('UPDATE', data, expires)
      await redisClient.set(id, JSON.stringify(data), {
        PXAT: expires?.valueOf(),
        XX: true
      })
    },
    async deleteData(id) {
      console.log('DELETE')
      await redisClient.del(id)
    }
  })
