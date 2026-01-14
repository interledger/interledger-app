import type { Identity, Session } from '@ory/kratos-client'
import { redirect } from '@remix-run/node'
import { KRATOS_URL, getUserSession } from './kratos.server'

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
  '/totp/challenge',
  '/recovery',
  '/recovery/password'
]

/**
 * Routes that can be accessed without verified email
 */
export const NON_VERIFIED_EMAIL_ROUTES = ['/logout', '/verify']

/**
 * Check if TOTP is available in Kratos settings flow
 * This determines if 2FA is configured and available for the current user
 */
export async function isTotpAvailable(request: Request): Promise<boolean> {
  try {
    const cookie = String(request.headers.get('cookie') ?? '')
    const response = await fetch(
      `${KRATOS_URL}/self-service/settings/browser`,
      {
        headers: {
          Accept: 'application/json',
          cookie
        }
      }
    )

    if (!response.ok) {
      return false
    }

    const flow = await response.json()
    const nodes = flow?.ui?.nodes ?? []

    // Check if there are any TOTP-related nodes in the flow
    const hasTotpNodes = nodes.some((node: any) => node.group === 'totp')

    return hasTotpNodes
  } catch (error) {
    console.error('Error checking TOTP availability:', error)
    return false
  }
}

// create a server function that will return true if the user has a totp enabled
export async function isTotpSet(
  session: Session,
  headers: Headers
): Promise<boolean> {
  if (session?.authenticator_assurance_level === 'aal2')
    return Promise.resolve(true)
  try {
    const response = await fetch(
      `${KRATOS_URL}/admin/identities/${session.identity.id}`,
      {
        headers: headers
      }
    )
    if (!response.ok) {
      throw new Error(`Failed to fetch identity`)
    }
    const identity = await response.json()
    return !!identity.credentials?.totp
  } catch (error) {
    console.error('Error checking if TOTP is enabled:', error)
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

export function sessionRequiresAAL2(session: Session): boolean {
  return (session as any).error.id === 'session_aal2_required'
}

export function getSessionIdentity(session: Session): Identity {
  return session.identity
}

export async function emailVerificationGuard(
  pathname: string,
  request: Request
) {
  if (NON_VERIFIED_EMAIL_ROUTES.includes(pathname)) return

  // Request a session WITHOUT AAL2 redirect, so we dont get redirected
  const session = await getUserSession(request, true)
  console.log("user session:", session)
  const identity = getSessionIdentity(session)

  if (!identity) {
    // Kratos requires AAL2 session, so skip email verification guard
    return
  }

  if (identity && !isEmailVerified(session)) {
    throw redirect('/verify')
  }
}

export async function withAAL2Guard(pathname: string, request: Request, fn: () => Promise<void>) {
  if (NON_FULL_SESSION_ROUTES.includes(pathname)) {
    return
  }

  const session = await getUserSession(request)
  const totpAvailable = await isTotpSet(session, request.headers)
  if (!totpAvailable) {
    throw redirect('/totp/two-factor-authentication')
  }

  await fn();
}

export async function recoveryLinkSessionInvalidationGuard(pathname: string, request: Request) {
  console.log(`[recoveryLinkSessionInvalidationGuard] path: ${pathname} : started`)
  if (PASSWORD_RECOVERY_ALLOWED_ROUTES.includes(pathname)) {
    console.log(`[recoveryLinkSessionInvalidationGuard] path: ${pathname}: passw recovery`)
    return
  }

  const session = await getUserSession(request)
  console.log(`[recoveryLinkSessionInvalidationGuard] path: ${pathname} session`, session)

  const isLinkRecoverySession = !!session.authentication_methods?.find((method) => method.method === 'link_recovery')
  console.log(`[recoveryLinkSessionInvalidationGuard] path: ${pathname} isLinkRecoverySession`, isLinkRecoverySession)
  if (isLinkRecoverySession) {
    console.log(`[recoveryLinkSessionInvalidationGuard] path: ${pathname} redirecting to login`)
    throw redirect('/login')
  }
}