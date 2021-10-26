import { SelfServiceRecoveryFlow } from '@ory/kratos-client'
import { Dispatch, SetStateAction, useEffect, useState } from 'react'
import { AxiosError } from 'axios'
import { handleGetFlowError, kratos } from 'lib/kratos'
import { useRouter } from 'next/router'
import { Routes } from 'components'

export const useRecoveryFlow = (): [
  SelfServiceRecoveryFlow | undefined,
  Dispatch<SetStateAction<SelfServiceRecoveryFlow | undefined>>
] => {
  const [flow, setFlow] = useState<SelfServiceRecoveryFlow>()

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
        .getSelfServiceRecoveryFlow(String(flowId))
        .then(({ data }) => {
          setFlow(data)
        })
        .catch(handleGetFlowError(router, Routes.recovery, setFlow))
      return
    }

    // Otherwise we initialize it
    kratos
      .initializeSelfServiceRecoveryFlowForBrowsers()
      .then(({ data }) => {
        setFlow(data)
      })
      .catch(handleGetFlowError(router, Routes.recovery, setFlow))
      .catch((err: AxiosError) => {
        // If the previous handler did not catch the error it's most likely a form validation error
        if (err.response?.status === 400) {
          // Yup, it is!
          setFlow(err.response?.data)
          return
        }

        return Promise.reject(err)
      })
  }, [flowId, router, router.isReady, returnTo, flow])

  return [flow, setFlow]
}
