import { Code } from '@bufbuild/connect'
import type { ActionFunctionArgs, LoaderFunctionArgs, MetaFunction } from 'react-router';
import { Form, useActionData, useLoaderData } from 'react-router';
import { href } from 'react-router'
import type { ApplicationProps } from '~/components'
import { Button, Card, Layouts, TextField } from '~/components'
import { getPublicWalletDetails, getWalletInfo } from '~/data/wallet.server'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { mergeMeta } from '~/lib/meta'
import { redirectWithSnackbar } from '~/lib/snackbar.server'

export async function loader({ request }: LoaderFunctionArgs) {
  const walletInfo = await getWalletInfo(request)
  const wallet = await getPublicWalletDetails(request, walletInfo.walletID)

  return jsonWithCSRF(request, {
    name: wallet.publicName
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: href('/settings/profile-public'),
      title: 'Edit public name'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Edit public name'
  }
])

export default function Page() {
  const { name, csrfToken } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()
  return (
    <>
      <Form
        id='edit-public-name'
        action={href('/settings/profile-public/name')}
        method='post'
        className='hidden'
      />
      <input
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

export async function action({ request }: ActionFunctionArgs) {
  const form = await request.formData()
  const name = form.get('name') as string

  await validateCSRFToken(request, form)

  const errors = {
    form: '',
    name: ''
  }

  const response = await grpc.setWalletName(request, {
    name
  })
  if (isConnectError(response)) {
    if (response.code == Code.InvalidArgument) {
      return response.error({ errors })
    } else return response.error({ errors }, {}, { action: 'Contact support' })
  }

  return redirectWithSnackbar(request, href('/settings/profile-public'), {
    message: 'Your public name was updated.',
    icon: 'close'
  })
}
