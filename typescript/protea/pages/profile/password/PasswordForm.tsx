import React, { FC } from 'react'
import { useForm } from 'react-hook-form'
import { useRouter } from 'next/router'
import { AxiosError } from 'axios'
import { SubmitSelfServiceSettingsFlowBody } from '@ory/kratos-client'
import { getCsrfTokenFromFlow, handleGetFlowError, kratos } from 'lib/kratos'
import { useProfileSettingsFlow } from 'hooks'
import { Button, Routes, TextField } from 'components'

type PasswordInputs = {
  password: string
  csrf_token: string
}

const transformToFlowBody = ({
  csrf_token,
  password
}: PasswordInputs): SubmitSelfServiceSettingsFlowBody => {
  // TODO: only doing password for now
  // webauthn/totp etc is being released in 0.8.0-alpha.1
  return {
    csrf_token,
    password,
    method: 'password'
  }
}

export const PasswordForm: FC = () => {
  const [flow, setFlow] = useProfileSettingsFlow()
  const router = useRouter()

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isValid }
  } = useForm<PasswordInputs>()

  const onSubmit = (values: PasswordInputs) =>
    router
      // On submission, add the flow ID to the URL but do not navigate. This prevents the user loosing
      // his data when she/he reloads the page.
      .push(
        {
          pathname: Routes.profilePassword,
          query: {
            flow: flow?.id
          }
        },
        undefined,
        { shallow: true }
      )
      .then(() =>
        kratos
          .submitSelfServiceSettingsFlow(
            String(flow?.id),
            undefined,
            transformToFlowBody(values)
          )
          .then(({ data }) => {
            router.push(Routes.profile)
          })
          .catch(handleGetFlowError(router, Routes.profile, setFlow))
          .catch(async (err: AxiosError) => {
            // If the previous handler did not catch the error it's most likely a form validation error
            if (err.response?.status === 400) {
              // Yup, it is!
              setFlow(err.response?.data)
              return
            }

            return Promise.reject(err)
          })
      )

  // TODO: Kratos will return error / success messages that our ui will need to display.
  // this needs to be dug out of the flow object

  // TODO: webauthn/totp being released in 0.8.0-alpha.1
  return (
    <>
      <form
        className='flex flex-col min-w-full space-y-4 items-end'
        onSubmit={handleSubmit(onSubmit)}
      >
        <TextField
          {...register('password', {
            required: 'Password is required.',
            minLength: { value: 8, message: 'Min length is 8.' }
          })}
          id='password'
          label='New password'
          type='password'
          isValid={isValid}
          errorMessage={errors.password?.message}
        />

        {/* flow might still be loading when page loads */}
        {flow && (
          <input
            {...register('csrf_token', { value: getCsrfTokenFromFlow(flow) })}
            type='hidden'
          />
        )}

        <Button type='submit'>Save password</Button>
      </form>
    </>
  )
}
