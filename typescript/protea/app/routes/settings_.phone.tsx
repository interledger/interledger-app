import type { Route } from './+types/settings_.phone'
import { Code } from '@bufbuild/connect'
import { redirect } from 'react-router';
import type { ShouldRevalidateFunction } from 'react-router';
import { Form, useActionData, useFetcher, useLoaderData } from 'react-router';
import { useEffect, useState } from 'react'
import { href } from 'react-router'
import type { ApplicationProps, PhoneAutocompleteOptions } from '~/components'
import {
  Button,
  Card,
  CardContent,
  CardHeader,
  Dialog,
  Icon,
  Layouts,
  PhoneTextField,
  TextButton,
  TextField
} from '~/components'
import { Label } from '~/components/Label'
import { jsonWithCSRF } from '~/lib/csrf.server'
import { error, isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  getUserSession,
  handleFlowError,
  kratosErrorMapping
} from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import type { action as sendOtpAction } from '~/routes/api_.sendOtp'
import styles from '~/styles/flags.css?url'

// The loader generates a new 3ds session. This must only be called on initial page load
// and not after submitting actions.
export const shouldRevalidate: ShouldRevalidateFunction = ({
  defaultShouldRevalidate,
  currentUrl
}) => {
  // don't initialise a new 3DS session.
  if (currentUrl.searchParams.has('flow')) {
    return false
  }

  return defaultShouldRevalidate
}

export async function loader({ request }: Route.LoaderArgs) {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const cookie = String(request.headers.get('cookie'))
  const session = await getUserSession(request)

  let flow
  if (flowId) {
    // If ?flow=.. was in the URL, we fetch it
    const flowRes = await fetch(
      `${KRATOS_URL}/self-service/settings/flows?id=${flowId}`,
      {
        headers: {
          cookie: cookie,
          Accept: 'application/json'
        }
      }
    )
    flow = await flowRes.json()
    if (flowRes.status >= 400) handleFlowError(flow, 'settings/password')
  } else {
    // If we don't have a flow, the user hasn't confirmed their old phone yet
    throw redirect(href('/otp/challenge'))
  }

  let response = await grpc.getCountries(request, {})

  if (isConnectError(response)) throw response.errorResponse

  const countries = response.countries

  return jsonWithCSRF(request, {
    flow,
    countryCode: session.identity.traits.countryCode,
    countries,
    csrf_token: getCsrfTokenFromFlow(flow)
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: href('/settings'),
      title: 'Set new mobile number'
    }
  }
}

export const meta = mergeMeta(() => [
  {
    title: 'Set new mobile number'
  }
])

export function links() {
  return [{ rel: 'stylesheet', href: styles }]
}

export default function Page() {
  const otpFetcher = useFetcher<typeof sendOtpAction>()
  const actionData = useActionData()
  const { flow, countryCode, countries, csrfToken, csrf_token } =
    useLoaderData()
  const [showDialog, setShowDialog] = useState<boolean>(false)

  useEffect(() => {
    if (
      !showDialog &&
      otpFetcher.state == 'loading' &&
      otpFetcher?.data?.success
    ) {
      setShowDialog(true)
    }
  }, [otpFetcher?.data, otpFetcher.state, showDialog])

  return (
    <>
      <otpFetcher.Form
        id='settings-phone-otp'
        action={href('/api/sendOtp')}
        method='post'
        className='hidden'
      />
      <input
        form='settings-phone-otp'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      <Form
        id='settings-phone'
        action={`/settings/phone?flow=${flow.id}`}
        method='post'
        className='hidden'
      />
      <input
        form='settings-phone'
        defaultValue={csrf_token}
        name='csrf_token'
        type='hidden'
      />
      <Card>
        <CardContent>
          <p>Set a new phone number to continue.</p>
        </CardContent>
        <PhoneTextField
          id='phone'
          form={showDialog ? 'settings-phone' : 'settings-phone-otp'}
          name='phone'
          defaultCountry={countryCode}
          options={countries as PhoneAutocompleteOptions[]}
          label='Mobile number'
          className='mt-2'
          aria-invalid={Boolean(otpFetcher.data?.errors?.phone) || undefined}
          aria-describedby={
            otpFetcher.data?.errors?.phone ? 'phone-error' : undefined
          }
          errorMessage={otpFetcher.data?.errors?.phone}
        />
      </Card>
      <Button form='settings-phone-otp' type='submit'>
        Continue
      </Button>
      <Dialog open={showDialog} setOpen={setShowDialog}>
        <CardHeader>
          <h1 className='text-xl font-medium'>Two-step verification</h1>
        </CardHeader>
        <CardContent>
          <span className='text-medium'>
            Enter the six digit code sent to your mobile number.
          </span>
        </CardContent>
        <Label className='mt-2'>Your mobile phone number</Label>
        <div className='mt-1 flex space-x-2 rounded-xl bg-nav p-3 text-medium'>
          <Icon>phone_android</Icon>
          <span>{otpFetcher?.data?.phone}</span>
        </div>

        <input
          form='signup-phone-otp-validation'
          value={otpFetcher?.data?.phone}
          name='phone'
          type='hidden'
        />
        <TextField
          id='otp'
          form='signup-phone-otp-validation'
          label='Verification code'
          name='otp'
          type='number'
          className='mt-4'
          aria-invalid={Boolean(actionData?.errors?.otp) || undefined}
          aria-describedby={actionData?.errors?.otp ? 'email-error' : undefined}
          required
          errorMessage={actionData?.errors?.otp}
        />
        <CardContent className='mt-2 flex w-full justify-end space-x-6'>
          <TextButton type='submit' form='settings-phone-otp'>
            Resend code
          </TextButton>
          <TextButton type='submit' form='settings-phone'>
            Verify
          </TextButton>
        </CardContent>
      </Dialog>
    </>
  )
}

export async function action({ request }: Route.ActionArgs) {
  const session = await getUserSession(request)
  const cookie = request.headers.get('Cookie') as string
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')

  const form = await request.formData()
  const csrfToken = form.get('csrf_token') as string
  const phone = form.get('phone') as string
  const otp = form.get('otp') as string

  const errors = {
    form: '',
    phone: '',
    otp: ''
  }
  const mapping = {
    phone: 'To'
  }

  const response = grpc.checkPhoneVerification(request, {
    to: phone,
    otp
  })

  if (isConnectError(response)) {
    if (response.code == Code.InvalidArgument) {
      return response.error({ errors }, mapping)
    } else
      return response.error({ errors }, mapping, { action: 'Contact support' })
  }

  const res = await fetch(
    `${KRATOS_URL}/self-service/settings?flow=${flowId}`,
    {
      method: 'POST',
      body: JSON.stringify({
        method: 'profile',
        traits: {
          ...session.identity.traits,
          phone
        },
        csrf_token: csrfToken
      }),
      headers: {
        'Content-type': 'application/json',
        Accept: 'application/json',
        cookie
      }
    }
  )
  const data = await res.json()

  if (res.status > 400) handleFlowError(data, 'settings/phone')
  if (res.status == 400) {
    const errs = await kratosErrorMapping(res, errors)
    return error(request, { errors: errs })
  }

  return redirectWithSnackbar(request, href('/settings/profile-contact'), {
    message: 'New mobile number successfully saved.',
    icon: 'close'
  })
}
