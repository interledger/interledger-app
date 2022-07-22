import { useEffect, useState } from 'react'
import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import {
  Form,
  useActionData,
  useFetcher,
  useLoaderData
} from '@remix-run/react'
import { Autocomplete, Button, Icon, TextField } from '~/components'
import { getCurrentFlow, stepFlow } from '~/lib/flows.server'
import { apolloClient } from '~/lib/apollo.server'
import type { SignupQuery, SignupQueryVariables } from '~/generated/types'
import { SignupDocument } from '~/generated/types'

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

  const country = countries.find((country) => country.id == 'US')

  return json({
    country,
    flow
  })
}

export default function Page() {
  const actionData = useActionData<typeof action>()
  const { country, flow } = useLoaderData<typeof loader>()

  const placeAutocompleteFetcher = useFetcher()
  const geocodeFetcher = useFetcher()

  const [query, setQuery] = useState<string>('')
  const [placeId, setPlaceId] = useState<string>('')

  useEffect(() => {
    if (query.length >= 0) {
      console.log('Calls query', country.id)
      // TODO debounce this: https://developers.google.com/maps/documentation/places/web-service/autocomplete#cost_best_practices
      placeAutocompleteFetcher.load(
        `/api/maps/placesAutocomplete?country=${country.id}&query=${query}`
      )
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query])

  useEffect(() => {
    // When place is set, fetch geocoded address
    if (placeId) geocodeFetcher.load(`/api/maps/geocode?place-id=${placeId}`)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [placeId])

  return (
    <>
      <div className='col-span-full flex flex-col space-y-2 pt-4 pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-display text-2xl font-medium'>
          Physical address
        </span>
      </div>
      <Form
        id='unit-address'
        action={`/flows/${flow.id}/unit-onboarding/address`}
        method='post'
        className='hidden'
      />

      <Autocomplete
        id='street'
        value={geocodeFetcher.data?.street}
        // TODO figure out how to set the value
        onChange={(place) => setPlaceId(place.id)}
        onQuery={setQuery}
        options={
          placeAutocompleteFetcher.data?.predictions || [
            { id: 'skjdhba', name: 'Searching for addresses' }
          ]
        }
        label='Street'
        button={false}
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.fieldErrors?.street) || undefined}
        aria-describedby={
          actionData?.fieldErrors?.street ? 'street-error' : undefined
        }
        errorMessage={actionData?.fieldErrors?.street}
      />
      <input
        form='unit-address'
        value={String(geocodeFetcher.data?.street.name)}
        name='street'
        type='hidden'
      />

      <TextField
        id='apartment'
        form='unit-address'
        label='Apartment, unit, suite or floor number'
        name='apartment'
        defaultValue={actionData?.fields?.apartment || flow?.data.apartment}
        type='text'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.fieldErrors?.apartment) || undefined}
        aria-describedby={
          actionData?.fieldErrors?.apartment ? 'apartment-error' : undefined
        }
        required
        errorMessage={actionData?.fieldErrors?.apartment}
      />
      <TextField
        id='city'
        form='unit-address'
        label='City'
        name='city'
        defaultValue={
          actionData?.fields?.city ||
          flow?.data.city ||
          geocodeFetcher.data?.city
        }
        type='text'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.fieldErrors?.city) || undefined}
        aria-describedby={
          actionData?.fieldErrors?.city ? 'city-error' : undefined
        }
        required
        errorMessage={actionData?.fieldErrors?.city}
      />
      <TextField
        id='state'
        form='unit-address'
        label='State'
        name='state'
        defaultValue={
          actionData?.fields?.state ||
          flow?.data.state ||
          geocodeFetcher.data?.state
        }
        type='text'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.fieldErrors?.state) || undefined}
        aria-describedby={
          actionData?.fieldErrors?.state ? 'state-error' : undefined
        }
        required
        errorMessage={actionData?.fieldErrors?.state}
      />
      <div className='col-span-full mb-4 flex items-center justify-between rounded-xl bg-container p-3 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <div className='flex space-x-3'>
          <Icon>flag</Icon>
          <span className='font-sans text-base font-normal'>
            {country.name}
          </span>
        </div>
      </div>
      <input
        type='hidden'
        form='unit-address'
        name='country'
        value={country.id}
      />
      <TextField
        id='zip'
        form='unit-address'
        label='Postal code'
        name='zip'
        defaultValue={
          actionData?.fields?.zip || flow?.data.zip || geocodeFetcher.data?.zip
        }
        type='text'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.fieldErrors?.zip) || undefined}
        aria-describedby={
          actionData?.fieldErrors?.zip ? 'zip-error' : undefined
        }
        required
        errorMessage={actionData?.fieldErrors?.zip}
      />
      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Button form='unit-address' type='submit'>
          Continue
        </Button>
      </div>
    </>
  )
}

export async function action({ request, params }: ActionArgs) {
  const form = await request.formData()
  const street = form.get('street') as string
  const apartment = form.get('apartment') as string
  const city = form.get('city') as string
  const state = form.get('state') as string
  const country = form.get('country') as string
  const zip = form.get('zip') as string

  const data = {
    street,
    apartment,
    city,
    state,
    country,
    zip
  }

  await stepFlow(request, data)
}
