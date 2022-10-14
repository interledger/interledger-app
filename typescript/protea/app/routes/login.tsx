import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Button, HomeShapes, Router, TextField } from '~/components'
import { route } from 'routes-gen'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError,
  requireNoUserSession
} from '~/lib/kratos.server'
import { trimHeaders } from '~/lib/headers.server'

export async function loader({ request }: LoaderArgs) {
  await requireNoUserSession(request)
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const cookie = String(request.headers.get('cookie'))

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
    if (flowRes.status >= 400) handleFlowError(flow, 'login')
  } else {
    // Otherwise we initialize it
    const flowRes = await fetch(
      `${KRATOS_URL}/self-service/login/browser?${url.searchParams}`,
      { headers: { Accept: 'application/json' } }
    )
    flow = await flowRes.json()
    if (flowRes.status >= 400) handleFlowError(flow, 'login')
    return redirect(`/login?flow=${flow.id}`, {
      headers: trimHeaders(flowRes.headers, ['set-cookie'])
    })
  }
  return json({ flow, csrfToken: getCsrfTokenFromFlow(flow) })
}

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { flow, csrfToken } = useLoaderData<typeof loader>()
  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <div className='mt-2'>
        <HomeShapes />
      </div>
      <h1 className='mt-6 font-display text-2xl font-medium'>Log in</h1>
      <p className='mt-6 text-medium'>
        New to Fynbos?{' '}
        <Router className='text-primary' to={route('/signup')}>
          Sign up
        </Router>
      </p>
      <Form
        id='login'
        action={`/login?flow=${flow.id}`}
        method='post'
        className='hidden'
      />
      <TextField
        id='email'
        label='Email'
        name='email'
        type='email'
        form='login'
        className='mt-6'
        aria-invalid={Boolean(actionData?.errors?.email) || undefined}
        aria-describedby={actionData?.errors?.email ? 'email-error' : undefined}
        required
        errorMessage={actionData?.errors?.email}
      />
      <TextField
        id='password'
        label='Password'
        name='password'
        type='password'
        form='login'
        className='mt-1'
        aria-invalid={Boolean(actionData?.errors?.password) || undefined}
        aria-describedby={
          actionData?.errors?.password ? 'password-error' : undefined
        }
        required
        errorMessage={actionData?.errors?.password}
      />

      <input
        form='login'
        defaultValue={csrfToken}
        name='csrf_token'
        type='hidden'
      />

      <Button className='mt-12' form='login' type='submit'>
        Login
      </Button>
      <div className='mt-4 flex justify-end'>
        <Router to={route('/recovery')} aria-label='Forgot password?'>
          <span className='text-primary'>Forgot password?</span>
        </Router>
      </div>
    </div>
  )
}

export async function action({ request }: ActionArgs) {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const returnTo = url.searchParams.get('return_to')

  const form = await request.formData()
  const csrfToken = form.get('csrf_token')
  const email = form.get('email')
  const password = form.get('password')

  const fieldErrors = {
    email: '',
    password: ''
  }

  const res = await fetch(`${KRATOS_URL}/self-service/login?flow=${flowId}`, {
    method: 'POST',
    body: JSON.stringify({
      method: 'password',
      password_identifier: email,
      password: password,
      csrf_token: csrfToken
    }),
    headers: {
      'Content-type': 'application/json',
      cookie: String(request.headers.get('cookie'))
    }
  })

  const data = await res.json()
  if (res.status >= 400) {
    for (let node of data.ui.nodes) {
      if (node.messages.length > 0) {
        Object.assign(fieldErrors, {
          [node.attributes.name]: node.messages[0].text
        })
      }
    }
    return json({ errors: { ...fieldErrors } }, { status: 400 })
  }

  // Remove all headers besides set-cookie
  const headers = trimHeaders(res.headers, ['set-cookie'])

  if (returnTo) {
    return redirect(returnTo, {
      headers: headers
    })
  }
  return redirect(route('/'), {
    headers: headers
  })
}
