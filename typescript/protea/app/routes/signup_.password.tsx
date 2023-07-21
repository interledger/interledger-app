import type { SuccessfulSelfServiceRegistrationWithoutBrowser } from '@ory/kratos-client'
import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Button,
  Card,
  CardContent,
  Checkbox,
  Layouts,
  Router,
  Snackbar,
  TextField
} from '~/components'
import { exitFlow, flowType, requireFlow } from '~/lib/flows.server'
import { trimHeaders } from '~/lib/headers.server'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError,
  kratosErrorMapping,
  requireNoUserSession
} from '~/lib/kratos.server'
import { grpcClient } from '~/lib/proto.server'
import { canSignup, setWaitlistSignupComplete } from '~/lib/signupCheck.server'
import { flashSnackbar } from '~/lib/snackbar.server'

export async function loader({ request }: LoaderArgs) {
  await requireNoUserSession(request)
  await canSignup(request)
  await requireFlow(request, flowType.Signup)
  const cookie = String(request.headers.get('cookie'))

  const url = new URL(request.url)
  const kratosFlowId = url.searchParams.get('flow')

  let kratosFlow
  if (kratosFlowId) {
    // If ?flow=.. was in the URL, we fetch it
    const flowRes = await fetch(
      `${KRATOS_URL}/self-service/registration/flows?id=${kratosFlowId}`,
      {
        headers: {
          cookie: cookie,
          Accept: 'application/json'
        }
      }
    )
    kratosFlow = await flowRes.json()
    if (flowRes.status >= 400) handleFlowError(kratosFlow, 'signup')
  } else {
    // Otherwise we initialize it
    const flowRes = await fetch(
      `${KRATOS_URL}/self-service/registration/browser?${url.searchParams}`,
      { headers: { Accept: 'application/json' } }
    )
    kratosFlow = await flowRes.json()
    if (flowRes.status >= 400) handleFlowError(kratosFlow, 'signup')
    return redirect(`/signup/password?flow=${kratosFlow.id}`, {
      headers: trimHeaders(flowRes.headers, ['set-cookie'])
    })
  }
  return json({
    kratosFlowId,
    csrfToken: getCsrfTokenFromFlow(kratosFlow)
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/signup/phone'),
      title: 'Password'
    }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Sign up | Password'
  }
}

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { kratosFlowId, csrfToken } = useLoaderData<typeof loader>()

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
        id='signup-password'
        action={`/signup/password?flow=${kratosFlowId}`}
        method='post'
        className='hidden'
      />
      <Card>
        <CardContent>
          <p>Create a password to log in to your account securely.</p>
        </CardContent>
        <TextField
          id='password'
          label='Password'
          name='password'
          form='signup-password'
          type='password'
          className='mt-2'
          aria-invalid={Boolean(actionData?.errors.password) || undefined}
          aria-describedby={
            actionData?.errors.password ? 'password-error' : undefined
          }
          required
          errorMessage={actionData?.errors.password}
        />
      </Card>
      <Card>
        <CardContent>
          <Checkbox
            id='service-agreement'
            name='service-agreement'
            form='signup-password'
            className='flex'
            aria-invalid={
              Boolean(actionData?.errors.serviceAgreement) || undefined
            }
            aria-describedby={
              actionData?.errors.serviceAgreement
                ? 'serviceAgreement-error'
                : undefined
            }
            errorMessage={actionData?.errors.serviceAgreement}
          >
            I agree to the Fynbos&nbsp;
            <Router className='text-primary' to='/legal/privacy-policy'>
              Privacy Policy
            </Router>
            ,&nbsp;
            <Router className='text-primary' to='/legal/terms-of-service'>
              Terms of Use
            </Router>
            , and&nbsp;
            <Router className='text-primary' to='/legal/us/e-sign-agreement'>
              E-sign Agreement
            </Router>
            .
          </Checkbox>
        </CardContent>

        <input
          defaultValue={csrfToken}
          name='csrf_token'
          form='signup-password'
          type='hidden'
        />
      </Card>
      <Button form='signup-password' type='submit'>
        Confirm
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
    </>
  )
}

export async function action({ request }: ActionArgs) {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow') as string

  const form = await request.formData()
  const csrfToken = form.get('csrf_token') as string
  const password = form.get('password') as string
  const serviceAgreement = form.get('service-agreement') as string

  const fieldErrors = {
    form: '',
    serviceAgreement: '',
    password: ''
  }

  if (serviceAgreement == null) {
    fieldErrors.serviceAgreement = 'You are required to agree to continue.'
    return json(
      {
        errors: {
          ...fieldErrors
        }
      },
      { status: 400 }
    )
  }

  const flow = await requireFlow(request, flowType.Signup)

  const email = flow?.data.email

  const res = await fetch(
    `${KRATOS_URL}/self-service/registration?flow=${flowId}`,
    {
      method: 'POST',
      body: JSON.stringify({
        method: 'password',
        traits: {
          email: email,
          phone: flow?.data.phone,
          firstName: flow?.data.firstName,
          lastName: flow?.data.lastName,
          countryCode: flow?.data.country
        },
        password: password,
        csrf_token: csrfToken
      }),
      headers: {
        'Content-type': 'application/json',
        cookie: String(request.headers.get('cookie'))
      }
    }
  )
  if (res.status >= 400) {
    return kratosErrorMapping(res, fieldErrors)
  }

  const data = await res.json()
  // The SuccessfulSelfServiceRegistrationWithoutBrowser is correct here. The OpenAPI spec for kratos
  // has some weird naming for types....
  const successData = data as SuccessfulSelfServiceRegistrationWithoutBrowser

  // Mark signup complete
  const userId = successData.identity.id
  // TODO: also handle via kratos webhook, add retry here and error handling
  await grpcClient.completeSignup({
    id: flow.data.id,
    userId: userId
  })
  await setWaitlistSignupComplete(request, userId)
  await exitFlow(request, flowType.Signup)

  // Delete all the values from the header instead of set-cookie
  const resHeaders = trimHeaders(res.headers, ['set-cookie'])

  const sessionHeaders = await flashSnackbar(request, {
    message: 'Your account was created successfully.',
    icon: 'close'
  })
  resHeaders.append('set-cookie', sessionHeaders)

  return redirect(route('/wallet-address'), {
    headers: resHeaders
  })
}
