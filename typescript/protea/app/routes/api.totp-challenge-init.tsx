import { json, type ActionFunctionArgs } from '@remix-run/node'
import { printKratosError, UserDisplayableError } from '~/lib/kratos/error.server'
import { kratosPublic } from '~/lib/kratos/kratos-client.server'
import { withCookie } from '~/lib/kratos/cookie.util'
import { getNodeValueFromFlow, isNodeInFlow } from '~/lib/kratos/flow.util'
import { CreateBrowserLoginFlowResponse } from '~/lib/kratos/types.server'
import logger, { addRequestId } from '~/lib/logger.server'
import { extractOrGenerateRequestId } from '~/lib/requestContext.server'

/**
 * Initialize a TOTP challenge flow (AAL2)
 * This is used for inline TOTP verification without page redirects
 */
export async function action({ request }: ActionFunctionArgs) {
  try {
    const cookie = request.headers.get('cookie') ?? ''

    let initAAL2FlowResponse: CreateBrowserLoginFlowResponse
    try {
      initAAL2FlowResponse = await kratosPublic.createBrowserLoginFlow({
        aal: 'aal2',
        refresh: true // reauthenticate even if there is a valid session
      }, withCookie(cookie))
    } catch (err) {
      const message = printKratosError(err)
      throw new Error(message)
    }

    // if ([302, 303].includes(initAAL2FlowResponse.status)) {} // not sure this is applicable anymore
    const flow = initAAL2FlowResponse.data
    const csrfToken = getNodeValueFromFlow(flow, "csrf_token")
    const hasTotpSetUp = isNodeInFlow(flow, "totp_code")
    if (!hasTotpSetUp) {
      throw new UserDisplayableError("TOTP is not configured for this account")
    }

    return json({
      flowId: flow.id,
      csrfToken,
      method: flow.ui?.method ?? 'POST',
      action: flow.ui?.action ?? '',
      shouldRevalidate: false
    })
  } catch (error) {
    const requestId = extractOrGenerateRequestId(request)
    const errorDetails =
      error instanceof Error
        ? { error }
        : { error: String(error) }
    logger.error(
      { ...addRequestId(requestId), ...errorDetails },
      'Error initializing TOTP challenge'
    )
    if (error instanceof UserDisplayableError) {
      return json({ error: error.message })
    }
    return json({ error: 'An unexpected error occurred' })
  }
}
