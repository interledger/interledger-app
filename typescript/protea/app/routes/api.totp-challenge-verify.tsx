import { data } from 'react-router'
import {
  buildHeadersWithCookies,
  getCookie,
  withCookie
} from '~/lib/kratos/cookie.server'
import { mapFlowToFieldErrors } from '~/lib/kratos/error.server'
import { kratosPublic } from '~/lib/kratos/kratos-client.server'
import type { KratosError } from '~/lib/kratos/types.server'
import logger, { addRequestId, withErrorLog } from '~/lib/logger.server'
import { extractOrGenerateRequestId } from '~/lib/requestContext.server'
import type { Route } from './+types/api.totp-challenge-verify'

/**
 * Verify TOTP code for AAL2 challenge
 * This handles the inline TOTP verification without redirects
 * IMPORTANT: Forwards Set-Cookie header to update the session
 */
export async function action({ request }: Route.ActionArgs) {
  try {
    const formData = await request.formData()
    const flow = formData.get('flow') as string
    const totpCode = formData.get('totp_code') as string
    const csrfToken = formData.get('csrf_token') as string
    const cookie = getCookie(request)

    if (!flow || !totpCode) {
      return data(
        { success: false, error: 'Please reinitialize the verification.' },
        { status: 400 }
      )
    }

    const submitTotpResponse = await kratosPublic.updateLoginFlow(
      {
        flow,
        updateLoginFlowBody: {
          method: 'totp',
          totp_code: totpCode,
          csrf_token: csrfToken
        }
      },
      withCookie(cookie)
    )
    const headers = buildHeadersWithCookies(submitTotpResponse)

    return data({ success: true, shouldRevalidate: false }, { headers })
  } catch (error) {
    const kratosError = error as KratosError
    const flowStatus = kratosError.response?.status
    const flowData = kratosError.response?.data
    const requestId = extractOrGenerateRequestId(request)
    logger.error(
      {
        ...addRequestId(requestId),
        ...withErrorLog(error)
      },
      'Unexpected error verifying TOTP challenge'
    )

    if (flowStatus === 400 && flowData) {
      const errorMapping: Record<string, string> = {}
      mapFlowToFieldErrors(flowData, errorMapping)
      return data({
        success: false,
        error:
          errorMapping.form || errorMapping.totp_code || 'Invalid TOTP code',
        flowId: flowData.id
      })
    }

    if (flowStatus === 410) {
      return data({
        success: false,
        error: 'Flow expired. Please reinitialize the verification.'
      })
    }

    return data({
      success: false,
      error: 'An unexpected error occurred'
    })
  }
}
