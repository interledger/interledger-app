import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import type { ShouldRevalidateFunction } from '@remix-run/react'
import {
  Outlet,
  useFetcher,
  useLoaderData,
  useLocation,
  useSearchParams
} from '@remix-run/react'
import clsx from 'clsx'
import { useCallback, useEffect, useState } from 'react'
import { route } from 'routes-gen'
import type { ApplicationProps } from '~/components'
import {
  Card,
  CardContent,
  CardHeader,
  CardIcon,
  CardLink,
  CardTitle,
  Chip,
  ChipColor,
  Fab,
  GridColumn,
  Icon,
  InterledgerIcon,
  Layouts,
  Router,
  SlackIcon,
  TwitterIcon,
  WalletGrid
} from '~/components'
import { Label } from '~/components/Label'
import { getKycStatus, getTransactionsWithPending } from '~/data/wallet.server'
import type { Transaction } from '~/generated/connect/backend/v1/backend_pb'
import { grpc } from '~/lib/grpc.server'

import { isConnectError } from '~/lib/error.server'
import { mergeMeta } from '~/lib/meta'
import { KycStatus } from '~/routes/_index/route'

/**
 * Allows us to change the searchParams without revalidating the pages data
 * This is useful for pagination.
 */
export const shouldRevalidate: ShouldRevalidateFunction = ({
  currentUrl,
  defaultShouldRevalidate,
  nextUrl
}) => {
  if (currentUrl.search !== nextUrl.search) return false
  return defaultShouldRevalidate
}

export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url)
  const pages = parseInt(url.searchParams.get('pages') || '1')

  const kycStatus = await getKycStatus(request)
  let pageInfo = {
    // pageToken is only set by fetcher so initial page loads this should be blank
    pageToken: url.searchParams.get('pageToken') || '',
    pageSize: 30
  }

  let allTransactions: Transaction[] = []

  const webmo = await grpc.listPendingWebMonetization(request, {})
  if (isConnectError(webmo)) throw webmo.errorResponse

  allTransactions = webmo.transactions

  /**
   * We can loop over pages as fetcher should omit this.
   * This allows us to pull all necessary data when navigating back to this page.
   */
  for (let i = 0; i < pages; i++) {
    const { transactions, nextPageToken } = await getTransactionsWithPending(
      request,
      pageInfo
    )
    pageInfo.pageToken = nextPageToken
    allTransactions = [...allTransactions, ...transactions]
    if (nextPageToken == '') break
  }

  const dateGroupedTransactions = Object.values(
    [...allTransactions].reduce<{
      [date: string]: Transaction[]
    }>((prev, current) => {
      prev[current.formattedDate] = prev[current.formattedDate] || []
      prev[current.formattedDate].push(current)
      return prev
    }, Object.create(null))
  )

  return json({
    kycStatus: kycStatus.kycStatus,
    transactions: dateGroupedTransactions,
    nextPageToken: pageInfo.pageToken
  })
}

export const handle: ApplicationProps = {
  layout: Layouts.Wallet,
  scaffold: {
    header: {
      title: 'Payments'
    },
    fab: Fab.Pay
  }
}

export const meta: MetaFunction = mergeMeta(() => [
  {
    title: 'Payments'
  }
])

