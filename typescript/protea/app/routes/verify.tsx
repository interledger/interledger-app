import type { Session } from '@ory/kratos-client'
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
import { trimHeaders } from '~/lib/headers.server'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError
} from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'
import { useCountdown } from '~/lib/useCountdown'
import { useDebounceAction } from '~/lib/useDebounceAction'

const RESEND_DELAY = 60 * 1000 // 1 minute

type ActionData = {
  success: boolean
}

export async function loader({ request }: LoaderFunctionArgs) {
  console.log('(verify)✅ loader, req url: ', request.url)
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const cookie = String(request.headers.get('cookie'))

  const session = await fetch(`${KRATOS_URL}/sessions/whoami`, {
    headers: request.headers
  })

  switch (session.status) {
    case 401:
    case 500:
      throw redirect(route('/login'))
    case 403:
    case 422: // Need to complete 2FA.
      console.log('(verify)✅ redirecting to login with aal2 for route', request.url)
      throw redirect(route('/login') + '?aal=aal2')
  }

  const userSession: Session = await session.json()
  if (session.status >= 400) handleFlowError(session, 'verify')

  // Check the user has at least one verifiable address.
  if (!userSession.identity.verifiable_addresses) {
    console.log('✅ user has no verifiable addresses')
    console.log('✅ userSession', userSession.identity)
    // user
  }
  // return redirect(route('/signup'))
  // We currently only allow one email per user.
  if (userSession.identity.verifiable_addresses[0].verified) {
    return redirect(route('/'))
  }

  // Ensure any redirects are thrown
  if (userSession instanceof Response) return session

  let flow
  if (flowId) {
    // If ?flow=.. was in the URL, we fetch it
    const flowRes = await fetch(
      `${KRATOS_URL}/self-service/verification/flows?id=${flowId}`,
      {
        headers: {
          cookie: cookie,
          Accept: 'application/json'
        }
      }
    )
    flow = await flowRes.json()
    if (flowRes.status >= 400) handleFlowError(flow, 'verify')
  } else {
    // Otherwise we initialize it
    const flowRes = await fetch(
      `${KRATOS_URL}/self-service/verification/browser?${url.searchParams}`,
      { headers: { cookie: cookie, Accept: 'application/json' } }
    )
    flow = await flowRes.json()
    if (flowRes.status >= 400) handleFlowError(flow, 'verify')
    return redirect(`/verify?flow=${flow.id}`, {
      headers: trimHeaders(flowRes.headers, ['set-cookie'])
    })
  }
  return json({
    flow,
    email: userSession.identity.verifiable_addresses[0].value,
    csrfToken: getCsrfTokenFromFlow(flow)
  })
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
  const email = form.get('email') as string

  let verificationResponse: Response
  try {
    verificationResponse = await fetch(
      `${KRATOS_URL}/self-service/verification?flow=${flowId}`,
      {
        method: 'POST',
        redirect: 'manual',
        body: JSON.stringify({
          method: 'link',
          email,
          csrf_token: csrfToken
        }),
        headers: {
          'Content-type': 'application/json',
          cookie: String(request.headers.get('cookie'))
        }
      }
    )
  } catch (error) {
    console.error('❌ Error sending email verification', error)
    return json<ActionData>(
      {
        success: false
      },
      { status: 500 }
    )
  }

  if (verificationResponse.status >= 400) {
    return json<ActionData>(
      {
        success: false
      },
      { status: 500 }
    )
  }

  return json<ActionData>({ success: true }, { status: 200 })
}
