import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Button,
  Card,
  CardContent,
  Icon,
  Layouts,
  TextField
} from '~/components'
import { Label } from '~/components/Label'
import { validateCSRFToken } from '~/lib/csrf.server'
import { KRATOS_URL } from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'

export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const returnTo = url.searchParams.get('redirectTo')
  const cookie = String(request.headers.get('cookie'))

  if (!flowId) {
    const initRes = await fetch(
      `${KRATOS_URL}/self-service/login/browser?aal=aal2`,
      {
        headers: {
          cookie
        },
        redirect: 'manual'
      }
    )

    if (initRes.status !== 303 && initRes.status !== 302) {
      throw new Error('Expected redirect response from Kratos')
    }

    const redirectTo = initRes.headers.get('location')
    if (!redirectTo) {
      throw new Error('Expected redirect with flow ID, but got none.')
    }

    const flowFromRedirect = new URL(redirectTo).searchParams.get('flow')
    if (!flowFromRedirect) {
      throw new Error('Redirect did not include flow parameter')
    }

    return redirect(
      `/totp/challenge?flow=${flowFromRedirect}&returnTo=${returnTo ?? ''}`
    )
  }

  const flowRes = await fetch(
    `${KRATOS_URL}/self-service/login/flows?id=${flowId}`,
    {
      headers: {
        cookie: request.headers.get('cookie') ?? '',
        Accept: 'application/json'
      }
    }
  )

  if (!flowRes.ok) {
    return redirect('/error')
  }

  const flow = await flowRes.json()

  const csrfNode = flow.ui.nodes.find(
    (n: any) => n.attributes?.name === 'csrf_token'
  )
  const csrfToken = csrfNode?.attributes?.value

  return json({ flowId, csrfToken })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/'),
      title: 'Enter Authenticator Code'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Enter TOTP Code'
  }
])

export default function Page() {
  const { flowId, csrfToken } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()

  return (
    <>
      <Form method='post'>
        <input type='hidden' name='csrf_token' value={csrfToken} />
        <input type='hidden' name='flow' value={flowId} />
        <Card>
          <CardContent>
            <p>
              Enter the 6-digit code from your authenticator app to continue.
            </p>
          </CardContent>
          <Label className='mt-2'>Authenticator App</Label>
          <div className='mt-1 flex space-x-2 rounded-xl bg-nav p-3 text-medium'>
            <Icon>shield</Icon>
            <span>Time-based One-Time Password (TOTP)</span>
          </div>
          <TextField
            id='totp'
            label='Authenticator Code'
            name='totp_code'
            type='text'
            className='mt-4'
            aria-invalid={Boolean(actionData?.errors?.totp_code) || undefined}
            aria-describedby={
              actionData?.errors?.totp_code ? 'totp-error' : undefined
            }
            required
            errorMessage={actionData?.errors?.totp_code}
          />
        </Card>
        <Button type='submit' className='mt-4'>
          Verify
        </Button>
      </Form>
    </>
  )
}

export async function action({ request }: ActionFunctionArgs) {
  const form = await request.formData()
  const flow = form.get('flow')
  const totp_code = form.get('totp_code')
  const csrf_token = form.get('csrf_token')

  await validateCSRFToken(request, form)

  const res = await fetch(`${KRATOS_URL}/self-service/login?flow=${flow}`, {
    method: 'POST',
    headers: {
      Accept: 'application/json',
      'Content-type': 'application/json',
      cookie: String(request.headers.get('cookie'))
    },
    body: JSON.stringify({
      method: 'totp',
      totp_code,
      csrf_token
    })
  })

  if (res.status === 400) {
    const data = await res.json()
    return json({
      errors: {
        totp_code:
          data.ui?.messages?.find((m: any) => m.type === 'error')?.text ||
          'Invalid code'
      }
    })
  }

  if (!res.ok) {
    throw new Response('Unexpected error', { status: res.status })
  }

  const redirectTo = new URL(request.url).searchParams.get('returnTo') || '/'
  return redirect(redirectTo, {
    headers: request.headers
  })
}
