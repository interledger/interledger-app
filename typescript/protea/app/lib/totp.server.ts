import type { Session } from '@ory/client'
import { redirect } from '@remix-run/node'
import { getUserSession } from './kratos/session.util'
import { kratosPublic } from './kratos/kratos-client.server'
import { withCookie } from './kratos/cookie.util'

/**
 * Routes that can be accessed without a session with highest AAL
 */
export const NON_FULL_SESSION_ROUTES = [
  '/totp/two-factor-authentication',
  '/totp/challenge',
  '/login',
  '/logout',
  '/signup',
  '/recovery/password',
  '/unavailable',
  '/verify'
]

const PASSWORD_RECOVERY_ALLOWED_ROUTES = [
  '/login',
  '/logout', // temporary fix for logging out when on challenge page
  '/totp/challenge',
  '/recovery',
  '/recovery/password'
]

/**
 * Routes that can be accessed without verified email
 */
export const NON_VERIFIED_EMAIL_ROUTES = ['/logout', '/verify']

/**
 * Check if the user has TOTP enabled
 */
export async function isTotpSet(
  session: Session,
  headers: Headers
): Promise<boolean> {
  if (session?.authenticator_assurance_level === 'aal2') {
    return true
  }

  try {
    const cookie = headers.get('cookie') ?? ''
    const { data: flow } = await kratosPublic.createBrowserSettingsFlow(
      undefined,
      withCookie(cookie)
    )

    const nodes = flow.ui.nodes ?? []
    // If TOTP is configured, the settings flow contains totp group nodes
    // with an "unlink" action. If not configured, the nodes offer "enable".
    const isSet = nodes.some(
      (node: any) => node.group === 'totp' && node.attributes?.name === 'totp_unlink'
    )
    return isSet
  } catch (error) {
    return false
  }
}

export async function ensureTOTP(
  session: Session,
  headers: Headers,
  returnTo: string
): Promise<void> {
  const hasTotp = await isTotpSet(session, headers)
  if (!hasTotp) {
    throw redirect(
      `/totp/two-factor-authentication?refresh=true&returnTo=${encodeURIComponent(
        returnTo
      )}`,
      {
        headers: headers
      }
    )
  }
}

/**
 * Check if user's email is verified
 * Returns true if the user has verified their email address
 */
export function isEmailVerified(session: Session): boolean {
  return !!(
    session.identity?.verifiable_addresses &&
    session.identity.verifiable_addresses.length > 0 &&
    session.identity.verifiable_addresses[0]?.verified
  )
}

export async function emailVerificationGuard(
  pathname: string,
  request: Request
) {
  if (NON_VERIFIED_EMAIL_ROUTES.includes(pathname)) return

  const session = await getUserSession(request, true)

  if (!session) {
    // Session requires AAL2 upgrade — skip email verification guard
    return
  }

  if (!isEmailVerified(session)) {
    throw redirect('/verify')
  }
}

export async function withAAL2Guard(pathname: string, request: Request, fn: () => Promise<void>) {
  if (NON_FULL_SESSION_ROUTES.includes(pathname)) {
    return
  }

  const session = await getUserSession(request)
  if (!session) throw redirect('/login')
  const totpAvailable = await isTotpSet(session, request.headers)
  if (!totpAvailable) {
    throw redirect('/totp/two-factor-authentication')
  }

  await fn();
}

export async function recoveryLinkSessionInvalidationGuard(pathname: string, request: Request) {
  if (PASSWORD_RECOVERY_ALLOWED_ROUTES.includes(pathname)) {
    return
  }

  const session = await getUserSession(request)
  if (!session) throw redirect('/login')

  const isLinkRecoverySession = !!session.authentication_methods?.some((method: any) => method.method === 'link_recovery')
  if (isLinkRecoverySession) {
    throw redirect('/login', {
      headers: {
        'Set-Cookie': 'ory_kratos_session=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax'
      }
    })
  }
}