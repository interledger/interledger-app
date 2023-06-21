import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Button,
  Card,
  CardContent,
  Layouts,
  Snackbar,
  TextField
} from '~/components'
import { trimHeaders } from '~/lib/headers.server'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError,
  kratosErrorMapping
} from '~/lib/kratos.server'
import { flashSnackbar } from '~/lib/snackbar.server'

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
  return json({
    flow,
    csrfToken: getCsrfTokenFromFlow(flow)
    // backTo: route('/settings')
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/settings'),
      title: 'Set password'
    }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Settings | Set password'
  }
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
          <TextField
            id='new-password'
            form='settings-password'
            label='New password'
            name='new-password'
            type='password'
            className='mt-4'
            aria-invalid={Boolean(actionData?.errors?.password) || undefined}
            aria-describedby={
              actionData?.errors?.password ? 'password-error' : undefined
            }
            required
            errorMessage={actionData?.errors?.password}
          />
        </CardContent>
      </Card>
      <Button form='settings-password' type='submit'>
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
    </>
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

  if (res.status > 400) handleFlowError(data, 'settings/password')
  if (res.status == 400) {
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
