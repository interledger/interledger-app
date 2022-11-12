import type { LoaderArgs } from '@remix-run/node'

import { WalletGrid } from '~/components'
import { grpcClient } from '~/lib/proto.server'
import { json } from '@remix-run/node'
import { useFetcher, useLoaderData } from '@remix-run/react'

export async function loader({ request }: LoaderArgs) {
  const signups = await grpcClient.listWaitlistSignups(
    {},
    {
      meta: {
        cookies: String(request.headers.get('cookie')) || ''
      }
    }
  )

  return json({
    signups: signups.response.signups
  })
}

export default function Page() {
  const allowSignup = useFetcher()
  const { signups } = useLoaderData<typeof loader>()
  const onClick = async (id: string, canSignup: boolean) => {
    if (!canSignup) {
      if (allowSignup.state === 'idle') {
        allowSignup.submit(
          { id: id },
          {
            method: 'post',
            action: '/api/allowSignup'
          }
        )
      }
    } else {
      await navigator.clipboard.writeText(
        `https://fynbos.app/signup?waitlistSignupId=${id}`
      )
    }
  }

  return (
    <WalletGrid>
      <div className='col-span-full flex flex-col rounded-2xl bg-page p-4 pb-8 sm:col-span-12 sm:col-start-2 lg:col-start-1'>
        <div className='px-4 sm:px-6 lg:px-8'>
          <div className='sm:flex sm:items-center'>
            <div className='sm:flex-auto'>
              <h1 className='text-xl font-semibold text-gray-900'>Waitlist</h1>
            </div>
          </div>
          <div className='mt-8 flex flex-col'>
            <div className='-my-2 -mx-4 sm:-mx-6 lg:-mx-8'>
              <div className='inline-block min-w-full py-2 align-middle'>
                <div className='shadow-sm ring-1 ring-black ring-opacity-5'>
                  <table
                    className='min-w-full border-separate'
                    style={{ borderSpacing: 0 }}
                  >
                    <thead className='bg-gray-50'>
                      <tr>
                        <th
                          scope='col'
                          className='sticky top-0 z-10 border-b border-gray-300 bg-gray-50 bg-opacity-75 py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 backdrop-blur backdrop-filter sm:pl-6 lg:pl-8'
                        >
                          Name
                        </th>
                        <th
                          scope='col'
                          className='sticky top-0 z-10 hidden border-b border-gray-300 bg-gray-50 bg-opacity-75 px-3 py-3.5 text-left text-sm font-semibold text-gray-900 backdrop-blur backdrop-filter sm:table-cell'
                        >
                          Email
                        </th>
                        <th
                          scope='col'
                          className='sticky top-0 z-10 hidden border-b border-gray-300 bg-gray-50 bg-opacity-75 px-3 py-3.5 text-left text-sm font-semibold text-gray-900 backdrop-blur backdrop-filter lg:table-cell'
                        >
                          Beta Opt In
                        </th>
                        <th
                          scope='col'
                          className='sticky top-0 z-10 hidden border-b border-gray-300 bg-gray-50 bg-opacity-75 px-3 py-3.5 text-left text-sm font-semibold text-gray-900 backdrop-blur backdrop-filter lg:table-cell'
                        >
                          Mug Id
                        </th>
                        <th
                          scope='col'
                          className='sticky top-0 z-10 hidden border-b border-gray-300 bg-gray-50 bg-opacity-75 px-3 py-3.5 text-left text-sm font-semibold text-gray-900 backdrop-blur backdrop-filter lg:table-cell'
                        >
                          Can Signup
                        </th>
                      </tr>
                    </thead>
                    <tbody className='bg-white'>
                      {signups.map((signup) => {
                        return (
                          <tr key={signup.id}>
                            <td className='whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6 lg:pl-8'>
                              {signup.name}
                            </td>
                            <td className='whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6 lg:pl-8'>
                              {signup.email}
                            </td>
                            <td className='whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6 lg:pl-8'>
                              {signup.betaOptIn ? 'TRUE' : 'FALSE'}
                            </td>
                            <td className='whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6 lg:pl-8'>
                              {signup.mugId}
                            </td>
                            <td
                              onClick={() =>
                                onClick(signup.id, signup.canSignup)
                              }
                              className='whitespace-nowrap cursor-pointer py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6 lg:pl-8'
                            >
                              {signup.canSignup ? 'TRUE' : 'FALSE'}
                            </td>
                          </tr>
                        )
                      })}
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
