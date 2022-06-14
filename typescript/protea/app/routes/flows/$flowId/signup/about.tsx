import React, { useEffect, useState } from 'react'
import type { ActionFunction, LoaderFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Autocomplete, Button, Router, TextField } from '~/components'
import { getCurrentFlow, stepFlow } from '~/lib/flows.server'
import type { GrpcError } from '~/lib/proto.server'
import { grpcClient, StatusError, isGrpcError } from '~/lib/proto.server'
import type { SignupQuery, SignupQueryVariables } from '~/generated/types'
import { SignupDocument } from '~/generated/types'
import { apolloClient } from '~/lib/apollo.server'
import type { FinishedUnaryCall } from '@protobuf-ts/runtime-rpc'
import type { Onboarding } from '~/generated/protobuf-ts/backend/v1/backend'
import { route } from 'routes-gen'

type Country = {
  id: string
  name: string
}

export const loader: LoaderFunction = async ({ request, params }) => {
  const flow = await getCurrentFlow(request, params)
  const countries = await apolloClient
    .query<SignupQuery, SignupQueryVariables>({
      query: SignupDocument,
      context: {
        headers: request.headers
      }
    })
    .then((val) => val.data.countries)

  return json({
    flow,
    countries
  })
}

export default function Page() {
  const actionData = useActionData<ActionData>()
  const { flow, countries } = useLoaderData()

  const [country, setCountry] = useState<Country>(
    countries.find(
      (country: Country) =>
        country.id == actionData?.fields?.country ||
        country.id == flow?.data.country
    )
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
        defaultValue={actionData?.fields?.firstName || flow?.data.firstName}
        type='text'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.fieldErrors?.firstName) || undefined}
        aria-describedby={
          actionData?.fieldErrors?.firstName ? 'firstName-error' : undefined
        }
        required
        errorMessage={actionData?.fieldErrors?.firstName}
      />

      <TextField
        id='lastName'
        form='signup-about-details'
        label='Last name'
        name='lastName'
        defaultValue={actionData?.fields?.lastName || flow?.data.lastName}
        type='text'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.fieldErrors?.lastName) || undefined}
        aria-describedby={
          actionData?.fieldErrors?.lastName ? 'lastName-error' : undefined
        }
        required
        errorMessage={actionData?.fieldErrors?.lastName}
      />
      <Autocomplete
        id='country'
        value={country}
        onChange={setCountry}
        onQuery={setQuery}
        options={filteredCountries}
        label='Country'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.fieldErrors?.country) || undefined}
        aria-describedby={
          actionData?.fieldErrors?.country ? 'country-error' : undefined
        }
        errorMessage={actionData?.fieldErrors?.country}
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
        defaultValue={actionData?.fields?.email || flow?.data.email}
        type='text'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.fieldErrors?.email) || undefined}
        aria-describedby={
          actionData?.fieldErrors?.email ? 'email-error' : undefined
        }
        required
        errorMessage={actionData?.fieldErrors?.email}
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

type ActionData = {
  formError?: string
  fieldErrors?: {
    firstName: string | undefined
    lastName: string | undefined
    country: string | undefined
    email: string | undefined
  }
  fields?: {
    firstName: string
    lastName: string
    country: string
    email: string
  }
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

/**
 * parseError handles potention errors from grpc client calls.
 * @param response The response from a grpc call
 * @param fields Any data passed to the grpc call
 * @returns ActionData response for field validation errors, throws other errors or null if not an error
 */
function parseError(response: any, fields: any): Response | null {
  if (isGrpcError(response)) {
    if (response.code == 3) {
      let fieldErrors: ActionData['fieldErrors'] = {
        firstName: undefined,
        lastName: undefined,
        country: undefined,
        email: undefined
      }
      for (let violation of (response as GrpcError).details[0]
        .fieldViolations) {
        const field = mapper(violation.field as fieldErrorsMap)
        if (field != null) fieldErrors[field] = violation.description
      }
      return json({
        fields,
        fieldErrors
      })
    } else throw response
  }
  return null
}

export const action: ActionFunction = async ({ request }) => {
  const form = await request.formData()
  const firstName = form.get('firstName') as string
  const lastName = form.get('lastName') as string
  const country = form.get('country') as string
  const email = form.get('email') as string

  let call = await grpcClient
    .updateOnboarding({
      firstName,
      lastName,
      countryOfResidence: country,
      email
    })
    .then((v) => v)
    .catch(StatusError)

  const actionData = parseError(call, {
    firstName,
    lastName,
    country,
    email
  })

  if (actionData != null) return actionData
  const id = (call as FinishedUnaryCall<Onboarding, Onboarding>).response.id
  await stepFlow(request, {
    id,
    firstName,
    lastName,
    country,
    email
  })
}
