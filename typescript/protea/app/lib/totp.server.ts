import { KRATOS_URL } from './kratos.server'

/**
 * Routes that do not require TOTP enabled check when user is logged in
 */
export const NON_TOTP_ROUTES = [
  '/totp/two-factor-authentication',
  '/totp/challenge',
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
