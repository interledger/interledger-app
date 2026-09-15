import type { LoaderFunctionArgs } from 'react-router'

import { Router, Grid, TextField } from '~/components'
import { data, href, Form, useLoaderData, useNavigation } from 'react-router'
import { ListWallets } from '~/lib/wallet.server'

const FILTER_FIELDS = [
  { name: 'firstName', label: 'First name (KYC)' },
  { name: 'lastName', label: 'Last name (KYC)' },
  { name: 'walletAddress', label: 'Wallet account' },
  { name: 'email', label: 'Email' },
  { name: 'phoneNumber', label: 'Phone number' },
  { name: 'providerId', label: 'Provider ID' }
] as const
type FilterField = (typeof FILTER_FIELDS)[number]['name']

// Empty/missing values render as a clear placeholder rather than a blank cell.
const EMPTY_PLACEHOLDER = '—'
const orDash = (value?: string) =>
  value && value.trim() !== '' ? value : EMPTY_PLACEHOLDER

async function countWallets(
  request: Request,
  pageSize: number,
  filters: Record<FilterField, string>
) {
  const firstPage = await ListWallets(
    request,
    { pageSize: pageSize },
    Object.values(filters).some((v) => v !== '') ? filters : undefined
  )

  let totalResults = firstPage.wallets.length
  let nextPageToken = firstPage.nextPageToken
  const pageTokens = ['']

  while (nextPageToken) {
    pageTokens.push(nextPageToken)

    const nextPage = await ListWallets(
      request,
      { pageSize: pageSize, pageToken: nextPageToken },
      Object.values(filters).some((v) => v !== '') ? filters : undefined
    )

    totalResults += nextPage.wallets.length
    nextPageToken = nextPage.nextPageToken
  }

  return { totalResults, pageTokens }
}

export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url)
  const pageSize = url.searchParams.get('pageSize') || '50'
  const pageToken = url.searchParams.get('pageToken') || ''
  const currentPage = Math.max(
    1,
    Number(url.searchParams.get('page') || (pageToken ? '2' : '1'))
  )

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

  const { totalResults, pageTokens } = await countWallets(
    request,
    parseInt(pageSize),
    filters
  )

  return data({
    wallets,
    pageSize,
    filters,
    hasFilter,
    totalResults,
    currentPage,
    pageTokens
  })
}

