import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useActionData, useLoaderData, useParams } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Button, Card, Layouts, TextField } from '~/components'
import { Code } from '~/generated/protobuf-ts/google/rpc/code'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import type { GrpcError } from '~/lib/proto.server'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { getLinkedAccount } from '~/lib/wallet.server'

export async function loader({ request, params }: LoaderArgs) {
  const account = await getLinkedAccount(request, params.accountId as string)
  return jsonWithCSRF(request, {
    name: account.nickname,
    type: account.type
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/accounts'),
      title: (match) =>
        match.data.type == 'bank' ? 'Bank account nickname' : 'Card nickname'
    }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Account name'
  }
}

export default function Page() {
  const { name, csrfToken } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()
  const params = useParams()

  return (
    <>
      <Form
        id='edit-linked-account-name'
        action={route('/accounts/:accountId/name', {
          accountId: params.accountId as string
        })}
        method='post'
        className='hidden'
      />
      <input
        form='edit-linked-account-name'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      <Card>
        <TextField
          id='name'
          label='Nickname'
          name='name'
          form='edit-linked-account-name'
          type='text'
          defaultValue={name}
          aria-invalid={Boolean(actionData?.errors.name) || undefined}
          aria-describedby={
            actionData?.errors.name ? 'password-error' : undefined
          }
          errorMessage={actionData?.errors.name}
        />
      </Card>
      <Button form='edit-linked-account-name' type='submit'>
        Save
      </Button>
    </>
  )
}

// The field names given by the backend for field violations
type fieldErrorsMap = 'Nickname'

function mapper(field: fieldErrorsMap): 'name' | null {
  switch (field) {
    case 'Nickname':
      return 'name'
    default:
      return null
  }
}

export async function action({ request, params }: ActionArgs) {
  const cookie = String(request.headers.get('cookie'))
  const form = await request.formData()
  const nickname = form.get('name') as string

  await validateCSRFToken(request, form)

  const fieldErrors = {
    form: '',
    name: ''
  }

  const response = await grpcClient
    .setNicknameLinkedAccount(
      {
        id: params.accountId as string,
        nickname
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

  return redirectWithSnackbar(
    request,
    route('/accounts/:accountId', {
      accountId: params.accountId as string
    }),
    {
      message: 'Linked account nickname updated.',
      icon: 'close'
    }
  )
}
