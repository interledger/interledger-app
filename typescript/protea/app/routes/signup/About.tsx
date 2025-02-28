import { useFetcher, useLoaderData } from '@remix-run/react'
import { useEffect, useState } from 'react'
import { route } from 'routes-gen'
import {
  Autocomplete,
  Button,
  Card,
  CardContent,
  CardIcon,
  Icon,
  TextField
} from '~/components'
import { SignupStep, useSignupStore } from '~/lib/useSignupStore'
import type { detailsAction, loader } from './route'

export function isEUCountry(countryCode: string) {
  const euCountryCodes = [
    'AT',
    'BE',
    'BG',
    'HR',
    'CY',
    'CZ',
    'DK',
    'EE',
    'FI',
    'FR',
    'DE',
    'GR',
    'HU',
    'IE',
    'IT',
    'LV',
    'LT',
    'LU',
    'MT',
    'NL',
    'PL',
    'PT',
    'RO',
    'SK',
    'SI',
    'ES',
    'SE'
  ]

  return euCountryCodes.includes(countryCode.toUpperCase())
}

export function About() {
  const details = useFetcher<typeof detailsAction>()
  const { csrfToken } = useLoaderData<typeof loader>()

  const [
    firstName,
    lastName,
    email,
    country,
    countries,
    setCountry,
    setDetails,
    setStep
  ] = useSignupStore((state) => [
    state.firstName,
    state.lastName,
    state.email,
    state.country,
    state.countries,
    state.setCountry,
    state.setDetails,
    state.setStep
  ])

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

  useEffect(() => {
    if (details.data?.id) {
      setDetails(
        details.data?.id,
        details.data?.firstName.replace(/\s+/g, ' ').trim(),
        details.data?.lastName.replace(/\s+/g, ' ').trim(),
        details.data?.email
      )
      setStep(SignupStep.PHONE)
    }
  }, [details.data, setDetails, setStep])

  return (
    <>
      <details.Form
        id='signup-about-details'
        action={route('/signup')}
        method='post'
        className='hidden'
      />
      <input
        form='signup-about-details'
        value={csrfToken}
        name='csrfToken'
        type='hidden'
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
          defaultValue={firstName}
          type='text'
          className='mt-2'
          aria-invalid={Boolean(details.data?.errors?.firstName) || undefined}
          aria-describedby={
            details.data?.errors?.firstName ? 'firstName-error' : undefined
          }
          required
          errorMessage={details.data?.errors?.firstName}
        />

        <TextField
          id='lastName'
          form='signup-about-details'
          label='Last name'
          name='lastName'
          labelSuffix='&dagger;'
          defaultValue={lastName}
          type='text'
          className='mt-4'
          aria-invalid={Boolean(details.data?.errors?.lastName) || undefined}
          aria-describedby={
            details.data?.errors?.lastName ? 'lastName-error' : undefined
          }
          required
          errorMessage={details.data?.errors?.lastName}
        />

        <TextField
          id='email'
          form='signup-about-details'
          label='Email address'
          name='email'
          defaultValue={email}
          type='text'
          className='mt-4'
          aria-invalid={Boolean(details.data?.errors?.email) || undefined}
          aria-describedby={
            details.data?.errors?.email ? 'email-error' : undefined
          }
          required
          errorMessage={details.data?.errors?.email}
        />

        <Autocomplete
          id='country'
          value={country ?? undefined}
          onChange={setCountry}
          onQuery={setQuery}
          options={filteredCountries}
          label='Country of residence'
          className='mt-4'
          aria-invalid={Boolean(details.data?.errors?.country) || undefined}
          aria-describedby={
            details.data?.errors?.country ? 'country-error' : undefined
          }
          errorMessage={details.data?.errors?.country}
        />
        <input
          form='signup-about-details'
          value={String(country?.id)}
          name='country'
          type='hidden'
        />
      </Card>
      {country &&
        country?.id != 'CA' &&
        country?.id !== 'US' &&
        country?.id !== 'ZA' &&
        !isEUCountry(country?.id) && (
          <Card>
            <CardContent>
              <div className='flex items-center space-x-4'>
                <CardIcon className='!bg-error'>
                  <Icon className='text-error'>exclamation</Icon>
                </CardIcon>
                <div className='flex flex-col space-y-1'>
                  <p className='text-sm text-medium'>
                    Country not yet supported. Click "Continue" to join the
                    waitlist.
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
      <Button
        form='signup-about-details'
        name='formName'
        value='details'
        type='submit'
      >
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
