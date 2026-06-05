import { Code } from '@bufbuild/connect'
import type { UIMatch } from 'react-router'
import {
  Form,
  href,
  useActionData,
  useLoaderData,
  useParams
} from 'react-router'
import type { ApplicationProps } from '~/components'
import { Button, Card, Layouts, TextField } from '~/components'
import { getLinkedAccount } from '~/data/accounts.server'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import type { Route } from './+types/accounts_.$accountId.name'

export async function loader({ request, params }: Route.LoaderArgs) {
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
      back: href('/accounts'),
      title: (match: UIMatch<Route.ComponentProps['loaderData']>) =>
        match.loaderData?.type == 'bank'
          ? 'Bank account nickname'
          : 'Card nickname'
    }
  }
}

export const meta = mergeMeta(({ data }) => {
  const d = data as Route.ComponentProps['loaderData'] | undefined
  return [
    { title: d?.type == 'bank' ? 'Bank account nickname' : 'Card nickname' }
  ]
})

export default function Page() {
  const { name, csrfToken } = useLoaderData<typeof loader>()
  const actionData = useActionData()
  const params = useParams()

  return (
    <>
      <Form
        id='edit-linked-account-name'
        action={href('/accounts/:accountId/name', {
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

export async function action({ request, params }: Route.ActionArgs) {
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
    href('/accounts/:accountId', {
      accountId: params.accountId as string
    }),
    {
      message: 'Linked account nickname updated.',
      icon: 'close'
    }
  )
}
