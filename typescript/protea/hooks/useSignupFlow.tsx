import { SelfServiceRegistrationFlow } from '@ory/kratos-client'
import { Dispatch, SetStateAction, useEffect, useState } from 'react'
import { handleGetFlowError, kratos } from 'lib/kratos'
import { useRouter } from 'next/router'
import { Routes } from 'components'

export const useSignupFlow = (): [
  SelfServiceRegistrationFlow | undefined,
  Dispatch<SetStateAction<SelfServiceRegistrationFlow | undefined>>
] => {
  const router = useRouter()

  // The "flow" represents a registration process and contains
  // information about the form we need to render (e.g. username + password)
  const [flow, setFlow] = useState<SelfServiceRegistrationFlow>()

  // Get ?flow=... from the URL
  const { flow: flowId, return_to: returnTo } = router.query

  // In this effect we either initiate a new registration flow, or we fetch an existing registration flow.
  useEffect(() => {
    // If the router is not ready yet, or we already have a flow, do nothing.
    if (!router.isReady || flow) {
      return
    }

    // If ?flow=.. was in the URL, we fetch it
    if (flowId) {
      kratos
        .getSelfServiceRegistrationFlow(String(flowId))
        .then(({ data }) => {
          // We received the flow - let's use its data and render the form!
          setFlow(data)
        })
        .catch(handleGetFlowError(router, Routes.signup, setFlow))
      return
    }

    // Otherwise we initialize it
    kratos
      .initializeSelfServiceRegistrationFlowForBrowsers(
        returnTo ? String(returnTo) : undefined
      )
      .then(({ data }) => {
        setFlow(data)
      })
      .catch(handleGetFlowError(router, Routes.signup, setFlow))
  }, [flowId, router, router.isReady, returnTo, flow])

  return [flow, setFlow]
}
