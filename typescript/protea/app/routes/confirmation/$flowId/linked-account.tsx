import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Button } from '~/components'
import { exitFlow, getCurrentFlow } from '~/lib/flows.server'
import { grpcClient, isGrpcError, StatusError } from '~/lib/proto.server'

export async function loader({ request, params }: LoaderArgs) {
  const flow = await getCurrentFlow(request, params)
  const cookie = request.headers.get('cookie') || ''
  let rpc = await grpcClient.getBankAccountDetails(
    {
      fundingsourceId: flow?.data.fundingsourceId,
    }, 
    {
      meta: {
        cookies: cookie
      }
    }
  ).then((v) => v)
    .catch(StatusError)
  if (isGrpcError(rpc)) {
    throw rpc
  }

  return json({
    flow,
    accountNumber: rpc.response.mask,
    institution: rpc.response.institution,
    type: rpc.response.type,
  })
}

export default function Page() {
  const { flow, accountNumber, institution, type } = useLoaderData<typeof loader>()
  const { nickname } = flow?.data
  return (
    <>
      <Form
        id='linked-account-confirmation'
        action={`/confirmation/${flow.id}/linked-account`}
        method='post'
        className='hidden'
      />
      <div className='col-span-full flex flex-col pb-8 pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='font-display text-4xl font-medium'>
          Payment method added
        </span>
      </div>
      <div className='col-span-full flex flex-col pb-4 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='text-sm font-medium'>Account Type</span>
        <span>{type}</span>
      </div>
      <div className='col-span-full flex flex-col pb-4 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='text-sm font-medium'>Institution</span>
        <span>{institution}</span>
      </div>
      <div className='col-span-full flex flex-col pb-4 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='text-sm font-medium'>Account number</span>
        <span>{accountNumber}</span>
      </div>
      <div className='col-span-full flex flex-col pb-4 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <span className='text-sm font-medium'>Nickname</span>
        <span>{nickname}</span>
      </div>

      <div className='col-span-full flex justify-end pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <Button form='linked-account-confirmation' type='submit'>
          Continue
        </Button>
      </div>
    </>
  )
}

export async function action({ request }: ActionArgs) {
  const headers = await exitFlow(request)
  return redirect(route('/settings'), { headers })
}
