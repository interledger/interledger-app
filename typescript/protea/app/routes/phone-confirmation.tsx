import { Code } from '@bufbuild/connect'
import { useEffect, useState } from 'react'
import {
  Form,
  href,
  redirect,
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
  TextButton,
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
import { useCountdown } from '~/lib/useCountdown'
import { safeReturnTo } from '~/lib/url.server'
import styles from '~/styles/flags.css?url'
import type { Route } from './+types/phone-confirmation'

const RESEND_DELAY = 60 * 1000 // 1 minute

export async function loader({ request }: Route.LoaderArgs) {
  const url = new URL(request.url)
  const returnTo = safeReturnTo(url.searchParams.get('returnTo'))
  const session = await getUserSession(request, true)
  const traits = getSessionTraits(session)
  const { phone, countryCode } = traits

  if (traits.phoneVerified) {
    throw redirect(returnTo)
  }

  const countriesResponse = await grpc.getCountries(request, {})
  if (isConnectError(countriesResponse)) throw countriesResponse.errorResponse

  return jsonWithCSRF(request, {
    phone,
    countryCode,
    returnTo,
    countries: countriesResponse.countries
  })
}

export function links() {
  return [{ rel: 'stylesheet', href: styles }]
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: href('/'),
      title: 'Phone confirmation'
    }
  }
}

export const meta = mergeMeta(() => [
  {
    title: 'Phone confirmation'
  }
])

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { csrfToken, phone, countryCode, countries, returnTo } =
    useLoaderData<typeof loader>()
  const resendFetcher = useFetcher()
  const updateFetcher = useFetcher()
  const { start, isActive, remainingSeconds } = useCountdown()
  const [otpSent, setOtpSent] = useState(false)
  const [currentPhone, setCurrentPhone] = useState(phone)
  const [showChangePhone, setShowChangePhone] = useState(false)
  const otpError =
    actionData?.errors && 'otp' in actionData.errors
      ? actionData.errors.otp
      : undefined

  // Start countdown after a successful send/resend
  useEffect(() => {
    if (resendFetcher.data?.codeSent) {
      setOtpSent(true)
      start(RESEND_DELAY)
    }
  }, [resendFetcher.data, start])

  // After phone update: refresh displayed number, hide form, show OTP field
  useEffect(() => {
    if (updateFetcher.data?.codeSent) {
      setCurrentPhone(updateFetcher.data.phone)
      setShowChangePhone(false)
      setOtpSent(true)
      start(RESEND_DELAY)
    }
  }, [updateFetcher.data, start])

  const isResendDisabled =
    (otpSent && isActive) || resendFetcher.state !== 'idle'
  const actionPath = `/phone-confirmation?returnTo=${encodeURIComponent(returnTo)}`

  return (
    <>
      <Form
        id='phone-confirmation'
        action={actionPath}
        method='post'
        className='hidden'
      />
      <input
        form='phone-confirmation'
        defaultValue={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      <input
        form='phone-confirmation'
        value='verify'
        name='intent'
        type='hidden'
      />
      <Card>
        <CardContent>
          <p>Enter the six digit code sent to your mobile number.</p>
        </CardContent>
        <Label className='mt-2'>Your mobile phone number</Label>
        <div className='mt-1 flex items-center justify-between rounded-xl bg-nav p-3 text-medium'>
          <div className='flex space-x-2'>
            <Icon>phone_android</Icon>
            <span>{currentPhone}</span>
          </div>
          {!showChangePhone && (
            <TextButton
              type='button'
              className='text-sm'
              onClick={() => setShowChangePhone(true)}
            >
              Change
            </TextButton>
          )}
        </div>
        {showChangePhone && (
          <ChangePhoneForm
            fetcher={updateFetcher}
            csrfToken={csrfToken}
            defaultCountry={countryCode}
            countries={countries as PhoneAutocompleteOptions[]}
            action={actionPath}
            onCancel={() => setShowChangePhone(false)}
            className='mt-3'
          />
        )}
        {otpSent && (
          <TextField
            id='otp'
            form='phone-confirmation'
            label='Verification code'
            name='otp'
            type='number'
            className='mt-4'
            aria-invalid={Boolean(otpError) || undefined}
            aria-describedby={otpError ? 'otp-error' : undefined}
            required
            errorMessage={otpError}
          />
        )}
      </Card>
      {otpSent && (
        <Button form='phone-confirmation' type='submit'>
          Verify
        </Button>
      )}
      {!showChangePhone && (
        <resendFetcher.Form
          method='post'
          action={actionPath}
          className='mt-2'
        >
          <input type='hidden' name='intent' value='resend' />
          <input type='hidden' name='csrfToken' value={csrfToken} />
          <Button type='submit' disabled={isResendDisabled} className='w-full'>
            {!otpSent
              ? resendFetcher.state !== 'idle'
                ? 'Sending...'
                : 'Send code'
              : isActive
                ? `Resend in ${remainingSeconds}s`
                : 'Resend code'}
          </Button>
        </resendFetcher.Form>
      )}
    </>
  )
}

export async function action({ request }: Route.ActionArgs) {
  const url = new URL(request.url)
  const returnTo = safeReturnTo(url.searchParams.get('returnTo'))
  const form = await request.formData()
  const intent = form.get('intent') as string

  await validateCSRFToken(request, form)
  const session = await getUserSession(request, true)
  const { phone } = getSessionTraits(session)

  if (intent === 'updatePhone') {
    const newPhone = form.get('phone') as string
    const country = form.get('country') as string
    const parsedPhone = parseUserPhone(newPhone, country)

    if (parsedPhone.error) {
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

  const errors: Partial<TwillioError> = {
    otp: ''
  }

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

  throw redirect(returnTo)
}
