import type { Route } from './+types/login'
import { data, redirect, href } from 'react-router'
import {
  Form,
  useActionData,
  useLoaderData,
  useSearchParams
} from 'react-router'
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
import {
  kratosPublic
} from '~/lib/kratos/kratos-client.server'
import {
  getCookie,
  isSessionAlreadyExistsMessage,
  buildHeadersWithCookies,
  withCookie
} from '~/lib/kratos/cookie.server'
import { getCsrfTokenFromFlow } from '~/lib/kratos/flow.server'
import { mapFlowToFieldErrors, handleFlowError } from '~/lib/kratos/error.server'
import { flashSnackbar } from '~/lib/snackbar.server'
import { type KratosError } from '~/lib/kratos/types.server'
import { requireNoUserSession } from '~/lib/kratos/session.server'
import logger from '~/lib/logger.server'
import { mergeMeta } from '~/lib/meta'
import { safeReturnTo } from '~/lib/url.server'

export async function loader({ request }: Route.LoaderArgs) {
  await requireNoUserSession(request)
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const cookie = getCookie(request)

  try {
    let flow
    let responseHeaders: Headers | undefined

    if (flowId) {
      // If ?flow=.. was in the URL, we fetch it
      const { data: flowData } = await kratosPublic.getLoginFlow({ id: flowId, cookie })
      flow = flowData
    } else {
      // Otherwise we initialize it
      const returnTo = safeReturnTo(url.searchParams.get('returnTo'))
      const aal = url.searchParams.get('aal') as 'aal1' | 'aal2' | undefined
      const refresh = url.searchParams.get('refresh') === 'true'

      const response = await kratosPublic.createBrowserLoginFlow({
        returnTo,
        aal,
        refresh
      })
      flow = response.data

      responseHeaders = buildHeadersWithCookies(response)
    }

    return data(
      {
        flowId: flow.id,
        csrfToken: getCsrfTokenFromFlow(flow)
      },
      {
        headers: responseHeaders
      }
    )
  } catch (error) {
    handleFlowError(error, 'login')
    logger.error({ error, route: 'login' }, 'Failed to load login flow')
    throw new Error('Failed to load login flow')
  }
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {}
  }
}

export const meta = mergeMeta(() => [
  {
    title: 'Log in'
  }
])

export default function Page() {
  const actionData = useActionData()
  const { csrfToken, flowId } = useLoaderData()
  const searchParams = useSearchParams()

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
          labelLinkTo={href('/recovery')}
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
      <Button form='login' type='submit'>
        Log in
      </Button>
      <p className='text-center text-sm font-medium text-medium'>
        New to the Interledger Wallet?{' '}
        <Router className='text-primary' to={href('/signup')}>
          Sign up
        </Router>
      </p>
    </>
  )
}

export async function action({ request }: Route.ActionArgs) {
  const form = await request.formData()
  const csrfToken = form.get('csrf_token')
  const email = form.get('email')
  const password = form.get('password')
  const flowId = form.get('flow_id')
  const cookie = getCookie(request)

  if (!flowId) {
    return redirect('/login')
  }

  // Theoretically, this should 100% safe, since in Remix `request.url` should
  // always be a valid URL.
  const requestUrl = new URL(request.url)
  const returnTo = safeReturnTo(requestUrl.searchParams.get('returnTo'))
  const searchParams = new URLSearchParams()
  searchParams.set('returnTo', returnTo)

  try {
    const response = await kratosPublic.updateLoginFlow({
      flow: flowId as string,
      updateLoginFlowBody: {
        method: 'password' as const,
        identifier: email as string,
        password: password as string,
        csrf_token: csrfToken as string
      },
    },
      withCookie(cookie)
    )

    const headers = buildHeadersWithCookies(response)

    // Check for AAL1 (needs TOTP setup)
    if (response.data.session?.authenticator_assurance_level === 'aal1') {
      return redirect(
        `${href('/totp/two-factor-authentication')}?${searchParams.toString()}`,
        { headers }
      )
    }

    return redirect(returnTo, { headers })
  } catch (err) {
    const errResponse = (err as KratosError).response

    // Handle validation errors
    if (errResponse.status === 400) {
      const flowData = errResponse.data
      const fieldErrors = {
        form: '',
        email: '',
        password: ''
      }
      const errors = mapFlowToFieldErrors(flowData, fieldErrors)

      // In case the user clicks a link from the email or from another website,
      // our cookie will not be sent (we have SameSite=Strict). Therefore, the
      // user will be redirected to the login page. Once the users logs in,
      // and if by any chance they already have a valid session, Kratos will throw
      // an error (no error ID but ...) similar to the one for the TOTP challenge
      // when someone double submits.
      if (isSessionAlreadyExistsMessage(errors.form)) {
        return redirect(returnTo)
      }

      // Redirect with snackbar instead of returning error codes — we need fresh
      // flow and CSRF token on retry, otherwise we have stale data in the loader.
      const redirectParams = new URLSearchParams(searchParams)
      redirectParams.set('flow', flowId as string)
      const snackbarCookie = await flashSnackbar(request, {
        message: errors.form || 'An error occurred, please retry.',
        icon: 'close',
        action: 'Contact support'
      })
      const redirectHeaders = buildHeadersWithCookies(errResponse)
      redirectHeaders.append('Set-Cookie', snackbarCookie)
      return redirect(`/login?${redirectParams.toString()}`, { headers: redirectHeaders })
    }

    // Handle AAL2 required
    if (errResponse.status === 422) {
      const headers = buildHeadersWithCookies(errResponse)
      return redirect(
        `${href('/totp/challenge')}?${searchParams.toString()}`,
        { headers }
      )
    }

    logger.error({ error: err, route: 'login' }, 'Failed to submit login')
    throw new Error('Failed to submit login')
  }
}
