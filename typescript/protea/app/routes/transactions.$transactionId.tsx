import type { LoaderArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  AnimatedSchedule,
  Card,
  CardButton,
  CardContent,
  CardIcon,
  Chip,
  ChipColor,
  FynbosIcon,
  Icon,
  Layouts,
  TwitterIcon
} from '~/components'
import { Label } from '~/components/Label'
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
        <CardContent>
          <div className='flex items-center justify-between'>
            <h2 className='text-4xl font-medium text-error'>
              - {transaction.total}
            </h2>
            {transaction.icon && (
              <CardIcon>
                {transaction.icon === 'wallet' && <FynbosIcon />}
                {transaction.icon === 'twitter' && <TwitterIcon />}
              </CardIcon>
            )}
          </div>
          <Label className='-mb-5 mt-4'>Payment to</Label>
        </CardContent>
        <CardButton>
          <div className='flex w-full items-center justify-between text-medium'>
            <span>{transaction.title}</span>
            <Icon>navigate_next</Icon>
          </div>
        </CardButton>
      </Card>
      <Card>
        <CardContent>
          <div className='flex w-full justify-between'>
            <span className='text-weak'>Total fees</span>
            <span className='text-medium'>
              Free <sup>*</sup>
            </span>
          </div>
          <div className='mt-2 flex w-full justify-between'>
            <span className='text-weak'>They receive</span>
            <span className='text-medium'>{transaction.total}</span>
          </div>
          <div className='mt-4 flex w-full space-x-2'>
            <span className='text-xs text-medium'>*</span>
            <span className='text-xs text-medium'>
              For a limited time, Fynbos will absorb the fees associated with
              making a payment.
            </span>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent>
          <div className='flex w-full flex-col justify-between space-y-1'>
            <span className='text-weak'>Source</span>
            <span className='text-medium'>{transaction.accountTitle}</span>
          </div>
          {transaction.reference && (
            <div className='mt-4 flex w-full flex-col space-y-1'>
              <span className='text-weak'>Reference</span>
              <span className='text-medium'>{transaction.reference}</span>
            </div>
          )}
          <div className='mt-4 flex w-full flex-col justify-between space-y-1'>
            <span className='text-weak'>Date</span>
            <span className='text-medium'>
              {transaction.date} at {transaction.time}
            </span>
          </div>
          <div className='mt-4 flex w-full flex-col justify-between space-y-1'>
            <span className='text-weak'>Transaction ID</span>
            <span className='text-medium'>{transaction.id}</span>
          </div>
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
        <CardContent>
          <div className='flex items-center justify-between'>
            <h2 className='text-4xl font-medium text-strong'>
              {transaction.total}
            </h2>
            {transaction.icon && (
              <CardIcon>
                {transaction.icon === 'schedule' && <AnimatedSchedule />}
                {transaction.icon === 'wallet' && <FynbosIcon />}
                {transaction.icon === 'twitter' && <TwitterIcon />}
              </CardIcon>
            )}
          </div>
          <Label className='-mb-5 mt-4'>Payment from</Label>
        </CardContent>
        <CardButton>
          <div className='flex w-full items-center justify-between text-medium'>
            <span>{transaction.title}</span>
            <Icon>navigate_next</Icon>
          </div>
        </CardButton>
      </Card>
      <Card>
        <CardContent>
          <div className='flex w-full flex-col justify-between space-y-1'>
            <span className='text-weak'>Paid to</span>
            <span className='text-medium'>{transaction.accountTitle}</span>
          </div>
          {transaction.reference && (
            <div className='mt-4 flex w-full flex-col space-y-1'>
              <span className='text-weak'>Their reference</span>
              <span className='text-medium'>{transaction.reference}</span>
            </div>
          )}
          <div className='mt-4 flex w-full flex-col justify-between space-y-1'>
            <span className='text-weak'>Date</span>
            <span className='text-medium'>
              {transaction.date} at {transaction.time}
            </span>
          </div>
          <div className='mt-4 flex w-full flex-col justify-between space-y-1'>
            <span className='text-weak'>Transaction ID</span>
            <span className='text-medium'>{transaction.id}</span>
          </div>
        </CardContent>
      </Card>
    </>
  )
}
