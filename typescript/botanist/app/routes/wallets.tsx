import type { LoaderFunctionArgs } from 'react-router'

import { Router, Grid, TextField } from '~/components'
import { data, href, Form, useLoaderData, useNavigation } from 'react-router'
import { ListWallets } from '~/lib/wallet.server'

const FILTER_FIELDS = [
  { name: 'firstName', label: 'First name' },
  { name: 'lastName', label: 'Last name' },
  { name: 'walletAddress', label: 'Wallet account' },
  { name: 'email', label: 'Email' },
  { name: 'phoneNumber', label: 'Phone number' },
  { name: 'providerId', label: 'Provider ID' }
] as const
type FilterField = (typeof FILTER_FIELDS)[number]['name']

export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url)
  const pageSize = url.searchParams.get('pageSize') || '50'
  const pageToken = url.searchParams.get('pageToken') || ''

  const filters = Object.fromEntries(
    FILTER_FIELDS.map(({ name }) => [
      name,
      (url.searchParams.get(name) || '').trim()
    ])
  ) as Record<FilterField, string>

  const hasFilter = Object.values(filters).some((v) => v !== '')

  const wallets = await ListWallets(
    request,
    {
      pageSize: parseInt(pageSize),
      pageToken: pageToken || undefined
    },
    hasFilter ? filters : undefined
  )

  return data({
    wallets,
    pageSize,
    filters,
    hasFilter
  })
}

export default function Page() {
  const { wallets, pageSize, filters, hasFilter } =
    useLoaderData<typeof loader>()
  const navigation = useNavigation()
  const isSearching = navigation.state === 'loading'

  const nextPageParams = new URLSearchParams()
  if (wallets.nextPageToken) {
    nextPageParams.set('pageToken', wallets.nextPageToken)
    nextPageParams.set('pageSize', pageSize)
    for (const { name } of FILTER_FIELDS) {
      if (filters[name]) nextPageParams.set(name, filters[name])
    }
  }

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

        <div className='mt-4'>
          <Form method='get' action='/wallets'>
            <input type='hidden' name='pageSize' value={pageSize} />
            <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3'>
              {FILTER_FIELDS.map(({ name, label }) => (
                <TextField
                  key={`${name}-${filters[name]}`}
                  id={`wallet-search-${name}`}
                  name={name}
                  type='search'
                  label={label}
                  defaultValue={filters[name]}
                />
              ))}
            </div>
            <div className='mt-2 flex items-center gap-3'>
              <button
                type='submit'
                className='rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white'
              >
                Search
              </button>
              <Router to='/wallets' className='text-sm text-primary'>
                Clear
              </Router>
            </div>
          </Form>
          {hasFilter && (
            <p className='mt-2 text-xs text-medium'>
              {wallets.wallets.length} result
              {wallets.wallets.length !== 1 ? 's' : ''}
            </p>
          )}
        </div>

        <div className='mt-8 flex flex-col'>
          <div className='-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8'>
            <div className='inline-block min-w-full py-2 align-middle md:px-6 lg:px-8'>
              <div className='overflow-hidden ring-2 ring-base md:rounded-lg'>
                <table
                  className={`min-w-full divide-y divide-base${
                    isSearching ? 'opacity-50' : ''
                  }`}
                >
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
                        className='px-4 py-3.5 text-left text-sm font-medium text-strong'
                      >
                        Email
                      </th>
                      <th
                        scope='col'
                        className='px-4 py-3.5 text-left text-sm font-medium text-strong'
                      >
                        Phone number
                      </th>
                      <th scope='col' className='relative px-4 py-3.5'>
                        <span className='sr-only'>Edit</span>
                      </th>
                    </tr>
                  </thead>
                  <tbody className='divide-y divide-gray-200 bg-white'>
                    {wallets.wallets.length === 0 && hasFilter && (
                      <tr>
                        <td
                          colSpan={5}
                          className='p-4 text-center text-sm text-weak'
                        >
                          No users found
                        </td>
                      </tr>
                    )}
                    {wallets.wallets.map((wallet) => (
                      <tr key={wallet.walletID}>
                        <td className='p-4 text-sm font-medium text-gray-900'>
                          {wallet.walletID}
                        </td>
                        <td className='p-4 text-sm font-medium text-gray-900'>
                          {wallet.walletName}
                        </td>
                        <td className='whitespace-nowrap p-4 text-sm text-gray-500'>
                          {wallet.users[0]?.email ?? ''}
                        </td>
                        <td className='whitespace-nowrap p-4 text-sm text-gray-500'>
                          {wallet.users[0]?.phoneNumber ?? ''}
                        </td>
                        <td className='relative whitespace-nowrap p-4 text-right text-sm font-medium'>
                          <Router
                            to={href('/wallet/:id/profile', {
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
                          Showing{' '}
                          <span className='font-medium'>
                            {wallets.wallets.length === 0 ? 0 : 1}
                          </span>{' '}
                          to{' '}
                          <span className='font-medium'>
                            {wallets.wallets.length}
                          </span>{' '}
                          results
                        </p>
                      </td>
                      <td colSpan={3}>
                        <div className='flex flex-1 justify-between pr-3 sm:justify-end'>
                          {wallets.nextPageToken && (
                            <Router
                              to={`/wallets?${nextPageParams.toString()}`}
                              className='relative ml-3 inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50'
                            >
                              Next
                            </Router>
                          )}
                        </div>
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
