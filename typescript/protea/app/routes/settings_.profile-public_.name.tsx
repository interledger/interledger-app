import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Button, Card, Layouts, TextField } from '~/components'
import { Code } from '~/generated/protobuf-ts/google/rpc/code'
import type { GrpcError } from '~/lib/proto.server'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'
import { flashSnackbar } from '~/lib/snackbar.server'
import {
  getPublicWalletDetails,
  getWalletPaymentPointer
} from '~/lib/wallet.server'

export async function loader({ request }: LoaderArgs) {
  const paymentPointer = await getWalletPaymentPointer(request)
  const wallet = await getPublicWalletDetails(request, paymentPointer.walletID)

  return json({
    name: wallet.publicName
  })
}

export const handle = {
  title: 'Edit public name',
  layout: Layouts.Focus
}

export const meta: MetaFunction = () => {
  return {
    title: 'Edit public name'
  }
}

export default function Page() {
  const { name } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()

  return (
    <>
      <Form
        id='edit-public-name'
        action='/settings/profile-public/name'
        method='post'
        className='hidden'
      />
      <Card>
        <TextField
          id='name'
          label='Public name'
          name='name'
          form='edit-public-name'
          type='text'
          defaultValue={name}
          aria-invalid={Boolean(actionData?.errors.name) || undefined}
          aria-describedby={
            actionData?.errors.name ? 'password-error' : undefined
          }
          required
          errorMessage={actionData?.errors.name}
        />
      </Card>
      <Button form='edit-public-name' type='submit'>
        Save
      </Button>
    </>
  )
}

// The field names given by the backend for field violations
type fieldErrorsMap = 'Name'

function mapper(field: fieldErrorsMap): 'name' | null {
  switch (field) {
    case 'Name':
      return 'name'
    default:
      return null
  }
}

export async function action({ request }: ActionArgs) {
  const cookie = String(request.headers.get('cookie'))
  const form = await request.formData()
  const name = form.get('name') as string

  const fieldErrors = {
    form: '',
    name: ''
  }

  const response = await grpcClient
    .setWalletName(
      {
        name
      },
      {
        meta: {
          cookies: cookie || ''
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
    message: 'Your public name was updated.',
    icon: 'close'
  })

  return redirect(route('/settings/profile-public'))
}
