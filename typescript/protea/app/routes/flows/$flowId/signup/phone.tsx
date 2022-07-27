import { useEffect, useState } from 'react'
import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { redirect } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Autocomplete, Button, TextField } from '~/components'
import { getCurrentFlow, updateFlow } from '~/lib/flows.server'
import type { GrpcError } from '~/lib/proto.server'
import { grpcClient, StatusError, isGrpcError } from '~/lib/proto.server'
import type { CountryCode } from 'libphonenumber-js'
import { parsePhoneNumber } from 'libphonenumber-js'
import { apolloClient } from '~/lib/apollo.server'
import type { SignupQuery, SignupQueryVariables } from '~/generated/types'
import { SignupDocument } from '~/generated/types'
import { route } from 'routes-gen'

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
    .then((val) => val.data.countries)

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
          Verify your phone number
        </span>
        <span>
          We require a verified phone number so that we can contact you if we
          need to.
        </span>
      </div>
      <Form
        id='signup-phone-details'
        action={`/flows/${flow.id}/signup/phone`}
        method='post'
        className='hidden'
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
        form='signup-phone-details'
        value={String(country?.id)}
        name='country'
        type='hidden'
      />

      <TextField
        id='phone'
        form='signup-phone-details'
        label='Phone number'
        name='phone'
        defaultValue={flow?.data.phone}
        type='tel'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.errors.phone) || undefined}
        aria-describedby={actionData?.errors.phone ? 'phone-error' : undefined}
        required
        errorMessage={actionData?.errors.phone}
      />

      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Button form='signup-phone-details' type='submit'>
          Send SMS
        </Button>
      </div>
    </>
  )
}

// The field names given by the backend for field violations
type fieldErrorsMap = 'To'

function mapper(field: fieldErrorsMap): 'phone' | null {
  switch (field) {
    case 'To':
      return 'phone'
    default:
      return null
  }
}

export async function action({ request, params }: ActionArgs) {
  const form = await request.formData()
  const country = form.get('country') as string
  const phone = form.get('phone') as string

  const fieldErrors = {
    country: '',
    phone: ''
  }

  const flow = await getCurrentFlow(request, params)
  const onboardingId = flow?.data.id

  const phoneNumber = parsePhoneNumber(phone, country as CountryCode)

  let response = await grpcClient
    .sendPhoneVerification({
      to: phoneNumber.number,
      onboardingId
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

  const headers = await updateFlow(request, {
    phone: phoneNumber.number
  })
  return redirect(
    route('/flows/:flowId/signup/sms', {
      flowId: flow?.id as string
    }),
    { headers }
  )
}
