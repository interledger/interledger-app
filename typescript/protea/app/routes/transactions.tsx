import type { LoaderArgs, MetaFunction } from '@remix-run/node'
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
  AnimatedSchedule,
  Card,
  CardContent,
  CardHeader,
  CardIcon,
  CardLink,
  CardTitle,
  Chip,
  ChipColor,
  Fab,
  FynbosIcon,
  GridColumn,
  Icon,
  Layouts,
  Router,
  TwitterIcon,
  WalletGrid
} from '~/components'
import { Label } from '~/components/Label'
import type { Transaction } from '~/generated/protobuf-ts/backend/v1/backend'
import { getKycStatus, getTransactionsWithPending } from '~/lib/wallet.server'
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

export async function loader({ request }: LoaderArgs) {
  const url = new URL(request.url)
  const pages = parseInt(url.searchParams.get('pages') || '1')

  const kycStatus = await getKycStatus(request)
  let pageInfo = {
    // pageToken is only set by fetcher so initial page loads this should be blank
    pageToken: url.searchParams.get('pageToken') || '',
    pageSize: 30
  }

  let allTransactions: Transaction[] = []

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
      title: 'Transactions',
      actions: [{ type: 'shapes' }]
    },
    fab: Fab.Pay
  }
}

export const meta: MetaFunction = () => {
  return {
    title: 'Transactions'
  }
}

export default function Page() {
  const initialPage = useLoaderData<typeof loader>()
  let [, setSearchParams] = useSearchParams()
  const fetcher = useFetcher()
  const [transactions, setTransactions] = useState(initialPage.transactions)
  const [nextPageToken, setNextPageToken] = useState<string>(
    initialPage.nextPageToken
  )
  const [scrollPosition, setScrollPosition] = useState(0)
  const [clientHeight, setClientHeight] = useState(0)
  const [height, setHeight] = useState(null)
  const [shouldFetch, setShouldFetch] = useState(true)

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
      fetcher.load(`/transactions?pageToken=${nextPageToken}`)
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

  const location = useLocation()
  const pathSegments = location.pathname.split('/').filter(Boolean)

  return (
    <WalletGrid ref={divHeight}>
      <GridColumn
        hideOnMobile={pathSegments[pathSegments.length - 1] !== 'transactions'}
        className='col-span-full lg:col-span-5'
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
        {initialPage.kycStatus == KycStatus.Verified &&
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
              {transactionGroup.map((transaction) => (
                <CardLink
                  preventScrollReset
                  prefetch='none'
                  key={transaction.id}
                  to={route('/transactions/:transactionId', {
                    transactionId: transaction.id
                  })}
                  className='justify-between space-x-4'
                >
                  <div className='flex w-7/12 items-center space-x-2'>
                    {transaction.state == 'Pending' && <AnimatedSchedule />}
                    {transaction.state != 'Pending' &&
                      transaction.destinationIdentityType == 'wallet' && (
                        <FynbosIcon />
                      )}
                    {transaction.state != 'Pending' &&
                      transaction.destinationIdentityType == 'twitter' && (
                        <TwitterIcon />
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
                        transaction.type.includes('outgoing')
                          ? 'text-error'
                          : 'text-medium'
                      )}
                    >
                      {transaction.type.includes('outgoing') && '- '}
                      {transaction.formattedAmount}
                    </span>
                    <Icon>navigate_next</Icon>
                  </div>
                </CardLink>
              ))}
            </Card>
          ))}
      </GridColumn>
      <GridColumn sticky className='col-span-full lg:col-span-7 lg:col-start-6'>
        <Outlet />
      </GridColumn>
    </WalletGrid>
  )
}
