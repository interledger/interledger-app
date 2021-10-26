import { SubmitSelfServiceLoginFlowBody } from '@ory/kratos-client'
import { AxiosError } from 'axios'
import { Routes } from 'components'
import { useLoginFlow } from 'hooks'
import { getCsrfTokenFromFlow, handleGetFlowError, kratos } from 'lib/kratos'
import { useRouter } from 'next/router'
import { FunctionComponent } from 'react'
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

export const LoginForm: FunctionComponent = () => {
  const [flow, setFlow] = useLoginFlow()
  const router = useRouter()
  const {
    register,
    handleSubmit,
    formState: { errors }
  } = useForm<LoginInputs>()
  const onSubmit = (inputs: LoginInputs) =>
    router
      // On submission, add the flow ID to the URL but do not navigate. This prevents the user loosing
      // his data when she/he reloads the page.
      .push({
        pathname: Routes.login,
        query: {
          flow: flow?.id,
        },
      }, undefined, { shallow: true })
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
            router.push(Routes.organisation)
          })
          .then(() => {})
          .catch(handleGetFlowError(router, Routes.login, setFlow))
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
  // flow might still be loading
  if (!flow) return null

  // TODO: Kratos will return validation errors (e.g. password has been pwned) in the flow data.
  // our frontend will need to display these errors as returned from Kratos
  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <label>Email</label>
      <input
        {...register('email', { required: true })}
        type='text'
        className='border-2'
      />

      <label>Password</label>
      <input
        {...register('password', { required: true, minLength: 6 })}
        className='border-2'
      />
      {/* errors will return when field validation fails  */}
      {errors.password && <span>This field is required</span>}

      <input
        {...register('csrf_token', { value: getCsrfTokenFromFlow(flow) })}
        type='hidden'
      />

      <input type='submit' value='Login' />
    </form>
  )
}
