import {
  Configuration,
  SelfServiceRegistrationFlow,
  SelfServiceLoginFlow,
  SelfServiceVerificationFlow,
  SelfServiceSettingsFlow,
  SelfServiceRecoveryFlow,
  UiNodeInputAttributes,
  V0alpha2Api
} from '@ory/kratos-client'
import { NextRouter } from 'next/router'
import { Dispatch, SetStateAction } from 'react'
import { AxiosError } from 'axios'
import { Routes } from 'components'

const KRATOS_URL = process.env.NEXT_PUBLIC_ORY_KRATOS_PUBLIC || "http://127.0.0.1:4433"

export const kratos = new V0alpha2Api(
  new Configuration({
    basePath: KRATOS_URL,
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

export function handleGetFlowError<S>(
  router: NextRouter,
  flowType: Routes.signup | Routes.login | Routes.verify | Routes.profile | Routes.recovery,
  resetFlow: Dispatch<SetStateAction<S | undefined>>
) {
  return async (err: AxiosError) => {
    switch (err.response?.data.error?.id) {
      case 'has_session_already':
        // User is already signed in, let's redirect them home!
        await router.push(Routes.home)
        return
      case 'forbidden_return_to':
        // The flow expired, let's request a new one.
        alert('The return_to address is not allowed.')
        resetFlow(undefined)
        await router.push(flowType)
        return
      case 'self_service_flow_expired':
        // The flow expired, let's request a new one.
        alert('Your interaction expired, please fill out the form again.')
        resetFlow(undefined)
        await router.push(flowType)
        return
      case 'csrf_violation':
        // A CSRF violation occurred. Best to just refresh the flow!
        alert(
          'A security violation was detected, please fill out the form again.'
        )
        resetFlow(undefined)
        await router.push(flowType)
        return
      case 'intended_for_someone_else':
        // The requested item was intended for someone else. Let's request a new flow...
        resetFlow(undefined)
        await router.push(flowType)
        return
      case 'browser_location_change_required':
        // Ory Kratos asked us to point the user to this URL.
        window.location.href = err.response.data.redirect_browser_to
        return
      case 'needs_privileged_session':
        // We need to re-authenticate to perform this action
        window.location.href = err.response?.data.redirect_browser_to
        return
    }

    switch (err.response?.status) {
      case 410:
        // The flow expired, let's request a new one.
        resetFlow(undefined)
        await router.push(flowType)
        return
    }

    // We are not able to handle the error? Return it.
    return Promise.reject(err)
  }
}
