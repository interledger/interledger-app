import type { Session } from '@ory/kratos-client'
import { redirect } from '@remix-run/node'
import { KRATOS_URL } from './kratos.server'

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
  '/settings'
]

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

  if(session?.authenticator_assurance_level === 'aal2')
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
