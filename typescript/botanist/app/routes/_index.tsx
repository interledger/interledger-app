import type { LoaderFunctionArgs } from 'react-router'

import { Grid } from '~/components'
import { data } from 'react-router'
import { useLoaderData } from 'react-router'
import { ListWallets } from '~/lib/wallet.server'
import { grpcClient } from '~/lib/proto.server'

export async function loader({ request }: LoaderFunctionArgs) {
  try {
    const wallets = await ListWallets(request, {
      pageSize: 10000
    })

    const signups = await grpcClient.listWaitlistSignups(
      {},
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )

    const stats = await grpcClient.getUserStats(
      {},
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )

    const transactionStats = await grpcClient.getTransactionStats(
      {},
      {
        meta: {
          cookies: String(request.headers.get('cookie')) || ''
        }
      }
    )

    return data({
      error: '',
      wallets,
      signups: signups.response.signups,
      userCount: stats.response.totalUsers,
      usersThisYear: stats.response.usersThisYear,
      quarterlyUsers: stats.response.quarterlyUsers,
      transactionCount: transactionStats.response.totalTransactions,
      transactionsThisyear: transactionStats.response.transactionsThisYear,
      quarterlyTransactions: transactionStats.response.quarterlyTransactions,
      transactionsByType: transactionStats.response.typeTransactions
    })
  } catch (e) {
    console.log('Failed to retrieve wallets: ', e)
    return data({
      error: 'there was an error retrieving wallets',
      wallets: { wallets: [] },
      signups: [],
      userCount: 0,
      usersThisYear: 0,
      quarterlyUsers: [],
      transactionCount: 0,
      transactionsThisyear: 0,
      quarterlyTransactions: [],
      transactionsByType: []
    })
  }
}

