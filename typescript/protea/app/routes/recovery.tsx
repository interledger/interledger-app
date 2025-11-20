import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import type { ApplicationProps } from '~/components'
import { Button, Card, CardContent, Layouts, TextField } from '~/components'
import { error } from '~/lib/error.server'
import { EnumIlpHeaders } from '~/lib/headers.types'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError,
  kratosErrorMapping,
  requireNoUserSession
} from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'

export async function loader({ request }: LoaderFunctionArgs) {
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

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      title: 'Recover account'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Recover account'
  }
])

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { flow, csrfToken } = useLoaderData<typeof loader>()

  return (
    <>
      <Form
        id='recovery'
        action={`/recovery?flow=${flow.id}`}
        method='post'
        className='hidden'
      />
      <input
        form='recovery'
        defaultValue={csrfToken}
        name='csrf_token'
        type='hidden'
      />
      <Card>
        <CardContent>
          <span>
            Enter your email address and we will email you a link to change your
            password.
          </span>
        </CardContent>
        <TextField
          id='email'
          form='recovery'
          label='Email'
          name='email'
          type='email'
          className='mt-2'
          aria-invalid={Boolean(actionData?.errors?.email) || undefined}
          aria-describedby={
            actionData?.errors?.email ? 'email-error' : undefined
          }
          required
          errorMessage={actionData?.errors?.email}
        />
      </Card>
      <Button form='recovery' type='submit'>
        Recover account
      </Button>
    </>
  )
}

export async function action({ request }: ActionFunctionArgs) {
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
        [EnumIlpHeaders.email]: email?.toString() ?? '',
        'Content-type': 'application/json',
        cookie: String(request.headers.get('cookie'))
      }
    }
  )
  if (res.status >= 400) {
    const errs = await kratosErrorMapping(res, fieldErrors)
    return error(request, { errors: errs })
  }

  return redirectWithSnackbar(request, `/recovery?flow=${flowId}`, {
    message: 'Recovery email successfully sent.',
    icon: 'close'
  })
}
