import { SelfServiceSettingsFlow } from '@ory/kratos-client'
import { Dispatch, SetStateAction, useEffect, useState } from 'react'
import { handleGetFlowError, kratos } from 'lib/kratos'
import { useRouter } from 'next/router'
import { Routes } from 'components'

export const useProfileSettingsFlow = (): [
  SelfServiceSettingsFlow | undefined,
  Dispatch<SetStateAction<SelfServiceSettingsFlow | undefined>>
] => {
  const [flow, setFlow] = useState<SelfServiceSettingsFlow>()

  // Get ?flow=... from the URL
  const router = useRouter()
  const { flow: flowId, return_to: returnTo } = router.query

  useEffect(() => {
    // If the router is not ready yet, or we already have a flow, do nothing.
    if (!router.isReady || flow) {
      return
    }

    // If ?flow=.. was in the URL, we fetch it
    if (flowId) {
      kratos
        .getSelfServiceSettingsFlow(String(flowId))
        .then(({ data }) => {
          setFlow(data)
        })
        .catch(handleGetFlowError(router, Routes.settings, setFlow))
      return
    }

    // Otherwise we initialize it
    kratos
      .initializeSelfServiceSettingsFlowForBrowsers(
        returnTo ? String(returnTo) : undefined
      )
      .then(({ data }) => {
        setFlow(data)
      })
      .catch(handleGetFlowError(router, Routes.settings, setFlow))
  }, [flowId, router, router.isReady, returnTo, flow])

  return [flow, setFlow]
}