export default function Page() {
  const {
    wallets,
    signups,
    userCount,
    usersThisYear,
    quarterlyUsers,
    transactionCount,
    transactionsThisyear,
    quarterlyTransactions,
    transactionsByType,
    error
  } = useLoaderData<typeof loader>()

  return (
    <Grid>
      {!error ? (
        <>
          <div className='col-span-2 flex flex-col rounded-2xl bg-page p-4'>
            <h2 className='font-display text-lg font-medium'>Total wallets</h2>
            <h1 className='mt-2 text-3xl font-medium'>
              {wallets.wallets.length}
            </h1>
          </div>
          <div className='col-span-2 flex flex-col rounded-2xl bg-page p-4'>
            <h2 className='font-display text-lg font-medium'>
              People on waitlist
            </h2>
            <h1 className='mt-2 text-3xl font-medium'>{signups.length}</h1>
          </div>
          <div className='col-span-2 col-start-1 flex flex-col rounded-2xl bg-page p-4'>
            <h2 className='font-display text-lg font-medium'>
              % waitlist can sign up
            </h2>
            <h1 className='mt-2 text-3xl font-medium'>
              {(
                (signups.filter((user) => user.canSignup).length /
                  signups.length) *
                100
              ).toFixed(2)}
            </h1>
          </div>
          <div className='col-span-2 flex flex-col rounded-2xl bg-page p-4'>
            <h2 className='font-display text-lg font-medium'>
              % waitlist beta opt in
            </h2>
            <h1 className='mt-2 text-3xl font-medium'>
              {(
                (signups.filter((user) => user.betaOptIn).length /
                  signups.length) *
                100
              ).toFixed(2)}
            </h1>
          </div>
          <div className='col-span-2 col-start-1 flex flex-col rounded-2xl bg-page p-4'>
            <h2 className='font-display text-lg font-medium'>Total users</h2>
            <h1 className='mt-2 text-3xl font-medium'>{userCount}</h1>
          </div>
          <div className='col-span-2 flex flex-col rounded-2xl bg-page p-4'>
            <h2 className='font-display text-lg font-medium'>
              Total users this year
            </h2>
            <h1 className='mt-2 text-3xl font-medium'>{usersThisYear}</h1>
          </div>
          <div className='col-span-full flex flex-col rounded-2xl bg-page p-4 pb-6'>
            <h2 className='font-display text-lg font-medium'>
              Users created this year
            </h2>

            <div className='mt-4 overflow-hidden rounded-lg shadow ring-1 ring-black ring-opacity-5'>
              <table className='min-w-full divide-y divide-gray-300'>
                <thead className='bg-gray-50'>
                  <tr>
                    <th className='px-4 py-3 text-left text-sm font-semibold text-gray-900'>
                      Quarter
                    </th>
                    <th className='px-4 py-3 text-left text-sm font-semibold text-gray-900'>
                      Users
                    </th>
                  </tr>
                </thead>

                <tbody className='divide-y divide-gray-200 bg-white'>
                  {quarterlyUsers.map((quarter) => (
                    <tr key={quarter.quarter}>
                      <td className='px-4 py-3 text-sm font-medium text-gray-900'>
                        Q{quarter.quarter}
                      </td>
                      <td className='px-4 py-3 text-sm text-gray-500'>
                        {quarter.count}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
          <div className='col-span-2 col-start-1 flex flex-col rounded-2xl bg-page p-4'>
            <h2 className='font-display text-lg font-medium'>
              Total Transactions
            </h2>
            <h1 className='mt-2 text-3xl font-medium'>{transactionCount}</h1>
          </div>
          <div className='col-span-2 flex flex-col rounded-2xl bg-page p-4'>
            <h2 className='font-display text-lg font-medium'>
              Total Transactions this year
            </h2>
            <h1 className='mt-2 text-3xl font-medium'>
              {transactionsThisyear}
            </h1>
          </div>
          <div className='col-span-full flex flex-col rounded-2xl bg-page p-4 pb-6'>
            <h2 className='font-display text-lg font-medium'>
              Transactions created this year
            </h2>

            <div className='mt-4 overflow-hidden rounded-lg shadow ring-1 ring-black ring-opacity-5'>
              <table className='min-w-full divide-y divide-gray-300'>
                <thead className='bg-gray-50'>
                  <tr>
                    <th className='px-4 py-3 text-left text-sm font-semibold text-gray-900'>
                      Quarter
                    </th>
                    <th className='px-4 py-3 text-left text-sm font-semibold text-gray-900'>
                      Transactions
                    </th>
                  </tr>
                </thead>

                <tbody className='divide-y divide-gray-200 bg-white'>
                  {quarterlyTransactions.map((quarter) => (
                    <tr key={quarter.quarter}>
                      <td className='px-4 py-3 text-sm font-medium text-gray-900'>
                        Q{quarter.quarter}
                      </td>
                      <td className='px-4 py-3 text-sm text-gray-500'>
                        {quarter.count}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
          <div className='col-span-full flex flex-col rounded-2xl bg-page p-4 pb-6'>
            <h2 className='font-display text-lg font-medium'>
              Transactions by type
            </h2>

            <div className='mt-4 overflow-hidden rounded-lg shadow ring-1 ring-black ring-opacity-5'>
              <table className='min-w-full divide-y divide-gray-300'>
                <thead className='bg-gray-50'>
                  <tr>
                    <th className='px-4 py-3 text-left text-sm font-semibold text-gray-900'>
                      Type
                    </th>
                    <th className='px-4 py-3 text-left text-sm font-semibold text-gray-900'>
                      Transactions
                    </th>
                  </tr>
                </thead>

                <tbody className='divide-y divide-gray-200 bg-white'>
                  {transactionsByType.map((type) => (
                    <tr key={type.type}>
                      <td className='px-4 py-3 text-sm font-medium text-gray-900'>
                        {type.type}
                      </td>
                      <td className='px-4 py-3 text-sm text-gray-500'>
                        {type.count}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </>
      ) : (
        <div className='col-span-4 flex flex-col rounded-2xl bg-page p-4'>
          <h2 className='font-display text-lg font-medium'>
            There was an error retrieving wallets
          </h2>
        </div>
      )}
    </Grid>
  )
}
