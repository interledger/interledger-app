import type { ActionFunction, LoaderFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Button, Logo, Router, TextField } from '~/components'
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
    password?: string
  }
  fields?: {
    password: string
    csrf_token: string
  }
}

const badRequest = (data: ActionData) => json(data, { status: 400 })

export const action: ActionFunction = async ({ request }) => {
  const cookie = request.headers.get('Cookie')
  const userSettings = await getSession(cookie)
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const form = await request.formData()
  const csrfToken = form.get('csrf_token')
  const password = form.get('new-password')
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
        Accept: 'application/json',
        cookie: String(request.headers.get('Cookie'))
      }
    }
  )

  const data = await res.json()
  if (res.status > 400) handleFlowError(data, 'recovery/password')
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

  userSettings.flash('snackbar', {
    message: 'New password successfully saved.',
    action: 'done'
  })
  return redirect(route('/settings'), {
    headers: {
      'Set-Cookie': await commitSession(userSettings)
    }
  })
}

export const loader: LoaderFunction = async ({ request }) => {
  await requireUserSession(request)
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
    if (flowRes.status >= 400) handleFlowError(flow, 'recovery/password')
  } else {
    // Otherwise we initialize it
    const flowRes = await fetch(
      `${KRATOS_URL}/self-service/settings/browser?${url.searchParams}`,
      { headers: { cookie: cookie, Accept: 'application/json' } }
    )
    flow = await flowRes.json()
    if (flowRes.status >= 400) handleFlowError(flow, 'recovery/password')
    return redirect(`/recovery/password?flow=${flow.id}`)
  }
  return json({ flow, csrfToken: getCsrfTokenFromFlow(flow) })
}

export default function RecoveryPasswordPage() {
  const actionData = useActionData<ActionData>()
  const { flow, csrfToken } = useLoaderData()

  return (
    <main className='mx-auto grid min-h-screen w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 lg:content-center xl:max-w-4xl'>
      <div className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Router to={route('/')}>
          <Logo className='h-8' />
        </Router>
      </div>
      <div className='col-span-full pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <h1 className='font-display text-4xl font-medium leading-normal text-strong'>
          Set a new password
        </h1>
      </div>
      <div className='col-span-full pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <p className='text-medium'>
          You've successfully recovered your account.
          <br />
          Set a new password to continue.
        </p>
      </div>

      <Form
        action={`/recovery/password?flow=${flow.id}`}
        method='post'
        className='col-span-full flex flex-col items-end space-y-2 sm:col-span-6 sm:col-start-2 lg:col-start-4'
      >
        <TextField
          id='new-password'
          label='New password'
          name='new-password'
          defaultValue={actionData?.fields?.password}
          type='password'
          aria-invalid={Boolean(actionData?.fieldErrors?.password) || undefined}
          aria-describedby={
            actionData?.fieldErrors?.password ? 'password-error' : undefined
          }
          required
          errorMessage={actionData?.fieldErrors?.password}
        />

        <input defaultValue={csrfToken} name='csrf_token' type='hidden' />

        <div className='pt-4'>
          <Button type='submit'>Save password</Button>
        </div>
      </Form>
    </main>
  )
}
