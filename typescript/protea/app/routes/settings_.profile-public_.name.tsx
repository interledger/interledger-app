import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Button, Card, Layouts, TextField } from '~/components'
import { Code } from '~/generated/protobuf-ts/google/rpc/code'
import { getSessionWithCSRFToken, validateCSRFToken } from '~/lib/csrf.server'
import type { GrpcError } from '~/lib/proto.server'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'
import { flashSnackbar } from '~/lib/snackbar.server'
import { getPublicWalletDetails, getWalletInfo } from '~/lib/wallet.server'
import { commitSession } from '~/session.server'

export async function loader({ request }: LoaderArgs) {
  const walletInfo = await getWalletInfo(request)
  const wallet = await getPublicWalletDetails(request, walletInfo.walletID)
  const session = await getSessionWithCSRFToken(request)
  return json(
    {
      name: wallet.publicName,
      csrfToken: session.get('csrf-token')
    },
    { headers: { 'Set-Cookie': await commitSession(session) } }
  )
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/settings/profile-public'),
      title: 'Edit public name'
    }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Edit public name'
  }
}

export default function Page() {
  const { name, csrfToken } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()
  return (
    <>
      <Form
        id='edit-public-name'
        action={route('/settings/profile-public/name')}
        method='post'
        className='hidden'
      />
      <input
        id='csrfToken'
        form='edit-public-name'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
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
  const csrfToken = form.get('csrfToken') as string
  const err = await validateCSRFToken(request, csrfToken).catch(
    (err: Error) => err
  )
  if (err) {
    return json(
      {
        errors: {
          name: 'Please try again.'
        }
      },
      { status: 422, statusText: 'Invalid CSRF token.' }
    )
  }

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
