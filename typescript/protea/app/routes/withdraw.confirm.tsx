import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, useLoaderData } from '@remix-run/react'
import { Button, Layouts } from '~/components'
import { exitFlow, flowType, requireFlow } from '~/lib/flows.server'
import { route } from 'routes-gen'
import {
  grpcClient,
  httpMapping,
  isGrpcError,
  StatusError
} from '~/lib/proto.server'
import { getClientIP } from '~/lib/ip.server'
import { getLinkedAccounts } from '~/lib/wallet.server'
import { randomUUID } from 'crypto'
import { flashSnackbar } from '~/lib/snackbar.server'

export async function loader({ request }: LoaderArgs) {
  const flow = await requireFlow(request, flowType.Withdraw)
  const { linkedAccounts } = await getLinkedAccounts(request)
  return json({
    flow,
    linkedAccount: linkedAccounts.find(
      (account) => account.id == flow.data.toLinkedAccountId
    )
  })
}

export const handle = {
  layout: Layouts.FocusLayout
}

export default function Page() {
  const { flow, linkedAccount } = useLoaderData<typeof loader>()

  return (
    <>
      <div className='flex w-full flex-col rounded-2xl bg-page p-4 pb-8'>
        <h1 className='mb-6 font-display text-2xl font-medium'>
          Confirm withdrawal
        </h1>
        <span>Please check the details and confirm your withdrawal.</span>

        <div className='mt-6 flex w-full flex-col space-y-1'>
          <span className='text-sm'>Withdraw to:</span>
          <span className='text-sm'>{linkedAccount?.name}</span>
        </div>
        <div className='mt-6 flex w-full justify-between'>
          <span className='text-sm'>You withdraw </span>
          <span className='text-sm font-medium text-strong'>
            {flow?.data.displayReceiveAmount || '$ 0.00'}
          </span>
        </div>
        <div className='mt-2 flex w-full justify-between'>
          <span className='text-sm'>Total fees</span>
          <span className='text-sm font-medium text-strong'>
            Free<sup>*</sup>
          </span>
        </div>
        <div className='mt-2 flex w-full justify-between'>
          <span className='text-sm'>You receive</span>
          <span className='text-sm text-2xl font-medium text-strong'>
            {flow?.data.displayReceiveAmount || '$ 0.00'}
          </span>
        </div>

        <Form
          id='withdraw-confirm'
          action='/withdraw/confirm'
          method='post'
          className='hidden'
        />
        <div className='mt-6'>
          <Button form='withdraw-confirm' type='submit'>
            Confirm
          </Button>
        </div>
      </div>
      <div className='mt-6 flex w-full space-x-2'>
        <span className='text-xs text-medium'>*</span>
        <span className='text-xs text-medium'>
          For a limited time, Fynbos will absorb the fees associated with making
          a withdrawal.
        </span>
      </div>
    </>
  )
}

export async function action({ request }: ActionArgs) {
  const flow = await requireFlow(request, flowType.Withdraw)

  const clientIpAddress = getClientIP(request)

  const response = await grpcClient
    .startWithdrawFromMachnetWallet(
      {
        idempotencyKey: flow.data.idempotencyKey || '',
        toLinkedAccountId: flow.data.toLinkedAccountId,
        amount: flow.data.receiveAmount,
        ipAddress: clientIpAddress
      },
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)
  if (isGrpcError(response)) {
    throw json({}, httpMapping(response.code))
  }
  await exitFlow(request, flowType.Withdraw)
  return redirect(route('/'), {
    headers: {
      'Set-Cookie': await flashSnackbar(request, {
        message: 'Withdrawal created successfully.',
        icon: 'close'
      })
    }
  })
}
