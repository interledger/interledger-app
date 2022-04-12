import React, { FC, useCallback, useEffect, useState } from 'react'
import {
  Link,
  LoaderFunction,
  json,
  useLoaderData,
  useFetcher,
  Form,
  redirect,
  ActionFunction
} from 'remix'
import { route } from 'routes-gen'
import {
  BackIcon,
  BankIcon,
  CardIcon,
  PendingIcon,
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

type LoaderData = {
  pageInfo: PageInfo
  transactions: Activity[]
  pages: number
}

export const loader: LoaderFunction = async ({ request }) => {
  await requireUserSession(request)
  const searchParams = new URL(request.url).searchParams

  const cursor = String(searchParams.get('cursor') || '')
  const cookie = request.headers.get('cookie')
  const userSettings = await getSession(cookie)

  const activitySettings = userSettings.get('activity')
  const pages = activitySettings?.pages || 1

  const activities = await apolloClient
    .query<GetActivityQuery, GetActivityQueryVariables>({
      query: GetActivityDocument,
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

  let transactions: LoaderData['transactions'] = activities?.edges.map(
    (trx) => ({
      id: trx.node.id,
      amount: trx.node.amount,
      transactionType: trx.node.type,
      title: activityTitle(trx.node.type),
      description: trx.node.description,
      status: trx.node.status,
      date: DateTime.fromISO(trx.node.timestamp).toFormat('dd LLLL yyyy')
    })
  )

  const data: LoaderData = {
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
    <button
      form='activity-control'
      type='submit'
      name='activity-id'
      value={activity.id}
      className='col-span-full flex items-center justify-between rounded-xl bg-container py-2 px-3 focus-visible:outline-2 focus-visible:outline-focus sm:col-span-6 sm:col-start-2 lg:col-start-4'
    >
      <div className='flex items-center justify-between space-x-3'>
        {activityIcon(activity.transactionType, activity.status)}
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

export default function ActivityPage() {
  const initialPage = useLoaderData<LoaderData>()
  const fetcher = useFetcher()

  const [transactions, setTransactions] = useState<LoaderData['transactions']>(
    initialPage.transactions
  )
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
      <div
        ref={divHeight}
        className='mx-auto grid w-full grid-cols-4 content-start gap-4 gap-y-2 overflow-y-auto p-4 pb-24 sm:max-w-lg sm:grid-cols-8 sm:px-0 lg:max-w-3xl lg:grid-cols-12 xl:max-w-4xl'
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
        {transactions.map((transaction, index) => (
          <React.Fragment key={transaction.id}>
            {(index == 0 ||
              transaction.date != transactions[index - 1].date) && (
              <span className='col-span-full ml-4 mt-2 font-display text-xs font-normal sm:col-span-6 sm:col-start-2 lg:col-start-4'>
                {transaction.date}
              </span>
            )}
            <ActivityCard key={transaction.id} activity={transaction} />
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

export const action: ActionFunction = async ({ request }) => {
  const form = await request.formData()
  const pages = form.get('pages')
  const activityId = form.get('activity-id')
  const userSettings = await getSession(request.headers.get('Cookie'))

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
