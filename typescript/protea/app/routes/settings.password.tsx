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
import { commitSession, getSession } from '~/session.server'
import { trimHeaders } from '~/lib/headers.server'

export async function loader({ request }: LoaderArgs) {
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
    if (flowRes.status >= 400) handleFlowError(flow, 'settings/password')
  } else {
    // Otherwise we initialize it
    const flowRes = await fetch(
      `${KRATOS_URL}/self-service/settings/browser?${url.searchParams}`,
      { headers: { cookie: cookie, Accept: 'application/json' } }
    )
    flow = await flowRes.json()
    if (flowRes.status >= 400) handleFlowError(flow, 'settings/password')
    return redirect(`/settings/password?flow=${flow.id}`, {
      headers: trimHeaders(flowRes.headers, ['set-cookie'])
    })
  }
  return json({ flow, csrfToken: getCsrfTokenFromFlow(flow) })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { flow, csrfToken } = useLoaderData<typeof loader>()

  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <h1 className='mb-6 font-display text-2xl font-medium'>Set password</h1>
      <span>Set a new password to continue.</span>
      <Form
        id='settings-password'
        action={`/settings/password?flow=${flow.id}`}
        method='post'
        className='hidden'
      />
      <TextField
        id='new-password'
        form='settings-password'
        label='New password'
        name='new-password'
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
        form='settings-password'
        defaultValue={csrfToken}
        name='csrf_token'
        type='hidden'
      />

      <Button className='mt-6' form='settings-password' type='submit'>
        Continue
      </Button>
    </div>
  )
}

export async function action({ request }: ActionArgs) {
  const cookie = request.headers.get('Cookie')
  const userSettings = await getSession(cookie)
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')

  const form = await request.formData()
  const csrfToken = form.get('csrf_token') as string
  const password = form.get('new-password') as string

  const fieldErrors = { password: '' }
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
  if (res.status > 400) handleFlowError(data, 'settings/password')
  if (res.status == 400) {
    for (let node of data.ui.nodes) {
      if (node.messages.length > 0) {
        Object.assign(fieldErrors, {
          [node.attributes.name]: node.messages[0].text
        })
      }
    }
    return json({ errors: { ...fieldErrors } }, { status: 400 })
  }

  userSettings.unset('challenge-flow')
  userSettings.flash('snackbar', {
    message: 'New password successfully saved.',
    icon: 'close'
  })
  return redirect(route('/settings'), {
    headers: {
      'Set-Cookie': await commitSession(userSettings)
    }
  })
}
