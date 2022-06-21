import React from 'react'
import type { ActionFunction, LoaderFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Button, TextField } from '~/components'
import { getCurrentFlow, stepFlow } from '~/lib/flows.server'

export const loader: LoaderFunction = async ({ request, params }) => {
  const flow = await getCurrentFlow(request, params)

  return json({
    flow
  })
}

export default function Page() {
  const actionData = useActionData<ActionData>()
  const { flow } = useLoaderData()

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

      <TextField
        id='street'
        form='unit-address'
        label='Street'
        name='street'
        defaultValue={actionData?.fields?.street || flow?.data.street}
        type='text'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.fieldErrors?.street) || undefined}
        aria-describedby={
          actionData?.fieldErrors?.street ? 'street-error' : undefined
        }
        required
        errorMessage={actionData?.fieldErrors?.street}
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
        defaultValue={actionData?.fields?.city || flow?.data.city}
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
        defaultValue={actionData?.fields?.state || flow?.data.state}
        type='text'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.fieldErrors?.state) || undefined}
        aria-describedby={
          actionData?.fieldErrors?.state ? 'state-error' : undefined
        }
        required
        errorMessage={actionData?.fieldErrors?.state}
      />
      <TextField
        id='zip'
        form='unit-address'
        label='Zip code'
        name='zip'
        defaultValue={actionData?.fields?.zip || flow?.data.zip}
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

type ActionData = {
  formError?: string
  fieldErrors?: {
    street: string | undefined
    apartment: string | undefined
    city: string | undefined
    state: string | undefined
    country: string | undefined
    zip: string | undefined
  }
  fields?: {
    street: string
    apartment: string
    city: string
    state: string
    country: string
    zip: string
  }
}
export const action: ActionFunction = async ({ request, params }) => {
  // Should have a fynbos user, that some of the data can be stored against.
  const form = await request.formData()
  // Figure out how the heck we validate an address.
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
