import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Link, useLoaderData } from '@remix-run/react'
import { DateTime } from 'luxon'
import type { FC } from 'react'
import { Fragment } from 'react'
import { route } from 'routes-gen'
import {
  Container,
  HomeDecor,
  Footer,
  Header,
  Icon,
  Router
} from '~/components'
import type { HomeQuery, HomeQueryVariables } from '~/generated/types'
import { HomeDocument, TransactionType } from '~/generated/types'
import { apolloClient } from '~/lib/apollo.server'
import { hasUserSession, requireUserSession } from '~/lib/kratos.server'

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

export async function loader({ request }: LoaderArgs) {
  const isUser = await hasUserSession(request)

  let data = {
    isUser: isUser,
    balance: '0',
    recentActivities: [] as Activities[],
    pendingTransactions: [] as Activity[]
  }

  if (isUser) {
    await requireUserSession(request)
    const cookie = request.headers.get('cookie')

    const account = await apolloClient
      .query<HomeQuery, HomeQueryVariables>({
        query: HomeDocument,
        context: {
          headers: {
            cookie: cookie
          }
        }
      })
      .then((val) => val.data.account)

    data.balance = account?.balance as string

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
          data.pendingTransactions.push(activity)
          continue
        }
        const date = DateTime.fromISO(trx.timestamp).toFormat('dd LLLL yyyy')
        const indexToPush = data.recentActivities.findIndex(
          (val) => val.date == date
        )
        if (indexToPush >= 0)
          data.recentActivities[indexToPush].activities.push(activity)
        else
          data.recentActivities.push({
            date: date,
            activities: [activity]
          })
      }
  }
  return json(data)
}

export default function Page() {
  const { isUser } = useLoaderData<typeof loader>()

  if (isUser) return <AppPage />
  else return <MarketingPage />
}

function MarketingPage() {
  return (
    <Container className='overflow-x-hidden'>
      <Header />
      <main className='flex-grow'>
        <div className='mt-20 mb-12 flex w-[340px] flex-col space-y-8 px-4 sm:mt-44 sm:mb-20 sm:p-8'>
          <span className='font-display text-4xl font-medium leading-normal'>
            Connecting
            <br />
            the Internet
            <br />
            economy
          </span>
        </div>
      </main>
      <Footer />
      <HomeDecor />
    </Container>
  )
}

function AppPage() {
  const { balance, recentActivities, pendingTransactions } =
    useLoaderData<typeof loader>()
  return (
    <div className='w-full'>
      {/* Header */}
      <header className='sticky top-0 flex h-16 min-w-full select-none items-center justify-between bg-app p-4 text-medium'>
        <div className='flex items-center justify-start font-display text-2xl font-medium'>
          Home
        </div>
        <Link className='sm:hidden' to={route('/settings')}>
          <div className='-mr-3 p-3 text-medium'>
            <Icon>settings</Icon>
          </div>
        </Link>
      </header>
      {/* Body */}
      <div className='mx-auto grid min-h-[calc(100vh-9rem)] w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'>
        {/* HOME */}
        <div className='col-span-full flex flex-col items-center px-3 pt-4 pb-2 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
          <span className='font-sans text-base font-normal'>Balance</span>
          <span className='font-display text-4xl font-normal'>{balance}</span>
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
        {recentActivities.length > 0 && (
          <div className='col-span-full flex justify-between pt-2 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
            <span className='font-display text-lg font-medium'>
              Recent activity
            </span>
            <Router to={route('/activity')}>
              <div className='flex items-center space-x-1 font-display text-sm font-medium text-primary'>
                <span>See all</span>
                <Icon>read_more</Icon>
              </div>
            </Router>
          </div>
        )}
        {/* Activity items */}
        {recentActivities.map((activities) => (
          <Fragment key={activities.date}>
            <span className='col-span-full ml-4 mt-2 font-display text-xs font-normal sm:col-span-6 sm:col-start-2 lg:col-start-4'>
              {activities.date}
            </span>
            {activities.activities.map((activity) => (
              <ActivityCard key={activity.id} activity={activity} />
            ))}
          </Fragment>
        ))}
        {pendingTransactions.length > 0 && (
          <div className='col-span-full flex justify-start pt-4 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
            <span className='font-display text-lg font-medium'>Pending</span>
          </div>
        )}
        {pendingTransactions.map((activity) => (
          <ActivityCard key={activity.id} activity={activity} />
        ))}
      </div>
    </div>
  )
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

const ActivityCard: FC<{ activity: Activity }> = ({ activity }) => {
  return (
    <Router
      to={route('/activity/transaction/:id', {
        id: activity.id
      })}
      className='col-span-full flex items-center justify-between rounded-xl bg-container py-2 px-3 sm:col-span-6 sm:col-start-2 lg:col-start-4'
    >
      <div className='flex items-center justify-between space-x-3'>
        <Icon>{activityIcon(activity.transactionType, activity.status)}</Icon>
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
