import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
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
import { error } from '~/lib/error.server'
import { getUserSession } from '~/lib/kratos.server'
import { PAYMENT_POINTER_BASE } from '~/lib/paymentPointer.server'
import type { GrpcError } from '~/lib/proto.server'
import {
  StatusError,
  httpMapping,
  isGrpcError,
  openPaymentsClient
} from '~/lib/proto.server'
import { redirectWithSnackbar } from '~/lib/snackbar.server'

export async function loader({ request }: LoaderArgs) {
  let response = await openPaymentsClient
    .listWalletPaymentPointers(
      {},
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  } else if (response.response.pointers.length > 0) {
    throw redirect(route('/'))
  }

  const session = await getUserSession(request)
  let usernameIsValid = false
  let attempts = 0
  let username = session.identity.traits.firstName
  let publicName =
    session.identity.traits.firstName + ' ' + session.identity.traits.lastName

  while (!usernameIsValid && attempts < 5) {
    let response = await openPaymentsClient
      .paymentPointerExists({
        url: `https://${PAYMENT_POINTER_BASE}/${username}`
      })
      .then((v) => v)
      .catch(StatusError)
    if (isGrpcError(response) || response.response.exists) {
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

export const meta: MetaFunction = () => {
  return {
    title: 'Wallet'
  }
}

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

// The field names given by the backend for field violations
type fieldErrorsMap = 'url'

function mapper(field: fieldErrorsMap): 'username' | null {
  switch (field) {
    case 'url':
      return 'username'
    default:
      return null
  }
}

export async function action({ request }: ActionArgs) {
  const cookie = String(request.headers.get('cookie'))
  const form = await request.formData()
  const username = form.get('username') as string
  const canSubmit = Boolean(form.get('canSubmit') as string)

  await validateCSRFToken(request, form)

  const fieldErrors = {
    form: '',
    username: ''
  }

  const publicName = username

  let response = await openPaymentsClient
    .paymentPointerExists({
      url: `https://${PAYMENT_POINTER_BASE}/${username}`
    })
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    if (response.code == 3) {
      for (let violation of (response as GrpcError).details[0]
        .fieldViolations) {
        const field = mapper(violation.field as fieldErrorsMap)
        if (field != null) fieldErrors[field] = violation.description
      }

      return error(request, fieldErrors)
    } else return error(request, fieldErrors, true, 'Contact support')
  }

  if (response.response.exists) {
    fieldErrors.username =
      'That wallet address has been taken. Please choose another.'
    return error(request, fieldErrors)
  }

  if (canSubmit) {
    let response = await openPaymentsClient
      .createPaymentPointer(
        {
          url: `https://${PAYMENT_POINTER_BASE}/${username}`,
          asset: 'USD',
          assetScale: 2,
          alias: publicName
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
      if (response.code == 3) {
        for (let violation of (response as GrpcError).details[0]
          .fieldViolations) {
          const field = mapper(violation.field as fieldErrorsMap)
          if (field != null) fieldErrors[field] = violation.description
        }
        return error(request, fieldErrors)
      } else return error(request, fieldErrors, true, 'Contact support')
    }

    return redirectWithSnackbar(request, route('/'), {
      message: 'Your wallet address is reserved.',
      icon: 'close'
    })
  } else return json({ errors: { ...fieldErrors } })
}
