import type { ActionArgs, LoaderArgs, MetaFunction } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import {
  Form,
  useActionData,
  useFetcher,
  useLoaderData
} from '@remix-run/react'
import type { AutocompleteOptions } from '~/components'
import {
  Autocomplete,
  Button,
  Card,
  Icon,
  Layouts,
  TextField
} from '~/components'
import { flowType, requireFlow } from '~/lib/flows.server'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'
import { getUserSession } from '~/lib/kratos.server'
import { useCallback } from 'react'
import type { Address } from '~/generated/protobuf-ts/backend/v1/backend'
import { getClientIP } from '~/lib/ip.server'
import { route } from 'routes-gen'
import { Code } from '~/generated/protobuf-ts/google/rpc/code'

export async function loader({ request }: LoaderArgs) {
  const session = await getUserSession(request)
  let flow = await requireFlow(request, flowType.PersonalDetails)

  return json({
    flow,
    country: { id: session.identity.traits.countryCode, name: '' }
  })
}

export const handle = {
  title: 'Address details',
  layout: Layouts.FocusLayout
}

export const meta: MetaFunction = () => {
  return {
    title: 'Address details'
  }
}

export default function Page() {
  const { flow, country } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()

  const placeAutocompleteFetcher = useFetcher()
  const geocodeFetcher = useFetcher()

  const _onQueryChange = useCallback(
    (query: string) => {
      if (query.length > 0) {
        // TODO debounce this: https://developers.google.com/maps/documentation/places/web-service/autocomplete#cost_best_practices
        placeAutocompleteFetcher.load(
          `/api/maps/placesAutocomplete?country=${country.id}&query=${query}`
        )
      }
    },
    [country.id, placeAutocompleteFetcher]
  )

  const _onPlaceChange = useCallback(
    (place: AutocompleteOptions) => {
      if (place.id)
        geocodeFetcher.load(`/api/maps/geocode?place-id=${place.id}`)
    },
    [geocodeFetcher]
  )

  const formattedAddress =
    geocodeFetcher.data?.formattedAddress ||
    flow?.data?.address.formattedAddress

  return (
    <>
      <Form
        id='personal-details-address'
        action='/personal-details/address'
        method='post'
        className='hidden'
      />
      <Card>
        <p className='text-medium'>Please provide your address.</p>

        <Autocomplete
          id='street'
          value={geocodeFetcher.data?.street}
          onChange={_onPlaceChange}
          onQuery={_onQueryChange}
          options={
            placeAutocompleteFetcher.data?.predictions || [
              { id: 'ID', name: 'Searching for addresses...' }
            ]
          }
          aria-invalid={Boolean(actionData?.errors.address) || undefined}
          aria-describedby={
            actionData?.errors.address ? 'address-error' : undefined
          }
          errorMessage={actionData?.errors.address}
          label='Address'
          prefixIcon={<Icon>search</Icon>}
          button={false}
          className='mt-6'
        />
        <input
          form='personal-details-address'
          value={geocodeFetcher.data?.line1 || flow?.data.address.line1 || ''}
          name='line1'
          type='hidden'
        />
        <input
          form='personal-details-address'
          value={geocodeFetcher.data?.line2 || flow?.data.address.line2 || ''}
          name='line2'
          type='hidden'
        />
        <input
          form='personal-details-address'
          value={
            geocodeFetcher.data?.building || flow?.data.address.building || ''
          }
          name='building'
          type='hidden'
        />
        <input
          form='personal-details-address'
          value={geocodeFetcher.data?.city || flow?.data.address.city || ''}
          name='city'
          type='hidden'
        />
        <input
          form='personal-details-address'
          value={geocodeFetcher.data?.state || flow?.data.address.state || ''}
          name='state'
          type='hidden'
        />
        <input
          form='personal-details-address'
          value={
            geocodeFetcher.data?.zipCode || flow?.data.address.zipCode || ''
          }
          name='zipCode'
          type='hidden'
        />
        <input
          type='hidden'
          form='personal-details-address'
          name='countryCode'
          value={
            geocodeFetcher.data?.countryCode ||
            flow?.data.address.countryCode ||
            ''
          }
        />
        <input
          form='personal-details-address'
          value={formattedAddress || ''}
          name='formattedAddress'
          type='hidden'
        />
        <input
          form='personal-details-address'
          value={
            geocodeFetcher.data?.placeID || flow?.data.address.placeID || ''
          }
          name='placeID'
          type='hidden'
        />
        <TextField
          id='apartment'
          form='personal-details-address'
          label='Building name / number'
          name='apartment'
          defaultValue={flow?.data.address.apartment}
          type='text'
          disabled={!formattedAddress}
          className='mt-6'
        />
      </Card>
      <Button form='personal-details-address' type='submit'>
        Continue
      </Button>
    </>
  )
}

export async function action({ request }: ActionArgs) {
  const form = await request.formData()

  const line1 = form.get('line1') as string
  const line2 = form.get('line2') as string
  const building = form.get('building') as string
  const apartment = form.get('apartment') as string
  const city = form.get('city') as string
  const state = form.get('state') as string
  const zipCode = form.get('zipCode') as string
  const countryCode = form.get('countryCode') as string
  const placeID = form.get('placeID') as string
  const formattedAddress = form.get('formattedAddress') as string

  const address: Address = {
    line1,
    line2,
    building,
    apartment,
    city,
    state,
    zipCode,
    countryCode,
    placeID,
    formattedAddress
  }

  const fieldErrors = {
    form: '',
    address: ''
  }

  const clientIpAddress = getClientIP(request)

  let updateResp = await grpcClient
    .updateIndividualKYC(
      {
        address,
        ipAddress: clientIpAddress
      },
      {
        meta: {
          cookies: request.headers.get('cookie') || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(updateResp)) {
    if (updateResp.code == Code.INVALID_ARGUMENT) {
      fieldErrors.address = 'Your address is not a valid address.'
      return json(
        {
          errors: {
            ...fieldErrors
          }
        },
        { status: 400 }
      )
    }

    throw json({}, httpMapping(updateResp.code))
  }

  // Approve KYC...
  let startKycResp = await grpcClient
    .startKYC(
      {},
      {
        meta: {
          cookies: request.headers.get('cookie') || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(startKycResp)) throw json({}, httpMapping(startKycResp.code))

  // NOTE Temporarily not exciting this flow so that if the user needs to fix something their data will be there.
  // We should find a better way to do this.
  // await exitFlow(request, flowType.PersonalDetails)
  return redirect(route('/'))
}
