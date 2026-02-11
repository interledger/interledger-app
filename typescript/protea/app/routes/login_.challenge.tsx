import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Button, Card, CardContent, Layouts, TextField } from '~/components'
import { error } from '~/lib/error.server'
import { kratosPublic } from '~/lib/kratos/kratos-client.server'
import { getCookie, withCookie, buildHeadersWithCookies } from '~/lib/kratos/cookie.util'
import { getCsrfTokenFromFlow } from '~/lib/kratos/flow.util'
import { handleFlowError, mapFlowToFieldErrors } from '~/lib/kratos/error'
import { getUserSession } from '~/lib/kratos/session.util'
import { mergeMeta } from '~/lib/meta'

export async function loader({ request }: LoaderFunctionArgs) {
  const session = await getUserSession(request)
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const cookie = getCookie(request)

  if (flowId) {
    try {
      const { data: kratosFlow } = await kratosPublic.getLoginFlow(
        { id: flowId },
        withCookie(cookie)
      )
      return json({
        flow: kratosFlow,
        csrfToken: getCsrfTokenFromFlow(kratosFlow),
        email: session.identity?.traits.email
      })
    } catch (err: any) {
      handleFlowError(err, 'login/challenge')
      throw err
    }
  }

  // No flow ID — initialize a new login flow with refresh
  try {
    const response = await kratosPublic.createBrowserLoginFlow(
      { refresh: true },
      withCookie(cookie)
    )
    return redirect(`/login/challenge?flow=${response.data.id}`, {
      headers: buildHeadersWithCookies(response)
    })
  } catch (err: any) {
    handleFlowError(err, 'login/challenge')
    throw err
  }
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/'),
      title: 'Confirmation'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: "Confirm it's you"
  }
])

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { flow, csrfToken, email } = useLoaderData<typeof loader>()

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

export async function action({ request }: ActionFunctionArgs) {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')!
  const cookie = getCookie(request)

  const form = await request.formData()
  const csrfToken = form.get('csrf_token') as string
  const email = form.get('email') as string
  const password = form.get('password') as string

  const fieldErrors = {
    form: '',
    email: '',
    password: ''
  }

  try {
    const response = await kratosPublic.updateLoginFlow(
      {
        flow: flowId,
        updateLoginFlowBody: {
          method: 'password',
          identifier: email,
          password,
          csrf_token: csrfToken
        }
      },
      withCookie(cookie)
    )
    return redirect(route('/settings/password'), {
      headers: buildHeadersWithCookies(response)
    })
  } catch (err: any) {
    const flowData = err.response?.data
    // 4000001 = user already has a privileged session — not an error
    if (flowData?.ui?.messages?.[0]?.id === 4000001) {
      return redirect(route('/settings/password'), {
        headers: buildHeadersWithCookies(err.response)
      })
    }
    const errs = mapFlowToFieldErrors(flowData, fieldErrors)
    return error(request, { errors: errs })
  }
}
