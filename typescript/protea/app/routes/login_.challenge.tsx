import type { Route } from './+types/login_.challenge'
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
  getUserSession,
  handleFlowError,
  kratosErrorMapping
} from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'

export async function loader({ request }: Route.LoaderArgs) {
  const session = await getUserSession(request)
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
  return data({
    flow: kratosFlow,
    csrfToken: getCsrfTokenFromFlow(kratosFlow),
    email: session.identity.traits.email
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: href('/'),
      title: 'Confirmation'
    }
  }
}

export const meta = mergeMeta(() => [
  {
    title: "Confirm it's you"
  }
])

export default function Page() {
  const actionData = useActionData()
  const { flow, csrfToken, email } = useLoaderData()

  return (
    <>
      <Form
        id='login-challenge'
        action={`/login/challenge?flow=${flow.id}`}
        method='post'
        className='hidden'
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
      <Card>
        <CardContent>
          <p>To continue, type your password.</p>
        </CardContent>
        <TextField
          id='password'
          form='login-challenge'
          label='Password'
          name='password'
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
      <Button form='login-challenge' type='submit'>
        Continue
      </Button>
    </>
  )
}

export async function action({ request }: Route.ActionArgs) {
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
    if (data.ui.messages[0].id !== 4000001) {
      const errs = await kratosErrorMapping(res, fieldErrors)
      return error(request, { errors: errs })
    }
  }

  const headers = trimHeaders(res.headers, ['set-cookie'])
  return redirect(href('/settings/password'), {
    headers: headers
  })
}
