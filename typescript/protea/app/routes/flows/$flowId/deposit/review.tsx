import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Button } from '~/components'
import { getCurrentFlow, updateFlow } from '~/lib/flows.server'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'

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
        id='deposit-review'
        action={`/flows/${flow.id}/deposit/review`}
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
        <Button form='deposit-review' type='submit'>
          Confirm
        </Button>
      </div>
    </>
  )
}

export async function action({ request, params }: ActionArgs) {
  const flow = await getCurrentFlow(request, params)
  const { linkedAccountId, amount } = flow?.data
  const response = await grpcClient
    .initiateDeposit(
      {
        amount,
        fundingsourceId: linkedAccountId
      },
      {
        meta: { cookies: String(request.headers.get('cookie')) }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }

  const headers = await updateFlow(request, null, true)
  return redirect(
    route('/confirmation/:flowId/deposit', { flowId: flow?.id as string }),
    { headers }
  )
}
