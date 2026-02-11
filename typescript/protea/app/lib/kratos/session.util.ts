import type { Session } from '@ory/client'
import { redirect } from '@remix-run/node'
import { route } from 'routes-gen'
import { safeReturnTo } from '../url.server'
import { kratosPublic } from './kratos-client.server'
import { getCookie } from './cookie.util'


/**
 * getUserSession allows fetching a user's kratos session.
 * @param request Request received in a loader function.
 * @returns boolean - if the user has a session.
 */
export async function getUserSession(
  request: Request,
  allowAal1 = false
): Promise<Session> {
  const cookie = getCookie(request)
  const requestUrl = new URL(request.url)
  const returnTo = safeReturnTo(requestUrl.pathname + requestUrl.search)
  const searchParams = new URLSearchParams()
  searchParams.set('returnTo', returnTo)

  try {
    const { data } = await kratosPublic.toSession({ cookie })
    console.log("🐳 [getUserSession] we have a session: ", data)
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
        // allowAal1 === true: return a partial object so callers can safely check .identity
        // The 403 error response is not a Session — it's an error body.
        // Callers (emailVerificationGuard, verify.tsx) already handle missing identity gracefully.
        console.log("🐳 [getUserSession] 403 or 422 with session ", err.response?.data)
        return (err.response?.data ?? {}) as Session
      default:
        throw redirect(`${route('/login')}?${searchParams.toString()}`)
    }
  }
}

/**
 * hasUserSession allows determining whether the user has a valid kratos session cookie.
 * @param request Request received in a loader function.
 * @returns boolean - if the user has a session cookie.
 */
export function hasUserSession(request: Request): boolean {
  return String(request.headers.get('cookie')).includes('ory_kratos_session')
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
