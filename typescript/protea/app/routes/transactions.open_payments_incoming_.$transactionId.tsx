import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Card, CardContent, Chip, ChipColor, Layouts } from '~/components'
import { getUserSession } from '~/lib/kratos.server'
import {
  StatusError,
  httpMapping,
  isGrpcError,
  openPaymentsClient
} from '~/lib/proto.server'
import { getPusherArgs } from '~/lib/pusher.server'
import { usePusher } from '~/lib/usePusher'
import { getTransaction } from '~/lib/wallet.server'

export async function loader({ request, params }: LoaderArgs) {
  const session = await getUserSession(request)
  const transaction = await getTransaction(
    request,
    params.type as string,
    params.transactionId as string
  )

  const pusherArgs = await getPusherArgs(request)

  const cookie = String(request.headers.get('cookie'))
  const incomingPayment = await openPaymentsClient
    .lookupIncomingPayment(
      { id: transaction.foreignId },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(incomingPayment)) {
    throw json({}, httpMapping(incomingPayment.code))
  }

  // For now handle just incoming & outgoing payments
  let beneficiaryName = ''
  const response = await openPaymentsClient
    .getPaymentPointer({ url: incomingPayment.response.fromPaymentPointer })
    .then((v) => v)
    .catch(StatusError)

  // Silently fail if it can't be found for now
  if (!isGrpcError(response)) {
    beneficiaryName = response.response.legalName
  }

  return json({
    // Always go to /transactions with back button even if we've just done a payment flow
    backTo: route('/transactions'),
    transaction,
    pusherArgs,
    beneficiaryName,
    paymentPointer: incomingPayment.response.fromPaymentPointer,
    note: incomingPayment.response.externalRef,
    traits: session.identity.traits
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/transactions'),
      title: 'Received payment'
      // TODO: chip as action
    }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Transaction | Incoming'
  }
}

export default function Page() {
  const { transaction, pusherArgs, beneficiaryName, paymentPointer } =
    useLoaderData<typeof loader>()

  usePusher(pusherArgs, ['transaction', 'kyc'])
  return (
    <>
      <Card>
        <CardContent>
          <div className='flex w-full flex-col space-y-1'>
            <span className='text-sm'>From</span>
            <span className='text-sm text-strong'>{paymentPointer}</span>
          </div>
          {beneficiaryName != '' && (
            <div className='mt-4 flex w-full flex-col space-y-1'>
              <span className='text-sm'>Sender name</span>
              <span className='text-sm text-strong'>{beneficiaryName}</span>
            </div>
          )}
          <div className='mt-4 flex w-full flex-col space-y-1'>
            <span className='text-sm'>They sent</span>
            <span className='text-sm font-medium text-strong'>
              {transaction.subTotal || '$ 0.00'}
            </span>
          </div>
          <div className='mt-4 flex w-full flex-col space-y-1'>
            <span className='text-sm'>Total fees</span>
            <span className='text-sm font-medium text-strong'>
              {transaction.fees || '$ 0.00'}
            </span>
          </div>
          <div className='mt-4 flex w-full flex-col space-y-1'>
            <span className='text-sm'>You received</span>
            <span className='text-2xl text-sm font-medium text-strong'>
              {transaction.total || '$ 0.00'}
            </span>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent>
          <div className='mt-4 flex w-full flex-col space-y-1'>
            <span className='text-sm text-medium'>Payment date</span>
            <span className='text-sm text-strong'>
              {transaction.date || 'Pending'}
            </span>
          </div>
          <div className='mt-4 items-center'>
            <span className='text-sm text-medium'>Status</span>
            {transaction.status == 'Completed' && (
              <Chip color={ChipColor.green}>Complete</Chip>
            )}
            {transaction.status == 'Pending' && (
              <Chip color={ChipColor.yellow}>Pending</Chip>
            )}
            {transaction.status == 'Failed' && (
              <Chip color={ChipColor.orange}>Failed</Chip>
            )}
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent>
          <div className='mt-4 flex w-full flex-col space-y-1'>
            <span className='text-sm text-medium'>Transaction ID</span>
            <span className='text-sm text-strong'>{transaction.id}</span>
          </div>
        </CardContent>
      </Card>
    </>
  )
}
