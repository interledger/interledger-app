import type { LoaderArgs } from '@remix-run/node'

import { Router, Grid } from '~/components'
import { json } from '@remix-run/node'
import { useLoaderData } from '@remix-run/react'
import { ListWallets } from '~/lib/wallet.server'
import { route } from 'routes-gen'

export async function loader({ request }: LoaderArgs) {
  const url = new URL(request.url)
  // let page = url.searchParams.get('page') || '1'
  let pageSize = url.searchParams.get('pageSize') || '50'
  const wallets = await ListWallets(request, {
    pageSize: parseInt(pageSize)
  })

  return json({
    wallets
  })
}

export default function Page() {
  const { wallets } = useLoaderData<typeof loader>()

  return (
    <Grid>
      <div className='col-span-full flex flex-col rounded-2xl bg-page p-4 pb-6'>
        <div className='sm:flex sm:items-center'>
          <div className='sm:flex-auto'>
            <h1 className='text-xl font-semibold text-gray-900'>Wallets</h1>
            <p className='mt-2 text-sm text-gray-700'>
              All the user wallets, and user details.
            </p>
          </div>
        </div>
        <div className='mt-8 flex flex-col'>
          <div className='-my-2 -mx-4 overflow-x-auto sm:-mx-6 lg:-mx-8'>
            <div className='inline-block min-w-full py-2 align-middle md:px-6 lg:px-8'>
              <div className='overflow-hidden ring-2 ring-base md:rounded-lg'>
                <table className='min-w-full divide-y divide-base'>
                  <thead className='bg-app'>
                    <tr>
                      <th
                        scope='col'
                        className='px-4 py-3.5 text-left text-sm font-medium text-strong'
                      >
                        ID
                      </th>
                      <th
                        scope='col'
                        className='px-4 py-3.5 text-left text-sm font-medium text-strong'
                      >
                        Name
                      </th>
                      <th
                        scope='col'
                        className='px-4 py-3.5 text-left text-sm  font-medium text-strong'
                      >
                        Email
                      </th>
                      <th
                        scope='col'
                        className='px-4 py-3.5 text-left text-sm font-medium text-strong'
                      >
                        Phone number
                      </th>
                      <th scope='col' className='relative py-3.5 px-4'>
                        <span className='sr-only'>Edit</span>
                      </th>
                    </tr>
                  </thead>
                  <tbody className='divide-y divide-gray-200 bg-white'>
                    {wallets.wallets.map((wallet) => (
                      <tr key={wallet.walletID}>
                        <td className='p-4 text-sm font-medium text-gray-900'>
                          {wallet.walletID}
                        </td>
                        <td className='p-4 text-sm font-medium text-gray-900'>
                          {wallet.walletName}
                        </td>
                        <td className='whitespace-nowrap p-4 text-sm text-gray-500'>
                          {wallet.users.length > 0 ? wallet.users[0].email : ''}
                        </td>
                        <td className='whitespace-nowrap p-4 text-sm text-gray-500'>
                          {wallet.users.length > 0
                            ? wallet.users[0].phoneNumber
                            : ''}
                        </td>
                        <td className='relative whitespace-nowrap p-4 text-right text-sm font-medium'>
                          <Router
                            to={route('/wallet/:id/profile', {
                              id: wallet.walletID
                            })}
                            className='text-primary'
                          >
                            View
                            <span className='sr-only'>, {wallet.walletID}</span>
                          </Router>
                        </td>
                      </tr>
                    ))}
                    <tr
                      className='items-center justify-between p-4'
                      aria-label='Pagination'
                    >
                      <td colSpan={2} className='p-4'>
                        <p className='text-sm text-weak'>
                          Showing <span className='font-medium'>1</span> to{' '}
                          <span className='font-medium'>
                            {wallets.wallets.length}
                          </span>{' '}
                          of{' '}
                          <span className='font-medium'>
                            {wallets.wallets.length > 10
                              ? 'unknown'
                              : wallets.wallets.length}
                          </span>{' '}
                          results
                        </p>
                      </td>
                      <td colSpan={3}>
                        {/*<div className='flex flex-1 justify-between pr-3 sm:justify-end'>*/}
                        {/*  {wallets.page?.page && wallets.page?.page > 1 && (*/}
                        {/*    <Router*/}
                        {/*      to={`/wallets?page=${wallets.page?.page - 1}`}*/}
                        {/*      className='relative inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50'*/}
                        {/*    >*/}
                        {/*      Previous*/}
                        {/*    </Router>*/}
                        {/*  )}*/}
                        {/*  {wallets.page?.hasNextPage && (*/}
                        {/*    <Router*/}
                        {/*      to={`/wallets?page=${wallets.page?.page + 1}`}*/}
                        {/*      className='relative ml-3 inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50'*/}
                        {/*    >*/}
                        {/*      Next*/}
                        {/*    </Router>*/}
                        {/*  )}*/}
                        {/*</div>*/}
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Grid>
  )
}
