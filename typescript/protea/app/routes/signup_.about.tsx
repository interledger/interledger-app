import type { FinishedUnaryCall } from '@protobuf-ts/runtime-rpc'
import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Autocomplete,
  Button,
  Card,
  CardContent,
  Layouts,
  TextField
} from '~/components'
import type {
  Country,
  SetSignupUserDataResponse
} from '~/generated/protobuf-ts/backend/v1/backend'
import { flowType, requireFlow, updateFlow } from '~/lib/flows.server'
import { requireNoUserSession } from '~/lib/kratos.server'
import type { GrpcError } from '~/lib/proto.server'
import {
  StatusError,
  grpcClient,
  httpMapping,
  isGrpcError
} from '~/lib/proto.server'
import { canSignup } from '~/lib/signupCheck.server'

export async function loader({ request, params }: LoaderArgs) {
  await requireNoUserSession(request)
  await canSignup(request)
  const flow = await requireFlow(request, flowType.Signup)

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

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/signup'),
      title: 'Profile details'
    }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Sign up | Profile details'
  }
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
    <>
      <Form
        id='signup-about-details'
        action='/signup/about'
        method='post'
        className='hidden'
      />
      <Card>
        <CardContent>
          <p className='text-medium'>
            Please provide a few details about yourself.
          </p>
        </CardContent>
        <TextField
          id='firstName'
          form='signup-about-details'
          label='First name'
          name='firstName'
          labelSuffix='&dagger;'
          defaultValue={flow?.data.firstName}
          type='text'
          className='mt-2'
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
          labelSuffix='&dagger;'
          defaultValue={flow?.data.lastName}
          type='text'
          className='mt-4'
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
          label='Email address'
          name='email'
          defaultValue={flow?.data.email}
          type='text'
          className='mt-4'
          aria-invalid={Boolean(actionData?.errors.email) || undefined}
          aria-describedby={
            actionData?.errors.email ? 'email-error' : undefined
          }
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
          className='mt-4'
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
      </Card>
      <Button form='signup-about-details' type='submit'>
        Continue
      </Button>
      <div className='flex w-full space-x-2'>
        <span className='text-xs text-medium'>
          <sup>&dagger;</sup>
        </span>
        <span className='text-xs text-medium'>
          First and last name as they appear on your government issued ID
          document.
        </span>
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

export async function action({ request }: ActionArgs) {
  await requireFlow(request, flowType.Signup)
  const form = await request.formData()
  const firstName = form.get('firstName') as string
  const lastName = form.get('lastName') as string
  const country = form.get('country') as string
  const email = form.get('email') as string

  // TODO: Determine what countries to let through.
  if (!(country == 'US' || country == 'GB')) {
    return redirect(
      `/waitlist?country=${country}&email=${email}&fullName=${firstName} ${lastName}`
    )
  }

  const fieldErrors = {
    form: '',
    firstName: '',
    lastName: '',
    country: '',
    email: ''
  }

  let response = await grpcClient
    .setSignupUserData({
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

  await updateFlow(request, flowType.Signup, {
    id,
    firstName,
    lastName,
    country,
    email
  })

  return redirect(route('/signup/phone'))
}
