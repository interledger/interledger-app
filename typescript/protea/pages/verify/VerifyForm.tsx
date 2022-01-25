import { SubmitSelfServiceVerificationFlowBody } from '@ory/kratos-client'
import { AxiosError } from 'axios'
import { useRouter } from 'next/router'
import React, { FC, useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useVerifyFlow } from 'hooks'
import { getCsrfTokenFromFlow, handleGetFlowError, kratos } from 'lib/kratos'
import { Button, Routes, TextField } from 'components'

type VerifyProps = {
  email: string
  createdAt: string
}

export const VerifyForm: FC<VerifyProps> = ({ email }) => {
  const [disabled, setDisabled] = useState<boolean>(true)
  const [flow, setFlow] = useVerifyFlow()
  const router = useRouter()

  useEffect(() => {
    if (disabled) {
      const timer = setTimeout(() => {
        setDisabled(false)
      }, 1000 * 60)
      return () => clearTimeout(timer)
    }
  }, [disabled])

  const { handleSubmit } = useForm()

  const onSubmit = () =>
    router
      // On submission, add the flow ID to the URL but do not navigate. This prevents the user loosing
      // his data when she/he reloads the page.
      .push(
        {
          pathname: Routes.verify,
          query: {
            flow: flow?.id
          }
        },
        undefined,
        { shallow: true }
      )
      .then(() =>
        kratos
          .submitSelfServiceVerificationFlow(String(flow?.id), undefined, {
            method: 'link',
            email: email,
            csrf_token: getCsrfTokenFromFlow(flow)
          })
          .then(({ data }) => {
            setDisabled(true)
            return router.push(flow?.return_to || Routes.profile)
          })
          .catch(handleGetFlowError(router, Routes.signup, setFlow))
          .catch((err: AxiosError) => {
            // If the previous handler did not catch the error it's most likely a form validation error
            if (err.response?.status === 400) {
              // Yup, it is!
              setFlow(err.response?.data)
              return
            }

            return Promise.reject(err)
          })
      )

  // TODO: Kratos will return validation errors (e.g. password has been pwned) in the flow data.
  // our frontend will need to display these errors as returned from Kratos
  return (
    <form
      className='flex min-w-full flex-col items-end space-y-4'
      onSubmit={handleSubmit(onSubmit)}
    >
      <Button disabled={disabled} type='submit'>
        Resend verification
      </Button>
    </form>
  )
}
