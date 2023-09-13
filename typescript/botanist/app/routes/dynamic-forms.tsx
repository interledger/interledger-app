import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Link, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Grid, Icon, Router } from '~/components'
import React from 'react'
import { ListFormSubmissionCounts } from '~/lib/wallet.server'

export async function loader({ request }: LoaderArgs) {
  const url = new URL(request.url)
  let pageSize = url.searchParams.get('pageSize') || '50'
  const subCounts = await ListFormSubmissionCounts(request, {
    pageSize: parseInt(pageSize)
  })

  return json({
    subCounts
  })
}

export default function Page() {
  const { subCounts } = useLoaderData<typeof loader>()

  return (
    <Grid>
      <div className='col-span-full flex flex-col rounded-2xl bg-page p-4 pb-6'>
        <div className='sm:flex sm:items-center'>
          <div className='sm:flex-auto'>
            <h1 className='text-xl font-semibold text-gray-900'>Forms</h1>
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
                        Form ID
                      </th>
                      <th
                        scope='col'
                        className='px-4 py-3.5 text-left text-sm  font-medium text-strong'
                      >
                        Responses
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
                    {subCounts.map((subCount) => (
                      <tr key={subCount.formId}>
                        <td className='p-4 text-sm font-medium text-gray-900'>
                          {subCount.formId}
                        </td>
                        <td className='whitespace-nowrap p-4 text-sm text-gray-500'>
                          {subCount.submissionCount}
                        </td>
                        <td className='whitespace-nowrap p-4 text-sm text-gray-500'>
                          <Link
                            to={route('/dynamic-form/:id.csv', {
                              id: subCount.formId
                            })}
                            className='text-primary'
                            reloadDocument
                          >
                            <span className='flex cursor-pointer items-center space-x-2 font-medium text-primary'>
                              <Icon>download</Icon>
                              <span>Download</span>
                            </span>
                          </Link>
                        </td>
                        <td className='relative whitespace-nowrap p-4 text-right text-sm font-medium'>
                          <Router
                            to={route('/dynamic-form/:id/submissions', {
                              id: subCount.formId
                            })}
                            className='text-primary'
                          >
                            View
                            <span className='sr-only'>, {subCount.formId}</span>
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
                            {subCounts.length}
                          </span>{' '}
                          of{' '}
                          <span className='font-medium'>
                            {subCounts.length > 10
                              ? 'unknown'
                              : subCounts.length}
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
