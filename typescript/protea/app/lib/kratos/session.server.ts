import type { Session } from '@ory/client'
import { redirect, href } from 'react-router'
import { safeReturnTo } from '../url.server'
import { kratosPublic, KRATOS_SESSION_COOKIE, CLEAR_SESSION_COOKIE_HEADER } from './kratos-client.server'
import { getCookie } from './cookie.server'
import type { KratosTraits } from './types.server'

const sessionPromiseCache = new WeakMap<
  Request,
  ReturnType<typeof kratosPublic.toSession>
>()

const WhoamiStatus = {
  SESSION_INVALID: 401,
  AAL_FORBIDDEN: 403,
  AAL_UNPROCESSABLE: 422,
} as const

function getCachedWhoamiSession(request: Request) {
  let cached = sessionPromiseCache.get(request)
  if (!cached) {
    cached = kratosPublic.toSession({ cookie: getCookie(request) })
    sessionPromiseCache.set(request, cached)
  }
  return cached
}

function buildLoginRedirectUrl(request: Request): string {
  const { pathname, search } = new URL(request.url)
  const target = `${pathname}${search}`
  const returnTo = safeReturnTo(target)
  return `${href('/login')}?${new URLSearchParams({ returnTo })}`
}

// TODO: migrate the remaining cookie-only callers routes to this
// function, then rename `hasUserSession` -> `hasSessionCookie`.
export async function isAuthenticated(request: Request): Promise<boolean> {
  if (!String(request.headers.get('cookie')).includes(KRATOS_SESSION_COOKIE)) {
    return false
  }

  try {
    await getCachedWhoamiSession(request)
    return true
  } catch (err: any) {
    const status = err.response?.status
    if (
      status === WhoamiStatus.AAL_FORBIDDEN ||
      status === WhoamiStatus.AAL_UNPROCESSABLE
    ) {
      return true
    }
    if (status === WhoamiStatus.SESSION_INVALID) {
      throw redirect(buildLoginRedirectUrl(request), {
        headers: { 'Set-Cookie': CLEAR_SESSION_COOKIE_HEADER }
      })
    }
    return false
  }
}

export async function getUserSession(
  request: Request,
  allowAal1 = false
): Promise<Session | null> {
  const loginUrl = buildLoginRedirectUrl(request)

  try {
    const { data } = await getCachedWhoamiSession(request)
    return data
  } catch (err: any) {
    const status = err.response?.status

    switch (status) {
      case WhoamiStatus.SESSION_INVALID:
        // TODO: possibly emit a sentry warning when this fires.
        // this path means a kratos cookie was present but rejected
        throw redirect(loginUrl, {
          headers: { 'Set-Cookie': CLEAR_SESSION_COOKIE_HEADER }
        })
      case WhoamiStatus.AAL_FORBIDDEN:
      case WhoamiStatus.AAL_UNPROCESSABLE:
        if (!allowAal1) {
          throw redirect(loginUrl)
        }
        return null
      default:
        throw redirect(loginUrl)
    }
  }
}

export function getSessionTraits(session: Session | null): KratosTraits {
  if (!session?.identity) {
    throw redirect(href('/login'))
  }
  return session.identity.traits as KratosTraits
}

// TODO: misleading name — rename to `hasSessionCookie`
// once routes using this migrate to `isAuthenticated`
export function hasUserSession(request: Request): boolean {
  return String(request.headers.get('cookie')).includes(KRATOS_SESSION_COOKIE)
}

export async function requireNoUserSession(request: Request): Promise<void> {
  if (!hasUserSession(request)) return

  const { pathname, search } = new URL(request.url)
  const target = `${pathname}${search}`

  try {
    await getCachedWhoamiSession(request)
    throw redirect(href('/'))
  } catch (err: any) {
    if (err instanceof Response) throw err

    const status = err.response?.status

    switch (status) {
      case WhoamiStatus.AAL_FORBIDDEN:
      case WhoamiStatus.AAL_UNPROCESSABLE:
        throw redirect(href('/totp/challenge'))

      case WhoamiStatus.SESSION_INVALID:
        throw redirect(target, {
          headers: { 'Set-Cookie': CLEAR_SESSION_COOKIE_HEADER }
        })
      default:
        throw redirect(target)
    }
  }
}
