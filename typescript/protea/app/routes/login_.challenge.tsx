import {
  Form,
  data,
  href,
  redirect,
  useActionData,
  useLoaderData
} from 'react-router'
import type { ApplicationProps } from '~/components'
import { Button, Card, CardContent, Layouts, TextField } from '~/components'
import { error } from '~/lib/error.server'
import {
  buildHeadersWithCookies,
  getCookie,
  withCookie
} from '~/lib/kratos/cookie.server'
import {
  handleFlowError,
  mapFlowToFieldErrors
} from '~/lib/kratos/error.server'
import { getCsrfTokenFromFlow } from '~/lib/kratos/flow.server'
import { kratosPublic } from '~/lib/kratos/kratos-client.server'
import { getSessionTraits, getUserSession } from '~/lib/kratos/session.server'
import logger from '~/lib/logger.server'
import { mergeMeta } from '~/lib/meta'
import type { Route } from './+types/login_.challenge'

// 4000001 represents the Kratos message ID when a user already has a privileged session
const ERROR_ID_SESSION_ALREADY_PRIVILEGED = 4000001

/**
 * The login challenge flow is triggered when the user is already authenticated
 * but attempts to access a highly sensitive route (like changing their password
 * or updating security settings).
 *
 * This functions as a "sudo mode" check or session elevation, forcing the user
 * to re-authenticate with their password to confirm their identity.
 *
 * It can also be triggered if the session is reported as `session_refresh_required`
 * by Kratos during certain critical operations.
 */
export async function loader({ request }: Route.LoaderArgs) {
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
      return data({
        flow: kratosFlow,
        csrfToken: getCsrfTokenFromFlow(kratosFlow),
        email: getSessionTraits(session).email
      })
    } catch (err: any) {
      handleFlowError(err, 'login/challenge')
      logger.error(
        { error: err, route: 'login.challenge' },
        'Failed to load login challenge flow'
      )
      throw new Error('Failed to load login challenge flow')
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
    logger.error(
      { error: err, route: 'login.challenge' },
      'Failed to initialize login challenge flow'
    )
    throw new Error('Failed to initialize login challenge flow')
  }
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
  if (!flowId) {
    return redirect(href('/login/challenge'))
  }
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
    return redirect(href('/settings/password'), {
      headers: buildHeadersWithCookies(response)
    })
  } catch (err: any) {
    const flowData = err.response?.data
    // User already has a privileged session — not an error
    if (
      flowData?.ui?.messages?.[0]?.id === ERROR_ID_SESSION_ALREADY_PRIVILEGED
    ) {
      return redirect(href('/settings/password'), {
        headers: buildHeadersWithCookies(err.response)
      })
    }
    const errs = mapFlowToFieldErrors(flowData, fieldErrors)
    return error(request, { errors: errs })
  }
}
