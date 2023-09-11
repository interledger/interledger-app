import { Code } from '@bufbuild/connect'
import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Button, Card, Layouts, TextField } from '~/components'
import { connectClient } from '~/lib/connect.server'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'
import { getPublicWalletDetails, getWalletInfo } from '~/lib/wallet.server'

export async function loader({ request }: LoaderArgs) {
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

export async function action({ request }: ActionArgs) {
  const form = await request.formData()
  const name = form.get('name') as string

  await validateCSRFToken(request, form)

  const errors = {
    form: '',
    name: ''
  }

  const response = await connectClient.setWalletName(request, {
    name
  })
  if (isConnectError(response)) {
    if (response.code == Code.InvalidArgument) {
      return response.error({ errors })
    } else return response.error({ errors }, {}, { action: 'Contact support' })
  }

  return redirectWithSnackbar(request, route('/settings/profile-public'), {
    message: 'Your public name was updated.',
    icon: 'close'
  })
}
