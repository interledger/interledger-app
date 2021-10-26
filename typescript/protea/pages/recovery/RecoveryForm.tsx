import React, { FC, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useRouter } from 'next/router'
import { AxiosError } from 'axios'
import { SubmitSelfServiceRecoveryFlowBody } from '@ory/kratos-client'
import { getCsrfTokenFromFlow, handleGetFlowError, kratos } from 'lib/kratos'
import { useRecoveryFlow } from 'hooks/useRecoveryFlow'
import { Button, Routes, TextField, Notification } from 'components'

type RecoveryInputs = {
  email: string
  csrf_token: string
}

const transformToFlowBody = ({
  csrf_token,
  email
}: RecoveryInputs): SubmitSelfServiceRecoveryFlowBody => {
  return {
    csrf_token,
    email,
    method: 'link'
  }
}

export const RecoveryForm: FC = () => {
  const [showNotification, setShowNotification] = useState<boolean>(false)
  const [flow, setFlow] = useRecoveryFlow()
  const router = useRouter()

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isValid }
  } = useForm<RecoveryInputs>()

  const onSubmit = (values: RecoveryInputs) =>
    router
      // On submission, add the flow ID to the URL but do not navigate. This prevents the user loosing
      // his data when she/he reloads the page.
      .push(
        {
          pathname: Routes.recovery,
          query: {
            flow: flow?.id
          }
        },
        undefined,
        { shallow: true }
      )
      .then(() =>
        kratos
          .submitSelfServiceRecoveryFlow(
            String(flow?.id),
            undefined,
            transformToFlowBody(values)
          )
          .then(({ data }) => {
            // Form submission was successful, show the message to the user!
            setShowNotification(true)
            setFlow(data)
          })
          .catch(handleGetFlowError(router, Routes.recovery, setFlow))
          .catch((err: AxiosError) => {
            switch (err.response?.status) {
              case 400:
                // Status code 400 implies the form validation had an error
                setFlow(err.response?.data)
                return
            }

            throw err
          })
      )

  // TODO: Kratos will return error / success messages that our ui will need to display.
  // this needs to be dug out of the flow object
  return (
    <>
      <form
        className='flex flex-col min-w-full space-y-4 items-end'
        onSubmit={handleSubmit(onSubmit)}
      >
        <TextField
          {...register('email', { required: 'Email is required.' })}
          id='email'
          label='Email'
          type='email'
          isValid={isValid}
          errorMessage={errors.email?.message}
        />

        {/* flow might still be loading when page loads */}
        {flow && (
          <input
            {...register('csrf_token', { value: getCsrfTokenFromFlow(flow) })}
            type='hidden'
          />
        )}

        <Button disabled={showNotification} type='submit'>
          Recover account
        </Button>
      </form>
      <Notification
        show={showNotification}
        setShow={setShowNotification}
        header='Recovery email sent'
        body='View email to complete recovery.'
        action='Done'
      />
    </>
  )
}
