import type { ActionFunction, LoaderFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import { Button, Icon } from '~/components'
import { exitFlow, getCurrentFlow } from '~/lib/flows.server'

export const loader: LoaderFunction = async ({ request, params }) => {
  const flow = await getCurrentFlow(request, params)
  return json({
    flow
  })
}

export default function Page() {
  const { flow } = useLoaderData()
  const { paymentMethodMask, displayAmount, displayFee, displayTotal } =
    flow?.data
  return (
    <>
      <Form
        id='deposit-confirmation'
        action={`/confirmation/${flow.id}/deposit`}
        method='post'
        className='hidden'
      />
      <div className='col-span-full flex flex-col pb-8 pt-4 text-strong sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-display text-4xl font-medium'>
          Deposit confirmed
        </span>
      </div>
      <div className='col-span-full flex justify-between pb-4 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-sans text-sm font-medium'>Payment method</span>
        <span className='font-sans text-base font-normal'>
          {paymentMethodMask}
        </span>
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
      <div className='col-span-full flex items-center justify-between space-x-3 rounded-xl bg-container p-3 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Icon>tips_and_updates</Icon>
        <span className='font-sans text-sm font-normal'>
          Your deposit may take some time to appear in your account.
        </span>
      </div>

      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Button form='deposit-confirmation' type='submit'>
          Continue
        </Button>
      </div>
    </>
  )
}

export const action: ActionFunction = async ({ request }) => {
  return exitFlow(request)
}
