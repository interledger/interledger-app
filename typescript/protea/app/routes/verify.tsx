import type { Session } from '@ory/kratos-client'
import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Button, Card, Layouts } from '~/components'
import { trimHeaders } from '~/lib/headers.server'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError
} from '~/lib/kratos.server'

export async function loader({ request }: LoaderArgs) {
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

export const handle = {
  title: 'Verify your email',
  layout: Layouts.Focus
}

export const meta: MetaFunction = () => {
  return {
    title: 'Verify your email'
  }
}

export default function Page() {
  const { flow, email, csrfToken } = useLoaderData<typeof loader>()

  return (
    <>
      <Form
        id='verify'
        action={`/verify?flow=${flow.id}`}
        method='post'
        className='hidden'
      />
      <Card>
        <span>
          We've sent a verification link to your email: <br /> {email}
        </span>
        <input
          form='verify'
          defaultValue={csrfToken}
          name='csrf_token'
          type='hidden'
        />
        <input form='verify' defaultValue={email} name='email' type='hidden' />
      </Card>
      <Button form='verify' type='submit'>
        Resend verification
      </Button>
    </>
  )
}

export async function action({ request }: ActionArgs) {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')

  const form = await request.formData()
  const csrfToken = form.get('csrf_token') as string
  const email = form.get('email') as string

  const res = await fetch(
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

  if (res.status >= 400) {
    throw json(
      { title: "Could't send email verification" },
      { status: res.status, statusText: res.statusText }
    )
  }
  return redirect(route('/verify'), {
    headers: trimHeaders(res.headers, ['set-cookie'])
  })
}
