import { useEffect, useState } from 'react'
import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Autocomplete, Button, Checkbox, Router, TextField } from '~/components'
import { getCurrentFlow, stepFlow } from '~/lib/flows.server'
import { apolloClient } from '~/lib/apollo.server'
import type { SignupQuery, SignupQueryVariables } from '~/generated/types'
import { SignupDocument } from '~/generated/types'
import { DateTime } from 'luxon'
import { grpcClient, StatusError } from '~/lib/proto.server'

type Country = {
  id: string
  name: string
}

export async function loader({ request, params }: LoaderArgs) {
  const flow = await getCurrentFlow(request, params)

  // TODO fetch the users country

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
      (country: Country) =>
        country.id == actionData?.fields?.country ||
        country.id == flow?.data.country
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
          Individual information
        </span>
        <span>We need to collect some identifying information about you.</span>
      </div>
      <Form
        id='unit-about'
        action={`/flows/${flow.id}/unit-onboarding/about`}
        method='post'
        className='hidden'
      />

      <TextField
        id='birth'
        form='unit-about'
        label='Birth date'
        name='birth'
        defaultValue={actionData?.fields?.birth || flow?.data.birth}
        type='date'
        max={DateTime.now().toFormat('yyyy-LL-dd')}
        className='col-span-full flex flex-col selection:bg-primary/50 sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.fieldErrors?.birth) || undefined}
        aria-describedby={
          actionData?.fieldErrors?.birth ? 'birth-error' : undefined
        }
        required
        errorMessage={actionData?.fieldErrors?.birth}
      />
      <Autocomplete
        id='country'
        value={country}
        onChange={setCountry}
        onQuery={setQuery}
        options={filteredCountries}
        label='Nationality'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.fieldErrors?.country) || undefined}
        aria-describedby={
          actionData?.fieldErrors?.country ? 'country-error' : undefined
        }
        errorMessage={actionData?.fieldErrors?.country}
      />
      <input
        form='unit-about'
        value={String(country?.id)}
        name='country'
        type='hidden'
      />
      <TextField
        id='ssn'
        form='unit-about'
        label={
          country?.id == 'US' ? 'Social service number' : 'Passport number'
        }
        name='ssn'
        defaultValue={actionData?.fields?.ssn || flow?.data.ssn}
        type='text'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.fieldErrors?.ssn) || undefined}
        aria-describedby={
          actionData?.fieldErrors?.ssn ? 'ssn-error' : undefined
        }
        required
        errorMessage={actionData?.fieldErrors?.ssn}
      />
      <Checkbox
        id='service-agreement'
        name='service-agreement'
        form='signup-password'
        className='col-span-full mt-4 flex sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={
          Boolean(actionData?.fieldErrors?.serviceAgreement) || undefined
        }
        aria-describedby={
          actionData?.fieldErrors?.serviceAgreement
            ? 'serviceAgreement-error'
            : undefined
        }
        errorMessage={actionData?.fieldErrors?.serviceAgreement}
      >
        I agree to the Fynbos&nbsp;
        <Router className='text-primary' to='/privacy-policy'>
          Deposit Terms &amp; Conditions
        </Router>
      </Checkbox>

      <span className='col-span-full justify-end pt-4 text-xs sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        By filling out this application, you understand and agree that Unit's
        use of your data is governed by its{' '}
        <Router.a
          className='text-primary'
          to='https://www.unit.co/privacy-policy'
        >
          Privacy Policy
        </Router.a>
        .
      </span>

      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Button form='unit-about' type='submit'>
          Continue
        </Button>
      </div>
    </>
  )
}

export async function action({ request, params }: ActionArgs) {
  const cookie = request.headers.get('Cookie') as string
  const form = await request.formData()
  const ssn = form.get('ssn') as string
  const dateOfBirth = form.get('birth') as string
  const nationality = form.get('country') as string
  const serviceAgreement = form.get('service-agreement') as string

  const flow = await getCurrentFlow(request, params)
  const { street, apartment, city, state, zip } = flow?.data
  const deviceFingerprints = ['TODO']
  console.log(flow)

  // This won't return data, but should notify success. Can just forward to a waiting page.
  let call = await grpcClient
    .initiateUnitOnboarding(
      {
        ssn: ssn,
        nationality: nationality,
        dateOfBirth: dateOfBirth,
        street: street,
        street2: apartment,
        city: city,
        state: state,
        postalCode: zip,
        ip: '41.71.7.119',
        deviceFingerprints: deviceFingerprints
      },
      {
        meta: {
          cookies: cookie
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)

  console.log(call)
  return null
  // TODO if call.status.code == 'OK' then redirect to waiting page.
  // if (call.status.code)

  // const data = {
  //   street,
  //   apartment,
  //   city,
  //   state,
  //   country,
  //   zip
  // }

  // await stepFlow(request, data)
}
