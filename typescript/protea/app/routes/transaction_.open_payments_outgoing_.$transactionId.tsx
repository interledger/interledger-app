import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import { Card, CardRow, Chip, ChipColor, Layouts } from '~/components'
import { getUserSession } from '~/lib/kratos.server'
import {
  StatusError,
  httpMapping,
  isGrpcError,
  openPaymentsClient
} from '~/lib/proto.server'
import { getPusherArgs } from '~/lib/pusher.server'
import { usePusher } from '~/lib/usePusher'
import { getLinkedAccount, getTransaction } from '~/lib/wallet.server'

export async function loader({ request, params }: LoaderArgs) {
  const session = await getUserSession(request)
  const transaction = await getTransaction(
    request,
    params.type as string,
    params.transactionId as string
  )

  const pusherArgs = await getPusherArgs(request)

  const cookie = String(request.headers.get('cookie'))
  const outgoingPayment = await openPaymentsClient
    .lookupOutgoingPayment(
      { id: transaction.foreignId },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)

  if (isGrpcError(outgoingPayment)) {
    throw json({}, httpMapping(outgoingPayment.code))
  }

  // For now handle just incoming & outgoing payments
  let beneficiaryName = '',
    toPaymentPointer = ''
  const response = await openPaymentsClient
    .getPaymentPointer({ url: outgoingPayment.response.toPaymentPointer })
    .then((v) => v)
    .catch(StatusError)

  // Silently fail if it can't be found for now
  if (!isGrpcError(response)) {
    beneficiaryName = response.response.legalName
    toPaymentPointer = response.response.formatted
  }

  const hasTopUp =
    transaction.transfers.filter((trx) => trx.type == 'credit_wallet').length >
    0

  let linkedAccountName: string, transfers

  if (hasTopUp) {
    const debitTransfer = transaction.transfers.find(
      (trx) => trx.type == 'debit_card'
    )
    const debitLinkedAccount = await getLinkedAccount(
      request,
      debitTransfer?.linkedAccountId || ''
    )
    linkedAccountName = debitLinkedAccount.name
    transfers = transaction.transfers
      .map((transfer) => {
        switch (transfer.type) {
          case 'debit_wallet':
            return {
              title: `Payment to ${toPaymentPointer}`,
              beneficiary: beneficiaryName,
              to: toPaymentPointer,
              from: 'Your Fynbos wallet',
              amount: transfer.amount,
              status: transfer.status
            }
          case 'credit_wallet':
            return {
              title: 'Top up',
              beneficiary: `${session.identity.traits.firstName} ${session.identity.traits.lastName}`,
              to: 'Your Fynbos wallet',
              from: linkedAccountName,
              amount: transfer.amount,
              status: transfer.status
            }
          default:
            return null
        }
      })
      .filter(Boolean)
  }

  return json({
    // Always go to /transactions with back button even if we've just done a payment flow
    backTo: route('/transactions'),
    hasTopUp,
    transfers,
    transaction,
    pusherArgs,
    beneficiaryName,
    paymentPointer: toPaymentPointer,
    note: outgoingPayment.response.description,
    traits: session.identity.traits
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Focus,
  scaffold: {
    header: {
      back: route('/transactions'),
      title: 'Sent payment'
      // TODO chip as action
    }
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Transaction | Outgoing'
  }
}

export default function Page() {
  const {
    transaction,
    transfers,
    pusherArgs,
    beneficiaryName,
    paymentPointer,
    note
  } = useLoaderData<typeof loader>()

  usePusher(pusherArgs, ['transaction', 'kyc'])
  return (
    <>
      <Card>
        <CardRow className='mt-6'>
          <span className='text-sm text-medium'>Payment from</span>
          <span className='text-sm text-strong'>
            {transfers ? transfers[0]?.from : 'Your Fynbos wallet'}
          </span>
        </CardRow>
        <CardRow className='mt-2'>
          <span className='text-sm text-medium'>To</span>
          <span className='text-sm text-strong'>{paymentPointer}</span>
        </CardRow>
        {beneficiaryName != '' && (
          <CardRow className='mt-2'>
            <span className='text-sm text-medium'>Beneficiary name</span>
            <span className='text-sm text-strong'>{beneficiaryName}</span>
          </CardRow>
        )}
        <CardRow className='mt-6'>
          <span className='text-sm text-medium'>You pay</span>
          <span className='text-sm font-medium text-strong'>
            {transaction.subTotal || '$ 0.00'}
          </span>
        </CardRow>
        <CardRow className='mt-2'>
          <span className='text-sm text-medium'>Total fees</span>
          <span className='text-sm font-medium text-strong'>
            {transaction.fees || '$ 0.00'}
          </span>
        </CardRow>
        <CardRow className='mt-2'>
          <span className='text-sm text-medium'>They receive</span>
          <span className='text-2xl text-sm font-medium text-strong'>
            {transaction.total || '$ 0.00'}
          </span>
        </CardRow>
        {note != '' && (
          <CardRow className='mt-6'>
            <span className='text-sm text-medium'>Note</span>
            <span className='text-sm text-strong'>{note}</span>
          </CardRow>
        )}
      </Card>
      <Card>
        <CardRow>
          <span className='text-sm text-medium'>Payment date</span>
          <span className='text-sm text-strong'>
            {transaction.date || 'Pending'}
          </span>
        </CardRow>
        <CardRow className='mt-4 items-center'>
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
        </CardRow>
      </Card>
      {transfers &&
        transfers.map((transfer) => (
          <Card key={transfer?.title} className='mt-6'>
            <h2 className='text-sm font-medium'>{transfer?.title}</h2>
            <CardRow className='mt-3'>
              <span className='text-sm text-medium'>Payment from</span>
              <span className='text-sm text-strong'>{transfer?.from}</span>
            </CardRow>
            <CardRow className='mt-2'>
              <span className='text-sm text-medium'>Payment to</span>
              <span className='text-sm text-strong'>{transfer?.to}</span>
            </CardRow>
            {transfer?.beneficiary != '' && (
              <CardRow className='mt-2'>
                <span className='text-sm text-medium'>Beneficiary name</span>
                <span className='text-sm text-strong'>
                  {transfer?.beneficiary}
                </span>
              </CardRow>
            )}
            <CardRow className='mt-2'>
              <span className='text-sm text-medium'>Payment amount</span>
              <span className='text-sm text-strong'>
                {transaction.subTotal || '$ 0.00'}
              </span>
            </CardRow>
            <CardRow className='mt-2'>
              <span className='text-sm text-medium'>Fees</span>
              <span className='text-sm text-strong'>
                {transaction.fees || '$ 0.00'}
              </span>
            </CardRow>
            <CardRow className='mt-2'>
              <span className='text-sm text-medium'>Total amount</span>
              <span className='text-sm text-strong'>
                {transaction.total || '$ 0.00'}
              </span>
            </CardRow>
            <CardRow className='mt-2'>
              <span className='text-sm text-medium'>Status</span>
              {transaction.status == 'Completed' && (
                // TODO token text colours
                <span className='text-sm text-green-800'>Complete</span>
              )}
              {transaction.status == 'Pending' && (
                <span className='text-sm text-yellow-800'>Pending</span>
              )}
              {transaction.status == 'Failed' && (
                <span className='text-sm text-orange-800'>Failed</span>
              )}
            </CardRow>
          </Card>
        ))}
      <Card>
        <CardRow>
          <span className='text-sm text-medium'>Transaction ID</span>
          <span className='text-sm text-strong'>{transaction.id}</span>
        </CardRow>
      </Card>
    </>
  )
}
