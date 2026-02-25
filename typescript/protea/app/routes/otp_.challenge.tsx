import { Code } from '@bufbuild/connect'
import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
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
import type { TwillioError} from '~/lib/error.mappers';
import { TwillioErrorMapper } from '~/lib/error.mappers'
import { isConnectError, isTwilioError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { kratosPublic } from '~/lib/kratos/kratos-client.server'
import { getCookie, withCookie, buildHeadersWithCookies } from '~/lib/kratos/cookie.util'
import { handleFlowError } from '~/lib/kratos/error.server'
import { getUserSession, getSessionTraits } from '~/lib/kratos/session.util.server'
import { mergeMeta } from '~/lib/meta'

export async function loader({ request }: LoaderFunctionArgs) {
  const session = await getUserSession(request)
  const { phone } = getSessionTraits(session)

  const len = phone.length
  const phoneMask = phone
    .substring(len - 4, len)
    .padStart(len, '*')

  let response = await grpc.sendPhoneVerification(request, {
    to: phone
  })
  if (isConnectError(response)) throw response.errorResponse

  const url = new URL(request.url)
  const returnTo = url.searchParams.get('returnTo')

  return jsonWithCSRF(request, {
    phoneMask,
    returnTo: returnTo ?? ''
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/'),
      title: 'Confirmation'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: "Confirm it's you"
  }
])

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { csrfToken, phoneMask, returnTo } = useLoaderData<typeof loader>()

  return (
    <>
      <Form
        id='otp-challenge'
        action='/otp/challenge'
        method='post'
        className='hidden'
      />
      <input
        form='otp-challenge'
        defaultValue={csrfToken}
        name='csrf_token'
        type='hidden'
      />
      <input
        form='otp-challenge'
        value={returnTo}
        name='return_to'
        type='hidden'
      />
      <Card>
        <CardContent>
          <p>Enter the six digit code sent to your current mobile number.</p>
        </CardContent>
        <Label className='mt-2'>Your mobile phone number</Label>
        <div className='mt-1 flex space-x-2 rounded-xl bg-nav p-3 text-medium'>
          <Icon>phone_android</Icon>
          <span>{phoneMask}</span>
        </div>
        <TextField
          id='otp'
          form='otp-challenge'
          label='Verification code'
          name='otp'
          type='number'
          className='mt-4'
          aria-invalid={Boolean(actionData?.errors.otp) || undefined}
          aria-describedby={actionData?.errors.otp ? 'email-error' : undefined}
          required
          errorMessage={actionData?.errors.otp}
        />
      </Card>
      <Button form='otp-challenge' type='submit'>
        Continue
      </Button>
    </>
  )
}

export async function action({ request }: ActionFunctionArgs) {
  const form = await request.formData()
  const otp = form.get('otp') as string
  const returnTo = form.get('return_to') as string

  const session = await getUserSession(request)
  const { phone } = getSessionTraits(session)

  await validateCSRFToken(request, form)

  const errors: Partial<TwillioError> = {
    otp: ''
  }

  const response = await grpc.checkPhoneVerification(request, {
    to: phone,
    otp
  })

  if (isConnectError(response)) {
    if (isTwilioError(response)) {
      const e = response.error(
        { errors },
        {
          otp: TwillioErrorMapper.otp
        },
        { action: 'Contact support', message: ErrorDescriptions.INVALID_OTP }
      )
      return e
    } else if (response.code == Code.InvalidArgument) {
      const e = response.error({ errors })
      return e
    } else {
      const e = response.error(
        { errors },
        {},
        { action: 'Contact support', message: ErrorDescriptions.DEFAULT }
      )
      return e
    }
  }

  if (returnTo) {
    return redirect(returnTo)
  }

  const cookie = getCookie(request)
  try {
    const response = await kratosPublic.createBrowserSettingsFlow(
      {},
      withCookie(cookie)
    )
    return redirect(`/settings/phone?flow=${response.data.id}`, {
      headers: buildHeadersWithCookies(response)
    })
  } catch (err: any) {
    handleFlowError(err, 'otp/challenge')
    throw err
  }
}
