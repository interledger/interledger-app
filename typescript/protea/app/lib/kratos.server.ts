import type {
  LoginFlow,
  RecoveryFlow,
  RegistrationFlow,
  SettingsFlow,
  VerificationFlow,
  Session,
  UiNodeInputAttributes,
  SuccessfulNativeRegistration
} from '@ory/kratos-client'
import { redirect } from '@remix-run/node'
import { route } from 'routes-gen'
import { safeReturnTo } from './url.server'
import {
  getCsrfTokenFromFlow as getCsrfToken,
  isUiNodeInputAttributes
} from './kratos-client.server'

// Export to ensure this is always evaluated server side.
export const KRATOS_URL = process.env.KRATOS_URL

// Re-export types for backward compatibility
export type {
  LoginFlow,
  RecoveryFlow,
  RegistrationFlow,
  SettingsFlow,
  VerificationFlow,
  Session,
  UiNodeInputAttributes,
  SuccessfulNativeRegistration
}

// Type aliases for backward compatibility with existing code
export type SelfServiceLoginFlow = LoginFlow
export type SelfServiceRecoveryFlow = RecoveryFlow
export type SelfServiceRegistrationFlow = RegistrationFlow
export type SelfServiceSettingsFlow = SettingsFlow
export type SelfServiceVerificationFlow = VerificationFlow
export type SuccessfulSelfServiceRegistrationWithoutBrowser = SuccessfulNativeRegistration

/**
 * Extract CSRF token from flow - re-exported from kratos-client.server.ts
 */
export const getCsrfTokenFromFlow = getCsrfToken

/**
 * getUserSession allows fetching a user's kratos session.
 * @param request Request received in a loader function.
 * @returns boolean - if the user has a session.
 */
export async function getUserSession(
  request: Request,
  allowAal1 = false
): Promise<Session> {
  const session = await fetch(`${KRATOS_URL}/sessions/whoami`, {
    headers: request.headers
  })

  const requestUrl = new URL(request.url)
  const returnTo = safeReturnTo(requestUrl.pathname + requestUrl.search)
  const searchParams = new URLSearchParams()
  searchParams.set('returnTo', returnTo)

  switch (session.status) {
    case 401:
    case 500:
      throw redirect(`${route('/login')}?${searchParams.toString()}`)
    case 403:
    case 422: // Need to complete 2FA.
      if (!allowAal1) {
        requestUrl.searchParams.set('aal', 'aal2')
        throw redirect(`${route('/login')}?${searchParams.toString()}`)
      }
  }

  return session.json()
}

/**
 * hasUserSession allows determining whether the user has a valid kratos session cookie.
 * @param request Request received in a loader function.
 * @returns boolean - if the user has a session cookie.
 */
export function hasUserSession(request: Request): boolean {
  return String(request.headers.get('cookie')).includes('ory_kratos_session')
}

/**
 * requireNoUserSession  will ensure the user doesn't already have a session.
 * @param request Request received in a loader function.
 * @returns void
 */
export async function requireNoUserSession(request: Request): Promise<void> {
  // Can immediately assume no session if there's no cookie
  if (!hasUserSession(request)) return

  const session = await fetch(`${KRATOS_URL}/sessions/whoami`, {
    headers: request.headers
  })

  switch (session.status) {
    // User shouldn't have session/cookies so don't catch unauthorised - 401.
    case 403:
    case 422: // Need to complete 2FA.
      throw redirect(route('/totp/challenge'))
  }

  const userSession = await session.json()
  if (typeof userSession.error == 'undefined') throw redirect(route('/'))
}

// This will only run on the server so don't need a router.
export function handleFlowError(
  flow: any,
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
  // This is the fynbos flow, for redirect purposes
  flowId?: string
): void {
  let redirectRoute = `/${flowType}`

  switch (flow.error.id) {
    case 'session_inactive':
      // The user doesn't have a session
      throw redirect(route('/login'), {
        headers: {
          'Clear-Site-Data': 'cookies'
        }
      })
    case 'session_aal2_required':
      // 2FA is enabled and enforced, but user did not perform 2fa yet!
      throw redirect(`/totp/challenge?returnTo=${redirectRoute}`)
    case 'session_already_available':
      // User is already signed in, let's redirect them home!
      throw redirect(route('/'))
    case 'session_refresh_required':
      // We need to re-authenticate to perform this action
      // NOTE: this is a last resort in case a user manualy bypasses the standard flow.
      throw redirect(`/login/challenge`)
    case 'self_service_flow_return_to_forbidden':
      // The return_to address is not allowed.
      throw redirect(redirectRoute)
    case 'self_service_flow_expired':
      // The flow expired, let's request a new one.
      throw redirect(redirectRoute, {
        headers: {
          'Clear-Site-Data': 'cookies'
        }
      })
    case 'security_csrf_violation':
      // A CSRF violation occurred. Best to just refresh the flow!
      throw redirect(redirectRoute, {
        headers: {
          'Clear-Site-Data': 'cookies'
        }
      })
    case 'security_identity_mismatch':
      // The requested item was intended for someone else. Let's request a new flow...
      throw redirect(redirectRoute, {
        headers: {
          'Clear-Site-Data': 'cookies'
        }
      })
    case 'browser_location_change_required':
      // Ory Kratos asked us to point the user to this URL.
      throw redirect(flow.error.redirect_browser_to)
  }
}

