import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import {
  Form,
  useActionData,
  useFetcher,
  useLoaderData
} from '@remix-run/react'
import {
  AnchorRouter,
  Autocomplete,
  Button,
  Checkbox,
  Icon,
  Layouts,
  Shape,
  TextField
} from '~/components'
import { exitFlow, flowType, requireFlow } from '~/lib/flows.server'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'
import { requireUserSession } from '~/lib/kratos.server'
import { useCallback } from 'react'
import type { Address } from '~/generated/protobuf-ts/backend/v1/backend'
import { getClientIP } from '~/lib/ip.server'

export async function loader({ request }: LoaderArgs) {
  const session = await requireUserSession(request)
  let flow = await requireFlow(request, flowType.PersonalDetails)

  return json({
    flow,
    country: { id: session.identity.traits.countryCode, name: '' }
  })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const { flow, country } = useLoaderData<typeof loader>()
  const actionData = useActionData<typeof action>()

  const placeAutocompleteFetcher = useFetcher()
  const geocodeFetcher = useFetcher()

  const _onQueryChange = useCallback(
    (query) => {
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
    (place) => {
      if (place.id)
        geocodeFetcher.load(`/api/maps/geocode?place-id=${place.id}`)
    },
    [geocodeFetcher]
  )

  const formattedAddress =
    geocodeFetcher.data?.formattedAddress ||
    flow?.data?.address.formattedAddress

  return (
    <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
      <div className='flex justify-between'>
        <h1 className='font-display text-2xl font-medium'>Address details</h1>
        <div className='hidden sm:flex'>
          <Shape
            width={'w-8'}
            radius={'rounded-full'}
            color={'bg-yellow-300'}
          />
          <Shape
            width={'w-8'}
            radius={'rounded-bl-full'}
            color={'bg-slate-300'}
          />
        </div>
      </div>
      <p className='mt-6 text-medium'>Please provide your address.</p>

      <Form
        id='personal-details-address'
        action='/personal-details/address'
        method='post'
        className='hidden'
      />
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
        value={geocodeFetcher.data?.zipCode || flow?.data.address.zipCode || ''}
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
        value={geocodeFetcher.data?.placeID || flow?.data.address.placeID || ''}
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
        className='mt-1'
      />
      <Checkbox
        id='service-agreement'
        name='service-agreement'
        form='personal-details-address'
        className='mt-4 flex'
        aria-invalid={Boolean(actionData?.errors.serviceAgreement) || undefined}
        aria-describedby={
          actionData?.errors.serviceAgreement
            ? 'serviceAgreement-error'
            : undefined
        }
        errorMessage={actionData?.errors.serviceAgreement}
      >
        I agree to the Machnet&nbsp;
        <AnchorRouter
          className='text-primary'
          to={
            'https://machnetservices.com/fynbos-technologies-llc-privacypolicy/'
          }
        >
          Privacy Policy
        </AnchorRouter>
        &nbsp;and&nbsp;
        <AnchorRouter
          className='text-primary'
          to={
            'https://machnetservices.com/fynbos-technologies-llc-termsofservice/'
          }
        >
          Terms of Service
        </AnchorRouter>
        .
      </Checkbox>

      <Button className='mt-12' form='personal-details-address' type='submit'>
        Continue
      </Button>
    </div>
  )
}

export async function action({ request }: ActionArgs) {
  const form = await request.formData()

  const serviceAgreement = form.get('service-agreement') as string
  // TODO: use actual service agreement service
  const fieldErrors = {
    serviceAgreement: '',
    address: ''
  }

  if (serviceAgreement == null) {
    fieldErrors.serviceAgreement = 'You are required to agree to continue.'
    return json(
      {
        errors: {
          ...fieldErrors
        }
      },
      { status: 400 }
    )
  }

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

  let isUpsAddressResponse = await grpcClient.isUSPSAddress(address, {
    meta: {
      cookies: request.headers.get('cookie') || ''
    }
  })

  if (!isUpsAddressResponse.response.valid) {
    fieldErrors.address = 'Your address is not a valid USPS address.'
    return json(
      {
        errors: {
          ...fieldErrors
        }
      },
      { status: 400 }
    )
  }

  const clientIpAddress = getClientIP(request)

  let response = await grpcClient
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

  if (isGrpcError(response)) throw json({}, httpMapping(response.code))

  let res = await grpcClient
    .startMachnetKYC(
      {},
      {
        meta: {
          cookies: request.headers.get('cookie') || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(res)) throw json({}, httpMapping(res.code))

  const flow = await requireFlow(request, flowType.PersonalDetails)
  await exitFlow(request, flowType.PersonalDetails)
  return redirect(flow.returnTo)
}
