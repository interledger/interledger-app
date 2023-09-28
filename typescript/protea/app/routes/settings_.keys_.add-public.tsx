import { Code } from '@bufbuild/connect'
import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
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
  TextArea,
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
      back: route('/settings/keys'),
      title: 'Add a public key'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Add a public key'
  }
])

export async function loader({ request }: LoaderFunctionArgs) {
  return jsonWithCSRF(request, {})
}

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { csrfToken } = useLoaderData<typeof loader>()

  return (
    <>
      <Form
        id='add-public-key'
        action='/settings/keys/add-public'
        method='post'
        className='hidden'
      />
      <input
        form='add-public-key'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      <Card>
        <CardContent>
          <p>
            Add the public key of the external application that is connecting to
            your wallet.
          </p>
        </CardContent>
        <TextField
          id='applicationName'
          label='Application Name'
          name='applicationName'
          form='add-public-key'
          type='text'
          defaultValue=''
          className='mt-2'
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
          className='mt-4'
          placeholder='-----BEGIN PUBLIC KEY-----'
          aria-invalid={Boolean(actionData?.errors.publicKey) || undefined}
          aria-describedby={
            actionData?.errors.publicKey ? 'publicKey-error' : undefined
          }
          required
          errorMessage={actionData?.errors.publicKey}
        />
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
          form='add-public-key'
          type='number'
          min='0'
          step='0.01'
          className='mt-2'
          defaultValue={100}
          prefix='$'
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
          form='add-public-key'
          type='number'
          min='0'
          step='0.01'
          defaultValue={10000}
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

      <Button form='add-public-key' type='submit'>
        Add key
      </Button>
    </>
  )
}

export async function action({ request }: ActionFunctionArgs) {
  const form = await request.formData()

  await validateCSRFToken(request, form)

  const errors = {
    form: '',
    applicationName: '',
    publicKey: '',
    dailyLimit: '',
    monthlyLimit: '',
    overallLimit: ''
  }

  const response = await grpc.createConnection(request, {
    applicationName: form.get('applicationName') as string,
    publicKey: form.get('publicKey') as string,
    dailyLimit: {
      amount: BigInt(
        Math.floor(parseFloat(form.get('dailyLimit') as string) * 100)
      ),
      asset: 'USD',
      assetScale: 2
    },
    monthlyLimit: {
      amount: BigInt(
        Math.floor(parseFloat(form.get('monthlyLimit') as string) * 100)
      ),
      asset: 'USD',
      assetScale: 2
    },
    overallLimit: {
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

  return redirectWithSnackbar(request, route('/settings/keys'), {
    message: 'Public key was added.',
    icon: 'close'
  })
}
