import type { Session } from '@ory/client'
import { redirect } from 'react-router'
import { withCookie } from './kratos/cookie.server'
import {
  CLEAR_SESSION_COOKIE_HEADER,
  kratosPublic
} from './kratos/kratos-client.server'
import { getUserSession } from './kratos/session.server'
import { safeReturnTo } from './url.server'

/**
 * Routes that can be accessed without a session with highest AAL
 */
export const NON_FULL_SESSION_ROUTES = [
  '/phone-confirmation',
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
  '/recovery/password',
  '/settings'
]

/**
 * Routes that can be accessed without verified email
 */
export const NON_VERIFIED_EMAIL_ROUTES = ['/logout', '/verify']

/**
 * Check if the user has TOTP enabled
 */
async function isTotpSet(session: Session, headers: Headers): Promise<boolean> {
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
      (node: any) =>
        node.group === 'totp' && node.attributes?.name === 'totp_unlink'
    )
    return isSet
  } catch (error) {
    return false
  }
}

function getSessionTraits(session: Session): {
  phone?: string
  phoneVerified?: boolean
} {
  return (session.identity?.traits as any) ?? {}
}

function isEmailVerified(session: Session): boolean {
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

export async function withAAL2Guard(
  pathname: string,
  request: Request,
  fn: (session: Session) => Promise<void>
) {
  if (NON_FULL_SESSION_ROUTES.includes(pathname)) {
    return
  }

  const url = new URL(request.url)
  if (pathname === '/settings' && url.searchParams.has('flow')) {
    return
  }

  const session = await getUserSession(request)
  if (!session) throw redirect('/login')
  const totpAvailable = await isTotpSet(session, request.headers)
  if (!totpAvailable) {
    throw redirect('/totp/two-factor-authentication')
  }

  await fn(session)
}

/**
 * Routes that can be accessed without phone confirmation
 */
export const NON_PHONE_CONFIRMED_ROUTES = [
  '/phone-confirmation',
  '/settings/phone',
  '/totp/challenge',
  '/login',
  '/logout',
  '/signup',
  '/recovery',
  '/recovery/password',
  '/verify',
  '/wallet-address',
  '/unavailable'
]

export async function phoneConfirmationGuard(
  pathname: string,
  request: Request
) {
  const session = await getUserSession(request, true)
  if (NON_PHONE_CONFIRMED_ROUTES.includes(pathname)) return
  if (!session) return // Not yet AAL2 — skip guard

  const { phone, phoneVerified } = getSessionTraits(session)
  // Skip guard for users without a phone number (legacy users)
  if (!phone) return

  // Read phoneVerified directly from Kratos session traits — no gRPC call needed
  if (!phoneVerified) {
    const url = new URL(request.url)
    const searchParams = new URLSearchParams()
    searchParams.set('returnTo', safeReturnTo(url.pathname + url.search))
    throw redirect(`/phone-confirmation?${searchParams.toString()}`)
  }
}

export async function recoveryLinkSessionInvalidationGuard(
  pathname: string,
  request: Request
) {
  if (PASSWORD_RECOVERY_ALLOWED_ROUTES.includes(pathname)) {
    return
  }

  const session = await getUserSession(request)
  if (!session) throw redirect('/login')

  const isLinkRecoverySession = !!session.authentication_methods?.some(
    (method: any) => method.method === 'link_recovery'
  )
  if (isLinkRecoverySession) {
    throw redirect('/login', {
      headers: {
        'Set-Cookie': CLEAR_SESSION_COOKIE_HEADER
      }
    })
  }
}
