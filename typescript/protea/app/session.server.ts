import type { Session } from '@remix-run/node'
import { createCookie, createSessionStorage } from '@remix-run/node'
import { randomUUID } from 'crypto'
import { v4 } from 'uuid'
import { redisClient } from '~/lib/redis.server'

const EXPIRATION_DURATION_IN_SECONDS = 60 * 60 * 24 // a day
const COOKIE_SECRETS = JSON.parse(
  process.env.COOKIE_SECRETS || '["TODO:secrets"]'
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
      const id = v4()
      await redisClient.set(id, JSON.stringify(data), {
        PXAT: expires?.valueOf()
      })
      return id
    },
    async readData(id) {
      const data = await redisClient.get(id)
      if (data == null) return null
      return JSON.parse(data)
    },
    async updateData(id, data, expires) {
      /**
       * NOTE: Don't set the update only flag here, because we want to always set the data.
       * Remix doesn't know if the data has changed or not, so we need to always set it.
       */
      await redisClient.set(id, JSON.stringify(data), {
        PXAT: expires?.valueOf()
      })
    },
    async deleteData(id) {
      await redisClient.del(id)
    }
  })

export function getCSRFToken(request: Request, session: Session): string {
  let token = randomUUID()
  session.set('csrf-token', token)
  return token
}

export function validateCSRFToken(token: string, session: Session): Boolean {
  return session.get('csrf-token') === token
}
