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
import {
  kratosPublic,
  getCookie,
  getCsrfTokenFromFlow,
  handleFlowError,
  isSessionAlreadyExistsMessage,
  mapFlowToFieldErrors,
  buildHeadersWithCookies,
  withCookie,
  type KratosError
} from '~/lib/kratos-client.server'
import { requireNoUserSession } from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'
import { safeReturnTo } from '~/lib/url.server'

export async function loader({ request }: LoaderFunctionArgs) {
  await requireNoUserSession(request)
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const cookie = getCookie(request)

  try {
    let flow
    let responseHeaders: Headers | undefined

    if (flowId) {
      // If ?flow=.. was in the URL, we fetch it
      const { data } = await kratosPublic.getLoginFlow({ id: flowId, cookie })
      flow = data
    } else {
      // Otherwise we initialize it
      const returnTo = url.searchParams.get('return_to') ?? undefined
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

    return json(
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
    throw error
  }
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
  const { csrfToken, flowId } = useLoaderData<typeof loader>()
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
      <Button form='login' type='submit'>
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
        `${route('/totp/two-factor-authentication')}?${searchParams.toString()}`,
        { headers }
      )
    }

    return redirect(returnTo || '/', { headers })
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
        return redirect(returnTo || '/')
      }
      return json({ errors }, { status: 400, headers: errResponse.headers as HeadersInit })
    }

    // Handle AAL2 required
    if (errResponse.status === 422) {
      const headers = buildHeadersWithCookies(errResponse)
      return redirect(
        `${route('/totp/challenge')}?${searchParams.toString()}`,
        { headers }
      )
    }

    throw err
  }
}
