import {
  Form,
  data,
  href,
  redirect,
  useActionData,
  useLoaderData
} from 'react-router'
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
import type { KratosError } from '~/lib/kratos/types.server'
import logger from '~/lib/logger.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import type { Route } from './+types/settings_.password'

export async function loader({ request }: Route.LoaderArgs) {
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  const cookie = getCookie(request)

  if (flowId) {
    try {
      const { data: flow } = await kratosPublic.getSettingsFlow(
        { id: flowId },
        withCookie(cookie)
      )
      return data({
        flow,
        csrfToken: getCsrfTokenFromFlow(flow)
      })
    } catch (err: any) {
      handleFlowError(err, 'settings/password')
      logger.error(
        { error: err, route: 'settings.password' },
        'Failed to load password settings flow'
      )
      throw new Error('Failed to load password settings flow')
    }
  }

  // Initialize new settings flow
  try {
    const response = await kratosPublic.createBrowserSettingsFlow(
      { returnTo: url.searchParams.get('returnTo') ?? undefined },
      withCookie(cookie)
    )
    return redirect(`/settings/password?flow=${response.data.id}`, {
      headers: buildHeadersWithCookies(response)
    })
  } catch (err: any) {
    handleFlowError(err, 'settings/password')
    logger.error(
      { error: err, route: 'settings.password' },
      'Failed to initialize password settings flow'
    )
    throw new Error('Failed to initialize password settings flow')
  }
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: href('/settings'),
      title: 'Set password'
    }
  }
}

export const meta = mergeMeta(() => [
  {
    title: 'Set password'
  }
])

export default function Page() {
  const actionData = useActionData()
  const { flow, csrfToken } = useLoaderData()

  return (
    <>
      <Form
        id='settings-password'
        action={`/settings/password?flow=${flow.id}`}
        method='post'
        className='hidden'
      />
      <input
        form='settings-password'
        defaultValue={csrfToken}
        name='csrf_token'
        type='hidden'
      />
      <Card>
        <CardContent>
          <p>Set a new password to continue.</p>
        </CardContent>
        <TextField
          id='new-password'
          form='settings-password'
          label='New password'
          name='new-password'
          type='password'
          className='mt-2'
          aria-invalid={Boolean(actionData?.errors?.password) || undefined}
          aria-describedby={
            actionData?.errors?.password ? 'password-error' : undefined
          }
          required
          errorMessage={actionData?.errors?.password}
        />
      </Card>
      <Button form='settings-password' type='submit'>
        Continue
      </Button>
    </>
  )
}

export async function action({ request }: Route.ActionArgs) {
  const cookie = getCookie(request)
  const url = new URL(request.url)
  const flowId = url.searchParams.get('flow')
  if (!flowId) return redirect('/settings/password')

  const form = await request.formData()
  const csrfToken = form.get('csrf_token') as string
  const password = form.get('new-password') as string

  const fieldErrors = {
    form: '',
    password: ''
  }

  try {
    await kratosPublic.updateSettingsFlow(
      {
        flow: flowId,
        updateSettingsFlowBody: {
          method: 'password',
          password,
          csrf_token: csrfToken
        }
      },
      withCookie(cookie)
    )
    return redirectWithSnackbar(request, href('/settings'), {
      message: 'New password successfully saved.',
      icon: 'close'
    })
  } catch (err: any) {
    const errResponse = (err as KratosError).response
    const status = errResponse?.status
    const flowData = errResponse?.data
    if (status === 400) {
      const errs = mapFlowToFieldErrors(flowData, fieldErrors)
      return error(request, { errors: errs })
    }
    handleFlowError(err, 'settings/password')
    logger.error(
      { error: err, route: 'settings.password' },
      'Failed to update password'
    )
    throw new Error('Failed to update password')
  }
}
