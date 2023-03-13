import type { ActionArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Button, Card, Layouts, TextArea, TextField } from '~/components'
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
    title: 'Connections | Add a public key'
  }
}

export default function Page() {
  const actionData = useActionData<typeof action>()

  return (
    <>
      <Form
        id='add-public-key'
        action={'/connections/add-public-key'}
        method='post'
        className='hidden'
      />

      <Card className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <h1 className='font-display text-2xl font-medium'>Add a public key</h1>
        <p className='mt-6 text-medium'>
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
          id='key'
          form='add-public-key'
          label='Public key'
          name='key'
          placeholder='base64 encoded public key'
          className='mt-6'
          aria-invalid={Boolean(actionData?.errors.publicKey) || undefined}
          aria-describedby={
            actionData?.errors.publicKey ? 'publicKey-error' : undefined
          }
          required
          errorMessage={actionData?.errors.publicKey}
        />
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
          form='add-public-key'
          type='text'
          defaultValue='100.00'
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
          type='text'
          defaultValue='1000.00'
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
          type='text'
          defaultValue='10000.00'
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

      <Button
        className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'
        form='add-public-key'
        type='submit'
      >
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
    .createPublicKey(
      {
        applicationName: form.get('applicationName') as string,
        publicKey: form.get('publicKey') as string,
        dailyLimit: {
          amount: form.get('dailyLimit') as string,
          currency: 'USD'
        },
        monthlyLimit: {
          amount: form.get('monthlyLimit') as string,
          currency: 'USD'
        },
        overallLimit: {
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
    message: 'Your public key was added.',
    icon: 'close'
  })

  return redirect(route('/connections'))
}
