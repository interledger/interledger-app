import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import {
  Form,
  useActionData,
  useLoaderData,
  useRevalidator
} from '@remix-run/react'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Button, Card, CardContent, Layouts, TextField } from '~/components'
import { error } from '~/lib/error.server'
import { kratosPublic, CLEAR_SESSION_COOKIE_HEADER } from '~/lib/kratos/kratos-client.server'
import { getCookie, withCookie, buildHeadersWithCookies } from '~/lib/kratos/cookie.server'
import { getCsrfTokenFromFlow } from '~/lib/kratos/flow.server'
import { handleFlowError, mapFlowToFieldErrors } from '~/lib/kratos/error.server'
import { hasUserSession } from '~/lib/kratos/session.server'
import { mergeMeta } from '~/lib/meta'
import { safeReturnTo } from '~/lib/url.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      title: 'Set password'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Set password'
  }
])

export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const cookie = getCookie(request)
  const isUser = hasUserSession(request)

  if (!isUser) {
    return json({ flowId: '', csrfToken: '' })
  }

  if (flowId) {
    try {
      const { data: flow } = await kratosPublic.getSettingsFlow(
        { id: flowId },
        withCookie(cookie)
      )
      return json({ flowId: flow.id, csrfToken: getCsrfTokenFromFlow(flow) })
    } catch (err: any) {
      handleFlowError(err, 'recovery/password')
      throw err
    }
  }

  // Initialize new settings flow
  try {
    const response = await kratosPublic.createBrowserSettingsFlow(
      { returnTo: safeReturnTo(url.searchParams.get('returnTo')) },
      withCookie(cookie)
    )
    return redirect(`/recovery/password?flow=${response.data.id}`, {
      headers: buildHeadersWithCookies(response)
    })
  } catch (err: any) {
    const flowData = err.response?.data
    if (flowData?.error?.id === 'session_aal2_required') {
      return redirect('/totp/challenge?returnTo=/recovery/password')
    }
    handleFlowError(err, 'recovery/password')
    throw err
  }
}

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { flowId, csrfToken } = useLoaderData<typeof loader>()
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

export async function action({ request }: ActionFunctionArgs) {
  const cookie = getCookie(request)
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  if (!flowId) return redirect('/recovery/password')

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

  try {
    await kratosPublic.updateSettingsFlow(
      {
        flow: flowId,
        updateSettingsFlowBody: {
          method: 'password',
          password,
          csrf_token: csrfToken
        }
      },
      withCookie(cookie)
    )
    return redirectWithSnackbar(
      request,
      route('/login'),
      {
        message:
          'New password successfully saved. Please log in with your new password.',
        icon: 'close'
      },
      {
        headers: {
          'Set-Cookie':
            CLEAR_SESSION_COOKIE_HEADER
        }
      }
    )
  } catch (err: any) {
    const status = err.response?.status
    const flowData = err.response?.data
    if (status === 400) {
      const errs = mapFlowToFieldErrors(flowData, fieldErrors)
      return error(request, { errors: errs })
    }
    handleFlowError(err, 'recovery/password')
    throw err
  }
}
