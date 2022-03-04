import { Configuration, V0alpha2Api } from '@ory/kratos-client'
import type {
  Session,
  UiNodeInputAttributes,
  SelfServiceLoginFlow,
  SelfServiceVerificationFlow,
  SelfServiceSettingsFlow,
  SelfServiceRecoveryFlow,
  SelfServiceRegistrationFlow
} from '@ory/kratos-client'
import { redirect } from 'remix'
import { AxiosError } from 'axios'
import { route } from 'routes-gen'

export const kratos = new V0alpha2Api(
  new Configuration({
    basePath:
      typeof window === 'undefined' ? 'http://kratos-public' : window.origin,
    baseOptions: {
      // Ensure we send credentials over CORSs
      withCredentials: true
    }
  })
)

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
 * requireUserSession allows gating loader functions that require a user to be authenticated.
 * @param request Request received in a loader function.
 * @returns the session if the user has a session or else a redirect.
 */
export async function requireUserSession(request: Request): Promise<Session> {
  const session = await fetch('http://kratos-public/sessions/whoami', {
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

  // Always redirect if the users email isn't verified
  if (
    userSession.identity.verifiable_addresses &&
    !userSession.identity.verifiable_addresses[0].verified
  ) {
    throw redirect(route('/verify'))
  }

  return userSession
}

/**
 * requireNoUserSession  will ensure the user doesn't already have a session.
 * @param request Request received in a loader function.
 * @returns void
 */
export async function requireNoUserSession(request: Request): Promise<void> {
  const session = await fetch('http://kratos-public/sessions/whoami', {
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
  flowType: 'login' | 'signup' | 'settings' | 'recovery' | 'verify'
) {
  return async (err: AxiosError) => {
    switch (err.response?.data.error?.id) {
      case 'session_aal2_required':
        // 2FA is enabled and enforced, but user did not perform 2fa yet!
        throw redirect(err.response?.data.redirect_browser_to)
      case 'session_already_available':
        // User is already signed in, let's redirect them home!
        throw redirect(route('/'))
      case 'session_refresh_required':
        // We need to re-authenticate to perform this action
        throw redirect(err.response?.data.redirect_browser_to)
      case 'self_service_flow_return_to_forbidden':
        // The return_to address is not allowed.
        throw redirect('/' + flowType)
      case 'self_service_flow_expired':
        // The flow expired, let's request a new one.
        throw redirect('/' + flowType)
      case 'security_csrf_violation':
        // A CSRF violation occurred. Best to just refresh the flow!
        throw redirect('/' + flowType)
      case 'security_identity_mismatch':
        // The requested item was intended for someone else. Let's request a new flow...
        throw redirect('/' + flowType)
      case 'browser_location_change_required':
        // Ory Kratos asked us to point the user to this URL.
        throw redirect(err.response.data.redirect_browser_to)
    }

    switch (err.response?.status) {
      case 410:
        // The flow expired, let's request a new one.
        throw redirect('/' + flowType)
    }

    // We are not able to handle the error? Return it.
    return Promise.reject(err)
  }
}