export default function Page() {
  const {
    wallets,
    pageSize,
    filters,
    hasFilter,
    totalResults,
    pageTokens,
    currentPage
  } = useLoaderData<typeof loader>()
  const navigation = useNavigation()
  const isSearching = navigation.state === 'loading'

  const pageParams = (pageNumber: number) => {
    const params = new URLSearchParams()
    const token = pageTokens[pageNumber - 1]

    params.set('page', pageNumber.toString())
    params.set('pageSize', pageSize.toString())

    if (token) params.set('pageToken', token)
    for (const { name } of FILTER_FIELDS) {
      if (filters[name]) params.set(name, filters[name])
    }

    return params
  }

  const pageCount = Math.max(1, Math.ceil(totalResults / parseInt(pageSize)))
  const previousPageParams = pageParams(Math.max(1, currentPage - 1))
  const nextPageParams = pageParams(currentPage + 1)

  const FIRST_PAGE = 1
  const VISIBLE_PAGES_AT_START = 3
  const VISIBLE_PAGES_AT_END = 3
  const PAGE_RANGE_THRESHOLD = 7
  const ADJACENT_PAGE_COUNT = 1

  const pageItems: Array<number | 'ellipsis'> =
    pageCount <= PAGE_RANGE_THRESHOLD
      ? Array.from({ length: pageCount }, (_, index) => index + FIRST_PAGE)
      : currentPage <= VISIBLE_PAGES_AT_START
        ? [
            ...Array.from(
              { length: VISIBLE_PAGES_AT_START },
              (_, index) => index + FIRST_PAGE
            ),
            'ellipsis',
            pageCount
          ]
        : currentPage >= pageCount - (VISIBLE_PAGES_AT_END - 1)
          ? [
              FIRST_PAGE,
              'ellipsis',
              ...Array.from(
                { length: VISIBLE_PAGES_AT_END },
                (_, index) => pageCount - (VISIBLE_PAGES_AT_END - 1) + index
              )
            ]
          : [
              FIRST_PAGE,
              'ellipsis',
              currentPage - ADJACENT_PAGE_COUNT,
              currentPage,
              currentPage + ADJACENT_PAGE_COUNT,
              'ellipsis',
              pageCount
            ]

  const hasPreviousPage = currentPage > 1
  const currentPageStart = (currentPage - 1) * parseInt(pageSize) + 1
  const currentPageEnd = Math.min(
    currentPageStart + wallets.wallets.length - 1,
    totalResults
  )

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
                        Internal ID
                      </th>
                      <th
                        scope='col'
                        className='px-4 py-3.5 text-left text-sm font-medium text-strong'
                      >
                        First name (KYC)
                      </th>
                      <th
                        scope='col'
                        className='px-4 py-3.5 text-left text-sm font-medium text-strong'
                      >
                        Last name (KYC)
                      </th>
                      <th
                        scope='col'
                        className='px-4 py-3.5 text-left text-sm font-medium text-strong'
                      >
                        Wallet name
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
                          colSpan={7}
                          className='p-4 text-center text-sm text-weak'
                        >
                          No users found
                        </td>
                      </tr>
                    )}
                    {wallets.wallets.map((wallet) => (
                      <tr key={wallet.walletID}>
                        <td className='p-4 text-sm font-medium text-gray-900'>
                          {orDash(wallet.walletID)}
                        </td>
                        <td className='whitespace-nowrap p-4 text-sm text-gray-500'>
                          {orDash(wallet.kycFirstName)}
                        </td>
                        <td className='whitespace-nowrap p-4 text-sm text-gray-500'>
                          {orDash(wallet.kycLastName)}
                        </td>
                        <td className='p-4 text-sm font-medium text-gray-900'>
                          {orDash(wallet.walletName)}
                        </td>
                        <td className='whitespace-nowrap p-4 text-sm text-gray-500'>
                          {orDash(wallet.users[0]?.email)}
                        </td>
                        <td className='whitespace-nowrap p-4 text-sm text-gray-500'>
                          {orDash(wallet.users[0]?.phoneNumber)}
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
                      <td colSpan={4} className='p-4'>
                        <p className='text-sm text-weak'>
                          Showing{' '}
                          <span className='font-medium text-[#14213d]'>
                            {wallets.wallets.length === 0
                              ? 0
                              : currentPageStart}
                          </span>{' '}
                          to{' '}
                          <span className='font-medium text-[#14213d]'>
                            {wallets.wallets.length === 0 ? 0 : currentPageEnd}
                          </span>{' '}
                          of{' '}
                          <span className='font-medium text-[#14213d]'>
                            {totalResults}
                          </span>{' '}
                          results
                        </p>
                      </td>
                      <td colSpan={3}>
                        <div className='flex flex-1 flex-wrap items-center justify-between gap-3 pr-3 sm:justify-end'>
                          {hasPreviousPage ? (
                            <Router
                              to={`/wallets?${previousPageParams.toString()}`}
                              className='relative inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50'
                            >
                              Previous
                            </Router>
                          ) : (
                            <span className='relative inline-flex cursor-not-allowed items-center rounded-md border border-gray-200 bg-gray-50 px-4 py-2 text-sm font-medium text-gray-400'>
                              Previous
                            </span>
                          )}

                          <div className='flex items-center gap-1'>
                            {pageItems.map((item, index) => {
                              if (item === 'ellipsis') {
                                return (
                                  <span
                                    key={`ellipsis-${index}`}
                                    aria-hidden='true'
                                    className='min-w-9 inline-flex h-9 items-center justify-center px-2 text-sm text-gray-500'
                                  >
                                    ...
                                  </span>
                                )
                              }

                              return item === currentPage ? (
                                <span
                                  key={item}
                                  aria-current='page'
                                  className='min-w-9 inline-flex h-9 items-center justify-center rounded-md bg-[#14213d] px-3 text-sm font-medium text-white'
                                >
                                  {item}
                                </span>
                              ) : (
                                <Router
                                  key={item}
                                  to={`/wallets?${pageParams(item).toString()}`}
                                  className='min-w-9 inline-flex h-9 items-center justify-center rounded-md border border-gray-300 bg-white px-3 text-sm font-medium text-gray-700 hover:bg-gray-50'
                                >
                                  {item}
                                </Router>
                              )
                            })}
                          </div>

                          {wallets.nextPageToken ? (
                            <Router
                              to={`/wallets?${nextPageParams.toString()}`}
                              className='relative inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50'
                            >
                              Next
                            </Router>
                          ) : (
                            <span className='relative inline-flex cursor-not-allowed items-center rounded-md border border-gray-200 bg-gray-50 px-4 py-2 text-sm font-medium text-gray-400'>
                              Next
                            </span>
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
