import { SubmitSelfServiceLoginFlowBody } from '@ory/kratos-client'
import { AxiosError } from 'axios'
import { Button, Router, Routes, TextField } from 'components'
import { useLoginFlow } from 'hooks'
import { getCsrfTokenFromFlow, handleGetFlowError, kratos } from 'lib/kratos'
import { useRouter } from 'next/router'
import React, { FC } from 'react'
import { useForm } from 'react-hook-form'

type LoginInputs = {
  email: string
  password: string
  csrf_token: string
}

const transformToFlowBody = (
  inputs: LoginInputs
): SubmitSelfServiceLoginFlowBody => {
  return {
    method: 'password',
    password_identifier: inputs.email,
    password: inputs.password,
    csrf_token: inputs.csrf_token
  }
}

export const LoginForm: FC = () => {
  const [flow, setFlow] = useLoginFlow()
  const router = useRouter()

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isValid }
  } = useForm<LoginInputs>()

  const onSubmit = (inputs: LoginInputs) =>
    router
      // On submission, add the flow ID to the URL but do not navigate. This prevents the user loosing
      // his data when she/he reloads the page.
      .push(
        {
          pathname: Routes.login,
          query: {
            flow: flow?.id
          }
        },
        undefined,
        { shallow: true }
      )
      .then(() =>
        kratos
          .submitSelfServiceLoginFlow(
            String(flow?.id),
            undefined,
            transformToFlowBody(inputs)
          )
          // We logged in successfully! Let's bring the user home.
          .then(() => {
            if (flow?.return_to) {
              window.location.href = flow?.return_to
              return
            }
            // TODO: catch expired session case. Currently just routes to organisation
            router.push(Routes.profile)
          })
          .then(() => {})
          .catch(handleGetFlowError(router, Routes.login, setFlow))
          .catch((err: AxiosError) => {
            // If the previous handler did not catch the error it's most likely a form validation error
            if (err.response?.status === 400) {
              // Yup, it is!
              setFlow(err.response?.data)
              setError('password', {
                type: 'manual',
                message: 'The provided credentials are invalid.'
              })
              return
            }

            return Promise.reject(err)
          })
      )

  // TODO: Kratos will return validation errors (e.g. password has been pwned) in the flow data.
  // our frontend will need to display these errors as returned from Kratos
  return (
    <form
      className='flex min-w-full flex-col space-y-4'
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
      <TextField
        {...register('password', {
          required: 'Password is required.',
          minLength: { value: 8, message: 'Min length is 8.' }
        })}
        id='password'
        label='Password'
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

      <div className='flex items-center justify-between'>
        <Router href={Routes.recovery} aria-label='Forgot password?'>
          <span className='text-primary'>Forgot password?</span>
        </Router>
        <Button type='submit'>Login</Button>
      </div>
    </form>
  )
}
