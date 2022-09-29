import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Button, Router, TextField } from '~/components'
import { route } from 'routes-gen'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError,
  requireNoUserSession
} from '~/lib/kratos.server'

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
      headers: flowRes.headers
    })
  }
  return json({ flow, csrfToken: getCsrfTokenFromFlow(flow) })
}

const shapes = [
  [
    'bg-slate-600 rounded-tl-full',
    'bg-transparent',
    'bg-yellow-400 rounded-tr-full',
    'bg-rose-300 rounded-tl-full',
    'bg-lime-400 rounded-full',
    'bg-transparent',
    'bg-rose-500 rounded-full',
    'bg-lime-300 rounded-tr-full',
    'bg-transparent',
    'bg-transparent'
  ],
  [
    'bg-transparent',
    'bg-rose-400 rounded-full',
    'bg-lime-500 rounded-bl-full',
    'bg-transparent',
    'bg-slate-300 rounded-tl-full',
    'bg-yellow-200 rounded-tl-full',
    'bg-slate-500 rounded-br-full',
    'bg-transparent',
    'bg-rose-100 rounded-full',
    'bg-rose-300 rounded-bl-full'
  ]
]

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { flow, csrfToken } = useLoaderData<typeof loader>()
  return (
    <div className='mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto rounded-2xl bg-page px-4 pb-16 pt-6 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:pt-12 xl:max-w-4xl'>
      <div className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2'>
        {shapes.map((shapeRow) => (
          <div className='flex' key={shapeRow.toString()}>
            {shapeRow.map((shape, index) => (
              <div
                key={shape + index}
                className={`aspect-square w-full ${shape}`}
              />
            ))}
          </div>
        ))}
      </div>
      <div className='col-span-full flex flex-col space-y-2 pt-4 pb-8 sm:col-span-6 sm:col-start-2'>
        <span className='font-display text-2xl font-medium'>Log in</span>
        <span className='text-medium'>
          New to Fynbos?{' '}
          <Router className='text-primary' to={route('/signup')}>
            Sign up
          </Router>
        </span>
      </div>
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
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2'
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
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2'
        aria-invalid={Boolean(actionData?.errors?.password) || undefined}
        aria-describedby={
          actionData?.errors?.password ? 'password-error' : undefined
        }
        required
        errorMessage={actionData?.errors?.password}
      />

      <input form='login' defaultValue={csrfToken} name='csrf_token' type='hidden' />

      <div className='col-span-full sm:col-span-6 sm:col-start-2'>
        <Button form='login' type='submit'>
          Login
        </Button>
      </div>
      <div className='col-span-full flex justify-end sm:col-span-6 sm:col-start-2'>
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
  if (returnTo) {
    return redirect(returnTo, {
      headers: res.headers
    })
  }
  return redirect(route('/'), {
    headers: res.headers
  })
}
