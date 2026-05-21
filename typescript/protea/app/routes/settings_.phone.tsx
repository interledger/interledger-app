import { Code } from '@bufbuild/connect'
import { useEffect, useState } from 'react'
import {
  Form,
  href,
  useActionData,
  useFetcher,
  useLoaderData
} from 'react-router'
import type { ApplicationProps, PhoneAutocompleteOptions } from '~/components'
import {
  Button,
  Card,
  CardContent,
  ChangePhoneForm,
  Icon,
  Layouts,
  TextField
} from '~/components'
import { Label } from '~/components/Label'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { ErrorDescriptions } from '~/lib/error.constants'
import type { TwillioError } from '~/lib/error.mappers'
import { isConnectError, isOtpValidationError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { getSessionTraits, getUserSession } from '~/lib/kratos/session.server'
import { mergeMeta } from '~/lib/meta'
import { parseUserPhone } from '~/lib/phone.server'
import { RateLimitKeys, getKey, rateLimit } from '~/lib/rateLimit.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { useCountdown } from '~/lib/useCountdown'
import styles from '~/styles/flags.css?url'
import type { Route } from './+types/settings_.phone'

const RESEND_DELAY = 60 * 1000

export async function loader({ request }: Route.LoaderArgs) {
  const session = await getUserSession(request)
  const { countryCode, phone } = getSessionTraits(session)

  const response = await grpc.getCountries(request, {})
  if (isConnectError(response)) throw response.errorResponse

  return jsonWithCSRF(request, {
    countryCode,
    phone,
    countries: response.countries
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: href('/settings'),
      title: 'Update mobile number'
    }
  }
}

export const meta = mergeMeta(() => [{ title: 'Update mobile number' }])

export function links() {
  return [{ rel: 'stylesheet', href: styles }]
}

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { csrfToken, countryCode, phone, countries } =
    useLoaderData<typeof loader>()
  const updateFetcher = useFetcher<typeof action>()
  const resendFetcher = useFetcher()
  const { start, isActive, remainingSeconds } = useCountdown()
  const [otpSent, setOtpSent] = useState(false)
  const [newPhone, setNewPhone] = useState<string | null>(null)
  const updateCodeSent =
    updateFetcher.data &&
    'codeSent' in updateFetcher.data &&
    Boolean(updateFetcher.data.codeSent)
  const updatedPhone =
    updateFetcher.data &&
      'phone' in updateFetcher.data &&
      typeof updateFetcher.data.phone === 'string'
      ? updateFetcher.data.phone
      : undefined
  const otpError =
    actionData?.errors && 'otp' in actionData.errors
      ? actionData.errors.otp
      : undefined

  useEffect(() => {
    if (updateCodeSent && updatedPhone) {
      setNewPhone(updatedPhone)
      setOtpSent(true)
      start(RESEND_DELAY)
    }
  }, [start, updateCodeSent, updatedPhone])

  useEffect(() => {
    if (resendFetcher.data?.codeSent) {
      start(RESEND_DELAY)
    }
  }, [resendFetcher.data, start])

  const isResendDisabled =
    (otpSent && isActive) || resendFetcher.state !== 'idle'

  return (
    <>
      <Form
        id='settings-phone-verify'
        action='/settings/phone'
        method='post'
        className='hidden'
      />
      <input
        form='settings-phone-verify'
        defaultValue={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      <input
        form='settings-phone-verify'
        value='verify'
        name='intent'
        type='hidden'
      />
      <Card>
        <CardContent>
          <p>
            Enter your new mobile number. We'll send a verification code to
            confirm it.
          </p>
        </CardContent>
        {!otpSent && (
          <>
            <Label className='mt-4'>Current mobile number</Label>
            <div className='mt-1 flex space-x-2 rounded-xl bg-nav p-3 text-medium'>
              <Icon>phone_android</Icon>
              <span>{phone}</span>
            </div>
          </>
        )}
        {!otpSent && (
          <ChangePhoneForm
            fetcher={updateFetcher}
            csrfToken={csrfToken}
            defaultCountry={countryCode}
            countries={countries as PhoneAutocompleteOptions[]}
            submitLabel='Continue'
            className='mt-3'
          />
        )}
        {otpSent && newPhone && (
          <>
            <Label className='mt-4'>New mobile number</Label>
            <div className='mt-1 flex space-x-2 rounded-xl bg-nav p-3 text-medium'>
              <Icon>phone_android</Icon>
              <span>{newPhone}</span>
            </div>
            <input
              form='settings-phone-verify'
              type='hidden'
              name='phone'
              value={newPhone}
            />
            <TextField
              id='otp'
              form='settings-phone-verify'
              label='Verification code'
              name='otp'
              type='number'
              className='mt-4'
              aria-invalid={Boolean(otpError) || undefined}
              aria-describedby={otpError ? 'otp-error' : undefined}
              required
              errorMessage={otpError}
            />
          </>
        )}
      </Card>
      {otpSent && (
        <Button form='settings-phone-verify' type='submit'>
          Verify
        </Button>
      )}
      {otpSent && (
        <resendFetcher.Form
          method='post'
          action='/settings/phone'
          className='mt-2'
        >
          <input type='hidden' name='intent' value='resend' />
          <input type='hidden' name='csrfToken' value={csrfToken} />
          <input type='hidden' name='phone' value={newPhone ?? ''} />
          <Button type='submit' disabled={isResendDisabled} className='w-full'>
            {isActive ? `Resend in ${remainingSeconds}s` : 'Resend code'}
          </Button>
        </resendFetcher.Form>
      )}
    </>
  )
}

export async function action({ request }: Route.ActionArgs) {
  const form = await request.formData()
  await validateCSRFToken(request, form)

  const intent = form.get('intent') as string

  if (intent === 'updatePhone') {
    const newPhone = form.get('phone') as string
    const country = form.get('country') as string
    const parsedPhone = parseUserPhone(newPhone, country)

    if (!parsedPhone.success) {
      return { errors: { phone: parsedPhone.error } }
    }

    const updateResponse = await grpc.updateUserPhone(request, {
      phone: parsedPhone.phone
    })
    if (isConnectError(updateResponse)) {
      if (updateResponse.code === Code.InvalidArgument) {
        return { errors: { phone: 'Invalid phone number. Please check the format.' } }
      }
      throw updateResponse.errorResponse
    }

    const sendResponse = await grpc.sendPhoneVerification(request, {
      to: parsedPhone.phone
    })
    if (isConnectError(sendResponse)) throw sendResponse.errorResponse

    return { codeSent: true, phone: parsedPhone.phone }
  }

  if (intent === 'resend') {
    const phone = form.get('phone') as string
    const rateLimitError = await rateLimit(
      getKey(RateLimitKeys.PhoneOTP, phone),
      { limit: 1, ttlSeconds: 60 }
    )
    if (rateLimitError) {
      return { codeSent: false, error: 'rateLimited', retryAfter: 60 }
    }
    const response = await grpc.sendPhoneVerification(request, { to: phone })
    if (isConnectError(response)) throw response.errorResponse
    return { codeSent: true }
  }

  // intent === 'verify'
  const otp = form.get('otp') as string
  const errors: Partial<TwillioError> = { otp: '' }

  const response = await grpc.confirmUserPhone(request, { otp })

  if (isConnectError(response)) {
    if (isOtpValidationError(response)) {
      return response.error(
        { errors },
        {},
        { action: 'Contact support', message: ErrorDescriptions.INVALID_OTP }
      )
    } else if (response.code === Code.InvalidArgument) {
      return response.error({ errors })
    } else {
      return response.error(
        { errors },
        {},
        { action: 'Contact support', message: ErrorDescriptions.DEFAULT }
      )
    }
  }

  return redirectWithSnackbar(request, href('/settings/profile-contact'), {
    message: 'Mobile number updated successfully.',
    icon: 'check'
  })
}
