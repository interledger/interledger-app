import {
  Autocomplete,
  Button,
  Checkbox,
  Layouts,
  TextField
} from '~/components'
import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import type { GrpcError } from '~/lib/proto.server'
import { httpMapping } from '~/lib/proto.server'
import { grpcClient, isGrpcError, StatusError } from '~/lib/proto.server'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { requireNoUserSession } from '~/lib/kratos.server'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'

type Country = {
  id: string
  name: string
}

export async function loader({ request }: LoaderArgs) {
  await requireNoUserSession(request)
  let response = await grpcClient
    .getCountries({})
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  const url = new URL(request.url)
  const countryCode = url.searchParams.get('country')
  const email = url.searchParams.get('email')
  const fullName = url.searchParams.get('fullName')

  return json({
    countryCode,
    countries: response.response.countries,
    email,
    fullName
  })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { countryCode, countries, email, fullName } =
    useLoaderData<typeof loader>()

  const [country, setCountry] = useState<Country>(
    countries.find((country: Country) => country.id == countryCode) as Country
  )

  const [query, setQuery] = useState<string>('')
  const [filteredCountries, setFilteredCountries] = useState(countries)

  useEffect(() => {
    if (query === '') setFilteredCountries(countries)
    else {
      setFilteredCountries(
        countries.filter((country: Country) => {
          return (
            country.name
              .toLowerCase()
              .replace(/\s+/g, '')
              .includes(query.toLowerCase().replace(/\s+/g, '')) ||
            country.id
              .toLowerCase()
              .replace(/\s+/g, '')
              .includes(query.toLowerCase().replace(/\s+/g, ''))
          )
        })
      )
    }
  }, [query, countries])

  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <span className='font-display text-2xl font-medium'>
        Join the waitlist
      </span>
      <span className='mt-6 text-medium'>
        Leave your details below and we will notify you as soon as enrollment
        opens.
      </span>

      <Form
        id='join-waitlist'
        action='/waitlist'
        method='post'
        className='hidden'
      />

      <TextField
        id='fullName'
        form='join-waitlist'
        label='Full name'
        name='fullName'
        type='text'
        defaultValue={fullName as string}
        className='mt-6'
        aria-invalid={Boolean(actionData?.errors.fullName) || undefined}
        aria-describedby={
          actionData?.errors.fullName ? 'lastName-error' : undefined
        }
        required
        errorMessage={actionData?.errors.fullName}
      />
      <TextField
        id='email'
        form='join-waitlist'
        label='Email address'
        name='email'
        defaultValue={email as string}
        type='text'
        className='mt-1'
        aria-invalid={Boolean(actionData?.errors.email) || undefined}
        aria-describedby={actionData?.errors.email ? 'email-error' : undefined}
        required
        errorMessage={actionData?.errors.email}
      />
      <Autocomplete
        id='country'
        value={country}
        onChange={setCountry}
        onQuery={setQuery}
        options={filteredCountries}
        label='Country of residence'
        className='mt-1'
        aria-invalid={Boolean(actionData?.errors.countryCode) || undefined}
        aria-describedby={
          actionData?.errors.countryCode ? 'country-error' : undefined
        }
        errorMessage={actionData?.errors.countryCode}
      />
      <input
        id='country'
        form='join-waitlist'
        value={String(country?.id)}
        name='country'
        type='hidden'
      />
      <span className='mt-4 font-medium text-strong'>Beta testing</span>
      <span className='mt-4 text-sm text-medium'>
        We are looking for users to help us test new features before we make
        them generally available. Beta testers will get access to pre-release
        features and in exchange will need to complete some tests to help us
        ensure they're ready for release.
      </span>
      <Checkbox
        id='beta'
        name='beta'
        form='join-waitlist'
        className='mt-6 flex'
      >
        <span className='text-medium'>Yes, sign me up for beta testing</span>
      </Checkbox>
      <Button className='mt-8' form='join-waitlist' type='submit'>
        Join now
      </Button>
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
  const fullName = form.get('fullName') as string
  const email = form.get('email') as string
  const country = form.get('country') as string
  const betaOptIn = form.get('beta') as string

  const fieldErrors = {
    fullName: '',
    countryCode: '',
    email: ''
  }

  let response = await grpcClient
    .joinWaitlist({
      email,
      countryCode: country,
      fullName,
      betaOptIn: betaOptIn != null
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

  return redirect(route('/waitlist/success'))
}
