import { useCallback } from 'react'
import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import {
  Form,
  useActionData,
  useFetcher,
  useLoaderData
} from '@remix-run/react'
import { Button, Icon, Shape, TextField } from '~/components'
import type { GrpcError } from '~/lib/proto.server'
import {
  httpMapping,
  isGrpcError,
  openPaymentsClient,
  StatusError
} from '~/lib/proto.server'
import { route } from 'routes-gen'
import { requireUserSession } from '~/lib/kratos.server'

export async function loader({ request }: LoaderArgs) {
  const session = await requireUserSession(request)
  console.log(session.identity.traits.firstName)
  return json({ username: session.identity.traits.firstName })
}

export default function Page() {
  /**
   * When a form needs to open a dialog to complete the process, the initial form can use a fetcher.Form - because the submission of this form doesn't cause navigation.
   */
  const fetcher = useFetcher()
  const actionData = useActionData<typeof action>()
  const { username } = useLoaderData<typeof loader>()
  console.log('actionData', actionData)

  const _onChangeInput = useCallback(
    (event) => {
      const username = event.target.value
      console.log('EVENT', username)
      if (username?.length >= 3) {
        fetcher.submit({ username }, { method: 'post' })
      }
      // actionData = {}
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
              radius={'rounded-tr-full'}
              color={'bg-lime-400'}
            />
            <Shape
              width={'w-8'}
              radius={'rounded-tl-full'}
              color={'bg-slate-300'}
            />
          </div>
        </div>
        <p>Create your unique, memorable payment pointer.</p>
      </div>
      <Form
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
          fetcher.data?.errors.username || actionData?.errors.username ? (
            <Icon className='text-error'>error</Icon>
          ) : (
            <Icon className='text-success'>check</Icon>
          )
        }
        defaultValue={username}
        onChange={_onChangeInput}
        type='text'
        className='mt-6'
        aria-invalid={
          Boolean(fetcher.data?.errors.username) ||
          Boolean(actionData?.errors.username) ||
          undefined
        }
        aria-describedby={
          fetcher.data?.errors.username || actionData?.errors.username
            ? 'username-error'
            : undefined
        }
        required
        errorMessage={
          fetcher.data?.errors.username ||
          actionData?.errors.username ||
          undefined
        }
        successMessage={
          fetcher.data?.errors.username || actionData?.errors.username
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
        Continue
      </Button>
    </div>
  )
}

// The field names given by the backend for field violations
type fieldErrorsMap = 'Url'

function mapper(field: fieldErrorsMap): 'username' | null {
  switch (field) {
    case 'Url':
      return 'username'
    default:
      return null
  }
}

export async function action({ request }: ActionArgs) {
  const cookie = String(request.headers.get('cookie'))
  const form = await request.formData()
  const prefix = 'https://fynbos.me/'
  const username = form.get('username') as string
  const canSubmit = Boolean(form.get('canSubmit') as string)

  console.log('PP: ', prefix + username)
  console.log('canSubmit: ', canSubmit)

  // TODO: check username is valid
  // Should just return whether it's valid or not, unless the form is submitted with a redirect flag, in which case it can create payment pointer and continue.

  const fieldErrors = {
    username: ''
  }
  // if (username.length < 4) {
  //   fieldErrors.username = 'Too short.'
  // }
  console.log('username', username)

  let response = await openPaymentsClient
    .paymentPointerExists({
      url: prefix + username
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
  // TODO: Ensure that the validation checks if it exists
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
    console.log('here')
    let response = await openPaymentsClient
      .createPaymentPointer(
        {
          url: prefix + username,
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
    return redirect(route('/'))
  } else return json({ errors: { ...fieldErrors } })
}
