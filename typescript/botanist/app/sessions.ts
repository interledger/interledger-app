import { createCookieSessionStorage } from 'react-router'

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
