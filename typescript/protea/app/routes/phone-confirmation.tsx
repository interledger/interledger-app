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
  OutlineButtonRouter,
  TextButton,
  TextField
} from '~/components'

import { Label } from '~/components/Label'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { getSessionTraits, getUserSession } from '~/lib/kratos/session.server'
import { mergeMeta } from '~/lib/meta'
import {
  handleResendOtp,
  handleUpdatePhone,
  handleVerifyOtp
} from '~/lib/phone.server'
import { safeReturnTo } from '~/lib/url.server'
import { formatCountdown, useCountdown } from '~/lib/useCountdown'
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
  const resendFetcher =
    useFetcher<Awaited<ReturnType<typeof handleResendOtp>>>()
  const updateFetcher =
    useFetcher<Awaited<ReturnType<typeof handleUpdatePhone>>>()
  const { start, isActive, remainingSeconds } = useCountdown()
  const [otpSent, setOtpSent] = useState(false)
  const [currentPhone, setCurrentPhone] = useState(phone)
  const [showChangePhone, setShowChangePhone] = useState(false)
  const otpError =
    actionData && 'errors' in actionData
      ? (actionData.errors as { otp?: string } | undefined)?.otp
      : undefined
  const resendError =
    resendFetcher.data && 'message' in resendFetcher.data
      ? resendFetcher.data.message
      : undefined

  // Start countdown after a successful send/resend or a rate-limit response.
  useEffect(() => {
    if (resendFetcher.data?.codeSent) {
      setOtpSent(true)
      start(RESEND_DELAY)
      return
    }

    if (
      resendFetcher.data?.error === 'rateLimited' &&
      typeof resendFetcher.data.retryAfter === 'number'
    ) {
      setOtpSent(true)
      start(resendFetcher.data.retryAfter * 1000)
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

  const isResendDisabled = isActive || resendFetcher.state !== 'idle'
  const actionPath = `/phone-confirmation?returnTo=${encodeURIComponent(
    returnTo
  )}`

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
        <resendFetcher.Form method='post' action={actionPath} className='mt-2'>
          <input type='hidden' name='intent' value='resend' />
          <input type='hidden' name='csrfToken' value={csrfToken} />
          <input type='hidden' name='phone' value={currentPhone} />
          <Button type='submit' disabled={isResendDisabled} className='w-full'>
            {resendFetcher.state !== 'idle'
              ? 'Sending...'
              : isActive
                ? `Resend in ${formatCountdown(remainingSeconds)}`
                : otpSent
                  ? 'Resend code'
                  : 'Send code'}
          </Button>
        </resendFetcher.Form>
      )}
      {!showChangePhone && resendError && (
        <p className='mt-2 text-sm text-error'>{resendError}</p>
      )}
      <OutlineButtonRouter to={href('/logout')} className='mt-4'>
        Log out
      </OutlineButtonRouter>
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
    return handleUpdatePhone(request, form)
  }

  if (intent === 'resend') {
    const resendPhone = (form.get('phone') as string) || phone
    return await handleResendOtp(request, resendPhone)
  }

  // intent === 'verify'
  const otp = form.get('otp') as string
  const result = await handleVerifyOtp(request, otp)
  if (result !== null) return result

  throw redirect(returnTo)
}
