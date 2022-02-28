import { json, Link, useLoaderData } from 'remix'
import type { LoaderFunction } from 'remix'
import { route } from 'routes-gen'
import {
  BankIcon,
  CardIcon,
  PendingIcon,
  ReadMoreIcon,
  ReceivedIcon,
  Router,
  SentIcon,
  SettingsIcon
} from '~/components'
import { requireUserSession } from '~/lib/kratos'
import React, { FC } from 'react'

enum TransactionType {
  Received = 'Received',
  Sent = 'Sent',
  Deposit = 'Deposit',
  Withdrawal = 'Withdrawal'
}

enum TransactionStatus {
  Pending = 'Pending',
  Complete = 'Complete',
  Cancelled = 'Cancelled',
  Failed = 'Failed'
}

type Activity = {
  id: string
  amount: string
  transactionType: TransactionType
  title: string
  description: string
  status: TransactionStatus
}

type Activities = {
  date: string
  activities: Activity[]
}

type LoaderData = {
  balance: string
  recentActivities: Activities[]
  pendingTransactions: Activity[]
}

export const loader: LoaderFunction = async ({ request }) => {
  await requireUserSession(request)
  const data: LoaderData = {
    balance: '$ 183.00',
    recentActivities: [
      {
        date: '24 January 2022',
        activities: [
          {
            id: '3',
            amount: '$ 1.00',
            transactionType: TransactionType.Received,
            title: 'Received',
            description: 'from Interledger',
            status: TransactionStatus.Complete
          },
          {
            id: '6',
            amount: '$ 1.00',
            transactionType: TransactionType.Withdrawal,
            title: 'Withdrawal',
            description: 'to xxxx bank account',
            status: TransactionStatus.Complete
          }
        ]
      },
      {
        date: '21 January 2022',
        activities: [
          {
            id: '2',
            amount: '$ 14.00',
            transactionType: TransactionType.Sent,
            title: 'Sent',
            description: 'to Interledger',
            status: TransactionStatus.Complete
          },
          {
            id: '1',
            amount: '$ 200.00',
            transactionType: TransactionType.Deposit,
            title: 'Deposit',
            description: 'from card ending 4242',
            status: TransactionStatus.Complete
          }
        ]
      }
    ],
    pendingTransactions: [
      {
        id: '7',
        amount: '$ 1.00',
        transactionType: TransactionType.Deposit,
        title: 'Incoming deposit',
        description: 'from card ending 4242',
        status: TransactionStatus.Pending
      },
      {
        id: '8',
        amount: '$ 1.00',
        transactionType: TransactionType.Withdrawal,
        title: 'Outgoing withdrawal',
        description: 'to xxxx bank account',
        status: TransactionStatus.Pending
      },
      {
        id: '9',
        amount: '$ 1.00',
        transactionType: TransactionType.Received,
        title: 'Incoming payment',
        description: 'from Interledger',
        status: TransactionStatus.Pending
      },
      {
        id: '10',
        amount: '$ 1.00',
        transactionType: TransactionType.Sent,
        title: 'Outgoing payment',
        description: 'to Interledger',
        status: TransactionStatus.Pending
      }
    ]
  }
  return json(data)
}

export default function Home() {
  const account = useLoaderData<LoaderData>()
  return (
    <div className='w-full'>
      {/* Header */}
      <header className='sticky top-0 flex h-16 min-w-full select-none items-center justify-between bg-white p-4 text-medium'>
        <div className='flex items-center justify-start font-display text-2xl font-medium'>
          Home
        </div>
        <Link className='sm:hidden' to={route('/settings')}>
          <div className='-mr-3 p-3 text-medium'>
            <SettingsIcon />
          </div>
        </Link>
      </header>
      {/* Body */}
      <div className='mx-auto grid min-h-[calc(100vh-9rem)] w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
        {/* HOME */}
        <div className='col-span-full flex flex-col items-center px-3 pt-4 pb-2 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <span className='font-sans text-base font-normal'>Balance</span>
          <span className='font-display text-4xl font-normal'>
            {account.balance}
          </span>
        </div>
        <div className='col-span-full flex justify-center space-x-3 py-4 px-3 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <Router to={route('/deposit')}>
            <div className='flex h-10 w-36 cursor-pointer items-center justify-center rounded-full bg-container-primary font-display text-sm font-medium text-medium hover:bg-container-primary-hover focus:outline-none  focus:ring-2 focus:ring-focus active:bg-container-primary-active'>
              Deposit
            </div>
          </Router>
          <Router to={route('/withdraw')}>
            <div className='flex h-10 w-36 cursor-pointer items-center justify-center rounded-full bg-container-primary font-display text-sm font-medium text-medium hover:bg-container-primary-hover focus:outline-none  focus:ring-2 focus:ring-focus active:bg-container-primary-active'>
              Withdraw
            </div>
          </Router>
        </div>
        <div className='col-span-full flex justify-between pt-2 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <span className='font-display text-lg font-medium'>
            Recent activity
          </span>
          <Router to={route('/activity')}>
            <div className='flex items-center space-x-1 font-display text-sm font-medium text-primary'>
              <span>See all</span>
              <ReadMoreIcon />
            </div>
          </Router>
        </div>
        {/* Activity items */}
        {account.recentActivities.map((activities) => (
          <React.Fragment key={activities.date}>
            <span className='col-span-full ml-4 mt-2 font-display text-xs font-normal sm:col-span-6 sm:col-start-2 lg:col-start-4'>
              {activities.date}
            </span>
            {activities.activities.map((activity) => (
              <ActivityCard key={activity.id} activity={activity} />
            ))}
          </React.Fragment>
        ))}
        <div className='col-span-full flex justify-start pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <span className='font-display text-lg font-medium'>Pending</span>
        </div>
        {account.pendingTransactions.map((activity) => (
          <ActivityCard key={activity.id} activity={activity} />
        ))}
      </div>
    </div>
  )
}

const activityIcon = (type: TransactionType, status: TransactionStatus) => {
  switch (status) {
    case TransactionStatus.Pending:
      return <PendingIcon />
    default:
      break
  }
  switch (type) {
    case TransactionType.Received:
      return <ReceivedIcon />
    case TransactionType.Sent:
      return <SentIcon />
    case TransactionType.Deposit:
      return <CardIcon />
    case TransactionType.Withdrawal:
      return <BankIcon />
    default:
      return null
  }
}

export const ActivityCard: FC<{ activity: Activity }> = ({ activity }) => {
  return (
    <Link
      to={route('/activity/transaction/:id', {
        id: activity.id
      })}
      className='col-span-full flex items-center justify-between rounded-xl bg-container py-2 px-3 sm:col-span-6 sm:col-start-2 lg:col-start-4'
    >
      <div className='flex items-center justify-between space-x-3'>
        {activityIcon(activity.transactionType, activity.status)}
        <div className='flex flex-col'>
          <span className='font-display text-base font-medium'>
            {activity.title}
          </span>
          <span className='font-sans text-xs font-normal'>
            {activity.description}
          </span>
        </div>
      </div>
      <span className='font-sans text-lg font-normal'>{activity.amount}</span>
    </Link>
  )
}
