import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { redirect } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import { Button } from '~/components'
import { updateFlow, getCurrentFlow } from '~/lib/flows.server'
import { apolloClient } from '~/lib/apollo.server'
import type {
  InitiateWithdrawalMutation,
  InitiateWithdrawalMutationVariables
} from '~/generated/types'
import { InitiateWithdrawalDocument } from '~/generated/types'
import { route } from 'routes-gen'

export async function loader({ request, params }: LoaderArgs) {
  const flow = await getCurrentFlow(request, params)
  return json({
    flow
  })
}

export default function Page() {
  const { flow } = useLoaderData<typeof loader>()
  const { linkedAccountMask, displayAmount, displayFee, displayTotal } =
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
        <span className='text-sm font-medium'>Payment method</span>
        <span>{linkedAccountMask}</span>
      </div>
      <div className='text medium col-span-full flex justify-between sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-display text-sm font-medium'>Amount</span>
        <span className='text-sm'>{displayAmount || '$ 0.00'}</span>
      </div>
      <div className='text medium col-span-full flex justify-between sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-display text-sm font-medium'>Fees</span>
        <span className='text-sm'>{displayFee || '$ 0.00'}</span>
      </div>
      <div className='col-span-full flex items-end justify-between py-3 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-display text-2xl font-medium'>Total</span>
        <span className='text-4xl font-medium'>{displayTotal || '$ 0.00'}</span>
      </div>

      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Button form='withdraw-review' type='submit'>
          Confirm
        </Button>
      </div>
    </>
  )
}

export async function action({ request, params }: ActionArgs) {
  const flow = await getCurrentFlow(request, params)
  const { linkedAccountId, total } = flow?.data
  const cookie = request.headers.get('cookie')
  const initiateWithdrawalMutationVariables = {
    input: {
      fundingSourceID: linkedAccountId,
      amount: total.toFixed(2).replace('.', '')
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

  const headers = await updateFlow(request, null, true)
  if (res.data?.initiateWithdrawal.success)
    return redirect(
      route('/confirmation/:flowId/send', {
        flowId: flow?.id as string
      }),
      { headers }
    )
}
