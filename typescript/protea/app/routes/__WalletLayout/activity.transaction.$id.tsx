import React from 'react'
import type { LoaderFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData, useNavigate } from '@remix-run/react'
import { Icon } from '~/components'
import { apolloClient } from '~/lib/apollo.server'
import { requireUserSession } from '~/lib/kratos.server'
import type {
  ActivityTransactionQuery,
  ActivityTransactionQueryVariables
} from '~/generated/types'
import { ActivityTransactionDocument, TransactionType } from '~/generated/types'
import { DateTime } from 'luxon'

type Activity = {
  id: string
  amount: string
  transactionType: TransactionType
  title: string
  description: string
  status: string
  date: string
}

type LoaderData = {
  transaction: Activity
}

export const loader: LoaderFunction = async ({ request, params }) => {
  await requireUserSession(request)
  const cookie = request.headers.get('cookie')
  const transaction = await apolloClient
    .query<ActivityTransactionQuery, ActivityTransactionQueryVariables>({
      query: ActivityTransactionDocument,
      variables: {
        id: String(params.id)
      },
      context: {
        headers: {
          cookie: cookie
        }
      }
    })
    .then((val) => val.data.transaction)

  return json({
    transaction: {
      id: transaction.id,
      amount: transaction.amount,
      transactionType: transaction.type,
      title: activityTitle(transaction.type),
      description: transaction.description,
      status: transaction.status,
      date: DateTime.fromISO(transaction.timestamp).toFormat(
        'dd LLLL yyyy HH:mm'
      )
    }
  })
}

const activityTitle = (type: TransactionType): string => {
  switch (type) {
    // case TransactionType.Received:
    //   return <ReceivedIcon />
    case TransactionType.Outgoingpayment:
      return 'Sent'
    case TransactionType.Deposit:
      return 'Deposit'
    case TransactionType.Withdrawal:
      return 'Withdrawal'
    default:
      return ''
  }
}

// TODO: Replace with implementation in the backend.
const activityIcon = (type: TransactionType, status: string) => {
  switch (status) {
    case 'pending':
      return 'hourglass_empty'
    default:
      break
  }
  switch (type) {
    // case TransactionType.Received:
    //   return <ReceivedIcon />
    case TransactionType.Outgoingpayment:
      return 'north_east'
    case TransactionType.Deposit:
      return 'account_balance'
    case TransactionType.Withdrawal:
      return 'account_balance'
    default:
      return null
  }
}

export default function Page() {
  const navigate = useNavigate()
  const { transaction } = useLoaderData<LoaderData>()
  return (
    <div className='w-full'>
      {/* Header */}
      <header className='sticky top-0 mx-auto flex h-16 w-full select-none items-center justify-start bg-white p-4 text-medium sm:max-w-lg lg:max-w-3xl xl:max-w-4xl'>
        <button
          onClick={() => {
            navigate(-1)
          }}
        >
          <div className='-ml-3 p-3 text-medium'>
            <Icon>arrow_back</Icon>
          </div>
        </button>
        <div className='flex items-center justify-start font-display text-2xl font-medium'>
          Transaction
        </div>
      </header>
      {/* Body */}
      <div className='mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
        <div className='col-span-full mb-4 flex h-12 items-center justify-center  sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <div className='flex items-center justify-between space-x-3'>
            <Icon>
              {activityIcon(transaction.transactionType, transaction.status)}
            </Icon>
            <div className='flex flex-col'>
              <span className='text-left font-display text-base font-medium'>
                {transaction.title}
              </span>
              <span className='font-sans text-xs font-normal'>
                {transaction.description}
              </span>
            </div>
          </div>
        </div>
        <div className='col-span-full mt-2 flex h-12 flex-col items-start justify-center  sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <span className='mb-1 font-display text-xs font-medium'>Amount</span>
          <span className='font-sans text-base font-normal'>
            {transaction.amount}
          </span>
        </div>
        <div className='col-span-full mt-2 flex h-12 flex-col items-start justify-center  sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <span className='mb-1 font-display text-xs font-medium'>Date</span>
          <span className='font-sans text-base font-normal'>
            {transaction.date}
          </span>
        </div>
        <div className='col-span-full mt-2 flex h-12 flex-col items-start justify-center  sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <span className='mb-1 font-display text-xs font-medium'>
            Transaction ID
          </span>
          <span className='font-sans text-base font-normal'>
            {transaction.id}
          </span>
        </div>
        <div className='col-span-full mt-2 flex h-12 flex-col items-start justify-center  sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <span className='mb-1 font-display text-xs font-medium'>Status</span>
          <span className='font-sans text-base font-normal capitalize'>
            {transaction.status}
          </span>
        </div>
      </div>
    </div>
  )
}
