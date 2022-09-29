import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Button, Icon, Logo, Router, TextField } from '~/components'
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

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { flow, csrfToken, email } = useLoaderData<typeof loader>()

  return (
    <main className='mx-auto grid min-h-screen w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 lg:content-center xl:max-w-4xl'>
      <div className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Router to={route('/')}>
          <Logo className='h-8' />
        </Router>
      </div>
      <div className='col-span-full pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <h1 className='font-display text-4xl font-medium leading-normal'>
          Confirm it's you
        </h1>
      </div>
      <div className='col-span-full pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <p className='text-medium'>To continue, first verify it's you.</p>
      </div>
      {/* Form */}
      <Form
        action={`/login/challenge?flow=${flow.id}`}
        method='post'
        className='col-span-full flex flex-col items-end space-y-2 sm:col-span-6 sm:col-start-2 lg:col-start-4'
      >
        <div className='flex min-w-full items-center space-x-3 rounded-xl bg-container p-4'>
          <div className='text-medium'>
            <Icon>mail</Icon>
          </div>
          <span className='text-small text-medium'>{email}</span>
        </div>
        <TextField
          id='password'
          label='Password'
          name='password'
          type='password'
          aria-invalid={Boolean(actionData?.errors?.password) || undefined}
          aria-describedby={
            actionData?.errors?.password ? 'password-error' : undefined
          }
          required
          errorMessage={actionData?.errors?.password}
        />

        <input defaultValue={email} name='email' type='hidden' />
        <input defaultValue={csrfToken} name='csrf_token' type='hidden' />

        <div className='flex min-w-full items-center justify-end pt-4'>
          {/* <Router to={route('/recovery')} aria-label='Forgot password?'>
            <span className='text-primary'>Forgot password?</span>
          </Router> */}
          <Button type='submit'>Continue</Button>
        </div>
      </Form>
    </main>
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
