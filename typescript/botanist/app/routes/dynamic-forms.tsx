import type { LoaderArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Link, useLoaderData } from '@remix-run/react'
import { route } from 'routes-gen'
import { Grid, Icon } from '~/components'
import { ListDynamicFormCounts } from '~/lib/wallet.server'
import React from 'react'

export async function loader({ request }: LoaderArgs) {
  const url = new URL(request.url)
  let pageSize = url.searchParams.get('pageSize') || '50'
  const formCounts = await ListDynamicFormCounts(request, {
    pageSize: parseInt(pageSize)
  })

  return json({
    formCounts
  })
}

export default function Page() {
  const { formCounts } = useLoaderData<typeof loader>()

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
                    </tr>
                  </thead>
                  <tbody className='divide-y divide-gray-200 bg-white'>
                    {formCounts.map((formCount) => (
                      <tr key={formCount.formId}>
                        <td className='p-4 text-sm font-medium text-gray-900'>
                          {formCount.formId}
                        </td>
                        <td className='whitespace-nowrap p-4 text-sm text-gray-500'>
                          {formCount.formCount}
                        </td>
                        <td className='whitespace-nowrap p-4 text-sm text-gray-500'>
                          <Link
                            to={route('/dynamic-forms/:id.csv', {
                              id: formCount.formId
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
                            {formCounts.length}
                          </span>{' '}
                          of{' '}
                          <span className='font-medium'>
                            {formCounts.length > 10
                              ? 'unknown'
                              : formCounts.length}
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
