import { json, type LoaderFunctionArgs } from '@remix-run/node'
import { KRATOS_URL } from '~/lib/kratos.server'

/**
 * Initialize a TOTP challenge flow (AAL2) via Kratos API endpoint
 * This is used for inline TOTP verification without page redirects
 */
export async function loader({ request }: LoaderFunctionArgs) {
  try {
    const cookie = request.headers.get('cookie') ?? ''

    // Step 1: Initialize an AAL2 login flow via BROWSER (not API)
    // We use browser mode because we need to send the existing session cookie
    // refresh=true tells Kratos we want to re-authenticate even though we have a valid session
    // redirect: "manual" prevents automatic redirects (we handle them ourselves)
    console.log('🍀 (totp-challenge-init) Initializing TOTP challenge flow')
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

    console.log(
      '🍀 (totp-challenge-init) Init response status:',
      initResponse.status
    )
    console.log(
      '🍀 (totp-challenge-init) Init response headers:',
      Object.fromEntries(initResponse.headers.entries())
    )

    // Step 2: Browser mode returns a redirect (302/303) with flow ID in Location header
    // Check for all possible redirect status codes
    if (initResponse.status >= 300 && initResponse.status < 400) {
      const location = initResponse.headers.get('location')
      console.log('🍀 (totp-challenge-init) Got redirect, location:', location)

      if (!location) {
        console.error('🍀 (totp-challenge-init) No location header in redirect')
        return json(
          { error: 'Failed to initialize TOTP challenge' },
          { status: 500 }
        )
      }

      // Extract flow ID from the redirect URL
      const flowId = new URL(location, KRATOS_URL).searchParams.get('flow')
      console.log('🍀 (totp-challenge-init) Extracted flow ID:', flowId)

      if (!flowId) {
        console.error('🍀 (totp-challenge-init) No flow ID in redirect URL')
        return json({ error: 'Failed to extract flow ID' }, { status: 500 })
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
        const error = await flowResponse.text()
        console.error('🍀 (totp-challenge-init) Failed to fetch flow:', error)
        return json(
          { error: 'Failed to fetch flow details' },
          { status: flowResponse.status }
        )
      }

      console.log('🍀 (totp-challenge-init) Successfully fetched flow details')
      const flow = await flowResponse.json()

      // Step 4: Extract necessary data from the flow
      const nodes = flow.ui?.nodes ?? []

      // Find the CSRF token node
      const csrfNode = nodes.find(
        (node: any) => node.attributes?.name === 'csrf_token'
      )
      const csrfToken = csrfNode?.attributes?.value ?? ''

      // Find the TOTP code input node
      const totpNode = nodes.find(
        (node: any) => node.attributes?.name === 'totp_code'
      )

      if (!totpNode) {
        console.log(
          '🍀 (totp-challenge-init) TOTP node not found in flow. User may not have TOTP set up.'
        )
        console.error(
          'TOTP node not found in flow. User may not have TOTP set up.'
        )
        return json(
          { error: 'TOTP is not configured for this account' },
          { status: 400 }
        )
      }

      console.log(
        '🍀 (totp-challenge-init) Successfully finalized TOTP challenge flow'
      )
      return json({
        flowId,
        csrfToken,
        method: flow.ui?.method ?? 'POST',
        action: flow.ui?.action ?? ''
      })
    }

    // If we got a 200, Kratos might have returned the flow directly (when Accept: application/json is set)
    if (initResponse.status === 200) {
      console.log(
        '🍀 (totp-challenge-init) Got 200 response, checking if flow data is present'
      )
      try {
        const flow = await initResponse.json()

        if (flow.id) {
          console.log(
            '🍀 (totp-challenge-init) Flow data present in 200 response, flow ID:',
            flow.id
          )

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
            console.log(
              '🍀 (totp-challenge-init) TOTP node not found in flow. User may not have TOTP set up.'
            )
            return json(
              { error: 'TOTP is not configured for this account' },
              { status: 400 }
            )
          }

          console.log(
            '🍀 (totp-challenge-init) Successfully extracted flow data from 200 response'
          )
          return json({
            flowId: flow.id,
            csrfToken,
            method: flow.ui?.method ?? 'POST',
            action: flow.ui?.action ?? ''
          })
        }
      } catch (e) {
        console.error(
          '🍀 (totp-challenge-init) Failed to parse 200 response as JSON:',
          e
        )
      }
    }

    // If we didn't get a redirect or valid flow, something went wrong
    console.error(
      '🍀 (totp-challenge-init) Unexpected response status:',
      initResponse.status
    )
    const error = await initResponse.text()
    console.error('🍀 (totp-challenge-init) Response body:', error)
    return json(
      { error: 'Unexpected response from Kratos' },
      { status: initResponse.status }
    )
  } catch (error) {
    console.error('Error initializing TOTP challenge:', error)
    return json({ error: 'An unexpected error occurred' }, { status: 500 })
  }
}
