import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import { Button, Logo, Router } from '~/components'
import { route } from 'routes-gen'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError
} from '~/lib/kratos.server'
import type {
  InitiateOnboardingMutation,
  InitiateOnboardingMutationVariables
} from '~/generated/types'
import { InitiateOnboardingDocument } from '~/generated/types'
import type { Session } from '@ory/kratos-client'
import { apolloClient } from '~/lib/apollo.server'

type ActionData = {
  formError?: string
  fieldErrors?: {
    email?: string
  }
  fields?: {
    email: string
    csrf_token: string
  }
}

const badRequest = (data: ActionData) => json(data, { status: 400 })

export async function action({ request }: ActionArgs) {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const form = await request.formData()
  const csrfToken = form.get('csrf_token')
  const email = form.get('email')

  if (typeof csrfToken !== 'string' || typeof email !== 'string') {
    return badRequest({
      formError: `Form not submitted correctly.`
    })
  }

  const fields = { csrf_token: csrfToken, email }
  const res = await fetch(
    `${KRATOS_URL}/self-service/verification?flow=${flowId}`,
    {
      method: 'POST',
      body: JSON.stringify({
        method: 'link',
        email,
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
    let fieldErrors: ActionData['fieldErrors'] = {}
    for (let node of data.ui.nodes) {
      if (node.messages.length > 0) {
        Object.assign(fieldErrors, {
          [node.attributes.name]: node.messages[0].text
        })
      }
    }
    return badRequest({ fieldErrors: fieldErrors, fields })
  }
  return redirect(route('/verify'), {
    headers: res.headers
  })
}

export async function loader({ request }: LoaderArgs) {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const cookie = String(request.headers.get('cookie'))

  const session = await fetch(`${KRATOS_URL}/sessions/whoami`, {
    headers: request.headers
  })

  switch (session.status) {
    case 401:
    case 500:
      throw redirect(route('/login'))
    case 403:
    case 422: // Need to complete 2FA.
      throw redirect(route('/login') + '?aal=aal2')
  }

  const userSession: Session = await session.json()
  if (session.status >= 400) handleFlowError(session, 'verify')

  const onboarding = await apolloClient
    .mutate<InitiateOnboardingMutation, InitiateOnboardingMutationVariables>({
      mutation: InitiateOnboardingDocument,
      context: {
        headers: {
          cookie: cookie
        }
      }
    })
    .then((val) => val.data?.initiateOnboarding)
  console.log(onboarding)

  // Check the user has at least one verifiable address.
  if (!userSession.identity.verifiable_addresses)
    return redirect(route('/signup'))
  // We currently only allow one email per user.
  if (userSession.identity.verifiable_addresses[0].verified) {
    if (onboarding?.success)
      return redirect(onboarding?.providerOnboarding.formUrl)
    return redirect(route('/home'))
  }

  // Ensure any redirects are thrown
  if (userSession instanceof Response) return session

  let flow
  if (flowId) {
    // If ?flow=.. was in the URL, we fetch it
    const flowRes = await fetch(
      `${KRATOS_URL}/self-service/verification/flows?id=${flowId}`,
      {
        headers: {
          cookie: cookie,
          Accept: 'application/json'
        }
      }
    )
    flow = await flowRes.json()
    if (flowRes.status >= 400) handleFlowError(flow, 'verify')
  } else {
    // Otherwise we initialize it
    const flowRes = await fetch(
      `${KRATOS_URL}/self-service/verification/browser?${url.searchParams}`,
      { headers: { cookie: cookie, Accept: 'application/json' } }
    )
    flow = await flowRes.json()
    if (flowRes.status >= 400) handleFlowError(flow, 'verify')
    return redirect(`/verify?flow=${flow.id}`, {
      headers: flowRes.headers
    })
  }
  return json({
    flow,
    email: userSession.identity.verifiable_addresses[0].value,
    csrfToken: getCsrfTokenFromFlow(flow)
  })
}

export default function Page() {
  const { flow, email, csrfToken } = useLoaderData<typeof loader>()

  return (
    <main className='mx-auto grid min-h-screen w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 lg:content-center xl:max-w-4xl'>
      <div className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Router to={route('/')}>
          <Logo className='h-8' />
        </Router>
      </div>
      <div className='col-span-full pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <h1 className='font-display text-4xl font-medium leading-normal text-strong'>
          Verify your email
        </h1>
      </div>
      <div className='col-span-full pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <p className='text-medium'>
          We’ve sent a verification link to your email:
          <br /> {email}
        </p>
      </div>
      {/* Form */}
      <Form
        action={`/verify?flow=${flow.id}`}
        method='post'
        className='col-span-full flex flex-col items-end space-y-2 sm:col-span-6 sm:col-start-2 lg:col-start-4'
      >
        <input defaultValue={csrfToken} name='csrf_token' type='hidden' />
        <input defaultValue={email} name='email' type='hidden' />
        <div className='pt-4'>
          <Button type='submit'>Resend verification</Button>
        </div>
      </Form>
    </main>
  )
}
