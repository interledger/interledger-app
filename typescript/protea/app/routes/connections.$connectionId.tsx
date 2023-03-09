import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Button, Card, Layouts, OutlineButton, TextField } from '~/components'
import { Code } from '~/generated/protobuf-ts/google/rpc/code'
import type { GrpcError } from '~/lib/proto.server'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'
import { flashSnackbar } from '~/lib/snackbar.server'

export const handle = {
  layout: Layouts.FocusLayout
}

export const meta: MetaFunction = () => {
  return {
    title: 'Connections'
  }
}

async function getKey(request: Request, id: string) {
  let rpc = await grpcClient
    .getPublicKey(
      { id },
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(rpc)) {
    throw json({}, httpMapping(rpc.code))
  }

  return rpc.response
}

async function getKeyLimits(request: Request, id: string) {
  let rpc = await grpcClient
    .getPublicKeyLimits(
      { id },
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(rpc)) {
    throw json({}, httpMapping(rpc.code))
  }

  return rpc.response
}

export async function loader({ request, params }: LoaderArgs) {
  let data = await Promise.all([
    getKey(request, params.connectionId as string),
    getKeyLimits(request, params.connectionId as string)
  ])

  return json({ key: data[0], limits: data[1] })
}

export default function Page() {
  const { key, limits } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()

  return (
    <>
      <Form
        id='update-key-limit'
        action={`/connections/${key.id}`}
        method='post'
        className='hidden'
      />
      <Form
        id='delete-key'
        action={`/connections/${key.id}`}
        method='post'
        className='hidden'
      />

      <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <h1 className='font-display text-2xl font-medium'>
          {key.applicationName}
        </h1>
        <p className='mt-6 text-medium'>
          Add the public key of the external application that is connecting to
          your wallet.
        </p>

        <div className='my-4 flex flex items-center justify-between rounded-xl bg-container p-4'>
          {key.publicKey}
        </div>

        <p className='text-sm'>Added {key.createdAt}</p>
        <p className='text-sm text-purple-500'>Last used {key.lastUsedAt}</p>
      </Card>

      <br />

      <Card className='col-span-full space-y-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <h1 className='font-display text-2xl font-medium'>Limits</h1>
        <p>
          Providing access to your Fynbos wallet allows the external application
          to make payments. Set the limits below.
        </p>
        <TextField
          id='dailyLimit'
          label='Daily'
          name='dailyLimit'
          form='update-key-limit'
          type='text'
          defaultValue={limits.daily?.amount}
          prefix='$'
          className='mt-6'
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
          form='update-key-limit'
          type='text'
          defaultValue={limits.monthly?.amount}
          prefix='$'
          className='mt-6'
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
          form='update-key-limit'
          type='text'
          defaultValue={limits.overall?.amount}
          prefix='$'
          className='mt-6'
          aria-invalid={Boolean(actionData?.errors.overallLimit) || undefined}
          aria-describedby={
            actionData?.errors.overallLimit ? 'overallLimit-error' : undefined
          }
          required
          errorMessage={actionData?.errors.overallLimit}
        />
      </Card>

      <br />

      <div className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <div className='grid grid-cols-3 gap-2'>
          <OutlineButton
            className='outline-red-700 focus-visible:outline-red-800 text-red-700'
            form='delete-key'
            type='submit'
          >
            Delete
          </OutlineButton>
          <Button className='col-span-2' form='add-public-key' type='submit'>
            Save
          </Button>
        </div>
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

  if (formName === 'delete-key') {
    const response = await grpcClient.deletePublicKey(
      { id: params.connectionId as string },
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
    if (isGrpcError(response)) {
      throw json({}, httpMapping(response.code))
    }

    await flashSnackbar(request, {
      message: 'Your public key was deleted.',
      icon: 'close'
    })

    return redirect(route('/connections'))
  }

  const fieldErrors = {
    dailyLimit: '',
    monthlyLimit: '',
    overallLimit: ''
  }

  const response = await grpcClient
    .updatePublicKeyLimit(
      {
        id: params.connectionId as string,
        daily: {
          amount: form.get('dailyLimit') as string,
          currency: 'USD'
        },
        monthly: {
          amount: form.get('monthlyLimit') as string,
          currency: 'USD'
        },
        overall: {
          amount: form.get('overallLimit') as string,
          currency: 'USD'
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

  await flashSnackbar(request, {
    message: 'Your public key was updated.',
    icon: 'close'
  })

  return redirect(route('/connections'))
}
