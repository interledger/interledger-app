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
import { Button, Card, CardContent, Layouts } from '~/components'
import { trimHeaders } from '~/lib/headers.server'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError
} from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'
import { useCountdown } from '~/lib/useCountdown'
import { useDebounceAction } from '~/lib/useDebounceAction'

export async function loader({ request }: LoaderFunctionArgs) {
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
      throw redirect(route('/login') + '?aal=aal2')
  }

  const userSession: Session = await session.json()
  if (session.status >= 400) handleFlowError(session, 'verify')

  // Check the user has at least one verifiable address.
  if (!userSession.identity.verifiable_addresses)
    return redirect(route('/signup'))
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
  const fetcher = useFetcher()

  const withDebounce = useDebounceAction(60000)
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

      start(60000)
    })
  }

  const isDisabled = isActive || fetcher.state !== 'idle'

  return (
    <>
      <Card>
        <CardContent>
          <span>
            We've sent a verification link to your email: <br /> {email}
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
    </>
  )
}

export async function action({ request }: ActionFunctionArgs) {
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
    if (verificationResponse.status >= 400) {
      throw json(
        { title: "Could't send email verification" },
        {
          status: verificationResponse.status,
          statusText: verificationResponse.statusText
        }
      )
    }
  } catch (error) {
    console.error('❌ Error sending email verification', error)
    // throw json(
    //   { title: "Could't send email verification" },
    //   { status: 500, statusText: 'Internal server error' }
    // )
  }

  // if (verificationResponse.status >= 400) {
  //   throw json(
  //     { title: "Could't send email verification" },
  //     {
  //       status: verificationResponse.status,
  //       statusText: verificationResponse.statusText
  //     }
  //   )
  // }

  console.log('✅ Sending email verification success')
  return json(
    { success: true }
    // { headers: trimHeaders(verificationResponse.headers, ['set-cookie']) }
  )
}
