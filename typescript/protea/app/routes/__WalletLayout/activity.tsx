import React, { FC } from 'react'
import { Link, LoaderFunction, json, useLoaderData } from 'remix'
import { route } from 'routes-gen'
import {
  BackIcon,
  BankIcon,
  CardIcon,
  PendingIcon,
  Router,
  SentIcon
} from '~/components'
import { apolloClient } from '~/lib/apollo.server'
import { requireUserSession } from '~/lib/kratos.server'
import {
  GetActivityQuery,
  GetActivityQueryVariables,
  GetActivityDocument,
  TransactionType,
  PageInfo
} from '~/generated/types'
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
  pageInfo: PageInfo
  transactions: Activities[]
}

export const loader: LoaderFunction = async ({ request }) => {
  await requireUserSession(request)
  const cookie = request.headers.get('cookie')
  const activities = await apolloClient
    .query<GetActivityQuery, GetActivityQueryVariables>({
      query: GetActivityDocument,
      variables: {
        input: {
          after: '',
          first: 10
        }
      },
      context: {
        headers: {
          cookie: cookie
        }
      }
    })
    .then((val) => val.data.transactions)

  let transactions: LoaderData['transactions'] = []
  if (typeof activities?.edges !== 'undefined')
    for (let trx of activities?.edges) {
      const activity = {
        id: trx.node.id,
        amount: trx.node.amount,
        transactionType: trx.node.type,
        title: activityTitle(trx.node.type),
        description: trx.node.description,
        status: trx.node.status
      }
      const date = DateTime.fromISO(trx.node.timestamp).toFormat('dd LLLL yyyy')
      const indexToPush = transactions.findIndex((val) => val.date == date)
      if (indexToPush >= 0) transactions[indexToPush].activities.push(activity)
      else
        transactions.push({
          date: date,
          activities: [activity]
        })
    }

  const data: LoaderData = {
    pageInfo: activities.pageInfo,
    transactions
  }
  return json(data)
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

export default function ActivityPage() {
  const { transactions } = useLoaderData<LoaderData>()
  return (
    <div className='w-full'>
      {/* Header */}
      <header className='sticky top-0 mx-auto flex h-16 w-full select-none items-center justify-start bg-white p-4 text-medium sm:max-w-lg lg:max-w-3xl xl:max-w-4xl'>
        <Link to={route('/home')}>
          <div className='-ml-3 p-3 text-medium'>
            <BackIcon />
          </div>
        </Link>
        <div className='flex items-center justify-start font-display text-2xl font-medium'>
          Activity
        </div>
      </header>
      {/* Body */}
      <div className='mx-auto grid min-h-[calc(100vh-9rem)] w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
        {/* Activity item */}
        {transactions.map((activities) => (
          <React.Fragment key={activities.date}>
            <span className='col-span-full ml-4 mt-2 font-display text-xs font-normal sm:col-span-6 sm:col-start-2 lg:col-start-4'>
              {activities.date}
            </span>
            {activities.activities.map((activity) => (
              <ActivityCard key={activity.id} activity={activity} />
            ))}
          </React.Fragment>
        ))}
      </div>
      {/* TODO enable FAB when filter page is ready. */}
      {/* <FAB
        onClick={() => navigate(route('/activity/filter'))}
        icon={<FilterIcon />}
      >
        Filter
      </FAB> */}
    </div>
  )
}
