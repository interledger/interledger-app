import { useCallback, useState } from 'react'
import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useFetcher, useLoaderData } from '@remix-run/react'
import { Button, Icon, Layouts, Shape, Snackbar, TextField } from '~/components'
import type { GrpcError } from '~/lib/proto.server'
import {
  httpMapping,
  isGrpcError,
  openPaymentsClient,
  StatusError
} from '~/lib/proto.server'
import { route } from 'routes-gen'
import { getUserSession } from '~/lib/kratos.server'
import { getSnackbar, flashSnackbar } from '~/lib/snackbar.server'

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
        url: 'https://fynbos.me/' + username
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

  return json({ username: username.toLowerCase(), snackbar })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const fetcher = useFetcher()
  const { username, snackbar } = useLoaderData<typeof loader>()
  const [showSnackbar, setSnackbar] = useState<boolean>(snackbar.show ?? false)

  const _onChangeInput = useCallback(
    (event) => {
      const username = event.target.value
      if (username?.length >= 3) {
        fetcher.submit({ username }, { method: 'post' })
      }
    },
    [fetcher]
  )

  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <div className='flex flex-col space-y-6'>
        <div className='flex justify-between'>
          <span className='font-display text-2xl font-medium'>
            Payment pointer
          </span>
          <div className='hidden sm:flex'>
            <Shape
              width={'w-8'}
              radius={'rounded-full'}
              color={'bg-rose-500'}
            />
            <Shape
              width={'w-8'}
              radius={'rounded-tr-full'}
              color={'bg-lime-300'}
            />
          </div>
        </div>
        <p>Create your unique, memorable payment pointer.</p>
      </div>
      <fetcher.Form
        id='payment-pointer'
        action='/payment-pointer'
        method='post'
        className='hidden'
      />

      <TextField
        id='username'
        form='payment-pointer'
        label='Payment pointer'
        name='username'
        prefix='$fynbos.me/'
        appendIcon={
          username == '' &&
          typeof fetcher.data == 'undefined' ? undefined : fetcher.data?.errors
              .username ? (
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

      <Button className='mt-12' form='payment-pointer' type='submit'>
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
    </div>
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
      url: 'https://fynbos.me/' + username
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
          url: 'https://fynbos.me/' + username,
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