enum kratosErrorId {
  ErrorValidation = 4000000,
  ErrorValidationGeneric,
  ErrorValidationRequired,
  ErrorValidationMinLength,
  ErrorValidationInvalidFormat,
  ErrorValidationPasswordPolicyViolation,
  ErrorValidationInvalidCredentials,
  ErrorValidationDuplicateCredentials,
  ErrorValidationTOTPVerifierWrong,
  ErrorValidationIdentifierMissing,
  ErrorValidationAddressNotVerified,
  ErrorValidationNoTOTPDevice,
  ErrorValidationLookupAlreadyUsed,
  ErrorValidationNoWebAuthnDevice,
  ErrorValidationNoLookup,
  ErrorValidationSuchNoWebAuthnUser,
  ErrorValidationLookupInvalid,
  ErrorValidationLogin = 4010000,
  ErrorValidationLoginFlowExpired,
  ErrorValidationLoginNoStrategyFound,
  ErrorValidationRegistrationNoStrategyFound,
  ErrorValidationSettingsNoStrategyFound,
  ErrorValidationRecoveryNoStrategyFound,
  ErrorValidationVerificationNoStrategyFound,
  ErrorValidationRegistration = 4040000,
  ErrorValidationRegistrationFlowExpired,
  ErrorValidationSettings = 4050000,
  ErrorValidationSettingsFlowExpired,
  ErrorValidationRecovery = 4060000,
  ErrorValidationRecoveryRetrySuccess,
  ErrorValidationRecoveryStateFailure,
  ErrorValidationRecoveryMissingRecoveryToken,
  ErrorValidationRecoveryTokenInvalidOrAlreadyUsed,
  ErrorValidationRecoveryFlowExpired,
  ErrorValidationRecoveryCodeInvalidOrAlreadyUsed,
  ErrorValidationVerification = 4070000,
  ErrorValidationVerificationTokenInvalidOrAlreadyUsed,
  ErrorValidationVerificationRetrySuccess,
  ErrorValidationVerificationStateFailure,
  ErrorValidationVerificationMissingVerificationToken,
  ErrorValidationVerificationFlowExpired,
  ErrorValidationVerificationCodeInvalidOrAlreadyUsed,
  ErrorSystem = 5000000,
  ErrorSystemGeneric
}

type KratosMessage = {
  id: kratosErrorId
  text: string
}

/**
 * Overrides the kratos error messages.
 * Will pass on the kratos error message if not overridden.
 * @param error The kratos error message.
 */
export function kratosErrorMessage(error: KratosMessage): string {
  switch (error.id) {
    case kratosErrorId.ErrorValidationInvalidCredentials:
      return 'The provided credentials are invalid.'
    case kratosErrorId.ErrorValidationDuplicateCredentials:
      return 'An account with the same identifier (email, phone, username, ...) exists already.'
    default:
      return error.text
  }
}

export async function kratosErrorMapping<T extends object>(
  response: Response,
  fieldErrors: T
): Promise<T> {
  const data = await response.json()

  if (data.ui) {
    for (let node of data.ui.nodes) {
      // Field validation errors
      if (node.messages.length > 0) {
        console.error(
          'Kratos error on node attribute',
          node.attributes.name,
          ' with message',
          node.messages[0].text
        )
        Object.assign(fieldErrors, {
          [node.attributes.name]: kratosErrorMessage(node.messages[0])
        })
      }
    }
  }
  if (data.ui.messages && data.ui.messages.length > 0) {
    // form message validation errors
    // This gets rendered in a snackbar - only use one.
    Object.assign(fieldErrors, {
      form: kratosErrorMessage(data.ui.messages[0]),
      kratosErrorId: data.id,
      kratosMessagesText: data.ui.messages[0].text,
      kratosMessagesContext: data.ui.messages[0].context,
      data: JSON.stringify(data)
    })
  }
  return fieldErrors
}

export function isSessionAlreadyExitsMessage(msg: string) {
  return (
    msg.includes('refresh=true') && msg.includes('valid session was detected')
  )
}
