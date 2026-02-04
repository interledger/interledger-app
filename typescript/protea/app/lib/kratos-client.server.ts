/**
 * Ory Kratos SDK Client
 *
 * This module provides a centralized Kratos SDK client for server-side use.
 * It wraps the official @ory/kratos-client SDK for use with Remix loaders/actions.
 */

import {
  Configuration,
  type LoginFlow,
  type LogoutFlow,
  type RecoveryFlow,
  type RegistrationFlow,
  type SettingsFlow,
  type VerificationFlow,
  type Session,
  type UiNodeInputAttributes,
  type UiNode,
  type UpdateLoginFlowWithPasswordMethod,
  type UpdateLoginFlowWithTotpMethod,
  type UpdateRegistrationFlowWithPasswordMethod,
  type UpdateSettingsFlowWithPasswordMethod,
  type UpdateSettingsFlowWithProfileMethod,
  type UpdateSettingsFlowWithTotpMethod,
  type UpdateRecoveryFlowWithLinkMethod,
  type UpdateVerificationFlowWithLinkMethod,
  FrontendApi,
  IdentityApi
} from '@ory/client'
import { redirect } from '@remix-run/node'
import { route } from 'routes-gen'
import { safeReturnTo } from './url.server'

// Validate environment
const KRATOS_PUBLIC_URL = process.env.KRATOS_URL
const KRATOS_ADMIN_URL = process.env.KRATOS_ADMIN_URL

if (!KRATOS_PUBLIC_URL) {
  throw new Error('KRATOS_URL environment variable is not set')
}

// Create SDK configuration
const publicConfig = new Configuration({
  basePath: KRATOS_PUBLIC_URL,
  baseOptions: {
    withCredentials: true
  }
})

const adminConfig = KRATOS_ADMIN_URL
  ? new Configuration({
      basePath: KRATOS_ADMIN_URL,
      baseOptions: {
        withCredentials: true
      }
    })
  : null

// Create API clients using factory pattern (SDK uses this approach)
export const kratosPublic = new FrontendApi(publicConfig)
export const kratosAdmin = adminConfig ? new IdentityApi(adminConfig) : null

// Re-export types for convenience
export type {
  LoginFlow,
  LogoutFlow,
  RecoveryFlow,
  RegistrationFlow,
  SettingsFlow,
  VerificationFlow,
  Session,
  UiNodeInputAttributes,
  UiNode,
  UpdateLoginFlowWithPasswordMethod,
  UpdateLoginFlowWithTotpMethod,
  UpdateRegistrationFlowWithPasswordMethod,
  UpdateSettingsFlowWithPasswordMethod,
  UpdateSettingsFlowWithProfileMethod,
  UpdateSettingsFlowWithTotpMethod,
  UpdateRecoveryFlowWithLinkMethod,
  UpdateVerificationFlowWithLinkMethod
}

// Flow type union for CSRF token extraction (flows with UI containing nodes)
export type KratosFlowWithUi =
  | LoginFlow
  | RecoveryFlow
  | RegistrationFlow
  | SettingsFlow
  | VerificationFlow

/**
 * Helper to extract cookie header from Request for SDK calls
 */
export function getCookieHeader(request: Request): string {
  return request.headers.get('cookie') ?? ''
}

/**
 * Helper to extract set-cookie header from Axios response
 */
export function extractSetCookieHeader(
  response: { headers?: Record<string, unknown> }
): string | undefined {
  const setCookie = response.headers?.['set-cookie']
  if (!setCookie) return undefined
  if (Array.isArray(setCookie)) return setCookie.join(', ')
  return typeof setCookie === 'string' ? setCookie : undefined
}

/**
 * Helper to build response headers with set-cookie from Kratos response
 */
export function buildHeadersWithCookies(
  response: { headers?: Record<string, unknown> }
): Headers | undefined {
  const setCookie = extractSetCookieHeader(response)
  if (!setCookie) return undefined

  const headers = new Headers()
  headers.set('Set-Cookie', setCookie)
  return headers
}

/**
 * Axios request config type (inline to avoid axios dependency)
 */
type RequestConfig = {
  headers?: Record<string, string>
  withCredentials?: boolean
}

/**
 * Creates axios request config with cookie header for authenticated requests
 */
export function withCookie(cookie: string): RequestConfig {
  return {
    headers: {
      Cookie: cookie
    },
    withCredentials: true
  }
}

/**
 * Check if a UI node has input attributes
 */
export function isUiNodeInputAttributes(n: unknown): n is UiNodeInputAttributes {
  return typeof n === 'object' && n !== null && 'name' in n
}

/**
 * Extract CSRF token from any Kratos flow with UI
 */
export function getCsrfTokenFromFlow(flow: KratosFlowWithUi | undefined): string {
  if (!flow?.ui?.nodes) return ''

  const node = flow.ui.nodes.find(
    (node) =>
      isUiNodeInputAttributes(node?.attributes) &&
      node.attributes.name === 'csrf_token'
  )

  return node ? (node.attributes as UiNodeInputAttributes).value ?? '' : ''
}

/**
 * Kratos error IDs for error message overrides
 */
enum KratosErrorId {
  ErrorValidationInvalidCredentials = 4000006,
  ErrorValidationDuplicateCredentials = 4000007
}

type KratosMessage = {
  id: number
  text: string
  context?: object
}

/**
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

/**
 * Updated handleFlowError for SDK exceptions
 * Use this when catching flow initialization/fetch errors
 */
export function handleFlowError(
  error: any,
  flowType:
    | 'login'
    | 'signup'
    | 'settings'
    | 'settings/password'
    | 'settings/phone'
    | 'otp/challenge'
    | 'login/challenge'
    | 'recovery'
    | 'recovery/password'
    | 'verify'
    | 'logout'
    | 'totp'
    | 'totp/challenge',
  flowId?: string
): void {
  const flowData = error.response?.data ?? error

  if (!flowData?.error?.id) return

  let redirectRoute = `/${flowType}`

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
      throw redirect(`/login/challenge`)
    case 'self_service_flow_expired':
    case 'security_csrf_violation':
    case 'security_identity_mismatch':
      throw redirect(redirectRoute, {
        headers: { 'Clear-Site-Data': 'cookies' }
      })
    case 'browser_location_change_required':
      throw redirect(flowData.error.redirect_browser_to)
  }
}

/**
 * Check if a message indicates a session already exists
 */
export function isSessionAlreadyExistsMessage(msg: string): boolean {
  return (
    msg.includes('refresh=true') && msg.includes('valid session was detected')
  )
}
