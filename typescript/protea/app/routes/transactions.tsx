import { Fragment } from 'react'
import type { LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { HomeShapes, Icon, Layouts, Router, WalletGrid } from '~/components'
import { requireUserSession } from '~/lib/kratos.server'
import { DateTime } from 'luxon'
import {
  httpMapping,
  isGrpcError,
  openPaymentsClient,
  StatusError
} from '~/lib/proto.server'
import { formatAmount } from '~/lib/wallet.server'

export async function loader({ request }: LoaderArgs) {
  await requireUserSession(request)
  const cookie = String(request.headers.get('cookie'))
  const url = new URL(request.url)

  const flowId = url.searchParams.get('flow')
  if (flowId) return redirect(`${route('/recovery/password')}?flow=${flowId}`)

  let pendingTransactionsResponse = openPaymentsClient
    .listPendingTransactions(
      { page: 1, pageSize: 20 },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)

  let transactionsResponse = openPaymentsClient
    .listTransactions(
      { page: 1, pageSize: 20 },
      {
        meta: {
          cookies: cookie || ''
        }
      }
    )
    .then((v) => v)
    .catch(StatusError)

  const responses = await Promise.all([
    pendingTransactionsResponse,
    transactionsResponse
  ])

  if (isGrpcError(responses[0])) {
    throw json({}, httpMapping(responses[0].code))
  }
  if (isGrpcError(responses[1])) {
    throw json({}, httpMapping(responses[1].code))
  }

  const pendingTransactions = responses[0].response.transactions.map((trx) => ({
    id: trx.id,
    icon: 'schedule',
    title: 'Sending',
    total: formatAmount(trx.amount),
    description:
      trx.type == 'outgoing' ? `to ${trx.destination}` : `from ${trx.source}`,
    date: DateTime.fromSeconds(parseInt(trx.timestamp?.seconds ?? '')).toFormat(
      'dd MMM yyyy'
    )
  }))

  const transactions = responses[1].response.transactions.map((trx) => ({
    id: trx.id,
    icon: trx.type == 'outgoing' ? 'north_east' : 'south_west',
    title: trx.type == 'outgoing' ? 'Sent' : 'Received',
    total: formatAmount(trx.amount),
    description:
      trx.type == 'outgoing' ? `to ${trx.destination}` : `from ${trx.source}`,
    date: DateTime.fromSeconds(parseInt(trx.timestamp?.seconds ?? '')).toFormat(
      'dd MMM yyyy'
    )
  }))

  return json({
    transactions: [...pendingTransactions, ...transactions]
  })
}

export const handle = {
  layout: Layouts.WalletLayout
}

export default function Page() {
  const { transactions } = useLoaderData<typeof loader>()
  return (
    <WalletGrid>
      <div className='col-span-full flex flex-col rounded-2xl bg-page p-4 pb-8 sm:col-span-6 sm:col-start-2 lg:col-start-4'>
        <div className='mt-2'>
          <HomeShapes />
        </div>
        <h1 className='mt-6 font-display text-2xl font-medium'>Transactions</h1>
        {transactions.length == 0 && (
          <div className='mt-4 flex flex-col space-y-4'>
            <span className='text-sm text-medium'>
              Your payment activity will appear here once you start using your
              payment pointer.
            </span>
            <Router
              to={route('/pay')}
              className='text-sm font-medium text-primary'
            >
              Send or receive payments now
            </Router>
          </div>
        )}
        {transactions.map((transaction, index) => (
          <Fragment key={transaction.id}>
            {(index == 0 ||
              transaction.date != transactions[index - 1].date) && (
              <span className='mt-6 text-xs text-medium'>
                {transaction.date}
              </span>
            )}
            <div className='mt-2 flex w-full justify-between'>
              <div className='mx-1 flex space-x-1'>
                <Icon className='text-medium'>{transaction.icon}</Icon>
                <div className='flex flex-col space-y-2'>
                  <span className='text-medium'>{transaction.title}</span>
                  <span className='text-xs text-medium'>
                    {transaction.description}
                  </span>
                </div>
              </div>
              <span className='font-medium'>{transaction.total}</span>
            </div>
          </Fragment>
        ))}
      </div>
    </WalletGrid>
  )
}
