import type { FinishedUnaryCall } from '@protobuf-ts/runtime-rpc'
import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import { Autocomplete, Button, Router, TextField } from '~/components'
import type { Onboarding } from '~/generated/protobuf-ts/backend/v1/backend'
import type { SignupQuery, SignupQueryVariables } from '~/generated/types'
import { SignupDocument } from '~/generated/types'
import { apolloClient } from '~/lib/apollo.server'
import { getCurrentFlow, updateFlow } from '~/lib/flows.server'
import type { GrpcError } from '~/lib/proto.server'
import { grpcClient, isGrpcError, StatusError } from '~/lib/proto.server'

type Country = {
  id: string
  name: string
}

export async function loader({ request, params }: LoaderArgs) {
  const flow = await getCurrentFlow(request, params)
  const countries = await apolloClient
    .query<SignupQuery, SignupQueryVariables>({
      query: SignupDocument,
      context: {
        headers: request.headers
      }
    })
    .then((val) => val.data.countries as Country[])
  return json({
    flow,
    countries
  })
}

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { flow, countries } = useLoaderData<typeof loader>()

  const [country, setCountry] = useState<Country>(
    countries.find(
      (country: Country) => country.id == flow?.data.country
    ) as Country
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
    <>
      <div className='col-span-full flex flex-col space-y-2 pt-4 pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-display text-2xl font-medium'>
          Create a new account
        </span>
        <span>We need to collect some basic information about you.</span>
      </div>
      <Form
        id='signup-about-details'
        action={`/flows/${flow.id}/signup/about`}
        method='post'
        className='hidden'
      />

      <TextField
        id='firstName'
        form='signup-about-details'
        label='First name'
        name='firstName'
        defaultValue={flow?.data.firstName}
        type='text'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.errors.firstName) || undefined}
        aria-describedby={
          actionData?.errors.firstName ? 'firstName-error' : undefined
        }
        required
        errorMessage={actionData?.errors.firstName}
      />

      <TextField
        id='lastName'
        form='signup-about-details'
        label='Last name'
        name='lastName'
        defaultValue={flow?.data.lastName}
        type='text'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.errors.lastName) || undefined}
        aria-describedby={
          actionData?.errors.lastName ? 'lastName-error' : undefined
        }
        required
        errorMessage={actionData?.errors.lastName}
      />
      <Autocomplete
        id='country'
        value={country}
        onChange={setCountry}
        onQuery={setQuery}
        options={filteredCountries}
        label='Country'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.errors.country) || undefined}
        aria-describedby={
          actionData?.errors.country ? 'country-error' : undefined
        }
        errorMessage={actionData?.errors.country}
      />
      <input
        form='signup-about-details'
        value={String(country?.id)}
        name='country'
        type='hidden'
      />

      <TextField
        id='email'
        form='signup-about-details'
        label='Email'
        name='email'
        defaultValue={flow?.data.email}
        type='text'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.errors.email) || undefined}
        aria-describedby={actionData?.errors.email ? 'email-error' : undefined}
        required
        errorMessage={actionData?.errors.email}
      />

      <div className='col-span-full flex items-center justify-between pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Router to={route('/login')}>
          <span className='text-primary'>Already have an account?</span>
        </Router>
        <Button form='signup-about-details' type='submit'>
          Continue
        </Button>
      </div>
    </>
  )
}

// The field names given by the backend for field violations
type fieldErrorsMap = 'FirstName' | 'LastName' | 'CountryOfResidence' | 'Email'

function mapper(
  field: fieldErrorsMap
): 'firstName' | 'lastName' | 'country' | 'email' | null {
  switch (field) {
    case 'FirstName':
      return 'firstName'
    case 'LastName':
      return 'lastName'
    case 'CountryOfResidence':
      return 'country'
    case 'Email':
      return 'email'
    default:
      return null
  }
}

export async function action({ request, params }: ActionArgs) {
  const form = await request.formData()
  const firstName = form.get('firstName') as string
  const lastName = form.get('lastName') as string
  const country = form.get('country') as string
  const email = form.get('email') as string

  const fieldErrors = {
    firstName: '',
    lastName: '',
    country: '',
    email: ''
  }

  let response = await grpcClient
    .updateOnboarding({
      firstName,
      lastName,
      countryOfResidence: country,
      email
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

  const id = (response as FinishedUnaryCall<Onboarding, Onboarding>).response.id

  const headers = await updateFlow(request, {
    id,
    firstName,
    lastName,
    country,
    email
  })
  const flow = await getCurrentFlow(request, params)

  if (country != 'US') {
    return redirect(route('/onboarding/country-access'), { headers })
  }
  return redirect(
    route('/flows/:flowId/signup/phone', {
      flowId: flow?.id as string
    }),
    { headers }
  )
}
