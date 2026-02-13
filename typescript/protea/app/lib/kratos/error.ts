import { redirect } from "@remix-run/node"
import { route } from "routes-gen"
import { KratosError, KratosErrorId, KratosMessage, type UiNode } from "./types"

export class UserDisplayableError extends Error {
  constructor(public message: string) {
    super(message)
  }
}

export const printKratosError = (err: any) => {
  const kratosError = err as KratosError
  const errorIdData = kratosError.response?.data?.error?.id
  const errorId = kratosError.id
  const status = kratosError.response?.status
  const message = `Error initializing flow, errorIdData ${errorIdData} and status ${status} and errorId ${errorId}`
  console.error(message)

  return message
}
/**
 * Use this when catching flow initialization/fetch errors
 */

export function handleFlowError(
  error: any,
  redirectTo: 'login' |
    'signup' |
    'settings' |
    'settings/password' |
    'settings/phone' |
    'otp/challenge' |
    'login/challenge' |
    'recovery' |
    'recovery/password' |
    'verify' |
    'logout' |
    'totp' |
    'totp/challenge'
): void {
  const flowData = error.response?.data ?? error

  if (!flowData?.error?.id) {
    console.error('[handleFlowError] Unrecognized error shape:', {
      status: error.response?.status,
      data: flowData
    })
    return
  }

  let redirectRoute = `/${redirectTo}`

  switch (flowData.error.id) {
    case 'session_inactive':
      throw redirect(route('/login'), {
        headers: { 'Clear-Site-Data': 'cookies' }
      })
    case 'session_aal2_required':
      throw redirect(`/totp/challenge?returnTo=${redirectRoute}`)
    case 'session_already_available':
      throw redirect(route('/'))
    case 'session_refresh_required':
      throw redirect(`/login/challenge`) // todo ????
    case 'self_service_flow_expired':
    case 'security_csrf_violation':
    case 'security_identity_mismatch':
      throw redirect(redirectRoute, {
        headers: { 'Clear-Site-Data': 'cookies' }
      })
    case 'browser_location_change_required':
      throw redirect(flowData.error.redirect_browser_to)
  }
}/**
 * Overrides the kratos error messages.
 * Will pass on the kratos error message if not overridden.
 * @param error The kratos error message.
 */


export function kratosErrorMessage(error: KratosMessage): string {
  switch (error.id) {
    case KratosErrorId.ErrorValidationInvalidCredentials:
      return 'The provided credentials are invalid.'
    case KratosErrorId.ErrorValidationDuplicateCredentials:
      return 'An account with the same identifier (email, phone, username, ...) exists already.'
    default:
      return error.text
  }
}
/**
 * Helper to map SDK flow errors to field errors
 * Use this when catching SDK exceptions in route actions
 */

export function mapFlowToFieldErrors<T extends object>(
  flowData: any,
  fieldErrors: T
): T {
  if (!flowData?.ui?.nodes) return fieldErrors

  for (const node of flowData.ui.nodes as UiNode[]) {
    if (node.messages && node.messages.length > 0) {
      const attrName = (node.attributes as any).name
      if (attrName) {
        (fieldErrors as any)[attrName] = kratosErrorMessage(node.messages[0])
      }
    }
  }

  if (flowData.ui.messages?.length > 0) {
    (fieldErrors as any).form = kratosErrorMessage(flowData.ui.messages[0])
  }

  return fieldErrors
}

