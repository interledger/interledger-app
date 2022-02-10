import { SelfServiceVerificationFlow } from '@ory/kratos-client'
import { useRouter } from 'next/router'
import { Dispatch, SetStateAction, useEffect, useState } from 'react'
import { kratos } from 'lib/kratos'
import { AxiosError } from 'axios'
import { Routes } from 'components'

export const useVerifyFlow = (): [
  SelfServiceVerificationFlow | undefined,
  Dispatch<SetStateAction<SelfServiceVerificationFlow | undefined>>
] => {
  const [flow, setFlow] = useState<SelfServiceVerificationFlow>()
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
        .getSelfServiceVerificationFlow(String(flowId))
        .then(({ data }) => {
          setFlow(data)
        })
        .catch((err: AxiosError) => {
          switch (err.response?.status) {
            case 410:
            // Status code 410 means the request has expired - so let's load a fresh flow!
            case 403:
              // Status code 403 implies some other issue (e.g. CSRF) - let's reload!
              return router.push(Routes.verify)
          }

          throw err
        })
      return
    }

    // Otherwise we initialize it
    kratos
      .initializeSelfServiceVerificationFlowForBrowsers(
        returnTo ? String(returnTo) : undefined
      )
      .then(({ data }) => {
        setFlow(data)
      })
      .catch((err: AxiosError) => {
        switch (err.response?.status) {
          case 400:
            // Status code 400 implies the user is already signed in
            return router.push(Routes.home)
        }

        throw err
      })
  }, [flowId, router, router.isReady, returnTo, flow])

  return [flow, setFlow]
}
