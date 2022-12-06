import { Fragment } from 'react'
import type { LoaderArgs } from '@remix-run/node'
import { json, redirect } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { HomeShapes, Icon, Layouts, Router, WalletGrid } from '~/components'
import { getPendingTransactions, getTransactions } from '~/lib/wallet.server'

export async function loader({ request }: LoaderArgs) {
  const url = new URL(request.url)

  const flowId = url.searchParams.get('flow')
  if (flowId) return redirect(`${route('/recovery/password')}?flow=${flowId}`)

  const [pendingTransactions, transactions] = await Promise.all([
    getPendingTransactions(request),
    getTransactions(request)
  ])

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
            <Router
              to={route('/transaction/:type/:transactionId', {
                type: transaction.type,
                transactionId: transaction.id
              })}
              className='mt-2 flex w-full justify-between'
            >
              <div className='flex space-x-1'>
                <Icon className='mt-0.5 text-medium'>{transaction.icon}</Icon>
                <div className='flex flex-col space-y-2'>
                  <span className='text-medium'>{transaction.title}</span>
                  <span className='text-xs text-medium'>
                    {transaction.time}
                  </span>
                </div>
              </div>
              <span className='font-medium'>{transaction.total}</span>
            </Router>
          </Fragment>
        ))}
      </div>
    </WalletGrid>
  )
}
