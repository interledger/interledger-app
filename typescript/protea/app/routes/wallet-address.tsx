import { Code } from '@bufbuild/connect'
import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useFetcher, useLoaderData } from '@remix-run/react'
import type { ChangeEventHandler } from 'react'
import { useCallback, useState } from 'react'
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

export async function loader({ request }: LoaderFunctionArgs) {
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

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Wallet'
  }
])

// we allow only alpha numeric characters, _, and a length between 3 & 17 characters
const validLength: { min: number, max: number } = { min: 3, max: 17 }
const regex = new RegExp(`^[a-zA-Z0-9_]{0,${validLength.max}}$`)
const isAllowed = (str: string): boolean => regex.test(str)

export default function Page() {
  const fetcher = useFetcher<typeof action>()
  const { paymentPointerBase, username, csrfToken } =
    useLoaderData<typeof loader>()
  const [userInput, setUserInput] = useState(username)

  const _onChangeInput = useCallback<ChangeEventHandler<HTMLInputElement>>(
    (event) => {
      const username = event.target.value
      if (isAllowed(username)) {
        setUserInput(username)
      }
      if (isAllowed(username) && username?.length >= validLength.min) {
        fetcher.submit({ username, csrfToken }, { method: 'post' })
      }
    },
    [csrfToken, fetcher]
  )

  const _validateInput = (username?: string) => {
    const fetcherUserNameError = fetcher.data?.errors.username
    const userInputError = (username || "").length < validLength.min
    const hasError = userInputError || fetcherUserNameError
    const displayMaxLength = validLength.max + paymentPointerBase.length + 1

    const appendIcon =
      hasError ? (
        <Icon className='text-error'>error</Icon>
      ) : (
        <Icon className='text-success'>check</Icon>
      )

    const ariaInvalid = Boolean(hasError) || undefined
    const ariaDescribedby =
      hasError ? 'username-error' : undefined

    const errorMessage = userInputError ? `Wallet address has to be between ${validLength.min} & ${displayMaxLength} characters long.` : (fetcherUserNameError || undefined)
    const successMessage =
      hasError
        ? undefined
        : 'This wallet address is available.'


    return { appendIcon, "aria-invalid": ariaInvalid, "aria-describedby": ariaDescribedby, errorMessage, successMessage }
  }

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
          value={userInput}
          onChange={_onChangeInput}
          type='text'
          className='mt-2'
          {..._validateInput(userInput)}
        />
      </Card>
      <Button form='wallet-address' type='submit'>
        Save
      </Button>
    </>
  )
}

export async function action({ request }: ActionFunctionArgs) {
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

  if (!isAllowed(username) || username.length < validLength.min) {
    const displayMaxLength = validLength.max + PAYMENT_POINTER_BASE.length + 1
    errors.username =
      `Wallet address has to be between ${validLength.min} & ${displayMaxLength} characters long.`
    return error(request, { errors })
  }

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
