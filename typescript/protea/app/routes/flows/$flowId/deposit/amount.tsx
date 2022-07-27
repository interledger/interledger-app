import type { ChangeEventHandler } from 'react'
import { useState } from 'react'
import type { ActionFunction, LoaderFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useFetcher, useLoaderData } from '@remix-run/react'
import { Button, TextField } from '~/components'
import { getCurrentFlow, stepFlow, updateFlowData } from '~/lib/flows.server'

// type ActionData = {
//   formError?: string
//   fieldErrors?: {
//     amount: string | undefined
//   }
//   fields?: {
//     amount: string
//   }
// }

export const loader: LoaderFunction = async ({ request, params }) => {
  const flow = await getCurrentFlow(request, params)
  return json({
    flow
  })
}

export default function Page() {
  // const actionData = useActionData<ActionData>()
  const { flow } = useLoaderData()
  const fetcher = useFetcher()

  const [amount, setAmount] = useState<string>(flow?.data.amount || '')

  const changeHandler: ChangeEventHandler<HTMLInputElement> = (e) => {
    const newVal = e.target.value
    if (newVal.split('.')[1] && newVal.split('.')[1].length > 2) {
      setAmount(newVal.slice(0, -1))
      fetcher.submit({ amount: newVal.slice(0, -1) }, { method: 'post' })
      return
    }
    setAmount(newVal)
    fetcher.submit({ amount: newVal }, { method: 'post' })
  }

  return (
    <>
      <Form
        id='amount-form'
        action={`/flows/${flow.id}/deposit/amount`}
        method='post'
        className='col-span-full flex flex-col items-end space-y-2 sm:col-span-6 sm:col-start-2 lg:col-start-4'
      />
      <TextField
        id='amount'
        form='amount-form'
        label='Amount'
        name='amount'
        value={amount}
        onChange={changeHandler}
        type='number'
        min='0'
        step='0.01'
        className='col-span-full flex flex-col sm:col-span-6 sm:col-start-2 lg:col-start-4'
        // aria-invalid={Boolean(actionData?.fieldErrors?.amount) || undefined}
        // aria-describedby={
        //   actionData?.fieldErrors?.amount ? 'amount-error' : undefined
        // }
        required
        // errorMessage={actionData?.fieldErrors?.amount}
      />

      <div className='text medium col-span-full flex justify-between sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-display text-sm font-medium'>Fees</span>
        <span className='font-sans text-sm font-normal'>
          {flow?.data.displayFee || '$ 0.00'}
        </span>
      </div>
      <div className='col-span-full flex items-end justify-between py-3 text-strong sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-display text-2xl font-medium'>Total</span>
        <span className='font-sans text-4xl font-medium'>
          {flow?.data.displayTotal || '$ 0.00'}
        </span>
      </div>
      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Button form='amount-form' type='submit' name='route-to' value='next'>
          Continue
        </Button>
      </div>
    </>
  )
}

export const action: ActionFunction = async ({ request }) => {
  // TODO: fetch fee from db
  const feeStructure = {
    fixed: 1.0,
    percentage: 0.02
  }
  const form = await request.formData()
  const amount = parseFloat(String(form.get('amount')))
  const routeTo = form.get('route-to')

  let fee = 0,
    total = 0
  if (!isNaN(amount)) {
    fee = amount * feeStructure.percentage + feeStructure.fixed
    total = amount + fee
    if (total < 0) total = 0
  }
  const data = {
    amount: amount,
    displayAmount: formatMoney(amount),
    fee: fee,
    displayFee: formatMoney(fee),
    total: total,
    displayTotal: formatMoney(total)
  }
  if (routeTo == 'next') {
    // TODO step to next flow step
    await stepFlow(request, data)
  } else {
    return updateFlowData(request, data)
  }
}

const formatMoney = (value: number): string => {
  return `$ ${value.toFixed(2)}`
}
