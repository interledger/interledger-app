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
import type { ApplicationProps } from '~/components'
import {
  Button,
  Card,
  CardContent,
  Icon,
  Layouts,
  TextField
} from '~/components'
import { Label } from '~/components/Label'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { ErrorDescriptions } from '~/lib/error.constants'
import type { TwillioError } from '~/lib/error.mappers'
import { TwillioErrorMapper } from '~/lib/error.mappers'
import { isConnectError, isTwilioError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { getSessionTraits, getUserSession } from '~/lib/kratos/session.server'
import { mergeMeta } from '~/lib/meta'
import { RateLimitKeys, getKey, rateLimit } from '~/lib/rateLimit.server'
import { useCountdown } from '~/lib/useCountdown'
import type { Route } from './+types/phone-confirmation'

const RESEND_DELAY = 60 * 1000 // 1 minute

export async function loader({ request }: Route.LoaderArgs) {
  const session = await getUserSession(request)
  const traits = getSessionTraits(session)
  const { phone } = traits

  if (traits.phoneVerified) {
    throw redirect('/wallet-address')
  }

  const len = phone.length
  const phoneMask = phone.substring(len - 4, len).padStart(len, '*')

  return jsonWithCSRF(request, { phoneMask })
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
  const { csrfToken, phoneMask } = useLoaderData<typeof loader>()
  const resendFetcher = useFetcher()
  const { start, isActive, remainingSeconds } = useCountdown()
  const [otpSent, setOtpSent] = useState(false)

  // Start countdown after a successful send/resend
  useEffect(() => {
    if (resendFetcher.data?.codeSent) {
      setOtpSent(true)
      start(RESEND_DELAY)
    }
  }, [resendFetcher.data, start])

  const isResendDisabled =
    (otpSent && isActive) || resendFetcher.state !== 'idle'

  return (
    <>
      <Form
        id='phone-confirmation'
        action='/phone-confirmation'
        method='post'
        className='hidden'
      />
      <input
        form='phone-confirmation'
        defaultValue={csrfToken}
        name='csrf_token'
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
        <div className='mt-1 flex space-x-2 rounded-xl bg-nav p-3 text-medium'>
          <Icon>phone_android</Icon>
          <span>{phoneMask}</span>
        </div>
        {otpSent && (
          <TextField
            id='otp'
            form='phone-confirmation'
            label='Verification code'
            name='otp'
            type='number'
            className='mt-4'
            aria-invalid={Boolean(actionData?.errors?.otp) || undefined}
            aria-describedby={actionData?.errors?.otp ? 'otp-error' : undefined}
            required
            errorMessage={actionData?.errors?.otp}
          />
        )}
      </Card>
      {otpSent && (
        <Button form='phone-confirmation' type='submit'>
          Verify
        </Button>
      )}
      <resendFetcher.Form
        method='post'
        action='/phone-confirmation'
        className='mt-2'
      >
        <input type='hidden' name='intent' value='resend' />
        <input type='hidden' name='csrf_token' value={csrfToken} />
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
    </>
  )
}

export async function action({ request }: Route.ActionArgs) {
  const form = await request.formData()
  const intent = form.get('intent') as string

  console.log('[phone-confirmation] action called, intent:', intent)

  await validateCSRFToken(request, form)

  console.log('[phone-confirmation] CSRF passed, getting session...')
  const session = await getUserSession(request)
  console.log(
    '[phone-confirmation] session ok, traits:',
    JSON.stringify(getSessionTraits(session))
  )
  const { phone } = getSessionTraits(session)

  if (intent === 'resend') {
    console.log(
      '[phone-confirmation] resend intent, checking rate limit for phone:',
      phone
    )
    const rateLimitError = await rateLimit(
      getKey(RateLimitKeys.PhoneOTP, phone),
      { limit: 1, ttlSeconds: 60 }
    )
    if (rateLimitError) {
      console.log('[phone-confirmation] rate limited:', rateLimitError)
      return { codeSent: false, error: 'rateLimited', retryAfter: 60 }
    }

    console.log('[phone-confirmation] calling sendPhoneVerification...')
    const response = await grpc.sendPhoneVerification(request, { to: phone })
    if (isConnectError(response)) {
      console.log(
        '[phone-confirmation] sendPhoneVerification connect error:',
        response.code
      )
      throw response.errorResponse
    }

    console.log('[phone-confirmation] resend success')
    return { codeSent: true }
  }

  // intent === 'verify'
  const otp = form.get('otp') as string

  const errors: Partial<TwillioError> = {
    otp: ''
  }

  const response = await grpc.confirmUserPhone(request, { otp })

  if (isConnectError(response)) {
    if (isTwilioError(response)) {
      return response.error(
        { errors },
        {
          otp: TwillioErrorMapper.otp
        },
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

  throw redirect('/wallet-address')
}
