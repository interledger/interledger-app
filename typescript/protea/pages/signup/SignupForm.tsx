import { SubmitSelfServiceRegistrationFlowBody } from '@ory/kratos-client'
import { AxiosError } from 'axios'
import { useRouter } from 'next/router'
import React, { FC } from 'react'
import { useForm } from 'react-hook-form'
import { useSignupFlow } from 'hooks'
import { getCsrfTokenFromFlow, handleGetFlowError, kratos } from 'lib/kratos'
import { Button, Routes, TextField } from 'components'

type SignupInputs = {
  email: string
  password: string
  csrf_token: string
}

const transformToFlowBody = ({
  email,
  password,
  csrf_token
}: SignupInputs): SubmitSelfServiceRegistrationFlowBody => {
  return {
    method: 'password',
    password: password,
    csrf_token: csrf_token,
    traits: {
      email: email
    }
  }
}

export const SignupForm: FC = () => {
  const [flow, setFlow] = useSignupFlow()
  const router = useRouter()

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isValid }
  } = useForm<SignupInputs>()

  const onSubmit = (values: SignupInputs) =>
    router
      // On submission, add the flow ID to the URL but do not navigate. This prevents the user loosing
      // his data when she/he reloads the page.
      .push(
        {
          pathname: Routes.signup,
          query: {
            flow: flow?.id
          }
        },
        undefined,
        { shallow: true }
      )
      .then(() =>
        kratos
          .submitSelfServiceRegistrationFlow(
            String(flow?.id),
            transformToFlowBody(values)
          )
          .then(({ data }) => {
            // If we ended up here, it means we are successfully signed up!
            //
            // You can do cool stuff here, like having access to the identity which just signed up:
            console.log('This is the user session: ', data, data.identity)

            // TODO: catch expired session case. Currently just routes to organisation
            // For now however we just want to redirect to organisation page
            return router.push(flow?.return_to || Routes.profile)
          })
          .catch(handleGetFlowError(router, Routes.signup, setFlow))
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
      className='flex min-w-full flex-col items-end space-y-4'
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

      <Button type='submit'>Create account</Button>
    </form>
  )
}
