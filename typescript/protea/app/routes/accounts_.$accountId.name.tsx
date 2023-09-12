import { Code } from '@bufbuild/connect'
import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { Form, useActionData, useLoaderData, useParams } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Button, Card, Layouts, TextField } from '~/components'
import { getLinkedAccount } from '~/data/wallet.server'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'

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

export async function action({ request, params }: ActionArgs) {
  const form = await request.formData()
  const nickname = form.get('name') as string

  await validateCSRFToken(request, form)

  const errors = {
    form: '',
    name: ''
  }
  const mapping = {
    name: 'Nickname'
  }

  const response = await grpc.setNicknameLinkedAccount(request, {
    id: params.accountId as string,
    nickname
  })
  if (isConnectError(response)) {
    if (response.code == Code.InvalidArgument) {
      return response.error({ errors }, mapping)
    } else
      return response.error({ errors }, mapping, {
        action: 'Contact support'
      })
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
