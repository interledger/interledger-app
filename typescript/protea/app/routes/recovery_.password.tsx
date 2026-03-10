import type { Route } from './+types/recovery_.password'
import { data, redirect } from 'react-router';
import { Form, useActionData, useLoaderData, useRevalidator } from 'react-router';
import { useEffect, useState } from 'react'
import { href } from 'react-router'
import type { ApplicationProps } from '~/components'
import { Button, Card, CardContent, Layouts, TextField } from '~/components'
import { error } from '~/lib/error.server'
import { trimHeaders } from '~/lib/headers.server'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError,
  hasUserSession,
  kratosErrorMapping
} from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      title: 'Set password'
    }
  }
}

export const meta = mergeMeta(() => [
  {
    title: 'Set password'
  }
])

export async function loader({ request }: Route.LoaderArgs) {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const cookie = String(request.headers.get('cookie'))
  const isUser = hasUserSession(request)

  if (!isUser) {
    return data({ flowId: '', csrfToken: '' })
  }

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
    if (flowRes.status >= 400) handleFlowError(flow, 'recovery/password', flowId)
  } else {
    // Otherwise we initialize it
    const flowRes = await fetch(
      `${KRATOS_URL}/self-service/settings/browser?${url.searchParams}`,
      { headers: { cookie: cookie, Accept: 'application/json' } }
    )
    flow = await flowRes.json()

    if (flow.error?.id === 'session_aal2_required') {
      const redirectURL = `/totp/challenge?returnTo=/recovery/password${flow.id ? encodeURIComponent(`?flow=${flow.id}`) : ''
        }`
      return redirect(redirectURL)
    }

    if (flowRes.status >= 400) handleFlowError(flow, 'recovery/password', flow.id)
    return redirect(`/recovery/password?flow=${flow.id}`, {
      headers: trimHeaders(flowRes.headers, ['set-cookie'])
    })
  }
  return data({ flowId: flow.id, csrfToken: getCsrfTokenFromFlow(flow) })
}

export default function Page() {
  const actionData = useActionData()
  const { flowId, csrfToken } = useLoaderData()
  const { revalidate, state } = useRevalidator()

  const [revalidateCount, setRevalidateCount] = useState<number>(0)

  useEffect(() => {
    if (
      flowId == '' &&
      csrfToken == '' &&
      state === 'idle' &&
      revalidateCount < 1
    ) {
      revalidate()
      setRevalidateCount(1)
    }
  }, [csrfToken, flowId, revalidate, revalidateCount, state])

  return (
    <>
      <Form
        id='recovery-password'
        action={`/recovery/password?flow=${flowId}`}
        method='post'
        className='hidden'
      />
      <Card>
        <CardContent>
          <span>
            You've successfully recovered your account. Set a new password to
            continue.
          </span>
        </CardContent>
        <TextField
          id='new-password'
          form='recovery-password'
          label='New password'
          name='new-password'
          type='password'
          className='mt-2'
          aria-invalid={Boolean(actionData?.errors?.password) || undefined}
          aria-describedby={
            actionData?.errors?.password ? 'password-error' : undefined
          }
          required
          errorMessage={actionData?.errors?.password}
        />
        <TextField
          id='confirm-new-password'
          form='recovery-password'
          label='Confifm New Password'
          name='confirm-new-password'
          type='password'
          className='mt-2'
          required
        />
        <input
          form='recovery-password'
          defaultValue={csrfToken}
          name='csrf_token'
          type='hidden'
        />
      </Card>
      <Button form='recovery-password' type='submit'>
        Continue
      </Button>
    </>
  )
}

export async function action({ request }: Route.ActionArgs) {
  const cookie = request.headers.get('Cookie') as string
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')

  const form = await request.formData()
  const csrfToken = form.get('csrf_token') as string
  const password = form.get('new-password') as string
  const confirmPassword = form.get('confirm-new-password') as string

  const fieldErrors = {
    form: '',
    password: ''
  }

  if (password !== confirmPassword) {
    fieldErrors.password = 'Passwords do not match'
    return error(request, { errors: fieldErrors })
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
  if (res.status > 400) {
    const data = await res.json()
    handleFlowError(data, 'recovery/password')
  }
  else if (res.status == 400) {
    const errs = await kratosErrorMapping(res, fieldErrors)
    return error(request, { errors: errs })
  }

  return redirectWithSnackbar(
    request,
    href('/login'),
    {
      message:
        'New password successfully saved. Please log in with your new password.',
      icon: 'close'
    },
    {
      headers: {
        'Set-Cookie':
          'ory_kratos_session=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax'
      }
    }
  )
}
