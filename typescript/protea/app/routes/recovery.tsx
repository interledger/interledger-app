import { data, redirect, useFetcher, useLoaderData } from 'react-router'
import type { ApplicationProps } from '~/components'
import { Button, Card, CardContent, Layouts, TextField } from '~/components'
import { error } from '~/lib/error.server'
import {
  buildHeadersWithCookies,
  getCookie,
  withCookie
} from '~/lib/kratos/cookie.server'
import {
  handleFlowError,
  mapFlowToFieldErrors
} from '~/lib/kratos/error.server'
import { getCsrfTokenFromFlow } from '~/lib/kratos/flow.server'
import { kratosPublic } from '~/lib/kratos/kratos-client.server'
import { requireNoUserSession } from '~/lib/kratos/session.server'
import type { KratosError } from '~/lib/kratos/types.server'
import logger from '~/lib/logger.server'
import { mergeMeta } from '~/lib/meta'
import { RateLimitKeys, getKey, rateLimit } from '~/lib/rateLimit.server'
import { safeReturnTo } from '~/lib/url.server'
import type { Route } from './+types/recovery'

type ActionResponse =
  | { success: true }
  | { errors: { form: string; email: string } }

export async function loader({ request }: Route.LoaderArgs) {
  await requireNoUserSession(request)
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const cookie = getCookie(request)

  if (flowId) {
    try {
      const { data: flow } = await kratosPublic.getRecoveryFlow(
        { id: flowId },
        withCookie(cookie)
      )
      return data({ flow, csrfToken: getCsrfTokenFromFlow(flow) })
    } catch (err: any) {
      const errResponse = (err as KratosError).response
      if (errResponse.status != 410) {
        handleFlowError(err, 'recovery')
        logger.error(
          { error: err, route: 'recovery' },
          'Failed to load recovery flow'
        )
        throw new Error('Failed to load recovery flow')
      }
      // 410 - generate a new flow
    }
  }

  // Initialize new recovery flow
  try {
    const response = await kratosPublic.createBrowserRecoveryFlow(
      { returnTo: safeReturnTo(url.searchParams.get('returnTo')) },
      withCookie(cookie)
    )
    return redirect(`/recovery?flow=${response.data.id}`, {
      headers: buildHeadersWithCookies(response)
    })
  } catch (err: any) {
    handleFlowError(err, 'recovery')
    logger.error(
      { error: err, route: 'recovery' },
      'Failed to initialize recovery flow'
    )
    throw new Error('Failed to initialize recovery flow')
  }
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      title: 'Recover account'
    }
  }
}

export const meta = mergeMeta(() => [
  {
    title: 'Recover account'
  }
])

export default function Page() {
  const { flow, csrfToken } = useLoaderData()
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

export async function action({ request }: Route.ActionArgs) {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const form = await request.formData()
  const csrfToken = form.get('csrf_token')
  const email = form.get('email') ?? ''

  if (!flowId || flowId === '' || !csrfToken) {
    throw redirect('/recovery')
  }

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

  const cookie = getCookie(request)

  try {
    await kratosPublic.updateRecoveryFlow(
      {
        flow: flowId!,
        updateRecoveryFlowBody: {
          method: 'link',
          email: email as string,
          csrf_token: csrfToken as string
        }
      },
      withCookie(cookie)
    )

    return data({ success: true })
  } catch (err: any) {
    const errResponse = (err as KratosError).response
    const flowData = errResponse.data
    const errs = mapFlowToFieldErrors(flowData, fieldErrors)
    return error(request, { errors: errs })
  }
}
