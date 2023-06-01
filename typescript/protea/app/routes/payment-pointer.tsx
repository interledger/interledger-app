import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useFetcher, useLoaderData } from '@remix-run/react'
import type { ChangeEventHandler } from 'react'
import { useCallback, useState } from 'react'
import { route } from 'routes-gen'
import { Button, Card, Icon, Layouts, Snackbar, TextField } from '~/components'
import { getUserSession } from '~/lib/kratos.server'
import { PAYMENT_POINTER_BASE } from '~/lib/paymentPointer.server'
import type { GrpcError } from '~/lib/proto.server'
import {
  StatusError,
  httpMapping,
  isGrpcError,
  openPaymentsClient
} from '~/lib/proto.server'
import { flashSnackbar, getSnackbar } from '~/lib/snackbar.server'

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

  const snackbar = await getSnackbar(request)

  return json({
    paymentPointerBase: PAYMENT_POINTER_BASE,
    username: username.toLowerCase(),
    snackbar
  })
}

export const handle = {
  title: 'Payment pointer',
  layout: Layouts.Focus
}

export const meta: MetaFunction = () => {
  return {
    title: 'Payment pointer'
  }
}

export default function Page() {
  const fetcher = useFetcher()
  const { paymentPointerBase, username, snackbar } =
    useLoaderData<typeof loader>()
  const [showSnackbar, setSnackbar] = useState<boolean>(snackbar.show ?? false)

  const _onChangeInput = useCallback<ChangeEventHandler<HTMLInputElement>>(
    (event) => {
      const username = event.target.value
      if (username?.length >= 3) {
        fetcher.submit({ username }, { method: 'post' })
      }
    },
    [fetcher]
  )

  return (
    <>
      <fetcher.Form
        id='payment-pointer'
        action='/payment-pointer'
        method='post'
        className='hidden'
      />
      <Card>
        <p>Create your unique, memorable payment pointer.</p>

        <TextField
          id='username'
          form='payment-pointer'
          label='Payment pointer'
          name='username'
          prefix={`$${paymentPointerBase}/`}
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
          className='mt-6'
          aria-invalid={Boolean(fetcher.data?.errors.username) || undefined}
          aria-describedby={
            fetcher.data?.errors.username ? 'username-error' : undefined
          }
          errorMessage={fetcher.data?.errors.username || undefined}
          successMessage={
            fetcher.data?.errors.username || username == ''
              ? undefined
              : 'This payment pointer is available.'
          }
        />
        <input
          form='payment-pointer'
          value='true'
          name='canSubmit'
          type='hidden'
        />
      </Card>
      <Button form='payment-pointer' type='submit'>
        Save
      </Button>
      <Snackbar
        message={snackbar.message}
        icon={snackbar.icon}
        action={snackbar.action}
        show={showSnackbar}
        id='cookie-snackbar'
        onClose={() => setSnackbar(false)}
        dismissAfter={3000}
      />
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

  const fieldErrors = {
    form: '',
    username: ''
  }

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

      return json({ errors: { ...fieldErrors } }, { status: 400 })
    } else throw json({}, httpMapping(response.code))
  }

  if (response.response.exists) {
    return json(
      {
        errors: {
          username:
            'That payment pointer has been taken. Please choose another.'
        }
      },
      { status: 400 }
    )
  }

  if (canSubmit) {
    let response = await openPaymentsClient
      .createPaymentPointer(
        {
          url: `https://${PAYMENT_POINTER_BASE}/${username}`,
          asset: 'USD',
          assetScale: 2,
          alias: 'default'
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
        return json({ errors: { ...fieldErrors } }, { status: 400 })
      } else throw json({}, httpMapping(response.code))
    }

    return redirect(route('/'), {
      headers: {
        'Set-Cookie': await flashSnackbar(request, {
          message: 'Payment pointer reserved.',
          icon: 'close'
        })
      }
    })
  } else return json({ errors: { ...fieldErrors } })
}
