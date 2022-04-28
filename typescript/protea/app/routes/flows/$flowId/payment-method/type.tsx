import React, { useState } from 'react'
import type { ActionFunction, LoaderFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import type { RadioGroupOption } from '~/components'
import { Icon } from '~/components'
import { Button, RadioGroup } from '~/components'
import { getCurrentFlow, stepFlow } from '~/lib/flows.server'

export const loader: LoaderFunction = async ({ request, params }) => {
  const flow = await getCurrentFlow(request, params)

  const paymentTypes: RadioGroupOption[] = [
    {
      id: 'bank',
      name: 'Bank account',
      description: '',
      icon: 'account_balance'
    },
    {
      id: 'card',
      name: 'Card',
      description: '',
      icon: 'credit_card',
      disabled: true
    }
  ]

  return json({
    paymentTypes,
    flow
  })
}

export default function Page() {
  const { paymentTypes, flow } = useLoaderData()

  const [selected, setSelected] = useState(paymentTypes[0])

  return (
    <>
      <Form
        id='payment-method-type'
        action={`/flows/${flow.id}/payment-method/type`}
        method='post'
        className='hidden'
      />
      <RadioGroup
        className='col-span-full sm:col-span-6 sm:col-start-2 lg:col-start-4'
        id='payment-type'
        label='Payment method type'
        value={selected}
        onChange={setSelected}
        options={paymentTypes}
      />
      <input
        form='payment-method-type'
        value={String(selected?.id)}
        name='payment-type'
        type='hidden'
      />
      <div className='col-span-full mt-4 flex items-center justify-between space-x-3 rounded-xl bg-container p-3 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Icon>tips_and_updates</Icon>
        <span className='font-sans text-sm font-normal'>
          We currently only support Bank accounts, more coming soon!
        </span>
      </div>

      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Button form='payment-method-type' type='submit'>
          Continue
        </Button>
      </div>
    </>
  )
}

export const action: ActionFunction = async ({ request }) => {
  const form = await request.formData()
  const paymentType = form.get('payment-type')
  await stepFlow(request, { paymentType })
}
