import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import {
  Form,
  useActionData,
  useLoaderData,
  useNavigation
} from '@remix-run/react'
import { useEffect, useState } from 'react'
import type { ApplicationProps } from '~/components'
import { Button, Card, Layouts, Snackbar, TextField } from '~/components'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError,
  kratosErrorMapping,
  requireNoUserSession
} from '~/lib/kratos.server'
import { flashSnackbar, getSnackbar } from '~/lib/snackbar.server'

export async function loader({ request }: LoaderArgs) {
  await requireNoUserSession(request)
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const cookie = String(request.headers.get('cookie'))

  let flow
  if (flowId) {
    // If ?flow=.. was in the URL, we fetch it
    const flowRes = await fetch(
      `${KRATOS_URL}/self-service/recovery/flows?id=${flowId}`,
      {
        headers: {
          cookie: cookie,
          Accept: 'application/json'
        }
      }
    )
    flow = await flowRes.json()
    if (flowRes.status >= 400) handleFlowError(flow, 'recovery')
  } else {
    // Otherwise we initialize it
    const flowRes = await fetch(
      `${KRATOS_URL}/self-service/recovery/browser?${url.searchParams}`,
      { headers: { Accept: 'application/json' } }
    )
    flow = await flowRes.json()
    if (flowRes.status >= 400) handleFlowError(flow, 'recovery')
    return redirect(`/recovery?flow=${flow.id}`, {
      headers: flowRes.headers
    })
  }

  const snackbar = await getSnackbar(request)

  return json({ flow, snackbar, csrfToken: getCsrfTokenFromFlow(flow) })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      title: 'Recover account'
    }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Recover account'
  }
}

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { flow, snackbar, csrfToken } = useLoaderData<typeof loader>()

  const navigation = useNavigation()
  const [showSnackbar, setSnackbar] = useState<boolean>(snackbar.show ?? false)

  useEffect(() => {
    if (navigation.state == 'idle') {
      setSnackbar(snackbar.show ?? false)
    }
  }, [navigation.state, snackbar.show])

  return (
    <>
      <Card>
        <span>
          Enter your email address and we will email you a link to change your
          password.
        </span>
        <Form
          id='recovery'
          action={`/recovery?flow=${flow.id}`}
          method='post'
          className='hidden'
        />
        <TextField
          id='email'
          form='recovery'
          label='Email'
          name='email'
          type='email'
          className='mt-6'
          aria-invalid={Boolean(actionData?.errors?.email) || undefined}
          aria-describedby={
            actionData?.errors?.email ? 'email-error' : undefined
          }
          required
          errorMessage={actionData?.errors?.email}
        />

        <input
          form='recovery'
          defaultValue={csrfToken}
          name='csrf_token'
          type='hidden'
        />
      </Card>
      <Button form='recovery' type='submit'>
        Recover account
      </Button>
      <Snackbar
        message={snackbar.message}
        icon={snackbar.icon}
        action={snackbar.action}
        show={showSnackbar}
        id='recovery-snackbar'
        onClose={() => setSnackbar(false)}
        dismissAfter={3000}
      />
    </>
  )
}

export async function action({ request }: ActionArgs) {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')

  const form = await request.formData()
  const csrfToken = form.get('csrf_token')
  const email = form.get('email')

  const fieldErrors = {
    form: '',
    email: ''
  }

  const res = await fetch(
    `${KRATOS_URL}/self-service/recovery?flow=${flowId}`,
    {
      method: 'POST',
      body: JSON.stringify({
        method: 'link',
        email: email,
        csrf_token: csrfToken
      }),
      headers: {
        'Content-type': 'application/json',
        cookie: String(request.headers.get('cookie'))
      }
    }
  )
  if (res.status >= 400) {
    return kratosErrorMapping(res, fieldErrors)
  }

  return redirect(`/recovery?flow=${flowId}`, {
    headers: {
      'Set-Cookie': await flashSnackbar(request, {
        message: 'Recovery email successfully sent.',
        icon: 'close'
      })
    }
  })
}
