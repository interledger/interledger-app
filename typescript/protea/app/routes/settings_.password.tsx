import type { Route } from './+types/settings_.password'
import { data, redirect } from 'react-router';
import { Form, useActionData, useLoaderData } from 'react-router';
import { href } from 'react-router'
import type { ApplicationProps } from '~/components'
import { Button, Card, CardContent, Layouts, TextField } from '~/components'
import { error } from '~/lib/error.server'
import { trimHeaders } from '~/lib/headers.server'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError,
  kratosErrorMapping
} from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'

export async function loader({ request }: Route.LoaderArgs) {
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
  return data({
    flow,
    csrfToken: getCsrfTokenFromFlow(flow)
    // backTo: href('/settings')
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: href('/settings'),
      title: 'Set password'
    }
  }
}

export const meta = mergeMeta(() => [
  {
    title: 'Set password'
  }
])

export default function Page() {
  const actionData = useActionData()
  const { flow, csrfToken } = useLoaderData()

  return (
    <>
      <Form
        id='settings-password'
        action={`/settings/password?flow=${flow.id}`}
        method='post'
        className='hidden'
      />
      <input
        form='settings-password'
        defaultValue={csrfToken}
        name='csrf_token'
        type='hidden'
      />
      <Card>
        <CardContent>
          <p>Set a new password to continue.</p>
        </CardContent>
        <TextField
          id='new-password'
          form='settings-password'
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
      </Card>
      <Button form='settings-password' type='submit'>
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

  if (res.status > 400) handleFlowError(data, 'settings/password')
  if (res.status == 400) {
    const errs = await kratosErrorMapping(res, fieldErrors)
    return error(request, { errors: errs })
  }

  return redirectWithSnackbar(request, href('/settings'), {
    message: 'New password successfully saved.',
    icon: 'close'
  })
}
