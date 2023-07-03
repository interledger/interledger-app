import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  AnimatedSchedule,
  Card,
  CardContent,
  Chip,
  ChipColor,
  Icon,
  Layouts
} from '~/components'
import { getPusherArgs } from '~/lib/pusher.server'
import { usePusher } from '~/lib/usePusher'
import { getTransaction } from '~/lib/wallet.server'

export async function loader({ request, params }: LoaderArgs) {
  const transaction = await getTransaction(
    request,
    params.type as string,
    params.transactionId as string
  )
  console.log('transaction', transaction)

  const pusherArgs = await getPusherArgs(request)

  return json({
    transaction,
    pusherArgs
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      back: route('/transactions'),
      title: 'Sent payment',
      actions: [
        {
          type: 'chip',
          content: (match) => {
            switch (match.data.transaction.status) {
              case 'Completed':
                return <Chip color={ChipColor.green}>Complete</Chip>
              case 'Pending':
                return <Chip color={ChipColor.orange}>Pending</Chip>
              case 'Failed':
                return <Chip color={ChipColor.red}>Failed</Chip>
              default:
                return null
            }
          }
        }
      ]
    },
    isNested: true
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Transaction | Outgoing'
  }
}

export default function Page() {
  const { transaction, pusherArgs } = useLoaderData<typeof loader>()

  usePusher(pusherArgs, ['transaction', 'kyc'])
  return (
    <>
      {transaction.type.includes('outgoing') && <Outgoing />}
      {transaction.type.includes('incoming') && <Incoming />}
    </>
  )
}

function Outgoing() {
  const { transaction } = useLoaderData<typeof loader>()
  return (
    <>
      <Card>
        <CardContent className='my-6 flex flex-col items-center justify-center space-y-4'>
          {transaction.icon == 'schedule' && (
            <div className='mt-0.5'>
              <AnimatedSchedule />
            </div>
          )}
          {transaction.icon != 'schedule' && (
            <Icon className='mt-0.5 text-medium'>{transaction.icon}</Icon>
          )}
          <h3 className='text-3xl font-medium text-error'>
            {transaction.total}
          </h3>
          <p className='text-weak'>Payment to {transaction.title}</p>
          <p className='text-weak'>
            Sent on {transaction.date} at {transaction.time}
          </p>
        </CardContent>
      </Card>
      <Card>
        <CardContent className='flex flex-col space-y-1'>
          <div className='flex w-full justify-between'>
            <span className='text-weak'>They receive</span>
            <span className='text-medium'>{transaction.subTotal}</span>
          </div>
          <div className='flex w-full justify-between'>
            <span className='text-weak'>Total fees</span>
            <span className='text-medium'>{transaction.fees}</span>
          </div>
          <div className='flex w-full justify-between'>
            <span className='text-weak'>You pay</span>
            <span className='text-medium'>{transaction.total}</span>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent className='flex flex-col space-y-4'>
          {transaction.accountTitle && (
            <div className='flex w-full flex-col space-y-1'>
              <span className='text-weak'>Source</span>
              <span className='text-medium'>{transaction.accountTitle}</span>
            </div>
          )}
          <div className='flex w-full  flex-col space-y-1'>
            <span className='text-weak'>Transaction ID</span>
            <span className='text-medium'>{transaction.id}</span>
          </div>
          {transaction.reference && (
            <div className='flex w-full  flex-col space-y-1'>
              <span className='text-weak'>Reference</span>
              <span className='text-medium'>{transaction.reference}</span>
            </div>
          )}
        </CardContent>
      </Card>
    </>
  )
}

function Incoming() {
  const { transaction } = useLoaderData<typeof loader>()
  return (
    <>
      <Card>
        <CardContent className='my-6 flex flex-col items-center justify-center space-y-4'>
          {transaction.icon == 'schedule' && (
            <div className='mt-0.5'>
              <AnimatedSchedule />
            </div>
          )}
          {transaction.icon != 'schedule' && (
            <Icon className='mt-0.5 text-medium'>{transaction.icon}</Icon>
          )}
          <h3 className='text-3xl font-medium text-strong'>
            {transaction.total}
          </h3>
          <p className='text-weak'>Payment from {transaction.title}</p>
          <p className='text-weak'>
            Received on {transaction.date} at {transaction.time}
          </p>
        </CardContent>
      </Card>
      <Card>
        <CardContent className='flex flex-col space-y-1'>
          <div className='flex w-full justify-between'>
            <span className='text-weak'>They sent</span>
            <span className='text-medium'>{transaction.subTotal}</span>
          </div>
          <div className='flex w-full justify-between'>
            <span className='text-weak'>Total fees</span>
            <span className='text-medium'>{transaction.fees}</span>
          </div>
          <div className='flex w-full justify-between'>
            <span className='text-weak'>You received</span>
            <span className='text-medium'>{transaction.total}</span>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent className='flex flex-col space-y-4'>
          {transaction.accountTitle && (
            <div className='flex w-full flex-col space-y-1'>
              <span className='text-weak'>Destination</span>
              <span className='text-medium'>{transaction.accountTitle}</span>
            </div>
          )}
          <div className='flex w-full  flex-col space-y-1'>
            <span className='text-weak'>Transaction ID</span>
            <span className='text-medium'>{transaction.id}</span>
          </div>
          {transaction.reference && (
            <div className='flex w-full  flex-col space-y-1'>
              <span className='text-weak'>Reference</span>
              <span className='text-medium'>{transaction.reference}</span>
            </div>
          )}
        </CardContent>
      </Card>
    </>
  )
}
