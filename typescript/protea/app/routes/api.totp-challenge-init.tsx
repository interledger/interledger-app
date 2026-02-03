import { json, type ActionFunctionArgs } from '@remix-run/node'
import { KRATOS_URL } from '~/lib/kratos.server'
import logger, { addRequestId } from '~/lib/logger.server'

/**
 * Initialize a TOTP challenge flow (AAL2) via Kratos API endpoint
 * This is used for inline TOTP verification without page redirects
 */
export async function action({ request }: ActionFunctionArgs) {
  try {
    const cookie = request.headers.get('cookie') ?? ''

    // Step 1: Initialize an AAL2 login flow via BROWSER (not API)
    // We use browser mode because we need to send the existing session cookie
    // refresh=true tells Kratos we want to re-authenticate even though we have a valid session
    // redirect: "manual" prevents automatic redirects (we handle them ourselves)
    const initResponse = await fetch(
      `${KRATOS_URL}/self-service/login/browser?aal=aal2&refresh=true`,
      {
        method: 'GET',
        headers: {
          Accept: 'application/json',
          cookie
        },
        redirect: 'manual'
      }
    )

    // Step 2: Browser mode returns a redirect (302/303) with flow ID in Location header
    // Check for all possible redirect status codes
    if (initResponse.status >= 300 && initResponse.status < 400) {
      const location = initResponse.headers.get('location')

      if (!location) {
        return json({ error: 'Failed to initialize TOTP challenge' })
      }

      // Extract flow ID from the redirect URL
      const flowId = new URL(location, KRATOS_URL).searchParams.get('flow')

      if (!flowId) {
        return json({ error: 'Failed to extract flow ID' })
      }

      // Step 3: Fetch the full flow details using the flow ID
      const flowResponse = await fetch(
        `${KRATOS_URL}/self-service/login/flows?id=${flowId}`,
        {
          headers: {
            Accept: 'application/json',
            cookie
          }
        }
      )

      if (!flowResponse.ok) {
        return json(
          { error: 'Failed to fetch flow details' },
          { status: flowResponse.status }
        )
      }

      const flow = await flowResponse.json()
      const nodes = flow.ui?.nodes ?? []
      const csrfNode = nodes.find(
        (node: any) => node.attributes?.name === 'csrf_token'
      )
      const csrfToken = csrfNode?.attributes?.value ?? ''
      const totpNode = nodes.find(
        (node: any) => node.attributes?.name === 'totp_code'
      )

      if (!totpNode) {
        return json({ error: 'TOTP is not configured for this account' })
      }

      return json({
        flowId,
        csrfToken,
        method: flow.ui?.method ?? 'POST',
        action: flow.ui?.action ?? '',
        shouldRevalidate: false
      })
    }

    // If we got a 200, Kratos might have returned the flow directly (when Accept: application/json is set)
    if (initResponse.status === 200) {
      try {
        const flow = await initResponse.json()

        if (flow.id) {
          // Extract necessary data from the flow
          const nodes = flow.ui?.nodes ?? []
          const csrfNode = nodes.find(
            (node: any) => node.attributes?.name === 'csrf_token'
          )
          const csrfToken = csrfNode?.attributes?.value ?? ''
          const totpNode = nodes.find(
            (node: any) => node.attributes?.name === 'totp_code'
          )

          if (!totpNode) {
            return json({ error: 'TOTP is not configured for this account' })
          }

          return json({
            flowId: flow.id,
            csrfToken,
            method: flow.ui?.method ?? 'POST',
            action: flow.ui?.action ?? '',
            shouldRevalidate: false
          })
        }
      } catch (e) {}
    }

    const errorText = await initResponse.text()
    return json({ error: `Unexpected response from Kratos ${errorText}` })
  } catch (error) {
    const requestId = 'unknown'
    logger.error(
      { ...addRequestId(requestId), error: error instanceof Error ? error.message : String(error) },
      'Error initializing TOTP challenge'
    )
    return json({ error: 'An unexpected error occurred' })
  }
}
