import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Button, Layouts, Snackbar, TextField } from '~/components'
import { route } from 'routes-gen'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError,
  kratosErrorMapping
} from '~/lib/kratos.server'
import { trimHeaders } from '~/lib/headers.server'
import { flashSnackbar } from '~/lib/snackbar.server'
import { useEffect, useState } from 'react'

export const handle = {
  layout: Layouts.FocusLayout
}

export async function loader({ request }: LoaderArgs) {
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
    return redirect(`/recovery/password?flow=${flow.id}`, {
      headers: trimHeaders(flowRes.headers, ['set-cookie'])
    })
  }
  return json({ flow, csrfToken: getCsrfTokenFromFlow(flow) })
}

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { flow, csrfToken } = useLoaderData<typeof loader>()

  const [snackbarMessage, setSnackbar] = useState<any>(actionData?.errors.form)
  const [showSnackbar, setShowSnackbar] = useState<boolean>(
    Boolean(actionData?.errors.form) ?? false
  )

  useEffect(() => {
    if (actionData?.errors.form) {
      setSnackbar(actionData?.errors.form)
      setShowSnackbar(true)
    }
  }, [actionData])

  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <h1 className='mb-6 font-display text-2xl font-medium'>Set password</h1>
      <span>
        You've successfully recovered your account. Set a new password to
        continue.
      </span>
      <Form
        id='recovery-password'
        action={`/recovery/password?flow=${flow.id}`}
        method='post'
        className='hidden'
      />
      <TextField
        id='new-password'
        form='recovery-password'
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
        form='recovery-password'
        defaultValue={csrfToken}
        name='csrf_token'
        type='hidden'
      />

      <Button className='mt-6' form='recovery-password' type='submit'>
        Continue
      </Button>
      <Snackbar
        message={snackbarMessage}
        icon='close'
        show={showSnackbar}
        id='error-snackbar'
        onClose={() => {
          setSnackbar('')
          setShowSnackbar(false)
        }}
      />
    </div>
  )
}

export async function action({ request }: ActionArgs) {
  const cookie = request.headers.get('Cookie') as string
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')

  const form = await request.formData()
  const csrfToken = form.get('csrf_token') as string
  const password = form.get('new-password') as string

  const fieldErrors = {
    form: '',
    password: ''
  }
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
        cookie
      }
    }
  )
  const data = await res.json()
  if (res.status > 400) handleFlowError(data, 'recovery/password')
  else if (res.status == 400) {
    return kratosErrorMapping(res, fieldErrors)
  }

  return redirect(route('/settings'), {
    headers: {
      'Set-Cookie': await flashSnackbar(request, {
        message: 'New password successfully saved.',
        icon: 'close'
      })
    }
  })
}
