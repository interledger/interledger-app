import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import {
  Form,
  useActionData,
  useLoaderData,
  useSearchParams
} from '@remix-run/react'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import {
  Button,
  Card,
  Layouts,
  Router,
  Snackbar,
  TextField
} from '~/components'
import { trimHeaders } from '~/lib/headers.server'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError,
  kratosErrorMapping,
  requireNoUserSession
} from '~/lib/kratos.server'
import { IS_SIGNUP_GATED } from '~/lib/signupCheck.server'

export async function loader({ request }: LoaderArgs) {
  await requireNoUserSession(request)
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const cookie = String(request.headers.get('cookie'))

  let flow
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
      `${KRATOS_URL}/self-service/login/browser?${url.searchParams}`,
      { headers: { Accept: 'application/json' } }
    )
    flow = await flowRes.json()
    if (flowRes.status >= 400) handleFlowError(flow, 'login')
    url.searchParams.set('flow', flow.id)
    return redirect(`/login?${url.searchParams}`, {
      headers: trimHeaders(flowRes.headers, ['set-cookie'])
    })
  }
  return json({
    csrfToken: getCsrfTokenFromFlow(flow),
    isSignupGated: IS_SIGNUP_GATED
  })
}

export const handle = {
  title: 'Log in',
  layout: Layouts.FocusLayout
}

export const meta: MetaFunction = () => {
  return {
    title: 'Log in'
  }
}

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { csrfToken, isSignupGated } = useLoaderData<typeof loader>()
  const searchParams = useSearchParams()

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
        id='login'
        action={`/login?${searchParams[0]}`}
        method='post'
        className='hidden'
      />
      <Card>
        <TextField
          id='email'
          label='Email'
          name='email'
          type='email'
          form='login'
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
          name='password'
          type='password'
          form='login'
          className='mt-1'
          aria-invalid={Boolean(actionData?.errors?.password) || undefined}
          aria-describedby={
            actionData?.errors?.password ? 'password-error' : undefined
          }
          required
          errorMessage={actionData?.errors?.password}
        />

        <input
          form='login'
          defaultValue={csrfToken}
          name='csrf_token'
          type='hidden'
        />
        <div className='mt-4'>
          <Router to={route('/recovery')} aria-label='Forgot password?'>
            <span className='text-sm font-medium text-primary'>
              Forgot password
            </span>
          </Router>
        </div>
      </Card>
      <Button form='login' type='submit'>
        Log in
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
      <p className='text-center text-sm font-medium text-medium'>
        New to Fynbos?{' '}
        {isSignupGated && (
          <Router className='text-primary' to={route('/waitlist')}>
            Join waitlist
          </Router>
        )}
        {!isSignupGated && (
          <Router className='text-primary' to={route('/signup')}>
            Sign up
          </Router>
        )}
      </p>
    </>
  )
}

export async function action({ request }: ActionArgs) {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const returnTo = url.searchParams.get('return_to')

  const form = await request.formData()
  const csrfToken = form.get('csrf_token')
  const email = form.get('email')
  const password = form.get('password')

  const fieldErrors = {
    form: '',
    email: '',
    password: ''
  }

  const res = await fetch(`${KRATOS_URL}/self-service/login?flow=${flowId}`, {
    method: 'POST',
    body: JSON.stringify({
      method: 'password',
      password_identifier: email,
      password: password,
      csrf_token: csrfToken
    }),
    headers: {
      'Content-type': 'application/json',
      cookie: String(request.headers.get('cookie'))
    }
  })
  if (res.status >= 400) {
    return kratosErrorMapping(res, fieldErrors)
  }

  // Remove all headers besides set-cookie
  const headers = trimHeaders(res.headers, ['set-cookie'])

  if (returnTo) {
    return redirect(returnTo, {
      headers: headers
    })
  }
  return redirect(route('/'), {
    headers: headers
  })
}
