import { json, type ActionFunctionArgs } from '@remix-run/node'
import { KRATOS_URL } from '~/lib/kratos.server'

/**
 * Verify TOTP code for AAL2 challenge
 * This handles the inline TOTP verification without redirects
 * IMPORTANT: Forwards Set-Cookie header to update the session
 */
export async function action({ request }: ActionFunctionArgs) {
  console.log('🍀 (totp-challenge-verify) Starting TOTP verification')

  try {
    const cookie = request.headers.get('cookie') ?? ''
    const formData = await request.formData()

    const flowId = formData.get('flow') as string
    const totpCode = formData.get('totp_code') as string
    const csrfToken = formData.get('csrf_token') as string

    console.log('🍀 (totp-challenge-verify) Flow ID:', flowId)
    console.log(
      '🍀 (totp-challenge-verify) TOTP code length:',
      totpCode?.length
    )

    if (!flowId || !totpCode) {
      console.log('🍀 (totp-challenge-verify) Missing required fields')
      return json(
        { success: false, error: 'Missing required fields' },
        { status: 400 }
      )
    }

    // Submit the TOTP code to Kratos
    console.log('🍀 (totp-challenge-verify) Submitting to Kratos...')
    const verifyTotpCodeResponse = await fetch(
      `${KRATOS_URL}/self-service/login?flow=${flowId}`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Accept: 'application/json',
          cookie
        },
        body: JSON.stringify({
          method: 'totp',
          totp_code: totpCode,
          csrf_token: csrfToken
        })
      }
    )

    console.log(
      '🍀 (totp-challenge-verify) Kratos response status:',
      verifyTotpCodeResponse.status
    )

    // Extract the new session cookie from Kratos response
    const setCookieHeader = verifyTotpCodeResponse.headers.get('set-cookie')
    console.log(
      '🍀 (totp-challenge-verify) Set-Cookie header present:',
      !!setCookieHeader
    )

    // Check if verification was successful
    if (verifyTotpCodeResponse.ok || verifyTotpCodeResponse.status === 200) {
      console.log('🍀 (totp-challenge-verify) ✅ TOTP verified successfully')
      // ✅ TOTP verified successfully
      // Forward the Set-Cookie header to update the browser's session
      return json(
        { success: true },
        setCookieHeader
          ? { headers: { 'Set-Cookie': setCookieHeader } }
          : undefined
      )
    }

    // Handle error responses
    console.log(
      '🍀 (totp-challenge-verify) Verification failed, parsing error...'
    )
    const errorData = await verifyTotpCodeResponse.json()
    console.log(
      '🍀 (totp-challenge-verify) Error data:',
      JSON.stringify(errorData, null, 2)
    )

    // Extract error message from Kratos flow UI
    let errorMessage = 'Invalid TOTP code'
    if (errorData?.ui?.messages && errorData.ui.messages.length > 0) {
      errorMessage = errorData.ui.messages[0].text
    } else if (errorData?.ui?.nodes) {
      // Check for node-specific errors (e.g., on totp_code field)
      const totpNode = errorData.ui.nodes.find(
        (node: any) => node.attributes?.name === 'totp_code'
      )
      if (totpNode?.messages && totpNode.messages.length > 0) {
        errorMessage = totpNode.messages[0].text
      }
    }

    console.log(
      '🍀 (totp-challenge-verify) Returning error response:',
      errorMessage
    )
    // Return 200 status with success: false so fetcher.data populates correctly
    return json({
      success: false,
      error: errorMessage,
      flowId: errorData.id // Return new flow ID in case of error
    })
  } catch (error) {
    console.error('🍀 (totp-challenge-verify) ❌ Unexpected error:', error)
    // Return 200 status with success: false so fetcher.data populates correctly
    return json({
      success: false,
      error: 'An unexpected error occurred'
    })
  }
}
