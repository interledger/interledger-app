import type { FC } from 'react'
import { useCallback, useEffect, useState, Fragment } from 'react'
import type { ActionArgs, LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { Form, Link, useFetcher, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Icon } from '~/components'
import { apolloClient } from '~/lib/apollo.server'
import { requireUserSession } from '~/lib/kratos.server'
import type {
  ActivityQuery,
  ActivityQueryVariables,
  PageInfo
} from '~/generated/types'
import { ActivityDocument, TransactionType } from '~/generated/types'
import { DateTime } from 'luxon'
import { commitSession, getSession } from '~/sessions'

type Activity = {
  id: string
  amount: string
  transactionType: TransactionType
  title: string
  description: string
  status: string
  date: string
}

export async function loader({ request }: LoaderArgs) {
  await requireUserSession(request)
  const searchParams = new URL(request.url).searchParams

  const cursor = String(searchParams.get('cursor') || '')
  const cookie = request.headers.get('cookie')
  const userSettings = await getSession(cookie)

  const activitySettings = userSettings.get('activity')
  const pages = activitySettings?.pages || 1

  const activities = await apolloClient
    .query<ActivityQuery, ActivityQueryVariables>({
      query: ActivityDocument,
      variables: {
        input: {
          after: cursor,
          first: 30 * pages
        }
      },
      context: {
        headers: {
          cookie: cookie
        }
      }
    })
    .then((val) => val.data.transactions)
  // TODO Handle empty case: val.data == null
  let transactions = activities?.edges.map((trx) => ({
    id: trx.node.id,
    amount: trx.node.amount,
    transactionType: trx.node.type,
    title: activityTitle(trx.node.type),
    description: trx.node.description,
    status: trx.node.status,
    date: DateTime.fromISO(trx.node.timestamp).toFormat('dd LLLL yyyy')
  }))

  const data = {
    pageInfo: activities.pageInfo,
    transactions,
    pages: pages
  }

  userSettings.unset('activity')
  return json(data, {
    headers: {
      'Set-Cookie': await commitSession(userSettings)
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

export const ActivityCard: FC<{ activity: Activity }> = ({ activity }) => {
  return (
    <button
      form='activity-control'
      type='submit'
      name='activity-id'
      value={activity.id}
      className='col-span-full flex items-center justify-between rounded-xl bg-container py-2 px-3 focus-visible:outline-2 focus-visible:outline-focus sm:col-span-6 sm:col-start-2 lg:col-start-4'
    >
      <div className='flex items-center justify-between space-x-3'>
        <Icon>{activityIcon(activity.transactionType, activity.status)}</Icon>
        <div className='flex flex-col'>
          <span className='text-left font-display text-base font-medium'>
            {activity.title}
          </span>
          <span className='font-sans text-xs font-normal'>
            {activity.description}
          </span>
        </div>
      </div>
      <span className='font-sans text-lg font-normal'>{activity.amount}</span>
    </button>
  )
}

export default function Page() {
  const initialPage = useLoaderData<typeof loader>()
  const fetcher = useFetcher()

  const [transactions, setTransactions] = useState<
    typeof initialPage.transactions
  >(initialPage.transactions)
  const [pageInfo, setPageInfo] = useState<PageInfo>(initialPage.pageInfo)
  const [pages, setPages] = useState<number>(initialPage.pages)

  const [scrollPosition, setScrollPosition] = useState(0)
  const [clientHeight, setClientHeight] = useState(0)
  const [height, setHeight] = useState(null)
  const [shouldFetch, setShouldFetch] = useState(true)

  const divHeight = useCallback(
    (node) => {
      if (node !== null) {
        setHeight(node.getBoundingClientRect().height)
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [transactions.length]
  )

  // Add Listeners to scroll and client resize
  useEffect(() => {
    const scrollListener = () => {
      setClientHeight(window.innerHeight)
      setScrollPosition(window.scrollY)
    }

    // Avoid running during SSR
    if (typeof window !== 'undefined') {
      window.addEventListener('scroll', scrollListener)
    }

    // Clean up
    return () => {
      if (typeof window !== 'undefined') {
        window.removeEventListener('scroll', scrollListener)
      }
    }
  }, [clientHeight, scrollPosition])

  useEffect(() => {
    if (!shouldFetch || !height) return
    if (clientHeight + scrollPosition + 100 < height) return

    if (pageInfo.hasNextPage) {
      fetcher.load(`/activity?cursor=${pageInfo.endCursor}`)
    }
    setShouldFetch(false)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [clientHeight, scrollPosition, fetcher])

  useEffect(() => {
    if (fetcher.data && fetcher.data.transactions.length > 0) {
      setTransactions((prevTransactions: Activity[]) => {
        return [...prevTransactions, ...fetcher.data.transactions]
      })
      setPages((pages) => pages + 1)
      setPageInfo(fetcher.data.pageInfo)
      setShouldFetch(true)
    }
  }, [fetcher.data])

  return (
    <div className='w-full'>
      {/* Header */}
      <header className='sticky top-0 flex h-16 min-w-full select-none items-center justify-between bg-white p-4 text-medium'>
        <div className='flex items-center justify-start font-display text-2xl font-medium'>
          Activity
        </div>
        <Link className='sm:hidden' to={route('/settings')}>
          <div className='-mr-3 p-3 text-medium'>
            <Icon>settings</Icon>
          </div>
        </Link>
      </header>
      {/* Body */}
      <div
        ref={divHeight}
        className='mx-auto grid min-h-[calc(100vh-9rem)] w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'
      >
        <Form
          id='activity-control'
          className='hidden'
          action={`/activity`}
          method='post'
        />
        <input
          form='activity-control'
          type='hidden'
          name='pages'
          value={pages}
        />
        {/* Activity item */}
        {transactions.length == 0 && (
          <div className='col-span-full flex items-center justify-start space-x-3 rounded-xl bg-container p-3 text-medium sm:col-span-6 sm:col-start-2 lg:col-start-4'>
            <Icon>tips_and_updates</Icon>
            <span className='font-sans text-sm font-normal'>
              You don't have any activity yet.
            </span>
          </div>
        )}
        {transactions.map((transaction, index) => (
          <Fragment key={transaction.id}>
            {(index == 0 ||
              transaction.date != transactions[index - 1].date) && (
              <span className='col-span-full ml-4 mt-2 font-display text-xs font-normal sm:col-span-6 sm:col-start-2 lg:col-start-4'>
                {transaction.date}
              </span>
            )}
            <ActivityCard key={transaction.id} activity={transaction} />
          </Fragment>
        ))}
      </div>
    </div>
  )
}

export async function action({ request }: ActionArgs) {
  const userSettings = await getSession(request.headers.get('Cookie'))

  const form = await request.formData()
  const pages = form.get('pages') as string
  const activityId = form.get('activity-id') as string

  userSettings.set('activity', { pages })

  return redirect(
    route('/activity/transaction/:id', {
      id: String(activityId)
    }),
    {
      headers: {
        'Set-Cookie': await commitSession(userSettings)
      }
    }
  )
}
