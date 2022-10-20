import type { LoaderArgs } from '@remix-run/node'

import {
  WalletGrid
} from '~/components'


export async function loader({ request }: LoaderArgs) {
  return null
}

export default function Page() {
  return (
    <WalletGrid>
      <div className='col-span-full flex flex-col rounded-2xl bg-page p-4 pb-8 sm:col-span-12 sm:col-start-2 lg:col-start-1'>
          <div className="px-4 sm:px-6 lg:px-8">
              <div className="sm:flex sm:items-center">
                  <div className="sm:flex-auto">
                      <h1 className="text-xl font-semibold text-gray-900">Waitlist</h1>
                  </div>
              </div>
              <div className="mt-8 flex flex-col">
                  <div className="-my-2 -mx-4 sm:-mx-6 lg:-mx-8">
                      <div className="inline-block min-w-full py-2 align-middle">
                          <div className="shadow-sm ring-1 ring-black ring-opacity-5">
                              <table className="min-w-full border-separate" style={{ borderSpacing: 0 }}>
                                  <thead className="bg-gray-50">
                                  <tr>
                                      <th
                                          scope="col"
                                          className="sticky top-0 z-10 border-b border-gray-300 bg-gray-50 bg-opacity-75 py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 backdrop-blur backdrop-filter sm:pl-6 lg:pl-8"
                                      >
                                          Name
                                      </th>
                                      <th
                                          scope="col"
                                          className="sticky top-0 z-10 hidden border-b border-gray-300 bg-gray-50 bg-opacity-75 px-3 py-3.5 text-left text-sm font-semibold text-gray-900 backdrop-blur backdrop-filter sm:table-cell"
                                      >
                                          Email
                                      </th>
                                      <th
                                          scope="col"
                                          className="sticky top-0 z-10 hidden border-b border-gray-300 bg-gray-50 bg-opacity-75 px-3 py-3.5 text-left text-sm font-semibold text-gray-900 backdrop-blur backdrop-filter lg:table-cell"
                                      >
                                          Beta Opt In
                                      </th>
                                  </tr>
                                  </thead>
                                  <tbody className="bg-white">
                                  <tr>
                                      <td
                                          className='whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6 lg:pl-8'
                                      >
                                          Matthew de Haast
                                      </td>
                                      <td
                                          className='whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6 lg:pl-8'
                                      >
                                          matt@fynbos.dev
                                      </td>
                                      <td
                                          className='whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6 lg:pl-8'
                                      >
                                          TRUE
                                      </td>
                                  </tr>
                                  </tbody>
                              </table>
                          </div>
                      </div>
                  </div>
              </div>
          </div>
      </div>
    </WalletGrid>
  )
}
