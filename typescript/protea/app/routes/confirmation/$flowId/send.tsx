import type { ActionFunction, LoaderFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import { Button } from '~/components'
import { exitFlow, getCurrentFlow } from '~/lib/flows.server'

export const loader: LoaderFunction = async ({ request, params }) => {
  const flow = await getCurrentFlow(request, params)
  return json({
    flow
  })
}

export default function Page() {
  const { flow } = useLoaderData()
  const { to, displayAmount, displayFee, displayTotal } = flow?.data
  return (
    <>
      <Form
        id='send-confirmation'
        action={`/confirmation/${flow.id}/send`}
        method='post'
        className='hidden'
      />
      <div className='col-span-full flex flex-col pb-8 pt-4 text-strong sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-display text-4xl font-medium'>
          Send confirmed
        </span>
      </div>
      <div className='col-span-full flex justify-between pb-4 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-sans text-sm font-medium'>Sent to</span>
        <span className='font-sans text-base font-normal'>{to}</span>
      </div>
      <div className='text medium col-span-full flex justify-between sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-display text-sm font-medium'>Amount</span>
        <span className='font-sans text-sm font-normal'>
          {displayAmount || '$ 0.00'}
        </span>
      </div>
      <div className='text medium col-span-full flex justify-between sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-display text-sm font-medium'>Fees</span>
        <span className='font-sans text-sm font-normal'>
          {displayFee || '$ 0.00'}
        </span>
      </div>
      <div className='col-span-full flex items-end justify-between py-3 text-strong sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-display text-2xl font-medium'>Total</span>
        <span className='font-sans text-4xl font-medium'>
          {displayTotal || '$ 0.00'}
        </span>
      </div>

      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Button form='send-confirmation' type='submit'>
          Continue
        </Button>
      </div>
    </>
  )
}

export const action: ActionFunction = async ({ request }) => {
  return exitFlow(request)
}
