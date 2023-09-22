import { Code } from '@bufbuild/connect'
import type { ActionArgs, LoaderArgs, V2_MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useFetcher, useLoaderData } from '@remix-run/react'
import type { ChangeEventHandler } from 'react'
import { useCallback } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Button,
  Card,
  CardContent,
  Icon,
  Layouts,
  TextField
} from '~/components'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { error, isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { getUserSession } from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'
import { PAYMENT_POINTER_BASE } from '~/lib/paymentPointer.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'

export async function loader({ request }: LoaderArgs) {
  let response = await grpc.getWalletInfo(request, {})
  if (isConnectError(response)) {
    throw response.errorResponse
  } else if (response.hasWalletAddress) {
    throw redirect(route('/'))
  }

  const session = await getUserSession(request)
  let usernameIsValid = false
  let attempts = 0
  let username = session.identity.traits.firstName
  let publicName =
    session.identity.traits.firstName + ' ' + session.identity.traits.lastName

  while (!usernameIsValid && attempts < 5) {
    let response = await grpc.walletAddressExists(request, {
      url: `https://${PAYMENT_POINTER_BASE}/${username}`
    })

    if (isConnectError(response) || response.exists) {
      attempts++
      username = session.identity.traits.firstName
      if (username.length < 4) username += session.identity.traits.lastName

      if (attempts > 1)
        username += String(Math.floor(Math.random() * 10000)).padStart(4, '0')

      if (attempts == 5) username = ''
    } else {
      usernameIsValid = true
    }
  }

  return jsonWithCSRF(request, {
    paymentPointerBase: PAYMENT_POINTER_BASE,
    username: username.toLowerCase(),
    publicName: publicName
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      title: 'Wallet'
    }
  }
}

export const meta: V2_MetaFunction = mergeMeta(() => [
  {
    title: 'Wallet'
  }
])

export default function Page() {
  const fetcher = useFetcher()
  const { paymentPointerBase, username, csrfToken } =
    useLoaderData<typeof loader>()

  const _onChangeInput = useCallback<ChangeEventHandler<HTMLInputElement>>(
    (event) => {
      const username = event.target.value
      if (username?.length >= 3) {
        fetcher.submit({ username, csrfToken }, { method: 'post' })
      }
    },
    [csrfToken, fetcher]
  )

  return (
    <>
      <fetcher.Form
        id='wallet-address'
        action='/wallet-address'
        method='post'
        className='hidden'
      />
      <input
        form='wallet-address'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      <input
        form='wallet-address'
        value='true'
        name='canSubmit'
        type='hidden'
      />
      <Card>
        <CardContent>
          <p className='text-medium'>
            Create your unique wallet address below.
          </p>
        </CardContent>
        <TextField
          id='username'
          form='wallet-address'
          label='Wallet address'
          name='username'
          prefix={`${paymentPointerBase}/`}
          appendIcon={
            username == '' &&
            typeof fetcher.data == 'undefined' ? undefined : fetcher.data
                ?.errors.username ? (
              <Icon className='text-error'>error</Icon>
            ) : (
              <Icon className='text-success'>check</Icon>
            )
          }
          defaultValue={username}
          onChange={_onChangeInput}
          type='text'
          className='mt-2'
          aria-invalid={Boolean(fetcher.data?.errors.username) || undefined}
          aria-describedby={
            fetcher.data?.errors.username ? 'username-error' : undefined
          }
          errorMessage={fetcher.data?.errors.username || undefined}
          successMessage={
            fetcher.data?.errors.username || username == ''
              ? undefined
              : 'This wallet address is available.'
          }
        />
      </Card>
      <Button form='wallet-address' type='submit'>
        Save
      </Button>
    </>
  )
}

export async function action({ request }: ActionArgs) {
  const form = await request.formData()
  const username = form.get('username') as string
  const canSubmit = Boolean(form.get('canSubmit') as string)

  await validateCSRFToken(request, form)

  const errors = {
    form: '',
    username: ''
  }
  const mapping = {
    username: 'url'
  }

  const publicName = username

  let response = await grpc.walletAddressExists(request, {
    url: `https://${PAYMENT_POINTER_BASE}/${username}`
  })
  if (isConnectError(response)) {
    if (response.code == Code.InvalidArgument) {
      return response.error({ errors }, mapping)
    } else
      return response.error({ errors }, mapping, { action: 'Contact support' })
  }

  if (response.exists) {
    errors.username =
      'That wallet address has been taken. Please choose another.'
    return error(request, { errors })
  }

  if (canSubmit) {
    let response = await grpc.createWalletAddress(request, {
      url: `https://${PAYMENT_POINTER_BASE}/${username}`,
      asset: 'USD',
      assetScale: 2,
      alias: publicName
    })
    if (isConnectError(response)) {
      if (response.code == Code.InvalidArgument) {
        return response.error({ errors }, mapping)
      } else
        return response.error({ errors }, mapping, {
          action: 'Contact support'
        })
    }

    return redirectWithSnackbar(request, route('/'), {
      message: 'Your wallet address is reserved.',
      icon: 'close'
    })
  } else return json({ errors })
}
