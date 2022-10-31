import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import {
  Form,
  useActionData,
  useLoaderData,
  useTransition
} from '@remix-run/react'
import { Button, Layouts, Snackbar, TextField } from '~/components'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError,
  requireNoUserSession
} from '~/lib/kratos.server'
import { useEffect, useState } from 'react'
import { commitSession, getSession } from '~/sessions'

export async function loader({ request }: LoaderArgs) {
  await requireNoUserSession(request)
  const userSettings = await getSession(request.headers.get('Cookie'))
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

  const snackbar = {
    // NOTE: userSettings.has must be called before userSettings.get
    show: userSettings.has('snackbar'),
    ...userSettings.get('snackbar')
  }

  return json(
    { flow, snackbar, csrfToken: getCsrfTokenFromFlow(flow) },
    {
      headers: { 'Set-Cookie': await commitSession(userSettings) }
    }
  )
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { flow, snackbar, csrfToken } = useLoaderData<typeof loader>()

  const transition = useTransition()
  const [showSnackbar, setSnackbar] = useState<boolean>(snackbar.show)

  useEffect(() => {
    if (transition.state == 'idle' && transition.type == 'idle') {
      setSnackbar(snackbar.show)
    }
  }, [transition.type, transition.state, snackbar.show])

  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <h1 className='mb-6 font-display text-2xl font-medium'>
        Recover account
      </h1>
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
        aria-describedby={actionData?.errors?.email ? 'email-error' : undefined}
        required
        errorMessage={actionData?.errors?.email}
      />

      <input
        form='recovery'
        defaultValue={csrfToken}
        name='csrf_token'
        type='hidden'
      />

      <Button className='mt-6' form='recovery' type='submit'>
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
    </div>
  )
}

export async function action({ request }: ActionArgs) {
  const userSettings = await getSession(request.headers.get('Cookie'))
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')

  const form = await request.formData()
  const csrfToken = form.get('csrf_token')
  const email = form.get('email')

  const fieldErrors = { email: '' }

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

  const data = await res.json()
  if (res.status >= 400) {
    for (let node of data.ui.nodes) {
      if (node.messages.length > 0) {
        Object.assign(fieldErrors, {
          [node.attributes.name]: node.messages[0].text
        })
      }
    }
    return json({ errors: { ...fieldErrors } }, { status: 400 })
  }

  userSettings.flash('snackbar', {
    message: 'Recovery email successfully sent.',
    icon: 'close'
  })

  return redirect(`/recovery?flow=${flowId}`, {
    headers: {
      'Set-Cookie': await commitSession(userSettings)
    }
  })
}
