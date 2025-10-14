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
    const sessionResponse = await fetch(`${KRATOS_URL}/sessions/whoami`, {
      headers: { cookie }
    })

    if (!sessionResponse.ok) {
      return json({ enabled: false, error: 'No valid session' })
    }

    const session = await sessionResponse.json()
    const headers = new Headers(request.headers)
    const hasTotpEnabled = await isTotpSet(session, headers)

    return json({ enabled: hasTotpEnabled })
  } catch (error) {
    return json({ enabled: false, error: 'Failed to check TOTP status' })
  }
}
