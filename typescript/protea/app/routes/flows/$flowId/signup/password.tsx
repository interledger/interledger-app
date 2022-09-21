import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { redirect } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Button, Checkbox, Router, TextField } from '~/components'
import { exitFlow, getCurrentFlow } from '~/lib/flows.server'
import type { GrpcError } from '~/lib/proto.server'
import { httpMapping } from '~/lib/proto.server'
import { grpcClient, StatusError, isGrpcError } from '~/lib/proto.server'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError
} from '~/lib/kratos.server'
import { route } from 'routes-gen'

export async function loader({ request, params }: LoaderArgs) {
  const flow = await getCurrentFlow(request, params)
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
    return redirect(
      `/flows/${flow?.id}/signup/password?flow=${kratosFlow.id}`,
      {
        headers: flowRes.headers
      }
    )
  }
  return json({
    flow,
    kratosFlowId,
    csrfToken: getCsrfTokenFromFlow(kratosFlow)
  })
}

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { flow, kratosFlowId, csrfToken } = useLoaderData<typeof loader>()

  return (
    <>
      <div className='col-span-full flex flex-col space-y-2 pt-4 pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-display text-2xl font-medium'>
          Create a password
        </span>
        <span>You will need your password to log in to your account.</span>
      </div>
      <Form
        id='signup-password'
        action={`/flows/${flow.id}/signup/password?flow=${kratosFlowId}`}
        method='post'
        className='hidden'
      />
      <TextField
        id='password'
        label='Password'
        name='password'
        form='signup-password'
        type='password'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.errors.password) || undefined}
        aria-describedby={
          actionData?.errors.password ? 'password-error' : undefined
        }
        required
        errorMessage={actionData?.errors.password}
      />

      <Checkbox
        id='service-agreement'
        name='service-agreement'
        form='signup-password'
        className='col-span-full mt-4 flex sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.errors.serviceAgreement) || undefined}
        aria-describedby={
          actionData?.errors.serviceAgreement
            ? 'serviceAgreement-error'
            : undefined
        }
        errorMessage={actionData?.errors.serviceAgreement}
      >
        I agree to the Fynbos&nbsp;
        <Router className='text-primary' to='/privacy-policy'>
          Privacy Policy
        </Router>
        ,&nbsp;
        <Router className='text-primary' to='/privacy-policy'>
          Consent to Electronic Disclosures
        </Router>
        ,&nbsp;
        <Router className='text-primary' to='/privacy-policy'>
          Deposit Terms &amp; Conditions
        </Router>
        ,&nbsp;
        <Router className='text-primary' to='/privacy-policy'>
          Client Terms of Service
        </Router>
      </Checkbox>

      <input
        defaultValue={csrfToken}
        name='csrf_token'
        form='signup-password'
        type='hidden'
      />
      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Button form='signup-password' type='submit'>
          Confirm
        </Button>
      </div>
    </>
  )
}

// The field names given by the backend for field violations
type fieldErrorsMap = 'ServiceAgreement'

function mapper(field: fieldErrorsMap): 'serviceAgreement' | null {
  switch (field) {
    case 'ServiceAgreement':
      return 'serviceAgreement'
    default:
      return null
  }
}

export async function action({ request, params }: ActionArgs) {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow') as string

  const form = await request.formData()
  const csrfToken = form.get('csrf_token') as string
  const password = form.get('password') as string
  const serviceAgreement = form.get('service-agreement') as string

  const fieldErrors = {
    serviceAgreement: '',
    password: ''
  }

  if (serviceAgreement == null) {
    fieldErrors.serviceAgreement = 'You are required to accept this to proceed.'
    return json(
      {
        errors: {
          ...fieldErrors
        }
      },
      { status: 400 }
    )
  }

  const flow = await getCurrentFlow(request, params)
  const onboardingId = flow?.data.id
  const email = flow?.data.email

  // TODO replace with service agreement service.
  let response = await grpcClient
    .updateOnboarding({
      id: onboardingId,
      serviceAgreement: true
    })
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(response)) {
    if (response.code == 3) {
      for (let violation of (response as GrpcError).details[0]
        .fieldViolations) {
        const field = mapper(violation.field as fieldErrorsMap)
        if (field != null) fieldErrors[field] = violation.description
      }
      return json({ errors: { ...fieldErrors } }, { status: 400 })
    } else throw json({}, httpMapping(response.code))
  }

  const res = await fetch(
    `${KRATOS_URL}/self-service/registration?flow=${flowId}`,
    {
      method: 'POST',
      body: JSON.stringify({
        method: 'password',
        traits: {
          email: email
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

  const data = await res.json()

  if (res.status >= 400) {
    for (let node of data.ui.nodes) {
      if (node.messages.length > 0) {
        Object.assign(fieldErrors, {
          [node.attributes.name]: node.messages[0].text
        })
      }
    }
    return json({ errors: { ...fieldErrors } }, { status: 400 })
  }

  const headers = await exitFlow(request)
  const flowSettings = headers.get('Set-Cookie') as string

  res.headers.append('Set-Cookie', flowSettings)
  return redirect(route('/onboarding/unit'), {
    headers: res.headers
  })
}
