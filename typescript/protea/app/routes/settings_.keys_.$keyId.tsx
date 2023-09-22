import { Code } from '@bufbuild/connect'
import type { ActionArgs, LoaderArgs, V2_MetaFunction } from '@remix-run/node'
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
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
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

export const meta: V2_MetaFunction = mergeMeta(() => [
  {
    title: 'Key'
  }
])

export async function loader({ request, params }: LoaderArgs) {
  const connection = await grpc.getConnection(request, {
    id: params.keyId as string
  })

  if (isConnectError(connection)) throw connection.errorResponse

  const response = await grpc.getConnectionLimits(request, {
    id: params.keyId as string
  })

  if (isConnectError(response)) throw response.errorResponse

  const limits = {
    daily:
      Number(response.daily?.amount) *
      Math.pow(10, -(response.daily?.assetScale || 0)),
    monthly:
      Number(response.monthly?.amount) *
      Math.pow(10, -(response.monthly?.assetScale || 0)),
    overall:
      Number(response.overall?.amount) *
      Math.pow(10, -(response.overall?.assetScale || 0))
  }

  return jsonWithCSRF(request, {
    connection,
    limits
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

export async function action({ request, params }: ActionArgs) {
  const form = await request.formData()
  const formName = form.get('formName')

  await validateCSRFToken(request, form)

  const errors = {
    form: '',
    dailyLimit: '',
    monthlyLimit: '',
    overallLimit: ''
  }

  if (formName === 'delete') {
    const response = await grpc.deleteConnection(request, {
      id: params.keyId as string
    })

    if (isConnectError(response)) {
      return response.error({ errors })
    }

    return redirectWithSnackbar(request, route('/settings/keys'), {
      message: 'Public key was deleted.',
      icon: 'close'
    })
  }

  const response = await grpc.updateConnectionLimits(request, {
    id: params.keyId as string,
    daily: {
      amount: BigInt(
        Math.floor(parseFloat(form.get('dailyLimit') as string) * 100)
      ),
      asset: 'USD',
      assetScale: 2
    },
    monthly: {
      amount: BigInt(
        Math.floor(parseFloat(form.get('monthlyLimit') as string) * 100)
      ),
      asset: 'USD',
      assetScale: 2
    },
    overall: {
      amount: BigInt(
        Math.floor(parseFloat(form.get('overallLimit') as string) * 100)
      ),
      asset: 'USD',
      assetScale: 2
    }
  })
  if (isConnectError(response)) {
    if (response.code == Code.InvalidArgument) {
      return response.error({ errors })
    } else return response.error({ errors }, {}, { action: 'Contact support' })
  }

  return json({ errors })
}
