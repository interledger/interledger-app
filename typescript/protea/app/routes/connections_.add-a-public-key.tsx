import type { ActionArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Button, Card, Layouts, TextArea, TextField } from '~/components'
import { Code } from '~/generated/protobuf-ts/google/rpc/code'
import type { GrpcError } from '~/lib/proto.server'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'
import { flashSnackbar } from '~/lib/snackbar.server'

export const handle = {
  title: 'Add a public key',
  layout: Layouts.Focus
}

export const meta: MetaFunction = () => {
  return {
    title: 'Connections | Add a public key'
  }
}

export default function Page() {
  const actionData = useActionData<typeof action>()

  return (
    <>
      <Form
        id='add-public-key'
        action={'/connections/add-a-public-key'}
        method='post'
        className='hidden'
      />

      <Card>
        <p className='text-medium'>
          Add the public key of the external application that is connecting to
          your wallet.
        </p>

        <TextField
          id='applicationName'
          label='Application Name'
          name='applicationName'
          form='add-public-key'
          type='text'
          defaultValue=''
          className='mt-6'
          aria-invalid={
            Boolean(actionData?.errors.applicationName) || undefined
          }
          aria-describedby={
            actionData?.errors.applicationName
              ? 'applicationName-error'
              : undefined
          }
          required
          errorMessage={actionData?.errors.applicationName}
        />

        <TextArea
          id='publicKey'
          form='add-public-key'
          label='Public key'
          name='publicKey'
          placeholder='-----BEGIN PUBLIC KEY-----'
          className='mt-6'
          aria-invalid={Boolean(actionData?.errors.publicKey) || undefined}
          aria-describedby={
            actionData?.errors.publicKey ? 'publicKey-error' : undefined
          }
          required
          errorMessage={actionData?.errors.publicKey}
        />
      </Card>

      <Card>
        <h2 className='font-display text-lg font-medium'>Limits</h2>
        <p className='mt-6'>
          Providing access to your Fynbos wallet allows the external application
          to make payments. Set the limits below.
        </p>
        <TextField
          id='dailyLimit'
          label='Daily'
          name='dailyLimit'
          form='add-public-key'
          type='number'
          min='0'
          step='0.01'
          defaultValue={100}
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
          form='add-public-key'
          type='number'
          min='0'
          step='0.01'
          defaultValue={1000}
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
          form='add-public-key'
          type='number'
          min='0'
          step='0.01'
          defaultValue={10000}
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

      <Button form='add-public-key' type='submit'>
        Add key
      </Button>
    </>
  )
}

// The field names given by the backend for field violations
type fieldErrorsMap =
  | 'ApplicationName'
  | 'PublicKey'
  | 'DailyLimit'
  | 'MonthlyLimit'
  | 'OverallLimit'

function mapper(
  field: fieldErrorsMap
):
  | 'applicationName'
  | 'publicKey'
  | 'dailyLimit'
  | 'monthlyLimit'
  | 'overallLimit'
  | null {
  switch (field) {
    case 'ApplicationName':
      return 'applicationName'
    case 'PublicKey':
      return 'publicKey'
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

export async function action({ request }: ActionArgs) {
  const form = await request.formData()
  const fieldErrors = {
    applicationName: '',
    publicKey: '',
    dailyLimit: '',
    monthlyLimit: '',
    overallLimit: ''
  }

  const response = await grpcClient
    .createConnection(
      {
        applicationName: form.get('applicationName') as string,
        publicKey: form.get('publicKey') as string,
        dailyLimit: {
          amount: String(
            Math.floor(parseFloat(form.get('dailyLimit') as string) * 100)
          ),
          asset: 'USD',
          assetScale: 2
        },
        monthlyLimit: {
          amount: String(
            Math.floor(parseFloat(form.get('monthlyLimit') as string) * 100)
          ),
          asset: 'USD',
          assetScale: 2
        },
        overallLimit: {
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

  await flashSnackbar(request, {
    message: 'Public key was added.',
    icon: 'close'
  })

  return redirect(route('/connections'))
}
