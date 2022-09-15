import { useEffect, useState } from 'react'
import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { redirect } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Autocomplete, Button, Checkbox, Router, TextField } from '~/components'
import { exitFlow, getCurrentFlow } from '~/lib/flows.server'
import { apolloClient } from '~/lib/apollo.server'
import type { SignupQuery, SignupQueryVariables } from '~/generated/types'
import { SignupDocument } from '~/generated/types'
import { DateTime } from 'luxon'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'
import { route } from 'routes-gen'
import { useScript } from '~/lib/useScript'

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

  const [blackbox, setBlackbox] = useState<string>('')
  const status = useScript('https://ci-mpsnare.iovation.com/snare.js')
  useEffect(() => {
    if (status == 'ready') {
      setBlackbox((window as any).ioGetBlackbox().blackbox)
    }
  }, [status])

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
        defaultValue={flow?.data.birth}
        type='date'
        max={DateTime.now().toFormat('yyyy-LL-dd')}
        className='col-span-full flex flex-col selection:bg-primary/50 sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.errors.birth) || undefined}
        aria-describedby={actionData?.errors.birth ? 'birth-error' : undefined}
        required
        errorMessage={actionData?.errors.birth}
      />
      <Autocomplete
        id='country'
        value={country}
        onChange={setCountry}
        onQuery={setQuery}
        options={filteredCountries}
        label='Nationality'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.errors.country) || undefined}
        aria-describedby={
          actionData?.errors.country ? 'country-error' : undefined
        }
        errorMessage={actionData?.errors.country}
      />
      <input
        form='unit-about'
        value={String(country?.id)}
        name='country'
        type='hidden'
      />
      <input form='unit-about' value={blackbox} name='blackbox' type='hidden' />
      <TextField
        id='ssn'
        form='unit-about'
        label={
          country?.id == 'US' ? 'Social service number' : 'Passport number'
        }
        name='ssn'
        defaultValue={flow?.data.ssn}
        type='text'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.errors.ssn) || undefined}
        aria-describedby={actionData?.errors.ssn ? 'ssn-error' : undefined}
        required
        errorMessage={actionData?.errors.ssn}
      />
      <Checkbox
        id='debit-card-agreement'
        name='debit-card-agreement'
        form='unit-about'
        className='col-span-full mt-4 flex sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.errors.serviceAgreement) || undefined}
        aria-describedby={
          actionData?.errors.serviceAgreement
            ? 'serviceAgreement-error'
            : undefined
        }
        errorMessage={actionData?.errors.serviceAgreement}
      >
        I have read and agree to the&nbsp;
        {/* TODO: Should route to legal */}
        <Router className='text-primary' to='/privacy-policy'>
          Debit card agreement
        </Router>
        &nbsp;, and I consent to the use of electronic records in connection
        with this application.
      </Checkbox>
      <Checkbox
        id='service-agreement'
        name='service-agreement'
        form='unit-about'
        className='col-span-full flex sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.errors.serviceAgreement) || undefined}
        aria-describedby={
          actionData?.errors.serviceAgreement
            ? 'serviceAgreement-error'
            : undefined
        }
        errorMessage={actionData?.errors.serviceAgreement}
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
  const blackbox = form.get('blackbox') as string
  // const serviceAgreement = form.get('service-agreement') as string

  // TODO: validate this somewhere
  const fieldErrors = {
    ssn: '',
    birth: '',
    country: '',
    serviceAgreement: ''
  }

  const flow = await getCurrentFlow(request, params)
  const { street, apartment, city, state, zip } = flow?.data
  const deviceFingerprints = [blackbox]

  // This won't return data, but should notify success. Can just forward to a waiting page.
  let response = await grpcClient
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

  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  const headers = await exitFlow(request)

  if (response.status.code == 'OK') return redirect(route('/'), { headers })
  return json({ errors: { ...fieldErrors } }, { status: 400 })
}