export default function Page() {
  const initialPage = useLoaderData<typeof loader>()
  let [, setSearchParams] = useSearchParams()
  const fetcher = useFetcher<typeof loader>()
  const [transactions, setTransactions] = useState(initialPage.transactions)
  const [nextPageToken, setNextPageToken] = useState<string>(
    initialPage.nextPageToken
  )
  const [scrollPosition, setScrollPosition] = useState(0)
  const [clientHeight, setClientHeight] = useState(0)
  const [height, setHeight] = useState(null)
  const [shouldFetch, setShouldFetch] = useState(true)

  const isMobile = typeof document !== 'undefined' && window.innerWidth < 1024

  const divHeight = useCallback(
    (node: any) => {
      if (node !== null) {
        setHeight(node.getBoundingClientRect().height)
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [transactions?.length]
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

  // Trigger fetching new data when the user scrolls near the bottom
  useEffect(() => {
    if (!shouldFetch || !height) return
    if (clientHeight + scrollPosition + 100 < height) return

    if (nextPageToken != '') {
      fetcher.load(`/payments?pageToken=${nextPageToken}`)
      setSearchParams(
        (old) => {
          old.set('pages', `${parseInt(old.get('pages') || '1') + 1}`)
          return old
        },
        { replace: true, preventScrollReset: true }
      )
    }

    setShouldFetch(false)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [clientHeight, scrollPosition])

  // Handle new transactions being fetched and add to transaction list
  useEffect(() => {
    if (fetcher.data && fetcher.data.transactions.length > 0) {
      setTransactions((currentTransactions) => {
        if (!fetcher.data) return currentTransactions

        const lastOfCurrent =
          currentTransactions[currentTransactions.length - 1]

        const newTransactions = fetcher.data.transactions
        // Verifies if current and new transaction sets have date group overlap
        if (
          lastOfCurrent[lastOfCurrent.length - 1].formattedDate ==
          newTransactions[0][0].formattedDate
        ) {
          const last = currentTransactions.pop() ?? []
          const first = newTransactions.shift()
          if (!first) return [...currentTransactions, ...newTransactions]
          return [
            ...currentTransactions,
            [...last, ...first],
            ...newTransactions
          ]
        }
        return [...currentTransactions, ...newTransactions]
      })

      setNextPageToken(fetcher.data.nextPageToken || '')
      setShouldFetch(true)
    }
  }, [fetcher.data])

  // Handle transactions being revalidated
  useEffect(() => {
    setTransactions(initialPage.transactions)
  }, [initialPage.transactions])

  const location = useLocation()
  const pathSegments = location.pathname.split('/').filter(Boolean)

  return (
    <WalletGrid ref={divHeight}>
      <GridColumn
        hideOnMobile={pathSegments[pathSegments.length - 1] !== 'payments'}
        className='col-span-full lg:col-span-6'
      >
        {initialPage.kycStatus == KycStatus.Unknown && (
          <Card>
            <CardHeader>
              <CardTitle>Wallet</CardTitle>
              <Chip color={ChipColor.orange}>Reserved</Chip>
            </CardHeader>
            <CardContent>
              <div className='flex items-start space-x-4'>
                <CardIcon>
                  <Icon>account_balance_wallet</Icon>
                </CardIcon>
                <div className='flex flex-col space-y-4'>
                  <p className='text-sm text-medium'>
                    Your wallet is reserved, we just need a few more details to
                    activate it.
                  </p>
                  <Router
                    prefetch='render'
                    className='text-sm font-medium text-primary'
                    to={route('/personal-details')}
                  >
                    Activate wallet
                  </Router>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
        {initialPage.kycStatus == KycStatus.Approved &&
          transactions &&
          transactions.length == 0 && (
            <Card>
              <CardContent>
                <div className='flex flex-col space-y-4'>
                  <span className='text-sm text-medium'>
                    Your payment activity will appear here once you start
                    transacting.
                  </span>
                  <Router
                    to={route('/pay')}
                    className='text-sm font-medium text-primary'
                  >
                    Send or receive payments now
                  </Router>
                </div>
              </CardContent>
            </Card>
          )}
        {transactions &&
          transactions.map((transactionGroup, index) => (
            <Card key={`group-${index}`}>
              <Label>{transactionGroup[0].formattedDate}</Label>
              {transactionGroup.map((transaction) =>
                transaction.type.includes('web_monetization_') &&
                transaction.state == 'Pending' ? (
                  // Can't disable a link :/
                  <div
                    key={transaction.id + transaction.state}
                    className='my-1 flex justify-between space-x-4 rounded-xl p-3 first:mt-0 last-of-type:mb-0'
                  >
                    <div className='flex w-7/12 items-center space-x-2'>
                      <Icon>schedule</Icon>
                      <div className='flex w-full flex-col space-y-1'>
                        <span className='truncate text-medium'>
                          {transaction.title}
                        </span>
                        <span className='text-xs text-weak'>
                          {transaction.formattedTime}
                        </span>
                      </div>
                    </div>
                    <div className='flex min-w-max flex-initial items-center space-x-2'>
                      <span
                        className={clsx(
                          'min-w-max font-medium',
                          transaction.type == 'web_monetization_outgoing'
                            ? 'text-error'
                            : 'text-medium'
                        )}
                      >
                        {transaction.subtotal}
                      </span>
                      <div className='h-6 w-6' />
                    </div>
                  </div>
                ) : (
                  <CardLink
                    preventScrollReset={!isMobile}
                    prefetch='none'
                    key={transaction.id + transaction.state}
                    to={route('/payments/:paymentId', {
                      paymentId: transaction.id
                    })}
                    className='justify-between space-x-4'
                  >
                    <div className='flex w-7/12 items-center space-x-2'>
                      {transaction.state == 'Pending' && <Icon>schedule</Icon>}
                      {transaction.state == 'Failed' && (
                        <Icon className='text-error'>exclamation</Icon>
                      )}
                      {transaction.state != 'Pending' &&
                        transaction.state != 'Failed' && (
                          <>
                            {transaction.destinationIdentityType ==
                              'wallet' && (
                              <>
                                {transaction.type != 'withdrawal' &&
                                  transaction.type != 'deposit' && (
                                    <InterledgerIcon />
                                  )}
                                {transaction.type == 'withdrawal' && (
                                  <Icon>south_west</Icon>
                                )}
                                {transaction.type == 'deposit' && (
                                  <Icon>north_east</Icon>
                                )}
                              </>
                            )}
                            {transaction.destinationIdentityType ==
                              'Twitter' && <TwitterIcon />}
                            {transaction.destinationIdentityType == 'Slack' && (
                              <SlackIcon />
                            )}
                            {transaction.type === 'card_transaction' && (
                              <Icon>credit_card</Icon>
                            )}
                          </>
                        )}
                      <div className='flex w-full flex-col space-y-1'>
                        <span className='truncate text-medium'>
                          {transaction.title}
                        </span>
                        <span className='text-xs text-weak'>
                          {transaction.formattedTime}
                        </span>
                      </div>
                    </div>
                    <div className='flex min-w-max flex-initial items-center space-x-2'>
                      <span
                        className={clsx(
                          'min-w-max font-medium',
                          transaction.type == 'sent' ||
                            transaction.type == 'web_monetization_outgoing' ||
                            (transaction.type == 'card_transaction' &&
                              transaction.state === 'Completed')
                            ? 'text-error'
                            : 'text-medium'
                        )}
                      >
                        {transaction.subtotal}
                      </span>
                      <Icon>navigate_next</Icon>
                    </div>
                  </CardLink>
                )
              )}
            </Card>
          ))}
      </GridColumn>
      <GridColumn sticky className='col-span-full lg:col-span-6 lg:col-start-7'>
        <Outlet />
      </GridColumn>
    </WalletGrid>
  )
}
