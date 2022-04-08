import React from 'react'
import { Form, LoaderFunction, json, useLoaderData } from 'remix'
import type { ActionFunction } from 'remix'
import { Button } from '~/components'
import { completeFlow, getCurrentFlow } from '~/lib/flows.server'
import { apolloClient } from '~/lib/apollo.server'
import {
  InitiateWithdrawalMutation,
  InitiateWithdrawalDocument,
  InitiateWithdrawalMutationVariables
} from '~/generated/types'

export const loader: LoaderFunction = async ({ request, params }) => {
  const flow = await getCurrentFlow(request, params)
  return json({
    flow
  })
}

export default function WithdrawReviewPage() {
  const { flow } = useLoaderData()
  const { paymentMethodMask, displayAmount, displayFee, displayTotal } =
    flow?.data
  return (
    <>
      <Form
        id='withdraw-review'
        action={`/flows/${flow.id}/withdraw/review`}
        method='post'
        className='hidden'
      />

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

      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Button form='withdraw-review' type='submit'>
          Confirm
        </Button>
      </div>
    </>
  )
}

export const action: ActionFunction = async ({ request, params }) => {
  const flow = await getCurrentFlow(request, params)
  const { paymentMethodId, amount } = flow?.data
  const cookie = request.headers.get('cookie')
  const initiateWithdrawalMutationVariables = {
    input: {
      fundingSourceID: paymentMethodId,
      amount: amount.toFixed(2).replace('.', '')
    }
  }
  const res = await apolloClient.mutate<
    InitiateWithdrawalMutation,
    InitiateWithdrawalMutationVariables
  >({
    mutation: InitiateWithdrawalDocument,
    variables: initiateWithdrawalMutationVariables,
    context: {
      headers: {
        cookie: cookie
      }
    }
  })
  if (res.data?.initiateWithdrawal.success) return completeFlow(request)
}
