import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useFetcher, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Button,
  Card,
  CardContent,
  Layouts,
  OutlineButtonRouter
} from '~/components'
import { kratosPublic } from '~/lib/kratos/kratos-client.server'
import { getCookie, withCookie, buildHeadersWithCookies } from '~/lib/kratos/cookie.util'
import { getCsrfTokenFromFlow } from '~/lib/kratos/flow.util'
import { handleFlowError } from '~/lib/kratos/error'
import { getUserSession } from '~/lib/kratos/session.util'
import { mergeMeta } from '~/lib/meta'
import { RateLimitKeys, getKey, rateLimit } from '~/lib/rateLimit.server'
import { useCountdown } from '~/lib/useCountdown'
import { useDebounceAction } from '~/lib/useDebounceAction'

const RESEND_DELAY = 60 * 1000 // 1 minute

type ActionData = {
  success: boolean
}

export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const cookie = getCookie(request)

  // getUserSession handles all error cases (401, 403, 422, 500) with appropriate redirects
  const userSession = await getUserSession(request, true)

  if (!userSession) {
    // Session requires AAL2 upgrade — redirect to TOTP challenge
    throw redirect(route('/totp/challenge'))
  }

  // We currently only allow one email per user.
  if (userSession.identity?.verifiable_addresses?.[0]?.verified) {
    return redirect(route('/'))
  }

  if (flowId) {
    try {
      const { data: flow } = await kratosPublic.getVerificationFlow(
        { id: flowId },
        withCookie(cookie)
      )
      return json({
        flow,
        email: userSession.identity?.verifiable_addresses?.[0]?.value,
        csrfToken: getCsrfTokenFromFlow(flow)
      })
    } catch (err: any) {
      handleFlowError(err, 'verify')
      throw err
    }
  }

  // Initialize new verification flow
  try {
    const response = await kratosPublic.createBrowserVerificationFlow(
      { returnTo: url.searchParams.get('returnTo') ?? undefined },
      withCookie(cookie)
    )
    return redirect(`/verify?flow=${response.data.id}`, {
      headers: buildHeadersWithCookies(response)
    })
  } catch (err: any) {
    handleFlowError(err, 'verify')
    throw err
  }
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      title: 'Verify your email'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Verify your email'
  }
])

export default function Page() {
  const { flow, email, csrfToken } = useLoaderData<typeof loader>()
  const fetcher = useFetcher<ActionData>()

  const withDebounce = useDebounceAction(RESEND_DELAY)
  const { start, isActive, remainingSeconds } = useCountdown()

  const handleResend = () => {
    withDebounce(() => {
      const formData = new FormData()
      formData.append('csrf_token', csrfToken)
      formData.append('email', email)

      fetcher.submit(formData, {
        method: 'post',
        action: `/verify?flow=${flow.id}`
      })

      start(RESEND_DELAY)
    })
  }

  const isDisabled = isActive || fetcher.state !== 'idle'

  const hasError = isActive && fetcher.data?.success === false

  const hasSuccess = isActive && fetcher.data?.success === true

  return (
    <>
      <Card>
        <CardContent>
          <span>
            We've sent a verification link to your email: <br /> <b>{email}</b>
            <br />
            <br />
            If you couldn't find it, check your spam folder or try resending.
          </span>
        </CardContent>
      </Card>
      <Button onClick={handleResend} disabled={isDisabled}>
        {isActive
          ? `Resend in ${remainingSeconds}s`
          : fetcher.state !== 'idle'
          ? 'Sending...'
          : 'Resend verification'}
      </Button>

      <OutlineButtonRouter to={route('/logout')} className='mt-4'>
        Log out
      </OutlineButtonRouter>

      {hasError && (
        <p className='mt-2 text-sm text-error'>
          Could not send email verification. Please try again or contact support
          at{' '}
          <a href='mailto:support@interledger.app'>
            <b>support@interledger.app</b>
          </a>
          .
        </p>
      )}

      {hasSuccess && (
        <p className='mt-2 text-sm text-success'>
          Verification email sent successfully.
        </p>
      )}
    </>
  )
}

export async function action({
  request
}: ActionFunctionArgs): Promise<ReturnType<typeof json<ActionData>>> {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')

  const form = await request.formData()
  const csrfToken = form.get('csrf_token') as string
  const email = (form.get('email') as string) ?? ''

  const key = getKey(RateLimitKeys.VerifyEmail, email)
  const rateError = await rateLimit(key)
  if (rateError) {
    return json<ActionData>(
      {
        success: false
      },
      { status: 500 }
    )
  }

  const cookie = getCookie(request)

  try {
    await kratosPublic.updateVerificationFlow(
      {
        flow: flowId!,
        updateVerificationFlowBody: {
          method: 'link',
          email,
          csrf_token: csrfToken
        }
      },
      withCookie(cookie)
    )
    return json<ActionData>({ success: true }, { status: 200 })
  } catch (_: any) {
    return json<ActionData>({ success: false }, { status: 400 })
  }
}
