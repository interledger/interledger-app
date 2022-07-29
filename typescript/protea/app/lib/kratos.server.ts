import type {
  Session,
  UiNodeInputAttributes,
  SelfServiceLoginFlow,
  SelfServiceVerificationFlow,
  SelfServiceSettingsFlow,
  SelfServiceRecoveryFlow,
  SelfServiceRegistrationFlow
} from '@ory/kratos-client'
import { redirect } from '@remix-run/node'
import { route } from 'routes-gen'

// Export to ensure this is always evaluated server side.
export const KRATOS_URL = process.env.KRATOS_URL

export const getCsrfTokenFromFlow = (
  flow:
    | SelfServiceRegistrationFlow
    | SelfServiceLoginFlow
    | SelfServiceVerificationFlow
    | SelfServiceSettingsFlow
    | SelfServiceRecoveryFlow
    | undefined
): string => {
  const node = flow?.ui.nodes.find(
    (node) => (node.attributes as UiNodeInputAttributes).name === 'csrf_token'
  )

  return node ? (node.attributes as UiNodeInputAttributes).value : ''
}

/**
 * hasUserSession allows determining whether there is a user, but not gate them.
 * requireUserSession should be preferred where gating is required.
 * @param request Request received in a loader function.
 * @returns boolean - if the user has a session.
 */
export async function hasUserSession(request: Request): Promise<boolean> {
  const session = await fetch(`${KRATOS_URL}/sessions/whoami`, {
    headers: request.headers
  })

  return session.status == 200
}

/**
 * requireUserSession allows gating loader functions that require a user to be authenticated.
 * @param request Request received in a loader function.
 * @returns the session if the user has a session or else a redirect.
 */
export async function requireUserSession(request: Request): Promise<Session> {
  const session = await fetch(`${KRATOS_URL}/sessions/whoami`, {
    headers: request.headers
  })

  switch (session.status) {
    case 401:
    case 500:
      throw redirect(route('/login'))
    case 403:
    case 422: // Need to complete 2FA.
      throw redirect(route('/login') + '?aal=aal2')
  }

  const userSession: Session = await session.json()

  //TODO: check if has provider account
  // if not call onboarding and redirect to form

  return userSession
}

/**
 * requireNoUserSession  will ensure the user doesn't already have a session.
 * @param request Request received in a loader function.
 * @returns void
 */
export async function requireNoUserSession(request: Request): Promise<void> {
  const session = await fetch(`${KRATOS_URL}/sessions/whoami`, {
    headers: request.headers
  })

  switch (session.status) {
    // User shouldn't have session/cookies so don't catch unauthorised - 401.
    case 403:
    case 422: // Need to complete 2FA.
      throw redirect(route('/login') + '?aal=aal2')
  }

  const userSession = await session.json()
  if (typeof userSession.error == 'undefined') throw redirect(route('/home'))
}

// This will only run on the server so don't need a router.
export function handleFlowError(
  flow: any,
  flowType:
    | 'login'
    | 'signup'
    | 'settings'
    | 'settings/password'
    | 'login/challenge'
    | 'recovery'
    | 'recovery/password'
    | 'verify'
    | 'logout',
  // This is the fynbos flow, for redirect purposes
  flowId?: string
): void {
  let redirectRoute =
    flowType == 'signup' ? `/flows/${flowId}/signup/password` : `/${flowType}`

  switch (flow.error.id) {
    case 'session_inactive':
      // The user doesn't have a session
      throw redirect(route('/login'))
    case 'session_aal2_required':
      // 2FA is enabled and enforced, but user did not perform 2fa yet!
      throw redirect(flow.error.redirect_browser_to)
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
      throw redirect(redirectRoute)
    case 'security_csrf_violation':
      // A CSRF violation occurred. Best to just refresh the flow!
      throw redirect(redirectRoute)
    case 'security_identity_mismatch':
      // The requested item was intended for someone else. Let's request a new flow...
      throw redirect(redirectRoute)
    case 'browser_location_change_required':
      // Ory Kratos asked us to point the user to this URL.
      throw redirect(flow.error.redirect_browser_to)
  }
}
