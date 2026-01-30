import { json, type ActionFunctionArgs } from '@remix-run/node'
import { KRATOS_URL } from '~/lib/kratos.server'
import { getLogger, addRequestId } from '~/lib/logger.server'

/**
 * Verify TOTP code for AAL2 challenge
 * This handles the inline TOTP verification without redirects
 * IMPORTANT: Forwards Set-Cookie header to update the session
 */
export async function action({ request }: ActionFunctionArgs) {
  try {
    const cookie = request.headers.get('cookie') ?? ''
    const formData = await request.formData()

    const flowId = formData.get('flow') as string
    const totpCode = formData.get('totp_code') as string
    const csrfToken = formData.get('csrf_token') as string

    if (!flowId || !totpCode) {
      return json(
        { success: false, error: 'Missing required fields' },
        { status: 400 }
      )
    }

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

    const setCookieHeader = verifyTotpCodeResponse.headers.get('set-cookie')
    if (verifyTotpCodeResponse.ok || verifyTotpCodeResponse.status === 200) {
      return json(
        { success: true, shouldRevalidate: false },
        setCookieHeader
          ? { headers: { 'Set-Cookie': setCookieHeader } }
          : undefined
      )
    }

    const errorData = await verifyTotpCodeResponse.json()

    let errorMessage = 'Invalid TOTP code'
    if (errorData?.ui?.messages && errorData.ui.messages.length > 0) {
      errorMessage = errorData.ui.messages[0].text
    } else if (errorData?.ui?.nodes) {
      const totpNode = errorData.ui.nodes.find(
        (node: any) => node.attributes?.name === 'totp_code'
      )
      if (totpNode?.messages && totpNode.messages.length > 0) {
        errorMessage = totpNode.messages[0].text
      }
    }

    return json({
      success: false,
      error: errorMessage,
      flowId: errorData.id,
    })
  } catch (error) {
    const logger = getLogger()
    const requestId = 'unknown' // In Remix we can get this from request headers if added
    logger.error(
      { ...addRequestId(requestId), error: error instanceof Error ? error.message : String(error) },
      'Unexpected error verifying TOTP challenge'
    )
    return json({
      success: false,
      error: 'An unexpected error occurred'
    })
  }
}
