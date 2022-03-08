import { ActionFunction, LoaderFunction, redirect, useLoaderData } from 'remix'
import { useActionData, json, Form } from 'remix'
import { Button, Logo, Router, TextField } from '~/components'
import React from 'react'
import { route } from 'routes-gen'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError,
  requireUserSession
} from '~/lib/kratos.server'

type ActionData = {
  formError?: string
  fieldErrors?: {
    password?: string
  }
  fields?: {
    password: string
    csrf_token: string
  }
}

const badRequest = (data: ActionData) => json(data, { status: 400 })

export const action: ActionFunction = async ({ request }) => {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const form = await request.formData()
  const csrfToken = form.get('csrf_token')
  const password = form.get('password')

  if (typeof csrfToken !== 'string' || typeof password !== 'string') {
    return badRequest({
      formError: `Form not submitted correctly.`
    })
  }

  const fields = { csrf_token: csrfToken, password }
  const res = await fetch(
    `${KRATOS_URL}/self-service/settings?flow=${flowId}`,
    {
      method: 'POST',
      body: JSON.stringify({
        method: 'password',
        password: password,
        csrf_token: csrfToken
      }),
      headers: {
        'Content-type': 'application/json',
        cookie: String(request.headers.get('cookie'))
      }
    }
  )

  const data = await res.json()
  if (res.status > 400) handleFlowError(data, 'settings/password')
  if (res.status == 400) {
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
  // TODO show a success notification
  return redirect(route('/settings'), {
    headers: res.headers
  })
}

export const loader: LoaderFunction = async ({ request }) => {
  requireUserSession(request)
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const cookie = String(request.headers.get('cookie'))

  let flow
  if (flowId) {
    // If ?flow=.. was in the URL, we fetch it
    const flowRes = await fetch(
      `${KRATOS_URL}/self-service/settings/flows?id=${flowId}`,
      {
        headers: {
          cookie: cookie,
          Accept: 'application/json'
        }
      }
    )
    flow = await flowRes.json()
    if (flowRes.status >= 400) handleFlowError(flow, 'settings')
  } else {
    // Otherwise we initialize it
    const flowRes = await fetch(
      `${KRATOS_URL}/self-service/settings/browser?${url.searchParams}`,
      { headers: { cookie: cookie, Accept: 'application/json' } }
    )
    flow = await flowRes.json()
    if (flowRes.status >= 400) handleFlowError(flow, 'settings')
    return redirect(`/settings/password?flow=${flow.id}`, {
      headers: flowRes.headers
    })
  }
  return json({ flow, csrfToken: getCsrfTokenFromFlow(flow) })
}

export default function SettingsPasswordPage() {
  const actionData = useActionData<ActionData>()
  const { flow, csrfToken } = useLoaderData()

  return (
    <main className='mx-auto flex h-screen max-w-sm flex-col items-start justify-center px-4'>
      <Router to={route('/')} aria-label='Fynbos logo'>
        <Logo className='h-8' />
      </Router>
      <h1 className='mt-6 mb-1 font-display text-4xl font-medium leading-normal text-strong'>
        Set a new password
      </h1>
      <p className='mb-10 text-medium'>
        You've successfully recovered your account.
        <br />
        Set a new password to continue.
      </p>
      <Form
        action={`/settings/password?flow=${flow.id}`}
        method='post'
        className='flex min-w-full flex-col items-end space-y-4'
      >
        <TextField
          id='password'
          label='New password'
          name='password'
          defaultValue={actionData?.fields?.password}
          type='password'
          aria-invalid={Boolean(actionData?.fieldErrors?.password) || undefined}
          aria-describedby={
            actionData?.fieldErrors?.password ? 'password-error' : undefined
          }
          required
          minLength={8}
          errorMessage={actionData?.fieldErrors?.password}
        />

        <input defaultValue={csrfToken} name='csrf_token' type='hidden' />

        <Button type='submit'>Save password</Button>
      </Form>
    </main>
  )
}
