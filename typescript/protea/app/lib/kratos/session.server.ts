import type { Session } from '@ory/client'
import { redirect } from '@remix-run/node'
import { route } from 'routes-gen'
import { safeReturnTo } from '../url.server'
import { kratosPublic, KRATOS_SESSION_COOKIE } from './kratos-client.server'
import { getCookie } from './cookie.server'
import type { KratosTraits } from './types.server'


/**
 * getUserSession allows fetching a user's kratos session.
 * @param request Request received in a loader function.
 * @returns Promise<Session | null> - the user's Kratos session, or null if no valid session is available and AAL1 is allowed.
 */
export async function getUserSession(
  request: Request,
  allowAal1 = false
): Promise<Session | null> {
  const cookie = getCookie(request)
  const requestUrl = new URL(request.url)
  const returnTo = safeReturnTo(requestUrl.pathname + requestUrl.search)
  const searchParams = new URLSearchParams()
  searchParams.set('returnTo', returnTo)

  try {
    const { data } = await kratosPublic.toSession({ cookie })
    return data
  } catch (err: any) {
    const status = err.response?.status

    switch (status) {
      case 401:
      case 500:
        throw redirect(`${route('/login')}?${searchParams.toString()}`)
      case 403:
      case 422: // Need to complete 2FA.
        if (!allowAal1) {
          throw redirect(`${route('/login')}?${searchParams.toString()}`)
        }
        return null
      default:
        throw redirect(`${route('/login')}?${searchParams.toString()}`)
    }
  }
}

/**
 * Extracts typed traits from a Kratos session.
 * Throws redirect to login if session or identity is missing.
 */
export function getSessionTraits(session: Session | null): KratosTraits {
  if (!session?.identity) {
    throw redirect(route('/login'))
  }
  return session.identity.traits as KratosTraits
}

/**
 * hasUserSession allows determining whether the user has a valid kratos session cookie.
 * @param request Request received in a loader function.
 * @returns boolean - if the user has a session cookie.
 */
export function hasUserSession(request: Request): boolean {
  return String(request.headers.get('cookie')).includes(KRATOS_SESSION_COOKIE)
}

/**
 * requireNoUserSession  will ensure the user doesn't already have a session.
 * @param request Request received in a loader function.
 * @returns void
 */
export async function requireNoUserSession(request: Request): Promise<void> {
  // Can immediately assume no session if there's no cookie
  if (!hasUserSession(request)) return

  const cookie = getCookie(request)

  try {
    await kratosPublic.toSession({ cookie })
    // Session is valid — user is already logged in
    throw redirect(route('/'))
  } catch (err: any) {
    // Re-throw redirects
    if (err instanceof Response) throw err

    const status = err.response?.status

    switch (status) {
      case 403:
      case 422: // Need to complete 2FA.
        throw redirect(route('/totp/challenge'))
      case 401:
        // No valid session — this is expected, continue
        return
    }
  }
}
