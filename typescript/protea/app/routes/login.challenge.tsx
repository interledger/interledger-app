import { useActionData, json, Form, redirect, useLoaderData } from 'remix'
import type { ActionFunction, LoaderFunction } from 'remix'
import { Button, Logo, MailIcon, Router, TextField } from '~/components'
import React from 'react'
import { route } from 'routes-gen'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError,
  requireUserSession
} from '~/lib/kratos.server'
import { commitSession, getSession } from '~/sessions'

type ActionData = {
  formError?: string
  fieldErrors?: {
    email?: string
    password?: string
  }
  fields?: {
    email: string
    password: string
    csrf_token: string
  }
}

const badRequest = (data: ActionData) => json(data, { status: 400 })

export const action: ActionFunction = async ({ request }) => {
  const userSettings = await getSession(request.headers.get('Cookie'))
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const form = await request.formData()
  const csrfToken = form.get('csrf_token')
  const email = form.get('email')
  const password = form.get('password')
  if (
    typeof csrfToken !== 'string' ||
    typeof email !== 'string' ||
    typeof password !== 'string'
  ) {
    return badRequest({
      // TODO: handle formError on client
      formError: `Form not submitted correctly.`
    })
  }
  const fields = { csrf_token: csrfToken, email, password }
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
    let fieldErrors: ActionData['fieldErrors'] = {}
    for (let node of data.ui.nodes) {
      if (node.messages.length > 0) {
        Object.assign(fieldErrors, {
          [node.attributes.name]: node.messages[0].text
        })
      }
    }
    return badRequest({ fieldErrors: fieldErrors, fields })
  }
  if (userSettings.has('challenge-flow')) {
    const { returnTo } = userSettings.get('challenge-flow')
    return redirect(returnTo, {
      headers: res.headers
    })
  }
  return redirect(route('/home'), {
    headers: res.headers
  })
}

export const loader: LoaderFunction = async ({ request, params }) => {
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
      headers: flowRes.headers
    })
  }
  return json({
    flow,
    csrfToken: getCsrfTokenFromFlow(flow),
    email: session.identity.traits.email
  })
}

export default function LoginChallengePage() {
  const actionData = useActionData<ActionData>()
  const { flow, csrfToken, email } = useLoaderData()

  return (
    <main className='mx-auto grid min-h-screen w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 lg:content-center xl:max-w-4xl'>
      <div className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Router to={route('/')}>
          <Logo className='h-8' />
        </Router>
      </div>
      <div className='col-span-full pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <h1 className='font-display text-4xl font-medium leading-normal text-strong'>
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
            <MailIcon />
          </div>
          <span className='text-small font-normal text-medium'>{email}</span>
        </div>
        <TextField
          id='password'
          label='Password'
          name='password'
          defaultValue={actionData?.fields?.password}
          type='password'
          aria-invalid={Boolean(actionData?.fieldErrors?.password) || undefined}
          aria-describedby={
            actionData?.fieldErrors?.password ? 'password-error' : undefined
          }
          required
          errorMessage={actionData?.fieldErrors?.password}
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
