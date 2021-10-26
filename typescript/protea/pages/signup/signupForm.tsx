import { SubmitSelfServiceRegistrationFlowBody } from '@ory/kratos-client'
import { AxiosError } from 'axios'
import { useRouter } from 'next/router'
import { FunctionComponent } from 'react'
import { useForm } from 'react-hook-form'
import { useSignupFlow } from 'hooks'
import { getCsrfTokenFromFlow, handleGetFlowError, kratos } from 'lib/kratos'
import { Routes } from 'components'

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

export const SignupForm: FunctionComponent = () => {
  const [flow, setFlow] = useSignupFlow()
  const router = useRouter()
  const {
    register,
    handleSubmit,
    formState: { errors }
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

            // For now however we just want to redirect to organisation page
            return router.push(flow?.return_to || Routes.organisation)
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

      <input type='submit' value='Signup' />
    </form>
  )
}
