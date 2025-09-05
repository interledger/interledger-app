import { Code } from '@bufbuild/connect'
import type {
  ActionFunctionArgs,
  LoaderFunctionArgs,
  MetaFunction
} from '@remix-run/node'
import { redirect } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Autocomplete,
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Checkbox,
  Layouts,
  TextField
} from '~/components'
import { jsonWithCSRF, validateCSRFToken } from '~/lib/csrf.server'
import { isConnectError } from '~/lib/error.server'
import { grpc } from '~/lib/grpc.server'
import { requireNoUserSession } from '~/lib/kratos.server'
import { mergeMeta } from '~/lib/meta'

type Country = {
  id: string
  name: string
}

export async function loader({ request }: LoaderFunctionArgs) {
  await requireNoUserSession(request)
  let response = await grpc.getCountries(request, {})

  if (isConnectError(response)) throw response.errorResponse

  const countries = response.countries

  const url = new URL(request.url)
  const mugId = url.searchParams.get('mug')
  const countryCode = url.searchParams.get('country')
  const email = url.searchParams.get('email')
  const fullName = url.searchParams.get('fullName')

  let isMugAvailable = false
  if (mugId != null) {
    let response = await grpc.isMugAvailable(request, {
      mugId: mugId
    })

    if (isConnectError(response)) throw response.errorResponse

    isMugAvailable = response.available
  }

  return jsonWithCSRF(request, {
    mug: {
      id: mugId ?? undefined,
      available: isMugAvailable
    },
    countryCode,
    countries,
    email,
    fullName
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/'),
      title: 'Join the waitlist'
    }
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Waitlist'
  }
])

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { mug, countryCode, countries, email, fullName, csrfToken } =
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
    <>
      <Form
        id='join-waitlist'
        action='/waitlist'
        method='post'
        className='hidden'
      />
      <input
        form='join-waitlist'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
      />
      <input
        id='country'
        form='join-waitlist'
        value={String(country?.id)}
        name='country'
        type='hidden'
      />
      <Card>
        <CardContent>
          {!mug.available && (
            <span className='text-medium'>
              Leave your details below and we will notify you as soon as
              enrollment opens.
            </span>
          )}
          {mug.available && (
            <>
              <span className='text-2xl font-medium'>Congratulations!</span>
              <div className='mt-4 flex flex-col space-y-4 sm:flex-row-reverse sm:items-center sm:space-x-6 sm:space-y-0 sm:space-x-reverse'>
                <span className='text-medium'>
                  You got your hands on a limited edition Interledger Wallet
                  mug.
                </span>
                <img
                  className='w-full sm:w-2/5'
                  alt='Interledger Wallet mug'
                  // TODO: Use our own CDN
                  src='https://cdn.fynbos.app/marketing/enamel-mug-waitlist.webp'
                />
              </div>
              <span className='mt-6 text-medium'>
                Each mug has a unique wallet address - now this one could be
                yours.
              </span>
              <span className='mt-4 text-medium'>
                Sign up to the waitlist and we'll link this mug's wallet address
                to your Interledger Wallet.
              </span>
              <span className='mt-4 text-medium'>
                If you’re already on the waitlist, submit your details again and
                we’ll link your mug to your existing details.
              </span>
              <span className='mt-6 font-medium'>Join the waitlist</span>
              <input
                type='hidden'
                form='join-waitlist'
                name='mugId'
                value={mug.id as string}
              />
            </>
          )}
        </CardContent>
        <TextField
          id='fullName'
          form='join-waitlist'
          label='Full name'
          name='fullName'
          type='text'
          defaultValue={fullName as string}
          className='mt-2'
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
          aria-invalid={Boolean(actionData?.errors.countryCode) || undefined}
          aria-describedby={
            actionData?.errors.countryCode ? 'country-error' : undefined
          }
          errorMessage={actionData?.errors.countryCode}
        />
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Beta testing</CardTitle>
        </CardHeader>
        <CardContent>
          <span className='text-sm text-medium'>
            We are looking for users to help us test new features before we make
            them generally available. Beta testers will get access to
            pre-release features and in exchange will need to complete some
            tests to help us ensure they're ready for release.
          </span>
          <Checkbox
            id='beta'
            name='beta'
            form='join-waitlist'
            className='mt-6 flex'
          >
            Yes, sign me up for beta testing
          </Checkbox>
        </CardContent>
      </Card>
      <Button form='join-waitlist' type='submit'>
        Join now
      </Button>
    </>
  )
}

export async function action({ request }: ActionFunctionArgs) {
  const form = await request.formData()
  const fullName = form.get('fullName') as string
  const email = form.get('email') as string
  const country = form.get('country') as string
  const betaOptIn = form.get('beta') as string

  await validateCSRFToken(request, form)

  const errors = {
    form: '',
    fullName: '',
    countryCode: '',
    email: ''
  }

  let response = await grpc.joinWaitlist(request, {
    email,
    countryCode: country,
    fullName,
    betaOptIn: betaOptIn != null
  })

  if (isConnectError(response)) {
    if (response.code == Code.InvalidArgument) {
      return response.error({ errors })
    } else return response.error({ errors }, {}, { action: 'Contact support' })
  }

  return redirect(route('/waitlist/success'))
}
