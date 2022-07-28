import { useEffect, useState } from 'react'
import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { redirect } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import { Button, Select, TextField } from '~/components'
import { getCurrentFlow, updateFlow } from '~/lib/flows.server'
import { route } from 'routes-gen'

export async function loader({ request, params }: LoaderArgs) {
  const flow = await getCurrentFlow(request, params)

  const accountTypes = [
    { id: '1', name: 'Checking' },
    { id: '2', name: 'Savings' }
  ]

  return json({
    accountTypes,
    flow
  })
}

export default function Page() {
  const { accountTypes, flow } = useLoaderData<typeof loader>()

  const [selected, setSelected] = useState(accountTypes[0])

  useEffect(() => {
    if (accountTypes && flow?.data.type) {
      const sel = accountTypes.find(
        (val: any) => val.name == flow?.data.type
      ) as { id: string; name: string }
      setSelected(sel)
    }
  }, [flow?.data.type, accountTypes])

  return (
    <>
      <Form
        id='payment-method-details'
        action={`/flows/${flow.id}/payment-method/details`}
        method='post'
        className='hidden'
      />
      <Select
        id='type'
        value={selected}
        onChange={setSelected}
        options={accountTypes}
        label='Account type'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
      />
      <input
        form='payment-method-details'
        value={String(selected.name)}
        name='type'
        type='hidden'
      />

      <TextField
        id='institution'
        form='payment-method-details'
        label='Institution'
        name='institution'
        defaultValue={flow?.data.institution}
        type='text'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        required
      />

      <TextField
        id='accountNumber'
        form='payment-method-details'
        label='Account number'
        name='accountNumber'
        defaultValue={flow?.data.accountNumber}
        type='text'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        required
      />

      <TextField
        id='routingNumber'
        form='payment-method-details'
        label='Routing number'
        name='routingNumber'
        defaultValue={flow?.data.routingNumber}
        type='text'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        required
      />

      <TextField
        id='name'
        form='payment-method-details'
        label='Nickname'
        name='name'
        defaultValue={flow?.data.name}
        type='text'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        required
      />

      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Button form='payment-method-details' type='submit'>
          Continue
        </Button>
      </div>
    </>
  )
}

export async function action({ request, params }: ActionArgs) {
  const form = await request.formData()
  const accountNumber = form.get('accountNumber') as string
  const institution = form.get('institution') as string
  const name = form.get('name') as string
  const routingNumber = form.get('routingNumber') as string
  const type = form.get('type') as string

  const headers = await updateFlow(request, {
    accountNumber: accountNumber,
    institution: institution,
    name: name,
    routingNumber: routingNumber,
    type: type
  })

  const flow = await getCurrentFlow(request, params)
  return redirect(
    route('/flows/:flowId/payment-method/review', {
      flowId: flow?.id as string
    }),
    { headers }
  )
}
