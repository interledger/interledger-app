import { Button, Icon, Logo, TextField } from '~/components'
import { ActionArgs, json, LoaderArgs, redirect } from '@remix-run/node'
import { apolloClient } from '~/lib/apollo.server'
import {
  SignupDocument,
  SignupQuery,
  SignupQueryVariables
} from '~/generated/types'
import {
  grpcClient,
  GrpcError,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { useState } from 'react'

type Country = {
  id: string
  name: string
}

export async function loader({ request, params }: LoaderArgs) {
  const countries = await apolloClient
    .query<SignupQuery, SignupQueryVariables>({
      query: SignupDocument,
      context: {
        headers: request.headers
      }
    })
    .then((val) => val.data.countries as Country[])

  const url = new URL(request.url)
  const countryCode = url.searchParams.get('country')
  const email = url.searchParams.get('email')

  return json({
    countryCode,
    countries,
    email
  })
}

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { countryCode, countries, email } = useLoaderData<typeof loader>()
  const [country, setCountry] = useState<Country>(
    countries.find((country: Country) => country.id == countryCode) as Country
  )

  return (
    <div className='w-full'>
      <header className='sticky top-0 mx-auto flex h-16 w-full select-none items-center justify-between bg-app p-4 text-medium sm:max-w-lg lg:max-w-3xl xl:max-w-4xl'>
        <div className='flex items-center'>
          <div className='flex items-center justify-start font-display text-2xl font-medium'>
            <Logo className='h-8' />
          </div>
        </div>
      </header>
      <div className='mx-auto grid min-h-[calc(100vh-9rem)] w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
        <div className='col-span-full flex flex-col space-y-2 pt-4 pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <span className='font-display text-4xl font-medium'>
            We aren't in your country yet
          </span>
          <span className='text-medium'>
            We unfortunately don't support the country you reside in.
          </span>
          <span className='text-medium'>
            You can sign up to our waitlist, and we will let you know when
            Fynbos becomes available to you.
          </span>
        </div>

        <Form
          id='join_waitlist'
          action={`/waitlist`}
          method='post'
          className='hidden'
        />
        <TextField
          id='email'
          form='join_waitlist'
          label='Email'
          name='email'
          defaultValue={email as string}
          type='text'
          className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
          aria-invalid={Boolean(actionData?.errors.email) || undefined}
          aria-describedby={
            actionData?.errors.email ? 'email-error' : undefined
          }
          required
          errorMessage={actionData?.errors.email}
        />
        <input
          id='country'
          form='join_waitlist'
          name='country'
          value={countryCode as string}
          type='hidden'
        />
        <div className='col-span-full mb-4 flex items-center justify-between rounded-xl bg-container p-3 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <div className='flex items-center space-x-3 text-medium'>
            <Icon>flag</Icon>
            <span className='font-sans text-base font-normal'>
              {country.name}
            </span>
          </div>
        </div>
        <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <Button form='join_waitlist' type='submit'>
            Join Waitlist
          </Button>
        </div>
      </div>
    </div>
  )
}

// The field names given by the backend for field violations
type fieldErrorsMap = 'CountryCode' | 'Email'

function mapper(field: fieldErrorsMap): 'countryCode' | 'email' | null {
  switch (field) {
    case 'CountryCode':
      return 'countryCode'
    case 'Email':
      return 'email'
    default:
      return null
  }
}

export async function action({ request, params }: ActionArgs) {
  const form = await request.formData()
  const email = form.get('email') as string
  const country = form.get('country') as string

  const fieldErrors = {
    countryCode: '',
    email: ''
  }

  let response = await grpcClient
    .joinWaitlist({
      email,
      countryCode: country
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
    } else throw response
  }

  // Redirect to joined waitlist
  return redirect('/waitlist/success')
}
