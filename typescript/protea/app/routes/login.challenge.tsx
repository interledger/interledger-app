import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Button, Layouts, TextField } from '~/components'
import { route } from 'routes-gen'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError,
  requireUserSession
} from '~/lib/kratos.server'
import { commitSession, getSession } from '~/sessions'
import { trimHeaders } from '~/lib/headers.server'

export async function loader({ request }: LoaderArgs) {
  const session = await requireUserSession(request)
  const userSettings = await getSession(request.headers.get('Cookie'))
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const challengeFlow = url.searchParams.get('challenge-flow')
  const cookie = String(request.headers.get('cookie'))

  if (challengeFlow == 'settings-password') {
    userSettings.set('challenge-flow', {
      type: challengeFlow,
      returnTo: '/settings/password',
      email: session.identity.traits.email
    })
  }

  let flow
  if (flowId) {
    // If ?flow=.. was in the URL, we fetch it
    const flowRes = await fetch(
      `${KRATOS_URL}/self-service/login/flows?id=${flowId}`,
      {
        headers: {
          cookie: cookie,
          Accept: 'application/json'
        }
      }
    )
    flow = await flowRes.json()
    if (flowRes.status >= 400) handleFlowError(flow, 'login/challenge')
  } else {
    // Otherwise we initialize it
    const flowRes = await fetch(
      `${KRATOS_URL}/self-service/login/browser?refresh=true`,
      { headers: { Accept: 'application/json' } }
    )
    if (flowRes.status >= 400) handleFlowError(flow, 'login/challenge')
    flow = await flowRes.json()
    flowRes.headers.append('Set-Cookie', await commitSession(userSettings))
    return redirect(`/login/challenge?flow=${flow.id}`, {
      headers: trimHeaders(flowRes.headers, ['set-cookie'])
    })
  }
  return json({
    flow,
    csrfToken: getCsrfTokenFromFlow(flow),
    email: session.identity.traits.email
  })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { flow, csrfToken, email } = useLoaderData<typeof loader>()

  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <h1 className='mb-6 font-display text-2xl font-medium'>
        Confirm it's you
      </h1>
      <span>You are trying to edit sensitive account information.</span>
      <Form
        id='login-challenge'
        action={`/login/challenge?flow=${flow.id}`}
        method='post'
        className='hidden'
      />
      <TextField
        id='password'
        form='login-challenge'
        label='Password'
        name='password'
        type='password'
        className='mt-6'
        aria-invalid={Boolean(actionData?.errors?.password) || undefined}
        aria-describedby={
          actionData?.errors?.password ? 'password-error' : undefined
        }
        required
        errorMessage={actionData?.errors?.password}
      />

      <input
        form='login-challenge'
        defaultValue={email}
        name='email'
        type='hidden'
      />
      <input
        form='login-challenge'
        defaultValue={csrfToken}
        name='csrf_token'
        type='hidden'
      />

      <Button className='mt-6' form='login-challenge' type='submit'>
        Continue
      </Button>
    </div>
  )
}

export async function action({ request }: ActionArgs) {
  const userSettings = await getSession(request.headers.get('Cookie'))
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')

  const form = await request.formData()
  const csrfToken = form.get('csrf_token') as string
  const email = form.get('email') as string
  const password = form.get('password') as string

  const fieldErrors = {
    email: '',
    password: ''
  }

  const res = await fetch(`${KRATOS_URL}/self-service/login?flow=${flowId}`, {
    method: 'POST',
    body: JSON.stringify({
      method: 'password',
      identifier: email,
      password: password,
      csrf_token: csrfToken
    }),
    headers: {
      'Content-type': 'application/json',
      Accept: 'application/json',
      Cookie: String(request.headers.get('cookie'))
    }
  })

  const data = await res.json()

  // 4000001 is an error if the user already has a privileged session.
  if (res.status >= 400 && data.ui.messages[0].id !== 4000001) {
    for (let node of data.ui.nodes) {
      if (node.messages.length > 0) {
        Object.assign(fieldErrors, {
          [node.attributes.name]: node.messages[0].text
        })
      }
    }
    return json({ errors: { ...fieldErrors } }, { status: 400 })
  }

  const headers = trimHeaders(res.headers, ['set-cookie'])
  if (userSettings.has('challenge-flow')) {
    const { returnTo } = userSettings.get('challenge-flow')
    return redirect(returnTo, {
      headers: headers
    })
  }
  return redirect(route('/'), {
    headers: headers
  })
}
