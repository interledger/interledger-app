import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Button, Layouts, Snackbar, TextField } from '~/components'
import {
  getCsrfTokenFromFlow,
  getUserSession,
  handleFlowError,
  KRATOS_URL,
  kratosErrorMapping
} from '~/lib/kratos.server'
import { trimHeaders } from '~/lib/headers.server'
import { exitFlow, flowType, requireFlow } from '~/lib/flows.server'
import { useEffect, useState } from 'react'

export async function loader({ request }: LoaderArgs) {
  const session = await getUserSession(request)
  await requireFlow(request, flowType.PasswordChallenge)
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')

  const cookie = String(request.headers.get('cookie'))

  let kratosFlow
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
    kratosFlow = await flowRes.json()
    if (flowRes.status >= 400) handleFlowError(kratosFlow, 'login/challenge')
  } else {
    // Otherwise we initialize it
    const flowRes = await fetch(
      `${KRATOS_URL}/self-service/login/browser?refresh=true`,
      { headers: { Accept: 'application/json' } }
    )
    if (flowRes.status >= 400) handleFlowError(kratosFlow, 'login/challenge')
    kratosFlow = await flowRes.json()
    return redirect(`/login/challenge?flow=${kratosFlow.id}`, {
      headers: trimHeaders(flowRes.headers, ['set-cookie'])
    })
  }
  return json({
    flow: kratosFlow,
    csrfToken: getCsrfTokenFromFlow(kratosFlow),
    email: session.identity.traits.email
  })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { flow, csrfToken, email } = useLoaderData<typeof loader>()

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
  const flow = await requireFlow(request, flowType.PasswordChallenge)
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')

  const form = await request.formData()
  const csrfToken = form.get('csrf_token') as string
  const email = form.get('email') as string
  const password = form.get('password') as string

  const fieldErrors = {
    form: '',
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
  if (res.status >= 400) {
    const data = await res.json()
    // 4000001 is an error if the user already has a privileged session.
    if (data.ui.messages[0].id !== 4000001)
      return kratosErrorMapping(res, fieldErrors)
  }

  const headers = trimHeaders(res.headers, ['set-cookie'])
  await exitFlow(request, flowType.PasswordChallenge)
  return redirect(flow.returnTo, {
    headers: headers
  })
}
