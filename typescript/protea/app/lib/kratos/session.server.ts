import type { Session } from '@ory/client'
import { redirect, href } from 'react-router'
import { safeReturnTo } from '../url.server'
import { kratosPublic, KRATOS_SESSION_COOKIE } from './kratos-client.server'
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

function getCachedToSession(request: Request) {
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

// TODO: migrate the remaining cookie-only callers (`_index/route.tsx`,
// `me_.$.tsx`, `confirmations.tsx`, `recovery_.password.tsx`) to this
// function, then rename `hasUserSession` -> `hasSessionCookie`.
export async function isAuthenticated(request: Request): Promise<boolean> {
  if (!String(request.headers.get('cookie')).includes(KRATOS_SESSION_COOKIE)) {
    return false
  }
  try {
    await getCachedToSession(request)
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
        headers: { 'Clear-Site-Data': '"cookies"' }
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
    const { data } = await getCachedToSession(request)
    return data
  } catch (err: any) {
    const status = err.response?.status

    switch (status) {
      case WhoamiStatus.SESSION_INVALID:
        // TODO: emit a Sentry/log event here — a 401 with cookie present
        // is the canary for stale sessions in prod.
        throw redirect(loginUrl, {
          headers: { 'Clear-Site-Data': '"cookies"' }
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

// TODO: misleading name — rename to `hasSessionCookie` once `_index/route.tsx`,
// `me_.$.tsx`, `confirmations.tsx`, `recovery_.password.tsx` migrate to
// `isAuthenticated`.
export function hasUserSession(request: Request): boolean {
  return String(request.headers.get('cookie')).includes(KRATOS_SESSION_COOKIE)
}

export async function requireNoUserSession(request: Request): Promise<void> {
  if (!hasUserSession(request)) return

  const { pathname, search } = new URL(request.url)
  const target = `${pathname}${search}`

  try {
    await getCachedToSession(request)
    throw redirect(href('/'))
  } catch (err: any) {
    if (err instanceof Response) throw err

    const status = err.response?.status

    switch (status) {
      case WhoamiStatus.AAL_FORBIDDEN:
      case WhoamiStatus.AAL_UNPROCESSABLE:
        throw redirect(href('/totp/challenge'))

      case WhoamiStatus.SESSION_INVALID:
      default:
        throw redirect(target, {
          headers: { 'Clear-Site-Data': '"cookies"' }
        })
    }
  }
}
