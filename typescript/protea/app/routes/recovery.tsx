import type { ActionFunctionArgs, LoaderFunctionArgs, MetaFunction } from 'react-router';
import { data, redirect } from 'react-router';
import { useFetcher, useLoaderData } from 'react-router';
import type { ApplicationProps } from '~/components'
import { Button, Card, CardContent, Layouts, TextField } from '~/components'
import { error } from '~/lib/error.server'
import {
  KRATOS_URL,
  getCsrfTokenFromFlow,
  handleFlowError,
  kratosErrorMapping,
  requireNoUserSession
} from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'
import { RateLimitKeys, getKey, rateLimit } from '~/lib/rateLimit.server'

type ActionResponse =
  | { success: true }
  | { errors: { form: string; email: string } }

export async function loader({ request }: LoaderFunctionArgs) {
  await requireNoUserSession(request)
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const cookie = String(request.headers.get('cookie'))

  let flow
  if (flowId) {
    // If ?flow=.. was in the URL, we fetch it
    const flowRes = await fetch(
      `${KRATOS_URL}/self-service/recovery/flows?id=${flowId}`,
      {
        headers: {
          cookie: cookie,
          Accept: 'application/json'
        }
      }
    )
    flow = await flowRes.json()
    if (flowRes.status >= 400) handleFlowError(flow, 'recovery')
  } else {
    // Otherwise we initialize it
    const flowRes = await fetch(
      `${KRATOS_URL}/self-service/recovery/browser?${url.searchParams}`,
      { headers: { Accept: 'application/json' } }
    )
    flow = await flowRes.json()
    if (flowRes.status >= 400) handleFlowError(flow, 'recovery')
    return redirect(`/recovery?flow=${flow.id}`, {
      headers: flowRes.headers
    })
  }

  return data({ flow, csrfToken: getCsrfTokenFromFlow(flow) })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      title: 'Recover account'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Recover account'
  }
])

export default function Page() {
  const { flow, csrfToken } = useLoaderData<typeof loader>()
  const fetcher = useFetcher<ActionResponse>()

  const isSubmitting = fetcher.state !== 'idle'
  const isSuccess =
    fetcher.data && 'success' in fetcher.data && fetcher.data.success
  const errors =
    fetcher.data && 'errors' in fetcher.data ? fetcher.data.errors : undefined

  return (
    <>
      <fetcher.Form
        id='recovery'
        action={`/recovery?flow=${flow.id}`}
        method='post'
        className='hidden'
      />
      <input
        form='recovery'
        defaultValue={csrfToken}
        name='csrf_token'
        type='hidden'
      />
      <Card>
        <CardContent>
          <span>
            Enter your email address and we will email you a link to change your
            password.
          </span>
        </CardContent>
        <TextField
          id='email'
          form='recovery'
          label='Email'
          name='email'
          type='email'
          className='mt-2'
          aria-invalid={Boolean(errors?.email) || undefined}
          aria-describedby={errors?.email ? 'email-error' : undefined}
          required
          errorMessage={errors?.email}
        />
      </Card>
      <Button
        form='recovery'
        type='submit'
        disabled={isSubmitting || isSuccess}
      >
        {isSubmitting ? 'Sending...' : 'Recover account'}
      </Button>
      {isSuccess && (
        <p className='mt-2 text-sm text-success'>
          Recovery email sent successfully.
        </p>
      )}
      {errors?.form && <p className='mt-2 text-sm text-error'>{errors.form}</p>}
    </>
  )
}

export async function action({ request }: ActionFunctionArgs) {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')

  const form = await request.formData()
  const csrfToken = form.get('csrf_token')
  const email = form.get('email') ?? ''
  const fieldErrors = {
    form: '',
    email: ''
  }
  const key = getKey(RateLimitKeys.RecoveryEmail, email.toString())
  const rateError = await rateLimit(key)
  if (rateError) {
    fieldErrors.form = rateError
    return error(request, { errors: fieldErrors })
  }

  const res = await fetch(
    `${KRATOS_URL}/self-service/recovery?flow=${flowId}`,
    {
      method: 'POST',
      body: JSON.stringify({
        method: 'link',
        email: email,
        csrf_token: csrfToken
      }),
      headers: {
        'Content-Type': 'application/json',
        cookie: String(request.headers.get('cookie'))
      }
    }
  )

  if (res.status >= 400) {
    const errs = await kratosErrorMapping(res, fieldErrors)
    return error(request, { errors: errs })
  }

  return data({ success: true })
}
