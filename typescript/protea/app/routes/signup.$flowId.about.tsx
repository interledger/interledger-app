import type { FinishedUnaryCall } from '@protobuf-ts/runtime-rpc'
import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import { Autocomplete, Button, Shape, TextField } from '~/components'
import {
  flowType,
  getCurrentFlow,
  requireFlow,
  updateFlow
} from '~/lib/flows.server'
import type { GrpcError } from '~/lib/proto.server'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'
import type {
  Country,
  SetSignupUserDataResponse
} from '~/generated/protobuf-ts/backend/v1/backend'

export async function loader({ request, params }: LoaderArgs) {
  await requireFlow(request, flowType.Signup, params)
  const flow = await getCurrentFlow(request, flowType.Signup)

  let response = await grpcClient
    .getCountries({})
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  return json({
    flow,
    countries: response.response.countries
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
        countries.filter((country) => {
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
      <div className='col-span-full flex justify-between pb-4 sm:col-span-6 sm:col-start-2'>
        <span className='font-display text-2xl font-medium'>
          Your personal details
        </span>
        <div className='hidden sm:flex'>
          <Shape
            width={'w-8'}
            radius={'rounded-tl-full'}
            color={'bg-yellow-300'}
          />
          <Shape
            width={'w-8'}
            radius={'rounded-bl-full'}
            color={'bg-slate-500'}
          />
        </div>
      </div>
      <Form
        id='signup-about-details'
        action={`/signup/${flow.id}/about`}
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
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2'
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
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2'
        aria-invalid={Boolean(actionData?.errors.lastName) || undefined}
        aria-describedby={
          actionData?.errors.lastName ? 'lastName-error' : undefined
        }
        required
        errorMessage={actionData?.errors.lastName}
      />

      <TextField
        id='email'
        form='signup-about-details'
        label='Email'
        name='email'
        defaultValue={flow?.data.email}
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

      <div className='col-span-full flex items-center justify-end pt-4 sm:col-span-6 sm:col-start-2'>
        <Button form='signup-about-details' type='submit'>
          Continue
        </Button>
      </div>
    </div>
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

  // TODO: Determine what countries to let through.
  if (country != 'US') {
    return redirect(
      `/waitlist?country=${country}&email=${email}&fullName=${firstName} ${lastName}`
    )
  }

  const fieldErrors = {
    firstName: '',
    lastName: '',
    country: '',
    email: ''
  }

  let response = await grpcClient
    .setSignupUserData({
      id: params.flowId,
      firstName,
      lastName,
      countryCode: country,
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
    } else throw json({}, httpMapping(response.code))
  }

  const id = (
    response as FinishedUnaryCall<
      SetSignupUserDataResponse,
      SetSignupUserDataResponse
    >
  ).response.id

  const headers = await updateFlow(request, flowType.Signup, {
    id,
    firstName,
    lastName,
    country,
    email
  })

  return redirect(
    route('/signup/:flowId/phone', {
      flowId: params.flowId as string
    }),
    { headers }
  )
}
