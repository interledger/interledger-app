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
  useSearchParams
} from '@remix-run/react'
import { useMemo } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Button,
  Card,
  CardHeader,
  CardTitle,
  Layouts,
  Router,
  TextField
} from '~/components'
import { error } from '~/lib/error.server'
import { trimHeaders } from '~/lib/headers.server'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError,
  isSessionAlreadyExitsMessage,
  kratosErrorMapping,
  requireNoUserSession
} from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'
import { useRecaptchaV2 } from '~/lib/useRecaptchaV2'

export async function loader({ request }: LoaderFunctionArgs) {
  await requireNoUserSession(request)
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const cookie = String(request.headers.get('cookie'))

  let flow
  let headers: Headers | null = null
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
    flow = await flowRes.json()
    if (flowRes.status >= 400) handleFlowError(flow, 'login')
  } else {
    // Otherwise we initialize it
    const flowRes = await fetch(
      `${KRATOS_URL}/self-service/login/browser${url.search}`,
      { headers: { Accept: 'application/json' } }
    )
    flow = await flowRes.json()
    if (flowRes.status >= 400) handleFlowError(flow, 'login')
    headers = trimHeaders(flowRes.headers, ['set-cookie'])
  }

  const returnTo = url.searchParams.get('returnTo')

  return json(
    {
      returnTo: returnTo ?? flow.ui.returnTo,
      flowId: flow.id,
      csrfToken: getCsrfTokenFromFlow(flow)
    },
    {
      headers: headers ?? undefined
    }
  )
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {}
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Log in'
  }
])

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { csrfToken, flowId, returnTo } = useLoaderData<typeof loader>()
  const searchParams = useSearchParams()
  const recaptchaSiteKey = useMemo(() => {
    return (
      (typeof window !== 'undefined' &&
        (window as any).ENV?.recaptchaSiteKey) ||
      ''
    )
  }, [])

  const { RecaptchaWidget: widget, token } = useRecaptchaV2(recaptchaSiteKey)

  return (
    <>
      <Form
        id='login'
        action={`/login?${searchParams[0]}`}
        method='post'
        className='hidden'
      />
      <input
        form='login'
        defaultValue={csrfToken}
        name='csrf_token'
        type='hidden'
      />
      <input form='login' defaultValue={flowId} name='flow_id' type='hidden' />
      <input
        form='login'
        defaultValue={returnTo}
        name='returnTo'
        type='hidden'
      />
      <input form='login' value={token} name='recaptcha_token' type='hidden' />
      <Card>
        <CardHeader>
          <CardTitle>Log in</CardTitle>
        </CardHeader>
        <TextField
          id='email'
          label='Email'
          name='email'
          type='email'
          form='login'
          className='mt-6'
          aria-invalid={Boolean(actionData?.errors?.email) || undefined}
          aria-describedby={
            actionData?.errors?.email ? 'email-error' : undefined
          }
          required
          errorMessage={actionData?.errors?.email}
        />
        <TextField
          id='password'
          label='Password'
          labelLink='Forgot password?'
          labelLinkTo={route('/recovery')}
          name='password'
          type='password'
          form='login'
          className='mt-4'
          aria-invalid={Boolean(actionData?.errors?.password) || undefined}
          aria-describedby={
            actionData?.errors?.password ? 'password-error' : undefined
          }
          required
          errorMessage={actionData?.errors?.password}
        />
      </Card>

      <div className='mb-4 mt-4 w-full'>{widget}</div>

      <Button form='login' type='submit' disabled={!token}>
        Log in
      </Button>

      <p className='text-center text-sm font-medium text-medium'>
        New to the Interledger Wallet?{' '}
        <Router className='text-primary' to={route('/signup')}>
          Sign up
        </Router>
      </p>
    </>
  )
}

export async function action({ request }: ActionFunctionArgs) {
  const form = await request.formData()
  const csrfToken = form.get('csrf_token')
  const email = form.get('email')
  const password = form.get('password')
  const flowId = form.get('flow_id')
  const returnTo = form.get('returnTo')?.toString()

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
      Accept: 'application/json',
      'Content-type': 'application/json',
      cookie: String(request.headers.get('cookie'))
    }
  })

  if (res.status >= 400 && res.status !== 422) {
    const errors = await kratosErrorMapping(res, fieldErrors)

    // In case the user clicks a link from the email or from another website,
    // our cookie will not be sent (we have SameSite=Strict). Therefore, the
    // user will be redirected to the login page. Once the users logs in,
    // and if by any chance they already have a valid session, Kratos will throw
    // an error (no error ID but ...) similar to the one for the TOTP challenge
    // when someone double submits.
    if (isSessionAlreadyExitsMessage(errors.form)) {
      return redirect(returnTo || '/')
    }
    return error(request, { errors }, { action: 'Contact support' })
  }

  // Remove all headers besides set-cookie
  const headers = trimHeaders(res.headers, ['set-cookie'])

  if (res.status === 422) {
    return redirect('/totp/challenge?returnTo=' + returnTo || '/', {
      headers
    })
  }

  try {
    const responseCopy = res.clone()
    const checkTOTP = await responseCopy.json()
    if (checkTOTP?.session?.authenticator_assurance_level === 'aal1') {
      return redirect(
        '/totp/two-factor-authentication?returnTo=' + returnTo || '/',
        {
          headers
        }
      )
    }
  } catch (error) {
    // If the response is not JSON, we can ignore it
  }

  if (returnTo) {
    return redirect(returnTo, {
      headers: headers
    })
  }

  return redirect(route('/'), {
    headers: headers
  })
}
