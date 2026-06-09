import type { ActionFunctionArgs, LoaderFunctionArgs } from 'react-router'

import { data, href, useLoaderData, useSubmit } from 'react-router'
import { Grid, Icon, Router } from '~/components'
import { LinkedAccountReviewState } from '~/lib/types'
import {
  CompleteLinkedAccountReview,
  ListLinkedAccountReviews
} from '~/lib/wallet.server'

export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url)
  let pageSize = url.searchParams.get('pageSize') || '50'
  const { reviews } = await ListLinkedAccountReviews(request, {
    pageSize: parseInt(pageSize)
  })

  return data({
    reviews
  })
}

export default function Page() {
  const { reviews } = useLoaderData<typeof loader>()
  const submit = useSubmit()

  return (
    <Grid>
      <div className='col-span-full flex flex-col rounded-2xl bg-page p-4 pb-6'>
        <div className='sm:flex sm:items-center'>
          <div className='sm:flex-auto'>
            <h1 className='text-xl font-semibold text-gray-900'>
              Linked account reviews
            </h1>
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
                        Linked account ID
                      </th>
                      <th
                        scope='col'
                        className='px-4 py-3.5 text-left text-sm  font-medium text-strong'
                      >
                        Current State
                      </th>
                      <th
                        scope='col'
                        className='px-4 py-3.5 text-left text-sm  font-medium text-strong'
                      >
                        Wallet ID
                      </th>
                      <th
                        scope='col'
                        className='px-4 py-3.5 text-left text-sm  font-medium text-strong'
                      >
                        Wallet Name
                      </th>
                      <th
                        scope='col'
                        className='px-4 py-3.5 text-left text-sm  font-medium text-strong'
                      >
                        Mask
                      </th>
                      <th
                        scope='col'
                        className='px-4 py-3.5 text-left text-sm font-medium text-strong'
                      >
                        Action
                      </th>
                      <th scope='col' className='relative py-3.5 px-4'>
                        <span className='sr-only'>Edit</span>
                      </th>
                    </tr>
                  </thead>
                  <tbody className='divide-y divide-gray-200 bg-white'>
                    {reviews.map((review) => (
                      <tr key={review.id}>
                        <td className='p-4 text-sm font-medium text-gray-900'>
                          {review.linkedAccountID}
                        </td>
                        <td className='whitespace-nowrap p-4 text-sm text-gray-500'>
                          {review.state}
                        </td>
                        <td className='p-4 text-sm font-medium text-gray-900'>
                          {review.walletID}
                        </td>
                        <td className='p-4 text-sm font-medium text-gray-900'>
                          {review.walletName}
                        </td>
                        <td className='p-4 text-sm font-medium text-gray-900'>
                          {review.mask}
                        </td>
                        <td
                          className='whitespace-nowrap p-4 text-sm text-gray-500'
                          onClick={() => {
                            let formData = new FormData()
                            formData.append('reviewID', review.id)
                            formData.append(
                              'newState',
                              LinkedAccountReviewState.Verified
                            )
                            submit(formData, {
                              action: href('/reviews'),
                              method: 'POST'
                            })
                          }}
                        >
                          <span className='flex cursor-pointer items-center space-x-2 font-medium text-primary'>
                            <Icon>approval_delegation</Icon>
                            <span>Approve</span>
                          </span>
                        </td>
                        <td className='relative whitespace-nowrap p-4 text-right text-sm font-medium'>
                          <Router
                            to={href('/review/:id/details', {
                              id: review.id
                            })}
                            className='text-primary'
                          >
                            View
                            <span className='sr-only'>, {review.id}</span>
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
                          <span className='font-medium'>{reviews.length}</span>{' '}
                          of{' '}
                          <span className='font-medium'>
                            {reviews.length > 10 ? 'unknown' : reviews.length}
                          </span>{' '}
                          results
                        </p>
                      </td>
                      <td colSpan={3}></td>
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

export async function action({ request }: ActionFunctionArgs) {
  const form = await request.formData()
  const reviewID = form.get('reviewID') as string
  const newState = form.get('newState') as LinkedAccountReviewState
  const reason = form.get('reason') as string

  await CompleteLinkedAccountReview(request, reviewID, newState, reason)

  return null
}
