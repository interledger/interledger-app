import { Autocomplete, Button, TextField } from '~/components'
import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { apolloClient } from '~/lib/apollo.server'
import type { SignupQuery, SignupQueryVariables } from '~/generated/types'
import { SignupDocument } from '~/generated/types'
import type { GrpcError } from '~/lib/proto.server'
import { httpMapping } from '~/lib/proto.server'
import { grpcClient, isGrpcError, StatusError } from '~/lib/proto.server'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { requireNoUserSession } from '~/lib/kratos.server'
import { useEffect, useState } from 'react'

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

const shapes = [
  [
    'bg-slate-600 rounded-tl-full',
    'bg-transparent',
    'bg-yellow-400 rounded-tr-full',
    'bg-rose-300 rounded-tl-full',
    'bg-lime-400 rounded-full',
    'bg-transparent',
    'bg-rose-500 rounded-full',
    'bg-lime-300 rounded-tr-full',
    'bg-transparent',
    'bg-transparent'
  ],
  [
    'bg-transparent',
    'bg-rose-400 rounded-full',
    'bg-lime-500 rounded-bl-full',
    'bg-transparent',
    'bg-slate-300 rounded-tl-full',
    'bg-yellow-200 rounded-tl-full',
    'bg-slate-500 rounded-br-full',
    'bg-transparent',
    'bg-rose-100 rounded-full',
    'bg-rose-300 rounded-bl-full'
  ]
]

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
    <div className='mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto rounded-2xl bg-page px-4 pb-16 pt-6 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:pt-12 xl:max-w-4xl'>
      <div className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2'>
        {shapes.map((shapeRow) => (
          <div className='flex' key={shapeRow.toString()}>
            {shapeRow.map((shape, index) => (
              <div
                key={shape + index}
                className={`aspect-square w-full ${shape}`}
              />
            ))}
          </div>
        ))}
      </div>
      <div className='col-span-full flex flex-col space-y-2 pt-4 pb-8 sm:col-span-6 sm:col-start-2'>
        <span className='font-display text-2xl font-medium'>
          Join the waitlist
        </span>
        <span className='text-medium'>
          Leave your details below and we will notify you as soon as enrollment
          opens.
        </span>
      </div>

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
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2'
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
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2'
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
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2'
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
      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2'>
        <Button form='join-waitlist' type='submit'>
          Join now
        </Button>
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
  const fullName = form.get('fullName') as string
  const email = form.get('email') as string
  const country = form.get('country') as string

  const fieldErrors = {
    fullName: '',
    countryCode: '',
    email: ''
  }

  let response = await grpcClient
    .joinWaitlist({
      email,
      countryCode: country,
      fullName
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

  // Redirect to joined waitlist
  return redirect('/waitlist/success')
}
