import type { LoaderFunctionArgs } from 'react-router'

import { data } from 'react-router'
import { useLoaderData } from 'react-router'
import { GetWalletAudits } from '~/lib/wallet.server'
import { DateTime } from 'luxon'

export async function loader({ request, params }: LoaderFunctionArgs) {
  const audit = await GetWalletAudits(request, params.id as string)
  return data({
    audit: audit.operations.map((op) => ({
      ...op,
      date: DateTime.fromSeconds(
        parseInt(op.timestamp?.seconds ?? '')
      ).toLocaleString(DateTime.DATETIME_FULL)
    }))
  })
}

export default function Page() {
  const { audit } = useLoaderData<typeof loader>()

  return (
    <div className='col-span-full flex flex-col rounded-2xl bg-page p-4'>
      <div className='sm:flex sm:items-center'>
        <div className='sm:flex-auto'>
          <h1 className='text-xl font-semibold text-gray-900'>Audit log</h1>
          <p className='mt-2 text-sm text-gray-700'>
            All the admin user interactions for this wallet.
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
                      Admin User
                    </th>
                    <th
                      scope='col'
                      className='px-4 py-3.5 text-left text-sm  font-medium text-strong'
                    >
                      Operation
                    </th>
                    <th
                      scope='col'
                      className='px-4 py-3.5 text-left text-sm font-medium text-strong'
                    >
                      Parameters
                    </th>
                    <th
                      scope='col'
                      className='px-4 py-3.5 text-left text-sm font-medium text-strong'
                    >
                      Timestamp
                    </th>
                  </tr>
                </thead>
                <tbody className='divide-y divide-gray-200 bg-white'>
                  {audit.map((operation) => (
                    <tr key={operation.walletID}>
                      <td className='whitespace-nowrap p-4 text-sm text-gray-500'>
                        {operation.adminUser}
                      </td>
                      <td className='whitespace-nowrap p-4 text-sm text-gray-500'>
                        {operation.operation}
                      </td>
                      <td className='whitespace-nowrap p-4 text-sm text-gray-500'>
                        {operation.parameters}
                      </td>
                      <td className='whitespace-nowrap p-4 text-sm text-gray-500'>
                        {operation.date}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
