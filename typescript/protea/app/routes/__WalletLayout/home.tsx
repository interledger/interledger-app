import { json, Link, useLoaderData } from 'remix'
import type { LoaderFunction } from 'remix'
import { route } from 'routes-gen'
import {
  BankIcon,
  CardIcon,
  PendingIcon,
  ReadMoreIcon,
  Router,
  SentIcon,
  SettingsIcon
} from '~/components'
import { requireUserSession } from '~/lib/kratos.server'
import React, { FC } from 'react'
import {
  GetHomeDocument,
  GetHomeQuery,
  GetHomeQueryVariables,
  TransactionType
} from '~/generated/types'
import { apolloClient } from '~/lib/apollo.server'
import { DateTime } from 'luxon'

type Activity = {
  id: string
  amount: string
  transactionType: TransactionType
  title: string
  description: string
  status: string
}

type Activities = {
  date: string
  activities: Activity[]
}

type LoaderData = {
  balance?: string
  recentActivities: Activities[]
  pendingTransactions: Activity[]
}

export const loader: LoaderFunction = async ({ request }) => {
  await requireUserSession(request)
  const cookie = request.headers.get('cookie')

  const account = await apolloClient
    .query<GetHomeQuery, GetHomeQueryVariables>({
      query: GetHomeDocument,
      context: {
        headers: {
          cookie: cookie
        }
      }
    })
    .then((val) => val.data.account)

  let recentActivities: LoaderData['recentActivities'] = []
  let pendingTransactions: LoaderData['pendingTransactions'] = []
  if (typeof account?.recentTransactions !== 'undefined')
    for (let trx of account?.recentTransactions) {
      const activity = {
        id: trx.id,
        amount: trx.amount,
        transactionType: trx.type,
        title: activityTitle(trx.type),
        description: trx.description,
        status: trx.status
      }
      if (activity.status == 'pending') {
        pendingTransactions.push(activity)
        continue
      }
      const date = DateTime.fromISO(trx.timestamp).toFormat('dd LLLL yyyy')
      const indexToPush = recentActivities.findIndex((val) => val.date == date)
      if (indexToPush >= 0)
        recentActivities[indexToPush].activities.push(activity)
      else
        recentActivities.push({
          date: date,
          activities: [activity]
        })
    }

  const data: LoaderData = {
    balance: account?.balance,
    recentActivities: recentActivities,
    pendingTransactions: pendingTransactions
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
          <Router
            className='rounded-full'
            to={route('/flows/:flowId/deposit/payment-method', {
              flowId: 'init'
            })}
          >
            <div className='flex h-10 w-36 cursor-pointer items-center justify-center rounded-full bg-container-primary font-display text-sm font-medium text-medium hover:bg-container-primary-hover active:bg-container-primary-active'>
              Deposit
            </div>
          </Router>
          <Router
            className='rounded-full'
            to={route('/flows/:flowId/withdraw/payment-method', {
              flowId: 'init'
            })}
          >
            <div className='flex h-10 w-36 cursor-pointer items-center justify-center rounded-full bg-container-primary font-display text-sm font-medium text-medium hover:bg-container-primary-hover active:bg-container-primary-active'>
              Withdraw
            </div>
          </Router>
        </div>
        {account.recentActivities.length > 0 && (
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
        )}
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
        {account.pendingTransactions.length > 0 && (
          <div className='col-span-full flex justify-start pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
            <span className='font-display text-lg font-medium'>Pending</span>
          </div>
        )}
        {account.pendingTransactions.map((activity) => (
          <ActivityCard key={activity.id} activity={activity} />
        ))}
      </div>
    </div>
  )
}

const activityIcon = (type: TransactionType, status: string) => {
  switch (status) {
    case 'pending':
      return <PendingIcon />
    default:
      break
  }
  switch (type) {
    // case TransactionType.Received:
    //   return <ReceivedIcon />
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

const activityTitle = (type: TransactionType): string => {
  switch (type) {
    // case TransactionType.Received:
    //   return <ReceivedIcon />
    case TransactionType.Sent:
      return 'Sent'
    case TransactionType.Deposit:
      return 'Deposit'
    case TransactionType.Withdrawal:
      return 'Withdrawal'
    default:
      return ''
  }
}

export const ActivityCard: FC<{ activity: Activity }> = ({ activity }) => {
  return (
    <Router
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
    </Router>
  )
}
