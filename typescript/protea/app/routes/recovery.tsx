import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Button, Logo, Router, TextField } from '~/components'
import { route } from 'routes-gen'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError,
  requireNoUserSession
} from '~/lib/kratos.server'

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
  return json({ flow, csrfToken: getCsrfTokenFromFlow(flow) })
}

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { flow, csrfToken } = useLoaderData<typeof loader>()

  return (
    <main className='mx-auto grid min-h-screen w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 lg:content-center xl:max-w-4xl'>
      <div className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Router to={route('/')}>
          <Logo className='h-8' />
        </Router>
      </div>
      {flow.state === 'sent_email' && (
        <>
          <div className='col-span-full pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
            <h1 className='font-display text-4xl font-medium leading-normal'>
              Email sent!
            </h1>
          </div>
          <div className='col-span-full pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
            <p className='text-medium'>
              We've sent you an email to change your password. Please click on
              the link in the email to continue.
            </p>
          </div>
        </>
      )}
      {flow.state === 'choose_method' && (
        <>
          <div className='col-span-full pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
            <h1 className='font-display text-4xl font-medium leading-normal'>
              Recover your account
            </h1>
          </div>
          <div className='col-span-full pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
            <p className='text-medium'>
              We'll send you an email to change your password.
            </p>
          </div>
        </>
      )}
      <Form
        action={`/recovery?flow=${flow.id}`}
        method='post'
        className='col-span-full flex flex-col items-end space-y-2 sm:col-span-6 sm:col-start-2 lg:col-start-4'
      >
        <TextField
          id='email'
          label='Email'
          name='email'
          type='email'
          disabled={flow.state === 'sent_email'}
          aria-invalid={Boolean(actionData?.errors?.email) || undefined}
          aria-describedby={
            actionData?.errors?.email ? 'email-error' : undefined
          }
          required
          errorMessage={actionData?.errors?.email}
        />

        <input defaultValue={csrfToken} name='csrf_token' type='hidden' />
        <div className='pt-4'>
          <Button disabled={flow.state === 'sent_email'} type='submit'>
            Recover account
          </Button>
        </div>
      </Form>
    </main>
  )
}

export async function action({ request }: ActionArgs) {
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
  return json(data)
}
