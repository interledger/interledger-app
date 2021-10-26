import { FunctionComponent } from 'react'
import { useForm } from 'react-hook-form'
import { useRouter } from 'next/router'
import { AxiosError } from 'axios'
import { SubmitSelfServiceSettingsFlowBody } from '@ory/kratos-client'
import { getCsrfTokenFromFlow, handleGetFlowError, kratos } from 'lib/kratos'
import { useProfileSettingsFlow } from 'hooks'
import { Routes } from 'components'

type ProfileInputs = {
  password: string
  csrf_token: string
}

const transformToFlowBody = ({
  csrf_token,
  password
}: ProfileInputs): SubmitSelfServiceSettingsFlowBody => {
  // TODO: only doing password for now
  // webauthn/totp etc is being released in 0.8.0-alpha.1
  return {
    csrf_token,
    password,
    method: 'password'
  }
}

export const ProfileForm: FunctionComponent = () => {
  const [flow, setFlow] = useProfileSettingsFlow()
  const { register, handleSubmit } = useForm<ProfileInputs>()
  const router = useRouter()

  const onSubmit = (values: ProfileInputs) =>
    router
      // On submission, add the flow ID to the URL but do not navigate. This prevents the user loosing
      // his data when she/he reloads the page.
      .push({
        pathname: Routes.profile,
        query: {
          flow: flow?.id,
        }
      }, undefined, { shallow: true })
      .then(() =>
        kratos
          .submitSelfServiceSettingsFlow(
            String(flow?.id),
            undefined,
            transformToFlowBody(values)
          )
          .then(({ data }) => {
            // The settings have been saved and the flow was updated. Let's show it to the user!
            setFlow(data)
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
    <form onSubmit={handleSubmit(onSubmit)}>
      <label>New password</label>
      <input
        {...register('password', { required: true })}
        type='text'
        className='border-2'
      />

      <input
        {...register('csrf_token', { value: getCsrfTokenFromFlow(flow) })}
        type='hidden'
      />

      <input type='submit' value='Send link' />
    </form>
  )
}
