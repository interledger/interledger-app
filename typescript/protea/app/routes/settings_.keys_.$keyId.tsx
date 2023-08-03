import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Layouts,
  OutlineButton,
  TextField
} from '~/components'
import { Code } from '~/generated/protobuf-ts/google/rpc/code'
import { getConnection, getConnectionLimits } from '~/lib/connections.server'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import type { GrpcError } from '~/lib/proto.server'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: '/settings/keys',
      title: (match) => match.data.connection.applicationName
    }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Key'
  }
}

export async function loader({ request, params }: LoaderArgs) {
  let data = await Promise.all([
    getConnection(request, params.keyId as string),
    getConnectionLimits(request, params.keyId as string)
  ])
  return jsonWithCSRF(request, {
    connection: data[0],
    limits: data[1]
  })
}

export default function Page() {
  const { connection, limits, csrfToken } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()

  return (
    <>
      <Form
        id='key-id'
        action={`/settings/keys/${connection.id}`}
        method='post'
        className='hidden'
      />
      <input form='key-id' value={csrfToken} name='csrfToken' type='hidden' />
      <Card>
        <CardContent>
          <code className='flex items-center justify-between break-all rounded-xl bg-nav p-2 font-mono text-medium'>
            {connection.publicKeyFingerprint}
          </code>

          <p className='mt-4 text-sm text-medium'>
            Added {connection.createdAt}
          </p>
          {/*TODO: implement last used*/}
          {/*<p className='mt-1 text-sm text-purple-500'>Last used {connection.lastUsedAt}</p>*/}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Limits</CardTitle>
        </CardHeader>
        <CardContent>
          <p>
            Providing access to your Fynbos wallet allows the external
            application to make payments. Set the limits below.
          </p>
        </CardContent>
        <TextField
          id='dailyLimit'
          label='Daily'
          name='dailyLimit'
          form='key-id'
          type='number'
          min='0'
          step='0.01'
          defaultValue={limits.daily}
          prefix='$'
          className='mt-2'
          aria-invalid={Boolean(actionData?.errors.dailyLimit) || undefined}
          aria-describedby={
            actionData?.errors.dailyLimit ? 'dailyLimit-error' : undefined
          }
          required
          errorMessage={actionData?.errors.dailyLimit}
        />

        <TextField
          id='monthlyLimit'
          label='Monthly'
          name='monthlyLimit'
          form='key-id'
          type='number'
          min='0'
          step='0.01'
          defaultValue={limits.monthly}
          prefix='$'
          className='mt-4'
          aria-invalid={Boolean(actionData?.errors.monthlyLimit) || undefined}
          aria-describedby={
            actionData?.errors.monthlyLimit ? 'monthlyLimit-error' : undefined
          }
          required
          errorMessage={actionData?.errors.monthlyLimit}
        />

        <TextField
          id='overallLimit'
          label='Overall'
          name='overallLimit'
          form='key-id'
          type='number'
          min='0'
          step='0.01'
          defaultValue={limits.overall}
          prefix='$'
          className='mt-4'
          aria-invalid={Boolean(actionData?.errors.overallLimit) || undefined}
          aria-describedby={
            actionData?.errors.overallLimit ? 'overallLimit-error' : undefined
          }
          required
          errorMessage={actionData?.errors.overallLimit}
        />
      </Card>

      <div className='flex w-full space-x-2'>
        <OutlineButton
          shrink
          // TODO error token colors
          className='text-red-700 outline-red-700 focus-visible:outline-red-800'
          form='key-id'
          name='formName'
          value='delete'
          type='submit'
        >
          Delete
        </OutlineButton>
        <Button
          className='col-span-2'
          form='key-id'
          name='formName'
          value='update'
          type='submit'
        >
          Save
        </Button>
      </div>
    </>
  )
}

// The field names given by the backend for field violations
type fieldErrorsMap = 'DailyLimit' | 'MonthlyLimit' | 'OverallLimit'

function mapper(
  field: fieldErrorsMap
): 'dailyLimit' | 'monthlyLimit' | 'overallLimit' | null {
  switch (field) {
    case 'DailyLimit':
      return 'dailyLimit'
    case 'MonthlyLimit':
      return 'monthlyLimit'
    case 'OverallLimit':
      return 'overallLimit'
    default:
      return null
  }
}

export async function action({ request, params }: ActionArgs) {
  const form = await request.formData()
  const formName = await form.get('formName')

  await validateCSRFToken(request, form)

  if (formName === 'delete') {
    const response = await grpcClient
      .deleteConnection(
        { id: params.keyId as string },
        {
          meta: {
            cookies: String(request.headers.get('cookie')) || ''
          }
        }
      )
      .then((v) => v)
      .catch((e) => {
        console.log(e)
        return StatusError(e)
      })
    if (isGrpcError(response)) {
      throw json({}, httpMapping(response.code))
    }

    return redirectWithSnackbar(request, route('/settings/keys'), {
      message: 'Public key was deleted.',
      icon: 'close'
    })
  }

  const fieldErrors = {
    dailyLimit: '',
    monthlyLimit: '',
    overallLimit: ''
  }

  const response = await grpcClient
    .updateConnectionLimits(
      {
        id: params.keyId as string,
        daily: {
          amount: String(
            Math.floor(parseFloat(form.get('dailyLimit') as string) * 100)
          ),
          asset: 'USD',
          assetScale: 2
        },
        monthly: {
          amount: String(
            Math.floor(parseFloat(form.get('monthlyLimit') as string) * 100)
          ),
          asset: 'USD',
          assetScale: 2
        },
        overall: {
          amount: String(
            Math.floor(parseFloat(form.get('overallLimit') as string) * 100)
          ),
          asset: 'USD',
          assetScale: 2
        }
      },
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    if (response.code == Code.INVALID_ARGUMENT) {
      for (let violation of (response as GrpcError).details[0]
        .fieldViolations) {
        const field = mapper(violation.field as fieldErrorsMap)
        if (field != null) fieldErrors[field] = violation.description
      }
      return json({ errors: { ...fieldErrors } }, { status: 400 })
    } else throw json({}, httpMapping(response.code))
  }

  return json({ errors: { ...fieldErrors } })
}
