import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Button, Checkbox, Router, Shape, TextField } from '~/components'
import { exitFlow, flowType, getCurrentFlow } from '~/lib/flows.server'
import {
  getCsrfTokenFromFlow,
  handleFlowError,
  KRATOS_URL
} from '~/lib/kratos.server'
import { route } from 'routes-gen'

export async function loader({ request }: LoaderArgs) {
  const flow = await getCurrentFlow(request, flowType.Signup)
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
    return redirect(`/signup/${flow?.id}/password?flow=${kratosFlow.id}`, {
      headers: flowRes.headers
    })
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
    <div className='mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto rounded-2xl bg-page px-4 pb-16 pt-6 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:pt-12 xl:max-w-4xl'>
      <div className='col-span-full flex flex-col space-y-4 pb-8 sm:col-span-6 sm:col-start-2 sm:space-y-6'>
        <div className='flex justify-between'>
          <span className='font-display text-2xl font-medium'>
            Create a password
          </span>
          <div className='hidden sm:flex'>
            <Shape
              width={'w-8'}
              radius={'rounded-bl-full'}
              color={'bg-rose-400'}
            />
            <Shape
              width={'w-8'}
              radius={'rounded-tl-full'}
              color={'bg-yellow-300'}
            />
          </div>
        </div>
        <span>You will need your password to log in to your account.</span>
      </div>

      <Form
        id='signup-password'
        action={`/signup/${flow.id}/password?flow=${kratosFlowId}`}
        method='post'
        className='hidden'
      />
      <TextField
        id='password'
        label='Password'
        name='password'
        form='signup-password'
        type='password'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2'
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
        className='col-span-full mt-4 flex sm:col-span-6 sm:col-start-2'
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
      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2'>
        <Button form='signup-password' type='submit'>
          Confirm
        </Button>
      </div>
    </div>
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

  const flow = await getCurrentFlow(request, flowType.Signup)

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
          countryCode: flow?.data.country,
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

  const headers = await exitFlow(request, flowType.Signup)

  // Delete all the values from the header instead of set-cookie
  const resHeaders = res.headers
  resHeaders.forEach((value,key) => {
    if(key.toLowerCase() !== 'set-cookie') {
      resHeaders.delete(key)
    }
  })

  // Append the exitFlow set-cookie
  const flowSetCookie = headers.get('set-cookie')
  if(flowSetCookie) {
    resHeaders.append('set-cookie', flowSetCookie)
  }

  return redirect(route('/'), {
    headers: resHeaders,
  })
}
