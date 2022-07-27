import { useEffect, useState } from 'react'
import type { ActionFunction, LoaderFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useActionData, useLoaderData } from '@remix-run/react'
import { Button, Select, TextField } from '~/components'
import { getCurrentFlow, stepFlow } from '~/lib/flows.server'

type ActionData = {
  formError?: string
  fieldErrors?: {
    accountNumber: string | undefined
    institution: string | undefined
    name: string | undefined
    routingNumber: string | undefined
    type: string | undefined
  }
  fields?: {
    accountNumber: string
    institution: string
    name: string
    routingNumber: string
    type: string
  }
}

export const loader: LoaderFunction = async ({ request, params }) => {
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
  const actionData = useActionData<ActionData>()
  const { accountTypes, flow } = useLoaderData()

  const [selected, setSelected] = useState(accountTypes[0])

  useEffect(() => {
    if (accountTypes && flow?.data.type) {
      const sel = accountTypes.find((val: any) => val.name == flow?.data.type)
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
        // defaultValue={actionData?.fields?.type}
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.fieldErrors?.type) || undefined}
        aria-describedby={
          actionData?.fieldErrors?.type ? 'type-error' : undefined
        }
        errorMessage={actionData?.fieldErrors?.type}
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
        defaultValue={actionData?.fields?.institution || flow?.data.institution}
        type='text'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={
          Boolean(actionData?.fieldErrors?.institution) || undefined
        }
        aria-describedby={
          actionData?.fieldErrors?.institution ? 'institution-error' : undefined
        }
        required
        errorMessage={actionData?.fieldErrors?.institution}
      />

      <TextField
        id='accountNumber'
        form='payment-method-details'
        label='Account number'
        name='accountNumber'
        defaultValue={
          actionData?.fields?.accountNumber || flow?.data.accountNumber
        }
        type='text'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={
          Boolean(actionData?.fieldErrors?.accountNumber) || undefined
        }
        aria-describedby={
          actionData?.fieldErrors?.accountNumber
            ? 'accountNumber-error'
            : undefined
        }
        required
        errorMessage={actionData?.fieldErrors?.accountNumber}
      />

      <TextField
        id='routingNumber'
        form='payment-method-details'
        label='Routing number'
        name='routingNumber'
        defaultValue={
          actionData?.fields?.routingNumber || flow?.data.routingNumber
        }
        type='text'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={
          Boolean(actionData?.fieldErrors?.routingNumber) || undefined
        }
        aria-describedby={
          actionData?.fieldErrors?.routingNumber
            ? 'routingNumber-error'
            : undefined
        }
        required
        errorMessage={actionData?.fieldErrors?.routingNumber}
      />

      <TextField
        id='name'
        form='payment-method-details'
        label='Nickname'
        name='name'
        defaultValue={actionData?.fields?.name || flow?.data.name}
        type='text'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        aria-invalid={Boolean(actionData?.fieldErrors?.name) || undefined}
        aria-describedby={
          actionData?.fieldErrors?.name ? 'name-error' : undefined
        }
        required
        errorMessage={actionData?.fieldErrors?.name}
      />

      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Button form='payment-method-details' type='submit'>
          Continue
        </Button>
      </div>
    </>
  )
}

const badRequest = (data: ActionData) => json(data, { status: 400 })

export const action: ActionFunction = async ({ request }) => {
  const form = await request.formData()
  const accountNumber = form.get('accountNumber')
  const institution = form.get('institution')
  const name = form.get('name')
  const routingNumber = form.get('routingNumber')
  const type = form.get('type')

  // TODO: proper validation of input based on providers requirements
  if (
    typeof accountNumber !== 'string' ||
    typeof institution !== 'string' ||
    typeof name !== 'string' ||
    typeof routingNumber !== 'string' ||
    typeof type !== 'string'
  ) {
    return badRequest({
      formError: `Form not submitted correctly.`
    })
  }

  await stepFlow(request, {
    accountNumber: accountNumber,
    institution: institution,
    name: name,
    routingNumber: routingNumber,
    type: type
  })
}
