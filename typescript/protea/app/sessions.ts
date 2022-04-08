import { createCookieSessionStorage } from 'remix'

export const { getSession, commitSession, destroySession } =
  createCookieSessionStorage({
    cookie: {
      name: 'user_settings',
      httpOnly: true,
      maxAge: 86400,
      path: '/',
      sameSite: 'lax'
    }
  })
