import { json, type LoaderFunctionArgs } from '@remix-run/node'
import { KRATOS_URL } from '~/lib/kratos.server'
import { isTotpSet } from '~/lib/totp.server'

/**
 * Check if the current user has TOTP enabled
 * This is called before showing the TOTP challenge popup
 * to ensure we don't prompt users who haven't set up 2FA yet
 */
export async function loader({ request }: LoaderFunctionArgs) {
  try {
    const cookie = request.headers.get('cookie') ?? ''

    // Get current session
    const sessionResponse = await fetch(`${KRATOS_URL}/sessions/whoami`, {
      headers: { cookie }
    })

    if (!sessionResponse.ok) {
      console.log('🍀 (check-totp-enabled) No valid session')
      // Return 200 with enabled: false so fetcher.data populates correctly
      return json({ enabled: false, error: 'No valid session' })
    }

    const session = await sessionResponse.json()

    // Check if TOTP is configured for this user
    const headers = new Headers(request.headers)
    const hasTotpEnabled = await isTotpSet(session, headers)

    console.log(
      '🍀 (check-totp-enabled) TOTP enabled:',
      hasTotpEnabled,
      'for user:',
      session.identity.id
    )

    return json({ enabled: hasTotpEnabled })
  } catch (error) {
    console.error('🍀 (check-totp-enabled) Error:', error)
    // Return 200 with enabled: false so fetcher.data populates correctly
    return json({ enabled: false, error: 'Failed to check TOTP status' })
  }
}
