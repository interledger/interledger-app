import {
  data,
  href,
  redirect,
  useFetcher,
  useLoaderData,
  useRouteLoaderData
} from 'react-router'
import type { ApplicationProps } from '~/components'
import {
  Button,
  Card,
  CardContent,
  Layouts,
  OutlineButtonRouter
} from '~/components'
import {
  buildHeadersWithCookies,
  getCookie,
  withCookie
} from '~/lib/kratos/cookie.server'
import { handleFlowError } from '~/lib/kratos/error.server'
import { getCsrfTokenFromFlow } from '~/lib/kratos/flow.server'
import { kratosPublic } from '~/lib/kratos/kratos-client.server'
import { getUserSession } from '~/lib/kratos/session.server'
import logger from '~/lib/logger.server'
import { mergeMeta } from '~/lib/meta'
import { RateLimitKeys, getKey, rateLimit } from '~/lib/rateLimit.server'
import { safeReturnTo } from '~/lib/url.server'
import { useCountdown } from '~/lib/useCountdown'
import { useDebounceAction } from '~/lib/useDebounceAction'
import { RootLoaderData } from '~/root'
import type { Route } from './+types/verify'

const RESEND_DELAY = 60 * 1000 // 1 minute

type ActionData = {
  success: boolean
}

export async function loader({ request }: Route.LoaderArgs) {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const cookie = getCookie(request)
  const userSession = await getUserSession(request, true)

  if (!userSession) {
    // Session requires AAL2 upgrade — redirect to TOTP challenge
    throw redirect(href('/totp/challenge'))
  }

  // We currently only allow one email per user.
  if (userSession.identity?.verifiable_addresses?.[0]?.verified) {
    const searchParams = new URLSearchParams()
    searchParams.set('returnTo', safeReturnTo(url.searchParams.get('returnTo')))
    return redirect(`${href('/phone-confirmation')}?${searchParams.toString()}`)
  }

  if (flowId) {
    try {
      const { data: flow } = await kratosPublic.getVerificationFlow(
        { id: flowId },
        withCookie(cookie)
      )
      return data({
        flow,
        email: userSession.identity?.verifiable_addresses?.[0]?.value ?? '',
        csrfToken: getCsrfTokenFromFlow(flow)
      })
    } catch (err: any) {
      handleFlowError(err, 'verify')
      logger.error(
        { error: err, route: 'verify' },
        'Failed to load verification flow'
      )
      throw new Error('Failed to load verification flow')
    }
  }

  // Initialize new verification flow
  try {
    const response = await kratosPublic.createBrowserVerificationFlow(
      { returnTo: safeReturnTo(url.searchParams.get('returnTo')) },
      withCookie(cookie)
    )
    return redirect(`/verify?flow=${response.data.id}`, {
      headers: buildHeadersWithCookies(response)
    })
  } catch (err: any) {
    handleFlowError(err, 'verify')
    logger.error(
      { error: err, route: 'verify' },
      'Failed to initialize verification flow'
    )
    throw new Error('Failed to initialize verification flow')
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

export const meta = mergeMeta(() => [
  {
    title: 'Verify your email'
  }
])

export default function Page() {
  const { flow, email, csrfToken } = useLoaderData()
  const fetcher = useFetcher<ActionData>()
  const { env } = useRouteLoaderData('root') as RootLoaderData
  const supportEmail = env.supportEmail

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

      <OutlineButtonRouter to={href('/logout')} className='mt-4'>
        Log out
      </OutlineButtonRouter>

      {hasError && (
        <p className='mt-2 text-sm text-error'>
          Could not send email verification. Please try again or contact support
          at{' '}
          <a href={`mailto:${supportEmail}`}>
            <b>{supportEmail}</b>
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

export async function action({ request }: Route.ActionArgs) {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  if (!flowId) {
    throw redirect('/verify')
  }

  const form = await request.formData()
  const csrfToken = form.get('csrf_token') as string
  const email = (form.get('email') as string) ?? ''

  const key = getKey(RateLimitKeys.VerifyEmail, email)
  const rateError = await rateLimit(key)
  if (rateError) {
    return data<ActionData>(
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
        flow: flowId,
        updateVerificationFlowBody: {
          method: 'link',
          email,
          csrf_token: csrfToken
        }
      },
      withCookie(cookie)
    )
    return data<ActionData>({ success: true }, { status: 200 })
  } catch (_: any) {
    return data<ActionData>({ success: false }, { status: 400 })
  }
}
